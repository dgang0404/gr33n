-- ============================================================
-- Queries: gr33nfertigation core resources
-- ============================================================

-- name: ListReservoirsByFarm :many
SELECT * FROM gr33nfertigation.reservoirs
WHERE farm_id = $1 AND deleted_at IS NULL
ORDER BY name ASC;

-- name: CreateReservoir :one
INSERT INTO gr33nfertigation.reservoirs (
    farm_id, zone_id, name, description, capacity_liters,
    current_volume_liters, status
) VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ListEcTargetsByFarm :many
SELECT * FROM gr33nfertigation.ec_targets
WHERE farm_id = $1
ORDER BY zone_id ASC, growth_stage ASC;

-- name: CreateEcTarget :one
INSERT INTO gr33nfertigation.ec_targets (
    farm_id, zone_id, growth_stage, ec_min_mscm, ec_max_mscm,
    ph_min, ph_max, notes, rationale
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: ListProgramsByFarm :many
SELECT * FROM gr33nfertigation.programs
WHERE farm_id = $1 AND deleted_at IS NULL
ORDER BY name ASC;

-- name: CreateProgram :one
INSERT INTO gr33nfertigation.programs (
    farm_id, name, description, application_recipe_id, reservoir_id,
    target_zone_id, schedule_id, ec_target_id, total_volume_liters,
    run_duration_seconds, ec_trigger_low, ph_trigger_low, ph_trigger_high,
    is_active
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9,
    $10, $11, $12, $13,
    $14
)
RETURNING *;

-- name: ListFertigationEventsByFarm :many
SELECT * FROM gr33nfertigation.fertigation_events
WHERE farm_id = $1
ORDER BY applied_at DESC;

-- name: CreateFertigationEvent :one
INSERT INTO gr33nfertigation.fertigation_events (
    farm_id, program_id, reservoir_id, zone_id, applied_at,
    growth_stage, volume_applied_liters, run_duration_seconds,
    ec_before_mscm, ec_after_mscm, ph_before, ph_after,
    trigger_source, notes, metadata
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8,
    $9, $10, $11, $12,
    $13, $14, $15
)
RETURNING *;

-- name: UpdateReservoir :one
UPDATE gr33nfertigation.reservoirs
SET name = $2, description = $3, capacity_liters = $4,
    current_volume_liters = $5, status = $6, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteReservoir :exec
UPDATE gr33nfertigation.reservoirs
SET deleted_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateProgram :one
UPDATE gr33nfertigation.programs
SET name = $2, description = $3, reservoir_id = $4,
    target_zone_id = $5, ec_target_id = $6,
    total_volume_liters = $7, is_active = $8, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteProgram :exec
UPDATE gr33nfertigation.programs
SET deleted_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;
