-- Phase 29 WS3 — Guardian action proposals (propose → confirm).
-- Phase 34 — revise/supersede chain + operator-supplied facts in meta.

-- name: InsertGuardianProposal :one
INSERT INTO gr33ncore.guardian_action_proposals (
    user_id, farm_id, session_id, tool_id, args, summary, risk_tier, expires_at,
    meta, supersedes_proposal_id, revision
) VALUES (
    sqlc.arg(user_id), sqlc.arg(farm_id), sqlc.arg(session_id), sqlc.arg(tool_id),
    sqlc.arg(args)::jsonb, sqlc.arg(summary), sqlc.arg(risk_tier), sqlc.arg(expires_at),
    sqlc.arg(meta)::jsonb, sqlc.narg(supersedes_proposal_id), sqlc.arg(revision)
)
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

-- name: SupersedeProposal :one
-- Phase 34 — mark a still-pending proposal as replaced by a later revision.
UPDATE gr33ncore.guardian_action_proposals
SET status = 'superseded'
WHERE proposal_id = $1
  AND user_id = $2
  AND status = 'pending'
RETURNING *;

-- name: GetLatestPendingProposalBySession :one
-- Phase 34 — the one live draft a correction turn should revise.
SELECT * FROM gr33ncore.guardian_action_proposals
WHERE user_id = $1
  AND session_id = sqlc.arg(session_id)
  AND status = 'pending'
  AND expires_at > NOW()
ORDER BY revision DESC, created_at DESC
LIMIT 1;

-- name: GetLatestLiveInChain :one
-- Phase 34 — given any proposal in a chain, return the newest still-pending revision
-- (its live successor), used to point a 410 at the confirmable draft.
WITH RECURSIVE descendants AS (
    SELECT a0.proposal_id, a0.supersedes_proposal_id, a0.status, a0.revision
    FROM gr33ncore.guardian_action_proposals a0
    WHERE a0.proposal_id = $1
    UNION ALL
    SELECT p.proposal_id, p.supersedes_proposal_id, p.status, p.revision
    FROM gr33ncore.guardian_action_proposals p
    JOIN descendants d ON p.supersedes_proposal_id = d.proposal_id
)
SELECT g.* FROM gr33ncore.guardian_action_proposals g
JOIN descendants d2 ON d2.proposal_id = g.proposal_id
WHERE g.status = 'pending'
ORDER BY g.revision DESC
LIMIT 1;

-- name: GetProposalChainRoot :one
-- Phase 34 — walk parent pointers to the first draft (revision 1) for audit lineage.
WITH RECURSIVE ancestors AS (
    SELECT a0.proposal_id, a0.supersedes_proposal_id, a0.revision
    FROM gr33ncore.guardian_action_proposals a0
    WHERE a0.proposal_id = $1
    UNION ALL
    SELECT p.proposal_id, p.supersedes_proposal_id, p.revision
    FROM gr33ncore.guardian_action_proposals p
    JOIN ancestors a ON a.supersedes_proposal_id = p.proposal_id
)
SELECT ancestors.proposal_id FROM ancestors ORDER BY ancestors.revision ASC LIMIT 1;

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

-- name: DismissGuardianProposal :one
UPDATE gr33ncore.guardian_action_proposals
SET status = 'dismissed'
WHERE proposal_id = $1
  AND user_id = $2
  AND status = 'pending'
RETURNING *;
