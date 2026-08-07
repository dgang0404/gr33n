-- Phase 125 LED bench — platform wiring put Veg Room Grow Light on GPIO 17,
-- same BCM as the discrete-sim heartbeat LED (gpio_leds.heartbeat_pin: 17).
-- Align grow light with discrete LED pixel 5 (GPIO 16 / phys 36).
-- Also wire Media Moisture Indoor (LED pixel 0) — was missing from Phase 50 backfill.

DO $$
DECLARE
    edge_device_id BIGINT;
BEGIN
    SELECT id INTO edge_device_id
    FROM gr33ncore.devices
    WHERE farm_id = 1 AND device_uid = 'demo-veg-relay-01'
    LIMIT 1;

    IF edge_device_id IS NULL THEN
        RETURN;
    END IF;

    -- Grow light: 17 → 16 (free GPIO 17 for heartbeat / indicators)
    UPDATE gr33ncore.actuators
    SET config = jsonb_set(
            config,
            '{wiring}',
            jsonb_build_object(
                'source', 'gpio_relay',
                'gpio_pin', 16,
                'device_id', edge_device_id,
                'notes', 'Veg Room Grow Light relay (BCM 16; 17 reserved for heartbeat LED on sim rig)'
            )
        )
    WHERE farm_id = 1
      AND name = 'Veg Room Grow Light'
      AND deleted_at IS NULL
      AND (config->'wiring'->>'gpio_pin') = '17'
      AND COALESCE((config->'wiring'->>'device_id')::bigint, edge_device_id) = edge_device_id;

    -- Media Moisture Indoor — used by LED sim Demo A; was unwired in Phase 50
    UPDATE gr33ncore.sensors
    SET config = COALESCE(config, '{}'::jsonb) || jsonb_build_object(
        'wiring', jsonb_build_object(
            'source', 'ads1115',
            'i2c_channel', 3,
            'device_id', edge_device_id,
            'notes', 'Media Moisture Indoor — ADS1115 A3 (LED sim indicator is discrete GPIO 18)'
        )
    )
    WHERE farm_id = 1
      AND name = 'Media Moisture Indoor'
      AND deleted_at IS NULL
      AND (config->'wiring') IS NULL
      AND edge_device_id IS NOT NULL;
END $$;
