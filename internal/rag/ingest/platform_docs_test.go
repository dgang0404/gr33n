package ingest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadPlatformDocManifest(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", "..", ".."))
	m, err := LoadPlatformDocManifest(repoRoot, "")
	if err != nil {
		t.Fatalf("LoadPlatformDocManifest: %v", err)
	}
	if len(m.Include) < 11 {
		t.Fatalf("expected >=11 include paths (incl. pattern-playbooks.md), got %d", len(m.Include))
	}
	files, err := m.ResolvePlatformDocFiles()
	if err != nil {
		t.Fatalf("ResolvePlatformDocFiles: %v", err)
	}
	if len(files) < 11 {
		t.Fatalf("expected >=11 files, got %d", len(files))
	}
	totalChunks := 0
	for _, f := range files {
		if f.Chunks < 1 {
			t.Fatalf("file %s has zero chunks", f.RelPath)
		}
		if f.SourceID <= 0 {
			t.Fatalf("file %s bad source id", f.RelPath)
		}
		totalChunks += f.Chunks
	}
	dry, err := DryRunPlatformDocs(repoRoot, "")
	if err != nil {
		t.Fatalf("DryRunPlatformDocs: %v", err)
	}
	if dry.TotalChunks != totalChunks {
		t.Fatalf("dry total %d != resolved %d", dry.TotalChunks, totalChunks)
	}
}

func TestPlatformDocSourceIDStable(t *testing.T) {
	a := PlatformDocSourceID("operator-tour.md")
	b := PlatformDocSourceID("operator-tour.md")
	if a != b || a <= 0 {
		t.Fatalf("unstable id %d %d", a, b)
	}
}

func TestChunkMarkdownSplitsLongDoc(t *testing.T) {
	long := strings.Repeat("word ", 5000)
	chunks := chunkMarkdown("# Title\n\n" + long)
	if len(chunks) < 2 {
		t.Fatalf("expected multiple chunks, got %d", len(chunks))
	}
	for i, ch := range chunks {
		if utf8Count(ch) > platformDocSplitThreshold+100 {
			t.Fatalf("chunk %d too large: %d runes", i, utf8Count(ch))
		}
	}
}

func TestPlatformDocumentFormat(t *testing.T) {
	doc := PlatformDocument("farm-guardian-architecture.md", "Confirm replays frozen args.", 0, 2)
	if !strings.Contains(doc, "platform_doc") || !strings.Contains(doc, "farm-guardian-architecture.md") {
		t.Fatalf("unexpected doc: %q", doc)
	}
}

// TestOperatorTourChunksIncludeGreenhouseClimate ensures Phase 36 §5b is present in
// chunk output so re-ingest picks up greenhouse guidance for Guardian RAG.
func TestOperatorTourChunksIncludeGreenhouseClimate(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", "..", ".."))
	path := filepath.Join(repoRoot, "docs", "operator-tour.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read operator-tour: %v", err)
	}
	chunks := chunkMarkdown(string(data))
	if len(chunks) < 2 {
		t.Fatalf("expected multiple chunks, got %d", len(chunks))
	}
	joined := strings.Join(chunks, "\n")
	for _, needle := range []string{
		"5b. Greenhouse shade",
		"greenhouse_climate",
		"summarize_zone_greenhouse_climate",
		"Block sun",
	} {
		if !strings.Contains(joined, needle) {
			t.Fatalf("operator-tour chunks missing %q", needle)
		}
	}
}

// TestOperatorTourChunksIncludePlantNeedsPhase38 ensures §4a is chunk-visible for Guardian RAG.
func TestOperatorTourChunksIncludePlantNeedsPhase38(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", "..", ".."))
	path := filepath.Join(repoRoot, "docs", "operator-tour.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read operator-tour: %v", err)
	}
	joined := strings.Join(chunkMarkdown(string(data)), "\n")
	for _, needle := range []string{
		"4a. Plant needs",
		"Water / Light / Climate",
		"without last-write-wins",
		"duration_seconds",
		"Phase 39",
	} {
		if !strings.Contains(joined, needle) {
			t.Fatalf("operator-tour chunks missing %q", needle)
		}
	}
}

func TestPatternPlaybooksInManifest(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", "..", ".."))
	m, err := LoadPlatformDocManifest(repoRoot, "")
	if err != nil {
		t.Fatalf("LoadPlatformDocManifest: %v", err)
	}
	found := false
	for _, p := range m.Include {
		if p == "pattern-playbooks.md" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("manifest must include pattern-playbooks.md for greenhouse_climate_v1")
	}
	files, err := m.ResolvePlatformDocFiles()
	if err != nil {
		t.Fatalf("ResolvePlatformDocFiles: %v", err)
	}
	var pb *PlatformDocFileSummary
	for i := range files {
		if files[i].RelPath == "pattern-playbooks.md" {
			pb = &files[i]
			break
		}
	}
	if pb == nil || pb.Chunks < 1 {
		t.Fatal("pattern-playbooks.md not resolved or has zero chunks")
	}
}

func utf8Count(s string) int {
	return len([]rune(s))
}
