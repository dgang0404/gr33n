package farmguardian

import (
	"strings"
	"testing"

	db "gr33n-api/internal/db"
)

// Run #8 archive excerpt — numbered list, markdown links, no [n] markers.
const run8UnreadAlertsNoCites = `You have three unread high severity alert notifications on your farm:

1. [Humidity High in Flower Room](https://gr33ncore.sensor_alerts/unread): The humidity sensor triggered an alert with a reading of 72.4% RH.

2. [OHN Batch Below Minimum](https://gr33ncore.sensor_alerts/unread): The OHN batch is below the minimum threshold.

3. [Light Schedule Change Reminder](https://gr33ncore.sensor_alerts/unread): A low severity alert notifies you of an upcoming light schedule change.`

func run8AlertChunks() []db.SearchRagNearestNeighborsFilteredRow {
	return []db.SearchRagNearestNeighborsFilteredRow{
		{ID: 5, SourceType: SourceTypeAlertNotification, ContentText: "severity: high\nsubject: Humidity high — Flower Room"},
		{ID: 4, SourceType: SourceTypeAlertNotification, ContentText: "severity: medium\nsubject: OHN batch below minimum"},
		{ID: 3, SourceType: SourceTypeAlertNotification, ContentText: "severity: low\nsubject: Light schedule change"},
	}
}

func TestInjectAlertCitationRefs_run8Style(t *testing.T) {
	t.Parallel()
	got, ok := InjectAlertCitationRefs(run8UnreadAlertsNoCites, run8AlertChunks())
	if !ok {
		t.Fatal("expected injection")
	}
	if !strings.Contains(got, "[1]") || !strings.Contains(got, "[2]") || !strings.Contains(got, "[3]") {
		t.Fatalf("expected [1][2][3] injected: %s", got)
	}
}

func TestInjectAlertCitationRefs_skipsWhenRefsPresent(t *testing.T) {
	t.Parallel()
	answer := "1. Humidity high [1]\n2. OHN low [2]"
	got, ok := InjectAlertCitationRefs(answer, run8AlertChunks())
	if ok || got != answer {
		t.Fatalf("expected no-op, ok=%v got=%q", ok, got)
	}
}

func TestInjectAlertCitationRefs_singleAlertNoOp(t *testing.T) {
	t.Parallel()
	chunks := []db.SearchRagNearestNeighborsFilteredRow{
		{ID: 1, SourceType: SourceTypeAlertNotification},
	}
	_, ok := InjectAlertCitationRefs("1. One alert only", chunks)
	if ok {
		t.Fatal("expected no injection for single alert chunk")
	}
}

func TestMissingNumberedCitationsNote_run8Style(t *testing.T) {
	t.Parallel()
	note := MissingNumberedCitationsNote(run8UnreadAlertsNoCites)
	if note != "missing_numbered_citations" {
		t.Fatalf("note=%q", note)
	}
}

func TestMissingNumberedCitationsNote_withRefsPasses(t *testing.T) {
	t.Parallel()
	answer := "1. Humidity high [1]\n2. OHN low [2]"
	if note := MissingNumberedCitationsNote(answer); note != "" {
		t.Fatalf("note=%q", note)
	}
}
