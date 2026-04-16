-- Phase 12: Insert Commons federation (farm-side sender persistence + history)

ALTER TABLE gr33ncore.farms
    ADD COLUMN IF NOT EXISTS insert_commons_last_attempt_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS insert_commons_last_delivery_status TEXT,
    ADD COLUMN IF NOT EXISTS insert_commons_last_error TEXT,
    ADD COLUMN IF NOT EXISTS insert_commons_backoff_until TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS insert_commons_consecutive_failures INT NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS gr33ncore.insert_commons_sync_events (
    id               BIGSERIAL PRIMARY KEY,
    farm_id          BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    idempotency_key  TEXT,
    status           TEXT NOT NULL,
    http_status      INT,
    error            TEXT,
    payload          JSONB NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_insert_commons_sync_farm_idem
    ON gr33ncore.insert_commons_sync_events (farm_id, idempotency_key)
    WHERE idempotency_key IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_insert_commons_sync_farm_created
    ON gr33ncore.insert_commons_sync_events (farm_id, created_at DESC);
