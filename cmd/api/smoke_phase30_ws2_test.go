// Phase 30 WS2 — proposal risk_tier on list + chat proposals.
package main

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPhase30WS2_ProposalRiskTierOnList(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	uid := uuid.MustParse(smokeDevUserUUID)
	args, _ := json.Marshal(map[string]any{"template": "jadam_indoor_photoperiod_v1"})
	var pid string
	err := testPool.QueryRow(ctx, `
INSERT INTO gr33ncore.guardian_action_proposals
    (user_id, farm_id, tool_id, args, summary, risk_tier, expires_at)
VALUES ($1, 1, 'apply_bootstrap_template', $2::jsonb, 'Apply bootstrap template', 'high', NOW() + INTERVAL '5 minutes')
RETURNING proposal_id::text`, uid, args).Scan(&pid)
	if err != nil {
		t.Fatalf("insert proposal: %v", err)
	}

	tok := smokeJWT(t)
	resp := authGet(t, tok, "/v1/chat/proposals?farm_id=1&status=pending")
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusOK)

	var body struct {
		Proposals []struct {
			ProposalID string `json:"proposal_id"`
			RiskTier   string `json:"risk_tier"`
			Tool       string `json:"tool"`
		} `json:"proposals"`
	}
	decodeJSON(t, resp.Body, &body)
	for _, p := range body.Proposals {
		if p.ProposalID == pid {
			if p.RiskTier != "high" {
				t.Fatalf("risk_tier %q want high", p.RiskTier)
			}
			return
		}
	}
	t.Fatalf("proposal %s not in list", pid)
}

func TestPhase30WS2_CreateTaskProposalIsMedium(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	tok := smokeJWT(t)
	chatResp := authPost(t, tok, "/v1/chat", map[string]any{
		"message": "Create a task to check Flower Room humidity",
		"farm_id": 1,
		"stream":  false,
	})
	defer chatResp.Body.Close()
	if chatResp.StatusCode == http.StatusServiceUnavailable {
		t.Skip("LLM not configured")
	}
	if chatResp.StatusCode != http.StatusOK {
		t.Fatalf("chat %d: %s", chatResp.StatusCode, readBodyPreview(chatResp))
	}
	var chatBody struct {
		Proposals []struct {
			Tool     string `json:"tool"`
			RiskTier string `json:"risk_tier"`
		} `json:"proposals"`
	}
	decodeJSON(t, chatResp.Body, &chatBody)
	if len(chatBody.Proposals) == 0 {
		t.Fatal("expected proposal")
	}
	if chatBody.Proposals[0].Tool != "create_task" {
		t.Fatalf("tool %q", chatBody.Proposals[0].Tool)
	}
	if chatBody.Proposals[0].RiskTier != "medium" {
		t.Fatalf("risk_tier %q want medium", chatBody.Proposals[0].RiskTier)
	}
}
