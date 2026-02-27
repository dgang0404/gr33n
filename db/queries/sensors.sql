-- ============================================================
-- Queries: gr33ncore.sensors
-- ============================================================

-- name: CreateSensor :one
INSERT INTO gr33ncore.sensors (
    device_id, farm_id, zone_id, name, sensor_type, unit_id,
    hardware_identifier, value_min_expected, value_max_expected,
    alert_threshold_low, alert_threshold_high, reading_interval_seconds,
    config, meta_data, created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, NOW(), NOW())
RETURNING *;

-- name: GetSensorByID :one
SELECT * FROM gr33ncore.sensors
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListSensorsByFarm :many
SELECT * FROM gr33ncore.sensors
WHERE farm_id = $1 AND deleted_at IS NULL
ORDER BY name ASC;

-- name: ListSensorsByZone :many
SELECT * FROM gr33ncore.sensors
WHERE zone_id = $1 AND deleted_at IS NULL
ORDER BY sensor_type, name ASC;

-- name: ListSensorsByDevice :many
SELECT * FROM gr33ncore.sensors
WHERE device_id = $1 AND deleted_at IS NULL
ORDER BY name ASC;

-- name: SoftDeleteSensor :exec
UPDATE gr33ncore.sensors
SET deleted_at = NOW(), updated_at = NOW(), updated_by_user_id = $2
WHERE id = $1;
