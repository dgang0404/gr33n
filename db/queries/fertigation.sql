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

-- name: ListProgramsByFarmUpdatedAfter :many
SELECT * FROM gr33nfertigation.programs
WHERE farm_id = sqlc.arg('farm_id') AND deleted_at IS NULL AND updated_at > sqlc.arg('updated_after')::timestamptz
ORDER BY updated_at ASC, id ASC;

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

-- name: ListFertigationEventsByFarmAndCropCycle :many
SELECT * FROM gr33nfertigation.fertigation_events
WHERE farm_id = $1 AND crop_cycle_id = $2
ORDER BY applied_at DESC;

-- name: CreateFertigationEvent :one
INSERT INTO gr33nfertigation.fertigation_events (
    farm_id, program_id, reservoir_id, zone_id, crop_cycle_id, applied_at,
    growth_stage, volume_applied_liters, run_duration_seconds,
    ec_before_mscm, ec_after_mscm, ph_before, ph_after,
    trigger_source, notes, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6,
    $7, $8, $9,
    $10, $11, $12, $13,
    $14, $15, $16
)
RETURNING *;

-- name: GetFertigationReservoirByID :one
SELECT * FROM gr33nfertigation.reservoirs
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetFertigationProgramByID :one
SELECT * FROM gr33nfertigation.programs
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListActivePrograms :many
-- Phase 22 WS1 — feeds the worker's program-tick. Only programs with a
-- bound schedule are dispatched automatically (unscheduled programs are
-- template-only and require an explicit "run now" API call, added later).
SELECT * FROM gr33nfertigation.programs
WHERE is_active = TRUE
  AND deleted_at IS NULL
  AND schedule_id IS NOT NULL;

-- name: MarkProgramTriggered :one
UPDATE gr33nfertigation.programs
SET last_triggered_time = $2, updated_at = NOW()
WHERE id = $1
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

-- name: GetMixingEventByID :one
SELECT * FROM gr33nfertigation.mixing_events WHERE id = $1;

-- name: ListMixingEventsByFarm :many
SELECT * FROM gr33nfertigation.mixing_events
WHERE farm_id = $1
ORDER BY mixed_at DESC;

-- name: ListMixingEventComponents :many
SELECT * FROM gr33nfertigation.mixing_event_components
WHERE mixing_event_id = $1
ORDER BY id ASC;

-- name: CreateMixingEvent :one
INSERT INTO gr33nfertigation.mixing_events (
    farm_id, reservoir_id, program_id, mixed_by_user_id, mixed_at,
    water_volume_liters, water_source, water_ec_mscm, water_ph,
    final_ec_mscm, final_ph, final_temp_celsius,
    ec_target_id, ec_target_met, notes, observations
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
RETURNING *;

-- name: CreateMixingEventComponent :one
INSERT INTO gr33nfertigation.mixing_event_components (
    mixing_event_id, input_definition_id, input_batch_id,
    volume_added_ml, dilution_ratio, notes
) VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;
