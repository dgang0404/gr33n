-- ============================================================
-- Queries: gr33ncore.units
-- ============================================================

-- name: GetUnitByID :one
SELECT * FROM gr33ncore.units WHERE id = $1;

-- name: GetUnitByName :one
SELECT * FROM gr33ncore.units WHERE name = $1;

-- name: ListUnitsByType :many
SELECT * FROM gr33ncore.units
WHERE unit_type = $1
ORDER BY is_base_unit DESC, name ASC;

-- name: ListAllUnits :many
SELECT * FROM gr33ncore.units
ORDER BY unit_type, is_base_unit DESC, name ASC;

-- name: GetBaseUnitForType :one
SELECT * FROM gr33ncore.units
WHERE unit_type = $1 AND is_base_unit = TRUE
LIMIT 1;
