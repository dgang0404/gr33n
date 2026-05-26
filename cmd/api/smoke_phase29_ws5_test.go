// Phase 29 WS5 — Guardian confirm RBAC + audit smoke.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func TestPhase29WS5_Confirm_ViewerForbidden(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	viewerID, viewerTok := seedSmokeViewerUser(t, ctx)
	t.Cleanup(func() {
		c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = testPool.Exec(c, `DELETE FROM gr33ncore.guardian_action_proposals WHERE user_id = $1`, viewerID)
		_, _ = testPool.Exec(c, `DELETE FROM gr33ncore.farm_memberships WHERE user_id = $1`, viewerID)
		_, _ = testPool.Exec(c, `DELETE FROM gr33ncore.profiles WHERE user_id = $1`, viewerID)
		_, _ = testPool.Exec(c, `DELETE FROM auth.users WHERE id = $1`, viewerID)
	})

	var alertID int64
	if err := testPool.QueryRow(ctx, `
SELECT id FROM gr33ncore.alerts_notifications
WHERE farm_id = 1 AND is_acknowledged = FALSE
ORDER BY id DESC LIMIT 1`).Scan(&alertID); err != nil || alertID == 0 {
		t.Skip("no unacknowledged alert on farm 1")
	}

	props, err := buildTestProposalForUser(ctx, viewerID, alertID, "ack_alert", "Viewer proposal smoke")
	if err != nil {
		t.Fatalf("build proposal: %v", err)
	}

	body, _ := json.Marshal(map[string]string{"proposal_id": props.ProposalID})
	req, _ := http.NewRequest(http.MethodPost, testServer.URL+"/v1/chat/confirm", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+viewerTok)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("confirm: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("viewer confirm want 403, got %d", resp.StatusCode)
	}
}

func TestPhase29WS5_Confirm_WritesAuditRow(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var alertID int64
	if err := testPool.QueryRow(ctx, `
SELECT id FROM gr33ncore.alerts_notifications
WHERE farm_id = 1 AND subject_rendered = 'OHN batch below minimum — reorder or brew soon'
ORDER BY id DESC LIMIT 1`).Scan(&alertID); err != nil || alertID == 0 {
		t.Skip("seed OHN alert missing")
	}

	tok := smokeJWT(t)
	props, err := buildTestProposal(ctx, alertID, "mark_alert_read", "Mark read: OHN inventory (audit smoke)")
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
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("confirm status %d", resp.StatusCode)
	}

	var actionType, kind, toolID string
	err = testPool.QueryRow(ctx, `
SELECT action_type::text, details->>'kind', details->>'tool_id'
FROM gr33ncore.user_activity_log
WHERE farm_id = 1
  AND action_type = 'guardian_tool_executed'
  AND details->>'proposal_id' = $1
ORDER BY activity_time DESC
LIMIT 1`, props.ProposalID).Scan(&actionType, &kind, &toolID)
	if err != nil {
		t.Fatalf("audit row: %v", err)
	}
	if kind != "guardian_tool_executed" || toolID != "mark_alert_read" {
		t.Fatalf("unexpected audit details kind=%q tool_id=%q", kind, toolID)
	}
}

func seedSmokeViewerUser(t *testing.T, ctx context.Context) (uuid.UUID, string) {
	t.Helper()
	viewerID := uuid.New()
	email := "viewer_" + viewerID.String()[:8] + "@test.local"
	hash, err := bcrypt.GenerateFromPassword([]byte(smokeDevPass), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}
	if _, err := testPool.Exec(ctx, `
INSERT INTO auth.users (id, email, password_hash, created_at)
VALUES ($1, $2, $3, NOW())`, viewerID, email, hash); err != nil {
		t.Fatalf("insert viewer auth: %v", err)
	}
	if _, err := testPool.Exec(ctx, `
INSERT INTO gr33ncore.profiles (user_id, full_name, email, created_at, updated_at)
VALUES ($1, 'Viewer Smoke', $2, NOW(), NOW())`, viewerID, email); err != nil {
		t.Fatalf("insert viewer profile: %v", err)
	}
	if _, err := testPool.Exec(ctx, `
INSERT INTO gr33ncore.farm_memberships (farm_id, user_id, role_in_farm, permissions, joined_at)
VALUES (1, $1, 'viewer', '{}'::jsonb, NOW())`, viewerID); err != nil {
		t.Fatalf("insert viewer membership: %v", err)
	}

	resp := postNoAuth("/auth/login", map[string]any{
		"username": email,
		"password": smokeDevPass,
	})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("viewer login status %d", resp.StatusCode)
	}
	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode login: %v", err)
	}
	tok, _ := body["token"].(string)
	if tok == "" {
		t.Fatal("viewer login missing token")
	}
	return viewerID, tok
}

func buildTestProposalForUser(ctx context.Context, userID uuid.UUID, alertID int64, toolID, summary string) (struct {
	ProposalID string
}, error) {
	var out struct {
		ProposalID string
	}
	args, _ := json.Marshal(map[string]any{"alert_id": alertID})
	var pid string
	err := testPool.QueryRow(ctx, `
INSERT INTO gr33ncore.guardian_action_proposals
    (user_id, farm_id, tool_id, args, summary, expires_at)
VALUES ($1, 1, $2, $3::jsonb, $4, NOW() + INTERVAL '5 minutes')
RETURNING proposal_id::text`,
		userID, toolID, args, summary,
	).Scan(&pid)
	out.ProposalID = pid
	return out, err
}
