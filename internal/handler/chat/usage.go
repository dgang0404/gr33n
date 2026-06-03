// Phase 28 WS5 — GET /v1/chat/usage. Surfaces the caller's rolling-window
// token totals + remaining budget so operators on shared deployments
// (multiple staff hitting one Guardian) can see who is burning the
// budget without diving into the DB.
//
// Request shape:
//
//	GET /v1/chat/usage              → user dimension only
//	GET /v1/chat/usage?farm_id=7    → user + farm dimension (farm-member gated)
//
// Response shape (200):
//
//	{
//	  "window_hours": 1,
//	  "ai_enabled": true,
//	  "user": {
//	    "used_tokens":      3200,
//	    "max_tokens":      10000,   // 0 = no cap configured
//	    "remaining_tokens": 6800,   // 0 when uncapped or used >= max
//	    "pct_used":         0.32,   // 0.0..1.0 (0 when uncapped)
//	    "warning_threshold_pct": 0.80
//	  },
//	  "farm": {                     // omitted when farm_id is absent
//	    "farm_id": 7,
//	    "used_tokens":     14000,
//	    "max_tokens":      50000,
//	    "remaining_tokens": 36000,
//	    "pct_used":         0.28
//	  }
//	}
//
// Auth: JWT. The farm dimension additionally requires farm membership.
// When AI is disabled at the server level the endpoint returns 503 so
// the UI can hide the card cleanly.

package chat

import (
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"gr33n-api/internal/authctx"
	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/farmguardian"
	"gr33n-api/internal/httputil"
)

// usageDimension is the JSON shape for a single budget axis.
type usageDimension struct {
	UsedTokens          int64    `json:"used_tokens"`
	MaxTokens           int64    `json:"max_tokens"`
	RemainingTokens     int64    `json:"remaining_tokens"`
	PctUsed             float64  `json:"pct_used"`
	WarningThresholdPct *float64 `json:"warning_threshold_pct,omitempty"`
}

type usageFarmDimension struct {
	FarmID int64 `json:"farm_id"`
	usageDimension
}

type usageResponse struct {
	WindowHours int                 `json:"window_hours"`
	AIEnabled   bool                `json:"ai_enabled"`
	User        usageDimension      `json:"user"`
	Farm        *usageFarmDimension `json:"farm,omitempty"`
}

// GetUsage handles GET /v1/chat/usage. Phase 28 WS5.
func (h *Handler) GetUsage(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.Enabled {
		// Same Lite-mode posture as the rest of the chat surface —
		// the UI gates on /capabilities so this 503 is a backstop.
		httputil.WriteError(w, http.StatusServiceUnavailable, "AI is disabled on this deployment")
		return
	}
	userID, hasUser := authctx.UserID(r.Context())
	if !hasUser {
		httputil.WriteError(w, http.StatusUnauthorized, "user context missing")
		return
	}

	cfg := h.costGuard
	windowHours := int(cfg.Window / time.Hour)
	if windowHours < 1 {
		windowHours = farmguardian.DefaultCostWindowHours
	}

	resp := usageResponse{
		WindowHours: windowHours,
		AIEnabled:   true,
	}

	// User dimension is always present. When the per-user cap is
	// disabled we still report used_tokens so operators can see
	// volume without committing to a cap.
	if h.q != nil {
		resp.User = buildUserDimension(r, h.q, userID, cfg)
	}

	// Optional ?farm_id=N. Must parse cleanly and pass the farm-member
	// check before we expose any per-farm numbers (multi-farm
	// deployments shouldn't leak utilisation across operators).
	if raw := r.URL.Query().Get("farm_id"); raw != "" {
		farmID, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || farmID <= 0 {
			httputil.WriteError(w, http.StatusBadRequest, "invalid farm_id")
			return
		}
		if !farmauthz.RequireFarmMember(w, r, h.q, farmID) {
			return
		}
		if h.q != nil {
			fd := buildFarmDimension(r, h.q, farmID, cfg)
			resp.Farm = &fd
		}
	}

	httputil.WriteJSON(w, http.StatusOK, resp)
}

// buildUserDimension is the per-user half of the response. Failures
// are logged at WARN and rendered as zeros — the endpoint never 500s on
// a transient SUM query hiccup.
func buildUserDimension(r *http.Request, q *db.Queries, userID uuid.UUID, cfg farmguardian.CostGuardConfig) usageDimension {
	since := time.Now().Add(-cfg.Window)
	totals, err := q.SumChatTokensSinceForUser(r.Context(), db.SumChatTokensSinceForUserParams{UserID: userID, Since: since})
	if err != nil {
		slog.Warn("chat usage per-user totals failed", "user_id", userID, "err", err)
		return usageDimension{MaxTokens: cfg.PerUserMaxTokens}
	}
	d := usageDimension{
		UsedTokens: totals.TotalTokens,
		MaxTokens:  cfg.PerUserMaxTokens,
	}
	if cfg.PerUserMaxTokens > 0 {
		d.RemainingTokens = cfg.PerUserMaxTokens - totals.TotalTokens
		if d.RemainingTokens < 0 {
			d.RemainingTokens = 0
		}
		d.PctUsed = roundPct(float64(totals.TotalTokens) / float64(cfg.PerUserMaxTokens))
		thr := farmguardian.WarningThresholdPct
		d.WarningThresholdPct = &thr
	}
	return d
}

// buildFarmDimension mirrors buildUserDimension but scopes the SUM to a
// farm. Plain (ungrounded) turns are excluded because their
// conversation_turns.farm_id is NULL — this is the same shape the cost
// guard's per-farm check uses.
func buildFarmDimension(r *http.Request, q *db.Queries, farmID int64, cfg farmguardian.CostGuardConfig) usageFarmDimension {
	since := time.Now().Add(-cfg.Window)
	totals, err := q.SumChatTokensSinceForFarm(r.Context(), db.SumChatTokensSinceForFarmParams{FarmID: &farmID, Since: since})
	if err != nil {
		slog.Warn("chat usage per-farm totals failed", "farm_id", farmID, "err", err)
		return usageFarmDimension{FarmID: farmID, usageDimension: usageDimension{MaxTokens: cfg.PerFarmMaxTokens}}
	}
	d := usageDimension{
		UsedTokens: totals.TotalTokens,
		MaxTokens:  cfg.PerFarmMaxTokens,
	}
	if cfg.PerFarmMaxTokens > 0 {
		d.RemainingTokens = cfg.PerFarmMaxTokens - totals.TotalTokens
		if d.RemainingTokens < 0 {
			d.RemainingTokens = 0
		}
		d.PctUsed = roundPct(float64(totals.TotalTokens) / float64(cfg.PerFarmMaxTokens))
	}
	return usageFarmDimension{FarmID: farmID, usageDimension: d}
}

// roundPct rounds to 4 decimal places so the JSON payload stays
// compact (0.3217 vs 0.32173913...). The UI renders it as a percentage
// so two extra digits past the percent point are plenty.
func roundPct(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0
	}
	return math.Round(v*10000) / 10000
}
