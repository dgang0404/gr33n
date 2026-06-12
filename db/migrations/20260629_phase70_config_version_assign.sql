-- Phase 70 WS5 — bump device config_version when actuator HAT channel / device assignment changes.

CREATE OR REPLACE FUNCTION gr33ncore.bump_device_config_version_from_actuator()
RETURNS TRIGGER AS $$
DECLARE
    dev_id BIGINT;
BEGIN
    IF TG_OP = 'UPDATE'
       AND OLD.hardware_identifier IS NOT DISTINCT FROM NEW.hardware_identifier
       AND OLD.device_id IS NOT DISTINCT FROM NEW.device_id THEN
        RETURN NEW;
    END IF;

    dev_id := NEW.device_id;
    IF dev_id IS NULL AND TG_OP = 'UPDATE' THEN
        dev_id := OLD.device_id;
    END IF;

    IF dev_id IS NOT NULL THEN
        UPDATE gr33ncore.devices
        SET config_version = config_version + 1, updated_at = NOW()
        WHERE id = dev_id AND deleted_at IS NULL;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_actuators_bump_device_config_version_assign ON gr33ncore.actuators;
CREATE TRIGGER trg_actuators_bump_device_config_version_assign
    AFTER INSERT OR UPDATE OF hardware_identifier, device_id ON gr33ncore.actuators
    FOR EACH ROW
    EXECUTE FUNCTION gr33ncore.bump_device_config_version_from_actuator();
