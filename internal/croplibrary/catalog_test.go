package croplibrary_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gr33n-api/internal/croplibrary"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}

func TestCropLibraryYAML_Validates(t *testing.T) {
	root := repoRoot(t)
	cat, err := croplibrary.LoadCatalog(root, croplibrary.DefaultCatalogPath)
	if err != nil {
		t.Fatal(err)
	}
	if cat.Version < 3 {
		t.Fatalf("want version >= 3, got %d", cat.Version)
	}
	if len(cat.Crops) < 46 {
		t.Fatalf("want >= 46 crops, got %d", len(cat.Crops))
	}
	withStages := cat.CropsWithStages()
	if len(withStages) < 46 {
		t.Fatalf("want >= 46 crops with stages (complete library), got %d", len(withStages))
	}
	if len(cat.Unsupported) < 3 {
		t.Fatalf("want >= 3 unsupported entries, got %d", len(cat.Unsupported))
	}
}

func TestCropLibraryYAML_RejectInvalidStage(t *testing.T) {
	root := repoRoot(t)
	bad := []byte(`
version: 2
crops:
  - key: test_crop
    display_name: Test
    category: leafy
    stages:
      - stage: vegetative
        ec_min: 1.0
        ec_target: 1.2
        ec_max: 1.4
unsupported: []
`)
	path := filepath.Join(t.TempDir(), "bad.yaml")
	if err := os.WriteFile(path, bad, 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := croplibrary.LoadCatalog(root, path)
	if err == nil || !strings.Contains(err.Error(), "growth_stage_enum") {
		t.Fatalf("want growth_stage_enum error, got %v", err)
	}
}

func TestCropLibraryYAML_RejectECOutOfRange(t *testing.T) {
	root := repoRoot(t)
	bad := []byte(`
version: 2
crops:
  - key: test_crop
    display_name: Test
    category: leafy
    stages:
      - stage: early_veg
        ec_min: 1.5
        ec_target: 2.0
        ec_max: 12.0
unsupported: []
`)
	path := filepath.Join(t.TempDir(), "bad-ec.yaml")
	if err := os.WriteFile(path, bad, 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := croplibrary.LoadCatalog(root, path)
	if err == nil || !strings.Contains(err.Error(), "mS/cm") {
		t.Fatalf("want mS/cm range error, got %v", err)
	}
}

func TestGenerateSeedSQL_IdempotentPattern(t *testing.T) {
	root := repoRoot(t)
	cat, err := croplibrary.LoadCatalog(root, croplibrary.DefaultCatalogPath)
	if err != nil {
		t.Fatal(err)
	}
	sql := croplibrary.GenerateSeedSQL(cat)
	for _, needle := range []string{
		"WHERE NOT EXISTS",
		"growth_stage_enum",
		"gr33ncrops.crop_profiles",
		"gr33ncrops.crop_profile_stages",
		"'cannabis'",
	} {
		if !strings.Contains(sql, needle) {
			t.Fatalf("generated SQL missing %q", needle)
		}
	}
}
