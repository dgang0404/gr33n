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

-- name: ListInputDefinitionsByFarmUpdatedAfter :many
SELECT * FROM gr33nnaturalfarming.input_definitions
WHERE farm_id = sqlc.arg('farm_id') AND deleted_at IS NULL AND updated_at > sqlc.arg('updated_after')::timestamptz
ORDER BY updated_at ASC, id ASC;

-- name: ListInputBatchesByFarmUpdatedAfter :many
SELECT * FROM gr33nnaturalfarming.input_batches
WHERE farm_id = sqlc.arg('farm_id') AND deleted_at IS NULL AND updated_at > sqlc.arg('updated_after')::timestamptz
ORDER BY updated_at ASC, id ASC;

-- name: CreateInputDefinition :one
INSERT INTO gr33nnaturalfarming.input_definitions (
  farm_id, name, category, description, typical_ingredients,
  preparation_summary, storage_guidelines, safety_precautions, reference_source,
  unit_cost, unit_cost_currency, unit_cost_unit_id
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) RETURNING *;

-- name: UpdateInputDefinition :one
UPDATE gr33nnaturalfarming.input_definitions SET
  name=$2, category=$3, description=$4, typical_ingredients=$5,
  preparation_summary=$6, storage_guidelines=$7, safety_precautions=$8,
  reference_source=$9, unit_cost=$10, unit_cost_currency=$11, unit_cost_unit_id=$12,
  updated_at=NOW(), updated_by_user_id=$13
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
  observations_notes, low_stock_threshold
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18) RETURNING *;

-- name: UpdateInputBatch :one
UPDATE gr33nnaturalfarming.input_batches SET
  batch_identifier=$2, status=$3, actual_ready_date=$4,
  current_quantity_remaining=$5, storage_location=$6,
  observations_notes=$7, low_stock_threshold=$8,
  updated_at=NOW(), updated_by_user_id=$9
WHERE id=$1 AND deleted_at IS NULL RETURNING *;

-- name: SoftDeleteInputBatch :exec
UPDATE gr33nnaturalfarming.input_batches
SET deleted_at=NOW(), updated_by_user_id=$2 WHERE id=$1;

-- Phase 20.7 WS2/WS3 — autologger deduct + refund. RETURNING the new
-- remaining quantity lets the caller detect a negative-stock
-- condition and log a warning without a second round-trip.
-- name: DecrementInputBatchQuantity :one
UPDATE gr33nnaturalfarming.input_batches
SET current_quantity_remaining = current_quantity_remaining - $2::NUMERIC,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, current_quantity_remaining;

-- name: IncrementInputBatchQuantity :one
UPDATE gr33nnaturalfarming.input_batches
SET current_quantity_remaining = current_quantity_remaining + $2::NUMERIC,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, current_quantity_remaining;

-- Phase 20.7 WS5 — low-stock sweep: any batch whose remaining stock
-- has dropped below its opt-in threshold. The worker fires one alert
-- per batch per day (dedupe enforced in the worker, not here).
-- name: ListLowStockBatchesByFarm :many
SELECT b.id, b.farm_id, b.input_definition_id, b.batch_identifier,
       b.current_quantity_remaining, b.low_stock_threshold,
       b.quantity_unit_id,
       d.name AS input_name
FROM gr33nnaturalfarming.input_batches b
JOIN gr33nnaturalfarming.input_definitions d ON d.id = b.input_definition_id
WHERE b.farm_id = $1
  AND b.deleted_at IS NULL
  AND b.low_stock_threshold IS NOT NULL
  AND b.current_quantity_remaining < b.low_stock_threshold
ORDER BY b.id ASC;
