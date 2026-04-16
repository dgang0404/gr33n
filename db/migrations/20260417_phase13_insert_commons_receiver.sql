-- Phase 13: optional Insert Commons receiver (pilot ingest persistence)

CREATE TABLE IF NOT EXISTS gr33ncore.insert_commons_received_payloads (
    id               BIGSERIAL PRIMARY KEY,
    received_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    payload_hash     TEXT        NOT NULL,
    farm_pseudonym   TEXT        NOT NULL,
    schema_version   TEXT        NOT NULL,
    generated_at     TIMESTAMPTZ NOT NULL,
    payload          JSONB       NOT NULL,
    CONSTRAINT uq_insert_commons_received_payload_hash UNIQUE (payload_hash)
);

CREATE INDEX IF NOT EXISTS idx_insert_commons_received_received_at
    ON gr33ncore.insert_commons_received_payloads (received_at DESC);
