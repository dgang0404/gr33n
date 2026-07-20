package eval

// RegressionFixtures returns additive grounded prompts for make guardian-qa-regression.
// Kept separate from SmokeFixtures — do not add smoke-cherry-forest here.
func RegressionFixtures() []Question {
	return []Question{
		{
			ID:         "regression-cherry-goldenrod-jlf",
			Category:   "natural_farming",
			Prompt:     CherryGoldenrodJLFPrompt(),
			Grounded:   true,
			Model:      "phi3:mini",
			ExpectTool: "suggest_process_from_material",
		},
	}
}

// CherryGoldenrodJLFPrompt is the grounded forest-garden + JLF variant (Phase 210 WS5).
func CherryGoldenrodJLFPrompt() string {
	return "I have a cherry tree with goldenrod and blackberries in the understory on this farm. Instead of removing the goldenrod, can I ferment it into JLF for the cherry? What dilution should I start with?"
}
