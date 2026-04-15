-- ============================================================
-- Queries: gr33nnaturalfarming
-- ============================================================

-- name: ListInputDefinitionsByFarm :many
SELECT * FROM gr33nnaturalfarming.input_definitions
WHERE farm_id = $1 AND deleted_at IS NULL
ORDER BY name ASC;

-- name: ListInputBatchesByFarm :many
SELECT * FROM gr33nnaturalfarming.input_batches
WHERE farm_id = $1 AND deleted_at IS NULL
ORDER BY creation_start_date DESC;
