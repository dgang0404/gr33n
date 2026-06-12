package croplibrary_test

import (
	"strings"
	"testing"

	"gr33n-api/internal/croplibrary"
)

func TestLoadFieldGuideSeeds(t *testing.T) {
	root := repoRoot(t)
	cat, err := croplibrary.LoadCatalog(root, croplibrary.DefaultCatalogPath)
	if err != nil {
		t.Fatal(err)
	}
	guides, err := croplibrary.LoadFieldGuideSeeds(root, croplibrary.DefaultFieldGuideManifest, cat)
	if err != nil {
		t.Fatal(err)
	}
	if len(guides) < 50 {
		t.Fatalf("want >= 50 field guides, got %d", len(guides))
	}
	var zucchini *croplibrary.FieldGuideSeed
	for i := range guides {
		if guides[i].Slug == "crop-zucchini-nutrition" {
			zucchini = &guides[i]
			break
		}
	}
	if zucchini == nil {
		t.Fatal("missing crop-zucchini-nutrition")
	}
	if zucchini.CropKey != "zucchini" {
		t.Fatalf("crop_key: %q", zucchini.CropKey)
	}
	if !strings.Contains(zucchini.BodyMD, "mS/cm") {
		t.Fatalf("body should mention mS/cm: %q", zucchini.BodyMD[:min(80, len(zucchini.BodyMD))])
	}
}

func TestGenerateCatalogSeedSQL(t *testing.T) {
	root := repoRoot(t)
	cat, err := croplibrary.LoadCatalog(root, croplibrary.DefaultCatalogPath)
	if err != nil {
		t.Fatal(err)
	}
	guides, err := croplibrary.LoadFieldGuideSeeds(root, croplibrary.DefaultFieldGuideManifest, cat)
	if err != nil {
		t.Fatal(err)
	}
	sql := croplibrary.GenerateCatalogSeedSQL(cat, guides)
	if !strings.Contains(sql, "crop_catalog_entries") {
		t.Fatal("missing catalog entries insert")
	}
	if !strings.Contains(sql, "'ramps'") || !strings.Contains(sql, "unsupported_reason") {
		t.Fatal("missing ramps unsupported seed")
	}
	if !strings.Contains(sql, "agronomy_field_guides") {
		t.Fatal("missing field guides insert")
	}
	if !strings.Contains(sql, "crop-zucchini-nutrition") {
		t.Fatal("missing zucchini guide seed")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
