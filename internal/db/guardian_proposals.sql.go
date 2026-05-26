// Phase 29 WS3 — Guardian action proposals. Hand-written binding (sqlc
// regen blocked by pre-existing conversation_turns.sql ambiguity).

package db

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Gr33ncoreGuardianProposalStatusEnum string

const (
	Gr33ncoreGuardianProposalStatusEnumPending    Gr33ncoreGuardianProposalStatusEnum = "pending"
	Gr33ncoreGuardianProposalStatusEnumConfirmed  Gr33ncoreGuardianProposalStatusEnum = "confirmed"
	Gr33ncoreGuardianProposalStatusEnumDismissed  Gr33ncoreGuardianProposalStatusEnum = "dismissed"
	Gr33ncoreGuardianProposalStatusEnumExpired    Gr33ncoreGuardianProposalStatusEnum = "expired"
)

type Gr33ncoreGuardianActionProposal struct {
	ProposalID  uuid.UUID                             `db:"proposal_id" json:"proposal_id"`
	UserID      uuid.UUID                             `db:"user_id" json:"user_id"`
	FarmID      int64                                 `db:"farm_id" json:"farm_id"`
	SessionID   *uuid.UUID                            `db:"session_id" json:"session_id,omitempty"`
	ToolID      string                                `db:"tool_id" json:"tool_id"`
	Args        json.RawMessage                       `db:"args" json:"args"`
	Summary     string                                `db:"summary" json:"summary"`
	Status      Gr33ncoreGuardianProposalStatusEnum   `db:"status" json:"status"`
	Result      json.RawMessage                       `db:"result" json:"result,omitempty"`
	CreatedAt   time.Time                             `db:"created_at" json:"created_at"`
	ExpiresAt   time.Time                             `db:"expires_at" json:"expires_at"`
	ConfirmedAt *time.Time                            `db:"confirmed_at" json:"confirmed_at,omitempty"`
}

const insertGuardianProposal = `-- name: InsertGuardianProposal :one
INSERT INTO gr33ncore.guardian_action_proposals (
    user_id, farm_id, session_id, tool_id, args, summary, expires_at
) VALUES ($1, $2, $3, $4, $5::jsonb, $6, $7)
RETURNING proposal_id, user_id, farm_id, session_id, tool_id, args, summary, status, result, created_at, expires_at, confirmed_at
`

type InsertGuardianProposalParams struct {
	UserID    uuid.UUID       `db:"user_id" json:"user_id"`
	FarmID    int64           `db:"farm_id" json:"farm_id"`
	SessionID *uuid.UUID      `db:"session_id" json:"session_id"`
	ToolID    string          `db:"tool_id" json:"tool_id"`
	Args      json.RawMessage `db:"args" json:"args"`
	Summary   string          `db:"summary" json:"summary"`
	ExpiresAt time.Time       `db:"expires_at" json:"expires_at"`
}

func scanGuardianProposal(row pgx.Row) (Gr33ncoreGuardianActionProposal, error) {
	var i Gr33ncoreGuardianActionProposal
	err := row.Scan(
		&i.ProposalID,
		&i.UserID,
		&i.FarmID,
		&i.SessionID,
		&i.ToolID,
		&i.Args,
		&i.Summary,
		&i.Status,
		&i.Result,
		&i.CreatedAt,
		&i.ExpiresAt,
		&i.ConfirmedAt,
	)
	return i, err
}

func (q *Queries) InsertGuardianProposal(ctx context.Context, arg InsertGuardianProposalParams) (Gr33ncoreGuardianActionProposal, error) {
	row := q.db.QueryRow(ctx, insertGuardianProposal,
		arg.UserID, arg.FarmID, arg.SessionID, arg.ToolID, arg.Args, arg.Summary, arg.ExpiresAt,
	)
	return scanGuardianProposal(row)
}

const getGuardianProposalByID = `-- name: GetGuardianProposalByID :one
SELECT proposal_id, user_id, farm_id, session_id, tool_id, args, summary, status, result, created_at, expires_at, confirmed_at
FROM gr33ncore.guardian_action_proposals WHERE proposal_id = $1
`

func (q *Queries) GetGuardianProposalByID(ctx context.Context, proposalID uuid.UUID) (Gr33ncoreGuardianActionProposal, error) {
	row := q.db.QueryRow(ctx, getGuardianProposalByID, proposalID)
	return scanGuardianProposal(row)
}

const confirmGuardianProposal = `-- name: ConfirmGuardianProposal :one
UPDATE gr33ncore.guardian_action_proposals
SET status = 'confirmed',
    result = $2::jsonb,
    confirmed_at = NOW()
WHERE proposal_id = $1
  AND user_id = $3
  AND status = 'pending'
  AND expires_at > NOW()
RETURNING proposal_id, user_id, farm_id, session_id, tool_id, args, summary, status, result, created_at, expires_at, confirmed_at
`

type ConfirmGuardianProposalParams struct {
	ProposalID uuid.UUID       `db:"proposal_id" json:"proposal_id"`
	Result     json.RawMessage `db:"result" json:"result"`
	UserID     uuid.UUID       `db:"user_id" json:"user_id"`
}

func (q *Queries) ConfirmGuardianProposal(ctx context.Context, arg ConfirmGuardianProposalParams) (Gr33ncoreGuardianActionProposal, error) {
	row := q.db.QueryRow(ctx, confirmGuardianProposal, arg.ProposalID, arg.Result, arg.UserID)
	return scanGuardianProposal(row)
}

const expireStaleGuardianProposals = `-- name: ExpireStaleGuardianProposals :exec
UPDATE gr33ncore.guardian_action_proposals
SET status = 'expired'
WHERE status = 'pending' AND expires_at <= NOW()
`

func (q *Queries) ExpireStaleGuardianProposals(ctx context.Context) error {
	_, err := q.db.Exec(ctx, expireStaleGuardianProposals)
	return err
}
