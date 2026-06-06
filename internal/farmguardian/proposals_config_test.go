package farmguardian

import "testing"

func TestMatchConfigToolIntent_CreateTaskHumidity(t *testing.T) {
	snap := Snapshot{
		ZoneNames: []string{"Flower Room", "Veg Room"},
		ActiveCycles: []ActiveCycle{{
			ID: 42, Name: "TomatoVeg", ZoneName: "Flower Room", Stage: "vegetative",
		}},
		UnreadAlertDetails: []UnreadAlertDetail{{
			ID: 9, Subject: "Humidity high — Flower Room",
		}},
	}
	tool, args, summary, ok := matchConfigToolIntent(
		"Create a task to check Flower Room humidity",
		snap,
	)
	if !ok || tool != "create_task" {
		t.Fatalf("got tool=%q ok=%v want create_task", tool, ok)
	}
	if args["title"] == nil || summary == "" {
		t.Fatalf("args=%v summary=%q", args, summary)
	}
}

func TestMatchConfigToolIntent_CreateTaskFromAlert(t *testing.T) {
	snap := Snapshot{
		UnreadAlertDetails: []UnreadAlertDetail{{
			ID: 4, Subject: "Humidity high — Flower Room",
		}},
	}
	tool, args, _, ok := matchConfigToolIntent("create a task from the humidity alert", snap)
	if !ok || tool != "create_task_from_alert" {
		t.Fatalf("got %q ok=%v", tool, ok)
	}
	id, okID := args["alert_id"].(int64)
	if !okID || id != 4 {
		t.Fatalf("alert_id %v", args["alert_id"])
	}
}

func TestTaskTitleFromQuestion_LowStockRefill(t *testing.T) {
	alert := UnreadAlertDetail{
		ID:         12,
		Subject:    "Inventory low: OHN at 1.00 (threshold 3.00)",
		SourceType: "inventory_low_stock",
	}
	title := taskTitleFromQuestion("Create a refill task from alert #12 for OHN", alert)
	if title != "Refill OHN" {
		t.Fatalf("got title %q want Refill OHN", title)
	}
}

func TestLowStockInputFromSubject(t *testing.T) {
	if got := lowStockInputFromSubject("Inventory low: OHN at 1.00 (threshold 3.00)"); got != "OHN" {
		t.Fatalf("got %q", got)
	}
}
