// Phase 128 — validate Phase 127 grounding suite wiring.

package main

import (
	"testing"

	"gr33n-api/internal/farmguardian/eval"
)

func TestPhase128_Phase127Suite(t *testing.T) {
	fixtures := eval.Phase127Fixtures()
	if len(fixtures) != 4 {
		t.Fatalf("expected 4 fixtures, got %d", len(fixtures))
	}
	ids := map[string]bool{}
	for _, q := range fixtures {
		ids[q.ID] = true
		if !q.Grounded {
			t.Fatalf("%s must be grounded", q.ID)
		}
		if q.Model != "phi3:mini" {
			t.Fatalf("%s model", q.ID)
		}
	}
	for _, want := range []string{"p128-devices", "p128-fert-manual", "p128-demo-pi", "p128-fert-triage"} {
		if !ids[want] {
			t.Fatalf("missing %s", want)
		}
	}
}

func TestPhase128_ScoreHeuristics(t *testing.T) {
	pass := eval.Score(eval.ScoreInput{
		Question: eval.Question{ID: "p128-demo-pi", ExpectCitation: true},
		Answer:   "Per [1] demo-farm-pi-layout, the veg grow light is on relay_1 on the Veg Relay Controller.",
		CitationCount: 1,
	})
	if !pass.Passed {
		t.Fatalf("demo pi score: %s", pass.Notes)
	}
}
