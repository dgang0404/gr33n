package eval

import (
	"strings"
	"testing"
)

func TestSmokeNaturalFarmingFixtures_count(t *testing.T) {
	if got := len(SmokeNaturalFarmingFixtures()); got != 10 {
		t.Fatalf("expected 10 natural farming smoke fixtures, got %d", got)
	}
}

func TestSmokeNaturalFarmingFixtures_jlfDocPrompt(t *testing.T) {
	for _, q := range SmokeNaturalFarmingFixtures() {
		if q.ID == "smoke-nf-jlf-doc" {
			if !strings.Contains(q.Prompt, "JLF from weeds and grass") {
				t.Fatalf("prompt=%q", q.Prompt)
			}
			if !q.ExpectCitation {
				t.Fatal("expected citation expectation")
			}
			return
		}
	}
	t.Fatal("smoke-nf-jlf-doc missing")
}

func TestFixturesForSuite_smokeNaturalFarming(t *testing.T) {
	if len(FixturesForSuite("smoke-natural-farming")) != 10 {
		t.Fatal("smoke-natural-farming suite size")
	}
}

func TestScore_smokeNFJlfDoc(t *testing.T) {
	res := Score(ScoreInput{
		Question:      Question{ID: "smoke-nf-jlf-doc", Category: "natural_farming"},
		Answer:        "Start JLF at 1:100 if unsure [1]. Ferment weeds 2/3 full 7–14 days; strain before drench.",
		CitationCount: 1,
	})
	if !res.Passed {
		t.Fatalf("expected pass, got %+v", res)
	}
}

func TestScore_smokeNFJmsDilution(t *testing.T) {
	res := Score(ScoreInput{
		Question: Question{ID: "smoke-nf-jms-dilution", Category: "natural_farming"},
		Answer:   "Soil drench JMS 1:10; foliar spray 1:20 with JWA for leaf coverage.",
	})
	if !res.Passed {
		t.Fatalf("expected pass, got %+v", res)
	}
}
