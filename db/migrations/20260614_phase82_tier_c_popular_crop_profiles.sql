-- Phase 82 WS4g — Tier C + popular crops (14 profiles). Completes crop_library.yaml v3 (36 crops).
-- Source: data/crop_library.yaml

INSERT INTO gr33ncrops.crop_profiles (farm_id, crop_key, display_name, category, source, version, is_builtin)
SELECT NULL, v.crop_key, v.display_name, v.category, v.source, 3, TRUE
FROM (VALUES
    ('rose',          'Rose (cut flower)',        'flower',     'Cut-flower rose; moderate EC; long photoperiod'),
    ('sunflower',     'Sunflower',                'flower',     'Short-cycle high-light flower; fast turnover'),
    ('hops',          'Hops (bines)',             'industrial', 'Long vegetative bines — not cannabis flower profile'),
    ('succulent',     'Succulents (general)',     'ornamental', 'Dry-down epiphyte/soilless — never constant wet'),
    ('san_pedro',     'San Pedro cactus',         'ornamental', 'Columnar cactus — minimal EC; dry winter rest'),
    ('houseplant',    'Houseplant (general)',     'ornamental', 'Conservative foliage baseline — many species; clone to customize'),
    ('chard',         'Swiss chard',              'leafy',      'Leafy beet relative; kale-class EC'),
    ('bok_choy',      'Bok choy / pak choi',      'leafy',      'Cool Asian brassica; fast head crop'),
    ('radish',        'Radish',                   'leafy',      'Fast root/leaf crop; low EC; 3–4 week turnover'),
    ('thyme',         'Thyme',                    'herb',       'Woody herb; lower EC; lean dry-down between feeds'),
    ('oregano',       'Oregano',                  'herb',       'Mediterranean herb; moderate EC; drier than basil'),
    ('rosemary',      'Rosemary',                 'herb',       'Woody Mediterranean herb; lean feed; excellent drainage'),
    ('lavender',      'Lavender',                 'herb',       'Dry-lean aromatic; low EC; alkaline-leaning ok'),
    ('chrysanthemum', 'Chrysanthemum (mum)',      'flower',     'Photoperiod-sensitive cut flower; moderate EC')
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
    ('rose', 'early_veg',    1.40, 1.80, 2.20, 5.5, 6.0, 0.80, 1.10, 18.0, 24.0, 55, 65, 25.0, 16.0, 'Vegetative cane growth'),
    ('rose', 'late_veg',     1.80, 2.20, 2.60, 5.5, 6.0, 0.90, 1.20, 18.0, 24.0, 50, 60, 30.0, 16.0, 'Pre-bud buildup'),
    ('rose', 'early_flower', 2.00, 2.40, 2.80, 5.5, 6.0, 1.00, 1.30, 18.0, 22.0, 45, 55, 30.0, 16.0, 'Bud / cut-flower production'),
    ('sunflower', 'seedling',  0.80, 1.00, 1.20, 5.5, 6.0, 0.60, 0.90, 20.0, 26.0, 60, 70, 20.0, 16.0, 'Transplant quickly — roots sensitive'),
    ('sunflower', 'early_veg', 1.20, 1.60, 2.00, 5.5, 6.0, 0.90, 1.20, 20.0, 28.0, 50, 60, 35.0, 16.0, 'Short cycle to bloom; high light'),
    ('hops', 'early_veg', 1.00, 1.40, 1.80, 5.8, 6.2, 0.90, 1.20, 18.0, 24.0, 55, 65, 35.0, 18.0, 'Bine establishment'),
    ('hops', 'late_veg',  1.40, 1.80, 2.20, 5.8, 6.2, 1.00, 1.30, 18.0, 24.0, 50, 60, 40.0, 18.0, 'Long veg before cone season — trellis required'),
    ('succulent', 'seedling',  0.30, 0.50, 0.70, 5.5, 6.5, 0.50, 0.90, 18.0, 26.0, 40, 55, 12.0, 14.0, 'Root lightly; excellent drainage'),
    ('succulent', 'early_veg', 0.40, 0.60, 0.80, 5.5, 6.5, 0.60, 1.00, 18.0, 28.0, 35, 50, 15.0, 14.0, 'Full dry-down between feeds'),
    ('san_pedro', 'seedling',  0.20, 0.35, 0.50, 5.5, 6.5, 0.40, 0.80, 18.0, 26.0, 35, 50, 10.0, 14.0, 'Seedling — almost no feed'),
    ('san_pedro', 'early_veg', 0.25, 0.40, 0.55, 5.5, 6.5, 0.50, 0.90, 18.0, 28.0, 30, 45, 15.0, 14.0, 'Active growth season only; rest dry in winter'),
    ('houseplant', 'seedling',  0.60, 0.80, 1.00, 5.5, 6.2, 0.50, 0.80, 20.0, 26.0, 55, 70, 8.0, 12.0, 'Low light tolerant baseline'),
    ('houseplant', 'early_veg', 0.80, 1.00, 1.30, 5.5, 6.2, 0.60, 0.90, 20.0, 26.0, 50, 65, 12.0, 12.0, 'Foliar houseplant — adjust per species'),
    ('chard', 'seedling',   0.80, 1.00, 1.20, 5.5, 6.0, 0.50, 0.80, 16.0, 22.0, 65, 75, 12.0, 16.0, 'Colorful stems — cool start'),
    ('chard', 'early_veg',  1.00, 1.20, 1.40, 5.5, 6.0, 0.60, 0.90, 16.0, 24.0, 60, 70, 15.0, 16.0, 'Continuous leaf harvest'),
    ('chard', 'late_veg',   1.10, 1.30, 1.50, 5.5, 6.0, 0.70, 1.00, 16.0, 24.0, 55, 65, 17.0, 16.0, 'Cut outer leaves; heat softens texture'),
    ('bok_choy', 'seedling',   0.70, 0.90, 1.10, 5.5, 6.0, 0.50, 0.80, 14.0, 20.0, 65, 75, 12.0, 14.0, 'Cool germination'),
    ('bok_choy', 'early_veg',  0.90, 1.10, 1.30, 5.5, 6.0, 0.60, 0.90, 14.0, 22.0, 60, 70, 15.0, 14.0, 'Head fill; bolts in heat'),
    ('bok_choy', 'late_veg',   1.00, 1.20, 1.40, 5.5, 6.0, 0.70, 1.00, 14.0, 22.0, 55, 65, 17.0, 14.0, 'Harvest whole heads promptly'),
    ('radish', 'seedling',  0.50, 0.70, 0.90, 5.5, 6.0, 0.50, 0.80, 15.0, 20.0, 65, 75, 12.0, 14.0, 'Direct or plug; cool'),
    ('radish', 'early_veg', 0.70, 0.90, 1.10, 5.5, 6.0, 0.60, 0.90, 15.0, 22.0, 60, 70, 15.0, 14.0, 'Bulb swell; harvest before pithy'),
    ('thyme', 'seedling',  0.60, 0.80, 1.00, 5.5, 6.0, 0.60, 0.90, 18.0, 24.0, 50, 60, 15.0, 14.0, 'Slow from seed — prefer cuttings'),
    ('thyme', 'early_veg', 0.80, 1.10, 1.30, 5.5, 6.0, 0.80, 1.10, 18.0, 26.0, 45, 55, 18.0, 14.0, 'Aromatic oils — do not over-feed'),
    ('oregano', 'seedling',  0.70, 0.90, 1.10, 5.5, 6.0, 0.60, 0.90, 18.0, 24.0, 50, 60, 15.0, 14.0, 'Warm germination'),
    ('oregano', 'early_veg', 1.00, 1.30, 1.50, 5.5, 6.0, 0.80, 1.10, 20.0, 28.0, 45, 55, 20.0, 14.0, 'Continuous harvest; trim to bush'),
    ('oregano', 'late_veg',  1.10, 1.40, 1.60, 5.5, 6.0, 0.90, 1.20, 20.0, 28.0, 40, 50, 22.0, 14.0, 'Lower RH reduces mildew on dense canopy'),
    ('rosemary', 'seedling',  0.60, 0.80, 1.00, 5.5, 6.2, 0.60, 0.90, 18.0, 24.0, 45, 55, 15.0, 14.0, 'Cuttings root faster than seed'),
    ('rosemary', 'early_veg', 0.80, 1.10, 1.30, 5.5, 6.2, 0.80, 1.10, 18.0, 26.0, 40, 50, 20.0, 14.0, 'Never soggy — root rot common'),
    ('lavender', 'seedling',  0.50, 0.70, 0.90, 5.8, 6.8, 0.60, 0.90, 18.0, 24.0, 40, 55, 15.0, 14.0, 'Slow from seed'),
    ('lavender', 'early_veg', 0.70, 0.90, 1.10, 5.8, 6.8, 0.80, 1.10, 18.0, 26.0, 35, 50, 20.0, 14.0, 'Dry-down between irrigation; high light for oil'),
    ('chrysanthemum', 'early_veg',    1.20, 1.60, 2.00, 5.5, 6.0, 0.80, 1.10, 18.0, 24.0, 55, 65, 25.0, 16.0, 'Long-day vegetative growth'),
    ('chrysanthemum', 'late_veg',     1.60, 2.00, 2.40, 5.5, 6.0, 0.90, 1.20, 18.0, 24.0, 50, 60, 28.0, 16.0, 'Pinch for branchiness'),
    ('chrysanthemum', 'early_flower', 1.80, 2.20, 2.60, 5.5, 6.0, 1.00, 1.30, 16.0, 22.0, 45, 55, 25.0, 12.0, 'Short-day bloom — photoperiod critical')
) AS s(crop_key, stage, ec_min, ec_target, ec_max, ph_min, ph_max,
         vpd_min_kpa, vpd_max_kpa, temp_min_c, temp_max_c, rh_min_pct, rh_max_pct,
         dli_target, photoperiod_hrs, notes)
  ON p.farm_id IS NULL AND p.is_builtin = TRUE AND p.crop_key = s.crop_key
WHERE NOT EXISTS (
    SELECT 1 FROM gr33ncrops.crop_profile_stages existing
    WHERE existing.crop_profile_id = p.id AND existing.stage = s.stage::gr33nfertigation.growth_stage_enum
);
