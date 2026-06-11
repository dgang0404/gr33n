-- Phase 82 WS4c — Tier B built-in crop profiles (9 crops).
-- Source of truth: data/crop_library.yaml

INSERT INTO gr33ncrops.crop_profiles (farm_id, crop_key, display_name, category, source, version, is_builtin)
SELECT NULL, v.crop_key, v.display_name, v.category, v.source, 2, TRUE
FROM (VALUES
    ('zucchini',    'Zucchini / summer squash', 'fruiting',   'Fruiting squash; modeled on cucumber'),
    ('green_bean',  'Green bean',               'fruiting',   'Moderate EC warm legume; modeled on pepper'),
    ('mint',        'Mint',                     'herb',       'Aggressive roots — contain in pots; modeled on basil'),
    ('parsley',     'Parsley',                  'herb',       'Biennial herb baseline; slightly cooler than basil'),
    ('blueberry',   'Blueberry',                'fruiting',   'Acidic pH band 4.5–5.5; modeled on strawberry'),
    ('hemp',        'Hemp (fiber/seed)',          'industrial', 'Separate from cannabis flower profile — vegetative fiber/seed baseline only'),
    ('broccoli',    'Broccoli',                 'leafy',      'Cool brassica; modeled on kale with lower temps'),
    ('melon',       'Melon / cantaloupe',         'fruiting',   'Warm vining melon; high transpiration; modeled on cucumber'),
    ('arugula',     'Arugula / rocket',         'leafy',      'Fast turnover leafy; bolts in heat; modeled on lettuce')
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
    ('zucchini', 'early_veg',    1.80, 2.20, 2.60, 5.5, 6.0, 0.80, 1.10, 22.0, 28.0, 65, 75, 30.0, 16.0, 'Bush or vining squash veg'),
    ('zucchini', 'late_veg',     2.20, 2.60, 3.00, 5.5, 6.0, 0.90, 1.20, 22.0, 28.0, 60, 70, 35.0, 16.0, 'Pre-fruit; support heavy squash'),
    ('zucchini', 'early_flower', 2.40, 2.80, 3.20, 5.5, 6.0, 1.00, 1.30, 22.0, 28.0, 55, 65, 35.0, 16.0, 'Fruit set; hand-pollinate if needed indoors'),
    ('green_bean', 'early_veg',    1.40, 1.80, 2.20, 5.5, 6.0, 0.80, 1.10, 22.0, 28.0, 55, 65, 25.0, 16.0, 'Warm veg; avoid cold irrigation'),
    ('green_bean', 'late_veg',     1.80, 2.20, 2.60, 5.5, 6.0, 0.90, 1.20, 22.0, 28.0, 50, 60, 30.0, 16.0, 'Runner / trellis'),
    ('green_bean', 'early_flower', 2.00, 2.40, 2.80, 5.5, 6.0, 1.00, 1.30, 22.0, 28.0, 45, 55, 30.0, 16.0, 'Pod set; moderate EC'),
    ('mint', 'seedling',   0.80, 1.00, 1.20, 5.5, 6.0, 0.60, 0.90, 20.0, 26.0, 60, 70, 15.0, 16.0, 'Root in contained module'),
    ('mint', 'early_veg',  1.00, 1.40, 1.60, 5.5, 6.0, 0.80, 1.10, 20.0, 28.0, 55, 65, 20.0, 16.0, 'Continuous harvest; isolate roots from other crops'),
    ('mint', 'late_veg',   1.20, 1.60, 1.80, 5.5, 6.0, 0.90, 1.20, 20.0, 28.0, 50, 60, 22.0, 16.0, 'Trim often; roots spread aggressively'),
    ('parsley', 'seedling',   0.80, 1.00, 1.20, 5.5, 6.0, 0.60, 0.90, 18.0, 24.0, 60, 70, 14.0, 14.0, 'Slow germination; cool start'),
    ('parsley', 'early_veg',  1.00, 1.30, 1.50, 5.5, 6.0, 0.80, 1.10, 18.0, 24.0, 55, 65, 18.0, 14.0, 'Leaf harvest; cooler than basil'),
    ('parsley', 'late_veg',   1.10, 1.50, 1.70, 5.5, 6.0, 0.90, 1.20, 18.0, 24.0, 50, 60, 20.0, 14.0, 'Second-year bolt if kept too warm'),
    ('blueberry', 'seedling',     0.80, 1.00, 1.20, 4.5, 5.5, 0.60, 0.90, 18.0, 22.0, 65, 75, 12.0, 14.0, 'Acidic root zone required'),
    ('blueberry', 'early_veg',    1.00, 1.30, 1.60, 4.5, 5.5, 0.80, 1.10, 18.0, 24.0, 60, 70, 18.0, 14.0, 'Crown / vegetative growth'),
    ('blueberry', 'early_flower', 1.10, 1.50, 1.90, 4.5, 5.5, 0.90, 1.20, 18.0, 24.0, 55, 65, 20.0, 14.0, 'Berry set; monitor pH drift'),
    ('hemp', 'early_veg', 0.80, 1.00, 1.20, 5.8, 6.2, 0.80, 1.00, 22.0, 26.0, 60, 70, 30.0, 18.0, 'Fiber/seed — not flower EC curve'),
    ('hemp', 'late_veg',  1.00, 1.30, 1.50, 5.8, 6.2, 0.90, 1.10, 22.0, 26.0, 55, 65, 35.0, 18.0, 'Long veg; do not use cannabis flower targets'),
    ('broccoli', 'seedling',   0.80, 1.00, 1.20, 5.5, 6.0, 0.50, 0.80, 14.0, 20.0, 65, 75, 12.0, 16.0, 'Cool brassica seedling'),
    ('broccoli', 'early_veg',  1.00, 1.20, 1.40, 5.5, 6.0, 0.60, 0.90, 14.0, 20.0, 60, 70, 15.0, 16.0, 'Head formation; keep cool'),
    ('broccoli', 'late_veg',   1.10, 1.30, 1.50, 5.5, 6.0, 0.70, 1.00, 14.0, 20.0, 55, 65, 17.0, 16.0, 'Pre-harvest; bolt in heat'),
    ('melon', 'early_veg',    1.80, 2.20, 2.60, 5.5, 6.0, 0.80, 1.10, 24.0, 30.0, 65, 75, 30.0, 16.0, 'Warm vining veg'),
    ('melon', 'late_veg',     2.20, 2.60, 3.00, 5.5, 6.0, 0.90, 1.20, 24.0, 30.0, 60, 70, 35.0, 16.0, 'Runner growth; high water demand'),
    ('melon', 'early_flower', 2.50, 2.90, 3.30, 5.5, 6.0, 1.00, 1.30, 24.0, 30.0, 55, 65, 40.0, 16.0, 'Fruiting; very high transpiration'),
    ('arugula', 'seedling',   0.60, 0.80, 1.00, 5.5, 6.0, 0.50, 0.80, 16.0, 20.0, 65, 75, 12.0, 12.0, 'Fast germination'),
    ('arugula', 'early_veg',  0.80, 1.00, 1.20, 5.5, 6.0, 0.60, 0.90, 16.0, 22.0, 60, 70, 15.0, 12.0, '3–4 week crop; harvest young'),
    ('arugula', 'late_veg',   0.90, 1.10, 1.30, 5.5, 6.0, 0.70, 1.00, 16.0, 22.0, 55, 65, 17.0, 12.0, 'Bolt and bitterness in sustained heat')
) AS s(crop_key, stage, ec_min, ec_target, ec_max, ph_min, ph_max,
         vpd_min_kpa, vpd_max_kpa, temp_min_c, temp_max_c, rh_min_pct, rh_max_pct,
         dli_target, photoperiod_hrs, notes)
  ON p.farm_id IS NULL AND p.is_builtin = TRUE AND p.crop_key = s.crop_key
WHERE NOT EXISTS (
    SELECT 1 FROM gr33ncrops.crop_profile_stages existing
    WHERE existing.crop_profile_id = p.id AND existing.stage = s.stage::gr33nfertigation.growth_stage_enum
);
