-- Phase 14 WS4: receiver idempotency correlation + pilot stats queries

ALTER TABLE gr33ncore.insert_commons_received_payloads
    ADD COLUMN IF NOT EXISTS source_idempotency_key TEXT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_insert_commons_received_farm_idem
    ON gr33ncore.insert_commons_received_payloads (farm_pseudonym, source_idempotency_key)
    WHERE source_idempotency_key IS NOT NULL;
