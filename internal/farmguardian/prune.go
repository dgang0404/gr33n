package farmguardian

import (
	"context"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	db "gr33n-api/internal/db"
)

// Phase 27 WS5 follow-up — TTL pruning for Farm Guardian conversation
// history. Without this loop, gr33ncore.conversation_turns +
// gr33ncore.conversation_sessions grow unboundedly. The loop:
//
//   - reads CHAT_SESSION_TTL_DAYS (default 30, 0 disables, clamp 0..3650),
//   - reads CHAT_SESSION_PRUNE_INTERVAL_HOURS (default 24, clamp 1..168),
//   - runs an opening prune after CHAT_SESSION_PRUNE_STARTUP_DELAY_SECONDS
//     (default 30s) so we don't slam the DB on boot,
//   - then ticks once per interval until ctx is done.
//
// The default 30-day retention matches the operator-runbook guidance for
// "operational logs ≠ chat history" — chat history is operator-facing
// scratch, not an audit trail.

// PruneConfig controls the conversation TTL prune loop.
type PruneConfig struct {
	// TTLDays is the max age of a session before its turns + metadata row
	// are dropped. 0 disables the loop entirely.
	TTLDays int
	// Interval is how often the loop runs after the initial startup delay.
	Interval time.Duration
	// StartupDelay defers the very first prune so the API doesn't compete
	// with whatever else runs at boot (RAG ingest, automation worker, etc.).
	StartupDelay time.Duration
}

// PruneResult captures the per-run counts so callers can log.
type PruneResult struct {
	TurnsDeleted    int64
	SessionsDeleted int64
	Cutoff          time.Time
	Duration        time.Duration
}

// Enabled is true when the TTL loop should actually run.
func (c PruneConfig) Enabled() bool {
	return c.TTLDays > 0 && c.Interval > 0
}

// Defaults — adjusted via env in LoadPruneConfigFromEnv.
const (
	DefaultPruneTTLDays         = 30
	DefaultPruneIntervalHours   = 24
	DefaultPruneStartupDelaySec = 30
	MaxPruneTTLDays             = 3650 // ~10 years; anything beyond is just "never"
	MaxPruneIntervalHours       = 168  // 1 week
)

// LoadPruneConfigFromEnv parses the three env vars with clamping and falls
// back to safe defaults so a fresh install gets sane retention without any
// .env edits.
func LoadPruneConfigFromEnv() PruneConfig {
	return PruneConfig{
		TTLDays:      pruneTTLFromEnv(),
		Interval:     pruneIntervalFromEnv(),
		StartupDelay: pruneStartupDelayFromEnv(),
	}
}

func pruneTTLFromEnv() int {
	raw := strings.TrimSpace(os.Getenv("CHAT_SESSION_TTL_DAYS"))
	if raw == "" {
		return DefaultPruneTTLDays
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return DefaultPruneTTLDays
	}
	if n > MaxPruneTTLDays {
		return MaxPruneTTLDays
	}
	return n
}

