-- Phase 29 WS3 — Guardian action proposals (propose → confirm).

-- name: InsertGuardianProposal :one
INSERT INTO gr33ncore.guardian_action_proposals (
    user_id, farm_id, session_id, tool_id, args, summary, expires_at
) VALUES ($1, $2, $3, $4, $5::jsonb, $6, $7)
RETURNING *;

-- name: GetGuardianProposalByID :one
SELECT * FROM gr33ncore.guardian_action_proposals WHERE proposal_id = $1;

-- name: ConfirmGuardianProposal :one
UPDATE gr33ncore.guardian_action_proposals
SET status = 'confirmed',
    result = $2::jsonb,
    confirmed_at = NOW()
WHERE proposal_id = $1
  AND user_id = $3
  AND status = 'pending'
  AND expires_at > NOW()
RETURNING *;

-- name: ExpireStaleGuardianProposals :exec
UPDATE gr33ncore.guardian_action_proposals
SET status = 'expired'
WHERE status = 'pending' AND expires_at <= NOW();
