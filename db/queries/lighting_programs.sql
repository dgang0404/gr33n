-- ============================================================
-- Queries: gr33ncore.lighting_programs (Phase 35)
-- ============================================================

-- name: ListLightingProgramsByFarm :many
SELECT * FROM gr33ncore.lighting_programs
WHERE farm_id = $1
ORDER BY name ASC;

-- name: GetLightingProgramByID :one
SELECT * FROM gr33ncore.lighting_programs
WHERE id = $1;

-- name: GetLightingProgramZoneBySchedule :one
-- Phase 159 WS1 follow-up — schedule citation via lighting_program ON/OFF pair.
SELECT zone_id
FROM gr33ncore.lighting_programs
WHERE farm_id = sqlc.arg(farm_id)
  AND (schedule_on_id = sqlc.arg(schedule_id) OR schedule_off_id = sqlc.arg(schedule_id))
  AND zone_id IS NOT NULL
ORDER BY is_active DESC, id ASC
LIMIT 1;

-- name: CreateLightingProgram :one
INSERT INTO gr33ncore.lighting_programs (
    farm_id, zone_id, actuator_id, name, description,
    on_hours, off_hours, lights_on_at, timezone,
    schedule_on_id, schedule_off_id, crop_cycle_id,
    is_active, metadata
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
RETURNING *;

-- name: UpdateLightingProgram :one
UPDATE gr33ncore.lighting_programs
SET name = $2, description = $3,
    on_hours = $4, off_hours = $5,
    lights_on_at = $6, timezone = $7,
    schedule_on_id = $8, schedule_off_id = $9,
    crop_cycle_id = $10, is_active = $11,
    metadata = $12, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateLightingProgramActive :one
UPDATE gr33ncore.lighting_programs
SET is_active = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateLightingProgramSchedules :one
UPDATE gr33ncore.lighting_programs
SET schedule_on_id = $2, schedule_off_id = $3, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteLightingProgram :exec
DELETE FROM gr33ncore.lighting_programs WHERE id = $1;
