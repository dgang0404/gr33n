// Phase 32 WS8 — platform doc RAG manifest + chunk storage smoke.
package main

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pgvector/pgvector-go"

	"gr33n-api/internal/rag/embed"
	"gr33n-api/internal/rag/ingest"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("repo root: %v", err)
	}
	return root
}

func TestPhase32WS8_PlatformDocManifestDryRun(t *testing.T) {
	dry, err := ingest.DryRunPlatformDocs(repoRoot(t), "")
	if err != nil {
		t.Fatalf("DryRunPlatformDocs: %v", err)
	}
	if len(dry.Files) < 10 {
		t.Fatalf("expected >=10 manifest files, got %d", len(dry.Files))
	}
	if dry.TotalChunks < len(dry.Files) {
		t.Fatalf("expected >=1 chunk per file, total=%d files=%d", dry.TotalChunks, len(dry.Files))
	}
}

func TestPhase32WS8_PlatformDocChunkUpsert(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	relPath := "farm-guardian-architecture.md"
	sourceID := ingest.PlatformDocSourceID(relPath)
	content := ingest.PlatformDocument(relPath, "High-tier actuator PRs require operator Confirm on the proposal card before enqueue_actuator_command runs.", 0, 1)
	vec := make([]float32, embed.DefaultExpectedDims)
	vec[0] = 1
	vec[2] = 0.5

	_, err := testPool.Exec(ctx, `
INSERT INTO gr33ncore.rag_embedding_chunks
    (farm_id, source_type, source_id, chunk_index, content_text, embedding, model_id, metadata)
VALUES (1, 'platform_doc', $1, 0, $2, $3, 'smoke-embed', $4::jsonb)
ON CONFLICT ON CONSTRAINT uq_rag_embedding_chunks_source_chunk
DO UPDATE SET content_text = EXCLUDED.content_text, embedding = EXCLUDED.embedding, updated_at = NOW()`,
		sourceID, content, pgvector.NewVector(vec), `{"module":"platform_doc","doc_path":"farm-guardian-architecture.md"}`)
	if err != nil {
		if strings.Contains(err.Error(), "rag_embedding_chunks") && strings.Contains(err.Error(), "does not exist") {
			t.Skip("rag_embedding_chunks table not migrated")
		}
		t.Fatalf("upsert platform_doc chunk: %v", err)
	}
	t.Cleanup(func() {
		c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = testPool.Exec(c, `
DELETE FROM gr33ncore.rag_embedding_chunks
WHERE farm_id = 1 AND source_type = 'platform_doc' AND source_id = $1`, sourceID)
	})

	var count int
	err = testPool.QueryRow(ctx, `
SELECT COUNT(*) FROM gr33ncore.rag_embedding_chunks
WHERE farm_id = 1 AND source_type = 'platform_doc' AND source_id = $1`, sourceID).Scan(&count)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 chunk, got %d", count)
	}
}
