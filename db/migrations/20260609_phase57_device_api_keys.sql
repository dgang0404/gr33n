-- Phase 57 WS1 — per-device Pi API keys (bcrypt hash only; plaintext shown once at issue).

CREATE TABLE IF NOT EXISTS gr33ncore.device_api_keys (
    id            BIGSERIAL PRIMARY KEY,
    device_id     BIGINT NOT NULL REFERENCES gr33ncore.devices(id) ON DELETE CASCADE,
    key_hash      TEXT NOT NULL,
    label         TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at    TIMESTAMPTZ,
    last_used_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_device_api_keys_device_active
    ON gr33ncore.device_api_keys (device_id)
    WHERE revoked_at IS NULL;

COMMENT ON TABLE gr33ncore.device_api_keys IS
    'Phase 57 — per-edge-device credentials; Pi sends gdev_{device_id}_{secret} via X-Device-Key.';
