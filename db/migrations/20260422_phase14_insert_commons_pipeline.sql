-- Phase 14 WS2: optional human approval queue + bundle export metadata for Insert Commons.

ALTER TABLE gr33ncore.farms
    ADD COLUMN IF NOT EXISTS insert_commons_require_approval BOOLEAN NOT NULL DEFAULT FALSE;

CREATE TABLE IF NOT EXISTS gr33ncore.insert_commons_bundles (
    id                  BIGSERIAL PRIMARY KEY,
    farm_id             BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    idempotency_key     TEXT,
    payload_hash        TEXT        NOT NULL,
    payload             JSONB       NOT NULL,
    status              TEXT        NOT NULL CHECK (status IN (
        'pending_approval', 'approved', 'rejected', 'delivered', 'delivery_failed'
    )),
    reviewer_user_id    UUID REFERENCES gr33ncore.profiles(user_id) ON DELETE SET NULL,
    reviewed_at         TIMESTAMPTZ,
    review_note         TEXT,
    delivery_http_status INT,
    delivery_error      TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_insert_commons_bundles_farm_hash
    ON gr33ncore.insert_commons_bundles (farm_id, payload_hash);

CREATE INDEX IF NOT EXISTS idx_insert_commons_bundles_farm_status_created
    ON gr33ncore.insert_commons_bundles (farm_id, status, created_at DESC);

ALTER TABLE gr33ncore.insert_commons_sync_events
    ADD COLUMN IF NOT EXISTS bundle_id BIGINT REFERENCES gr33ncore.insert_commons_bundles(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_insert_commons_sync_bundle
    ON gr33ncore.insert_commons_sync_events (bundle_id)
    WHERE bundle_id IS NOT NULL;
