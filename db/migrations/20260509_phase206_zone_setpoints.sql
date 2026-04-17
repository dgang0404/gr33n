-- Phase 20.6 WS1 — stage-scoped setpoints.
-- Additive only. No existing tables change. No enum changes.
--
-- A zone-scoped row (crop_cycle_id NULL) applies to any cycle running in
-- the zone. A cycle-scoped row (crop_cycle_id NOT NULL) overrides. A
-- NULL `stage` means "all stages for this scope" (the fallback inside
-- that scope). Resolution order at eval time: cycle+stage > cycle-any >
-- zone+stage > zone-any > nothing.
--
-- `stage` is TEXT (not the gr33nfertigation.growth_stage_enum) so the
-- same table can carry setpoints for non-crop zones — drying rooms,
-- propagation areas, aquaponics loops that don't have a fertigation
-- crop cycle.

CREATE TABLE IF NOT EXISTS gr33ncore.zone_setpoints (
    id              BIGSERIAL PRIMARY KEY,
    farm_id         BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    zone_id         BIGINT REFERENCES gr33ncore.zones(id) ON DELETE CASCADE,
    crop_cycle_id   BIGINT REFERENCES gr33nfertigation.crop_cycles(id) ON DELETE CASCADE,
    stage           TEXT,
    sensor_type     TEXT NOT NULL,
    min_value       NUMERIC,
    max_value       NUMERIC,
    ideal_value     NUMERIC,
    meta            JSONB NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_setpoint_scope CHECK (zone_id IS NOT NULL OR crop_cycle_id IS NOT NULL),
    CONSTRAINT chk_setpoint_numeric_coherent CHECK (
        (min_value IS NULL OR max_value IS NULL OR min_value <= max_value) AND
        (ideal_value IS NULL OR min_value IS NULL OR ideal_value >= min_value) AND
        (ideal_value IS NULL OR max_value IS NULL OR ideal_value <= max_value)
    )
);

CREATE INDEX IF NOT EXISTS idx_zone_setpoints_zone_stage
    ON gr33ncore.zone_setpoints (zone_id, stage, sensor_type);
CREATE INDEX IF NOT EXISTS idx_zone_setpoints_cycle_stage
    ON gr33ncore.zone_setpoints (crop_cycle_id, stage, sensor_type);
CREATE INDEX IF NOT EXISTS idx_zone_setpoints_farm
    ON gr33ncore.zone_setpoints (farm_id);

DROP TRIGGER IF EXISTS trg_zone_setpoints_updated_at ON gr33ncore.zone_setpoints;
CREATE TRIGGER trg_zone_setpoints_updated_at
    BEFORE UPDATE ON gr33ncore.zone_setpoints
    FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();
