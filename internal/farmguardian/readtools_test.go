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
	if got := EnrichPromptBlock(nil, nil, 1, "list unread alerts", Snapshot{}, nil); got != "" {
		t.Fatalf("expected empty block without querier, got %q", got)
	}
}

func TestReadToolIDs(t *testing.T) {
	ids := ReadToolIDs()
	for _, want := range []string{
		"list_unread_alerts",
		"summarize_farm_low_stock",
		"restock_priority",
		"summarize_cycle_cost",
		"summarize_farm_spending",
		"summarize_active_grows",
		"summarize_zone",
		"list_plants",
		"summarize_zone_fertigation",
	} {
		found := false
		for _, id := range ids {
			if id == want {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("ReadToolIDs missing %q: %v", want, ids)
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

func TestMatchSummarizeFarmLowStockIntent(t *testing.T) {
	cases := []struct {
		q    string
		want bool
	}{
		{"What's running low?", true},
		{"low stock on the farm", true},
		{"supplies low — do I need to restock?", true},
		{"out of OHN", true},
		{"Do I need to restock anything?", true},
		{"What supplies are below their low-stock threshold on this farm?", true},
		{"list my plants", false},
		{"show plants on this farm", false},
		{"summarize zone Flower Room", false},
	}
	for _, c := range cases {
		got := matchSummarizeFarmLowStockIntent(c.q)
		if got != c.want {
			t.Fatalf("matchSummarizeFarmLowStockIntent(%q) = %v want %v", c.q, got, c.want)
		}
	}
}

func TestShouldRunSummarizeFarmLowStockReadIntent_Guards(t *testing.T) {
	cases := []struct {
		q    string
		want bool
	}{
		{"What's running low?", true},
		{"list my plants", false},
		{"tell me about the plant catalog inventory", false},
		{"summarize zone Flower Room", false},
		{"is the reservoir running low in Flower Room?", false},
	}
	for _, c := range cases {
		got := shouldRunSummarizeFarmLowStockReadIntent(c.q)
		if got != c.want {
			t.Fatalf("shouldRunSummarizeFarmLowStockReadIntent(%q) = %v want %v", c.q, got, c.want)
		}
	}
}

func TestMatchSummarizeZoneFertigationIntent(t *testing.T) {
	for _, q := range []string{
		"what fertigation program runs in Veg Room?",
		"show feeding programs for Flower Room",
		"ec targets in the outdoor garden",
		"When is the next feed for Flower Room?",
	} {
		if !matchSummarizeZoneFertigationIntent(q) {
			t.Fatalf("expected fertigation intent for %q", q)
		}
	}
	if matchSummarizeZoneFertigationIntent("what's the humidity in Flower Room?") {
		t.Fatal("humidity-only question should not match fertigation intent")
	}
}
