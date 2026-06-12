package ingest

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// FieldGuidesSource returns db or file from AGRONOMY_FIELD_GUIDES_SOURCE (default db after Phase 84 WS-G).
func FieldGuidesSource() string {
	s := strings.ToLower(strings.TrimSpace(os.Getenv("AGRONOMY_FIELD_GUIDES_SOURCE")))
	switch s {
	case "file", "yaml":
		return "file"
	default:
		return "db"
	}
}

// FieldGuidesSourceLabel returns a log-friendly source name.
func FieldGuidesSourceLabel() string {
	if FieldGuidesSource() == "db" {
		return "database (gr33ncrops.agronomy_field_guides)"
	}
	return "filesystem (docs/rag/field-guide-manifest.yaml — deprecated authoring path)"
}

// IngestFieldGuidesFromDB embeds published rows from gr33ncrops.agronomy_field_guides.
func (w *Worker) IngestFieldGuidesFromDB(ctx context.Context, farmID int64) (int, error) {
	if w == nil || w.Q == nil || w.Embedder == nil {
		return 0, fmt.Errorf("ingest worker not configured")
	}
	guides, err := w.Q.ListAgronomyFieldGuides(ctx)
	if err != nil {
		return 0, fmt.Errorf("list agronomy field guides: %w", err)
	}
	if len(guides) == 0 {
		return 0, fmt.Errorf("agronomy_field_guides empty — run catalog seed migration")
	}
	total := 0
	for _, g := range guides {
		chunks := chunkMarkdown(strings.TrimSpace(g.BodyMd))
		domain := ""
		if g.Domain != nil {
			domain = strings.TrimSpace(*g.Domain)
		}
		safety := strings.TrimSpace(g.SafetyTier)
		if safety == "" {
			safety = "safe"
		}
		relPath := g.Slug + ".md"
		n, err := w.upsertFieldGuideFile(ctx, farmID, relPath, g.ID, chunks, domain, safety, cropKeyFromFieldGuideSlug(g.Slug), int(g.CatalogVersion))
		if err != nil {
			return total, fmt.Errorf("%s: %w", g.Slug, err)
		}
		total += n
	}
	return total, nil
}

// DryRunFieldGuidesFromDB returns chunk estimates from DB guides.
func DryRunFieldGuidesFromDB(ctx context.Context, q FieldGuideQuerier) (FieldGuideDryRun, error) {
	if q == nil {
		return FieldGuideDryRun{}, fmt.Errorf("querier required")
	}
	guides, err := q.ListAgronomyFieldGuides(ctx)
	if err != nil {
		return FieldGuideDryRun{}, err
	}
	var files []FieldGuideFileSummary
	total := 0
	for _, g := range guides {
		chunks := chunkMarkdown(strings.TrimSpace(g.BodyMd))
		domain := ""
		if g.Domain != nil {
			domain = *g.Domain
		}
		files = append(files, FieldGuideFileSummary{
			RelPath:  g.Slug + ".md",
			SourceID: g.ID,
			Bytes:    len(g.BodyMd),
			Chunks:   len(chunks),
			Domain:   domain,
			Safety:   g.SafetyTier,
		})
		total += len(chunks)
	}
	return FieldGuideDryRun{Files: files, TotalChunks: total}, nil
}
