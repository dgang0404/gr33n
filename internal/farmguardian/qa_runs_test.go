package farmguardian

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSummarizeQARun_passCounts(t *testing.T) {
	arch := QARunArchive{
		UpdatedAt: "2026-01-02T12:00:00Z",
		Suite:     "smoke",
		Model:     "phi3:mini",
		Scores: []EvalQuestionScore{
			{ID: "a", Passed: true},
			{ID: "b", Passed: false},
			{ID: "c", Passed: true},
		},
	}
	sum := SummarizeQARun(arch)
	if sum.Passed != 2 || sum.Total != 3 || sum.AllPassed {
		t.Fatalf("summary: %+v", sum)
	}
}

func TestLoadLatestQARun_picksNewest(t *testing.T) {
	dir := t.TempDir()
	oldPath := filepath.Join(dir, "20260101T120000_smoke_phi3-mini.json")
	newPath := filepath.Join(dir, "20260102T120000_smoke_phi3-mini.json")
	writeQARunFile(t, oldPath, "2026-01-01T12:00:00Z", "old")
	writeQARunFile(t, newPath, "2026-01-02T12:00:00Z", "new")
	arch, path, err := LoadLatestQARun(dir)
	if err != nil {
		t.Fatal(err)
	}
	if path != newPath {
		t.Fatalf("path %q want %q", path, newPath)
	}
	if len(arch.Scores) != 1 || arch.Scores[0].ID != "new" {
		t.Fatalf("arch: %+v", arch)
	}
}

func writeQARunFile(t *testing.T, path, updatedAt, id string) {
	t.Helper()
	arch := QARunArchive{
		UpdatedAt: updatedAt,
		Suite:     "smoke",
		Model:     "phi3:mini",
		Scores:    []EvalQuestionScore{{ID: id, Passed: true}},
	}
	raw, err := json.Marshal(arch)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadLatestQARun_missingDir(t *testing.T) {
	_, _, err := LoadLatestQARun(filepath.Join(t.TempDir(), "missing"))
	if !os.IsNotExist(err) {
		t.Fatalf("want not exist, got %v", err)
	}
}
