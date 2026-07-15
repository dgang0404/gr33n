package eval

import "strings"

// ChangeRequestUIScenarios returns multi-turn Guardian dialogues for Pending-tab UI
// testing: some scenarios refine back-and-forth then confirm via API; others end
// left pending so operators can exercise Confirm / Refine / Dismiss in the UI.
func ChangeRequestUIScenarios() []Scenario {
	return []Scenario{
		{
			ID:               "scenario-feed-revise-confirm",
			Category:         "write_intent",
			Grounded:         true,
			ExpectProposal:   true,
			VerifyFixtureID:  "write-feed",
			ConfirmFinal:     true,
			WantVolumeLiters: 0.3,
			MinRevision:      2,
			Turns: []ScenarioTurn{
				{Prompt: "Set the feed volume to 0.5 liters for the Veg Tent program."},
				{Prompt: "Please revise — use 0.3 L instead of 0.5."},
			},
		},
		{
			ID:               "scenario-feed-revise-pending",
			Category:         "write_intent",
			Grounded:         true,
			ExpectProposal:   true,
			LeavePending:     true,
			MinRevision:      2,
			WantVolumeLiters: 0.3,
			Turns: []ScenarioTurn{
				{Prompt: "Set the feed volume to 0.5 liters for the Veg Tent program."},
				{Prompt: "Please revise — use 0.3 L instead of 0.5."},
			},
		},
		{
			ID:             "scenario-task-dialogue-pending",
			Category:       "write_intent",
			Grounded:       true,
			ExpectProposal: true,
			LeavePending:   true,
			MinRevision:      4,
			RequireTaskZone:  true,
			WantTitle:        "Refill calcium nitrate",
			WantDueDate:      "2026-07-20",
			Turns: []ScenarioTurn{
				{Prompt: "Create a task to refill calcium nitrate when stock is low."},
				{Prompt: "Put it in Veg Room — that is the zone for this task."},
				{Prompt: "call it Refill calcium nitrate instead"},
				{Prompt: "set the due date to 2026-07-20"},
			},
		},
		{
			ID:             "scenario-schedule-pending",
			Category:       "write_intent",
			Grounded:       true,
			ExpectProposal: true,
			LeavePending:   true,
			Turns: []ScenarioTurn{
				{Prompt: "Pause the lights schedule for Veg Tent until tomorrow."},
			},
		},
		{
			ID:             "scenario-ack-pending",
			Category:       "write_intent",
			Grounded:       true,
			ExpectProposal: true,
			LeavePending:   true,
			Turns: []ScenarioTurn{
				{Prompt: "Acknowledge the highest severity unread alert."},
			},
		},
	}
}

// ChangeRequestUIScenariosQuick is a shorter mix for faster Pending-tab demos (~2 single-turn scenarios).
func ChangeRequestUIScenariosQuick() []Scenario {
	all := ChangeRequestUIScenarios()
	want := map[string]bool{
		"scenario-ack-pending":       true,
		"scenario-schedule-pending":  true,
	}
	out := make([]Scenario, 0, len(want))
	for _, sc := range all {
		if want[sc.ID] {
			out = append(out, sc)
		}
	}
	return out
}

// ScenariosForSuite returns multi-turn scenarios for UI smoke suites.
func ScenariosForSuite(suite string) []Scenario {
	switch strings.ToLower(strings.TrimSpace(suite)) {
	case "change-requests-ui", "change_requests_ui", "pr-ui", "pr_ui":
		return ChangeRequestUIScenarios()
	case "change-requests-ui-quick", "change_requests_ui_quick", "pr-ui-quick":
		return ChangeRequestUIScenariosQuick()
	default:
		return nil
	}
}
