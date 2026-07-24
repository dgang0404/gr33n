package reciperevision

import (
	"encoding/json"
	"testing"
)

func TestFormulaSummaryFromSnapshot(t *testing.T) {
	raw, err := json.Marshal(Snapshot{
		Recipe: RecipeSnapshot{
			Name:          "JMS Soil Drench",
			DilutionRatio: strPtr("1:10"),
		},
		Components: []ComponentSnapshot{{
			InputDefinitionID: 5,
			InputName:         "JMS",
			PartValue:         2,
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	summary, err := FormulaSummaryFromSnapshot(raw)
	if err != nil {
		t.Fatal(err)
	}
	if summary["dilution_ratio"] != "1:10" {
		t.Fatalf("dilution_ratio: %v", summary["dilution_ratio"])
	}
	comps, ok := summary["components"].([]map[string]any)
	if !ok || len(comps) != 1 {
		t.Fatalf("components: %T %v", summary["components"], summary["components"])
	}
}

func strPtr(s string) *string { return &s }
