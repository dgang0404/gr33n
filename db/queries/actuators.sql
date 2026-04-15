-- ============================================================
-- Queries: gr33ncore.actuators + actuator_events
-- ============================================================

-- name: CreateActuator :one
INSERT INTO gr33ncore.actuators (
    device_id, farm_id, zone_id, name, actuator_type,
    hardware_identifier, feedback_sensor_id, config, meta_data, created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
RETURNING *;

-- name: GetActuatorByID :one
SELECT * FROM gr33ncore.actuators
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListActuatorsByFarm :many
SELECT * FROM gr33ncore.actuators
WHERE farm_id = $1 AND deleted_at IS NULL
ORDER BY name ASC;

-- name: UpdateActuatorState :one
UPDATE gr33ncore.actuators
SET current_state_numeric = $2, current_state_text = $3,
    last_known_state_time = NOW(), updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: InsertActuatorEvent :one
INSERT INTO gr33ncore.actuator_events (
    event_time, actuator_id, command_sent, parameters_sent,
    triggered_by_user_id, triggered_by_schedule_id, triggered_by_rule_id,
    source, execution_status, meta_data
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: ListActuatorEventsByActuator :many
SELECT * FROM gr33ncore.actuator_events
WHERE actuator_id = $1
  AND event_time >= $2
ORDER BY event_time DESC
LIMIT $3;

-- name: ListActuatorEventsBySchedule :many
SELECT * FROM gr33ncore.actuator_events
WHERE triggered_by_schedule_id = $1
  AND event_time >= $2
ORDER BY event_time DESC
LIMIT $3;
