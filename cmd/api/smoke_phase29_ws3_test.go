// Phase 29 WS3 — propose → confirm smoke (ack_alert on seeded humidity alert).
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPhase29WS3_ConfirmAckHumidityAlert(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var alertID int64
	err := testPool.QueryRow(ctx, `
SELECT id FROM gr33ncore.alerts_notifications
WHERE farm_id = 1 AND is_acknowledged = FALSE AND subject_rendered = 'Humidity high — Flower Room'
ORDER BY id DESC LIMIT 1`).Scan(&alertID)
	if err != nil || alertID == 0 {
		t.Skip("seed humidity alert missing — run master_seed.sql")
	}

	tok := smokeJWT(t)
	props, err := buildTestProposal(ctx, alertID, "ack_alert", "Acknowledge humidity alert (smoke)")
	if err != nil {
		t.Fatalf("build proposal: %v", err)
	}

	body, _ := json.Marshal(map[string]string{"proposal_id": props.ProposalID})
	req, _ := http.NewRequest(http.MethodPost, testServer.URL+"/v1/chat/confirm", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("confirm: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("confirm status %d", resp.StatusCode)
	}
	var cr struct {
		Summary string         `json:"summary"`
		Result  map[string]any `json:"result"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&cr)
	if cr.Summary == "" {
		t.Fatal("expected summary in confirm response")
	}

	var ack bool
	if err := testPool.QueryRow(ctx, `SELECT is_acknowledged FROM gr33ncore.alerts_notifications WHERE id = $1`, alertID).Scan(&ack); err != nil || !ack {
		t.Fatalf("alert %d should be acknowledged, ack=%v err=%v", alertID, ack, err)
	}

	// Idempotent second confirm
	req2, _ := http.NewRequest(http.MethodPost, testServer.URL+"/v1/chat/confirm", bytes.NewReader(body))
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

func buildTestProposal(ctx context.Context, alertID int64, toolID, summary string) (struct {
	ProposalID string
}, error) {
	var out struct {
		ProposalID string
	}
	if testPool == nil {
		return out, context.Canceled
	}
	uid := uuid.MustParse(smokeDevUserUUID)
	args, _ := json.Marshal(map[string]any{"alert_id": alertID})
	var pid string
	err := testPool.QueryRow(ctx, `
INSERT INTO gr33ncore.guardian_action_proposals
    (user_id, farm_id, tool_id, args, summary, risk_tier, expires_at)
VALUES ($1, 1, $2, $3::jsonb, $4, 'low', NOW() + INTERVAL '5 minutes')
RETURNING proposal_id::text`,
		uid, toolID, args, summary,
	).Scan(&pid)
	out.ProposalID = pid
	return out, err
}
