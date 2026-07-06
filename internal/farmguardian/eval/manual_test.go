package eval

import (
	"os"
	"strings"
	"testing"
)

func TestManualChecklist_smokeFixtureCount(t *testing.T) {
	fixtures := FixturesForSuite("smoke")
	if len(fixtures) != 4 {
		t.Fatalf("expected 4 smoke fixtures, got %d", len(fixtures))
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
		Answer:   "You have one unread seed alert about Veg Tent EC — review and acknowledge when ready.",
	})
	if !pass.Passed {
		t.Fatalf("expected pass: %s", pass.Notes)
	}
}
