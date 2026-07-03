// Phase 117 — handler gap smokes (auth modes, RBAC deny paths, fileattach limits).
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func TestPhase117_AuthRegisterLoginRoundtrip(t *testing.T) {
	email := fmt.Sprintf("phase117_%d@test.local", rand.Int())
	pass := "longpassword117"
	resp := postNoAuth("/auth/register", map[string]any{
		"email":     email,
		"password":  pass,
		"full_name": "Phase 117 Register",
	})
	expectStatus(t, resp, http.StatusCreated)
	reg := decodeMap(t, resp)
	tok, _ := reg["token"].(string)
	if tok == "" {
		t.Fatal("register should return token")
	}

	resp = postNoAuth("/auth/login", map[string]any{
		"username": email,
		"password": pass,
	})
	expectStatus(t, resp, http.StatusOK)
	login := decodeMap(t, resp)
	if login["token"] == nil {
		t.Fatal("login should return token")
	}
}

func TestPhase117_ViewerCostAndReceiptRBACDeny(t *testing.T) {
	ctx := context.Background()
	_, viewerTok := seedSmokeViewerUser(t, ctx)

	resp := authGet(t, viewerTok, "/farms/1/costs?limit=5")
	expectStatus(t, resp, http.StatusForbidden)

	resp = authMultipartPost(t, viewerTok, "/farms/1/cost-receipts", "file", "deny.pdf", "application/pdf",
		[]byte("%PDF-1.4 minimal"), nil)
	expectStatus(t, resp, http.StatusForbidden)
}

func TestPhase117_FinanceUserCanReadCosts(t *testing.T) {
	ctx := context.Background()
	_, financeTok := seedSmokeFinanceUser(t, ctx)
	resp := authGet(t, financeTok, "/farms/1/costs/summary")
	expectStatus(t, resp, http.StatusOK)
}

func TestPhase117_OrganizationNonMemberDenied(t *testing.T) {
	tok := smokeJWT(t)
	resp := authPost(t, tok, "/organizations", map[string]any{"name": uniqueName("phase117_org")})
	expectStatus(t, resp, http.StatusCreated)
	org := decodeMap(t, resp)
	orgID := int64(org["id"].(float64))

	email := fmt.Sprintf("org_outsider_%d@test.local", rand.Int())
	resp = postNoAuth("/auth/register", map[string]any{
		"email":     email,
		"password":  "longpassword117",
		"full_name": "Org Outsider",
	})
	expectStatus(t, resp, http.StatusCreated)
	outsider := decodeMap(t, resp)
	outsiderTok, _ := outsider["token"].(string)

	resp = authGet(t, outsiderTok, fmt.Sprintf("/organizations/%d", orgID))
	expectStatus(t, resp, http.StatusForbidden)
}

func TestPhase117_GuardianProposalsListAndDismiss(t *testing.T) {
	tok := smokeJWT(t)
	ctx := context.Background()

	listResp := authGet(t, tok, "/v1/chat/proposals?farm_id=1&status=pending&limit=20")
	expectStatus(t, listResp, http.StatusOK)
	listBody := decodeMap(t, listResp)
	pending, _ := listBody["proposals"].([]any)

	if len(pending) == 0 {
		alertID := seedSmokeAlertForProposal(t, ctx)
		proposalID := insertSmokeGuardianProposal(t, ctx, alertID, "ack_alert", "Phase 117 dismiss smoke")
		defer func() {
			_, _ = testPool.Exec(ctx, `DELETE FROM gr33ncore.guardian_action_proposals WHERE proposal_id = $1`, proposalID)
		}()
	} else {
		first := pending[0].(map[string]any)
		proposalID, _ := first["proposal_id"].(string)
		if proposalID == "" {
			t.Fatalf("pending proposal missing proposal_id: %#v", first)
		}
		dismissResp := authPost(t, tok, fmt.Sprintf("/v1/chat/proposals/%s/dismiss", proposalID), nil)
		expectStatus(t, dismissResp, http.StatusOK)
		return
	}

	listAgain := authGet(t, tok, "/v1/chat/proposals?farm_id=1&status=pending&limit=20")
	expectStatus(t, listAgain, http.StatusOK)
	againBody := decodeMap(t, listAgain)
	rows, _ := againBody["proposals"].([]any)
	if len(rows) == 0 {
		t.Fatal("expected at least one pending proposal after seed")
	}
	prop := rows[0].(map[string]any)
	proposalID, _ := prop["proposal_id"].(string)
	dismissResp := authPost(t, tok, fmt.Sprintf("/v1/chat/proposals/%s/dismiss", proposalID), nil)
	expectStatus(t, dismissResp, http.StatusOK)
}

