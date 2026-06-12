-- Phase 103 WS1 — read-only legacy plant audit (pre/post merge).
\set QUIET on
\pset format aligned
\pset tuples_only off

\echo '==> Active plants missing crop_key (need manual catalog pick or alias)'
SELECT pl.id, pl.farm_id, pl.display_name, pl.variety_or_cultivar, pl.crop_profile_id
FROM gr33ncrops.plants pl
WHERE pl.deleted_at IS NULL
  AND pl.crop_key IS NULL
ORDER BY pl.farm_id, pl.id;

\echo ''
\echo '==> Duplicate crop_key slots per farm (should be 0 after merge)'
SELECT pl.farm_id, pl.crop_key, count(*) AS plant_rows, array_agg(pl.id ORDER BY pl.id) AS plant_ids
FROM gr33ncrops.plants pl
WHERE pl.deleted_at IS NULL
  AND pl.crop_key IS NOT NULL
GROUP BY pl.farm_id, pl.crop_key
HAVING count(*) > 1
ORDER BY pl.farm_id, pl.crop_key;

\echo ''
\echo '==> Active cycles with plant_id pointing at missing/deleted plant'
SELECT c.id AS cycle_id, c.farm_id, c.plant_id, c.name, c.is_active
FROM gr33nfertigation.crop_cycles c
LEFT JOIN gr33ncrops.plants p ON p.id = c.plant_id AND p.deleted_at IS NULL
WHERE c.plant_id IS NOT NULL
  AND p.id IS NULL
ORDER BY c.farm_id, c.id;

\echo ''
\echo '==> Active cycles missing plant_id (informational)'
SELECT c.id, c.farm_id, c.zone_id, c.name, c.is_active
FROM gr33nfertigation.crop_cycles c
WHERE c.is_active = TRUE
  AND (c.plant_id IS NULL OR c.plant_id <= 0)
ORDER BY c.farm_id, c.id;

\echo ''
\echo '==> Summary counts'
SELECT
  (SELECT count(*) FROM gr33ncrops.plants WHERE deleted_at IS NULL) AS active_plants,
  (SELECT count(*) FROM gr33ncrops.plants WHERE deleted_at IS NULL AND crop_key IS NULL) AS missing_crop_key,
  (SELECT count(*) FROM (
      SELECT 1 FROM gr33ncrops.plants
      WHERE deleted_at IS NULL AND crop_key IS NOT NULL
      GROUP BY farm_id, crop_key HAVING count(*) > 1
  ) d) AS duplicate_crop_key_groups;
