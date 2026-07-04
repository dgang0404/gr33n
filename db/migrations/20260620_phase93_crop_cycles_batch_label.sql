-- Phase 93 WS2 — cycle batch label (genetics/room note), not crop identity.
-- Fresh schema (v2) may already define batch_label; skip rename when strain_or_variety is absent.

DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'gr33nfertigation' AND table_name = 'crop_cycles' AND column_name = 'strain_or_variety'
  ) THEN
    ALTER TABLE gr33nfertigation.crop_cycles
      RENAME COLUMN strain_or_variety TO batch_label;
  END IF;
END $$;

COMMENT ON COLUMN gr33nfertigation.crop_cycles.batch_label IS
    'Optional batch or genetics label for this grow run (Phase 93). Crop identity is plants.crop_key via plant_id.';
