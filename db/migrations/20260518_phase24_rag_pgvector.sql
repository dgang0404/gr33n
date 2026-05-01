-- ============================================================
-- Phase 24 WS2 — pgvector extension + farm-scoped embedding chunks
-- ============================================================
-- Enables semantic retrieval indexes in Postgres (same trust boundary as
-- relational data). Requires pgvector installed on the server — see INSTALL.md
-- and Docker db image (db/Dockerfile).
--
-- Embedding dimension 1536 matches OpenAI text-embedding-3-small / ada-002
-- family defaults; WS3 must use the same model_id ↔ dimension pairing or add
-- a follow-up migration.
-- ============================================================

CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS gr33ncore.rag_embedding_chunks (
    id             BIGSERIAL PRIMARY KEY,
    farm_id        BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    source_type    TEXT NOT NULL,
    source_id      BIGINT NOT NULL,
    chunk_index    INTEGER NOT NULL DEFAULT 0 CHECK (chunk_index >= 0),
    content_text   TEXT NOT NULL,
    embedding      vector(1536) NOT NULL,
    model_id       TEXT NOT NULL,
    metadata       JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_rag_embedding_chunks_source_chunk UNIQUE (
        farm_id,
        source_type,
        source_id,
        chunk_index
    )
);

COMMENT ON TABLE gr33ncore.rag_embedding_chunks IS
  'Farm-scoped text chunks and embedding vectors for Phase 24 RAG; '
  'source_type + source_id + chunk_index form the dedupe key per farm.';

COMMENT ON COLUMN gr33ncore.rag_embedding_chunks.source_type IS
  'Stable embed source label, e.g. task, crop_cycle, automation_run.';

COMMENT ON COLUMN gr33ncore.rag_embedding_chunks.metadata IS
  'Optional filters: module, zone_id, date bounds, etc. — no secrets.';

CREATE INDEX IF NOT EXISTS idx_rag_embedding_chunks_farm
    ON gr33ncore.rag_embedding_chunks (farm_id);

CREATE INDEX IF NOT EXISTS idx_rag_embedding_chunks_farm_source
    ON gr33ncore.rag_embedding_chunks (farm_id, source_type, source_id);

CREATE INDEX IF NOT EXISTS idx_rag_embedding_chunks_embedding_hnsw
    ON gr33ncore.rag_embedding_chunks
    USING hnsw (embedding vector_cosine_ops)
    WITH (m = 16, ef_construction = 64);

DROP TRIGGER IF EXISTS trg_rag_embedding_chunks_updated_at ON gr33ncore.rag_embedding_chunks;
CREATE TRIGGER trg_rag_embedding_chunks_updated_at
    BEFORE UPDATE ON gr33ncore.rag_embedding_chunks
    FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();
