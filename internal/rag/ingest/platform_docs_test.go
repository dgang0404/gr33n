package ingest

import (
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
	if len(m.Include) < 10 {
		t.Fatalf("expected >=10 include paths, got %d", len(m.Include))
	}
	files, err := m.ResolvePlatformDocFiles()
	if err != nil {
		t.Fatalf("ResolvePlatformDocFiles: %v", err)
	}
	if len(files) < 10 {
		t.Fatalf("expected >=10 files, got %d", len(files))
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

func utf8Count(s string) int {
	return len([]rune(s))
}
