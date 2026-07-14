package eval

import "testing"

func TestScoreScenario_multiTurnPassesWithPendingProposal(t *testing.T) {
	sc := Scenario{
		ID:             "scenario-task-dialogue-pending",
		ExpectProposal: true,
		Turns:          []ScenarioTurn{{Prompt: "create task"}, {Prompt: "which zone?"}},
	}
	in := ScoreInput{
		Question:      scenarioScoreQuestion(sc, "which zone?"),
		Answer:        "completely off-topic gibberish about literature",
		ProposalCount: 1,
		ProposalIDs:   []string{"p-1"},
	}
	res := scoreScenario(in, sc)
	if !res.Passed {
		t.Fatalf("want pass when pending proposal enriched, got %q", res.Notes)
	}
}

func TestScoreScenario_singleTurnStillRequiresProposal(t *testing.T) {
	sc := Scenario{
		ID:             "scenario-ack-pending",
		ExpectProposal: true,
		Turns:          []ScenarioTurn{{Prompt: "ack alert"}},
	}
	in := ScoreInput{
		Question:      scenarioScoreQuestion(sc, "ack alert"),
		Answer:        "no proposal here",
		ProposalCount: 0,
	}
	res := scoreScenario(in, sc)
	if res.Passed {
		t.Fatal("single-turn should fail without proposal")
	}
}
