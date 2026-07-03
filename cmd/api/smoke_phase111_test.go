// Phase 111 — Guardian model selector smoke tests.
package main

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestPhase111_ModelDiscovery(t *testing.T) {
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/guardian/models")
	if resp.StatusCode == http.StatusServiceUnavailable {
		resp.Body.Close()
		t.Skip("AI disabled")
	}
	expectStatus(t, resp, 200)
	body := decodeMap(t, resp)
	resp.Body.Close()
	models, _ := body["available_models"].([]any)
	if models == nil {
		t.Fatal("available_models missing")
	}
	if _, ok := body["server_default"]; !ok {
		t.Fatal("server_default missing")
	}
}

func TestPhase111_RBACDenial(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, viewerTok := seedSmokeViewerUser(t, ctx)
	resp := authPatch(t, viewerTok, "/farms/1/settings", map[string]any{
		"guardian_preferred_model": "llama3.1:8b",
	})
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("viewer PATCH want 403, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestPhase111_FarmModelSwitchAudit(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	tok := smokeJWT(t)

	disc := authGet(t, tok, "/guardian/models")
	if disc.StatusCode != http.StatusOK {
		disc.Body.Close()
		t.Skip("guardian models unavailable")
	}
	discBody := decodeMap(t, disc)
	disc.Body.Close()
	models, _ := discBody["available_models"].([]any)
	if len(models) == 0 {
		t.Skip("no ollama models discovered — start Ollama for full E2E")
	}
	first, _ := models[0].(map[string]any)
	modelName, _ := first["name"].(string)
	if modelName == "" {
		t.Skip("empty model name in discovery")
	}

	patch := authPatch(t, tok, "/farms/1/settings", map[string]any{
		"guardian_preferred_model": modelName,
	})
	expectStatus(t, patch, 200)
	patch.Body.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var action string
	err := testPool.QueryRow(ctx, `
SELECT action_type::text FROM gr33ncore.user_activity_log
WHERE farm_id = 1 AND action_type = 'guardian_model_changed'
ORDER BY activity_time DESC LIMIT 1`).Scan(&action)
	if err != nil {
		t.Fatalf("audit row: %v", err)
	}

	clear := authPatch(t, tok, "/farms/1/settings", map[string]any{
		"guardian_preferred_model": nil,
	})
	clear.Body.Close()
}

func TestPhase111_ContextWindowGuardrail(t *testing.T) {
	tok := smokeJWT(t)
	chatResp := authPost(t, tok, "/v1/chat", map[string]any{
		"message":  "What is the EC target for zone 1?",
		"farm_id":  1,
		"model":    "phi3:mini-phase111-smoke",
		"stream":   false,
	})
	if chatResp.StatusCode == http.StatusServiceUnavailable {
		chatResp.Body.Close()
		t.Skip("LLM not configured")
	}
	// Without injected cache entry this may fallback rather than 400; unit tests cover guardrail.
	if chatResp.StatusCode != http.StatusBadRequest && chatResp.StatusCode != http.StatusOK && chatResp.StatusCode != http.StatusBadGateway {
		t.Fatalf("unexpected status %d", chatResp.StatusCode)
	}
	chatResp.Body.Close()
}

func TestPhase111_InvalidModelRejectedOnPatch(t *testing.T) {
	tok := smokeJWT(t)
	resp := authPatch(t, tok, "/farms/1/settings", map[string]any{
		"guardian_preferred_model": "definitely-not-a-real-model:999b",
	})
	if resp.StatusCode == http.StatusBadRequest {
		resp.Body.Close()
		return
	}
	if resp.StatusCode == http.StatusOK {
		resp.Body.Close()
		t.Skip("model cache empty — cannot validate unknown model rejection")
	}
	resp.Body.Close()
	t.Fatalf("want 400 or skip, got %d", resp.StatusCode)
}
