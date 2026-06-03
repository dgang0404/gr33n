-- Phase 34 WS1 — Guardian PR iteration: revise/supersede a pending proposal.
-- A revised proposal is a NEW frozen row that supersedes its parent; only the
-- latest pending row in a chain is confirmable. operator_provided facts live in
-- meta and are never merged into args as if they were measurements.

-- New status value for proposals replaced by a later revision.
ALTER TYPE gr33ncore.guardian_proposal_status_enum ADD VALUE IF NOT EXISTS 'superseded';

ALTER TABLE gr33ncore.guardian_action_proposals
    ADD COLUMN IF NOT EXISTS supersedes_proposal_id UUID NULL
        REFERENCES gr33ncore.guardian_action_proposals (proposal_id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS revision INT NOT NULL DEFAULT 1,
    ADD COLUMN IF NOT EXISTS meta JSONB NOT NULL DEFAULT '{}'::jsonb;

COMMENT ON COLUMN gr33ncore.guardian_action_proposals.supersedes_proposal_id IS
  'Phase 34 — parent proposal this revision replaces (NULL for the first draft in a chain).';
COMMENT ON COLUMN gr33ncore.guardian_action_proposals.revision IS
  'Phase 34 — 1-based revision number within a supersede chain.';
COMMENT ON COLUMN gr33ncore.guardian_action_proposals.meta IS
  'Phase 34 — non-arg proposal metadata, e.g. operator_provided[] facts Guardian cannot sense.';

-- "latest live draft in this session" lookup for the revise router.
CREATE INDEX IF NOT EXISTS idx_guardian_proposals_session_live
    ON gr33ncore.guardian_action_proposals (user_id, session_id, status, revision DESC)
    WHERE session_id IS NOT NULL;

-- Walk a supersede chain by parent pointer.
CREATE INDEX IF NOT EXISTS idx_guardian_proposals_supersedes
    ON gr33ncore.guardian_action_proposals (supersedes_proposal_id)
    WHERE supersedes_proposal_id IS NOT NULL;
