-- Phase 93 WS2 — cycle batch label (genetics/room note), not crop identity.

ALTER TABLE gr33nfertigation.crop_cycles
    RENAME COLUMN strain_or_variety TO batch_label;

COMMENT ON COLUMN gr33nfertigation.crop_cycles.batch_label IS
    'Optional batch or genetics label for this grow run (Phase 93). Crop identity is plants.crop_key via plant_id.';
