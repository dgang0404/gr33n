package eval

// SmokeNaturalFarmingFixtures — Phase 211 follow-on: grounded natural-farming counsel (~10 prompts).
// Requires farm context ON, field guides ingested, and demo natural-farming rows on the farm.
func SmokeNaturalFarmingFixtures() []Question {
	return []Question{
		{
			ID:             "smoke-nf-jlf-doc",
			Category:       "natural_farming",
			Prompt:         `I'm reading "JLF from weeds and grass (general)" from our indexed knowledge. What should I know from this doc for my farm right now?`,
			Grounded:       true,
			ExpectCitation: true,
			Model:          "phi3:mini",
		},
		{
			ID:       "smoke-nf-jms-dilution",
			Category: "natural_farming",
			Prompt:   "For JMS on this farm — what dilution do we use for a soil drench versus a foliar spray?",
			Grounded: true,
			Model:    "phi3:mini",
		},
		{
			ID:         "smoke-nf-ready-batches",
			Category:   "natural_farming",
			Prompt:     "What natural farming ferments or input batches do I have ready on hand right now?",
			Grounded:   true,
			ExpectTool: "summarize_natural_farming_inventory",
			Model:      "phi3:mini",
		},
		{
			ID:       "smoke-nf-jms-make",
			Category: "natural_farming",
			Prompt:   "Walk me through making JMS — ingredients and timing — using our field guides.",
			Grounded: true,
			Model:    "phi3:mini",
		},
		{
			ID:       "smoke-nf-jlf-start",
			Category: "natural_farming",
			Prompt:   "I'm new to JLF on this farm. What dilution should I start at before going stronger?",
			Grounded: true,
			Model:    "phi3:mini",
		},
		{
			ID:       "smoke-nf-combined-drench",
			Category: "natural_farming",
			Prompt:   "How should I combine JLF and JMS in the same drench tank on this farm?",
			Grounded: true,
			Model:    "phi3:mini",
		},
		{
			ID:       "smoke-nf-ffj-flower",
			Category: "natural_farming",
			Prompt:   "When would I reach for FFJ instead of JLF here, especially in flower?",
			Grounded: true,
			Model:    "phi3:mini",
		},
		{
			ID:       "smoke-nf-wca-foliar",
			Category: "natural_farming",
			Prompt:   "How do we dilute WCA for a foliar spray according to our natural farming docs?",
			Grounded: true,
			Model:    "phi3:mini",
		},
		{
			ID:         "smoke-nf-goldenrod",
			Category:   "natural_farming",
			Prompt:     CherryGoldenrodJLFPrompt(),
			Grounded:   true,
			ExpectTool: "suggest_process_from_material",
			Model:      "phi3:mini",
		},
		{
			ID:       "smoke-nf-lab",
			Category: "natural_farming",
			Prompt:   "What is LAB in our natural farming setup and when would I use it on soil?",
			Grounded: true,
			Model:    "phi3:mini",
		},
		{
			ID:         "smoke-nf-recipe-outcomes",
			Category:   "natural_farming",
			Prompt:     "Which recipe worked best for chrysanthemum based on history — compare yield and cost per gram from our harvested cycles.",
			Grounded:   true,
			ExpectTool: "summarize_recipe_outcomes",
			Model:      "phi3:mini",
		},
	}
}
