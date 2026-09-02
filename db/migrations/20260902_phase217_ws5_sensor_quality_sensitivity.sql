-- Phase 217 WS5 — per-sensor override for the 5-tier quality scale's band
-- width. NULL means "use the farm-wide default" (farms.meta_data.default_sensor_sensitivity).
ALTER TABLE gr33ncore.sensors
    ADD COLUMN IF NOT EXISTS quality_sensitivity TEXT
    CHECK (quality_sensitivity IS NULL OR quality_sensitivity IN ('narrow', 'normal', 'wide'));
