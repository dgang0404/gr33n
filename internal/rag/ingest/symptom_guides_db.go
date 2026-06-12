package ingest

import (
	"context"
	"fmt"
	"strings"

	"github.com/pgvector/pgvector-go"

	db "gr33n-api/internal/db"
)

// IngestSymptomGuidesFromDB embeds published agronomy_symptom_entries (source_type symptom_guide).
func (w *Worker) IngestSymptomGuidesFromDB(ctx context.Context, farmID int64) (int, error) {
	if w == nil || w.Q == nil || w.Embedder == nil {
		return 0, fmt.Errorf("ingest worker not configured")
	}
	entries, err := w.Q.ListAgronomySymptomEntries(ctx)
	if err != nil {
		return 0, fmt.Errorf("list agronomy symptom entries: %w", err)
	}
	if len(entries) == 0 {
		return 0, fmt.Errorf("agronomy_symptom_entries empty — run phase 106 migration")
	}
	total := 0
	for _, e := range entries {
		body := strings.TrimSpace(e.DisplayName) + "\n\n" + strings.TrimSpace(e.BodyMd)
		chunks := chunkMarkdown(body)
		relPath := "symptoms/" + e.SymptomKey + ".md"
		n, err := w.upsertSymptomGuideFile(ctx, farmID, relPath, e.ID, chunks)
		if err != nil {
			return total, fmt.Errorf("%s: %w", e.SymptomKey, err)
		}
		total += n
	}
	return total, nil
}

func (w *Worker) upsertSymptomGuideFile(ctx context.Context, farmID int64, relPath string, sourceID int64, chunks []string) (int, error) {
	if len(chunks) == 0 {
		return 0, nil
	}
	if err := w.Q.DeleteRagChunksByFarmSource(ctx, db.DeleteRagChunksByFarmSourceParams{
		FarmID:     farmID,
		SourceType: SourceTypeSymptomGuide,
		SourceID:   sourceID,
	}); err != nil {
		return 0, err
	}
	texts := make([]string, len(chunks))
	for i, ch := range chunks {
		texts[i] = FieldGuideDocument(relPath, ch, i, len(chunks))
	}
	vecs, err := w.Embedder.Embed(ctx, texts)
	if err != nil {
		return 0, err
	}
	if len(vecs) != len(texts) {
		return 0, fmt.Errorf("embed count %d != chunk count %d", len(vecs), len(texts))
	}
	meta := fieldGuideMetadata(relPath, "agronomy", "safe", "", 0)
	modelID := w.Embedder.ModelID()
	n := 0
	for i, text := range texts {
		_, err := w.Q.UpsertRagEmbeddingChunk(ctx, db.UpsertRagEmbeddingChunkParams{
			FarmID:      farmID,
			SourceType:  SourceTypeSymptomGuide,
			SourceID:    sourceID,
			ChunkIndex:  int32(i),
			ContentText: text,
			Embedding:   pgvector.NewVector(vecs[i]),
			ModelID:     modelID,
			Metadata:    meta,
		})
		if err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}
