package eval

import "testing"

func TestRegressionFixtures_CherryGoldenrodJLF(t *testing.T) {
	fixtures := RegressionFixtures()
	if len(fixtures) != 1 {
		t.Fatalf("len=%d want 1", len(fixtures))
	}
	q := fixtures[0]
	if q.ID != "regression-cherry-goldenrod-jlf" {
		t.Fatalf("id=%q", q.ID)
	}
	if !q.Grounded || q.ExpectTool != "suggest_process_from_material" {
		t.Fatalf("fixture=%+v", q)
	}
	last := Fixtures()[len(Fixtures())-1]
	if last.ID != q.ID {
		t.Fatalf("Fixtures() tail=%q want %q", last.ID, q.ID)
	}
}

func TestScore_regressionCherryGoldenrodJLF_pass(t *testing.T) {
	res := Score(ScoreInput{
		Question: Question{ID: "regression-cherry-goldenrod-jlf", Category: "natural_farming"},
		Answer: `Goldenrod biomass fits the extension-method JLF approach on this farm — not a Cho-named goldenrod recipe.
Use fermented plant juice from the process catalog at 1:100 to start on the cherry understory; stronger 1:30 only after the tree looks hungry.
I can draft an application recipe if you want to Confirm it in Natural farming.`,
	})
	if !res.Passed {
		t.Fatalf("expected pass, got %+v", res)
	}
}

func TestScore_regressionCherryGoldenrodJLF_failMissingJLF(t *testing.T) {
	res := Score(ScoreInput{
		Question: Question{ID: "regression-cherry-goldenrod-jlf", Category: "natural_farming"},
		Answer:   stringsRepeat("Remove goldenrod now. ", 20) + "Use 1:100 water on cherry.",
	})
	if res.Passed {
		t.Fatalf("expected fail without JLF framing")
	}
}

func TestScore_regressionCherryGoldenrodJLF_failChoRecipe(t *testing.T) {
	res := Score(ScoreInput{
		Question: Question{ID: "regression-cherry-goldenrod-jlf", Category: "natural_farming"},
		Answer:   "Use Cho's goldenrod recipe for JLF at 1:100 on your cherry tree understory with fermented plant juice.",
	})
	if res.Passed {
		t.Fatalf("expected fail on Cho-named recipe")
	}
}

func TestScore_smokeCherryForest_unchanged(t *testing.T) {
	res := Score(ScoreInput{
		Question: Question{ID: "smoke-cherry-forest", Category: "ungrounded"},
		Answer:   stringsRepeat("Cherry and goldenrod understory with blackberry canes need thoughtful management. ", 3),
	})
	if !res.Passed {
		t.Fatalf("smoke bar should remain unchanged: %+v", res)
	}
}

func stringsRepeat(s string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}
