package tools

import (
	"strings"
	"testing"
)

func TestNaturalFarmingDraftToolsRegistered(t *testing.T) {
	for _, id := range []string{"draft_input_definition", "draft_application_recipe", "draft_input_batch"} {
		if _, err := Lookup(id); err != nil {
			t.Fatalf("lookup %s: %v", id, err)
		}
		if got := RiskTierForTool(id, nil); got != RiskMedium {
			t.Fatalf("%s risk %q want medium", id, got)
		}
	}
}

func TestDraftInputDefinitionArgsValidation(t *testing.T) {
	_, err := execDraftInputDefinition(t.Context(), ExecutorDeps{FarmID: 1}, map[string]any{})
	if err == nil || !strings.Contains(err.Error(), "name") {
		t.Fatalf("expected name/material error, got %v", err)
	}
	_, err = execDraftInputDefinition(t.Context(), ExecutorDeps{FarmID: 1}, map[string]any{
		"farm_id": float64(99),
		"name":    "JMS",
		"category": "microbial_inoculant",
	})
	if err == nil || !strings.Contains(err.Error(), "farm_id") {
		t.Fatalf("expected farm_id rejected, got %v", err)
	}
}

func TestDraftApplicationRecipeArgsValidation(t *testing.T) {
	_, err := execDraftApplicationRecipe(t.Context(), ExecutorDeps{FarmID: 1}, map[string]any{
		"name": "Goldenrod drench",
	})
	if err == nil || !strings.Contains(err.Error(), "target_application_type") {
		t.Fatalf("expected target_application_type required, got %v", err)
	}
	_, err = execDraftApplicationRecipe(t.Context(), ExecutorDeps{FarmID: 1}, map[string]any{
		"name":                    "Goldenrod drench",
		"target_application_type": "soil_drench",
	})
	if err == nil || !strings.Contains(err.Error(), "dilution_ratio") {
		t.Fatalf("expected dilution_ratio required, got %v", err)
	}
	_, err = execDraftApplicationRecipe(t.Context(), ExecutorDeps{FarmID: 1}, map[string]any{
		"name":                    "Goldenrod drench",
		"target_application_type": "not_a_target",
		"dilution_ratio":          "1:100",
	})
	if err == nil || !strings.Contains(err.Error(), "invalid target_application_type") {
		t.Fatalf("expected invalid target, got %v", err)
	}
}

func TestDraftInputBatchArgsValidation(t *testing.T) {
	_, err := execDraftInputBatch(t.Context(), ExecutorDeps{FarmID: 1}, map[string]any{})
	if err == nil || !strings.Contains(err.Error(), "input_definition_id") {
		t.Fatalf("expected input_definition_id required, got %v", err)
	}
}

func TestResolveDraftInputFromCatalog_Goldenrod(t *testing.T) {
	res, err := resolveDraftInputFromCatalog(map[string]any{"material_id": "goldenrod"})
	if err != nil {
		t.Fatal(err)
	}
	if res.category != "other_ferment" {
		t.Fatalf("category=%q want other_ferment", res.category)
	}
	if res.sourceTier != "extension_method" {
		t.Fatalf("source_tier=%q want extension_method", res.sourceTier)
	}
}
