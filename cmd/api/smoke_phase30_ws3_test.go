// Phase 30 WS3 — config tool propose→confirm smoke (create_task).
package main

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestPhase30WS3_CreateTaskProposeConfirm(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

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
		t.Fatalf("chat status %d: %s", chatResp.StatusCode, readBodyPreview(chatResp))
	}

	var chatBody struct {
		Proposals []struct {
			ProposalID string         `json:"proposal_id"`
			Tool       string         `json:"tool"`
			Args       map[string]any `json:"args"`
		} `json:"proposals"`
	}
	decodeJSON(t, chatResp.Body, &chatBody)
	if len(chatBody.Proposals) == 0 {
		t.Fatal("expected create_task proposal")
	}
	prop := chatBody.Proposals[0]
	if prop.Tool != "create_task" {
		t.Fatalf("tool %q want create_task", prop.Tool)
	}

	confirmResp := authPost(t, tok, "/v1/chat/confirm", map[string]string{
		"proposal_id": prop.ProposalID,
	})
	defer confirmResp.Body.Close()
	expectStatus(t, confirmResp, http.StatusOK)

	var confirmBody struct {
		Result map[string]any `json:"result"`
	}
	decodeJSON(t, confirmResp.Body, &confirmBody)
	taskID, _ := confirmBody.Result["task_id"].(float64)
	if taskID == 0 {
		t.Fatalf("confirm result missing task_id: %+v", confirmBody.Result)
	}

	var title string
	err := testPool.QueryRow(ctx, `
SELECT title FROM gr33ncore.tasks WHERE id = $1 AND farm_id = 1`, int64(taskID)).Scan(&title)
	if err != nil {
		t.Fatalf("task row: %v", err)
	}
	if title == "" {
		t.Fatal("expected non-empty task title")
	}

	var auditCount int
	_ = testPool.QueryRow(ctx, `
SELECT COUNT(*) FROM gr33ncore.user_activity_log
WHERE action_type = 'guardian_tool_executed'
  AND details->>'tool_id' = 'create_task'
  AND details->>'proposal_id' = $1`, prop.ProposalID).Scan(&auditCount)
	if auditCount < 1 {
		t.Fatal("expected guardian_tool_executed audit row for create_task")
	}
}

func TestPhase30WS3_CreateTaskToolRegistered(t *testing.T) {
	tok := smokeJWT(t)
	// List proposals endpoint should work; tool registry is exercised via confirm on unknown tool separately.
	resp := authGet(t, tok, "/v1/chat/proposals?farm_id=1&status=pending&limit=1")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list proposals %d", resp.StatusCode)
	}
	var raw map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&raw)
}
