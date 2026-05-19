// Phase 28 WS4 — Farm Guardian alert detail query. Hand-written Go
// binding (same pattern as Phase 27 conversation_turns + Phase 28 WS1
// crop_cycle_analytics) to avoid a repo-wide sqlc regen. The SQL
// definition lives in db/queries/alerts.sql alongside the other alert
// queries — when the next routine sqlc pass happens this file folds
// cleanly back into alerts.sql.go.

package db

import (
	"context"
	"time"
)

const listRecentUnreadAlertsByFarm = `-- name: ListRecentUnreadAlertsByFarm :many
SELECT
    id,
    severity,
    subject_rendered,
    message_text_rendered,
    triggering_event_source_type,
    triggering_event_source_id,
    created_at
FROM gr33ncore.alerts_notifications
WHERE farm_id = $1 AND is_read = FALSE
ORDER BY severity DESC NULLS LAST, created_at DESC
LIMIT $2
`

// RecentUnreadAlertSummary is the prompt-ready projection of an
// alerts_notifications row for Farm Guardian's live snapshot. Only the
// fields the LLM actually needs to explain the alert are surfaced — the
// rest of the row (delivery_attempts, status, html, scheduled_send_at,
// recipient_user_id, etc.) is intentionally omitted to keep the struct
// allocation-light and the prompt budget predictable.
type RecentUnreadAlertSummary struct {
	ID                        int64                                 `db:"id" json:"id"`
	Severity                  NullGr33ncoreNotificationPriorityEnum `db:"severity" json:"severity"`
	SubjectRendered           *string                               `db:"subject_rendered" json:"subject_rendered"`
	MessageTextRendered       *string                               `db:"message_text_rendered" json:"message_text_rendered"`
	TriggeringEventSourceType *string                               `db:"triggering_event_source_type" json:"triggering_event_source_type"`
	TriggeringEventSourceID   *int64                                `db:"triggering_event_source_id" json:"triggering_event_source_id"`
	CreatedAt                 time.Time                             `db:"created_at" json:"created_at"`
}

// ListRecentUnreadAlertsByFarm returns the top `limit` unread alerts for
// a farm, ordered by severity DESC then created_at DESC. Used by Farm
// Guardian's snapshot builder (Phase 28 WS4) to feed the LLM enough
// detail to *explain* the alert ("you have a high-humidity alert in
// the Flower Room triggered 2h ago — 72% RH vs 65% threshold") instead
// of just reporting the count.
func (q *Queries) ListRecentUnreadAlertsByFarm(ctx context.Context, farmID int64, limit int32) ([]RecentUnreadAlertSummary, error) {
	rows, err := q.db.Query(ctx, listRecentUnreadAlertsByFarm, farmID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []RecentUnreadAlertSummary{}
	for rows.Next() {
		var i RecentUnreadAlertSummary
		if err := rows.Scan(
			&i.ID,
			&i.Severity,
			&i.SubjectRendered,
			&i.MessageTextRendered,
			&i.TriggeringEventSourceType,
			&i.TriggeringEventSourceID,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
