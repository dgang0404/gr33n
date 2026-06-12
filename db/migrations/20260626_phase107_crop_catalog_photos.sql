-- Phase 107 — optional catalog thumbnails for picker / commons API.

ALTER TABLE gr33ncrops.crop_catalog_entries
    ADD COLUMN IF NOT EXISTS image_url TEXT;

UPDATE gr33ncrops.crop_catalog_entries
SET image_url = '/assets/crops/' || crop_key || '.svg'
WHERE crop_key IN ('san_pedro', 'succulent', 'phalaenopsis', 'chrysanthemum', 'rose');
