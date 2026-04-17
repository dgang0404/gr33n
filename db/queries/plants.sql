-- name: ListPlantsByFarm :many
SELECT * FROM gr33ncrops.plants
WHERE farm_id = $1 AND deleted_at IS NULL
ORDER BY display_name;

-- name: GetPlant :one
SELECT * FROM gr33ncrops.plants WHERE id = $1 AND deleted_at IS NULL;

-- name: CreatePlant :one
INSERT INTO gr33ncrops.plants (farm_id, display_name, variety_or_cultivar, meta)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: UpdatePlant :one
UPDATE gr33ncrops.plants
SET display_name = $2, variety_or_cultivar = $3, meta = $4, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL RETURNING *;

-- name: SoftDeletePlant :exec
UPDATE gr33ncrops.plants SET deleted_at = NOW() WHERE id = $1;
