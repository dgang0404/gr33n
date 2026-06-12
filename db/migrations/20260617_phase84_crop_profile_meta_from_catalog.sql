-- Phase 84 WS-I — copy substrate/watering metadata from crop_catalog_entries into builtin crop_profiles.meta.

UPDATE gr33ncrops.crop_profiles p
SET meta = jsonb_strip_nulls(COALESCE(p.meta, '{}'::jsonb) || jsonb_build_object(
    'substrate', NULLIF(BTRIM(c.substrate), ''),
    'watering_style', NULLIF(BTRIM(c.watering_style), ''),
    'runoff_pct_target', NULLIF(BTRIM(c.runoff_pct_target), ''),
    'moisture_guidance', NULLIF(BTRIM(c.moisture_guidance), ''),
    'catalog_version', c.catalog_version
))
FROM gr33ncrops.crop_catalog_entries c
WHERE p.farm_id IS NULL
  AND p.is_builtin = TRUE
  AND p.crop_key = c.crop_key
  AND c.supported = TRUE;
