-- Phase 125 — Veg Room Irrigation Pump was added in seed (Phase 39) after the
-- Phase 50 wiring backfill, so the UI showed "Not wired yet" while the pump was
-- still bound to demo-veg-relay-01 and accepted pulse/commands. LED bench found it.
-- Idempotent: only sets wiring when config->'wiring' is absent.
-- gpio_pin 20 matches discrete LED sim pixel 6 (scripts/run-led-simulation.sh).

DO $$
DECLARE
    edge_device_id BIGINT;
BEGIN
    SELECT id INTO edge_device_id
    FROM gr33ncore.devices
    WHERE farm_id = 1 AND device_uid = 'demo-veg-relay-01'
    LIMIT 1;

    UPDATE gr33ncore.actuators SET config = config || jsonb_build_object('wiring', jsonb_build_object(
        'source', 'gpio_relay', 'gpio_pin', 20, 'device_id', edge_device_id,
        'notes', 'Veg Room Irrigation Pump relay'
    ))
    WHERE farm_id = 1 AND name = 'Veg Room Irrigation Pump' AND deleted_at IS NULL
      AND (config->'wiring') IS NULL AND edge_device_id IS NOT NULL;
END $$;
