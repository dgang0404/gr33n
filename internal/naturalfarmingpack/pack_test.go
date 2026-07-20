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
	if _, ok := catalog.Packs["mericle_veg_to_jlf_v1"]; !ok {
		t.Fatal("missing mericle_veg_to_jlf_v1")
	}
	if _, ok := catalog.Packs["mericle_flower_to_ffj_v1"]; !ok {
		t.Fatal("missing mericle_flower_to_ffj_v1")
	}
	if _, ok := catalog.Packs["livestock_comfrey_feed_v1"]; !ok {
		t.Fatal("missing livestock_comfrey_feed_v1")
	}
}

func TestLoadLivestockPackBody(t *testing.T) {
	root, err := croplibrary.FindRepoRoot()
	if err != nil {
		t.Fatal(err)
	}
	body, err := LoadPackBody(root, "livestock_comfrey_feed_v1")
	if err != nil {
		t.Fatal(err)
	}
	if len(body.InputDefinitions) != 2 {
		t.Fatalf("inputs %d want 2", len(body.InputDefinitions))
	}
	if body.InputDefinitions[0].Category != "animal_feed" {
		t.Fatalf("category %q", body.InputDefinitions[0].Category)
	}
	if err := catalogpack.ValidatePublishBody(body); err != nil {
		t.Fatal(err)
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
	filtered := FilterStarterPack(starter, catalog.Packs["mericle_veg_to_jlf_v1"])
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
	filtered := FilterStarterPack(starter, catalog.Packs["mericle_flower_to_ffj_v1"])
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

func TestLivestockPackApplyFull(t *testing.T) {
	root, err := croplibrary.FindRepoRoot()
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := LoadSwitchoverCatalog(root)
	if err != nil {
		t.Fatal(err)
	}
	spec := catalog.Packs["livestock_comfrey_feed_v1"]
	if !spec.ApplyFull {
		t.Fatal("expected apply_full on livestock pack")
	}
	body, err := LoadPackBody(root, effectiveSourcePack(spec, catalog.DefaultSourcePack))
	if err != nil {
		t.Fatal(err)
	}
	if len(body.ApplicationRecipes) != 2 {
		t.Fatalf("recipes %d want 2", len(body.ApplicationRecipes))
	}
}
