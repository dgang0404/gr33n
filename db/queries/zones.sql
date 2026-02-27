-- ============================================================
-- Queries: gr33ncore.zones
-- ============================================================

-- name: CreateZone :one
INSERT INTO gr33ncore.zones (
    farm_id, parent_zone_id, name, description, zone_type,
    area_sqm, meta_data, created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
RETURNING *;

-- name: GetZoneByID :one
SELECT * FROM gr33ncore.zones
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListZonesByFarm :many
SELECT * FROM gr33ncore.zones
WHERE farm_id = $1 AND deleted_at IS NULL
ORDER BY name ASC;

-- name: ListZonesByParent :many
SELECT * FROM gr33ncore.zones
WHERE parent_zone_id = $1 AND deleted_at IS NULL
ORDER BY name ASC;

-- name: UpdateZone :one
UPDATE gr33ncore.zones
SET name = $2, description = $3, zone_type = $4, area_sqm = $5,
    meta_data = $6, updated_by_user_id = $7, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteZone :exec
UPDATE gr33ncore.zones
SET deleted_at = NOW(), updated_at = NOW(), updated_by_user_id = $2
WHERE id = $1;
