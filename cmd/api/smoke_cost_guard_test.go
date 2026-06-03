// Phase 27 WS5 follow-up — smoke test for chat cost guards.
//
// Exercises farmguardian.CheckBudget against a real Postgres so we know the
// SUM(prompt_tokens + completion_tokens) query rolls up correctly and the
// per-user / per-farm dimensions are wired the way the chat handler expects.
//
// The HTTP path is skipped because the smoke harness has no LLM configured
// (POST /v1/chat returns 503 before the guard ever runs). The cost-guard
// module unit tests cover the 429 response shape with mocks; this test
// pins down the SQL contract against the real DB.

package main

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmguardian"
)

func TestCostGuard_BlocksWhenUserBudgetExceeded(t *testing.T) {
	if testPool == nil {
		t.Skip("test DB not initialised")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	userID := uuid.MustParse(smokeDevUserUUID)
	sessionA := uuid.New()
	sessionB := uuid.New()

	t.Cleanup(func() {
		_, _ = testPool.Exec(context.Background(),
			`DELETE FROM gr33ncore.conversation_turns
			   WHERE session_id IN ($1, $2)`,
			sessionA, sessionB)
		_, _ = testPool.Exec(context.Background(),
			`DELETE FROM gr33ncore.conversation_sessions
			   WHERE id IN ($1, $2)`,
			sessionA, sessionB)
	})

	insertTurn := func(session uuid.UUID, turnIdx int, prompt, completion int64) {
		t.Helper()
		if _, err := testPool.Exec(ctx, `
INSERT INTO gr33ncore.conversation_sessions (id, user_id)
VALUES ($1, $2)
ON CONFLICT (id) DO UPDATE SET updated_at = NOW()`,
			session, userID); err != nil {
			t.Fatalf("insert session %d: %v", turnIdx, err)
		}
		if _, err := testPool.Exec(ctx, `
INSERT INTO gr33ncore.conversation_turns
    (session_id, user_id, turn_index, user_message, assistant_message,
     llm_model, grounded, context_count, citations,
     prompt_tokens, completion_tokens)
VALUES ($1, $2, $3, 'q', 'a', 'test-model', false, 0, '[]'::jsonb, $4, $5)`,
			session, userID, turnIdx, prompt, completion); err != nil {
			t.Fatalf("insert turn %d: %v", turnIdx, err)
		}
	}

	insertTurn(sessionA, 0, 1500, 500) // 2k tokens
	insertTurn(sessionA, 1, 2500, 500) // +3k → 5k running total in session A
	insertTurn(sessionB, 0, 500, 100)  // +600 → 5.6k across the user

	q := db.New(testPool)

	// Sanity: the SUM query rolls up across sessions for the same user.
	totals, err := q.SumChatTokensSinceForUser(ctx, db.SumChatTokensSinceForUserParams{UserID: userID, Since: time.Now().Add(-time.Hour)})
	if err != nil {
		t.Fatalf("SumChatTokensSinceForUser: %v", err)
	}
	if totals.TotalTokens != 5600 {
		t.Fatalf("rolled-up total = %d, want 5600 (3 turns: 2000+3000+600)", totals.TotalTokens)
	}
	if totals.PromptTokens != 4500 || totals.CompletionTokens != 1100 {
		t.Fatalf("prompt/completion split: got %d / %d, want 4500 / 1100",
			totals.PromptTokens, totals.CompletionTokens)
	}

	// Cap below the rolled-up total → guard should fire with per_user reason.
	tight := farmguardian.CostGuardConfig{
		Window:           time.Hour,
		PerUserMaxTokens: 5_000,
	}
	tightDecision, err := farmguardian.CheckBudget(ctx, q, tight, userID, 0)
	if err != nil {
		t.Fatalf("CheckBudget (tight): %v", err)
	}
	if tightDecision.Allowed {
		t.Fatalf("expected guard to block at 5k cap when total is 5.6k")
	}
	if tightDecision.Reason != "per_user" {
		t.Fatalf("reason = %q want per_user", tightDecision.Reason)
	}
	if tightDecision.UsedTokens != 5600 || tightDecision.MaxTokens != 5_000 {
		t.Fatalf("decision counters: used=%d max=%d (want 5600 / 5000)",
			tightDecision.UsedTokens, tightDecision.MaxTokens)
	}
	if tightDecision.RetryAfter != time.Hour {
		t.Fatalf("retry_after = %s, want 1h", tightDecision.RetryAfter)
	}

	// Cap above the rolled-up total → guard should let the request through.
	loose := farmguardian.CostGuardConfig{
		Window:           time.Hour,
		PerUserMaxTokens: 10_000,
	}
	looseDecision, err := farmguardian.CheckBudget(ctx, q, loose, userID, 0)
	if err != nil {
		t.Fatalf("CheckBudget (loose): %v", err)
	}
	if !looseDecision.Allowed {
		t.Fatalf("expected guard to allow at 10k cap when total is 5.6k (decision=%+v)", looseDecision)
	}

	// Disabled config (both caps zero) → fast path, no error, no DB error
	// even if the DB were stubbed empty.
	disabled, err := farmguardian.CheckBudget(ctx, q, farmguardian.CostGuardConfig{}, userID, 0)
	if err != nil {
		t.Fatalf("CheckBudget (disabled): %v", err)
	}
	if !disabled.Allowed {
		t.Fatalf("disabled guard must always allow")
	}
}

func TestCostGuard_PerFarmDimensionRollsUpByFarm(t *testing.T) {
	if testPool == nil {
		t.Skip("test DB not initialised")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	userID := uuid.MustParse(smokeDevUserUUID)
	const farmID = int64(1)
	sessionGrounded := uuid.New()

	t.Cleanup(func() {
		_, _ = testPool.Exec(context.Background(),
			`DELETE FROM gr33ncore.conversation_turns WHERE session_id = $1`,
			sessionGrounded)
		_, _ = testPool.Exec(context.Background(),
			`DELETE FROM gr33ncore.conversation_sessions WHERE id = $1`,
			sessionGrounded)
	})

	if _, err := testPool.Exec(ctx, `
INSERT INTO gr33ncore.conversation_sessions (id, user_id)
VALUES ($1, $2)
ON CONFLICT (id) DO UPDATE SET updated_at = NOW()`,
		sessionGrounded, userID); err != nil {
		t.Fatalf("insert session: %v", err)
	}

	for i := 0; i < 3; i++ {
		if _, err := testPool.Exec(ctx, `
INSERT INTO gr33ncore.conversation_turns
    (session_id, user_id, farm_id, turn_index, user_message, assistant_message,
     llm_model, grounded, context_count, citations,
     prompt_tokens, completion_tokens)
VALUES ($1, $2, $3, $4, 'q', 'a', 'test-model', true, 0, '[]'::jsonb, 1000, 200)`,
			sessionGrounded, userID, farmID, i); err != nil {
			t.Fatalf("insert grounded turn %d: %v", i, err)
		}
	}

	q := db.New(testPool)
	farmIDPtr := farmID
	farmTotals, err := q.SumChatTokensSinceForFarm(ctx, db.SumChatTokensSinceForFarmParams{FarmID: &farmIDPtr, Since: time.Now().Add(-time.Hour)})
	if err != nil {
		t.Fatalf("SumChatTokensSinceForFarm: %v", err)
	}
	if farmTotals.TotalTokens < 3*1200 {
		t.Fatalf("farm rollup = %d, want >= 3600 (3 turns × 1200)", farmTotals.TotalTokens)
	}

	cfg := farmguardian.CostGuardConfig{
		Window:           time.Hour,
		PerUserMaxTokens: 0, // disabled → must not short-circuit
		PerFarmMaxTokens: farmTotals.TotalTokens - 1,
	}
	d, err := farmguardian.CheckBudget(ctx, q, cfg, userID, farmID)
	if err != nil {
		t.Fatalf("CheckBudget: %v", err)
	}
	if d.Allowed {
		t.Fatalf("expected per_farm guard to fire")
	}
	if d.Reason != "per_farm" {
		t.Fatalf("reason = %q want per_farm", d.Reason)
	}

	// farm_id = 0 → per-farm dimension must be skipped, request allowed.
	noFarm, err := farmguardian.CheckBudget(ctx, q, cfg, userID, 0)
	if err != nil {
		t.Fatalf("CheckBudget (no farm): %v", err)
	}
	if !noFarm.Allowed {
		t.Fatalf("farm cap must not apply when farmID = 0")
	}
}

func TestCostGuard_WindowExcludesOldTurns(t *testing.T) {
	if testPool == nil {
		t.Skip("test DB not initialised")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	userID := uuid.MustParse(smokeDevUserUUID)
	session := uuid.New()

	t.Cleanup(func() {
		_, _ = testPool.Exec(context.Background(),
			`DELETE FROM gr33ncore.conversation_turns WHERE session_id = $1`, session)
		_, _ = testPool.Exec(context.Background(),
			`DELETE FROM gr33ncore.conversation_sessions WHERE id = $1`, session)
	})

	if _, err := testPool.Exec(ctx, `
INSERT INTO gr33ncore.conversation_sessions (id, user_id)
VALUES ($1, $2)
ON CONFLICT (id) DO UPDATE SET updated_at = NOW()`,
		session, userID); err != nil {
		t.Fatalf("insert session: %v", err)
	}

	// One ancient turn (way outside the 1h window) and one fresh turn.
	if _, err := testPool.Exec(ctx, `
INSERT INTO gr33ncore.conversation_turns
    (session_id, user_id, turn_index, user_message, assistant_message,
     llm_model, grounded, context_count, citations,
     prompt_tokens, completion_tokens, created_at)
VALUES ($1, $2, 0, 'q', 'a', 'test-model', false, 0, '[]'::jsonb,
        100000, 100000, NOW() - INTERVAL '24 hours')`,
		session, userID); err != nil {
		t.Fatalf("insert ancient turn: %v", err)
	}
	if _, err := testPool.Exec(ctx, `
INSERT INTO gr33ncore.conversation_turns
    (session_id, user_id, turn_index, user_message, assistant_message,
     llm_model, grounded, context_count, citations,
     prompt_tokens, completion_tokens)
VALUES ($1, $2, 1, 'q', 'a', 'test-model', false, 0, '[]'::jsonb, 50, 50)`,
		session, userID); err != nil {
		t.Fatalf("insert fresh turn: %v", err)
	}

	q := db.New(testPool)

	totals, err := q.SumChatTokensSinceForUser(ctx, db.SumChatTokensSinceForUserParams{UserID: userID, Since: time.Now().Add(-time.Hour)})
	if err != nil {
		t.Fatalf("SumChatTokensSinceForUser: %v", err)
	}
	if totals.TotalTokens != 100 {
		t.Fatalf("window must exclude the 24h-old 200k-token turn; got total=%d (want 100)",
			totals.TotalTokens)
	}

	cfg := farmguardian.CostGuardConfig{
		Window:           time.Hour,
		PerUserMaxTokens: 1_000, // way above the 100 inside the window
	}
	d, err := farmguardian.CheckBudget(ctx, q, cfg, userID, 0)
	if err != nil {
		t.Fatalf("CheckBudget: %v", err)
	}
	if !d.Allowed {
		t.Fatalf("expected Allowed; 1h-window total is 100, cap 1000 (decision=%+v)", d)
	}
}
