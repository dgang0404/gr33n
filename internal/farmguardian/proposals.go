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

// ActionProposal is returned on chat `done` and confirm flows (Phase 29 WS3).
type ActionProposal struct {
	ProposalID string         `json:"proposal_id"`
	Tool       string         `json:"tool"`
	Args       map[string]any `json:"args"`
	Summary    string         `json:"summary"`
	RiskTier   string         `json:"risk_tier"`
	ExpiresAt  time.Time      `json:"expires_at"`
}

var (
	alertIDPattern = regexp.MustCompile(`(?i)(?:alert\s*#?|#)\s*(\d+)`)
	ackIntent      = regexp.MustCompile(`(?i)\b(acknowledge|ack)\b.*\balert\b|\balert\b.*\b(acknowledge|ack)\b`)
	readIntent     = regexp.MustCompile(`(?i)\b(mark\s+.*read|read)\b.*\balert\b|\balert\b.*\b(mark\s+.*read|read)\b`)
)

// BuildRuleAssistedProposals templates a single low-risk write when the operator
// message matches action intent and the snapshot lists unread alerts (v1).
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

	var toolID string
	var args map[string]any
	var summary string
	var ok bool

	if len(snap.UnreadAlertDetails) > 0 {
		toolID, ok = matchAlertToolIntent(question)
		if ok {
			alert := pickAlertForIntent(question, snap.UnreadAlertDetails)
			if alert.ID == 0 {
				return nil, nil
			}
			args = map[string]any{"alert_id": alert.ID}
			summary = proposalSummary(toolID, alert)
		}
	}
	if !ok {
		if packArgs, packSummary, okPack := matchSetupPackIntent(ctx, q, farmID, question, snap); okPack {
			toolID = "apply_grow_setup_pack"
			args = packArgs
			summary = packSummary
			ok = true
		}
	}
	if !ok {
		toolID, args, summary, ok = matchConfigToolIntent(question, snap)
		if !ok {
			return nil, nil
		}
	}

	argsJSON, _ := json.Marshal(args)
	expires := time.Now().UTC().Add(ProposalTTL)
	var sessUUID pgtype.UUID
	if sessionID != uuid.Nil {
		sessUUID = pgtype.UUID{Bytes: sessionID, Valid: true}
	}
	row, err := q.InsertGuardianProposal(ctx, db.InsertGuardianProposalParams{
		UserID:    userID,
		FarmID:    farmID,
		SessionID: sessUUID,
		ToolID:    toolID,
		Column5:   argsJSON,
		Summary:   summary,
		RiskTier:  tools.RiskTierForTool(toolID, args),
		ExpiresAt: expires,
	})
	if err != nil {
		return nil, err
	}
	return []ActionProposal{ActionProposalFromRow(row)}, nil
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
