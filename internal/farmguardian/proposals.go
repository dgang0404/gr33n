package farmguardian

import (
	"context"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmguardian/tools"
)

// ProposalTTL is how long a frozen proposal stays confirmable.
const ProposalTTL = 5 * time.Minute

// MaxProposalRevisions caps a single supersede chain so a refine conversation
// cannot grow unbounded (Phase 34 WS1/WS2).
const MaxProposalRevisions = 8

// ActionProposal is returned on chat `done` and confirm flows (Phase 29 WS3;
// revision lineage + blind-spot facts + impact added Phase 34).
type ActionProposal struct {
	ProposalID string         `json:"proposal_id"`
	Tool       string         `json:"tool"`
	Args       map[string]any `json:"args"`
	Summary    string         `json:"summary"`
	RiskTier   string         `json:"risk_tier"`
	ExpiresAt  time.Time      `json:"expires_at"`
	// Phase 34 — revise/supersede + operator-supplied facts + impact explanation.
	Revision             int            `json:"revision,omitempty"`
	SupersedesProposalID string         `json:"supersedes_proposal_id,omitempty"`
	Status               string         `json:"status,omitempty"`
	OperatorProvided     []OperatorFact `json:"operator_provided,omitempty"`
	ImpactSummary        []string       `json:"impact_summary,omitempty"`
	LLMSourced           bool           `json:"llm_sourced,omitempty"`
}

// OperatorFact is a ground-truth value the operator asserts that Guardian cannot
// sense (Phase 34 WS3). It is stored in proposal meta and labeled in the UI/audit
// as operator-stated, never merged into args as if it were a measurement.
type OperatorFact struct {
	Field   string `json:"field"`
	Value   any    `json:"value"`
	Basis   string `json:"basis"` // always "operator_stated"
	Label   string `json:"label"` // e.g. "RH 60% (operator-stated, not measured)"
	TurnRef string `json:"turn_ref,omitempty"`
}

// proposalMeta is the JSONB shape persisted in guardian_action_proposals.meta.
type proposalMeta struct {
	OperatorProvided []OperatorFact `json:"operator_provided,omitempty"`
	LLMSourced       bool           `json:"llm_sourced,omitempty"`
}

var (
	alertIDPattern = regexp.MustCompile(`(?i)(?:alert\s*#?|#)\s*(\d+)`)
	ackIntent      = regexp.MustCompile(`(?i)\b(acknowledge|ack)\b.*\balert\b|\balert\b.*\b(acknowledge|ack)\b`)
	readIntent     = regexp.MustCompile(`(?i)\b(mark\s+.*read|read)\b.*\balert\b|\balert\b.*\b(mark\s+.*read|read)\b`)
)

// BuildRuleAssistedProposals templates a single reviewed write when the operator
// message matches action intent. When an active pending proposal exists in the
// session and the turn reads as a correction, it revises that draft (Phase 34 WS2)
// instead of starting over; otherwise it builds a fresh proposal (Phase 29/32).
func BuildRuleAssistedProposals(
	ctx context.Context,
	q db.Querier,
	userID uuid.UUID,
	farmID int64,
	sessionID uuid.UUID,
	question string,
	snap Snapshot,
) ([]ActionProposal, error) {
	if q == nil || farmID <= 0 {
		return nil, nil
	}

	// Phase 34 — a correction turn against the live draft revises it in place.
	if sessionID != uuid.Nil {
		if revised, handled, err := tryReviseActiveProposal(ctx, q, userID, sessionID, question, snap); err != nil {
			return nil, err
		} else if handled {
			return revised, nil
		}
	}

	toolID, args, summary, ok := matchFreshProposal(ctx, q, farmID, question, snap)
	if !ok {
		return nil, nil
	}

	row, err := insertProposal(ctx, q, insertProposalInput{
		userID:    userID,
		farmID:    farmID,
		sessionID: sessionID,
		toolID:    toolID,
		args:      args,
		summary:   summary,
		revision:  1,
	})
	if err != nil {
		return nil, err
	}
	return []ActionProposal{ActionProposalFromRow(row)}, nil
}

// matchFreshProposal runs the Phase 29/30/32 intent matchers for a brand-new draft.
func matchFreshProposal(
	ctx context.Context,
	q db.Querier,
	farmID int64,
	question string,
	snap Snapshot,
) (toolID string, args map[string]any, summary string, ok bool) {
	if len(snap.UnreadAlertDetails) > 0 {
		toolID, ok = matchAlertToolIntent(question)
		if ok {
			alert := pickAlertForIntent(question, snap.UnreadAlertDetails)
			if alert.ID == 0 {
				return "", nil, "", false
			}
			args = map[string]any{"alert_id": alert.ID}
			summary = proposalSummary(toolID, alert)
			return toolID, args, summary, true
		}
	}
	if packArgs, packSummary, okPack := matchSetupPackIntent(ctx, q, farmID, question, snap); okPack {
		return "apply_grow_setup_pack", packArgs, packSummary, true
	}
	if toolID, args, summary, okFeed := matchFeedingProgramIntent(ctx, q, farmID, question, snap); okFeed {
		return toolID, args, summary, true
	}
	if toolID, args, summary, okCfg := matchConfigToolIntent(question, snap); okCfg {
		return toolID, args, summary, true
	}
	return matchComfortAutomationIntent(ctx, q, farmID, question, snap)
}

