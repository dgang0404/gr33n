package eval

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// PendingProposal mirrors the fields of chat.proposalListItem this client
// actually needs — kept local to eval instead of importing internal/handler/chat
// (which already imports farmguardian and would create an import cycle).
type PendingProposal struct {
	ProposalID string         `json:"proposal_id"`
	Tool       string         `json:"tool"`
	Summary    string         `json:"summary"`
	RiskTier   string         `json:"risk_tier"`
	Status     string         `json:"status"`
	SessionID  string         `json:"session_id,omitempty"`
	Revision   int            `json:"revision,omitempty"`
	Args       map[string]any `json:"args"`
}

// FetchPendingProposals calls GET /v1/chat/proposals?farm_id=...&status=pending
// — the same endpoint the UI's Guardian change-request ("PR") queue reads
// from — so a smoke script can confirm a write-intent prompt actually landed
// a row there, not just that the chat response echoed a proposal inline.
func (c *APIClient) FetchPendingProposals(ctx context.Context) ([]PendingProposal, error) {
	if c == nil || c.HTTP == nil {
		return nil, fmt.Errorf("eval API client not configured")
	}
	url := strings.TrimRight(c.BaseURL, "/") + "/v1/chat/proposals?status=pending&limit=100"
	if c.FarmID > 0 {
		url += "&farm_id=" + strconv.FormatInt(c.FarmID, 10)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("GET /v1/chat/proposals HTTP %d: %s", resp.StatusCode, truncate(string(body), 200))
	}
	var parsed struct {
		Proposals []PendingProposal `json:"proposals"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	return parsed.Proposals, nil
}

// VerifyPendingProposalIDs confirms each proposal_id from this prompt is still
// in the pending queue. Call immediately after the chat turn — proposals expire
// after ProposalTTL (5m) while eval prompts can take 20+ minutes each.
func VerifyPendingProposalIDs(ctx context.Context, client *APIClient, proposalIDs []string) error {
	var ids []string
	for _, id := range proposalIDs {
		if id = strings.TrimSpace(id); id != "" {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return fmt.Errorf("no proposal_id(s) to verify")
	}
	pending, err := client.FetchPendingProposals(ctx)
	if err != nil {
		return err
	}
	pendingSet := make(map[string]PendingProposal, len(pending))
	for _, p := range pending {
		if id := strings.TrimSpace(p.ProposalID); id != "" {
			pendingSet[id] = p
		}
	}
	var missing []string
	for _, id := range ids {
		if _, ok := pendingSet[id]; !ok {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("proposal_id(s) not in pending queue: %s", strings.Join(missing, ", "))
	}
	return nil
}

// BumpProposalExpiry extends expires_at on pending proposals so they remain visible
// in the UI Pending tab after long eval runs (default ProposalTTL is only 5m).
func BumpProposalExpiry(ctx context.Context, proposalIDs []string, ttl time.Duration) (int, error) {
	var ids []uuid.UUID
	for _, raw := range proposalIDs {
		id, err := uuid.Parse(strings.TrimSpace(raw))
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return 0, fmt.Errorf("no valid proposal_id(s) to bump")
	}
	dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dsn == "" {
		return 0, fmt.Errorf("DATABASE_URL required for -leave-pending")
	}
	if ttl <= 0 {
		ttl = LeavePendingTTLFromEnv()
	}
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return 0, err
	}
	defer conn.Close(ctx)
	expires := time.Now().UTC().Add(ttl)
	tag, err := conn.Exec(ctx, `
		UPDATE gr33ncore.guardian_action_proposals
		SET expires_at = $1
		WHERE status = 'pending' AND proposal_id = ANY($2::uuid[])`,
		expires, ids)
	if err != nil {
		return 0, err
	}
	return int(tag.RowsAffected()), nil
}
