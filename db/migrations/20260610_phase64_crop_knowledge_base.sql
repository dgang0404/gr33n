-- Phase 64 WS1+WS2 — crop knowledge base: profiles, stages, plants FK, built-in seed.

CREATE TABLE IF NOT EXISTS gr33ncrops.crop_profiles (
    id            BIGSERIAL PRIMARY KEY,
    farm_id       BIGINT REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    crop_key      TEXT NOT NULL,
    display_name  TEXT NOT NULL,
    category      TEXT,
    source        TEXT,
    version       INTEGER NOT NULL DEFAULT 1,
    is_builtin    BOOLEAN NOT NULL DEFAULT FALSE,
    meta          JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT crop_profiles_farm_key_unique UNIQUE NULLS NOT DISTINCT (farm_id, crop_key)
);

CREATE INDEX IF NOT EXISTS idx_crop_profiles_farm
    ON gr33ncrops.crop_profiles (farm_id)
    WHERE farm_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_crop_profiles_builtin
    ON gr33ncrops.crop_profiles (is_builtin)
    WHERE is_builtin = TRUE;

CREATE TABLE IF NOT EXISTS gr33ncrops.crop_profile_stages (
    id              BIGSERIAL PRIMARY KEY,
    crop_profile_id BIGINT NOT NULL REFERENCES gr33ncrops.crop_profiles(id) ON DELETE CASCADE,
    stage           gr33nfertigation.growth_stage_enum NOT NULL,
    ec_min          NUMERIC(4,2),
    ec_target       NUMERIC(4,2),
    ec_max          NUMERIC(4,2),
    ph_min          NUMERIC(3,1),
    ph_max          NUMERIC(3,1),
    vpd_min_kpa     NUMERIC(3,2),
    vpd_max_kpa     NUMERIC(3,2),
    temp_min_c      NUMERIC(4,1),
    temp_max_c      NUMERIC(4,1),
    rh_min_pct      NUMERIC(4,1),
    rh_max_pct      NUMERIC(4,1),
    dli_target      NUMERIC(4,1),
    photoperiod_hrs NUMERIC(3,1),
    notes           TEXT,
    CONSTRAINT crop_profile_stages_unique UNIQUE (crop_profile_id, stage)
);

CREATE INDEX IF NOT EXISTS idx_crop_profile_stages_profile
    ON gr33ncrops.crop_profile_stages (crop_profile_id);

