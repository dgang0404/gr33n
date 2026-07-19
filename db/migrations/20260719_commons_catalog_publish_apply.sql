-- Commons catalog: user publish provenance + auto-apply on import (Phase 207)

ALTER TABLE gr33ncore.commons_catalog_entries
    ADD COLUMN IF NOT EXISTS published_by_user_id UUID REFERENCES gr33ncore.profiles(user_id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS source_farm_id BIGINT REFERENCES gr33ncore.farms(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_commons_catalog_published_by
    ON gr33ncore.commons_catalog_entries (published_by_user_id)
    WHERE published_by_user_id IS NOT NULL;

COMMENT ON COLUMN gr33ncore.commons_catalog_entries.published_by_user_id IS
  'User who published this pack via POST /commons/catalog (NULL for migration-seeded packs).';
COMMENT ON COLUMN gr33ncore.commons_catalog_entries.source_farm_id IS
  'Optional farm that exported this pack (recipe export from farm programs).';
