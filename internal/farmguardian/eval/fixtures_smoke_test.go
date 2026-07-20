package eval

import (
	"strings"
	"testing"
)

func TestSmokeFixtures_morningWalkUsesMorningCheckEntry(t *testing.T) {
	fixtures := SmokeFixtures()
	var walk Question
	for _, q := range fixtures {
		if q.ID == "smoke-morning-walk" {
			walk = q
			break
		}
	}
	if walk.ContextRef == nil || walk.ContextRef.GuardianMode != "morning_walkthrough" {
		t.Fatal("smoke-morning-walk should use morning_walkthrough context_ref")
	}
	if !strings.Contains(walk.Prompt, "walk_farm") {
		t.Fatal("smoke-morning-walk prompt should mention walk_farm")
	}
}

func TestSmokeFixtures_count(t *testing.T) {
	if len(SmokeFixtures()) != 5 {
		t.Fatalf("expected 5 smoke fixtures, got %d", len(SmokeFixtures()))
	}
}

func TestFixturesForSuite_smoke(t *testing.T) {
	if len(FixturesForSuite("smoke")) != 5 {
		t.Fatal("smoke suite size")
	}
}

func TestSmokeFixtures_cherryJLFIsFifthStep(t *testing.T) {
	fixtures := SmokeFixtures()
	if fixtures[0].ID != "smoke-cherry-forest" {
		t.Fatalf("step 1 must stay smoke-cherry-forest, got %q", fixtures[0].ID)
	}
	last := fixtures[len(fixtures)-1]
	if last.ID != "smoke-cherry-jlf" {
		t.Fatalf("expected smoke-cherry-jlf as step 5, got %q", last.ID)
	}
	if !last.Grounded || last.ExpectTool != "suggest_process_from_material" {
		t.Fatalf("smoke-cherry-jlf fixture=%+v", last)
	}
}

func TestFixturesForSuite_regression(t *testing.T) {
	if len(FixturesForSuite("regression")) != len(Fixtures()) {
		t.Fatal("regression should match Fixtures()")
	}
}

func TestScore_smokeCherryJLF(t *testing.T) {
	res := Score(ScoreInput{
		Question: Question{ID: "smoke-cherry-jlf", Category: "natural_farming"},
		Answer: `Goldenrod fits extension-method JLF on this farm — use fermented plant juice from the process catalog at 1:100 to start on the cherry understory.`,
	})
	if !res.Passed {
		t.Fatalf("expected pass, got %+v", res)
	}
}

func TestScore_smokeCherryForest(t *testing.T) {
	res := Score(ScoreInput{
		Question: Question{ID: "smoke-cherry-forest", Category: "ungrounded"},
		Answer:   "Your cherry tree understory can keep blackberries if you manage goldenrod for dyes separately.",
	})
	if !res.Passed {
		t.Fatalf("expected pass, got %+v", res)
	}
}
