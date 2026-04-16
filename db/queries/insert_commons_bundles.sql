-- ============================================================
-- Queries: gr33ncore.insert_commons_bundles (approval + export)
-- ============================================================

-- name: InsertInsertCommonsBundle :one
INSERT INTO gr33ncore.insert_commons_bundles (
    farm_id, idempotency_key, payload_hash, payload, status
) VALUES ($1, $2, $3, $4, 'pending_approval')
RETURNING *;

-- name: GetInsertCommonsBundleByID :one
SELECT * FROM gr33ncore.insert_commons_bundles
WHERE id = $1;

-- name: GetInsertCommonsBundlePendingByFarmIdempotencyKey :one
SELECT * FROM gr33ncore.insert_commons_bundles
WHERE farm_id = $1
  AND idempotency_key = $2
  AND status = 'pending_approval';

-- name: ListInsertCommonsBundlesByFarm :many
SELECT * FROM gr33ncore.insert_commons_bundles
WHERE farm_id = $1
  AND ($2 = '' OR status = $2)
ORDER BY created_at DESC, id DESC
LIMIT $3 OFFSET $4;

-- name: ApproveInsertCommonsBundle :one
UPDATE gr33ncore.insert_commons_bundles
SET status = 'approved',
    reviewer_user_id = $2,
    reviewed_at = NOW(),
    review_note = $3,
    updated_at = NOW()
WHERE id = $1
  AND farm_id = $4
  AND status = 'pending_approval'
RETURNING *;

-- name: RejectInsertCommonsBundle :one
UPDATE gr33ncore.insert_commons_bundles
SET status = 'rejected',
    reviewer_user_id = $2,
    reviewed_at = NOW(),
    review_note = $3,
    updated_at = NOW()
WHERE id = $1
  AND farm_id = $4
  AND status = 'pending_approval'
RETURNING *;

-- name: MarkInsertCommonsBundleDelivered :one
UPDATE gr33ncore.insert_commons_bundles
SET status = 'delivered',
    delivery_http_status = $2,
    delivery_error = NULL,
    updated_at = NOW()
WHERE id = $1 AND farm_id = $3
RETURNING *;

-- name: MarkInsertCommonsBundleDeliveryFailed :one
UPDATE gr33ncore.insert_commons_bundles
SET status = 'delivery_failed',
    delivery_http_status = $2,
    delivery_error = $3,
    updated_at = NOW()
WHERE id = $1 AND farm_id = $4
RETURNING *;
