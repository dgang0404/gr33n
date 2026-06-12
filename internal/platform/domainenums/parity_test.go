package domainenums

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"gr33n-api/internal/croplibrary"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
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

func openAPIGrowthStages(t *testing.T) []string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(repoRoot(t), "openapi.yaml"))
	if err != nil {
		t.Fatalf("read openapi.yaml: %v", err)
	}
	re := regexp.MustCompile(`(?ms)GrowthStageEnum:\s*\n\s*type: string\s*\n\s*enum: \[([^\]]+)\]`)
	m := re.FindSubmatch(b)
	if len(m) != 2 {
		t.Fatal("GrowthStageEnum not found in openapi.yaml")
	}
	raw := string(m[1])
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func growthStageValuesFromAll() []string {
	stages := All().GrowthStages
	out := make([]string, len(stages))
	for i, s := range stages {
		out[i] = s.Value
	}
	return out
}

func TestParity_GrowthStagesMatchCropLibrary(t *testing.T) {
	got := growthStageValuesFromAll()
	if len(got) != len(croplibrary.ValidGrowthStages) {
		t.Fatalf("growth stage count: domainenums=%d croplibrary=%d", len(got), len(croplibrary.ValidGrowthStages))
	}
	for _, stage := range got {
		if _, ok := croplibrary.ValidGrowthStages[stage]; !ok {
			t.Fatalf("domain enum stage %q missing from croplibrary.ValidGrowthStages", stage)
		}
	}
	for stage := range croplibrary.ValidGrowthStages {
		found := false
		for _, g := range got {
			if g == stage {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("croplibrary stage %q missing from domainenums", stage)
		}
	}
}

func TestParity_GrowthStagesMatchOpenAPI(t *testing.T) {
	got := growthStageValuesFromAll()
	want := openAPIGrowthStages(t)
	if len(got) != len(want) {
		t.Fatalf("growth stage count: domainenums=%d openapi=%d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("growth stage order mismatch at %d: domainenums=%q openapi=%q", i, got[i], want[i])
		}
	}
}
