package naturalfarmingpack

import (
	"testing"

	catalogpack "gr33n-api/internal/commonscatalog"
	"gr33n-api/internal/croplibrary"
)

func TestLoadSwitchoverCatalog(t *testing.T) {
	root, err := croplibrary.FindRepoRoot()
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := LoadSwitchoverCatalog(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := catalog["mericle_veg_to_jlf_v1"]; !ok {
		t.Fatal("missing mericle_veg_to_jlf_v1")
	}
	if _, ok := catalog["mericle_flower_to_ffj_v1"]; !ok {
		t.Fatal("missing mericle_flower_to_ffj_v1")
	}
}

func TestFilterVegSwitchoverPack(t *testing.T) {
	root, err := croplibrary.FindRepoRoot()
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := LoadSwitchoverCatalog(root)
	if err != nil {
		t.Fatal(err)
	}
	starter, err := LoadStarterPackBody(root)
	if err != nil {
		t.Fatal(err)
	}
	filtered := FilterStarterPack(starter, catalog["mericle_veg_to_jlf_v1"])
	if len(filtered.InputDefinitions) != 2 {
		t.Fatalf("inputs %d want 2", len(filtered.InputDefinitions))
	}
	if len(filtered.ApplicationRecipes) != 2 {
		t.Fatalf("recipes %d want 2", len(filtered.ApplicationRecipes))
	}
	if len(filtered.RecipeInputComponents) != 3 {
		t.Fatalf("components %d want 3", len(filtered.RecipeInputComponents))
	}
	if err := catalogpack.ValidatePublishBody(filtered); err != nil {
		t.Fatal(err)
	}
}

func TestFilterFlowerSwitchoverPack(t *testing.T) {
	root, err := croplibrary.FindRepoRoot()
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := LoadSwitchoverCatalog(root)
	if err != nil {
		t.Fatal(err)
	}
	starter, err := LoadStarterPackBody(root)
	if err != nil {
		t.Fatal(err)
	}
	filtered := FilterStarterPack(starter, catalog["mericle_flower_to_ffj_v1"])
	if len(filtered.InputDefinitions) != 2 {
		t.Fatalf("inputs %d want 2", len(filtered.InputDefinitions))
	}
	if len(filtered.ApplicationRecipes) != 1 {
		t.Fatalf("recipes %d want 1", len(filtered.ApplicationRecipes))
	}
	if err := catalogpack.ValidatePublishBody(filtered); err != nil {
		t.Fatal(err)
	}
}
