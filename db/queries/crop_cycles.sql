-- ============================================================
-- Queries: gr33nfertigation.crop_cycles
-- ============================================================

-- name: CreateCropCycle :one
INSERT INTO gr33nfertigation.crop_cycles (
    farm_id, zone_id, name, strain_or_variety, current_stage,
    is_active, started_at, cycle_notes, primary_program_id
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: ListCropCyclesByFarm :many
SELECT * FROM gr33nfertigation.crop_cycles
WHERE farm_id = $1
ORDER BY started_at DESC;

-- name: ListCropCyclesByFarmUpdatedAfter :many
SELECT * FROM gr33nfertigation.crop_cycles
WHERE farm_id = sqlc.arg('farm_id') AND updated_at > sqlc.arg('updated_after')::timestamptz
ORDER BY updated_at ASC, id ASC;

-- name: GetCropCycleByID :one
SELECT * FROM gr33nfertigation.crop_cycles WHERE id = $1;

-- name: UpdateCropCycle :one
UPDATE gr33nfertigation.crop_cycles SET
    name = $2,
    strain_or_variety = $3,
    zone_id = $4,
    is_active = $5,
    cycle_notes = $6,
    harvested_at = $7,
    yield_grams = $8,
    yield_notes = $9,
    primary_program_id = $10,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateCropCycleStage :one
UPDATE gr33nfertigation.crop_cycles SET
    current_stage = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: SoftDeleteCropCycle :exec
UPDATE gr33nfertigation.crop_cycles SET is_active = FALSE, updated_at = NOW() WHERE id = $1;

-- name: GetActiveCropCycleForZone :one
-- Phase 20.6 WS3 — the rule engine calls this to translate a rule's zone
-- into the active cycle so the setpoint resolver has a `current_stage`
-- to match against. At most one crop cycle per zone is active at a time
-- (enforced by uq_active_crop_cycle).
SELECT * FROM gr33nfertigation.crop_cycles
WHERE zone_id = $1 AND is_active = TRUE
LIMIT 1;
