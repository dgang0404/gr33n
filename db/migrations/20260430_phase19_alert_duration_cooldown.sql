-- Phase 19 WS2: alert duration + cooldown on sensors.
-- Adds three columns to gr33ncore.sensors:
--   alert_duration_seconds   : minimum sustained breach before an alert fires (0 = fire on first reading).
--   alert_cooldown_seconds   : minimum quiet window after an alert before another can fire.
--   alert_breach_started_at  : evaluator state — when the current out-of-range streak started (NULL = in bounds).
-- All are additive, default-bearing, and safe to re-run.

ALTER TABLE gr33ncore.sensors
    ADD COLUMN IF NOT EXISTS alert_duration_seconds  INTEGER     NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS alert_cooldown_seconds  INTEGER     NOT NULL DEFAULT 300,
    ADD COLUMN IF NOT EXISTS alert_breach_started_at TIMESTAMPTZ NULL;

-- Guard against absurd / negative values; keep ceilings generous for future "once per day" use cases.
DO $$ BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'chk_sensor_alert_duration_nonneg'
    ) THEN
        ALTER TABLE gr33ncore.sensors
            ADD CONSTRAINT chk_sensor_alert_duration_nonneg
            CHECK (alert_duration_seconds >= 0 AND alert_duration_seconds <= 86400);
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'chk_sensor_alert_cooldown_nonneg'
    ) THEN
        ALTER TABLE gr33ncore.sensors
            ADD CONSTRAINT chk_sensor_alert_cooldown_nonneg
            CHECK (alert_cooldown_seconds >= 0 AND alert_cooldown_seconds <= 604800);
    END IF;
END $$;