func pruneIntervalFromEnv() time.Duration {
	raw := strings.TrimSpace(os.Getenv("CHAT_SESSION_PRUNE_INTERVAL_HOURS"))
	if raw == "" {
		return time.Duration(DefaultPruneIntervalHours) * time.Hour
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 {
		return time.Duration(DefaultPruneIntervalHours) * time.Hour
	}
	if n > MaxPruneIntervalHours {
		n = MaxPruneIntervalHours
	}
	return time.Duration(n) * time.Hour
}

func pruneStartupDelayFromEnv() time.Duration {
	raw := strings.TrimSpace(os.Getenv("CHAT_SESSION_PRUNE_STARTUP_DELAY_SECONDS"))
	if raw == "" {
		return time.Duration(DefaultPruneStartupDelaySec) * time.Second
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return time.Duration(DefaultPruneStartupDelaySec) * time.Second
	}
	if n > 600 { // 10-minute cap, just in case someone fat-fingers it
		n = 600
	}
	return time.Duration(n) * time.Second
}

// prunableQuerier is the slice of *db.Queries we actually need so tests can
// supply a fake without touching every method on Queries.
type prunableQuerier interface {
	DeleteStaleConversationTurns(ctx context.Context, cutoff time.Time) (int64, error)
	DeleteStaleConversationSessions(ctx context.Context, cutoff time.Time) (int64, error)
}

// PruneOnce removes turns + session rows whose latest activity is older than
// `ttl` from now. Returns counts + cutoff for logging. Errors from the turn
// pass short-circuit (the session pass would otherwise leak dangling FK-free
// metadata rows, but they're cheap and the next loop iteration retries).
func PruneOnce(ctx context.Context, q prunableQuerier, ttl time.Duration) (PruneResult, error) {
	start := time.Now()
	cutoff := start.Add(-ttl)
	res := PruneResult{Cutoff: cutoff}
	turns, err := q.DeleteStaleConversationTurns(ctx, cutoff)
	if err != nil {
		return res, err
	}
	res.TurnsDeleted = turns
	sessions, err := q.DeleteStaleConversationSessions(ctx, cutoff)
	if err != nil {
		res.Duration = time.Since(start)
		return res, err
	}
	res.SessionsDeleted = sessions
	res.Duration = time.Since(start)
	return res, nil
}

// StartPruneLoop runs the TTL loop until ctx is done. Caller decides
// whether to spawn it as a goroutine. A nil logger falls back to the
// default slog logger so callers can wire structured logging without
// boilerplate.
//
// The loop:
//  1. sleeps cfg.StartupDelay,
//  2. runs PruneOnce, logs result,
//  3. ticks cfg.Interval, repeating step 2,
//  4. returns when ctx is cancelled.
func StartPruneLoop(ctx context.Context, q prunableQuerier, cfg PruneConfig, logger *slog.Logger) {
	if !cfg.Enabled() {
		return
	}
	if logger == nil {
		logger = slog.Default()
	}
	ttl := time.Duration(cfg.TTLDays) * 24 * time.Hour

	// Startup delay — respect ctx so a fast SIGTERM doesn't wait it out.
	if cfg.StartupDelay > 0 {
		t := time.NewTimer(cfg.StartupDelay)
		select {
		case <-ctx.Done():
			t.Stop()
			return
		case <-t.C:
		}
	}

	pruneAndLog(ctx, q, ttl, cfg.TTLDays, logger)

	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pruneAndLog(ctx, q, ttl, cfg.TTLDays, logger)
		}
	}
}

func pruneAndLog(ctx context.Context, q prunableQuerier, ttl time.Duration, ttlDays int, logger *slog.Logger) {
	res, err := PruneOnce(ctx, q, ttl)
	if err != nil {
		logger.Warn("chat session prune failed",
			"ttl_days", ttlDays,
			"turns_deleted", res.TurnsDeleted,
			"sessions_deleted", res.SessionsDeleted,
			"err", err,
		)
		return
	}
	// Only emit an info line when we actually removed something — quiet
	// happy path avoids 30-day log spam on small installs.
	if res.TurnsDeleted == 0 && res.SessionsDeleted == 0 {
		logger.Debug("chat session prune ran (no-op)",
			"ttl_days", ttlDays,
			"duration_ms", res.Duration.Milliseconds(),
		)
		return
	}
	logger.Info("chat session prune",
		"ttl_days", ttlDays,
		"turns_deleted", res.TurnsDeleted,
		"sessions_deleted", res.SessionsDeleted,
		"cutoff", res.Cutoff.UTC().Format(time.RFC3339),
		"duration_ms", res.Duration.Milliseconds(),
	)
}

// Compile-time sanity check: *db.Queries satisfies prunableQuerier.
var _ prunableQuerier = (*db.Queries)(nil)
