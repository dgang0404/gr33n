package croplibrary_test

import (
	"testing"

	"gr33n-api/internal/croplibrary"
)

func TestBuildPicker_GroupsAndTargets(t *testing.T) {
	root := repoRoot(t)
	cat, err := croplibrary.LoadCatalog(root, croplibrary.DefaultCatalogPath)
	if err != nil {
		t.Fatal(err)
	}
	idTomato := int64(10)
	idEggplant := int64(11)
	profiles := []croplibrary.ProfileRow{
		{ID: idTomato, CropKey: "tomato", DisplayName: "Tomato", IsBuiltin: true, StageCount: 5},
		{ID: idEggplant, CropKey: "eggplant", DisplayName: "Eggplant", IsBuiltin: true, StageCount: 4},
	}
	out := croplibrary.BuildPicker(cat, profiles)
	if out.Counts.WithTargets < 2 {
		t.Fatalf("want >= 2 with targets, got %d", out.Counts.WithTargets)
	}
	if len(out.Groups) == 0 {
		t.Fatal("want groups")
	}
	var tomato *croplibrary.PickerItem
	for _, g := range out.Groups {
		for i := range g.Items {
			if g.Items[i].CropKey == "tomato" {
				tomato = &g.Items[i]
			}
		}
	}
	if tomato == nil || !tomato.HasTargets || tomato.CropProfileID == nil || *tomato.CropProfileID != idTomato {
		t.Fatalf("tomato picker item: %+v", tomato)
	}
}

func TestBuildPicker_CatalogOnlyHasCousin(t *testing.T) {
	root := repoRoot(t)
	cat, err := croplibrary.LoadCatalog(root, croplibrary.DefaultCatalogPath)
	if err != nil {
		t.Fatal(err)
	}
	out := croplibrary.BuildPicker(cat, nil)
	var rose *croplibrary.PickerItem
	for _, g := range out.Groups {
		for i := range g.Items {
			if g.Items[i].CropKey == "rose" {
				rose = &g.Items[i]
			}
		}
	}
	if rose == nil {
		t.Fatal("rose not in picker")
	}
	if rose.HasTargets {
		t.Fatal("rose should be catalog-only without DB seed")
	}
	if rose.CousinOf == nil || *rose.CousinOf != "tomato" {
		t.Fatalf("cousin_of: %+v", rose.CousinOf)
	}
}
