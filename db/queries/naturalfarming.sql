-- ============================================================
-- Queries: gr33nnaturalfarming
-- ============================================================

-- name: GetInputDefinitionByID :one
SELECT * FROM gr33nnaturalfarming.input_definitions
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetInputBatchByID :one
SELECT * FROM gr33nnaturalfarming.input_batches
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListInputDefinitionsByFarm :many
SELECT * FROM gr33nnaturalfarming.input_definitions
WHERE farm_id = $1 AND deleted_at IS NULL
ORDER BY name ASC;

-- name: ListInputBatchesByFarm :many
SELECT * FROM gr33nnaturalfarming.input_batches
WHERE farm_id = $1 AND deleted_at IS NULL
ORDER BY creation_start_date DESC;

-- name: CreateInputDefinition :one
INSERT INTO gr33nnaturalfarming.input_definitions (
  farm_id, name, category, description, typical_ingredients,
  preparation_summary, storage_guidelines, safety_precautions, reference_source
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING *;

-- name: UpdateInputDefinition :one
UPDATE gr33nnaturalfarming.input_definitions SET
  name=$2, category=$3, description=$4, typical_ingredients=$5,
  preparation_summary=$6, storage_guidelines=$7, safety_precautions=$8,
  reference_source=$9, updated_at=NOW(), updated_by_user_id=$10
WHERE id=$1 AND deleted_at IS NULL RETURNING *;

-- name: SoftDeleteInputDefinition :exec
UPDATE gr33nnaturalfarming.input_definitions
SET deleted_at=NOW(), updated_by_user_id=$2 WHERE id=$1;

-- name: CreateInputBatch :one
INSERT INTO gr33nnaturalfarming.input_batches (
  farm_id, input_definition_id, batch_identifier, creation_start_date,
  creation_end_date, expected_ready_date, quantity_produced, quantity_unit_id,
  current_quantity_remaining, status, storage_location, shelf_life_days,
  ph_value, ec_value_ms_cm, ingredients_used, procedure_followed,
  observations_notes
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17) RETURNING *;

-- name: UpdateInputBatch :one
UPDATE gr33nnaturalfarming.input_batches SET
  batch_identifier=$2, status=$3, actual_ready_date=$4,
  current_quantity_remaining=$5, storage_location=$6,
  observations_notes=$7, updated_at=NOW(), updated_by_user_id=$8
WHERE id=$1 AND deleted_at IS NULL RETURNING *;

-- name: SoftDeleteInputBatch :exec
UPDATE gr33nnaturalfarming.input_batches
SET deleted_at=NOW(), updated_by_user_id=$2 WHERE id=$1;
