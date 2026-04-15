-- ============================================================
-- Queries: gr33ncore.farms
-- ============================================================

-- name: CreateFarm :one
INSERT INTO gr33ncore.farms (
    name, description, location_text, size_hectares, farm_type,
    scale_tier, owner_user_id, timezone, currency, operational_status,
    created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
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
    operational_status = $10, updated_by_user_id = $11, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: ListAllFarms :many
SELECT * FROM gr33ncore.farms
WHERE deleted_at IS NULL
ORDER BY name ASC;

-- name: SoftDeleteFarm :exec
UPDATE gr33ncore.farms
SET deleted_at = NOW(), updated_at = NOW(), updated_by_user_id = $2
WHERE id = $1;
