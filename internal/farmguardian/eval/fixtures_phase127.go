package eval

// Phase127Fixtures returns the Phase 128 four-prompt validation suite (Phase 127 grounding).
// Wording matches docs/plans/phase_128_validate_phase127_guardian.plan.md for manual UI parity.
func Phase127Fixtures() []Question {
	return []Question{
		{
			ID:         "p128-devices",
			Category:   "farm_state",
			Prompt:     "Are any edge devices offline?",
			Grounded:   true,
			Model:      "phi3:mini",
			ExpectTool: "summarize_device_health",
		},
		{
			ID:       "p128-fert-manual",
			Category: "farm_state",
			Prompt:   "Which fertigation programs are manual-only?",
			Grounded: true,
			Model:    "phi3:mini",
		},
		{
			ID:             "p128-demo-pi",
			Category:       "field_guide",
			Prompt:         "Which relay channel is the veg grow light on the demo farm?",
			ExpectCitation: true,
			Grounded:       true,
			Model:          "phi3:mini",
		},
		{
			ID:             "p128-fert-triage",
			Category:       "field_guide",
			Prompt:         "Program active but no dose — what to check first?",
			ExpectCitation: true,
			Grounded:       true,
			Model:          "phi3:mini",
		},
	}
}
