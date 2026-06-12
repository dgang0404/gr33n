-- name: ListCropProfilesForFarm :many
SELECT *
FROM gr33ncrops.crop_profiles
WHERE farm_id IS NULL AND is_builtin = TRUE
   OR farm_id = sqlc.arg(farm_id)
ORDER BY is_builtin ASC, display_name;

-- name: GetCropProfile :one
SELECT * FROM gr33ncrops.crop_profiles WHERE id = $1;

-- name: GetCropProfileByKey :one
SELECT *
FROM gr33ncrops.crop_profiles
WHERE crop_key = $1
  AND (farm_id IS NULL AND is_builtin = TRUE OR farm_id = sqlc.arg(farm_id))
ORDER BY is_builtin ASC
LIMIT 1;

-- name: GetBuiltinCropProfileByKey :one
SELECT *
FROM gr33ncrops.crop_profiles
WHERE crop_key = $1 AND farm_id IS NULL AND is_builtin = TRUE;

-- name: DeleteFarmCropProfileByKey :exec
DELETE FROM gr33ncrops.crop_profiles
WHERE farm_id = $1 AND crop_key = $2 AND is_builtin = FALSE;

-- name: ListCropProfileStages :many
SELECT *
FROM gr33ncrops.crop_profile_stages
WHERE crop_profile_id = $1
ORDER BY stage;

-- name: GetCropProfileStage :one
SELECT s.*
FROM gr33ncrops.crop_profile_stages s
JOIN gr33ncrops.crop_profiles p ON p.id = s.crop_profile_id
WHERE s.crop_profile_id = $1 AND s.stage = $2;

-- name: CreateCropProfile :one
INSERT INTO gr33ncrops.crop_profiles (
    farm_id, crop_key, display_name, category, source, version, is_builtin, meta
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: CreateCropProfileStage :one
INSERT INTO gr33ncrops.crop_profile_stages (
    crop_profile_id, stage, ec_min, ec_target, ec_max, ph_min, ph_max,
    vpd_min_kpa, vpd_max_kpa, temp_min_c, temp_max_c, rh_min_pct, rh_max_pct,
    dli_target, photoperiod_hrs, notes
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
) RETURNING *;
