-- Phase 30 WS2 — risk tier on Guardian action proposals.

ALTER TABLE gr33ncore.guardian_action_proposals
    ADD COLUMN IF NOT EXISTS risk_tier TEXT NOT NULL DEFAULT 'medium'
        CHECK (risk_tier IN ('low', 'medium', 'high'));

COMMENT ON COLUMN gr33ncore.guardian_action_proposals.risk_tier IS
  'Operator-facing impact tier: low (read/ack), medium (config/tasks), high (bootstrap, actuators, disable rules).';

UPDATE gr33ncore.guardian_action_proposals
SET risk_tier = 'low'
WHERE tool_id IN ('mark_alert_read', 'ack_alert');

UPDATE gr33ncore.guardian_action_proposals
SET risk_tier = 'high'
WHERE tool_id = 'apply_bootstrap_template';
