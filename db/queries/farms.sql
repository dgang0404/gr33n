-- ============================================================
-- Queries: gr33ncore.farms
-- ============================================================

-- name: CreateFarm :one
INSERT INTO gr33ncore.farms (
    name, description, location_text, size_hectares, farm_type,
    scale_tier, owner_user_id, timezone, currency, operational_status,
    organization_id,
    created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
RETURNING *;

-- name: GetFarmByID :one
SELECT * FROM gr33ncore.farms
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListFarmsByOwner :many
SELECT * FROM gr33ncore.farms
WHERE owner_user_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: ListFarmsForUser :many
SELECT f.*
FROM gr33ncore.farms f
JOIN gr33ncore.farm_memberships m ON m.farm_id = f.id
WHERE m.user_id = $1 AND f.deleted_at IS NULL
ORDER BY f.name ASC;

-- name: UpdateFarm :one
UPDATE gr33ncore.farms
SET name = $2, description = $3, location_text = $4, size_hectares = $5,
    farm_type = $6, scale_tier = $7, timezone = $8, currency = $9,
    operational_status = $10, updated_by_user_id = $11,
    organization_id = $12,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: ListAllFarms :many
SELECT * FROM gr33ncore.farms
WHERE deleted_at IS NULL
ORDER BY name ASC;

-- name: UserHasFarmAccess :one
SELECT (
 EXISTS (
        SELECT 1 FROM gr33ncore.farm_memberships m
        WHERE m.farm_id = $1 AND m.user_id = $2
    )
    OR EXISTS (
        SELECT 1 FROM gr33ncore.farms f
        WHERE f.id = $1 AND f.owner_user_id = $2 AND f.deleted_at IS NULL
    )
) AS user_has_farm_access;

-- name: SoftDeleteFarm :exec
UPDATE gr33ncore.farms
SET deleted_at = NOW(), updated_at = NOW(), updated_by_user_id = $2
WHERE id = $1;

-- name: SetFarmOrganization :one
UPDATE gr33ncore.farms
SET organization_id = $2,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SetFarmInsertCommonsOptIn :one
UPDATE gr33ncore.farms
SET insert_commons_opt_in = $2,
    insert_commons_backoff_until = CASE WHEN $2 THEN insert_commons_backoff_until ELSE NULL END,
    insert_commons_consecutive_failures = CASE WHEN $2 THEN insert_commons_consecutive_failures ELSE 0 END,
    insert_commons_last_error = CASE WHEN $2 THEN insert_commons_last_error ELSE NULL END,
    insert_commons_last_delivery_status = CASE WHEN $2 THEN insert_commons_last_delivery_status ELSE NULL END,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: MarkFarmInsertCommonsAttempt :one
UPDATE gr33ncore.farms
SET insert_commons_last_attempt_at = NOW(), updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL AND insert_commons_opt_in = TRUE
RETURNING *;

-- name: MarkFarmInsertCommonsDelivered :one
UPDATE gr33ncore.farms
SET insert_commons_last_sync_at = NOW(),
    insert_commons_last_attempt_at = NOW(),
    insert_commons_last_delivery_status = $2,
    insert_commons_last_error = NULL,
    insert_commons_backoff_until = NULL,
    insert_commons_consecutive_failures = 0,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL AND insert_commons_opt_in = TRUE
RETURNING *;

-- name: MarkFarmInsertCommonsSkippedReceiver :one
UPDATE gr33ncore.farms
SET insert_commons_last_attempt_at = NOW(),
    insert_commons_last_delivery_status = $2,
    insert_commons_last_error = NULL,
    insert_commons_backoff_until = NULL,
    insert_commons_consecutive_failures = 0,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL AND insert_commons_opt_in = TRUE
RETURNING *;

-- name: MarkFarmInsertCommonsSyncFailure :one
UPDATE gr33ncore.farms
SET insert_commons_last_attempt_at = NOW(),
    insert_commons_last_delivery_status = $2,
    insert_commons_last_error = $3,
    insert_commons_backoff_until = $4,
    insert_commons_consecutive_failures = insert_commons_consecutive_failures + 1,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL AND insert_commons_opt_in = TRUE
RETURNING *;
