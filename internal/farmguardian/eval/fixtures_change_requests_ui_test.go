package eval

import "testing"

func TestChangeRequestUIScenarios_count(t *testing.T) {
	all := ChangeRequestUIScenarios()
	if len(all) != 5 {
		t.Fatalf("want 5 UI scenarios, got %d", len(all))
	}
	var confirm, pending int
	for _, sc := range all {
		if len(sc.Turns) == 0 {
			t.Fatalf("%s: no turns", sc.ID)
		}
		if sc.ConfirmFinal && sc.LeavePending {
			t.Fatalf("%s: cannot confirm and leave pending", sc.ID)
		}
		if sc.ConfirmFinal {
			confirm++
		}
		if sc.LeavePending {
			pending++
		}
	}
	if confirm != 1 {
		t.Fatalf("want 1 confirm scenario, got %d", confirm)
	}
	if pending != 4 {
		t.Fatalf("want 4 leave-pending scenarios, got %d", pending)
	}
}

func TestChangeRequestUIScenariosQuick_subset(t *testing.T) {
	quick := ChangeRequestUIScenariosQuick()
	if len(quick) != 2 {
		t.Fatalf("want 2 quick scenarios, got %d", len(quick))
	}
	ids := map[string]bool{}
	for _, sc := range quick {
		ids[sc.ID] = true
	}
	if !ids["scenario-ack-pending"] || !ids["scenario-schedule-pending"] {
		t.Fatalf("quick subset should be ack + schedule, got %+v", ids)
	}
}

func TestChangeRequestUIScenarios_taskReviseTurns(t *testing.T) {
	all := ChangeRequestUIScenarios()
	var task *Scenario
	for i := range all {
		if all[i].ID == "scenario-task-dialogue-pending" {
			task = &all[i]
			break
		}
	}
	if task == nil {
		t.Fatal("missing scenario-task-dialogue-pending")
	}
	if len(task.Turns) != 4 {
		t.Fatalf("task scenario turns = %d want 4", len(task.Turns))
	}
	if task.MinRevision != 4 {
		t.Fatalf("MinRevision = %d want 4", task.MinRevision)
	}
	if !task.RequireTaskZone {
		t.Fatal("expected RequireTaskZone on task scenario")
	}
	if task.WantDueDate != "2026-07-20" {
		t.Fatalf("WantDueDate = %q", task.WantDueDate)
	}
	if task.WantTitle != "Refill calcium nitrate" {
		t.Fatalf("WantTitle = %q", task.WantTitle)
	}
}

func TestFilterScenariosByIDs(t *testing.T) {
	all := ChangeRequestUIScenarios()
	got := FilterScenariosByIDs(all, "scenario-ack-pending")
	if len(got) != 1 || got[0].ID != "scenario-ack-pending" {
		t.Fatalf("filter: %+v", got)
	}
}

func TestIsScenarioSuite(t *testing.T) {
	if !IsScenarioSuite("change-requests-ui") {
		t.Fatal("expected change-requests-ui")
	}
	if !IsScenarioSuite("change-requests-ui-quick") {
		t.Fatal("expected change-requests-ui-quick")
	}
	if IsScenarioSuite("change-requests") {
		t.Fatal("change-requests is single-turn fixtures, not scenarios")
	}
}
