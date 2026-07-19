-- Phase 183 — animal zones (coop, pasture) had no way to model a gate/door
-- actuator: feeding (feeder_hopper) and watering (pump/water_valve) already
-- existed in the Phase 90 registry, but "open/shut a gate" had no type_key
-- at all, so the UI's actuator-add wizard never offered it.
INSERT INTO gr33ncore.device_type_registry
    (type_key, device_class, plant_need, display_label, supports_pulse, sort_order)
VALUES
    ('gate', 'actuator', 'water', 'Gate', true, 118)
ON CONFLICT (type_key) DO UPDATE SET
    device_class = EXCLUDED.device_class,
    plant_need = EXCLUDED.plant_need,
    display_label = EXCLUDED.display_label,
    supports_pulse = EXCLUDED.supports_pulse,
    sort_order = EXCLUDED.sort_order;
