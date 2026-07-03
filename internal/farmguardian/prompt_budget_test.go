package farmguardian

import "testing"

func TestComputePromptBudget_largeWindowUnchanged(t *testing.T) {
	budget, log := ComputePromptBudget(32768, 20)
	if budget.RAGTopK != RAGTopK || budget.MaxHistoryTurns != 20 {
		t.Fatalf("expected full budget, got %+v", budget)
	}
	if len(log) != 0 {
		t.Fatalf("expected no trim log, got %v", log)
	}
}

func TestComputePromptBudget_smallWindowShrinks(t *testing.T) {
	budget, log := ComputePromptBudget(2048, 20)
	if budget.RAGTopK != 3 {
		t.Fatalf("RAGTopK: got %d want 3", budget.RAGTopK)
	}
	if budget.MaxHistoryTurns > 4 {
		t.Fatalf("history: got %d want <=4", budget.MaxHistoryTurns)
	}
	if budget.Snapshot.MaxZones > 4 {
		t.Fatalf("snapshot zones: got %d", budget.Snapshot.MaxZones)
	}
	if len(log) == 0 {
		t.Fatal("expected trim log entries")
	}
	est := EstimatePromptTokens("x")
	if est < 1 {
		t.Fatal("estimate should be positive")
	}
}

func TestApplyBudgetLimits_truncatesSnapshot(t *testing.T) {
	s := Snapshot{
		ZoneNames: []string{"A", "B", "C", "D", "E"},
		UnreadAlertDetails: []UnreadAlertDetail{
			{Subject: "1"}, {Subject: "2"}, {Subject: "3"},
		},
	}
	s.ApplyBudgetLimits(SnapshotBudgetLimits{MaxZones: 2, MaxAlertDetails: 1})
	if len(s.ZoneNames) != 2 || len(s.UnreadAlertDetails) != 1 {
		t.Fatalf("truncate failed: zones=%d alerts=%d", len(s.ZoneNames), len(s.UnreadAlertDetails))
	}
}