// insertProposalInput carries the fields for one frozen proposal row.
type insertProposalInput struct {
	userID     uuid.UUID
	farmID     int64
	sessionID  uuid.UUID
	toolID     string
	args       map[string]any
	summary    string
	revision   int32
	supersedes uuid.UUID // uuid.Nil for a first draft
	facts      []OperatorFact
	llmSourced bool
}

func insertProposal(ctx context.Context, q db.Querier, in insertProposalInput) (db.Gr33ncoreGuardianActionProposal, error) {
	argsJSON, _ := json.Marshal(in.args)
	metaJSON := marshalProposalMeta(in.facts, in.llmSourced)
	expires := time.Now().UTC().Add(ProposalTTL)
	var sessUUID pgtype.UUID
	if in.sessionID != uuid.Nil {
		sessUUID = pgtype.UUID{Bytes: in.sessionID, Valid: true}
	}
	var supersedes pgtype.UUID
	if in.supersedes != uuid.Nil {
		supersedes = pgtype.UUID{Bytes: in.supersedes, Valid: true}
	}
	rev := in.revision
	if rev < 1 {
		rev = 1
	}
	return q.InsertGuardianProposal(ctx, db.InsertGuardianProposalParams{
		UserID:               in.userID,
		FarmID:               in.farmID,
		SessionID:            sessUUID,
		ToolID:               in.toolID,
		Args:                 argsJSON,
		Summary:              in.summary,
		RiskTier:             tools.RiskTierForTool(in.toolID, in.args),
		ExpiresAt:            expires,
		Meta:                 metaJSON,
		SupersedesProposalID: supersedes,
		Revision:             rev,
	})
}

func marshalProposalMeta(facts []OperatorFact, llmSourced bool) []byte {
	if len(facts) == 0 && !llmSourced {
		return []byte("{}")
	}
	b, err := json.Marshal(proposalMeta{OperatorProvided: facts, LLMSourced: llmSourced})
	if err != nil || len(b) == 0 {
		return []byte("{}")
	}
	return b
}

func matchAlertToolIntent(question string) (string, bool) {
	q := strings.TrimSpace(question)
	if ackIntent.MatchString(q) {
		return "ack_alert", true
	}
	if readIntent.MatchString(q) {
		return "mark_alert_read", true
	}
	// Demo-friendly: "acknowledge the humidity alert" without the word "alert" twice.
	lower := strings.ToLower(q)
	if strings.Contains(lower, "acknowledge") || strings.Contains(lower, " ack ") {
		if strings.Contains(lower, "humidity") || strings.Contains(lower, "alert") || strings.Contains(lower, "ohn") {
			return "ack_alert", true
		}
	}
	if strings.Contains(lower, "mark") && strings.Contains(lower, "read") {
		return "mark_alert_read", true
	}
	return "", false
}

func pickAlertForIntent(question string, details []UnreadAlertDetail) UnreadAlertDetail {
	if m := alertIDPattern.FindStringSubmatch(question); len(m) > 1 {
		var id int64
		for _, c := range m[1] {
			id = id*10 + int64(c-'0')
		}
		if id > 0 {
			for _, a := range details {
				if a.ID == id {
					return a
				}
			}
		}
	}
	lower := strings.ToLower(question)
	if strings.Contains(lower, "restock") || strings.Contains(lower, "refill") ||
		strings.Contains(lower, "reorder") || strings.Contains(lower, "low stock") || strings.Contains(lower, "low-stock") {
		for _, a := range details {
			if a.SourceType == "inventory_low_stock" {
				return a
			}
		}
		for _, a := range details {
			subj := strings.ToLower(a.Subject)
			if strings.Contains(subj, "inventory low") || strings.Contains(subj, "low stock") {
				return a
			}
		}
	}
	keywords := []string{"humidity", "ohn", "inventory", "light", "schedule"}
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			for _, a := range details {
				if strings.Contains(strings.ToLower(a.Subject), kw) {
					return a
				}
			}
		}
	}
	return details[0]
}

func proposalSummary(toolID string, a UnreadAlertDetail) string {
	subj := a.Subject
	if subj == "" {
		subj = "alert #" + strconv.FormatInt(a.ID, 10)
	}
	switch toolID {
	case "ack_alert":
		return "Acknowledge: " + subj
	case "mark_alert_read":
		return "Mark as read: " + subj
	default:
		return subj
	}
}
