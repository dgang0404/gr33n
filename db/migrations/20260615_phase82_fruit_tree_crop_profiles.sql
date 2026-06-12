-- Phase 82 WS4h — Fruit tree crop profiles (10 crops). crop_library.yaml v4 (46 crops).
-- Source: data/crop_library.yaml

INSERT INTO gr33ncrops.crop_profiles (farm_id, crop_key, display_name, category, source, version, is_builtin)
SELECT NULL, v.crop_key, v.display_name, v.category, v.source, v.version, TRUE
FROM (VALUES
    ('apple', 'Apple (nursery / young tree)', 'fruit_tree', 'Container or greenhouse nursery — multi-year to bearing; not full orchard automation', 4),
    ('citrus', 'Citrus (lemon / orange nursery)', 'fruit_tree', 'Warm greenhouse citrus nursery; container citrus production', 4),
    ('fig', 'Fig (container)', 'fruit_tree', 'Warm container fig; breba/main crop cultivar-dependent', 4),
    ('peach', 'Peach / nectarine (nursery)', 'fruit_tree', 'Deciduous stone fruit nursery; chill hours required', 4),
    ('cherry', 'Cherry (nursery / sweet)', 'fruit_tree', 'Sweet cherry nursery; high light; chill dependent', 4),
    ('grape', 'Grape (vine / nursery)', 'fruit_tree', 'Container or greenhouse vine nursery; trellis from year 1', 4),
    ('avocado', 'Avocado (container nursery)', 'fruit_tree', 'Sensitive roots; long juvenile phase; warm greenhouse', 4),
    ('pear', 'Pear (nursery)', 'fruit_tree', 'Deciduous nursery pear; similar to apple bench culture', 4),
    ('plum', 'Plum (nursery / stone fruit)', 'fruit_tree', 'Stone fruit nursery; similar to peach', 4),
    ('mango', 'Mango (container nursery)', 'fruit_tree', 'Tropical greenhouse nursery; warm only', 4)
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
    ('apple', 'seedling', 0.80, 1, 1.20, 5.50, 6.50, 0.60, 0.90, 15, 24, 55, 70, 15, 16, 'Bench graft / liner'),
    ('apple', 'early_veg', 1, 1.40, 1.70, 5.50, 6.50, 0.80, 1.10, 15, 26, 50, 65, 25, 16, 'Year 1–2 container; winter chill cultivar-dependent'),
    ('apple', 'late_veg', 1.20, 1.60, 2, 5.50, 6.50, 0.90, 1.20, 15, 26, 45, 60, 30, 16, 'Young tree training — fruiting years 3–5+'),
    ('citrus', 'seedling', 0.80, 1, 1.20, 5.50, 6.20, 0.60, 0.90, 20, 28, 55, 70, 18, 14, 'Warm bench; avoid cold root zone'),
    ('citrus', 'early_veg', 1, 1.40, 1.70, 5.50, 6.20, 0.80, 1.10, 20, 30, 50, 65, 25, 14, 'Container citrus; iron at low pH'),
    ('citrus', 'late_veg', 1.20, 1.60, 2, 5.50, 6.20, 0.90, 1.20, 18, 30, 45, 60, 30, 14, 'Young tree — fruit set year 2–4 in pots'),
    ('fig', 'seedling', 0.70, 0.90, 1.10, 5.50, 6.50, 0.60, 0.90, 18, 28, 50, 65, 18, 14, 'Warm germination'),
    ('fig', 'early_veg', 1, 1.30, 1.60, 5.50, 6.50, 0.80, 1.10, 20, 30, 45, 60, 25, 14, 'Container vigor management'),
    ('fig', 'late_veg', 1.20, 1.50, 1.90, 5.50, 6.50, 0.90, 1.20, 18, 30, 40, 55, 30, 14, 'Reduce feed pre-dormancy if applicable'),
    ('peach', 'seedling', 0.80, 1, 1.20, 5.50, 6.50, 0.60, 0.90, 12, 24, 55, 70, 18, 16, 'Graft / liner'),
    ('peach', 'early_veg', 1, 1.40, 1.70, 5.50, 6.50, 0.80, 1.10, 15, 28, 45, 60, 28, 16, 'Spring flush — monitor bacterial spot risk'),
    ('peach', 'late_veg', 1.20, 1.60, 2, 5.50, 6.50, 0.90, 1.20, 15, 28, 40, 55, 32, 16, 'Young tree — stone fruit years 2–4'),
    ('cherry', 'seedling', 0.80, 1, 1.20, 5.50, 6.50, 0.60, 0.90, 12, 22, 55, 70, 18, 16, 'Cool-start bench'),
    ('cherry', 'early_veg', 1, 1.30, 1.60, 5.50, 6.50, 0.80, 1.10, 15, 26, 45, 60, 28, 16, 'High light for quality wood'),
    ('cherry', 'late_veg', 1.10, 1.50, 1.80, 5.50, 6.50, 0.90, 1.20, 15, 26, 40, 55, 32, 16, 'Bird / crack risk at fruit — outdoor scale caveat'),
    ('grape', 'seedling', 0.70, 0.90, 1.10, 5.50, 6.50, 0.60, 0.90, 15, 26, 55, 70, 18, 16, 'Bench graft'),
    ('grape', 'early_veg', 1, 1.40, 1.70, 5.50, 6.50, 0.80, 1.10, 18, 28, 45, 60, 28, 16, 'Cane training'),
    ('grape', 'late_veg', 1.20, 1.60, 2, 5.50, 6.50, 0.90, 1.20, 18, 28, 40, 55, 32, 16, 'First crop year 2–3 in greenhouse vine'),
    ('avocado', 'seedling', 0.60, 0.80, 1, 5.50, 6.50, 0.60, 0.90, 20, 28, 55, 70, 15, 14, 'Never waterlogged — root rot'),
    ('avocado', 'early_veg', 0.90, 1.20, 1.50, 5.50, 6.50, 0.80, 1.10, 20, 30, 50, 65, 25, 14, 'Graft management; chloride-sensitive'),
    ('avocado', 'late_veg', 1.10, 1.40, 1.80, 5.50, 6.50, 0.90, 1.20, 18, 30, 45, 60, 30, 14, 'Years to production — patience on fruit'),
    ('pear', 'seedling', 0.80, 1, 1.20, 5.50, 6.50, 0.60, 0.90, 15, 24, 55, 70, 15, 16, 'Bench liner'),
    ('pear', 'early_veg', 1, 1.30, 1.60, 5.50, 6.50, 0.80, 1.10, 15, 26, 50, 65, 25, 16, 'Fire blight risk — airflow matters'),
    ('pear', 'late_veg', 1.10, 1.50, 1.90, 5.50, 6.50, 0.90, 1.20, 15, 26, 45, 60, 30, 16, 'Young tree training'),
    ('plum', 'seedling', 0.80, 1, 1.20, 5.50, 6.50, 0.60, 0.90, 12, 24, 55, 70, 18, 16, 'Bench graft'),
    ('plum', 'early_veg', 1, 1.30, 1.60, 5.50, 6.50, 0.80, 1.10, 15, 28, 45, 60, 28, 16, 'Spring growth flush'),
    ('plum', 'late_veg', 1.10, 1.50, 1.80, 5.50, 6.50, 0.90, 1.20, 15, 28, 40, 55, 32, 16, 'Young tree — fruit years 3–5'),
    ('mango', 'seedling', 0.80, 1, 1.20, 5.50, 6.50, 0.70, 1, 22, 32, 55, 75, 20, 14, 'Warm bench only — no cold roots'),
    ('mango', 'early_veg', 1, 1.40, 1.70, 5.50, 6.50, 0.90, 1.20, 24, 34, 50, 70, 30, 14, 'Container vigor; anthracnose in high RH'),
    ('mango', 'late_veg', 1.20, 1.60, 2, 5.50, 6.50, 1, 1.30, 22, 34, 45, 65, 35, 14, 'Juvenile years — fruit 3–5+ in pots')
) AS s(crop_key, stage, ec_min, ec_target, ec_max, ph_min, ph_max,
         vpd_min_kpa, vpd_max_kpa, temp_min_c, temp_max_c, rh_min_pct, rh_max_pct,
         dli_target, photoperiod_hrs, notes)
  ON p.farm_id IS NULL AND p.is_builtin = TRUE AND p.crop_key = s.crop_key
WHERE NOT EXISTS (
    SELECT 1 FROM gr33ncrops.crop_profile_stages existing
    WHERE existing.crop_profile_id = p.id AND existing.stage = s.stage::gr33nfertigation.growth_stage_enum
);
