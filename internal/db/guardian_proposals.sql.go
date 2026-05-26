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
	Gr33ncoreGuardianProposalStatusEnumPending   Gr33ncoreGuardianProposalStatusEnum = "pending"
	Gr33ncoreGuardianProposalStatusEnumConfirmed Gr33ncoreGuardianProposalStatusEnum = "confirmed"
	Gr33ncoreGuardianProposalStatusEnumDismissed Gr33ncoreGuardianProposalStatusEnum = "dismissed"
	Gr33ncoreGuardianProposalStatusEnumExpired   Gr33ncoreGuardianProposalStatusEnum = "expired"
)

type Gr33ncoreGuardianActionProposal struct {
	ProposalID  uuid.UUID                           `db:"proposal_id" json:"proposal_id"`
	UserID      uuid.UUID                           `db:"user_id" json:"user_id"`
	FarmID      int64                               `db:"farm_id" json:"farm_id"`
	SessionID   *uuid.UUID                          `db:"session_id" json:"session_id,omitempty"`
	ToolID      string                              `db:"tool_id" json:"tool_id"`
	Args        json.RawMessage                     `db:"args" json:"args"`
	Summary     string                              `db:"summary" json:"summary"`
	RiskTier    string                              `db:"risk_tier" json:"risk_tier"`
	Status      Gr33ncoreGuardianProposalStatusEnum `db:"status" json:"status"`
	Result      json.RawMessage                     `db:"result" json:"result,omitempty"`
	CreatedAt   time.Time                           `db:"created_at" json:"created_at"`
	ExpiresAt   time.Time                           `db:"expires_at" json:"expires_at"`
	ConfirmedAt *time.Time                          `db:"confirmed_at" json:"confirmed_at,omitempty"`
}

const insertGuardianProposal = `-- name: InsertGuardianProposal :one
INSERT INTO gr33ncore.guardian_action_proposals (
    user_id, farm_id, session_id, tool_id, args, summary, risk_tier, expires_at
) VALUES ($1, $2, $3, $4, $5::jsonb, $6, $7, $8)
RETURNING proposal_id, user_id, farm_id, session_id, tool_id, args, summary, risk_tier, status, result, created_at, expires_at, confirmed_at
`

type InsertGuardianProposalParams struct {
	UserID    uuid.UUID       `db:"user_id" json:"user_id"`
	FarmID    int64           `db:"farm_id" json:"farm_id"`
	SessionID *uuid.UUID      `db:"session_id" json:"session_id"`
	ToolID    string          `db:"tool_id" json:"tool_id"`
	Args      json.RawMessage `db:"args" json:"args"`
	Summary   string          `db:"summary" json:"summary"`
	RiskTier  string          `db:"risk_tier" json:"risk_tier"`
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
		&i.RiskTier,
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
		arg.UserID, arg.FarmID, arg.SessionID, arg.ToolID, arg.Args, arg.Summary, arg.RiskTier, arg.ExpiresAt,
	)
	return scanGuardianProposal(row)
}

const getGuardianProposalByID = `-- name: GetGuardianProposalByID :one
SELECT proposal_id, user_id, farm_id, session_id, tool_id, args, summary, risk_tier, status, result, created_at, expires_at, confirmed_at
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
RETURNING proposal_id, user_id, farm_id, session_id, tool_id, args, summary, risk_tier, status, result, created_at, expires_at, confirmed_at
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

const listGuardianProposalsByUser = `-- name: ListGuardianProposalsByUser :many
SELECT proposal_id, user_id, farm_id, session_id, tool_id, args, summary, risk_tier, status, result, created_at, expires_at, confirmed_at
FROM gr33ncore.guardian_action_proposals
WHERE user_id = $1
  AND ($2::bigint IS NULL OR farm_id = $2::bigint)
  AND ($3::text IS NULL OR status::text = $3::text)
ORDER BY created_at DESC
LIMIT $4 OFFSET $5
`

type ListGuardianProposalsByUserParams struct {
	UserID uuid.UUID `db:"user_id" json:"user_id"`
	FarmID *int64    `db:"farm_id" json:"farm_id"`
	Status *string   `db:"status" json:"status"`
	Limit  int32     `db:"limit" json:"limit"`
	Offset int32     `db:"offset" json:"offset"`
}

func (q *Queries) ListGuardianProposalsByUser(ctx context.Context, arg ListGuardianProposalsByUserParams) ([]Gr33ncoreGuardianActionProposal, error) {
	rows, err := q.db.Query(ctx, listGuardianProposalsByUser,
		arg.UserID, arg.FarmID, arg.Status, arg.Limit, arg.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Gr33ncoreGuardianActionProposal
	for rows.Next() {
		i, err := scanGuardianProposal(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

const countGuardianProposalsByUser = `-- name: CountGuardianProposalsByUser :one
SELECT COUNT(*)::bigint FROM gr33ncore.guardian_action_proposals
WHERE user_id = $1
  AND ($2::bigint IS NULL OR farm_id = $2::bigint)
  AND ($3::text IS NULL OR status::text = $3::text)
`

type CountGuardianProposalsByUserParams struct {
	UserID uuid.UUID `db:"user_id" json:"user_id"`
	FarmID *int64    `db:"farm_id" json:"farm_id"`
	Status *string   `db:"status" json:"status"`
}

func (q *Queries) CountGuardianProposalsByUser(ctx context.Context, arg CountGuardianProposalsByUserParams) (int64, error) {
	row := q.db.QueryRow(ctx, countGuardianProposalsByUser, arg.UserID, arg.FarmID, arg.Status)
	var n int64
	err := row.Scan(&n)
	return n, err
}
