package farmguardian

import (
	"strings"
	"testing"

	"gr33n-api/internal/cropcycle/recipeoutcomes"
)

func TestShouldRunSummarizeRecipeOutcomesReadIntent(t *testing.T) {
	t.Parallel()
	cases := []struct {
		q    string
		want bool
	}{
		{"which recipe worked best for tomato?", true},
		{"predict my yield based on history", true},
		{"what is the weather today?", false},
	}
	for _, tc := range cases {
		if got := shouldRunSummarizeRecipeOutcomesReadIntent(tc.q, nil); got != tc.want {
			t.Fatalf("q=%q got=%v want=%v", tc.q, got, tc.want)
		}
	}
}

func TestShouldRunSummarizeRecipeOutcomesReadIntent_historyCompare(t *testing.T) {
	t.Parallel()
	q := "Using only our farm data, compare chrysanthemum recipe outcome history (track record: avg yield and cost per gram by application recipe) with our last harvested chrysanthemum cycles in crop analytics."
	if !shouldRunSummarizeRecipeOutcomesReadIntent(q, nil) {
		t.Fatal("expected recipe outcomes intent for history-compare prompt")
	}
}

func TestFormatRecipeOutcomeLine_includesAvgAndN(t *testing.T) {
	t.Parallel()
	avgY := 182.0
	minY := 140.0
	maxY := 210.0
	avgC := 0.21
	rev := int64(3)
	row := recipeoutcomes.RecipeOutcome{
		RecipeName:                  "JMS Foliar",
		CropKey:                     "tomato",
		CycleCount:                  4,
		AvgYieldGrams:               &avgY,
		MinYieldGrams:               &minY,
		MaxYieldGrams:               &maxY,
		AvgCostPerGram:              &avgC,
		CostCurrency:                "USD",
		ApplicationRecipeRevisionID: &rev,
	}
	line := formatRecipeOutcomeLine(row, true)
	for _, part := range []string{"4 harvested cycles", "avg yield 182g", "range 140–210g", "avg 0.21 USD/g", "rev #3"} {
		if !strings.Contains(line, part) {
			t.Fatalf("line missing %q: %s", part, line)
		}
	}
}

func TestRecipeOutcomeToolGroundingNote_formattedLinePasses(t *testing.T) {
	t.Parallel()
	avgY := 182.0
	minY := 140.0
	maxY := 210.0
	rev := int64(3)
	row := recipeoutcomes.RecipeOutcome{
		RecipeName:                  "JMS Foliar",
		CropKey:                     "tomato",
		CycleCount:                  4,
		AvgYieldGrams:               &avgY,
		MinYieldGrams:               &minY,
		MaxYieldGrams:               &maxY,
		ApplicationRecipeRevisionID: &rev,
	}
	block := "summarize_recipe_outcomes — Demo Farm\n" + formatRecipeOutcomeLine(row, false)
	if note := RecipeOutcomeToolGroundingNote(block); note != "" {
		t.Fatalf("unexpected note %q", note)
	}
}

func TestRecipeOutcomeToolGroundingNote_bareNumberFails(t *testing.T) {
	t.Parallel()
	block := "summarize_recipe_outcomes — Demo Farm\nJMS Foliar (tomato): yield 182g last run"
	if note := RecipeOutcomeToolGroundingNote(block); note == "" {
		t.Fatal("expected recipe_outcome_bare_number note")
	}
}

func TestRecipeOutcomeGroundingRule_steersSmallModels(t *testing.T) {
	t.Parallel()
	for _, part := range []string{
		"Only name recipes that appear",
		"Never invent a second recipe",
		"if we assume",
		"averaged",
	} {
		if !strings.Contains(RecipeOutcomeGroundingRule, part) {
			t.Fatalf("RecipeOutcomeGroundingRule missing %q", part)
		}
	}
	if !strings.Contains(recipeOutcomeToolFooter, "Grow/cycle names are not recipes") {
		t.Fatal("tool footer missing cycle-name warning")
	}
}

func TestReadToolIDsIncludesSummarizeRecipeOutcomes(t *testing.T) {
	t.Parallel()
	ids := ReadToolIDs()
	found := false
	for _, id := range ids {
		if id == "summarize_recipe_outcomes" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("summarize_recipe_outcomes missing from ReadToolIDs")
	}
}
