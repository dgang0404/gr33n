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
