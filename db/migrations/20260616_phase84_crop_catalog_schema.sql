-- Phase 84 WS-B — platform crop catalog + agronomy field guides (enterprise DB source of truth).
-- Seed data: db/migrations/20260616_phase84_crop_catalog_seed.sql (generated).

CREATE TABLE IF NOT EXISTS gr33ncrops.crop_catalog_entries (
    crop_key            TEXT PRIMARY KEY,
    display_name        TEXT NOT NULL,
    supported           BOOLEAN NOT NULL DEFAULT TRUE,
    category            TEXT,
    source              TEXT,
    substrate           TEXT,
    watering_style      TEXT,
    runoff_pct_target   TEXT,
    moisture_guidance   TEXT,
    cousin_of           TEXT REFERENCES gr33ncrops.crop_catalog_entries (crop_key),
    unsupported_reason  TEXT,
    catalog_version     INTEGER NOT NULL DEFAULT 1,
    meta                JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT crop_catalog_unsupported_reason_chk CHECK (
        (supported = TRUE AND unsupported_reason IS NULL)
        OR (supported = FALSE AND unsupported_reason IS NOT NULL)
    )
);

CREATE INDEX IF NOT EXISTS idx_crop_catalog_entries_supported
    ON gr33ncrops.crop_catalog_entries (supported);

CREATE INDEX IF NOT EXISTS idx_crop_catalog_entries_category
    ON gr33ncrops.crop_catalog_entries (category)
    WHERE category IS NOT NULL;

CREATE TABLE IF NOT EXISTS gr33ncrops.crop_catalog_aliases (
    alias     TEXT PRIMARY KEY,
    crop_key  TEXT NOT NULL REFERENCES gr33ncrops.crop_catalog_entries (crop_key) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_crop_catalog_aliases_crop_key
    ON gr33ncrops.crop_catalog_aliases (crop_key);

CREATE TABLE IF NOT EXISTS gr33ncrops.agronomy_field_guides (
    id              BIGSERIAL PRIMARY KEY,
    slug            TEXT NOT NULL UNIQUE,
    title           TEXT NOT NULL,
    crop_key        TEXT REFERENCES gr33ncrops.crop_catalog_entries (crop_key),
    guide_kind      TEXT NOT NULL DEFAULT 'crop_nutrition',
    domain          TEXT,
    safety_tier     TEXT NOT NULL DEFAULT 'safe',
    body_md         TEXT NOT NULL,
    catalog_version INTEGER NOT NULL DEFAULT 1,
    published       BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order      INTEGER NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_agronomy_field_guides_crop_key
    ON gr33ncrops.agronomy_field_guides (crop_key)
    WHERE crop_key IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_agronomy_field_guides_kind
    ON gr33ncrops.agronomy_field_guides (guide_kind);

DROP TRIGGER IF EXISTS trg_crop_catalog_entries_updated_at ON gr33ncrops.crop_catalog_entries;
CREATE TRIGGER trg_crop_catalog_entries_updated_at
    BEFORE UPDATE ON gr33ncrops.crop_catalog_entries
    FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();

DROP TRIGGER IF EXISTS trg_agronomy_field_guides_updated_at ON gr33ncrops.agronomy_field_guides;
CREATE TRIGGER trg_agronomy_field_guides_updated_at
    BEFORE UPDATE ON gr33ncrops.agronomy_field_guides
    FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();
