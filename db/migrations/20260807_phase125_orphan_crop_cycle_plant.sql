-- Phase 125 LED bench — active Veg Room crop cycle pointed at soft-deleted
-- plant id 24 (Spinach), causing UI GET /plants/24 → 404 console noise.
-- Clear dead plant_ids; deactivate leftover phase93 smoke cycles on zone 1.

UPDATE gr33nfertigation.crop_cycles cc
SET plant_id = NULL,
    updated_at = NOW()
WHERE cc.plant_id IS NOT NULL
  AND EXISTS (
      SELECT 1 FROM gr33ncrops.plants p
      WHERE p.id = cc.plant_id AND p.deleted_at IS NOT NULL
  );

UPDATE gr33nfertigation.crop_cycles
SET is_active = false,
    updated_at = NOW()
WHERE farm_id = 1
  AND zone_id = 1
  AND is_active = true
  AND name LIKE 'phase93_%';
