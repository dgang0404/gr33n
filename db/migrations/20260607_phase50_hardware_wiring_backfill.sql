-- Phase 50 WS1 — backfill demo farm wiring from pi_client/config.yaml mapping (by name, not serial id).
-- Idempotent: only sets wiring when config->'wiring' is absent.

DO $$
DECLARE
    edge_device_id BIGINT;
    flower_device_id BIGINT;
BEGIN
    SELECT id INTO edge_device_id
    FROM gr33ncore.devices
    WHERE farm_id = 1 AND device_uid = 'demo-veg-relay-01'
    LIMIT 1;

    SELECT id INTO flower_device_id
    FROM gr33ncore.devices
    WHERE farm_id = 1 AND device_uid = 'demo-flower-relay-01'
    LIMIT 1;

    -- Sensors (farm 1 demo names ↔ config.yaml sources)
    UPDATE gr33ncore.sensors SET config = config || jsonb_build_object('wiring', jsonb_build_object(
        'source', 'dht22', 'gpio_pin', 4, 'device_id', edge_device_id, 'notes', 'Air Temp Indoor — DHT22 data line'
    ))
    WHERE farm_id = 1 AND name = 'Air Temp Indoor' AND deleted_at IS NULL
      AND (config->'wiring') IS NULL AND edge_device_id IS NOT NULL;

    UPDATE gr33ncore.sensors SET config = config || jsonb_build_object('wiring', jsonb_build_object(
        'source', 'dht22', 'gpio_pin', 4, 'device_id', edge_device_id, 'notes', 'Air Humidity Indoor — shared DHT22'
    ))
    WHERE farm_id = 1 AND name = 'Air Humidity Indoor' AND deleted_at IS NULL
      AND (config->'wiring') IS NULL AND edge_device_id IS NOT NULL;

    UPDATE gr33ncore.sensors SET config = config || jsonb_build_object('wiring', jsonb_build_object(
        'source', 'ads1115', 'i2c_channel', 0, 'device_id', edge_device_id, 'notes', 'Soil Moisture Outdoor — ADS1115 A0'
    ))
    WHERE farm_id = 1 AND name = 'Soil Moisture Outdoor' AND deleted_at IS NULL
      AND (config->'wiring') IS NULL AND edge_device_id IS NOT NULL;

    UPDATE gr33ncore.sensors SET config = config || jsonb_build_object('wiring', jsonb_build_object(
        'source', 'mhz19', 'serial_port', '/dev/ttyS0', 'device_id', edge_device_id, 'notes', 'CO2 Sensor Indoor'
    ))
    WHERE farm_id = 1 AND name = 'CO2 Sensor Indoor' AND deleted_at IS NULL
      AND (config->'wiring') IS NULL AND edge_device_id IS NOT NULL;

    UPDATE gr33ncore.sensors SET config = config || jsonb_build_object('wiring', jsonb_build_object(
        'source', 'ads1115', 'i2c_channel', 1, 'device_id', edge_device_id, 'notes', 'EC Sensor — ADS1115 A1'
    ))
    WHERE farm_id = 1 AND name = 'EC Sensor' AND deleted_at IS NULL
      AND (config->'wiring') IS NULL AND edge_device_id IS NOT NULL;

    UPDATE gr33ncore.sensors SET config = config || jsonb_build_object('wiring', jsonb_build_object(
        'source', 'ads1115', 'i2c_channel', 2, 'device_id', edge_device_id, 'notes', 'pH Sensor — ADS1115 A2'
    ))
    WHERE farm_id = 1 AND name = 'pH Sensor' AND deleted_at IS NULL
      AND (config->'wiring') IS NULL AND edge_device_id IS NOT NULL;

    UPDATE gr33ncore.sensors SET config = config || jsonb_build_object('wiring', jsonb_build_object(
        'source', 'bh1750', 'device_id', edge_device_id, 'notes', 'PAR Sensor Indoor — BH1750 I2C'
    ))
    WHERE farm_id = 1 AND name = 'PAR Sensor Indoor' AND deleted_at IS NULL
      AND (config->'wiring') IS NULL AND edge_device_id IS NOT NULL;

    -- Actuators
    UPDATE gr33ncore.actuators SET config = config || jsonb_build_object('wiring', jsonb_build_object(
        'source', 'gpio_relay', 'gpio_pin', 17, 'device_id', edge_device_id, 'notes', 'Veg Room Grow Light relay'
    ))
    WHERE farm_id = 1 AND name = 'Veg Room Grow Light' AND deleted_at IS NULL
      AND (config->'wiring') IS NULL AND edge_device_id IS NOT NULL;

    UPDATE gr33ncore.actuators SET config = config || jsonb_build_object('wiring', jsonb_build_object(
        'source', 'gpio_relay', 'gpio_pin', 27, 'device_id', flower_device_id, 'notes', 'Flower Room Irrigation Pump relay'
    ))
    WHERE farm_id = 1 AND name = 'Flower Room Irrigation Pump' AND deleted_at IS NULL
      AND (config->'wiring') IS NULL AND flower_device_id IS NOT NULL;
END $$;
