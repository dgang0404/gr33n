package tools

import "testing"

func TestCreatePlantArgsValidation(t *testing.T) {
	_, err := Lookup("create_plant")
	if err != nil {
		t.Fatal(err)
	}
	_, err = execCreatePlant(t.Context(), ExecutorDeps{FarmID: 1}, map[string]any{})
	if err == nil {
		t.Fatal("expected display_name required")
	}
}

func TestCreateCropCycleArgsValidation(t *testing.T) {
	_, err := Lookup("create_crop_cycle")
	if err != nil {
		t.Fatal(err)
	}
	_, err = execCreateCropCycle(t.Context(), ExecutorDeps{FarmID: 1}, map[string]any{
		"zone_id": 1,
		"name":    "Test cycle",
	})
	if err == nil {
		t.Fatal("expected current_stage required")
	}
}

func TestCreateFertigationProgramArgsValidation(t *testing.T) {
	_, err := Lookup("create_fertigation_program")
	if err != nil {
		t.Fatal(err)
	}
	_, err = execCreateFertigationProgram(t.Context(), ExecutorDeps{FarmID: 1}, map[string]any{
		"name":           "Light feed",
		"target_zone_id": 1,
	})
	if err == nil {
		t.Fatal("expected volume/trigger fields required")
	}
}

func TestRiskTierGrowCreateTools(t *testing.T) {
	for _, id := range []string{"create_plant", "create_crop_cycle", "create_fertigation_program"} {
		if got := RiskTierForTool(id, nil); got != RiskMedium {
			t.Fatalf("%s risk %q want medium", id, got)
		}
	}
}
