-- ============================================================
-- Phase 20.8 WS1 — Animal husbandry flesh-out
--
-- Phase 20.95 WS4 already pre-shipped the operational columns on
-- animal_groups (count, primary_zone_id, active, archived_at,
-- archived_reason) and the topology FKs on aquaponics.loops
-- (fish_tank_zone_id, grow_bed_zone_id) -- see
-- 20260508_phase2095_animal_aquaponics_scope.sql.
--
-- 20.8 WS1 is the rest:
--   * gr33nanimals.animal_lifecycle_events  (NEW table)
--   * gr33naquaponics.loops.active          (matching active flag)
--
-- Strictly additive. event_type is TEXT (not enum) deliberately —
-- the vocabulary will settle from real usage before we tighten it.
-- delta_count is signed (+ for added/born, - for died/sold/culled,
-- NULL for note/health). animal_groups.count stays manually edited;
-- the UI surfaces "sum of deltas vs stored count" as a sanity nudge.
-- ============================================================

CREATE TABLE IF NOT EXISTS gr33nanimals.animal_lifecycle_events (
    id              BIGSERIAL PRIMARY KEY,
    farm_id         BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    animal_group_id BIGINT NOT NULL REFERENCES gr33nanimals.animal_groups(id) ON DELETE CASCADE,
    event_type      TEXT NOT NULL,
    event_time      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    delta_count     INTEGER,
    notes           TEXT,
    recorded_by     UUID REFERENCES gr33ncore.profiles(user_id) ON DELETE SET NULL,
    related_task_id BIGINT REFERENCES gr33ncore.tasks(id) ON DELETE SET NULL,
    meta            JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_animal_lifecycle_events_group_time
    ON gr33nanimals.animal_lifecycle_events (animal_group_id, event_time DESC);
CREATE INDEX IF NOT EXISTS idx_animal_lifecycle_events_farm_time
    ON gr33nanimals.animal_lifecycle_events (farm_id, event_time DESC);

ALTER TABLE gr33naquaponics.loops
    ADD COLUMN IF NOT EXISTS active BOOLEAN NOT NULL DEFAULT TRUE;
