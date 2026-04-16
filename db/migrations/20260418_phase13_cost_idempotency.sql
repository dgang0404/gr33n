-- Phase 13: Idempotent cost creation (offline sync / safe retries)

CREATE TABLE IF NOT EXISTS gr33ncore.cost_transaction_idempotency (
    farm_id              BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    idempotency_key      TEXT NOT NULL,
    cost_transaction_id  BIGINT NOT NULL REFERENCES gr33ncore.cost_transactions(id) ON DELETE CASCADE,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (farm_id, idempotency_key)
);

CREATE INDEX IF NOT EXISTS idx_cost_idem_transaction
    ON gr33ncore.cost_transaction_idempotency (cost_transaction_id);
