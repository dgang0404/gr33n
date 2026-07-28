package farmguardian

import (
	"strings"
	"testing"

	db "gr33n-api/internal/db"
)

func TestMergeLiveUnreadAlertChunks_seedsWhenRAGEmpty(t *testing.T) {
	t.Parallel()
	high := db.Gr33ncoreNotificationPriorityEnumHigh
	med := db.Gr33ncoreNotificationPriorityEnumMedium
	subj1, subj2 := "Humidity high — Flower Room", "OHN batch below minimum"
	alerts := []db.ListRecentUnreadAlertsByFarmRow{
		{ID: 10, Severity: &high, SubjectRendered: &subj1},
		{ID: 11, Severity: &med, SubjectRendered: &subj2},
	}
	out := MergeLiveUnreadAlertChunks(nil, alerts)
	if countAlertChunks(out) != 2 {
		t.Fatalf("alert chunks=%d want 2", countAlertChunks(out))
	}
	if out[0].SourceID != 10 || out[0].SourceType != SourceTypeAlertNotification {
		t.Fatalf("first chunk: %+v", out[0])
	}
	if !strings.Contains(out[0].ContentText, "Humidity high") {
		t.Fatalf("content=%q", out[0].ContentText)
	}
	answer := "1. High humidity in Flower Room\n2. OHN batch low"
	got, ok := InjectAlertCitationRefs(answer, out)
	if !ok || !strings.Contains(got, "[1]") || !strings.Contains(got, "[2]") {
		t.Fatalf("inject ok=%v got=%q", ok, got)
	}
}

func TestMergeLiveUnreadAlertChunks_dedupesExisting(t *testing.T) {
	t.Parallel()
	existing := []db.SearchRagNearestNeighborsFilteredRow{
		{ID: 10, SourceType: SourceTypeAlertNotification, SourceID: 10, ContentText: "already"},
	}
	high := db.Gr33ncoreNotificationPriorityEnumHigh
	subj := "New alert"
	alerts := []db.ListRecentUnreadAlertsByFarmRow{
		{ID: 10, Severity: &high, SubjectRendered: &subj},
		{ID: 12, Severity: &high, SubjectRendered: &subj},
	}
	out := MergeLiveUnreadAlertChunks(existing, alerts)
	if countAlertChunks(out) != 2 {
		t.Fatalf("want 2 after dedupe, got %d", countAlertChunks(out))
	}
}
