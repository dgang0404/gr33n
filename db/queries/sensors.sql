-- ============================================================
-- Queries: gr33ncore.sensors
-- ============================================================

-- name: CreateSensor :one
INSERT INTO gr33ncore.sensors (
    device_id, farm_id, zone_id, name, sensor_type, unit_id,
    hardware_identifier, value_min_expected, value_max_expected,
    alert_threshold_low, alert_threshold_high, reading_interval_seconds,
    alert_duration_seconds, alert_cooldown_seconds,
    config, meta_data, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6,
    $7, $8, $9,
    $10, $11, $12,
    $13, $14,
    $15, $16, NOW(), NOW()
)
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

-- name: UpdateSensor :one
-- Patch-style update: each field overwritten when the caller passes a non-NULL value;
-- pass NULL to leave the existing value untouched. alert_breach_started_at is managed
-- by the evaluator and is not editable via this query.
UPDATE gr33ncore.sensors
SET
    zone_id                  = COALESCE(sqlc.narg('zone_id'), zone_id),
    device_id                = COALESCE(sqlc.narg('device_id'), device_id),
    name                     = COALESCE(sqlc.narg('name'), name),
    sensor_type              = COALESCE(sqlc.narg('sensor_type'), sensor_type),
    unit_id                  = COALESCE(sqlc.narg('unit_id'), unit_id),
    hardware_identifier      = COALESCE(sqlc.narg('hardware_identifier'), hardware_identifier),
    value_min_expected       = COALESCE(sqlc.narg('value_min_expected'), value_min_expected),
    value_max_expected       = COALESCE(sqlc.narg('value_max_expected'), value_max_expected),
    alert_threshold_low      = COALESCE(sqlc.narg('alert_threshold_low'), alert_threshold_low),
    alert_threshold_high     = COALESCE(sqlc.narg('alert_threshold_high'), alert_threshold_high),
    reading_interval_seconds = COALESCE(sqlc.narg('reading_interval_seconds'), reading_interval_seconds),
    alert_duration_seconds   = COALESCE(sqlc.narg('alert_duration_seconds'), alert_duration_seconds),
    alert_cooldown_seconds   = COALESCE(sqlc.narg('alert_cooldown_seconds'), alert_cooldown_seconds),
    updated_at               = NOW(),
    updated_by_user_id       = sqlc.narg('updated_by_user_id')
WHERE id = sqlc.arg('id') AND deleted_at IS NULL
RETURNING *;

-- name: SetSensorAlertBreachStart :exec
UPDATE gr33ncore.sensors
SET alert_breach_started_at = $2, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: ClearSensorAlertBreachStart :exec
UPDATE gr33ncore.sensors
SET alert_breach_started_at = NULL, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL AND alert_breach_started_at IS NOT NULL;

-- name: SoftDeleteSensor :exec
UPDATE gr33ncore.sensors
SET deleted_at = NOW(), updated_at = NOW(), updated_by_user_id = $2
WHERE id = $1;
