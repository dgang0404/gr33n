-- Phase 90 — Platform device type registry (sensor/actuator → plant need, labels, GH roles).

CREATE TABLE IF NOT EXISTS gr33ncore.device_type_registry (
    type_key        TEXT PRIMARY KEY,
    device_class    TEXT NOT NULL CHECK (device_class IN ('sensor', 'actuator')),
    plant_need      TEXT NOT NULL CHECK (plant_need IN ('water', 'light', 'air')),
    display_label   TEXT NOT NULL,
    supports_pulse  BOOLEAN NOT NULL DEFAULT false,
    gh_role         TEXT CHECK (gh_role IS NULL OR gh_role IN ('shade', 'vent', 'fan')),
    wiring_sources  JSONB NOT NULL DEFAULT '[]'::jsonb,
    sort_order      INT NOT NULL DEFAULT 0
);

-- Sensors — water
INSERT INTO gr33ncore.device_type_registry
    (type_key, device_class, plant_need, display_label, wiring_sources, sort_order)
VALUES
    ('soil_moisture', 'sensor', 'water', 'Soil moisture', '["ads1115"]', 10),
    ('moisture', 'sensor', 'water', 'Moisture', '["ads1115"]', 11),
    ('ec', 'sensor', 'water', 'EC', '["ads1115"]', 12),
    ('ph', 'sensor', 'water', 'pH', '["ads1115"]', 13),
    ('water_level', 'sensor', 'water', 'Water level', '["gpio_digital"]', 14),
    ('flow_rate', 'sensor', 'water', 'Flow rate', '["gpio_digital"]', 15),
    ('water_temp', 'sensor', 'water', 'Water temperature', '["dht22"]', 16),
    ('dissolved_oxygen', 'sensor', 'water', 'Dissolved oxygen', '["ads1115"]', 17),
    ('tds', 'sensor', 'water', 'TDS', '["ads1115"]', 18)
ON CONFLICT (type_key) DO UPDATE SET
    device_class = EXCLUDED.device_class,
    plant_need = EXCLUDED.plant_need,
    display_label = EXCLUDED.display_label,
    wiring_sources = EXCLUDED.wiring_sources,
    sort_order = EXCLUDED.sort_order;

-- Sensors — light
INSERT INTO gr33ncore.device_type_registry
    (type_key, device_class, plant_need, display_label, wiring_sources, sort_order)
VALUES
    ('lux', 'sensor', 'light', 'Light level', '["bh1750"]', 30),
    ('par', 'sensor', 'light', 'PAR', '["bh1750"]', 31),
    ('par_umol', 'sensor', 'light', 'PAR', '["bh1750"]', 32),
    ('ppfd', 'sensor', 'light', 'PPFD', '["bh1750"]', 33),
    ('light_level', 'sensor', 'light', 'Light level', '["bh1750"]', 34)
ON CONFLICT (type_key) DO UPDATE SET
    device_class = EXCLUDED.device_class,
    plant_need = EXCLUDED.plant_need,
    display_label = EXCLUDED.display_label,
    wiring_sources = EXCLUDED.wiring_sources,
    sort_order = EXCLUDED.sort_order;

-- Sensors — air / climate
INSERT INTO gr33ncore.device_type_registry
    (type_key, device_class, plant_need, display_label, wiring_sources, sort_order)
VALUES
    ('air_temp', 'sensor', 'air', 'Air temperature', '["dht22"]', 50),
    ('temperature', 'sensor', 'air', 'Temperature', '["dht22"]', 51),
    ('temp', 'sensor', 'air', 'Temperature', '["dht22"]', 52),
    ('temp_f', 'sensor', 'air', 'Temperature (°F)', '["dht22"]', 53),
    ('humidity', 'sensor', 'air', 'Humidity', '["dht22"]', 54),
    ('rh', 'sensor', 'air', 'Humidity', '["dht22"]', 55),
    ('co2', 'sensor', 'air', 'CO₂', '["mhz19"]', 56),
    ('vpd', 'sensor', 'air', 'VPD', '["derived"]', 57),
    ('dew_point', 'sensor', 'air', 'Dew point', '["derived"]', 58),
    ('barometric_pressure', 'sensor', 'air', 'Barometric pressure', '["derived"]', 59),
    ('pressure', 'sensor', 'air', 'Pressure', '["derived"]', 60)
