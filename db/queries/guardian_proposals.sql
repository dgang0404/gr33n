-- Phase 29 WS3 — Guardian action proposals (propose → confirm).

-- name: InsertGuardianProposal :one
INSERT INTO gr33ncore.guardian_action_proposals (
    user_id, farm_id, session_id, tool_id, args, summary, risk_tier, expires_at
) VALUES ($1, $2, $3, $4, $5::jsonb, $6, $7, $8)
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

-- name: ListGuardianProposalsByUser :many
SELECT * FROM gr33ncore.guardian_action_proposals
WHERE user_id = $1
  AND (sqlc.narg('farm_id')::bigint IS NULL OR farm_id = sqlc.narg('farm_id')::bigint)
  AND (sqlc.narg('status')::text IS NULL OR status::text = sqlc.narg('status')::text)
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountGuardianProposalsByUser :one
SELECT COUNT(*)::bigint FROM gr33ncore.guardian_action_proposals
WHERE user_id = $1
  AND (sqlc.narg('farm_id')::bigint IS NULL OR farm_id = sqlc.narg('farm_id')::bigint)
  AND (sqlc.narg('status')::text IS NULL OR status::text = sqlc.narg('status')::text);
