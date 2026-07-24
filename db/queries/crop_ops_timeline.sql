-- Phase 211.02 WS5 — crop cycle ops timeline queries

-- name: ListCropCycleStageEventsInRange :many
SELECT * FROM gr33nfertigation.crop_cycle_stage_events
WHERE crop_cycle_id = sqlc.arg('crop_cycle_id')
  AND entered_at >= sqlc.arg('from_ts')::timestamptz
  AND entered_at <= sqlc.arg('to_ts')::timestamptz
ORDER BY entered_at ASC, id ASC;

-- name: ListFertigationEventsForCropCycleInRange :many
SELECT * FROM gr33nfertigation.fertigation_events
WHERE farm_id = sqlc.arg('farm_id')
  AND crop_cycle_id = sqlc.arg('crop_cycle_id')
  AND applied_at >= sqlc.arg('from_ts')::timestamptz
  AND applied_at <= sqlc.arg('to_ts')::timestamptz
ORDER BY applied_at ASC, id ASC;

-- name: ListProgramAutomationRunsForZoneInRange :many
SELECT
    ar.id,
    ar.farm_id,
    ar.schedule_id,
    ar.rule_id,
    ar.program_id,
    ar.status,
    ar.message,
    ar.details,
    ar.executed_at,
    p.name AS program_name,
    p.target_zone_id
FROM gr33ncore.automation_runs ar
INNER JOIN gr33nfertigation.programs p ON p.id = ar.program_id
WHERE ar.farm_id = sqlc.arg('farm_id')
  AND p.target_zone_id = sqlc.arg('zone_id')
  AND p.deleted_at IS NULL
  AND ar.program_id IS NOT NULL
  AND ar.executed_at >= sqlc.arg('from_ts')::timestamptz
  AND ar.executed_at <= sqlc.arg('to_ts')::timestamptz
ORDER BY ar.executed_at ASC, ar.id ASC;

-- name: ListMixingEventsForZoneInRange :many
SELECT
    me.id,
    me.farm_id,
    me.reservoir_id,
    me.program_id,
    me.mixed_at,
    me.water_volume_liters,
    me.final_ec_mscm,
    me.final_ph,
    me.metadata,
    COALESCE(p.name, '') AS program_name
FROM gr33nfertigation.mixing_events me
LEFT JOIN gr33nfertigation.programs p ON p.id = me.program_id AND p.deleted_at IS NULL
LEFT JOIN gr33nfertigation.reservoirs r ON r.id = me.reservoir_id AND r.deleted_at IS NULL
WHERE me.farm_id = sqlc.arg('farm_id')
  AND me.mixed_at >= sqlc.arg('from_ts')::timestamptz
  AND me.mixed_at <= sqlc.arg('to_ts')::timestamptz
  AND (
    p.target_zone_id = sqlc.arg('zone_id')
    OR r.zone_id = sqlc.arg('zone_id')
  )
ORDER BY me.mixed_at ASC, me.id ASC;

-- name: ListLightingAutomationRunsForCropCycleInRange :many
SELECT
    ar.id,
    ar.schedule_id,
    ar.status,
    ar.message,
    ar.details,
    ar.executed_at,
    lp.id AS lighting_program_id,
    lp.name AS lighting_program_name,
    lp.on_hours,
    lp.off_hours,
    lp.lights_on_at
FROM gr33ncore.automation_runs ar
INNER JOIN gr33ncore.lighting_programs lp
    ON (lp.schedule_on_id = ar.schedule_id OR lp.schedule_off_id = ar.schedule_id)
WHERE ar.farm_id = sqlc.arg('farm_id')
  AND lp.crop_cycle_id = sqlc.arg('crop_cycle_id')
  AND ar.executed_at >= sqlc.arg('from_ts')::timestamptz
  AND ar.executed_at <= sqlc.arg('to_ts')::timestamptz
ORDER BY ar.executed_at ASC, ar.id ASC;
