package eval

import "strings"

// SmokeFixtures returns the Phase 131 four-step smoke suite (sequential, one at a time).
func SmokeFixtures() []Question {
	return []Question{
		{
			ID:       "smoke-cherry-forest",
			Category: "ungrounded",
			Prompt: "I have a cherry tree with plants growing under it — I want this forest garden situation but I think the Canadian goldenrod is not good; I'll use it for dyes but maybe I need to get rid of it now. The blackberries would be nice if they could stay; they have thorns.",
			Grounded: false,
			Model:    "phi3:mini",
		},
		{
			ID:       "smoke-morning-walk",
			Category: "farm_state",
			Prompt:   "What should I check first on a morning walkthrough of this farm today?",
			Grounded: true,
			Model:    "phi3:mini",
			ExpectTool: "walk_farm",
		},
		{
			ID:       "smoke-unread-alerts",
			Category: "farm_state",
			Prompt:   "Summarize my unread alerts and what I should do about each one.",
			Grounded: true,
			Model:    "phi3:mini",
		},
		{
			ID:       "smoke-ec-ph",
			Category: "field_guide",
			Prompt:   "What does our operational documentation say about EC and pH targets for leafy greens here?",
			ExpectCitation: true,
			Grounded: true,
			Model:    "phi3:mini",
		},
	}
}

// FixturesForSuite returns prompts for smoke, phase127, regression (default), or all.
func FixturesForSuite(suite string) []Question {
	switch strings.ToLower(strings.TrimSpace(suite)) {
	case "smoke":
		return SmokeFixtures()
	case "phase127", "phase128", "p128":
		return Phase127Fixtures()
	case "all":
		out := make([]Question, 0, len(Fixtures())+len(SmokeFixtures())+len(Phase127Fixtures()))
		out = append(out, SmokeFixtures()...)
		out = append(out, Phase127Fixtures()...)
		out = append(out, Fixtures()...)
		return out
	default:
		return Fixtures()
	}
}
