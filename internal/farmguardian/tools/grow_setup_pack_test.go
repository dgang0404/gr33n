package tools

import "testing"

func TestParseGrowSetupPackArgs(t *testing.T) {
	pack, err := parseGrowSetupPackArgs(map[string]any{
		"profile":    "house_plant",
		"zone_id":    float64(12),
		"zone_name":  "Living Room",
		"plant":      map[string]any{"display_name": "Philodendron", "variety_or_cultivar": "heartleaf"},
		"cycle":      map[string]any{"name": "Philodendron — Living Room", "current_stage": "vegetative", "started_at": "2026-05-27"},
		"program":    map[string]any{"name": "Light feed", "total_volume_liters": 0.5, "ec_trigger_low": 0.8, "ph_trigger_low": 5.8, "ph_trigger_high": 6.5},
		"optional_task": map[string]any{"title": "Monitor first two weeks"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if pack.ZoneID != 12 || pack.Profile != "house_plant" {
		t.Fatalf("unexpected pack: %+v", pack)
	}
	if pack.Plant == nil || pack.Cycle == nil || pack.Program == nil || pack.OptTask == nil {
		t.Fatal("expected all setup pack sections")
	}
}

func TestParseGrowSetupPackArgs_MissingCycle(t *testing.T) {
	_, err := parseGrowSetupPackArgs(map[string]any{
		"zone_id": float64(1),
		"program": map[string]any{"name": "x", "total_volume_liters": 1, "ec_trigger_low": 1, "ph_trigger_low": 5, "ph_trigger_high": 6},
	})
	if err == nil {
		t.Fatal("expected cycle required")
	}
}

func TestGrowSetupPackSummary(t *testing.T) {
	got := GrowSetupPackSummary(map[string]any{
		"zone_name": "Living Room",
		"plant":     map[string]any{"display_name": "Philodendron"},
	})
	if got != "Setup pack: Philodendron in Living Room (plant + cycle + program)" {
		t.Fatalf("got %q", got)
	}
}

func TestApplyGrowSetupPackRegisteredHighRisk(t *testing.T) {
	if _, err := Lookup("apply_grow_setup_pack"); err != nil {
		t.Fatal(err)
	}
	if got := RiskTierForTool("apply_grow_setup_pack", nil); got != RiskHigh {
		t.Fatalf("risk %q want high", got)
	}
}

func TestCycleArgsFromSetupPack_BatchLabelFallback(t *testing.T) {
	args, err := cycleArgsFromSetupPack(growSetupPack{
		ZoneID: 3,
		Plant:  map[string]any{"display_name": "Philodendron"},
		Cycle: map[string]any{
			"name":          "Philodendron — Living Room",
			"current_stage": "vegetative",
			"started_at":    "2026-05-27",
		},
	}, "heartleaf")
	if err != nil {
		t.Fatal(err)
	}
	if args["batch_label"] != "heartleaf" {
		t.Fatalf("batch_label %#v", args["batch_label"])
	}
}
