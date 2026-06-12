-- name: GetGeneticsProfileLink :one
SELECT *
FROM gr33ncrops.plant_genetics_profiles
WHERE farm_id = $1 AND crop_key = $2 AND variety_slug = $3;

-- name: InsertGeneticsProfileLink :one
INSERT INTO gr33ncrops.plant_genetics_profiles (
    farm_id, crop_key, variety_slug, variety_label, crop_profile_id
) VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: DeleteGeneticsProfileLink :exec
DELETE FROM gr33ncrops.plant_genetics_profiles
WHERE farm_id = $1 AND crop_key = $2 AND variety_slug = $3;

-- name: DeleteCropProfileByID :exec
DELETE FROM gr33ncrops.crop_profiles WHERE id = $1;
