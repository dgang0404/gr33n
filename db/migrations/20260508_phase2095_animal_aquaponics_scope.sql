-- ============================================================
-- Phase 20.95 WS4 — animal_groups + aquaponics.loops scope columns
--
-- Additive only. All new columns are nullable or have sane defaults so
-- existing rows do not need a backfill. Phase 20.6 WS3/WS4 will add UI
-- and handlers that actually populate these columns.
-- ============================================================

-- animal_groups scope ---------------------------------------------------
ALTER TABLE gr33nanimals.animal_groups
    ADD COLUMN IF NOT EXISTS count            INTEGER,
    ADD COLUMN IF NOT EXISTS primary_zone_id  BIGINT REFERENCES gr33ncore.zones(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS active           BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS archived_at      TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS archived_reason  TEXT;

CREATE INDEX IF NOT EXISTS idx_gr33nanimals_groups_primary_zone
    ON gr33nanimals.animal_groups (primary_zone_id)
    WHERE deleted_at IS NULL;

-- aquaponics loop topology ----------------------------------------------
ALTER TABLE gr33naquaponics.loops
    ADD COLUMN IF NOT EXISTS fish_tank_zone_id BIGINT REFERENCES gr33ncore.zones(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS grow_bed_zone_id  BIGINT REFERENCES gr33ncore.zones(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_gr33naquaponics_loops_fish_tank_zone
    ON gr33naquaponics.loops (fish_tank_zone_id)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_gr33naquaponics_loops_grow_bed_zone
    ON gr33naquaponics.loops (grow_bed_zone_id)
    WHERE deleted_at IS NULL;
