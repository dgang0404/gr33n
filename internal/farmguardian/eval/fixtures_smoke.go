package eval

import "strings"

// SmokeFixtures returns the Phase 131 smoke suite plus Phase 211 WS5 cherry+JLF step (sequential).
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
			ID:         "smoke-morning-walk",
			Category:   "farm_state",
			Prompt:     MorningWalkPrompt(),
			Grounded:   true,
			Model:      "phi3:mini",
			ExpectTool: "walk_farm",
			ContextRef: MorningWalkContextRef(),
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
		{
			ID:         "smoke-cherry-jlf",
			Category:   "natural_farming",
			Prompt:     CherryGoldenrodJLFPrompt(),
			Grounded:   true,
			Model:      "phi3:mini",
			ExpectTool: "suggest_process_from_material",
		},
	}
}

// SmokeFullFixtures runs core smoke (5) then natural-farming smoke (10) in one sequential eval.
func SmokeFullFixtures() []Question {
	out := make([]Question, 0, len(SmokeFixtures())+len(SmokeNaturalFarmingFixtures()))
	out = append(out, SmokeFixtures()...)
	out = append(out, SmokeNaturalFarmingFixtures()...)
	return out
}

func FixturesForSuite(suite string) []Question {
	switch strings.ToLower(strings.TrimSpace(suite)) {
	case "smoke":
		return SmokeFixtures()
	case "smoke-full", "smoke_full":
		return SmokeFullFixtures()
	case "smoke-natural-farming", "smoke_natural_farming", "smoke-nf":
		return SmokeNaturalFarmingFixtures()
	case "phase127", "phase128", "p128":
		return Phase127Fixtures()
	case "change-requests", "change_requests", "proposals", "pr":
		return ChangeRequestFixtures()
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
