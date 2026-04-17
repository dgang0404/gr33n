-- ============================================================
-- Phase 20.8 WS2 — gr33naquaponics.loops queries
-- ============================================================

-- name: ListAquaponicsLoopsByFarm :many
SELECT * FROM gr33naquaponics.loops
WHERE farm_id = $1 AND deleted_at IS NULL
ORDER BY active DESC, label ASC, id ASC;

-- name: GetAquaponicsLoopByID :one
SELECT * FROM gr33naquaponics.loops
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateAquaponicsLoop :one
INSERT INTO gr33naquaponics.loops (
    farm_id, label, fish_tank_zone_id, grow_bed_zone_id, meta
) VALUES (
    sqlc.arg(farm_id),
    sqlc.arg(label),
    sqlc.narg(fish_tank_zone_id),
    sqlc.narg(grow_bed_zone_id),
    COALESCE(sqlc.narg(meta)::jsonb, '{}'::jsonb)
)
RETURNING *;

-- name: UpdateAquaponicsLoop :one
UPDATE gr33naquaponics.loops SET
    label             = sqlc.arg(label),
    fish_tank_zone_id = sqlc.narg(fish_tank_zone_id),
    grow_bed_zone_id  = sqlc.narg(grow_bed_zone_id),
    active            = sqlc.arg(active),
    meta              = COALESCE(sqlc.narg(meta)::jsonb, meta),
    updated_at        = NOW()
WHERE id = sqlc.arg(id) AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteAquaponicsLoop :exec
UPDATE gr33naquaponics.loops
SET deleted_at = NOW()
WHERE id = $1;
