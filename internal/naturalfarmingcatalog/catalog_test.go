package naturalfarmingcatalog

import (
	"strings"
	"testing"

	"gr33n-api/internal/croplibrary"
)

func TestLoadMaterialCatalog(t *testing.T) {
	root, err := croplibrary.FindRepoRoot()
	if err != nil {
		t.Fatal(err)
	}
	cat, err := LoadMaterialCatalog(root)
	if err != nil {
		t.Fatal(err)
	}
	if !yamlVersionAtLeast1(cat["version"]) {
		t.Fatalf("version: %v", cat["version"])
	}
	mat, ok := MaterialByID(cat, "goldenrod")
	if !ok {
		t.Fatal("goldenrod not found")
	}
	if mat["source_tier"] != "extension_method" {
		t.Fatalf("goldenrod source_tier: %v", mat["source_tier"])
	}
	if _, ok := MaterialByID(cat, "not-a-material"); ok {
		t.Fatal("expected missing material")
	}
	matches := MaterialsMatchingQuery(cat, "Can I use Canadian goldenrod for JLF?")
	if len(matches) != 1 {
		t.Fatalf("goldenrod matches: %d", len(matches))
	}
	if matches[0]["id"] != "goldenrod" {
		t.Fatalf("id: %v", matches[0]["id"])
	}
}

func TestLoadRecipeCanon_JMSSoilDilution(t *testing.T) {
	root, err := croplibrary.FindRepoRoot()
	if err != nil {
		t.Fatal(err)
	}
	cat, err := LoadRecipeCanon(root)
	if err != nil {
		t.Fatal(err)
	}
	recipes, ok := cat["application_recipes"].([]any)
	if !ok || len(recipes) < 14 {
		t.Fatalf("application_recipes: %T len=%d", cat["application_recipes"], len(recipes))
	}
	var jmsSoil map[string]any
	for _, item := range recipes {
		r, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if r["seed_name"] == "JMS Soil Drench" {
			jmsSoil = r
			break
		}
	}
	if jmsSoil == nil {
		t.Fatal("JMS Soil Drench not found")
	}
	dil, _ := jmsSoil["dilution"].(string)
	if !strings.Contains(dil, "1:10") {
		t.Fatalf("dilution: %q", dil)
	}
	if strings.Contains(dil, "1:500") {
		t.Fatalf("unexpected 1:500 in dilution: %q", dil)
	}
}
