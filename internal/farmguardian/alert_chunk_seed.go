package farmguardian

import (
	"context"
	"strings"

	db "gr33n-api/internal/db"
)

const liveUnreadAlertSeedLimit = 8

// SeedAlertChunksFromLiveUnread prepends citeable alert_notification rows from
// live unread alerts when an alert-summary question retrieved fewer than 2 RAG
// alert chunks. Without this, list_unread_alerts supplies prose that is marked
// "do not cite" and InjectAlertCitationRefs / BuildCitations stay empty —
// smoke-unread-alerts fails despite good farm content.
func SeedAlertChunksFromLiveUnread(
	ctx context.Context,
	q db.Querier,
	farmID int64,
	question string,
	chunks []db.SearchRagNearestNeighborsFilteredRow,
) []db.SearchRagNearestNeighborsFilteredRow {
	if !MatchAlertSummaryIntent(question) || countAlertChunks(chunks) >= 2 {
		return chunks
	}
	if q == nil || farmID <= 0 {
		return chunks
	}
	alerts, err := q.ListRecentUnreadAlertsByFarm(ctx, db.ListRecentUnreadAlertsByFarmParams{
		FarmID: farmID,
		Limit:  liveUnreadAlertSeedLimit,
	})
	if err != nil || len(alerts) < 2 {
		return chunks
	}
	return MergeLiveUnreadAlertChunks(chunks, alerts)
}

// MergeLiveUnreadAlertChunks builds synthetic RAG-shaped rows from live unread
// alerts and places them ahead of existing chunks (deduping by SourceID).
func MergeLiveUnreadAlertChunks(
	chunks []db.SearchRagNearestNeighborsFilteredRow,
	alerts []db.ListRecentUnreadAlertsByFarmRow,
) []db.SearchRagNearestNeighborsFilteredRow {
	if len(alerts) == 0 {
		return chunks
	}
	seen := map[int64]bool{}
	for _, c := range chunks {
		if c.SourceType == SourceTypeAlertNotification {
			seen[c.SourceID] = true
		}
	}
	var seeded []db.SearchRagNearestNeighborsFilteredRow
	for _, a := range alerts {
		if seen[a.ID] {
			continue
		}
		seen[a.ID] = true
		seeded = append(seeded, db.SearchRagNearestNeighborsFilteredRow{
			ID:          a.ID, // stable positive id for citation routing
			FarmID:      0,
			SourceType:  SourceTypeAlertNotification,
			SourceID:    a.ID,
			ChunkIndex:  0,
			ContentText: unreadAlertCitationText(a),
		})
	}
	if len(seeded) == 0 {
		return chunks
	}
	out := make([]db.SearchRagNearestNeighborsFilteredRow, 0, len(seeded)+len(chunks))
	out = append(out, seeded...)
	out = append(out, chunks...)
	return out
}

func unreadAlertCitationText(a db.ListRecentUnreadAlertsByFarmRow) string {
	var b strings.Builder
	b.WriteString("alert_notification\n")
	if a.Severity != nil {
		b.WriteString("severity: ")
		b.WriteString(string(*a.Severity))
		b.WriteByte('\n')
	}
	if a.SubjectRendered != nil && strings.TrimSpace(*a.SubjectRendered) != "" {
		b.WriteString("subject: ")
		b.WriteString(strings.TrimSpace(*a.SubjectRendered))
		b.WriteByte('\n')
	}
	if a.MessageTextRendered != nil && strings.TrimSpace(*a.MessageTextRendered) != "" {
		b.WriteString("message: ")
		b.WriteString(strings.TrimSpace(*a.MessageTextRendered))
		b.WriteByte('\n')
	}
	if a.TriggeringEventSourceType != nil && strings.TrimSpace(*a.TriggeringEventSourceType) != "" {
		b.WriteString("triggering_event_source_type: ")
		b.WriteString(strings.TrimSpace(*a.TriggeringEventSourceType))
		b.WriteByte('\n')
	}
	return strings.TrimSpace(b.String())
}
