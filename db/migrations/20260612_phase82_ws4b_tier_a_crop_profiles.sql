-- Phase 82 WS4b — Tier A built-in crop profiles (eggplant, cucumber, kale, spinach, cilantro, microgreens).
-- Source of truth: data/crop_library.yaml — regenerate via ./scripts/generate-crop-seed.sql.sh

INSERT INTO gr33ncrops.crop_profiles (farm_id, crop_key, display_name, category, source, version, is_builtin)
SELECT NULL, v.crop_key, v.display_name, v.category, v.source, 2, TRUE
FROM (VALUES
    ('eggplant',   'Eggplant',            'fruiting', 'Solanaceous fruiting; ~10% lower EC than tomato; hand-pollinate indoors'),
    ('cucumber',   'Cucumber',            'fruiting', 'Vining fruiting cucumber; higher RH than tomato'),
    ('kale',       'Kale',                'leafy',    'Leafy brassica; slightly higher EC than lettuce'),
    ('spinach',    'Spinach',             'leafy',    'Cool-season leafy; bolts in heat'),
    ('cilantro',   'Cilantro / coriander', 'herb',     'Cool-leaning herb; bolts in heat like basil but prefers lower temps'),
    ('microgreens','Microgreens',         'leafy',    'Very low EC; 10–14 day turnover; shallow moisture')
) AS v(crop_key, display_name, category, source)
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
    ('eggplant', 'seedling',     0.90, 1.10, 1.30, 5.5, 6.0, 0.60, 0.90, 22.0, 26.0, 65, 75, 15.0, 16.0, 'Transplant; warm'),
    ('eggplant', 'early_veg',    1.60, 2.00, 2.40, 5.5, 6.0, 0.80, 1.10, 22.0, 28.0, 60, 70, 25.0, 16.0, 'Vegetative; hand-pollinate when flowering indoors'),
    ('eggplant', 'late_veg',     2.00, 2.40, 2.70, 5.5, 6.0, 0.90, 1.20, 22.0, 28.0, 55, 65, 30.0, 16.0, 'Pre-fruit'),
    ('eggplant', 'early_flower', 2.20, 2.50, 2.90, 5.5, 6.0, 1.00, 1.30, 22.0, 28.0, 50, 60, 30.0, 16.0, 'Fruit set; vibrate flowers if no pollinators'),
    ('cucumber', 'early_veg',    1.80, 2.20, 2.60, 5.5, 6.0, 0.80, 1.10, 22.0, 28.0, 65, 75, 30.0, 16.0, 'Vining veg; higher humidity than tomato'),
    ('cucumber', 'late_veg',     2.20, 2.60, 3.00, 5.5, 6.0, 0.90, 1.20, 22.0, 28.0, 60, 70, 35.0, 16.0, 'Runner / trellis growth'),
    ('cucumber', 'early_flower', 2.40, 2.80, 3.20, 5.5, 6.0, 1.00, 1.30, 22.0, 28.0, 55, 65, 35.0, 16.0, 'Fruiting; monitor transpiration on long vines'),
    ('kale', 'seedling',         0.80, 1.00, 1.20, 5.5, 6.0, 0.50, 0.80, 16.0, 22.0, 65, 75, 12.0, 16.0, 'Cool seedling'),
    ('kale', 'early_veg',        1.00, 1.20, 1.40, 5.5, 6.0, 0.60, 0.90, 16.0, 22.0, 60, 70, 15.0, 16.0, 'Leaf expansion; tolerates cooler than basil'),
    ('kale', 'late_veg',         1.10, 1.30, 1.50, 5.5, 6.0, 0.70, 1.00, 16.0, 22.0, 55, 65, 17.0, 16.0, 'Pre-harvest; baby or full leaf'),
    ('spinach', 'seedling',      0.60, 0.80, 1.00, 5.5, 6.0, 0.50, 0.80, 15.0, 20.0, 65, 75, 12.0, 14.0, 'Cool seedling; avoid heat'),
    ('spinach', 'early_veg',     0.80, 1.00, 1.20, 5.5, 6.0, 0.60, 0.90, 15.0, 20.0, 60, 70, 15.0, 14.0, 'Leaf crop; bolts above ~24 °C'),
    ('spinach', 'late_veg',      0.90, 1.10, 1.30, 5.5, 6.0, 0.70, 1.00, 15.0, 20.0, 55, 65, 17.0, 14.0, 'Harvest before bolting in warm rooms'),
    ('cilantro', 'seedling',     0.80, 1.00, 1.20, 5.5, 6.0, 0.60, 0.90, 18.0, 22.0, 60, 70, 15.0, 14.0, 'Cool germination; slow in heat'),
    ('cilantro', 'early_veg',    1.00, 1.40, 1.60, 5.5, 6.0, 0.80, 1.10, 18.0, 24.0, 55, 65, 18.0, 14.0, 'Leaf harvest; bolts in sustained heat'),
    ('cilantro', 'late_veg',     1.20, 1.60, 1.80, 5.5, 6.0, 0.90, 1.20, 18.0, 24.0, 50, 60, 20.0, 14.0, 'Succession plant; do not let go to seed indoors'),
    ('microgreens', 'seedling',  0.40, 0.55, 0.70, 5.5, 6.0, 0.50, 0.80, 18.0, 22.0, 60, 75, 10.0, 16.0, 'Germination; mist or shallow water only'),
    ('microgreens', 'early_veg', 0.55, 0.70, 0.85, 5.5, 6.0, 0.60, 0.90, 18.0, 22.0, 55, 70, 12.0, 16.0, '10–14 d cycle; harvest at cotyledon/first true leaf')
) AS s(crop_key, stage, ec_min, ec_target, ec_max, ph_min, ph_max,
         vpd_min_kpa, vpd_max_kpa, temp_min_c, temp_max_c, rh_min_pct, rh_max_pct,
         dli_target, photoperiod_hrs, notes)
  ON p.farm_id IS NULL AND p.is_builtin = TRUE AND p.crop_key = s.crop_key
WHERE NOT EXISTS (
    SELECT 1 FROM gr33ncrops.crop_profile_stages existing
    WHERE existing.crop_profile_id = p.id AND existing.stage = s.stage::gr33nfertigation.growth_stage_enum
);
