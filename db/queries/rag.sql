-- ============================================================
-- Queries: gr33ncore.rag_embedding_chunks (Phase 24 RAG)
-- ============================================================

-- name: UpsertRagEmbeddingChunk :one
INSERT INTO gr33ncore.rag_embedding_chunks (
    farm_id,
    source_type,
    source_id,
    chunk_index,
    content_text,
    embedding,
    model_id,
    metadata
) VALUES (
    $1, $2, $3, $4,
    $5, $6, $7, $8
)
ON CONFLICT ON CONSTRAINT uq_rag_embedding_chunks_source_chunk
DO UPDATE SET
    content_text = EXCLUDED.content_text,
    embedding = EXCLUDED.embedding,
    model_id = EXCLUDED.model_id,
    metadata = EXCLUDED.metadata,
    updated_at = NOW()
RETURNING *;

-- name: DeleteRagChunksByFarmSource :exec
DELETE FROM gr33ncore.rag_embedding_chunks
WHERE farm_id = $1
  AND source_type = $2
  AND source_id = $3;

-- name: DeleteRagChunksByFarmAndSourceType :exec
DELETE FROM gr33ncore.rag_embedding_chunks
WHERE farm_id = $1
  AND source_type = $2;

-- name: CountRagChunksByFarm :one
SELECT COUNT(*)::bigint AS cnt
FROM gr33ncore.rag_embedding_chunks
WHERE farm_id = $1;

-- name: CountRagChunksByFarmSourceType :one
SELECT COUNT(*)::bigint AS cnt
FROM gr33ncore.rag_embedding_chunks
WHERE farm_id = $1
  AND source_type = $2;

-- Phase 135 — corpus freshness aggregates per farm (max updated_at per tier).
-- name: GetRagCorpusStatsByFarm :one
SELECT
    COUNT(*) FILTER (WHERE source_type = 'field_guide')::bigint AS field_guide_chunks,
    MAX(updated_at) FILTER (WHERE source_type = 'field_guide') AS field_guide_last_ingested_at,
    COUNT(*) FILTER (WHERE source_type = 'platform_doc')::bigint AS platform_doc_chunks,
    MAX(updated_at) FILTER (WHERE source_type = 'platform_doc') AS platform_last_ingested_at,
    COUNT(*) FILTER (WHERE source_type NOT IN ('field_guide', 'platform_doc', 'symptom_guide'))::bigint AS operational_chunks,
    MAX(updated_at) FILTER (WHERE source_type NOT IN ('field_guide', 'platform_doc', 'symptom_guide')) AS operational_last_ingested_at
FROM gr33ncore.rag_embedding_chunks
WHERE farm_id = $1;

-- Farm-scoped nearest-neighbor search (caller supplies query embedding; WS4 retrieval API).
-- name: SearchRagNearestNeighbors :many
SELECT
    id,
    farm_id,
    source_type,
    source_id,
    chunk_index,
    content_text,
    embedding,
    model_id,
    metadata,
    created_at,
    updated_at,
    embedding <=> sqlc.arg(query_embedding)::vector AS distance
FROM gr33ncore.rag_embedding_chunks
WHERE farm_id = sqlc.arg(farm_id)
ORDER BY embedding <=> sqlc.arg(query_embedding)::vector
LIMIT sqlc.arg(match_limit);

-- Same as above with optional metadata module + created_at range (hybrid filters).
-- name: SearchRagNearestNeighborsFiltered :many
SELECT
    id,
    farm_id,
    source_type,
    source_id,
    chunk_index,
    content_text,
    model_id,
    metadata,
    created_at,
    updated_at,
    embedding <=> sqlc.arg(query_embedding)::vector AS distance
FROM gr33ncore.rag_embedding_chunks
WHERE farm_id = sqlc.arg(farm_id)
  AND (sqlc.narg('module')::text IS NULL OR metadata->>'module' = sqlc.narg('module')::text)
  AND (sqlc.narg('created_since')::timestamptz IS NULL OR created_at >= sqlc.narg('created_since')::timestamptz)
  AND (sqlc.narg('created_until')::timestamptz IS NULL OR created_at <= sqlc.narg('created_until')::timestamptz)
ORDER BY embedding <=> sqlc.arg(query_embedding)::vector
LIMIT sqlc.arg(match_limit);
