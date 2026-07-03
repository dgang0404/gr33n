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
SET status = $2,
    last_heartbeat = NOW(),
    updated_at = NOW(),
    config = coalesce(config, '{}'::jsonb)
      || CASE WHEN $3::text IS NOT NULL AND $3::text <> ''
         THEN jsonb_build_object('last_config_fetch_at', $3::text) ELSE '{}'::jsonb END
      || CASE WHEN $4::text IS NOT NULL AND $4::text <> ''
         THEN jsonb_build_object('config_sha256', $4::text) ELSE '{}'::jsonb END
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateDeviceStatusTelemetry :one
-- Pi-key heartbeat: status + optional config fetch timestamp, firmware/client version, uptime.
UPDATE gr33ncore.devices
SET status = $2,
    last_heartbeat = NOW(),
    updated_at = NOW(),
    firmware_version = COALESCE(NULLIF($4::text, ''), firmware_version),
    config = coalesce(config, '{}'::jsonb)
      || CASE WHEN $3::text IS NOT NULL AND $3::text <> ''
         THEN jsonb_build_object('last_config_fetch_at', $3::text) ELSE '{}'::jsonb END
      || CASE WHEN $5::text IS NOT NULL AND $5::text <> ''
         THEN jsonb_build_object('client_version', $5::text) ELSE '{}'::jsonb END
      || CASE WHEN $6::bigint >= 0
         THEN jsonb_build_object('client_uptime_seconds', $6::bigint) ELSE '{}'::jsonb END
      || CASE WHEN $7::text IS NOT NULL AND $7::text <> ''
         THEN jsonb_build_object('config_sha256', $7::text) ELSE '{}'::jsonb END
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: MarkStaleDevicesOffline :many
UPDATE gr33ncore.devices
SET status = 'offline', updated_at = NOW()
WHERE status = 'online'
  AND deleted_at IS NULL
  AND (
    last_heartbeat IS NULL
    OR last_heartbeat < NOW() - ($1::bigint * INTERVAL '1 second')
  )
RETURNING id, farm_id, name, device_uid;

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

-- name: CountDevicesByStatusForFarm :many
SELECT status, COUNT(*)::bigint AS cnt
FROM gr33ncore.devices
WHERE farm_id = $1 AND deleted_at IS NULL
GROUP BY status
ORDER BY status ASC;

-- name: BumpDeviceConfigVersion :one
UPDATE gr33ncore.devices
SET config_version = config_version + 1,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;
