package eval

import (
	"strings"
	"testing"
)

func TestSmokeNaturalFarmingFixtures_count(t *testing.T) {
	if got := len(SmokeNaturalFarmingFixtures()); got != 11 {
		t.Fatalf("expected 11 natural farming smoke fixtures, got %d", got)
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
	if len(FixturesForSuite("smoke-natural-farming")) != 11 {
		t.Fatal("smoke-natural-farming suite size")
	}
}

func TestSmokeNaturalFarmingFixtures_recipeOutcomes(t *testing.T) {
	for _, q := range SmokeNaturalFarmingFixtures() {
		if q.ID == "smoke-nf-recipe-outcomes" {
			if q.ExpectTool != "summarize_recipe_outcomes" {
				t.Fatalf("ExpectTool=%q", q.ExpectTool)
			}
			if !strings.Contains(strings.ToLower(q.Prompt), "based on history") {
				t.Fatalf("prompt=%q", q.Prompt)
			}
			return
		}
	}
	t.Fatal("smoke-nf-recipe-outcomes missing")
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

func TestScore_smokeNFRecipeOutcomes(t *testing.T) {
	res := Score(ScoreInput{
		Question: Question{ID: "smoke-nf-recipe-outcomes", Category: "natural_farming"},
		Answer:   "FFJ and WCA Flowering Boost on chrysanthemum: 2 harvested cycles — avg yield 396g, avg 0.22 USD/g. Correlational only.",
	})
	if !res.Passed {
		t.Fatalf("expected pass, got %+v", res)
	}
	bad := Score(ScoreInput{
		Question: Question{ID: "smoke-nf-recipe-outcomes", Category: "natural_farming"},
		Answer:   "This recipe will produce 400g next run for sure.",
	})
	if bad.Passed {
		t.Fatal("expected forecast claim to fail")
	}
}
