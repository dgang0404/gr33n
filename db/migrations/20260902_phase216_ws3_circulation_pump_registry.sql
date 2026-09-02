-- Phase 216 WS3 — register circulation_pump in the device type registry so
-- the UI shows the "Run pulse" control (same declarative flag pattern as
-- pump/return_pump/air_pump). Bound to a reservoir_id, not a zone_id.

INSERT INTO gr33ncore.device_type_registry
    (type_key, device_class, plant_need, display_label, supports_pulse, sort_order)
VALUES
    ('circulation_pump', 'actuator', 'water', 'Circulation pump', true, 118)
ON CONFLICT (type_key) DO UPDATE SET
    device_class = EXCLUDED.device_class,
    plant_need = EXCLUDED.plant_need,
    display_label = EXCLUDED.display_label,
    supports_pulse = EXCLUDED.supports_pulse,
    sort_order = EXCLUDED.sort_order;