ALTER TABLE gr33ncrops.plants
    ADD COLUMN IF NOT EXISTS crop_profile_id BIGINT
        REFERENCES gr33ncrops.crop_profiles(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_gr33ncrops_plants_crop_profile
    ON gr33ncrops.plants (crop_profile_id)
    WHERE crop_profile_id IS NOT NULL AND deleted_at IS NULL;

DROP TRIGGER IF EXISTS trg_gr33ncrops_crop_profiles_updated_at ON gr33ncrops.crop_profiles;
CREATE TRIGGER trg_gr33ncrops_crop_profiles_updated_at
    BEFORE UPDATE ON gr33ncrops.crop_profiles
    FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();

COMMENT ON TABLE gr33ncrops.crop_profiles IS
    'Per-crop target profiles — farm_id NULL + is_builtin for platform library (Phase 64).';
COMMENT ON COLUMN gr33ncrops.plants.crop_profile_id IS
    'Optional crop knowledge profile for Guardian targets and grow UI (Phase 64).';

-- Built-in profiles (idempotent — skip if crop_key already exists as builtin).
INSERT INTO gr33ncrops.crop_profiles (farm_id, crop_key, display_name, category, source, version, is_builtin)
SELECT NULL, v.crop_key, v.display_name, v.category, v.source, 1, TRUE
FROM (VALUES
    ('cannabis',      'Cannabis',           'flower',   'Curated indoor ranges; verify against your genetics'),
    ('tomato',        'Tomato',             'fruiting', 'Hydroponic fruiting tomato references'),
    ('pepper',        'Pepper (bell/chili)', 'fruiting', 'Similar to tomato, lower EC headroom'),
    ('lettuce',       'Lettuce / leafy greens', 'leafy', 'Low-EC leafy production'),
    ('phalaenopsis',  'Orchid (Phalaenopsis)', 'epiphyte', 'Epiphyte — very low EC, high RH'),
    ('basil',         'Basil / herbs',      'herb',     'Warm-weather herb baseline'),
    ('strawberry',    'Strawberry',         'fruiting', 'Day-neutral strawberry baseline')
) AS v(crop_key, display_name, category, source)
WHERE NOT EXISTS (
    SELECT 1 FROM gr33ncrops.crop_profiles p
    WHERE p.farm_id IS NULL AND p.crop_key = v.crop_key AND p.is_builtin = TRUE
);

-- Stage rows (insert only when profile exists and stage missing).
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
    -- cannabis
    ('cannabis', 'early_veg',    0.80, 1.00, 1.20, 5.8, 6.2, 0.80, 1.00, 22.0, 26.0, 60, 70, 30.0, 18.0, 'Veg photoperiod'),
    ('cannabis', 'late_veg',     1.00, 1.40, 1.60, 5.8, 6.2, 0.90, 1.10, 22.0, 26.0, 55, 65, 35.0, 18.0, 'Ramp EC before flip'),
    ('cannabis', 'early_flower', 1.40, 1.60, 1.80, 5.8, 6.2, 1.00, 1.20, 20.0, 24.0, 45, 55, 40.0, 12.0, '12/12 flip'),
    ('cannabis', 'mid_flower',   1.60, 1.80, 2.00, 5.8, 6.2, 1.10, 1.30, 20.0, 24.0, 40, 50, 45.0, 12.0, 'Peak EC window'),
    ('cannabis', 'late_flower',  1.20, 1.40, 1.60, 5.8, 6.2, 1.00, 1.20, 18.0, 22.0, 40, 50, 35.0, 12.0, 'Pre-flush taper'),
    ('cannabis', 'flush',        0.00, 0.20, 0.40, 6.0, 6.5, 1.00, 1.30, 18.0, 22.0, 40, 50, 30.0, 12.0, 'Plain water flush'),
    -- tomato
    ('tomato', 'seedling',       1.00, 1.20, 1.40, 5.5, 6.0, 0.60, 0.90, 20.0, 24.0, 65, 75, 15.0, 16.0, 'Seedling / transplant'),
    ('tomato', 'early_veg',      1.80, 2.20, 2.60, 5.5, 6.0, 0.80, 1.10, 20.0, 26.0, 60, 70, 25.0, 16.0, 'Vegetative'),
    ('tomato', 'late_veg',       2.20, 2.60, 3.00, 5.5, 6.0, 0.90, 1.20, 20.0, 26.0, 55, 65, 30.0, 16.0, 'Pre-fruit'),
    ('tomato', 'early_flower',   2.40, 2.80, 3.20, 5.5, 6.0, 1.00, 1.30, 20.0, 26.0, 50, 60, 30.0, 16.0, 'Fruit set'),
    ('tomato', 'mid_flower',     2.80, 3.20, 3.50, 5.5, 6.0, 1.10, 1.40, 20.0, 26.0, 50, 60, 35.0, 16.0, 'Heavy fruiting — high EC'),
    -- pepper
    ('pepper', 'early_veg',      1.40, 1.80, 2.20, 5.5, 6.0, 0.80, 1.10, 22.0, 28.0, 55, 65, 25.0, 16.0, 'Warm veg'),
    ('pepper', 'late_veg',       1.80, 2.20, 2.60, 5.5, 6.0, 0.90, 1.20, 22.0, 28.0, 50, 60, 30.0, 16.0, 'Pre-fruit'),
    ('pepper', 'early_flower',   2.00, 2.40, 2.80, 5.5, 6.0, 1.00, 1.30, 22.0, 28.0, 45, 55, 30.0, 16.0, 'Fruit set'),
    ('pepper', 'mid_flower',     2.20, 2.60, 3.00, 5.5, 6.0, 1.10, 1.40, 22.0, 28.0, 45, 55, 35.0, 16.0, 'Fruiting — lower than tomato peak'),
    -- lettuce
    ('lettuce', 'seedling',      0.60, 0.80, 1.00, 5.5, 6.0, 0.50, 0.80, 18.0, 22.0, 65, 75, 12.0, 16.0, 'Cool seedling'),
    ('lettuce', 'early_veg',     0.80, 1.00, 1.20, 5.5, 6.0, 0.60, 0.90, 18.0, 22.0, 60, 70, 15.0, 16.0, 'Leaf expansion'),
    ('lettuce', 'late_veg',      0.90, 1.10, 1.30, 5.5, 6.0, 0.70, 1.00, 18.0, 22.0, 55, 65, 17.0, 16.0, 'Pre-harvest'),
    -- phalaenopsis
    ('phalaenopsis', 'seedling', 0.30, 0.50, 0.70, 5.5, 6.0, 0.40, 0.70, 20.0, 24.0, 70, 85, 8.0,  12.0, 'Low light epiphyte'),
    ('phalaenopsis', 'early_veg', 0.40, 0.60, 0.80, 5.5, 6.0, 0.50, 0.80, 20.0, 26.0, 65, 80, 10.0, 12.0, 'Vegetative spike growth'),
    ('phalaenopsis', 'early_flower', 0.40, 0.60, 0.80, 5.5, 6.0, 0.60, 0.90, 20.0, 26.0, 60, 75, 12.0, 12.0, 'Spike / bloom — very low EC'),
    -- basil
    ('basil', 'seedling',        0.80, 1.00, 1.20, 5.5, 6.0, 0.60, 0.90, 22.0, 26.0, 60, 70, 15.0, 16.0, 'Warm herb seedling'),
    ('basil', 'early_veg',       1.00, 1.40, 1.60, 5.5, 6.0, 0.80, 1.10, 22.0, 28.0, 55, 65, 20.0, 16.0, 'Vegetative harvest'),
    ('basil', 'late_veg',        1.20, 1.60, 1.80, 5.5, 6.0, 0.90, 1.20, 22.0, 28.0, 50, 60, 22.0, 16.0, 'Continuous harvest'),
    -- strawberry
    ('strawberry', 'seedling',   0.80, 1.00, 1.20, 5.5, 6.0, 0.60, 0.90, 18.0, 22.0, 65, 75, 12.0, 14.0, 'Transplant'),
    ('strawberry', 'early_veg',  1.00, 1.40, 1.60, 5.5, 6.0, 0.80, 1.10, 18.0, 24.0, 60, 70, 18.0, 14.0, 'Runner / crown growth'),
    ('strawberry', 'early_flower', 1.20, 1.60, 2.00, 5.5, 6.0, 0.90, 1.20, 18.0, 24.0, 55, 65, 20.0, 14.0, 'Day-neutral fruiting')
) AS s(crop_key, stage, ec_min, ec_target, ec_max, ph_min, ph_max,
         vpd_min_kpa, vpd_max_kpa, temp_min_c, temp_max_c, rh_min_pct, rh_max_pct,
         dli_target, photoperiod_hrs, notes)
  ON p.farm_id IS NULL AND p.is_builtin = TRUE AND p.crop_key = s.crop_key
WHERE NOT EXISTS (
    SELECT 1 FROM gr33ncrops.crop_profile_stages existing
    WHERE existing.crop_profile_id = p.id AND existing.stage = s.stage::gr33nfertigation.growth_stage_enum
);
