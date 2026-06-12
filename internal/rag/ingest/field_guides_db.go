package ingest

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// FieldGuidesSource returns file or db from AGRONOMY_FIELD_GUIDES_SOURCE (default file).
func FieldGuidesSource() string {
	if strings.ToLower(strings.TrimSpace(os.Getenv("AGRONOMY_FIELD_GUIDES_SOURCE"))) == "db" {
		return "db"
	}
	return "file"
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
		n, err := w.upsertFieldGuideFile(ctx, farmID, relPath, g.ID, chunks, domain, safety)
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
