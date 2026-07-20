package commonscatalog

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"gr33n-api/internal/croplibrary"
)

func TestValidateJadamIndoorStarterPack(t *testing.T) {
	root, err := croplibrary.FindRepoRoot()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "data/natural-farming-packs/jadam_indoor_starter_recipes_v1.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	body, err := ParsePackBody(raw)
	if err != nil {
		t.Fatal(err)
	}
	if body.Kind != KindNaturalFarmingRecipePack {
		t.Fatalf("kind %q", body.Kind)
	}
	if body.PackKey != "jadam_indoor_starter_recipes_v1" {
		t.Fatalf("pack_key %q", body.PackKey)
	}
	if len(body.InputDefinitions) != 16 {
		t.Fatalf("inputs %d want 16", len(body.InputDefinitions))
	}
	if len(body.ApplicationRecipes) != 14 {
		t.Fatalf("recipes %d want 14", len(body.ApplicationRecipes))
	}
	if len(body.RecipeInputComponents) != 20 {
		t.Fatalf("components %d want 20", len(body.RecipeInputComponents))
	}
	if err := ValidatePublishBody(body); err != nil {
		t.Fatal(err)
	}
}

func TestValidatePublishNaturalFarmingPack(t *testing.T) {
	if err := ValidatePublishBody(PackBody{Kind: KindNaturalFarmingRecipePack}); err == nil {
		t.Fatal("expected error for empty inputs")
	}
	err := ValidatePublishBody(PackBody{
		Kind: KindNaturalFarmingRecipePack,
		InputDefinitions: []NFInputDefinitionSpec{{
			Name:     "JMS (JADAM Microbial Solution)",
			Category: "microbial_inoculant",
		}},
		ApplicationRecipes: []NFApplicationRecipeSpec{{
			Name:                  "JMS Soil Drench",
			TargetApplicationType: "soil_drench",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestParsePackBodyNaturalFarmingKind(t *testing.T) {
	raw := json.RawMessage(`{"catalog_version":"gr33n.commons_catalog.v1","kind":"natural_farming_recipe_pack","input_definitions":[{"name":"X","category":"other_ferment"}],"application_recipes":[{"name":"Y","target_application_type":"soil_drench"}]}`)
	b, err := ParsePackBody(raw)
	if err != nil {
		t.Fatal(err)
	}
	if b.Kind != KindNaturalFarmingRecipePack {
		t.Fatalf("kind %q", b.Kind)
	}
}
