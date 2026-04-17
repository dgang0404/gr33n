-- ============================================================
-- Queries: gr33ncore.sensor_readings (TimescaleDB hypertable)
-- ============================================================

-- name: InsertSensorReading :one
INSERT INTO gr33ncore.sensor_readings (
    reading_time, sensor_id, value_raw, value_text, value_json,
    battery_level_percent, signal_strength_dbm, is_valid, meta_data
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetLatestReadingBySensor :one
SELECT * FROM gr33ncore.sensor_readings
WHERE sensor_id = $1
ORDER BY reading_time DESC
LIMIT 1;

-- name: GetLatestReadingForZoneSensorType :one
-- Phase 20.6 WS3 — setpoint-typed predicates key off `sensor_type`
-- (e.g. "dew_point") rather than a specific sensor_id, so the evaluator
-- has to pick the freshest reading across every sensor of that type in
-- the zone. `gr33ncore.sensors` is small relative to sensor_readings so
-- the JOIN is cheap; the ORDER BY reading_time DESC LIMIT 1 uses the
-- existing per-sensor reading_time index.
SELECT sr.*
FROM gr33ncore.sensor_readings sr
JOIN gr33ncore.sensors s ON s.id = sr.sensor_id
WHERE s.zone_id = $1
  AND s.sensor_type = $2
  AND s.deleted_at IS NULL
ORDER BY sr.reading_time DESC
LIMIT 1;

-- name: ListReadingsBySensorAndTimeRange :many
SELECT * FROM gr33ncore.sensor_readings
WHERE sensor_id = $1
  AND reading_time >= $2
  AND reading_time <= $3
ORDER BY reading_time DESC;

-- name: ListLatestReadingsByFarm :many
SELECT DISTINCT ON (s.id)
    sr.reading_time, sr.sensor_id, sr.value_raw, sr.value_normalized,
    sr.normalized_unit_id, sr.is_valid, s.name as sensor_name,
    s.sensor_type, s.zone_id
FROM gr33ncore.sensor_readings sr
JOIN gr33ncore.sensors s ON s.id = sr.sensor_id
WHERE s.farm_id = $1 AND s.deleted_at IS NULL
ORDER BY s.id, sr.reading_time DESC;

-- name: GetSensorReadingStats :one
SELECT
    COUNT(*)                    AS total_readings,
    COALESCE(AVG(value_raw), 0) AS avg_value,
    MIN(value_raw)              AS min_value,
    MAX(value_raw)              AS max_value,
    MIN(reading_time)           AS first_reading,
    MAX(reading_time)           AS last_reading
FROM gr33ncore.sensor_readings
WHERE sensor_id = $1
  AND reading_time >= $2
  AND reading_time <= $3;
