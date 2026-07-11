package farmguardian

import (
	"testing"

	db "gr33n-api/internal/db"
)

func TestMatchAlertSummaryIntent_smokeFixture(t *testing.T) {
	t.Parallel()
	q := "Summarize my unread alerts and what I should do about each one."
	if !MatchAlertSummaryIntent(q) {
		t.Fatal("expected smoke-unread-alerts prompt to match")
	}
}

func TestMatchAlertSummaryIntent_listAlerts(t *testing.T) {
	t.Parallel()
	if !MatchAlertSummaryIntent("list my unread alerts") {
		t.Fatal("expected list intent")
	}
	if MatchAlertSummaryIntent("what is EC for lettuce") {
		t.Fatal("expected false for unrelated question")
	}
}

func TestFilterChunksForAlertSummary_alertOnlyWhenTwoPlus(t *testing.T) {
	t.Parallel()
	chunks := []db.SearchRagNearestNeighborsFilteredRow{
		{ID: 1, SourceType: "platform_doc", ContentText: "severity: n/a\nworkflow"},
		{ID: 2, SourceType: "alert_notification", ContentText: "severity: high\nsubject: Humidity high"},
		{ID: 3, SourceType: "alert_notification", ContentText: "severity: medium\nsubject: OHN low"},
		{ID: 4, SourceType: "alert_notification", ContentText: "severity: low\nsubject: Light schedule"},
	}
	q := "Summarize my unread alerts and what I should do about each one."
	out := FilterChunksForAlertSummary(q, chunks)
	if len(out) != 3 {
		t.Fatalf("expected 3 alert-only chunks, got %d", len(out))
	}
	for _, c := range out {
		if c.SourceType != SourceTypeAlertNotification {
			t.Fatalf("expected alert only, got %s", c.SourceType)
		}
	}
}

func TestFilterChunksForAlertSummary_singleAlertUnchanged(t *testing.T) {
	t.Parallel()
	chunks := []db.SearchRagNearestNeighborsFilteredRow{
		{ID: 1, SourceType: "platform_doc"},
		{ID: 2, SourceType: "alert_notification"},
	}
	q := "Summarize my unread alerts"
	out := FilterChunksForAlertSummary(q, chunks)
	if len(out) != 2 {
		t.Fatalf("expected unchanged 2 chunks, got %d", len(out))
	}
}

func TestFilterChunksForAlertSummary_nonAlertQuestionUnchanged(t *testing.T) {
	t.Parallel()
	chunks := []db.SearchRagNearestNeighborsFilteredRow{
		{ID: 1, SourceType: "field_guide"},
		{ID: 2, SourceType: "alert_notification"},
		{ID: 3, SourceType: "alert_notification"},
	}
	out := FilterChunksForAlertSummary("EC targets for lettuce", chunks)
	if len(out) != 3 {
		t.Fatalf("expected unchanged, got %d", len(out))
	}
}
