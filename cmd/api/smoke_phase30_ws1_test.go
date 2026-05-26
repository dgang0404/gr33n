// Phase 30 WS1 — GET /v1/chat/proposals inbox API smoke.
package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestPhase30WS1_ListProposalsUnauthorized(t *testing.T) {
	resp, err := http.Get(testServer.URL + "/v1/chat/proposals?farm_id=1&status=pending")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusUnauthorized)
}

func TestPhase30WS1_ListProposalsIncludesChatProposal(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	var alertID int64
	err := testPool.QueryRow(ctx, `
SELECT id FROM gr33ncore.alerts_notifications
WHERE farm_id = 1 AND subject_rendered = 'Humidity high — Flower Room'
ORDER BY id DESC LIMIT 1`).Scan(&alertID)
	if err != nil || alertID == 0 {
		t.Skip("seed humidity alert missing — run master_seed.sql")
	}

	tok := smokeJWT(t)
	chatResp := authPost(t, tok, "/v1/chat", map[string]any{
		"message": "acknowledge the humidity alert",
		"farm_id": 1,
		"stream":  false,
	})
	defer chatResp.Body.Close()
	if chatResp.StatusCode == http.StatusServiceUnavailable {
		t.Skip("LLM not configured — set LLM_BASE_URL and LLM_MODEL for full E2E")
	}
	if chatResp.StatusCode != http.StatusOK {
		t.Fatalf("chat status %d: %s", chatResp.StatusCode, readBodyPreview(chatResp))
	}

	var chatBody struct {
		Proposals []struct {
			ProposalID string `json:"proposal_id"`
		} `json:"proposals"`
	}
	decodeJSON(t, chatResp.Body, &chatBody)
	if len(chatBody.Proposals) == 0 {
		t.Fatal("expected proposals from chat turn")
	}
	proposalID := chatBody.Proposals[0].ProposalID

	listResp := authGet(t, tok, "/v1/chat/proposals?farm_id=1&status=pending")
	defer listResp.Body.Close()
	expectStatus(t, listResp, http.StatusOK)

	var listBody struct {
		Proposals []struct {
			ProposalID string `json:"proposal_id"`
			Tool       string `json:"tool"`
			Status     string `json:"status"`
		} `json:"proposals"`
		Total int64 `json:"total"`
	}
	decodeJSON(t, listResp.Body, &listBody)
	if listBody.Total < 1 {
		t.Fatalf("expected total >= 1, got %d", listBody.Total)
	}
	found := false
	for _, p := range listBody.Proposals {
		if p.ProposalID == proposalID {
			found = true
			if p.Tool != "ack_alert" {
				t.Fatalf("tool %q want ack_alert", p.Tool)
			}
			if p.Status != "pending" {
				t.Fatalf("status %q want pending", p.Status)
			}
		}
	}
	if !found {
		t.Fatalf("proposal %s not in list %+v", proposalID, listBody.Proposals)
	}

	// Confirm from inbox path (same POST /v1/chat/confirm as chat card).
	confirmResp := authPost(t, tok, "/v1/chat/confirm", map[string]string{"proposal_id": proposalID})
	defer confirmResp.Body.Close()
	expectStatus(t, confirmResp, http.StatusOK)

	listAfter := authGet(t, tok, "/v1/chat/proposals?farm_id=1&status=pending")
	defer listAfter.Body.Close()
	var afterBody struct {
		Proposals []struct {
			ProposalID string `json:"proposal_id"`
		} `json:"proposals"`
	}
	decodeJSON(t, listAfter.Body, &afterBody)
	for _, p := range afterBody.Proposals {
		if p.ProposalID == proposalID {
			t.Fatalf("confirmed proposal %s should not appear in pending list", proposalID)
		}
	}
}

func decodeJSON(t *testing.T, r io.Reader, dest any) {
	t.Helper()
	if err := json.NewDecoder(r).Decode(dest); err != nil {
		t.Fatalf("decode: %v", err)
	}
}
