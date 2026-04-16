-- ============================================================
-- Commons catalog (gr33n_inserts direction — browse / import audit)
-- ============================================================

-- name: ListPublishedCommonsCatalogEntries :many
SELECT id, slug, title, summary, contributor_display, contributor_uri,
       license_spdx, license_notes, tags, sort_order, created_at, updated_at
FROM gr33ncore.commons_catalog_entries
WHERE published = TRUE
  AND (
    $1::text = ''
    OR title ILIKE '%' || $1 || '%'
    OR summary ILIKE '%' || $1 || '%'
    OR EXISTS (
        SELECT 1 FROM unnest(tags) AS t(tag)
        WHERE t.tag ILIKE '%' || $1 || '%'
    )
  )
ORDER BY sort_order ASC, title ASC
LIMIT $2 OFFSET $3;

-- name: GetPublishedCommonsCatalogEntryBySlug :one
SELECT id, slug, title, summary, body, contributor_display, contributor_uri,
       license_spdx, license_notes, tags, sort_order, created_at, updated_at
FROM gr33ncore.commons_catalog_entries
WHERE published = TRUE AND slug = $1;

-- name: UpsertFarmCommonsCatalogImport :one
INSERT INTO gr33ncore.farm_commons_catalog_imports (
    farm_id, catalog_entry_id, imported_by, note
) VALUES (
    $1, $2, $3, $4
)
ON CONFLICT (farm_id, catalog_entry_id) DO UPDATE SET
    imported_at = NOW(),
    imported_by = EXCLUDED.imported_by,
    note = COALESCE(EXCLUDED.note, farm_commons_catalog_imports.note)
RETURNING id, farm_id, catalog_entry_id, imported_by, imported_at, note;

-- name: ListFarmCommonsCatalogImports :many
SELECT
    i.id,
    i.imported_at,
    i.note,
    e.id AS catalog_entry_id,
    e.slug,
    e.title,
    e.license_spdx,
    e.contributor_display
FROM gr33ncore.farm_commons_catalog_imports i
JOIN gr33ncore.commons_catalog_entries e ON e.id = i.catalog_entry_id
WHERE i.farm_id = $1
ORDER BY i.imported_at DESC;
