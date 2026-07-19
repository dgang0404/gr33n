-- name: ListCropCatalogEntries :many
SELECT *
FROM gr33ncrops.crop_catalog_entries
ORDER BY supported DESC, display_name;

-- name: ListCropCatalogAliases :many
SELECT *
FROM gr33ncrops.crop_catalog_aliases
ORDER BY alias;

-- name: GetCropCatalogEntry :one
SELECT *
FROM gr33ncrops.crop_catalog_entries
WHERE crop_key = $1;

-- name: ListAgronomyFieldGuides :many
SELECT *
FROM gr33ncrops.agronomy_field_guides
WHERE published = TRUE
ORDER BY sort_order, slug;

-- name: GetAgronomyFieldGuideBySlug :one
SELECT *
FROM gr33ncrops.agronomy_field_guides
WHERE slug = $1;

-- name: GetPublishedAgronomyFieldGuideBySlug :one
SELECT *
FROM gr33ncrops.agronomy_field_guides
WHERE slug = $1 AND published = TRUE;

-- name: GetBuiltinCropProfileIDByCropKey :one
SELECT id
FROM gr33ncrops.crop_profiles
WHERE farm_id IS NULL AND is_builtin = TRUE AND crop_key = $1;

-- name: CountCropCatalogEntries :one
SELECT COUNT(*)::bigint AS count FROM gr33ncrops.crop_catalog_entries;

-- name: CountSupportedCropCatalogEntries :one
SELECT COUNT(*)::bigint AS count FROM gr33ncrops.crop_catalog_entries WHERE supported = TRUE;

-- name: CountUnsupportedCropCatalogEntries :one
SELECT COUNT(*)::bigint AS count FROM gr33ncrops.crop_catalog_entries WHERE supported = FALSE;

-- name: CountBuiltinCropProfiles :one
SELECT COUNT(*)::bigint AS count FROM gr33ncrops.crop_profiles
WHERE farm_id IS NULL AND is_builtin = TRUE;

-- name: CountAgronomyFieldGuides :one
SELECT COUNT(*)::bigint AS count FROM gr33ncrops.agronomy_field_guides WHERE published = TRUE;
