-- Phase 39 WS1 — FIFO device command queue
-- Replaces the single devices.config.pending_command slot so that
-- concurrent writers (automation worker, operator API, Guardian Confirm)
-- can all enqueue safely and the Pi drains them in order.
--
-- Backward compat: devices.config.pending_command continues to mirror
-- the head payload for one Pi-client release so old clients don't lose
-- their command on upgrade day.

BEGIN;

CREATE TABLE IF NOT EXISTS gr33ncore.device_commands (
    id              BIGSERIAL PRIMARY KEY,
    device_id       BIGINT NOT NULL REFERENCES gr33ncore.devices(id) ON DELETE CASCADE,
    farm_id         BIGINT NOT NULL,   -- denorm for auth queries; not FK to avoid cross-schema join cost
    command_type    TEXT NOT NULL CHECK (command_type IN ('actuator', 'pulse', 'mix_batch')),
    payload         JSONB NOT NULL DEFAULT '{}',
    status          TEXT NOT NULL DEFAULT 'pending'
                        CHECK (status IN ('pending', 'in_progress', 'completed', 'failed', 'cancelled')),
    source          TEXT NOT NULL DEFAULT 'operator'
                        CHECK (source IN ('operator', 'schedule', 'rule', 'program', 'guardian')),
    -- nullable provenance links
    actuator_id     BIGINT,
    schedule_id     BIGINT,
    rule_id         BIGINT,
    program_id      BIGINT,
    -- audit
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    result          JSONB        -- Pi ack result payload (optional)
);

-- Ordered head-of-queue lookup: pending commands for one device, oldest first
CREATE INDEX IF NOT EXISTS device_commands_device_pending
    ON gr33ncore.device_commands (device_id, created_at ASC)
    WHERE status = 'pending';

-- Auth / list queries
CREATE INDEX IF NOT EXISTS device_commands_farm
    ON gr33ncore.device_commands (farm_id, created_at DESC);

COMMENT ON TABLE gr33ncore.device_commands IS
    'Phase 39 WS1: FIFO per-device command queue. Replaces single devices.config.pending_command slot.';

COMMIT;
