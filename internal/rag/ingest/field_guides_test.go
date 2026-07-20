package ingest

import (
	"context"
	"strings"
	"testing"
)

func TestDryRunFieldGuides_FileManifest(t *testing.T) {
	t.Setenv("AGRONOMY_FIELD_GUIDES_SOURCE", "file")
	root := findRepoRoot(t)
	dry, err := DryRunFieldGuides(context.Background(), nil, root, "")
	if err != nil {
		t.Fatal(err)
	}
	if dry.TotalChunks < 6 {
		t.Fatalf("expected chunks from manifest, got %d", dry.TotalChunks)
	}
}

func TestFieldGuideMetadata(t *testing.T) {
	meta := fieldGuideMetadata("crop-cannabis-nutrition.md", "general", "safe", "", "cannabis", 4)
	if !strings.Contains(string(meta), `"crop_key":"cannabis"`) {
		t.Fatalf("meta: %s", meta)
	}
	if !strings.Contains(string(meta), `"catalog_version":4`) {
		t.Fatalf("meta: %s", meta)
	}
	nf := fieldGuideMetadata("natural-farming-jms.md", "natural_farming", "safe", "jadam", "", 5)
	if !strings.Contains(string(nf), `"domain":"natural_farming"`) {
		t.Fatalf("meta: %s", nf)
	}
	if !strings.Contains(string(nf), `"tradition":"jadam"`) {
		t.Fatalf("meta: %s", nf)
	}
}

func TestFieldGuideMetaDefaultsNaturalFarming(t *testing.T) {
	domain, safety := fieldGuideMetaDefaults("natural-farming-jms.md", map[string]string{})
	if domain != "natural_farming" {
		t.Fatalf("domain=%q", domain)
	}
	if safety != "safe" {
		t.Fatalf("safety=%q", safety)
	}
}

func TestCropKeyFromFieldGuideSlug(t *testing.T) {
	if got := cropKeyFromFieldGuideSlug("crop-tomato-nutrition.md"); got != "tomato" {
		t.Fatalf("got %q", got)
	}
	if got := cropKeyFromFieldGuideSlug("pi-wiring-basics.md"); got != "" {
		t.Fatalf("got %q", got)
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
