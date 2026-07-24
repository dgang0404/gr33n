-- Phase 211.05 — recipe attribution for harvested crop cycles

-- name: ListHarvestedCyclesForRecipeOutcomes :many
SELECT
    cc.id,
    cc.farm_id,
    cc.zone_id,
    cc.name,
    cc.started_at,
    cc.harvested_at,
    cc.yield_grams,
    p.crop_key,
    p.display_name AS plant_display_name
FROM gr33nfertigation.crop_cycles cc
INNER JOIN gr33ncrops.plants p ON p.id = cc.plant_id AND p.deleted_at IS NULL
WHERE cc.farm_id = sqlc.arg('farm_id')
  AND cc.is_active = FALSE
  AND cc.harvested_at IS NOT NULL
  AND cc.yield_grams IS NOT NULL
  AND cc.yield_grams > 0
  AND p.crop_key IS NOT NULL
  AND TRIM(p.crop_key) <> ''
  AND (sqlc.narg('crop_key')::text IS NULL OR p.crop_key = sqlc.narg('crop_key')::text)
ORDER BY cc.harvested_at DESC, cc.id DESC;

-- name: ListRecipeAttributionHitsForCycle :many
SELECT
    'mix'::text AS source_kind,
    me.id AS source_id,
    me.mixed_at AS occurred_at,
    NULLIF(me.metadata->>'application_recipe_id', '')::bigint AS application_recipe_id,
    NULLIF(me.metadata->>'application_recipe_revision_id', '')::bigint AS application_recipe_revision_id
FROM gr33nfertigation.mixing_events me
LEFT JOIN gr33nfertigation.programs p ON p.id = me.program_id AND p.deleted_at IS NULL
LEFT JOIN gr33nfertigation.reservoirs r ON r.id = me.reservoir_id AND r.deleted_at IS NULL
WHERE me.farm_id = sqlc.arg('farm_id')
  AND me.mixed_at >= sqlc.arg('from_ts')::timestamptz
  AND me.mixed_at <= sqlc.arg('to_ts')::timestamptz
  AND NULLIF(me.metadata->>'application_recipe_id', '') IS NOT NULL
  AND (
    p.target_zone_id = sqlc.arg('zone_id')
    OR r.zone_id = sqlc.arg('zone_id')
  )

UNION ALL

SELECT
    'program_run'::text AS source_kind,
    ar.id AS source_id,
    ar.executed_at AS occurred_at,
    NULLIF(ar.details->>'application_recipe_id', '')::bigint AS application_recipe_id,
    NULLIF(ar.details->>'application_recipe_revision_id', '')::bigint AS application_recipe_revision_id
FROM gr33ncore.automation_runs ar
INNER JOIN gr33nfertigation.programs p ON p.id = ar.program_id AND p.deleted_at IS NULL
WHERE ar.farm_id = sqlc.arg('farm_id')
  AND ar.program_id IS NOT NULL
  AND ar.executed_at >= sqlc.arg('from_ts')::timestamptz
  AND ar.executed_at <= sqlc.arg('to_ts')::timestamptz
  AND p.target_zone_id = sqlc.arg('zone_id')
  AND NULLIF(ar.details->>'application_recipe_id', '') IS NOT NULL

ORDER BY occurred_at ASC, source_id ASC;
