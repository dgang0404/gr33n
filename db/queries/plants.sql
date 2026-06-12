-- name: ListPlantsByFarm :many
SELECT * FROM gr33ncrops.plants
WHERE farm_id = $1 AND deleted_at IS NULL
ORDER BY display_name;

-- name: GetPlant :one
SELECT * FROM gr33ncrops.plants WHERE id = $1 AND deleted_at IS NULL;

-- name: GetPlantByFarmCropKey :one
SELECT * FROM gr33ncrops.plants
WHERE farm_id = $1 AND crop_key = $2 AND deleted_at IS NULL;

-- name: CreatePlant :one
INSERT INTO gr33ncrops.plants (farm_id, display_name, variety_or_cultivar, crop_profile_id, crop_key, meta)
VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: UpdatePlant :one
UPDATE gr33ncrops.plants
SET display_name = $2, variety_or_cultivar = $3, crop_profile_id = $4, crop_key = $5, meta = $6, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL RETURNING *;

-- name: UpdatePlantVariety :one
UPDATE gr33ncrops.plants
SET variety_or_cultivar = $2, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL RETURNING *;

-- name: GetPlantCropProfileStage :one
SELECT s.*, p.display_name AS profile_display_name, p.crop_key
FROM gr33ncrops.crop_profile_stages s
JOIN gr33ncrops.crop_profiles p ON p.id = s.crop_profile_id
JOIN gr33ncrops.plants pl ON pl.crop_profile_id = p.id
WHERE pl.id = $1 AND s.stage = $2 AND pl.deleted_at IS NULL;

-- name: SoftDeletePlant :exec
UPDATE gr33ncrops.plants SET deleted_at = NOW() WHERE id = $1;
