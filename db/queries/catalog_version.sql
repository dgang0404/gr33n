-- Phase 109 — platform catalog version bump detection + farm admin notifications.

-- name: GetMaxCropCatalogVersion :one
SELECT COALESCE(MAX(catalog_version), 1)::int AS max_version
FROM gr33ncrops.crop_catalog_entries;

-- name: GetPlatformCatalogState :one
SELECT id, catalog_version, updated_at
FROM gr33ncore.platform_catalog_state
WHERE id = 1;

-- name: UpsertPlatformCatalogState :one
INSERT INTO gr33ncore.platform_catalog_state (id, catalog_version, updated_at)
VALUES (1, $1, NOW())
ON CONFLICT (id) DO UPDATE
SET catalog_version = EXCLUDED.catalog_version,
    updated_at = NOW()
RETURNING id, catalog_version, updated_at;

-- name: GetFarmCatalogVersionSeen :one
SELECT farm_id, catalog_version_seen, notified_at, updated_at
FROM gr33ncore.farm_catalog_version_seen
WHERE farm_id = $1;

-- name: UpsertFarmCatalogVersionSeen :one
INSERT INTO gr33ncore.farm_catalog_version_seen (farm_id, catalog_version_seen, notified_at, updated_at)
VALUES ($1, $2, $3, NOW())
ON CONFLICT (farm_id) DO UPDATE
SET catalog_version_seen = EXCLUDED.catalog_version_seen,
    notified_at = EXCLUDED.notified_at,
    updated_at = NOW()
RETURNING farm_id, catalog_version_seen, notified_at, updated_at;

-- name: ListAllFarmIDs :many
SELECT id FROM gr33ncore.farms WHERE deleted_at IS NULL ORDER BY id;

-- name: ListFarmCatalogNotifyAdminUserIDs :many
SELECT DISTINCT m.user_id
FROM gr33ncore.farm_memberships m
WHERE m.farm_id = $1
  AND m.role_in_farm IN ('owner', 'manager')
UNION
SELECT f.owner_user_id
FROM gr33ncore.farms f
WHERE f.id = $1 AND f.deleted_at IS NULL;

-- name: GetCatalogVersionBumpAlertForFarm :one
SELECT id FROM gr33ncore.alerts_notifications
WHERE farm_id = $1
  AND triggering_event_source_type = 'catalog_version_bump'
  AND triggering_event_source_id = $2
LIMIT 1;
