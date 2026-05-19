// Phase 27 WS5 follow-up — smoke test for conversation TTL pruning.
//
// Exercises farmguardian.PruneOnce against a real Postgres so we know the
// SQL contracts hold (DELETE … WHERE session_id IN (... HAVING MAX(...) <
// cutoff)) and that pruning only touches rows the operator scheduled it to
// touch. The smoke harness (smoke_test.go) already applies the Phase 27
// migrations.

package main

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmguardian"
)

func TestPruneOnce_RemovesStaleSessionsLeavesFresh(t *testing.T) {
	if testPool == nil {
		t.Skip("test DB not initialised")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	userID := uuid.MustParse(smokeDevUserUUID)
	staleSessionID := uuid.New()
	freshSessionID := uuid.New()

	// Make sure no leftover state from a previous run clouds the assertions.
	t.Cleanup(func() {
		_, _ = testPool.Exec(context.Background(),
			`DELETE FROM gr33ncore.conversation_turns
			   WHERE session_id IN ($1, $2)`,
			staleSessionID, freshSessionID)
		_, _ = testPool.Exec(context.Background(),
			`DELETE FROM gr33ncore.conversation_sessions
			   WHERE id IN ($1, $2)`,
			staleSessionID, freshSessionID)
	})

	// Insert one stale turn (created 100 days ago) and one fresh turn (now).
	// We bypass InsertConversationTurn here because we need to backdate
	// created_at; the column has DEFAULT NOW() but accepts an explicit value.
	staleAt := time.Now().Add(-100 * 24 * time.Hour)
	if _, err := testPool.Exec(ctx, `
INSERT INTO gr33ncore.conversation_turns
    (session_id, user_id, turn_index, user_message, assistant_message,
     llm_model, grounded, context_count, citations,
     prompt_tokens, completion_tokens, created_at)
VALUES ($1, $2, 0, 'old', 'old reply',
        'test-model', false, 0, '[]'::jsonb, 0, 0, $3)`,
		staleSessionID, userID, staleAt); err != nil {
		t.Fatalf("insert stale turn: %v", err)
	}
	if _, err := testPool.Exec(ctx, `
INSERT INTO gr33ncore.conversation_sessions (id, user_id, created_at, updated_at)
VALUES ($1, $2, $3, $3)
ON CONFLICT (id) DO UPDATE SET updated_at = EXCLUDED.updated_at`,
		staleSessionID, userID, staleAt); err != nil {
		t.Fatalf("insert stale session: %v", err)
	}

	if _, err := testPool.Exec(ctx, `
INSERT INTO gr33ncore.conversation_turns
    (session_id, user_id, turn_index, user_message, assistant_message,
     llm_model, grounded, context_count, citations,
     prompt_tokens, completion_tokens)
VALUES ($1, $2, 0, 'new', 'new reply', 'test-model', false, 0, '[]'::jsonb, 0, 0)`,
		freshSessionID, userID); err != nil {
		t.Fatalf("insert fresh turn: %v", err)
	}
	if _, err := testPool.Exec(ctx, `
INSERT INTO gr33ncore.conversation_sessions (id, user_id)
VALUES ($1, $2)
ON CONFLICT (id) DO UPDATE SET updated_at = NOW()`,
		freshSessionID, userID); err != nil {
		t.Fatalf("insert fresh session: %v", err)
	}

	// TTL = 30 days; the 100-day-old session should go, the fresh one stays.
	q := db.New(testPool)
	res, err := farmguardian.PruneOnce(ctx, q, 30*24*time.Hour)
	if err != nil {
		t.Fatalf("PruneOnce: %v", err)
	}
	if res.TurnsDeleted < 1 {
		t.Fatalf("expected at least 1 turn deleted, got %d", res.TurnsDeleted)
	}
	if res.SessionsDeleted < 1 {
		t.Fatalf("expected at least 1 session deleted, got %d", res.SessionsDeleted)
	}

	// Verify: stale session id is gone, fresh remains.
	var staleTurnCount, freshTurnCount int
	if err := testPool.QueryRow(ctx,
		`SELECT COUNT(*) FROM gr33ncore.conversation_turns WHERE session_id = $1`,
		staleSessionID).Scan(&staleTurnCount); err != nil {
		t.Fatalf("count stale turns: %v", err)
	}
	if staleTurnCount != 0 {
		t.Fatalf("stale turns survived prune: %d", staleTurnCount)
	}
	if err := testPool.QueryRow(ctx,
		`SELECT COUNT(*) FROM gr33ncore.conversation_turns WHERE session_id = $1`,
		freshSessionID).Scan(&freshTurnCount); err != nil {
		t.Fatalf("count fresh turns: %v", err)
	}
	if freshTurnCount != 1 {
		t.Fatalf("fresh turn was pruned (count=%d) — TTL clamp leaked", freshTurnCount)
	}

	var staleSessionCount, freshSessionCount int
	if err := testPool.QueryRow(ctx,
		`SELECT COUNT(*) FROM gr33ncore.conversation_sessions WHERE id = $1`,
		staleSessionID).Scan(&staleSessionCount); err != nil {
		t.Fatalf("count stale sessions: %v", err)
	}
	if staleSessionCount != 0 {
		t.Fatalf("stale session metadata survived prune: %d", staleSessionCount)
	}
	if err := testPool.QueryRow(ctx,
		`SELECT COUNT(*) FROM gr33ncore.conversation_sessions WHERE id = $1`,
		freshSessionID).Scan(&freshSessionCount); err != nil {
		t.Fatalf("count fresh sessions: %v", err)
	}
	if freshSessionCount != 1 {
		t.Fatalf("fresh session metadata was pruned (count=%d)", freshSessionCount)
	}
}
