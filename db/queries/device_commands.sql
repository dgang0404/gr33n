-- ============================================================
-- Queries: gr33ncore.device_commands (Phase 39 WS1)
-- ============================================================

-- name: EnqueueDeviceCommand :one
INSERT INTO gr33ncore.device_commands (
    device_id, farm_id, command_type, payload,
    source, actuator_id, schedule_id, rule_id, program_id
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetNextDeviceCommand :one
-- Returns the oldest pending command for a device and marks it in_progress atomically.
-- Pi-key path: Pi calls this, gets one command, executes, then calls AckDeviceCommand.
UPDATE gr33ncore.device_commands
SET status = 'in_progress', started_at = NOW()
WHERE id = (
    SELECT id FROM gr33ncore.device_commands
    WHERE device_id = $1 AND status = 'pending'
    ORDER BY created_at ASC
    LIMIT 1
    FOR UPDATE SKIP LOCKED
)
RETURNING *;

-- name: AckDeviceCommand :one
-- Pi-key: marks a command completed or failed and records the result payload.
UPDATE gr33ncore.device_commands
SET status = $2, completed_at = NOW(), result = $3
WHERE id = $1
RETURNING *;

-- name: ListDeviceCommands :many
-- Operator JWT: list commands for a device with optional status filter.
-- Pass NULL for status to get all.
SELECT * FROM gr33ncore.device_commands
WHERE device_id = $1
  AND ($2::text IS NULL OR status = $2)
ORDER BY created_at DESC
LIMIT 100;

-- name: ListDeviceCommandsByFarm :many
-- List all commands for a farm for dashboard / ops view.
SELECT * FROM gr33ncore.device_commands
WHERE farm_id = $1
  AND ($2::text IS NULL OR status = $2)
ORDER BY created_at DESC
LIMIT 100;

-- name: CancelDeviceCommand :one
-- Cancel a pending command (operator or worker safety valve).
UPDATE gr33ncore.device_commands
SET status = 'cancelled', completed_at = NOW()
WHERE id = $1 AND status = 'pending'
RETURNING *;

-- name: CountPendingCommandsByDevice :one
SELECT COUNT(*)::bigint AS cnt
FROM gr33ncore.device_commands
WHERE device_id = $1 AND status IN ('pending', 'in_progress');

-- name: CountPendingCommandsByFarm :many
-- Aggregate queue depth per device for a farm (Dashboard chip, Phase 41 WS1).
SELECT device_id, COUNT(*)::bigint AS cnt
FROM gr33ncore.device_commands
WHERE farm_id = $1 AND status IN ('pending', 'in_progress')
GROUP BY device_id
ORDER BY device_id;
