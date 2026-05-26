-- Phase 29 WS3 — frozen Guardian action proposals (propose → confirm).

DO $$ BEGIN
    CREATE TYPE gr33ncore.guardian_proposal_status_enum AS ENUM (
        'pending', 'confirmed', 'dismissed', 'expired'
    );
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS gr33ncore.guardian_action_proposals (
    proposal_id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    farm_id       BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    session_id    UUID NULL,
    tool_id       TEXT NOT NULL,
    args          JSONB NOT NULL DEFAULT '{}'::jsonb,
    summary       TEXT NOT NULL,
    status        gr33ncore.guardian_proposal_status_enum NOT NULL DEFAULT 'pending',
    result        JSONB NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at    TIMESTAMPTZ NOT NULL,
    confirmed_at  TIMESTAMPTZ NULL
);

COMMENT ON TABLE gr33ncore.guardian_action_proposals IS
  'Phase 29 — server-side frozen tool calls proposed by Farm Guardian; confirm replays stored args.';

CREATE INDEX IF NOT EXISTS idx_guardian_proposals_user_status
    ON gr33ncore.guardian_action_proposals (user_id, status, expires_at DESC);

CREATE INDEX IF NOT EXISTS idx_guardian_proposals_farm
    ON gr33ncore.guardian_action_proposals (farm_id, created_at DESC);
