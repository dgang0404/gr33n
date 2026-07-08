// Phase 149 WS1 — deterministic alert citation ordering.
//
// Run #6 showed phi3:mini attaching the wrong [n] to an alert claim when
// alert_notification chunks arrive in semantic-similarity order rather than
// severity order. Models tend to enumerate "most urgent first"; when [1] is
// not the most severe alert, the model's own list-writing instinct and the
// citation numbers it was given fall out of sync. Sorting alert_notification
// chunks to the front by severity before numbering removes that mismatch by
// construction instead of only detecting it after the fact (Phase 148).

package farmguardian

import (
	"strings"

	db "gr33n-api/internal/db"
)

var alertSeverityRank = map[string]int{
	"critical": 0,
	"high":     1,
	"medium":   2,
	"low":      3,
}

// PrioritizeAlertChunks moves alert_notification chunks to the front of the
// list, ordered by severity (most severe first), leaving all other chunks
// in their original relative order after. Non-alert retrieval and answers
// with fewer than two alerts are returned unchanged.
func PrioritizeAlertChunks(chunks []db.SearchRagNearestNeighborsFilteredRow) []db.SearchRagNearestNeighborsFilteredRow {
	if len(chunks) < 2 {
		return chunks
	}
	var alerts, rest []db.SearchRagNearestNeighborsFilteredRow
	for _, c := range chunks {
		if c.SourceType == SourceTypeAlertNotification {
			alerts = append(alerts, c)
		} else {
			rest = append(rest, c)
		}
	}
	if len(alerts) < 2 {
		return chunks
	}
	sortAlertsBySeverityDesc(alerts)
	out := make([]db.SearchRagNearestNeighborsFilteredRow, 0, len(chunks))
	out = append(out, alerts...)
	out = append(out, rest...)
	return out
}

// SourceTypeAlertNotification mirrors internal/rag/ingest's source type constant
// (kept local to avoid an import cycle with the ingest package).
const SourceTypeAlertNotification = "alert_notification"

func sortAlertsBySeverityDesc(alerts []db.SearchRagNearestNeighborsFilteredRow) {
	// Stable insertion sort — alert counts are small (<=5) and we want ties
	// to keep their original (semantic-relevance) relative order.
	for i := 1; i < len(alerts); i++ {
		for j := i; j > 0 && severityRank(alerts[j]) < severityRank(alerts[j-1]); j-- {
			alerts[j], alerts[j-1] = alerts[j-1], alerts[j]
		}
	}
}

func severityRank(c db.SearchRagNearestNeighborsFilteredRow) int {
	for _, line := range strings.Split(c.ContentText, "\n") {
		line = strings.TrimSpace(strings.ToLower(line))
		if strings.HasPrefix(line, "severity:") {
			sev := strings.TrimSpace(strings.TrimPrefix(line, "severity:"))
			if r, ok := alertSeverityRank[sev]; ok {
				return r
			}
			return 2
		}
	}
	return 2
}
