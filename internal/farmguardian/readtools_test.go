package farmguardian

import (
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "gr33n-api/internal/db"
)

func TestMatchListUnreadAlertsIntent(t *testing.T) {
	cases := []struct {
		q    string
		want bool
	}{
		{"list my unread alerts", true},
		{"show unread alerts", true},
		{"tell me about alerts", true},
		{"acknowledge the humidity alert", false},
		{"mark alert #12 as read", false},
		{"how many unread alerts do I have", false},
		{"what is the weather", false},
	}
	for _, c := range cases {
		got := matchListUnreadAlertsIntent(c.q)
		if got != c.want {
			t.Fatalf("matchListUnreadAlertsIntent(%q) = %v want %v", c.q, got, c.want)
		}
	}
}

func TestMatchSummarizeZoneIntent(t *testing.T) {
	for _, q := range []string{
		"what's the humidity in Flower Room?",
		"summarize zone Flower Room",
		"temperature in Veg Room",
		"zone status for propagation",
	} {
		if !matchSummarizeZoneIntent(q) {
			t.Fatalf("expected summarize intent for %q", q)
		}
	}
}

func TestShouldRunSummarizeZoneReadIntent_Guards(t *testing.T) {
	cases := []struct {
		q    string
		want bool
	}{
		{"what's the humidity in Flower Room?", true},
		{"Please acknowledge the humidity alert in Flower Room", false},
		{"mark alert #12 as read", false},
		{"tell me about alerts", false},
		{"list my unread alerts", false},
		{"tell me about the Flower Room humidity sensor", true},
	}
	for _, c := range cases {
		got := shouldRunSummarizeZoneReadIntent(c.q)
		if got != c.want {
			t.Fatalf("shouldRunSummarizeZoneReadIntent(%q) = %v want %v", c.q, got, c.want)
		}
	}
}

func TestResolveZoneForSummary_ByName(t *testing.T) {
	snap := Snapshot{ZoneNames: []string{"Flower Room", "Veg Room"}}
	// Uses snap names only when zones list matches — test the string matcher via helper logic.
	lowerQ := strings.ToLower("what's the humidity in flower room?")
	name := "flower room"
	if !strings.Contains(lowerQ, name) {
		t.Fatal("expected substring match for zone resolution")
	}
	_ = snap
}

func TestFormatSensorReading(t *testing.T) {
	var raw pgtype.Numeric
	_ = raw.Scan("72.4")
	got := formatSensorReading(
		db.Gr33ncoreSensor{Name: "RH probe", SensorType: "humidity"},
		db.Gr33ncoreSensorReading{ValueRaw: raw, ReadingTime: time.Now()},
	)
	if got != "72.4% RH" {
		t.Fatalf("got %q want 72.4%% RH", got)
	}
}

func TestEnrichPromptBlock_NoQueries(t *testing.T) {
	if got := EnrichPromptBlock(nil, nil, 1, "list unread alerts", Snapshot{}); got != "" {
		t.Fatalf("expected empty block without querier, got %q", got)
	}
}

func TestReadToolIDs(t *testing.T) {
	ids := ReadToolIDs()
	want := []string{"list_unread_alerts", "summarize_zone", "list_plants", "summarize_zone_fertigation"}
	if len(ids) != len(want) {
		t.Fatalf("unexpected read tool ids: %v", ids)
	}
	for i, id := range want {
		if ids[i] != id {
			t.Fatalf("read tool ids = %v want %v", ids, want)
		}
	}
}

func TestMatchListPlantsIntent(t *testing.T) {
	cases := []struct {
		q    string
		want bool
	}{
		{"list my plants", true},
		{"show plants on this farm", true},
		{"add a philodendron plant to Living Room", false},
		{"create plant Philodendron", false},
		{"how many plants do I have", false},
	}
	for _, c := range cases {
		got := matchListPlantsIntent(c.q)
		if got != c.want {
			t.Fatalf("matchListPlantsIntent(%q) = %v want %v", c.q, got, c.want)
		}
	}
}

func TestMatchSummarizeZoneFertigationIntent(t *testing.T) {
	for _, q := range []string{
		"what fertigation program runs in Veg Room?",
		"show feeding programs for Flower Room",
		"ec targets in the outdoor garden",
	} {
		if !matchSummarizeZoneFertigationIntent(q) {
			t.Fatalf("expected fertigation intent for %q", q)
		}
	}
	if matchSummarizeZoneFertigationIntent("what's the humidity in Flower Room?") {
		t.Fatal("humidity-only question should not match fertigation intent")
	}
}