ON CONFLICT (type_key) DO UPDATE SET
    device_class = EXCLUDED.device_class,
    plant_need = EXCLUDED.plant_need,
    display_label = EXCLUDED.display_label,
    wiring_sources = EXCLUDED.wiring_sources,
    sort_order = EXCLUDED.sort_order;

-- Actuators — water
INSERT INTO gr33ncore.device_type_registry
    (type_key, device_class, plant_need, display_label, supports_pulse, sort_order)
VALUES
    ('pump', 'actuator', 'water', 'Pump', true, 110),
    ('water_valve', 'actuator', 'water', 'Water valve', true, 111),
    ('return_pump', 'actuator', 'water', 'Return pump', true, 112),
    ('irrigation', 'actuator', 'water', 'Irrigation', false, 113),
    ('drip', 'actuator', 'water', 'Drip', false, 114),
    ('feeder_hopper', 'actuator', 'water', 'Feeder hopper', true, 115),
    ('relay', 'actuator', 'water', 'Relay', true, 116),
    ('air_pump', 'actuator', 'water', 'Air pump', true, 117)
ON CONFLICT (type_key) DO UPDATE SET
    device_class = EXCLUDED.device_class,
    plant_need = EXCLUDED.plant_need,
    display_label = EXCLUDED.display_label,
    supports_pulse = EXCLUDED.supports_pulse,
    sort_order = EXCLUDED.sort_order;

-- Actuators — light
INSERT INTO gr33ncore.device_type_registry
    (type_key, device_class, plant_need, display_label, sort_order)
VALUES
    ('light', 'actuator', 'light', 'Light', 130),
    ('grow_light', 'actuator', 'light', 'Grow light', 131)
ON CONFLICT (type_key) DO UPDATE SET
    device_class = EXCLUDED.device_class,
    plant_need = EXCLUDED.plant_need,
    display_label = EXCLUDED.display_label,
    sort_order = EXCLUDED.sort_order;

-- Actuators — air / climate (+ greenhouse roles)
INSERT INTO gr33ncore.device_type_registry
    (type_key, device_class, plant_need, display_label, gh_role, sort_order)
VALUES
    ('exhaust_fan', 'actuator', 'air', 'Exhaust fan', 'fan', 150),
    ('circulation_fan', 'actuator', 'air', 'Circulation fan', 'fan', 151),
    ('fan', 'actuator', 'air', 'Fan', 'fan', 152),
    ('ridge_vent', 'actuator', 'air', 'Ridge vent', 'vent', 153),
    ('glazing_panel', 'actuator', 'air', 'Glazing panel', 'vent', 154),
    ('shade_screen', 'actuator', 'air', 'Shade screen', 'shade', 155),
    ('shade_cloth_motor', 'actuator', 'air', 'Shade cloth motor', 'shade', 156),
    ('shade', 'actuator', 'air', 'Shade', 'shade', 157),
    ('vent', 'actuator', 'air', 'Vent', 'vent', 158),
    ('humidifier', 'actuator', 'air', 'Humidifier', NULL, 159),
    ('dehumidifier', 'actuator', 'air', 'Dehumidifier', NULL, 160),
    ('co2_injector', 'actuator', 'air', 'CO₂ injector', NULL, 161),
    ('heat_lamp', 'actuator', 'air', 'Heat lamp', NULL, 162),
    ('heater', 'actuator', 'air', 'Heater', NULL, 163),
    ('cooler', 'actuator', 'air', 'Cooler', NULL, 164)
ON CONFLICT (type_key) DO UPDATE SET
    device_class = EXCLUDED.device_class,
    plant_need = EXCLUDED.plant_need,
    display_label = EXCLUDED.display_label,
    gh_role = EXCLUDED.gh_role,
    sort_order = EXCLUDED.sort_order;
