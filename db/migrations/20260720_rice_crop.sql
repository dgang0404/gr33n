-- Rice (aquaponics / shallow water) — catalog, profile, field guide.
-- Source: data/crop_library.yaml + docs/field-guides/crop-rice-nutrition.md

INSERT INTO gr33ncrops.crop_catalog_entries (
    crop_key, display_name, supported, category, source, substrate, watering_style,
    runoff_pct_target, moisture_guidance, cousin_of, unsupported_reason, catalog_version
)
SELECT v.crop_key, v.display_name, v.supported, v.category, v.source, v.substrate, v.watering_style,
       v.runoff_pct_target, v.moisture_guidance, v.cousin_of, v.unsupported_reason, v.catalog_version
FROM (VALUES
    ('rice', 'Rice (aquaponics / shallow water)', true, 'grain', 'Warm shallow-water grain; aquaponics raft or paddy tray', 'aquaponics raft / shallow tray', 'constant_feed', NULL, NULL, NULL, NULL, 4)
) AS v(crop_key, display_name, supported, category, source, substrate, watering_style,
         runoff_pct_target, moisture_guidance, cousin_of, unsupported_reason, catalog_version)
ON CONFLICT (crop_key) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    supported = EXCLUDED.supported,
    category = EXCLUDED.category,
    source = EXCLUDED.source,
    substrate = EXCLUDED.substrate,
    watering_style = EXCLUDED.watering_style,
    catalog_version = EXCLUDED.catalog_version,
    updated_at = NOW();

UPDATE gr33ncrops.crop_catalog_entries SET cousin_of = 'lettuce' WHERE crop_key = 'rice';

INSERT INTO gr33ncrops.crop_catalog_aliases (alias, crop_key)
SELECT v.alias, v.crop_key
FROM (VALUES
    ('basmati', 'rice'),
    ('jasmine_rice', 'rice'),
    ('paddy', 'rice')
) AS v(alias, crop_key)
ON CONFLICT (alias) DO UPDATE SET crop_key = EXCLUDED.crop_key;

INSERT INTO gr33ncrops.crop_profiles (farm_id, crop_key, display_name, category, source, version, is_builtin)
SELECT NULL, v.crop_key, v.display_name, v.category, v.source, 4, TRUE
FROM (VALUES
    ('rice', 'Rice (aquaponics / shallow water)', 'grain', 'Warm shallow-water grain; aquaponics raft or paddy tray', 4)
) AS v(crop_key, display_name, category, source, version)
WHERE NOT EXISTS (
    SELECT 1 FROM gr33ncrops.crop_profiles p
    WHERE p.farm_id IS NULL AND p.crop_key = v.crop_key AND p.is_builtin = TRUE
);

INSERT INTO gr33ncrops.crop_profile_stages (
    crop_profile_id, stage, ec_min, ec_target, ec_max, ph_min, ph_max,
    vpd_min_kpa, vpd_max_kpa, temp_min_c, temp_max_c, rh_min_pct, rh_max_pct,
    dli_target, photoperiod_hrs, notes
)
SELECT p.id, s.stage::gr33nfertigation.growth_stage_enum, s.ec_min, s.ec_target, s.ec_max, s.ph_min, s.ph_max,
       s.vpd_min_kpa, s.vpd_max_kpa, s.temp_min_c, s.temp_max_c, s.rh_min_pct, s.rh_max_pct,
       s.dli_target, s.photoperiod_hrs, s.notes
FROM gr33ncrops.crop_profiles p
JOIN (VALUES
    ('rice', 'seedling', 0.40, 0.55, 0.70, 5.50, 6.50, 0.50, 0.80, 22, 30, 60, 85, 18, 14, 'Germination in warm shallow water; keep roots wet, not buried deep'),
    ('rice', 'early_veg', 0.55, 0.75, 0.95, 5.50, 6.50, 0.60, 0.90, 24, 32, 55, 80, 22, 14, 'Tillering; low EC — fish waste often supplies N'),
    ('rice', 'late_veg', 0.70, 0.90, 1.10, 5.50, 6.50, 0.70, 1.00, 24, 34, 50, 75, 28, 14, 'Panicle init — avoid cold nights below ~18 °C'),
    ('rice', 'harvest', 0.50, 0.65, 0.80, 5.50, 6.50, 0.80, 1.10, 22, 32, 45, 70, 30, 12, 'Grain fill; taper EC; shallow dry-down before harvest')
) AS s(crop_key, stage, ec_min, ec_target, ec_max, ph_min, ph_max,
         vpd_min_kpa, vpd_max_kpa, temp_min_c, temp_max_c, rh_min_pct, rh_max_pct,
         dli_target, photoperiod_hrs, notes)
  ON p.farm_id IS NULL AND p.is_builtin = TRUE AND p.crop_key = s.crop_key
WHERE NOT EXISTS (
    SELECT 1 FROM gr33ncrops.crop_profile_stages existing
    WHERE existing.crop_profile_id = p.id AND existing.stage = s.stage::gr33nfertigation.growth_stage_enum
);

INSERT INTO gr33ncrops.agronomy_field_guides (
    slug, title, crop_key, guide_type, audience, safety_tier, body_md, catalog_version, is_published, sort_order
)
SELECT v.slug, v.title, v.crop_key, v.guide_type, v.audience, v.safety_tier, v.body_md, v.catalog_version, v.is_published, v.sort_order
FROM (VALUES
    ('crop-rice-nutrition', 'Rice nutrition (aquaponics / shallow water)', 'rice', 'crop_nutrition', 'general', 'safe', $fg_crop_rice_nutrition$# Rice nutrition (aquaponics / shallow water)

Rice in bench-scale aquaponics or shallow trays runs **low EC** (~0.5–1.0 mS/cm) with **warm roots** and **constant shallow water**. Fish waste often supplies nitrogen — watch for **iron chlorosis** if water is too alkaline.

Assign the rice profile in Start grow or Plants so Guardian cites structured mS/cm targets. Use **Variety / cultivar** for the strain (Basmati, Jasmine, etc.) after picking Rice from the catalog.$fg_crop_rice_nutrition$, 4, TRUE, 56)
) AS v(slug, title, crop_key, guide_type, audience, safety_tier, body_md, catalog_version, is_published, sort_order)
ON CONFLICT (slug) DO UPDATE SET
    title = EXCLUDED.title,
    crop_key = EXCLUDED.crop_key,
    body_md = EXCLUDED.body_md,
    catalog_version = EXCLUDED.catalog_version,
    is_published = EXCLUDED.is_published,
    sort_order = EXCLUDED.sort_order,
    updated_at = NOW();
