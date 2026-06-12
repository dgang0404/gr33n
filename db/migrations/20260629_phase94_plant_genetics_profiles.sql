-- Phase 94 — per-variety EC profiles (genetics override above farm crop_key override).

CREATE TABLE IF NOT EXISTS gr33ncrops.plant_genetics_profiles (
    farm_id          BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    crop_key         TEXT NOT NULL,
    variety_slug     TEXT NOT NULL,
    variety_label    TEXT NOT NULL,
    crop_profile_id  BIGINT NOT NULL REFERENCES gr33ncrops.crop_profiles(id) ON DELETE CASCADE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (farm_id, crop_key, variety_slug)
);

CREATE INDEX IF NOT EXISTS plant_genetics_profiles_profile_idx
    ON gr33ncrops.plant_genetics_profiles (crop_profile_id);