func TestPhase117_AlertAcknowledgeReturnsRow(t *testing.T) {
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/farms/1/alerts?limit=5")
	expectStatus(t, resp, http.StatusOK)
	alerts := decodeSlice(t, resp)
	if len(alerts) == 0 {
		t.Skip("no alerts in seed data")
	}
	alertID := int64(alerts[0].(map[string]any)["id"].(float64))
	resp = authPatch(t, tok, fmt.Sprintf("/alerts/%d/acknowledge", alertID), map[string]any{})
	expectStatus(t, resp, http.StatusOK)
	body := decodeMap(t, resp)
	if body["id"] == nil {
		t.Fatalf("expected alert row in ack response: %#v", body)
	}
}

func TestPhase117_FileAttachInvalidMimeRejected(t *testing.T) {
	tok := smokeJWT(t)
	resp := authMultipartPost(t, tok, "/farms/1/cost-receipts", "file", "bad.txt", "text/plain",
		[]byte("not a receipt"), nil)
	expectStatus(t, resp, http.StatusBadRequest)
}

func seedSmokeFinanceUser(t *testing.T, ctx context.Context) (uuid.UUID, string) {
	t.Helper()
	financeID := uuid.New()
	email := "finance_" + financeID.String()[:8] + "@test.local"
	hash, err := bcrypt.GenerateFromPassword([]byte(smokeDevPass), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}
	if _, err := testPool.Exec(ctx, `
INSERT INTO auth.users (id, email, password_hash, created_at)
VALUES ($1, $2, $3, NOW())`, financeID, email, hash); err != nil {
		t.Fatalf("insert finance auth: %v", err)
	}
	if _, err := testPool.Exec(ctx, `
INSERT INTO gr33ncore.profiles (user_id, full_name, email, created_at, updated_at)
VALUES ($1, 'Finance Smoke', $2, NOW(), NOW())`, financeID, email); err != nil {
		t.Fatalf("insert finance profile: %v", err)
	}
	if _, err := testPool.Exec(ctx, `
INSERT INTO gr33ncore.farm_memberships (farm_id, user_id, role_in_farm, permissions, joined_at)
VALUES (1, $1, 'finance', '{}'::jsonb, NOW())`, financeID); err != nil {
		t.Fatalf("insert finance membership: %v", err)
	}
	resp := postNoAuth("/auth/login", map[string]any{
		"username": email,
		"password": smokeDevPass,
	})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("finance login status %d", resp.StatusCode)
	}
	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode login: %v", err)
	}
	tok, _ := body["token"].(string)
	if tok == "" {
		t.Fatal("finance login missing token")
	}
	return financeID, tok
}

func seedSmokeAlertForProposal(t *testing.T, ctx context.Context) int64 {
	t.Helper()
	var alertID int64
	err := testPool.QueryRow(ctx, `
INSERT INTO gr33ncore.alerts_notifications
  (farm_id, triggering_event_source_type, triggering_event_source_id,
   severity, subject_rendered, message_text_rendered, status, is_read, created_at)
VALUES (1, 'smoke_test', 0, 'medium'::gr33ncore.notification_priority_enum,
        'Phase 117', 'Smoke alert', 'pending', FALSE, NOW())
RETURNING id`).Scan(&alertID)
	if err != nil {
		t.Fatalf("insert alert: %v", err)
	}
	return alertID
}

func insertSmokeGuardianProposal(t *testing.T, ctx context.Context, alertID int64, toolID, summary string) string {
	t.Helper()
	uid := uuid.MustParse(smokeDevUserUUID)
	args, _ := json.Marshal(map[string]any{"alert_id": alertID})
	var proposalID string
	err := testPool.QueryRow(ctx, `
INSERT INTO gr33ncore.guardian_action_proposals
    (user_id, farm_id, tool_id, args, summary, risk_tier, expires_at)
VALUES ($1, 1, $2, $3::jsonb, $4, 'low', NOW() + INTERVAL '10 minutes')
RETURNING proposal_id::text`, uid, toolID, args, summary).Scan(&proposalID)
	if err != nil {
		t.Fatalf("insert proposal: %v", err)
	}
	return proposalID
}
