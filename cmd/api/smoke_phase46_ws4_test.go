// Phase 46 WS4 — LLM proposal safety smokes (matcher-first, validation rejects, confirm gates).
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmguardian"
)

func TestPhase46WS4_MatcherFirstIgnoresLLMJSON(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	q := db.New(testPool)
	snap, err := farmguardian.BuildSnapshot(ctx, q, 1)
	if err != nil {
		t.Fatalf("BuildSnapshot: %v", err)
	}

	var alertID int64
	if err := testPool.QueryRow(ctx, `
SELECT id FROM gr33ncore.alerts_notifications
WHERE farm_id = 1 AND subject_rendered = 'Humidity high — Flower Room'
ORDER BY id DESC LIMIT 1`).Scan(&alertID); err != nil || alertID == 0 {
		t.Skip("seed humidity alert missing")
	}
	_, _ = testPool.Exec(ctx, `
UPDATE gr33ncore.alerts_notifications
SET is_acknowledged = FALSE, is_read = FALSE WHERE id = $1`, alertID)

	uid := uuid.MustParse(smokeDevUserUUID)
	sessionID := uuid.New()
	question := "acknowledge the humidity alert"
	assistant := llmSmokeProposalJSON("patch_fertigation_program", map[string]any{
		"program_id":          1,
		"total_volume_liters": 0.3,
	}, "LLM should not win")

	props, err := phase46HybridAttach(ctx, q, uid, 1, sessionID, question, assistant, snap,
		farmguardian.LLMProposalPolicy{Enabled: true}, true, true)
	if err != nil {
		t.Fatalf("hybrid attach: %v", err)
	}
	if len(props) != 1 {
		t.Fatalf("expected single matcher proposal, got %+v", props)
	}
	if props[0].Tool != "ack_alert" {
		t.Fatalf("tool %q want ack_alert", props[0].Tool)
	}
	if props[0].LLMSourced {
		t.Fatal("matcher proposal must not be llm_sourced")
	}

	t.Cleanup(func() {
		c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = testPool.Exec(c, `DELETE FROM gr33ncore.guardian_action_proposals WHERE proposal_id = $1`, props[0].ProposalID)
	})
}

func TestPhase46WS4_LLMHappyPathPatchFertigationProgram(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	q := db.New(testPool)
	snap, err := farmguardian.BuildSnapshot(ctx, q, 1)
	if err != nil {
		t.Fatalf("BuildSnapshot: %v", err)
	}

	var programID int64
	if err := testPool.QueryRow(ctx, `
SELECT id FROM gr33nfertigation.programs
WHERE farm_id = 1 AND name = 'Flower Daily FFJ+WCA Program' AND deleted_at IS NULL
LIMIT 1`).Scan(&programID); err != nil || programID == 0 {
		t.Skip("seed flower fertigation program missing")
	}

	uid := uuid.MustParse(smokeDevUserUUID)
	sessionID := uuid.New()
	question := "Patch fertigation dilution per Guardian structured suggestion"
	if farmguardian.FreshMatcherMatches(ctx, q, 1, question, snap) {
		t.Fatal("question should miss matchers for LLM-only path")
	}

	assistant := llmSmokeProposalJSON("patch_fertigation_program", map[string]any{
		"program_id":          programID,
		"total_volume_liters": 0.25,
	}, "Set Flower Daily FFJ+WCA Program volume to 0.25 L")

	props, err := phase46HybridAttach(ctx, q, uid, 1, sessionID, question, assistant, snap,
		farmguardian.LLMProposalPolicy{Enabled: true}, true, true)
	if err != nil {
		t.Fatalf("hybrid attach: %v", err)
	}
	if len(props) != 1 {
		t.Fatalf("expected 1 LLM proposal, got %+v", props)
	}
	prop := props[0]
	if prop.Tool != "patch_fertigation_program" {
		t.Fatalf("tool %q", prop.Tool)
	}
	if !prop.LLMSourced {
		t.Fatal("expected llm_sourced proposal")
	}
	if int64(prop.Args["program_id"].(float64)) != programID {
		t.Fatalf("program_id %v want %d", prop.Args["program_id"], programID)
	}

	t.Cleanup(func() {
		c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = testPool.Exec(c, `DELETE FROM gr33ncore.guardian_action_proposals WHERE proposal_id = $1`, prop.ProposalID)
	})
}

func TestPhase46WS4_WrongProgramIDRejected(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	q := db.New(testPool)
	uid := uuid.MustParse(smokeDevUserUUID)
	before := countPendingProposals(ctx, uid)
	assistant := llmSmokeProposalJSON("patch_fertigation_program", map[string]any{
		"program_id":          999999999,
		"total_volume_liters": 0.25,
	}, "Bad program")

	props, err := farmguardian.TryBuildLLMProposalsFromAssistant(
		ctx, q, uid, 1, uuid.New(),
		"Patch fertigation dilution per Guardian structured suggestion",
		assistant,
		farmguardian.LLMProposalPolicy{Enabled: true},
		true, true, false, false,
	)
	if err != nil {
		t.Fatalf("TryBuildLLM: %v", err)
	}
	if len(props) != 0 {
		t.Fatalf("wrong program_id should not insert, got %+v", props)
	}
	after := countPendingProposals(ctx, uid)
	if after != before {
		t.Fatalf("pending proposals before=%d after=%d", before, after)
	}
}

