// Phase 134 — conversation turn feedback PATCH + export.

package main

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
)

func TestPhase134_TurnFeedbackPatchAndExport(t *testing.T) {
	tok := smokeJWT(t)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	sessionID := uuid.New()

	_, err := testPool.Exec(context.Background(), `
INSERT INTO gr33ncore.conversation_turns (
  session_id, user_id, farm_id, turn_index,
  user_message, assistant_message, llm_model, grounded
) VALUES ($1, $2, 1, 0, 'morning?', 'check alerts first', 'phi3:mini', true)
`, sessionID, userID)
	if err != nil {
		t.Fatalf("insert turn: %v", err)
	}
	t.Cleanup(func() {
		_, _ = testPool.Exec(context.Background(),
			`DELETE FROM gr33ncore.conversation_turns WHERE session_id = $1`, sessionID)
	})

	path := fmt.Sprintf("/v1/chat/sessions/%s/turns/0/feedback", sessionID)
	resp := authPatch(t, tok, path, map[string]any{
		"rating": "down",
		"reason": "Missed alert",
	})
	expectStatus(t, resp, http.StatusOK)
	body := decodeMap(t, resp)
	if body["feedback_rating"] != "down" {
		t.Fatalf("rating: %v", body["feedback_rating"])
	}

	resp = authGet(t, tok, "/v1/chat/feedback/export?farm_id=1&since=1d")
	expectStatus(t, resp, http.StatusOK)
	export := decodeMap(t, resp)
	rows, ok := export["rows"].([]any)
	if !ok || len(rows) == 0 {
		t.Fatalf("expected export rows, got %v", export["rows"])
	}
}
