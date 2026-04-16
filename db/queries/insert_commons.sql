-- ============================================================
-- Queries: gr33ncore.insert_commons_sync_events
-- ============================================================

-- name: GetInsertCommonsSyncEventByFarmIdempotencyKey :one
SELECT *
FROM gr33ncore.insert_commons_sync_events
WHERE farm_id = $1 AND idempotency_key = $2
LIMIT 1;

-- name: CountInsertCommonsSyncAttemptsSince :one
SELECT COUNT(*)::bigint AS cnt
FROM gr33ncore.insert_commons_sync_events
WHERE farm_id = $1 AND created_at >= $2;

-- name: ListInsertCommonsSyncEventsByFarm :many
SELECT id, farm_id, idempotency_key, status, http_status, error, bundle_id, created_at
FROM gr33ncore.insert_commons_sync_events
WHERE farm_id = $1
ORDER BY created_at DESC, id DESC
LIMIT $2 OFFSET $3;

-- name: UpsertInsertCommonsSyncEvent :one
INSERT INTO gr33ncore.insert_commons_sync_events (
    farm_id, idempotency_key, status, http_status, error, payload, bundle_id
) VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (farm_id, idempotency_key)
WHERE idempotency_key IS NOT NULL
DO UPDATE SET
    status = EXCLUDED.status,
    http_status = EXCLUDED.http_status,
    error = EXCLUDED.error,
    payload = EXCLUDED.payload,
    bundle_id = EXCLUDED.bundle_id,
    created_at = NOW()
RETURNING *;
