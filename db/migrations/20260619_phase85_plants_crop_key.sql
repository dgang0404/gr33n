-- Phase 85 WS1 — catalog-bound plants: one farm slot per crop_key.

ALTER TABLE gr33ncrops.plants
    ADD COLUMN IF NOT EXISTS crop_key TEXT;

-- Backfill from linked crop profile.
UPDATE gr33ncrops.plants pl
SET crop_key = cp.crop_key
FROM gr33ncrops.crop_profiles cp
WHERE pl.crop_profile_id = cp.id
  AND pl.crop_key IS NULL
  AND cp.crop_key IS NOT NULL
  AND pl.deleted_at IS NULL;

-- Dedupe: keep lowest id per (farm_id, crop_key); soft-delete extras.
WITH ranked AS (
    SELECT id,
           ROW_NUMBER() OVER (PARTITION BY farm_id, crop_key ORDER BY id) AS rn
    FROM gr33ncrops.plants
    WHERE deleted_at IS NULL
      AND crop_key IS NOT NULL
)
UPDATE gr33ncrops.plants p
SET deleted_at = NOW()
FROM ranked r
WHERE p.id = r.id
  AND r.rn > 1;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'plants_crop_key_fkey'
          AND conrelid = 'gr33ncrops.plants'::regclass
    ) THEN
        ALTER TABLE gr33ncrops.plants
            ADD CONSTRAINT plants_crop_key_fkey
            FOREIGN KEY (crop_key) REFERENCES gr33ncrops.crop_catalog_entries (crop_key);
    END IF;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS idx_plants_farm_crop_key
    ON gr33ncrops.plants (farm_id, crop_key)
    WHERE deleted_at IS NULL AND crop_key IS NOT NULL;

COMMENT ON COLUMN gr33ncrops.plants.crop_key IS
    'Platform catalog identity (Phase 85). One active row per (farm_id, crop_key). display_name mirrors catalog.';
