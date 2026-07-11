// Phase 151 WS4 — alert-summary retrieval narrowing and intent matching.

package farmguardian

import (
	"regexp"
	"strings"

	db "gr33n-api/internal/db"
)

var summarizeAlertsIntent = regexp.MustCompile(`(?i)\bsummar(?:y|ize|ies).*\b(unread\s+)?alerts?\b|\b(unread\s+)?alerts?\b.*\bwhat (should|to do)\b`)

// MatchAlertSummaryIntent reports whether the operator wants a per-alert summary
// (list, summarize, or what-to-do), including the smoke-unread-alerts fixture.
func MatchAlertSummaryIntent(question string) bool {
	q := strings.TrimSpace(question)
	if q == "" {
		return false
	}
	return matchListUnreadAlertsIntent(q) || summarizeAlertsIntent.MatchString(q)
}

// FilterChunksForAlertSummary returns only alert_notification chunks in the
// numbered Sources list when the question is an alert summary and retrieval
// returned 2+ alerts. Removes platform_doc / field_guide from numbering so
// item 1 → [1] is unambiguous (run #6 had workflow docs at [1]/[2]).
func FilterChunksForAlertSummary(question string, chunks []db.SearchRagNearestNeighborsFilteredRow) []db.SearchRagNearestNeighborsFilteredRow {
	if !MatchAlertSummaryIntent(question) {
		return chunks
	}
	alerts := alertChunksOnly(chunks)
	if len(alerts) < 2 {
		return chunks
	}
	return alerts
}

func alertChunksOnly(chunks []db.SearchRagNearestNeighborsFilteredRow) []db.SearchRagNearestNeighborsFilteredRow {
	var out []db.SearchRagNearestNeighborsFilteredRow
	for _, c := range chunks {
		if c.SourceType == SourceTypeAlertNotification {
			out = append(out, c)
		}
	}
	return out
}

func countAlertChunks(chunks []db.SearchRagNearestNeighborsFilteredRow) int {
	return len(alertChunksOnly(chunks))
}
