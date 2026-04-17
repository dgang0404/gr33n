-- Phase 20.6 WS1 — stage-scoped setpoints.
-- The precedence resolver (GetActiveSetpointForScope) is the query the
-- rule-engine calls on every setpoint-typed predicate evaluation.

-- name: ListSetpointsByFarm :many
SELECT * FROM gr33ncore.zone_setpoints
WHERE farm_id = $1
ORDER BY zone_id NULLS LAST, crop_cycle_id NULLS LAST, sensor_type, stage NULLS LAST;

-- name: ListSetpointsByFarmFiltered :many
-- Optional filters: zone_id ($2), crop_cycle_id ($3), sensor_type ($4).
-- Pass NULL / empty-string to skip a filter. Keeping one query with
-- nullable filter args avoids the handler having to branch.
SELECT * FROM gr33ncore.zone_setpoints
WHERE farm_id = $1
  AND (sqlc.narg('zone_id')::BIGINT       IS NULL OR zone_id       = sqlc.narg('zone_id')::BIGINT)
  AND (sqlc.narg('crop_cycle_id')::BIGINT IS NULL OR crop_cycle_id = sqlc.narg('crop_cycle_id')::BIGINT)
  AND (sqlc.narg('sensor_type')::TEXT     IS NULL OR sensor_type   = sqlc.narg('sensor_type')::TEXT)
ORDER BY zone_id NULLS LAST, crop_cycle_id NULLS LAST, sensor_type, stage NULLS LAST;

-- name: ListSetpointsByZone :many
SELECT * FROM gr33ncore.zone_setpoints
WHERE zone_id = $1
ORDER BY crop_cycle_id NULLS LAST, sensor_type, stage NULLS LAST;

-- name: ListSetpointsByCropCycle :many
SELECT * FROM gr33ncore.zone_setpoints
WHERE crop_cycle_id = $1
ORDER BY sensor_type, stage NULLS LAST;

-- name: GetSetpointByID :one
SELECT * FROM gr33ncore.zone_setpoints
WHERE id = $1;

-- name: CreateSetpoint :one
INSERT INTO gr33ncore.zone_setpoints (
    farm_id, zone_id, crop_cycle_id, stage, sensor_type,
    min_value, max_value, ideal_value, meta
) VALUES (
    sqlc.arg('farm_id'),
    sqlc.narg('zone_id'),
    sqlc.narg('crop_cycle_id'),
    sqlc.narg('stage'),
    sqlc.arg('sensor_type'),
    sqlc.narg('min_value'),
    sqlc.narg('max_value'),
    sqlc.narg('ideal_value'),
    COALESCE(sqlc.narg('meta')::jsonb, '{}'::jsonb)
)
RETURNING *;

-- name: UpdateSetpoint :one
UPDATE gr33ncore.zone_setpoints
SET zone_id       = sqlc.narg('zone_id'),
    crop_cycle_id = sqlc.narg('crop_cycle_id'),
    stage         = sqlc.narg('stage'),
    sensor_type   = sqlc.arg('sensor_type'),
    min_value     = sqlc.narg('min_value'),
    max_value     = sqlc.narg('max_value'),
    ideal_value   = sqlc.narg('ideal_value'),
    meta          = COALESCE(sqlc.narg('meta')::jsonb, meta),
    updated_at    = NOW()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: DeleteSetpoint :exec
DELETE FROM gr33ncore.zone_setpoints WHERE id = $1;

-- name: GetActiveSetpointForScope :one
-- Precedence resolver: pick the single most-specific setpoint row for
-- (zone, crop_cycle, stage, sensor_type). Rank 1 is most-specific; we
-- ORDER BY rank ASC and LIMIT 1.
--   rank 1 : cycle + stage         (exact cycle + exact stage)
--   rank 2 : cycle + any stage     (NULL stage row on the cycle)
--   rank 3 : zone  + stage         (no cycle row; zone default for stage)
--   rank 4 : zone  + any stage     (zone fallback for every stage)
-- sqlc.narg('crop_cycle_id') may be NULL (rule targets a zone with no
-- active cycle) — in that case rows of rank 1 and 2 naturally drop out.
-- sqlc.narg('stage') may be NULL (zone has no cycle and thus no stage)
-- — rank 1 and 3 drop out, leaving any zone-wide fallback.
SELECT *, (
    CASE
        WHEN crop_cycle_id IS NOT NULL AND stage IS NOT NULL THEN 1
        WHEN crop_cycle_id IS NOT NULL AND stage IS NULL     THEN 2
        WHEN zone_id       IS NOT NULL AND stage IS NOT NULL THEN 3
        WHEN zone_id       IS NOT NULL AND stage IS NULL     THEN 4
        ELSE 99
    END
) AS specificity_rank
FROM gr33ncore.zone_setpoints
WHERE sensor_type = sqlc.arg('sensor_type')::TEXT
  AND (
        (crop_cycle_id = sqlc.narg('crop_cycle_id')::BIGINT AND stage = sqlc.narg('stage')::TEXT)
     OR (crop_cycle_id = sqlc.narg('crop_cycle_id')::BIGINT AND stage IS NULL)
     OR (zone_id       = sqlc.narg('zone_id')::BIGINT       AND stage = sqlc.narg('stage')::TEXT)
     OR (zone_id       = sqlc.narg('zone_id')::BIGINT       AND stage IS NULL)
  )
ORDER BY specificity_rank ASC, updated_at DESC
LIMIT 1;
