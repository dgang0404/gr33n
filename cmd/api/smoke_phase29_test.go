// Phase 29 WS8 — OpenAPI + end-to-end propose→confirm smoke.
//
// Exercises the full chat → proposals[] → POST /v1/chat/confirm path when
// LLM_BASE_URL/LLM_MODEL are configured in the smoke process. Falls back to
// skipping the chat leg with t.Skip when the handler returns 503 (typical in
// CI without Ollama). Confirm-path coverage complements smoke_phase29_ws3/ws5.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestPhase29WS8_ConfirmUnauthorized(t *testing.T) {
	resp := postNoAuth("/v1/chat/confirm", map[string]string{
		"proposal_id": "550e8400-e29b-41d4-a716-446655440000",
	})
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusUnauthorized)
}

func TestPhase29WS8_ConfirmInvalidProposalID(t *testing.T) {
	tok := smokeJWT(t)
	resp := authPost(t, tok, "/v1/chat/confirm", map[string]string{"proposal_id": "not-a-uuid"})
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusBadRequest)
}

func TestPhase29WS8_ChatProposeConfirmAck_EndToEnd(t *testing.T) {
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

	// Ensure the alert is unacknowledged so ack_alert can run.
	_, _ = testPool.Exec(ctx, `
UPDATE gr33ncore.alerts_notifications
SET is_acknowledged = FALSE, is_read = FALSE
WHERE id = $1`, alertID)

	tok := smokeJWT(t)
	chatResp := authPost(t, tok, "/v1/chat", map[string]any{
		"message":  "acknowledge the humidity alert",
		"farm_id":  1,
		"stream":   false,
		"context_ref": map[string]any{
			"type": "alert",
			"id":   alertID,
		},
	})
	defer chatResp.Body.Close()

	if chatResp.StatusCode == http.StatusServiceUnavailable {
		t.Skip("LLM not configured in smoke process — set LLM_BASE_URL and LLM_MODEL for full E2E")
	}
	if chatResp.StatusCode != http.StatusOK {
		t.Fatalf("chat status %d: %s", chatResp.StatusCode, readBodyPreview(chatResp))
	}

	var chatBody struct {
		Proposals []struct {
			ProposalID string         `json:"proposal_id"`
			Tool       string         `json:"tool"`
			Args       map[string]any `json:"args"`
			Summary    string         `json:"summary"`
		} `json:"proposals"`
	}
	if err := json.NewDecoder(chatResp.Body).Decode(&chatBody); err != nil {
		t.Fatalf("decode chat: %v", err)
	}
	if len(chatBody.Proposals) == 0 {
		t.Fatal("expected proposals[] on grounded ack intent chat turn")
	}
	prop := chatBody.Proposals[0]
	if prop.Tool != "ack_alert" {
		t.Fatalf("expected ack_alert proposal, got %q", prop.Tool)
	}
	if int64(prop.Args["alert_id"].(float64)) != alertID {
		t.Fatalf("proposal alert_id=%v want %d", prop.Args["alert_id"], alertID)
	}

	confirmBody, _ := json.Marshal(map[string]string{"proposal_id": prop.ProposalID})
	req, _ := http.NewRequest(http.MethodPost, testServer.URL+"/v1/chat/confirm", bytes.NewReader(confirmBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("confirm: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("confirm status %d: %s", resp.StatusCode, readBodyPreview(resp))
	}

	var ack bool
	if err := testPool.QueryRow(ctx, `SELECT is_acknowledged FROM gr33ncore.alerts_notifications WHERE id = $1`, alertID).Scan(&ack); err != nil || !ack {
		t.Fatalf("alert %d should be acknowledged, ack=%v err=%v", alertID, ack, err)
	}

	var actionType string
	err = testPool.QueryRow(ctx, `
SELECT action_type::text
FROM gr33ncore.user_activity_log
WHERE farm_id = 1
  AND action_type = 'guardian_tool_executed'
  AND details->>'proposal_id' = $1
ORDER BY activity_time DESC
LIMIT 1`, prop.ProposalID).Scan(&actionType)
	if err != nil {
		t.Fatalf("audit row: %v", err)
	}

	// Idempotent second confirm
	req2, _ := http.NewRequest(http.MethodPost, testServer.URL+"/v1/chat/confirm", bytes.NewReader(confirmBody))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+tok)
	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		t.Fatalf("confirm retry: %v", err)
	}
	resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("idempotent confirm status %d", resp2.StatusCode)
	}
}

func readBodyPreview(resp *http.Response) string {
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
	return string(b)
}
