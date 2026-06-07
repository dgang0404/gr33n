-- Phase 48 — restore full showcase footprint on farm 1 (undo small_indoor trim).
SELECT gr33ncore.set_dev_seed_profile(1, 'demo_showcase');

UPDATE gr33ncore.zones
SET deleted_at = NULL, updated_at = NOW()
WHERE farm_id = 1 AND name = 'Outdoor Garden';

UPDATE gr33ncore.sensors
SET deleted_at = NULL, updated_at = NOW()
WHERE farm_id = 1
  AND name IN ('Soil Moisture Outdoor', 'CO2 Sensor Indoor');
