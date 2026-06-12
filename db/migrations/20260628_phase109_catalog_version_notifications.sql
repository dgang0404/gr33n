-- Phase 109 — platform catalog version tracking + farm last-seen for admin notifications.

CREATE TABLE IF NOT EXISTS gr33ncore.platform_catalog_state (
    id               SMALLINT    PRIMARY KEY CHECK (id = 1),
    catalog_version  INTEGER     NOT NULL DEFAULT 1,
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO gr33ncore.platform_catalog_state (id, catalog_version)
SELECT 1, COALESCE(MAX(catalog_version), 1)
FROM gr33ncrops.crop_catalog_entries
ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS gr33ncore.farm_catalog_version_seen (
    farm_id              BIGINT      PRIMARY KEY REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    catalog_version_seen INTEGER     NOT NULL DEFAULT 0,
    notified_at          TIMESTAMPTZ,
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_farm_catalog_version_seen_version
    ON gr33ncore.farm_catalog_version_seen (catalog_version_seen);
