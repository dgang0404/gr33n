// Phase 28 WS5 — chat token budget warning. When a chat turn pushes a
// user past `WarningThresholdPct` of their configured per-user cap, the
// platform fires a single notification into gr33ncore.alerts_notifications
// so the existing alert channel (UI, push, etc.) surfaces it without
// requiring operators to poll the new /v1/chat/usage endpoint.
//
// Debounce: at most one warning per user per cost-guard window.
// `MaybeFireBudgetWarning` checks the DB for an existing warning before
// inserting; the check fails open (returns no error when the warning
// can't be checked / inserted) so a transient hiccup never breaks chat.

package farmguardian

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "gr33n-api/internal/db"
)

// WarningThresholdPct is the proportion of the per-user cap at which a
// chat_budget_warning alert is fired. 0.80 == 80%.
const WarningThresholdPct = 0.80

// ChatBudgetWarningSourceType is the literal string written into
// alerts_notifications.triggering_event_source_type. The alerts UI / API
// can filter on this value to surface chat-budget-warnings distinctly
// from sensor / rule / program alerts.
const ChatBudgetWarningSourceType = "chat_budget_warning"

// budgetWarningQuerier is the slice of *db.Queries the warning hook
// needs — split out so tests can stub it without faking every method on
// Queries. Methods mirror the real signatures verbatim.
type budgetWarningQuerier interface {
	SumChatTokensSinceForUser(ctx context.Context, userID uuid.UUID, since time.Time) (db.ChatTokenTotals, error)
	GetRecentChatBudgetWarningForUser(ctx context.Context, recipientUserID uuid.UUID, since time.Time) (int64, error)
	CreateAlert(ctx context.Context, arg db.CreateAlertParams) (db.Gr33ncoreAlertsNotification, error)
}

// WarningResult reports what MaybeFireBudgetWarning did. It is the test
// seam — production callers can discard it.
type WarningResult struct {
	// Fired is true when a new chat_budget_warning row was inserted on
	// this call. False means either: no cap configured, below
	// threshold, debounce hit, no farm_id supplied, or a transient
	// query error (best-effort).
	Fired bool
	// PctUsed is the user's rolling-window utilisation expressed as a
	// fraction (0.85 == 85%). 0 when no cap is configured.
	PctUsed float64
	// UsedTokens is the rolling-window total at decision time.
	UsedTokens int64
	// MaxTokens echoes the per-user cap (0 when disabled).
	MaxTokens int64
	// AlertID is the new alert's id when Fired == true; 0 otherwise.
	AlertID int64
}

// MaybeFireBudgetWarning runs the threshold check + debounce check + alert
// insert as a single best-effort unit. Designed to be called from the
// chat handler immediately after a successful conversation_turns insert,
// when the new rolling total reflects the just-completed turn.
//
// Contract:
//   - Returns Fired=false when: cost guard per-user cap is disabled,
//     user under threshold, an existing warning is already in-window,
//     or there's no farm_id to pin the alert to.
//   - Returns Fired=true when a new alerts_notifications row was inserted.
//   - All DB errors are non-fatal: the function returns a zero
//     WarningResult and a nil error so the chat turn keeps flowing.
//     (A non-nil error is only returned for programmer mistakes like a
//     nil queries handle — easy to fail loud in tests.)
//
// farmID is required because gr33ncore.alerts_notifications.farm_id is
// NOT NULL. Plain (ungrounded) turns skip the warning entirely — they
// don't carry a farm context and warnings are scoped to the farm whose
// dashboard the operator will be looking at. This is a deliberate
// trade-off (see Phase 28 plan WS5 § "warning targeting").
func MaybeFireBudgetWarning(
	ctx context.Context,
	q budgetWarningQuerier,
	cfg CostGuardConfig,
	userID uuid.UUID,
	farmID int64,
) (WarningResult, error) {
	if q == nil {
		return WarningResult{}, errors.New("MaybeFireBudgetWarning: nil queries handle")
	}
	if cfg.PerUserMaxTokens <= 0 {
		return WarningResult{}, nil
	}
	if farmID <= 0 {
		// No farm context — alerts_notifications requires farm_id and
		// the operator wouldn't see the warning surfaced anywhere
		// useful anyway. /v1/chat/usage still reflects the totals.
		return WarningResult{}, nil
	}
	if userID == uuid.Nil {
		return WarningResult{}, nil
	}

	since := time.Now().Add(-cfg.Window)

	totals, err := q.SumChatTokensSinceForUser(ctx, userID, since)
	if err != nil {
		// Best-effort — the cost guard itself will catch the user at
		// the hard cap on the next turn even if we miss the 80%
		// warning here.
		return WarningResult{}, nil
	}

	pct := float64(totals.TotalTokens) / float64(cfg.PerUserMaxTokens)
	if pct < WarningThresholdPct {
		return WarningResult{
			PctUsed:    pct,
			UsedTokens: totals.TotalTokens,
			MaxTokens:  cfg.PerUserMaxTokens,
		}, nil
	}

	// Debounce — only one warning per user per window. We DON'T treat
	// a non-pgx.ErrNoRows error as "go ahead and fire": if the lookup
	// itself errors we'd rather skip the warning than risk spamming.
	if _, derr := q.GetRecentChatBudgetWarningForUser(ctx, userID, since); derr == nil {
		// Existing warning in-window → debounce hit.
		return WarningResult{
			PctUsed:    pct,
			UsedTokens: totals.TotalTokens,
			MaxTokens:  cfg.PerUserMaxTokens,
		}, nil
	} else if !errors.Is(derr, pgx.ErrNoRows) {
		return WarningResult{
			PctUsed:    pct,
			UsedTokens: totals.TotalTokens,
			MaxTokens:  cfg.PerUserMaxTokens,
		}, nil
	}

	subject := fmt.Sprintf("Chat token budget at %d%%", int(pct*100+0.5))
	message := fmt.Sprintf(
		"You've used %d of %d tokens in the last %d hour(s) of Farm Guardian chat. The cap will reject further turns once you hit 100%%.",
		totals.TotalTokens, cfg.PerUserMaxTokens, int(cfg.Window/time.Hour),
	)
	srcType := ChatBudgetWarningSourceType
	severity := db.NullGr33ncoreNotificationPriorityEnum{
		Gr33ncoreNotificationPriorityEnum: db.Gr33ncoreNotificationPriorityEnumMedium,
		Valid:                             true,
	}
	row, ierr := q.CreateAlert(ctx, db.CreateAlertParams{
		FarmID:                    farmID,
		RecipientUserID:           pgtype.UUID{Bytes: userID, Valid: true},
		TriggeringEventSourceType: &srcType,
		TriggeringEventSourceID:   nil,
		Severity:                  severity,
		SubjectRendered:           &subject,
		MessageTextRendered:       &message,
	})
	if ierr != nil {
		return WarningResult{
			PctUsed:    pct,
			UsedTokens: totals.TotalTokens,
			MaxTokens:  cfg.PerUserMaxTokens,
		}, nil
	}
	return WarningResult{
		Fired:      true,
		PctUsed:    pct,
		UsedTokens: totals.TotalTokens,
		MaxTokens:  cfg.PerUserMaxTokens,
		AlertID:    row.ID,
	}, nil
}

// Compile-time sanity check: *db.Queries satisfies budgetWarningQuerier.
var _ budgetWarningQuerier = (*db.Queries)(nil)
