package farmguardian

import (
	"context"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	db "gr33n-api/internal/db"
)

// Phase 27 WS5 follow-up — cost guards on Farm Guardian token usage.
//
// Cost guards are a rolling-window cap on the total tokens
// (prompt_tokens + completion_tokens) a user (or a farm) is allowed to
// burn in a given period. When the cap is hit, `POST /v1/chat` returns
// HTTP 429 with a Retry-After header pointing at the end of the window.
//
// Defaults are intentionally generous — a farm running a single operator
// chatting on a 70B model on-prem isn't trying to outrun OpenAI billing,
// so the goal here is "catch a runaway loop / scripted abuse" rather than
// "enforce a billing budget". Operators tighten the caps via env vars on
// shared / multi-farm deployments.
//
// Both caps default to 0 = disabled unless GUARDIAN_COST_GUARD enables prod defaults.

// CostGuardConfig controls per-user / per-farm token budgets.
type CostGuardConfig struct {
	// Window is how far back the cap is measured. 1h = "no more than X
	// tokens in any 60-minute window".
	Window time.Duration
	// PerUserMaxTokens caps the (prompt + completion) total per user
	// across every session they own. 0 disables the per-user guard.
	PerUserMaxTokens int64
	// PerFarmMaxTokens caps the per-farm total — only checked when the
	// chat request includes a farm_id. 0 disables the per-farm guard.
	PerFarmMaxTokens int64
}

// Defaults & clamps.
const (
	DefaultCostWindowHours      = 24
	MaxCostWindowHours          = 168 // 1 week
	MaxCostMaxTokens            = 100000000 // 100 M / window — anything beyond is "off"
	DefaultPerUserMaxTokens     = 200000
	DefaultPerFarmMaxTokens     = 0 // disabled unless set
)

// AnyEnabled is true when at least one of the two caps is set.
func (c CostGuardConfig) AnyEnabled() bool {
	return c.PerUserMaxTokens > 0 || c.PerFarmMaxTokens > 0
}

// LoadCostGuardConfigFromEnv reads env vars. When GUARDIAN_COST_GUARD is unset,
// guards are enabled in production-style installs (200k tokens/user/day) and
// disabled for dev/auth_test. Set GUARDIAN_COST_GUARD=off to opt out explicitly.
func LoadCostGuardConfigFromEnv() CostGuardConfig {
	if !costGuardEnabledFromEnv() {
		return CostGuardConfig{Window: costWindowFromEnv()}
	}
	return CostGuardConfig{
		Window:           costWindowFromEnv(),
		PerUserMaxTokens: costMaxFromEnv("CHAT_COST_MAX_TOKENS_PER_USER", DefaultPerUserMaxTokens),
		PerFarmMaxTokens: costMaxFromEnv("CHAT_COST_MAX_TOKENS_PER_FARM", DefaultPerFarmMaxTokens),
	}
}

func costGuardEnabledFromEnv() bool {
	raw := strings.ToLower(strings.TrimSpace(os.Getenv("GUARDIAN_COST_GUARD")))
	switch raw {
	case "off", "false", "0", "disabled":
		return false
	case "on", "true", "1", "enabled":
		return true
	default:
		mode := strings.ToLower(strings.TrimSpace(os.Getenv("AUTH_MODE")))
		return mode != "dev" && mode != "auth_test"
	}
}

func costWindowFromEnv() time.Duration {
	raw := strings.TrimSpace(os.Getenv("CHAT_COST_WINDOW_HOURS"))
	if raw == "" {
		return time.Duration(DefaultCostWindowHours) * time.Hour
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 {
		return time.Duration(DefaultCostWindowHours) * time.Hour
	}
	if n > MaxCostWindowHours {
		n = MaxCostWindowHours
	}
	return time.Duration(n) * time.Hour
}

func costMaxFromEnv(name string, fallback int64) int64 {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback
	}
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || n < 0 {
		return fallback
	}
	if n > MaxCostMaxTokens {
		return MaxCostMaxTokens
	}
	return n
}

// Decision is the outcome of a CheckBudget call. When Allowed is false,
// the caller should reject the chat request with HTTP 429.
type Decision struct {
	Allowed bool
	// Reason identifies which cap fired ("per_user" or "per_farm");
	// empty when Allowed is true.
	Reason string
	// UsedTokens is the rolling-window total that triggered the rejection.
	UsedTokens int64
	// MaxTokens is the configured cap for the firing dimension.
	MaxTokens int64
	// WindowSeconds is the window length, exposed to clients via the
	// response body so they can show "wait until X" without doing math.
	WindowSeconds int64
	// RetryAfter is how long the client should wait before retrying. We
	// use the configured window length — pessimistic but stable, and
	// avoids leaking the oldest-turn timestamp to the API surface.
	RetryAfter time.Duration
}

// costQuerier is the slice of *db.Queries the guard actually needs so
// tests can supply a fake without faking every Queries method.
type costQuerier interface {
	SumChatTokensSinceForUser(ctx context.Context, arg db.SumChatTokensSinceForUserParams) (db.SumChatTokensSinceForUserRow, error)
	SumChatTokensSinceForFarm(ctx context.Context, arg db.SumChatTokensSinceForFarmParams) (db.SumChatTokensSinceForFarmRow, error)
}

// CheckBudget runs the rolling-window queries the config asks for and
// returns the first cap that's blown (per-user takes precedence over
// per-farm so a runaway user can't hide behind a farm with headroom).
// When neither cap is configured, the call returns Allowed = true without
// touching the DB.
//
// farmID == 0 disables the per-farm check for this call (used for plain
// non-grounded turns).
func CheckBudget(ctx context.Context, q costQuerier, cfg CostGuardConfig, userID uuid.UUID, farmID int64) (Decision, error) {
	if !cfg.AnyEnabled() {
		return Decision{Allowed: true}, nil
	}
	windowSec := int64(cfg.Window / time.Second)
	since := time.Now().Add(-cfg.Window)

	if cfg.PerUserMaxTokens > 0 {
		totals, err := q.SumChatTokensSinceForUser(ctx, db.SumChatTokensSinceForUserParams{UserID: userID, Since: since})
		if err != nil {
			return Decision{}, err
		}
		if totals.TotalTokens >= cfg.PerUserMaxTokens {
			return Decision{
				Allowed:       false,
				Reason:        "per_user",
				UsedTokens:    totals.TotalTokens,
				MaxTokens:     cfg.PerUserMaxTokens,
				WindowSeconds: windowSec,
				RetryAfter:    cfg.Window,
			}, nil
		}
	}
	if cfg.PerFarmMaxTokens > 0 && farmID > 0 {
		totals, err := q.SumChatTokensSinceForFarm(ctx, db.SumChatTokensSinceForFarmParams{FarmID: &farmID, Since: since})
		if err != nil {
			return Decision{}, err
		}
		if totals.TotalTokens >= cfg.PerFarmMaxTokens {
			return Decision{
				Allowed:       false,
				Reason:        "per_farm",
				UsedTokens:    totals.TotalTokens,
				MaxTokens:     cfg.PerFarmMaxTokens,
				WindowSeconds: windowSec,
				RetryAfter:    cfg.Window,
			}, nil
		}
	}
	return Decision{Allowed: true}, nil
}

// Compile-time sanity check: *db.Queries satisfies costQuerier.
var _ costQuerier = (*db.Queries)(nil)
