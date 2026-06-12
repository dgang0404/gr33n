package ingest

import (
	"context"
	"strings"
	"testing"
)

func TestDryRunFieldGuides(t *testing.T) {
	root := findRepoRoot(t)
	dry, err := DryRunFieldGuides(context.Background(), nil, root, "")
	if err != nil {
		t.Fatal(err)
	}
	if dry.TotalChunks < 6 {
		t.Fatalf("expected chunks from manifest, got %d", dry.TotalChunks)
	}
}

func TestFieldGuideDocument(t *testing.T) {
	doc := FieldGuideDocument("pi-wiring-basics.md", "GPIO 17", 0, 1)
	if !strings.Contains(doc, "field_guide") || !strings.Contains(doc, "pi-wiring-basics") {
		t.Fatalf("doc: %s", doc)
	}
}

func TestSplitYAMLFrontmatter(t *testing.T) {
	raw := "---\ndomain: pi\nsafety_tier: safe\n---\n\n# Body\n"
	body, meta := splitYAMLFrontmatter(raw)
	if meta["domain"] != "pi" || meta["safety_tier"] != "safe" {
		t.Fatalf("meta: %v", meta)
	}
	if !strings.Contains(body, "# Body") {
		t.Fatalf("body: %q", body)
	}
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	for _, rel := range []string{".", "..", "../..", "../../.."} {
		if _, err := LoadFieldGuideManifest(rel, ""); err == nil {
			return rel
		}
	}
	t.Fatal("repo root not found")
	return ""
}