func TestPhase46WS4_ViewerNoLLMInsert(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	q := db.New(testPool)
	snap, _ := farmguardian.BuildSnapshot(ctx, q, 1)
	viewerID, _ := seedSmokeViewerUser(t, ctx)
	t.Cleanup(func() {
		c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = testPool.Exec(c, `DELETE FROM gr33ncore.guardian_action_proposals WHERE user_id = $1`, viewerID)
	})

	before := countPendingProposals(ctx, viewerID)
	assistant := llmSmokeProposalJSON("create_task", map[string]any{"title": "Viewer task"}, "Task")
	// Deliberately avoids the rule-based createTaskIntent matcher (add|create|make + task)
	// so this exercises the LLM-insertion path under test, not matcher-first rule handling
	// (which is intentionally role-agnostic — confirm-time RequireFarmOperate is the gate).
	props, err := phase46HybridAttach(ctx, q, viewerID, 1, uuid.New(),
		"Flag the reservoir for someone to check on later",
		assistant, snap,
		farmguardian.LLMProposalPolicy{Enabled: true}, false, false)
	if err != nil {
		t.Fatalf("hybrid attach: %v", err)
	}
	if len(props) != 0 {
		t.Fatalf("viewer should get no LLM proposal, got %+v", props)
	}
	if countPendingProposals(ctx, viewerID) != before {
		t.Fatal("viewer pending proposal count changed")
	}
}

func TestPhase46WS4_ConfirmExpiredProposalGone(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var alertID int64
	if err := testPool.QueryRow(ctx, `
SELECT id FROM gr33ncore.alerts_notifications
WHERE farm_id = 1 ORDER BY id DESC LIMIT 1`).Scan(&alertID); err != nil {
		t.Skip("no alerts on farm 1")
	}

	uid := uuid.MustParse(smokeDevUserUUID)
	args, _ := json.Marshal(map[string]any{"alert_id": alertID})
	var proposalID string
	err := testPool.QueryRow(ctx, `
INSERT INTO gr33ncore.guardian_action_proposals
    (user_id, farm_id, tool_id, args, summary, risk_tier, expires_at)
VALUES ($1, 1, 'ack_alert', $2::jsonb, 'Expired smoke', 'low', NOW() - INTERVAL '1 minute')
RETURNING proposal_id::text`, uid, args).Scan(&proposalID)
	if err != nil {
		t.Fatalf("insert expired proposal: %v", err)
	}
	t.Cleanup(func() {
		_, _ = testPool.Exec(context.Background(),
			`DELETE FROM gr33ncore.guardian_action_proposals WHERE proposal_id = $1`, proposalID)
	})

	tok := smokeJWT(t)
	body, _ := json.Marshal(map[string]string{"proposal_id": proposalID})
	req, _ := http.NewRequest(http.MethodPost, testServer.URL+"/v1/chat/confirm", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("confirm: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusGone {
		t.Fatalf("expired confirm want 410 Gone, got %d: %s", resp.StatusCode, readBodyPreview(resp))
	}
}

func phase46HybridAttach(
	ctx context.Context,
	q db.Querier,
	userID uuid.UUID,
	farmID int64,
	sessionID uuid.UUID,
	question string,
	assistantText string,
	snap farmguardian.Snapshot,
	policy farmguardian.LLMProposalPolicy,
	hasOperate, hasAdmin bool,
) ([]farmguardian.ActionProposal, error) {
	props, err := farmguardian.BuildRuleAssistedProposals(ctx, q, userID, farmID, sessionID, question, snap)
	if err != nil {
		return nil, err
	}
	if len(props) > 0 {
		return props, nil
	}
	return farmguardian.TryBuildLLMProposalsFromAssistant(
		ctx, q, userID, farmID, sessionID, question, assistantText, policy,
		hasOperate, hasAdmin, false,
		farmguardian.FreshMatcherMatches(ctx, q, farmID, question, snap),
	)
}

func countPendingProposals(ctx context.Context, userID uuid.UUID) int {
	var n int
	_ = testPool.QueryRow(ctx, `
SELECT COUNT(*) FROM gr33ncore.guardian_action_proposals
WHERE user_id = $1 AND status = 'pending'`, userID).Scan(&n)
	return n
}

func llmSmokeProposalJSON(tool string, args map[string]any, summary string) string {
	raw, _ := json.Marshal(map[string]any{
		"tool":       tool,
		"args":       args,
		"summary":    summary,
		"confidence": "high",
	})
	return fmt.Sprintf("```json\n%s\n```", raw)
}
