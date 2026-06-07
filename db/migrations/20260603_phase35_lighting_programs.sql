-- Phase 35 WS1: lighting_programs domain
-- Promotes photoperiod configuration from two orphan schedules into a
-- first-class gr33ncore.lighting_programs entity that owns the ON/OFF
-- schedule pair and the linked actuator.

BEGIN;

-- ── lighting_programs ────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS gr33ncore.lighting_programs (
    id              BIGSERIAL PRIMARY KEY,
    farm_id         BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    zone_id         BIGINT NOT NULL REFERENCES gr33ncore.zones(id) ON DELETE CASCADE,
    actuator_id     BIGINT NOT NULL REFERENCES gr33ncore.actuators(id),
    name            TEXT   NOT NULL,
    description     TEXT,
    -- Integer on/off hours (sum must equal 24 for a standard 24h photoperiod).
    on_hours        INTEGER NOT NULL CHECK (on_hours > 0 AND on_hours <= 24),
    off_hours       INTEGER NOT NULL CHECK (off_hours >= 0 AND off_hours < 24),
    -- "HH:MM" 24-hour anchor for the ON cron; OFF is derived.
    lights_on_at    TEXT NOT NULL DEFAULT '06:00',
    timezone        TEXT NOT NULL DEFAULT 'UTC',
    -- Back-refs to the generated schedule pair (NULL until materialised).
    schedule_on_id  BIGINT REFERENCES gr33ncore.schedules(id) ON DELETE SET NULL,
    schedule_off_id BIGINT REFERENCES gr33ncore.schedules(id) ON DELETE SET NULL,
    -- Optional link to the active crop cycle for stage-specific photoperiod.
    crop_cycle_id   BIGINT REFERENCES gr33nfertigation.crop_cycles(id) ON DELETE SET NULL,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    -- Stores preset_key and operator notes.
    metadata        JSONB NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_lighting_hours_cycle CHECK (on_hours + off_hours = 24)
);

DROP TRIGGER IF EXISTS trg_lighting_programs_updated_at ON gr33ncore.lighting_programs;
CREATE TRIGGER trg_lighting_programs_updated_at
    BEFORE UPDATE ON gr33ncore.lighting_programs
    FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();

COMMIT;
