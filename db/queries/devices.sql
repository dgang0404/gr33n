-- ============================================================
-- Queries: gr33ncore.devices
-- ============================================================

-- name: CreateDevice :one
INSERT INTO gr33ncore.devices (
    farm_id, zone_id, name, device_uid, device_type,
    ip_address, firmware_version, status, config, meta_data, created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
RETURNING *;

-- name: GetDeviceByID :one
SELECT * FROM gr33ncore.devices
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetDeviceByUID :one
SELECT * FROM gr33ncore.devices
WHERE device_uid = $1 AND deleted_at IS NULL;

-- name: ListDevicesByFarm :many
SELECT * FROM gr33ncore.devices
WHERE farm_id = $1 AND deleted_at IS NULL
ORDER BY name ASC;

-- name: ListDevicesByZone :many
SELECT * FROM gr33ncore.devices
WHERE zone_id = $1 AND deleted_at IS NULL
ORDER BY name ASC;

-- name: UpdateDeviceStatus :one
UPDATE gr33ncore.devices
SET status = $2, last_heartbeat = NOW(), updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: SetDevicePendingCommand :exec
UPDATE gr33ncore.devices
SET config = jsonb_set(
      coalesce(config, '{}'),
      '{pending_command}', $2::jsonb
    ), updated_at = NOW()
WHERE id = $1;

-- name: ClearDevicePendingCommand :exec
UPDATE gr33ncore.devices
SET config = config - 'pending_command', updated_at = NOW()
WHERE id = $1;

-- name: SoftDeleteDevice :exec
UPDATE gr33ncore.devices
SET deleted_at = NOW(), updated_at = NOW(), updated_by_user_id = $2
WHERE id = $1;
