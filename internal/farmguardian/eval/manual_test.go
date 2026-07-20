package eval

import (
	"os"
	"strings"
	"testing"
)

func TestManualChecklist_smokeFixtureCount(t *testing.T) {
	fixtures := FixturesForSuite("smoke")
	if len(fixtures) != 5 {
		t.Fatalf("expected 5 smoke fixtures, got %d", len(fixtures))
	}
	if fixtures[0].ID != "smoke-cherry-forest" {
		t.Fatalf("first fixture: %s", fixtures[0].ID)
	}
	if fixtures[0].Grounded {
		t.Fatal("cherry smoke step should be ungrounded")
	}
	if fixtures[1].ID != "smoke-morning-walk" || !fixtures[1].Grounded {
		t.Fatalf("second fixture: %+v", fixtures[1])
	}
	if fixtures[4].ID != "smoke-cherry-jlf" || !fixtures[4].Grounded {
		t.Fatalf("fifth fixture: %+v", fixtures[4])
	}
}

func TestManualPassHint_morningWalk(t *testing.T) {
	hint := manualPassHint(Question{ID: "smoke-morning-walk", ExpectTool: "walk_farm"})
	if !strings.Contains(hint, "alert") && !strings.Contains(hint, "walk_farm") {
		t.Fatalf("hint: %s", hint)
	}
}

func TestScrapeLogEvidence_tool(t *testing.T) {
	dir := t.TempDir()
	logPath := dir + "/api.log"
	content := "2026/07/06 INFO guardian: tool_id=walk_farm farm_id=1\n" +
		"2026/07/06 INFO smoke-morning-walk request done\n"
	if err := os.WriteFile(logPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	ev := ScrapeLogEvidence(logPath, "smoke-morning-walk", "walk_farm")
	if len(ev) == 0 {
		t.Fatal("expected log evidence")
	}
	found := false
	for _, e := range ev {
		if strings.Contains(e, "walk_farm") {
			found = true
		}
	}
	if !found {
		t.Fatalf("evidence: %v", ev)
	}
}

func TestScore_smokeUnreadAlerts(t *testing.T) {
	pass := Score(ScoreInput{
		Question: Question{ID: "smoke-unread-alerts"},
		Answer:   "1. Veg Tent EC alert [1] — review and acknowledge.\n2. Humidity high [2] — check dehumidifier.",
		CitationCount: 2,
	})
	if !pass.Passed {
		t.Fatalf("expected pass: %s", pass.Notes)
	}
}

func TestScore_smokeUnreadAlerts_run8UncitedFails(t *testing.T) {
	t.Parallel()
	res := Score(ScoreInput{
		Question: Question{ID: "smoke-unread-alerts", Category: "farm_state"},
		Answer: `You have three unread alerts on your farm:
1. Humidity high in Flower Room — 72.4% RH.
2. OHN batch below minimum — 0.35 L remaining.
3. Light schedule change in 48 hours.`,
		CitationCount: 0,
	})
	if res.Passed {
		t.Fatalf("run #8-style uncited list should fail: %+v", res)
	}
}
