package eval

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"gr33n-api/internal/farmguardian"
)

// ConfirmResult is the response from POST /v1/chat/confirm (Phase 162).
type ConfirmResult struct {
	ProposalID string
	Summary    string
	Result     map[string]any
}

// ConfirmProposal executes a frozen change request (same path as UI Confirm).
func (c *APIClient) ConfirmProposal(ctx context.Context, proposalID string) (ConfirmResult, error) {
	if c == nil || c.HTTP == nil {
		return ConfirmResult{}, fmt.Errorf("eval API client not configured")
	}
	proposalID = strings.TrimSpace(proposalID)
	if proposalID == "" {
		return ConfirmResult{}, fmt.Errorf("empty proposal_id")
	}
	body, _ := json.Marshal(map[string]string{"proposal_id": proposalID})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.BaseURL, "/")+"/v1/chat/confirm", bytes.NewReader(body))
	if err != nil {
		return ConfirmResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return ConfirmResult{}, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ConfirmResult{}, fmt.Errorf("POST /v1/chat/confirm HTTP %d: %s", resp.StatusCode, truncate(string(respBody), 300))
	}
	var parsed struct {
		Summary string         `json:"summary"`
		Result  map[string]any `json:"result"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return ConfirmResult{}, err
	}
	return ConfirmResult{
		ProposalID: proposalID,
		Summary:    parsed.Summary,
		Result:     parsed.Result,
	}, nil
}

// ProposalConfirmTarget links a smoke fixture to a proposal_id from this run.
type ProposalConfirmTarget struct {
	FixtureID  string
	ProposalID string
}

// ProposalConfirmTargets returns proposal_ids from passed write-intent fixtures.
func ProposalConfirmTargets(fixtures []Question, scores []farmguardian.EvalQuestionScore) []ProposalConfirmTarget {
	expectByID := make(map[string]bool, len(fixtures))
	for _, q := range fixtures {
		if q.ExpectProposal {
			expectByID[q.ID] = true
		}
	}
	var out []ProposalConfirmTarget
	for _, s := range scores {
		if !expectByID[s.ID] || !s.Passed {
			continue
		}
		for _, pid := range s.ProposalIDs {
			if pid = strings.TrimSpace(pid); pid != "" {
				out = append(out, ProposalConfirmTarget{FixtureID: s.ID, ProposalID: pid})
			}
		}
	}
	return out
}
