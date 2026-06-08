-- Phase 51 WS1 — device config_version for Pi sync + bump on wiring changes.

ALTER TABLE gr33ncore.devices
    ADD COLUMN IF NOT EXISTS config_version INTEGER NOT NULL DEFAULT 0;

COMMENT ON COLUMN gr33ncore.devices.config_version IS
    'Incremented when sensor/actuator wiring for this device changes; Pi polls GET /devices/by-uid/{uid}/config/version';

-- Seed non-zero version for devices that already have wired entities (demo farm).
UPDATE gr33ncore.devices d
SET config_version = GREATEST(d.config_version, 1)
WHERE d.deleted_at IS NULL
  AND (
    EXISTS (
        SELECT 1 FROM gr33ncore.sensors s
        WHERE s.deleted_at IS NULL
          AND (s.config->'wiring'->>'device_id')::bigint = d.id
    )
    OR EXISTS (
        SELECT 1 FROM gr33ncore.actuators a
        WHERE a.deleted_at IS NULL
          AND (a.config->'wiring'->>'device_id')::bigint = d.id
    )
  );

CREATE OR REPLACE FUNCTION gr33ncore.bump_device_config_version_from_entity()
RETURNS TRIGGER AS $$
DECLARE
    dev_id BIGINT;
BEGIN
    IF TG_OP = 'UPDATE' AND (OLD.config->'wiring') IS NOT DISTINCT FROM (NEW.config->'wiring') THEN
        RETURN NEW;
    END IF;

    dev_id := NULL;
    IF NEW.config ? 'wiring' AND (NEW.config->'wiring'->>'device_id') ~ '^[0-9]+$' THEN
        dev_id := (NEW.config->'wiring'->>'device_id')::bigint;
    END IF;
    IF dev_id IS NULL AND NEW.device_id IS NOT NULL THEN
        dev_id := NEW.device_id;
    END IF;

    IF dev_id IS NOT NULL THEN
        UPDATE gr33ncore.devices
        SET config_version = config_version + 1, updated_at = NOW()
        WHERE id = dev_id AND deleted_at IS NULL;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_sensors_bump_device_config_version ON gr33ncore.sensors;
CREATE TRIGGER trg_sensors_bump_device_config_version
    AFTER INSERT OR UPDATE OF config ON gr33ncore.sensors
    FOR EACH ROW
    WHEN (NEW.config ? 'wiring')
    EXECUTE FUNCTION gr33ncore.bump_device_config_version_from_entity();

DROP TRIGGER IF EXISTS trg_actuators_bump_device_config_version ON gr33ncore.actuators;
CREATE TRIGGER trg_actuators_bump_device_config_version
    AFTER INSERT OR UPDATE OF config ON gr33ncore.actuators
    FOR EACH ROW
    WHEN (NEW.config ? 'wiring')
    EXECUTE FUNCTION gr33ncore.bump_device_config_version_from_entity();
