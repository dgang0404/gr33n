-- Phase 48 — trim farm 1 to small_indoor profile after idempotent master_seed.
-- Run via: ./scripts/dev-reset-farm.sh --profile small_indoor

SELECT gr33ncore.set_dev_seed_profile(1, 'small_indoor');

-- Hide outdoor zone for sit-in / daily dev (data retained, soft-deleted).
UPDATE gr33ncore.zones
SET deleted_at = NOW(), updated_at = NOW()
WHERE farm_id = 1
  AND name = 'Outdoor Garden'
  AND deleted_at IS NULL;

-- Trim sensors not needed for two-room indoor dev.
UPDATE gr33ncore.sensors
SET deleted_at = NOW(), updated_at = NOW()
WHERE farm_id = 1
  AND deleted_at IS NULL
  AND name IN ('Soil Moisture Outdoor', 'CO2 Sensor Indoor');
