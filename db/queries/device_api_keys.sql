-- ============================================================
-- Queries: gr33ncore.device_api_keys (Phase 57)
-- ============================================================

-- name: InsertDeviceAPIKey :one
INSERT INTO gr33ncore.device_api_keys (device_id, key_hash, label)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListDeviceAPIKeysByDevice :many
SELECT id, device_id, label, created_at, revoked_at, last_used_at
FROM gr33ncore.device_api_keys
WHERE device_id = $1
ORDER BY created_at DESC;

-- name: ListActiveDeviceAPIKeyHashesByDevice :many
SELECT id, key_hash
FROM gr33ncore.device_api_keys
WHERE device_id = $1 AND revoked_at IS NULL
ORDER BY created_at DESC;

-- name: RevokeDeviceAPIKey :one
UPDATE gr33ncore.device_api_keys
SET revoked_at = NOW()
WHERE id = $1 AND device_id = $2 AND revoked_at IS NULL
RETURNING id, device_id, label, created_at, revoked_at, last_used_at;

-- name: TouchDeviceAPIKeyLastUsed :exec
UPDATE gr33ncore.device_api_keys
SET last_used_at = NOW()
WHERE id = $1 AND revoked_at IS NULL;

-- name: CountActiveDeviceAPIKeysByDevice :one
SELECT COUNT(*)::bigint AS cnt
FROM gr33ncore.device_api_keys
WHERE device_id = $1 AND revoked_at IS NULL;
