//go:build ollama

// Phase 112 — Guardian Ollama E2E smokes (requires LLM_BASE_URL + running Ollama).
package main

import (
	"context"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func requireOllamaE2E(t *testing.T) {
	t.Helper()
	if strings.TrimSpace(os.Getenv("LLM_BASE_URL")) == "" {
		t.Skip("LLM_BASE_URL not set")
	}
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
}

// pickLightChatModel prefers tinyllama/phi3 for ungrounded smokes on low-RAM laptops.
func pickLightChatModel(t *testing.T, models []any, serverDefault string) string {
	t.Helper()
	prefs := []string{"tinyllama", "phi3:mini", "phi3"}
	for _, pref := range prefs {
		for _, raw := range models {
			m, _ := raw.(map[string]any)
			name, _ := m["name"].(string)
			if strings.HasPrefix(name, pref) {
				return name
			}
		}
	}
	for _, raw := range models {
		m, _ := raw.(map[string]any)
		name, _ := m["name"].(string)
		if name != "" && name != serverDefault {
			return name
		}
	}
	if len(models) > 0 {
		m, _ := models[0].(map[string]any)
		name, _ := m["name"].(string)
		if name != "" {
			return name
		}
	}
	t.Skip("no chat model available for override test")
	return ""
}

func requireEnvModelPrefix(t *testing.T, prefix string) {
	t.Helper()
	envModel := strings.TrimSpace(os.Getenv("LLM_MODEL"))
	bare := strings.TrimSuffix(envModel, ":latest")
	if !strings.HasPrefix(bare, prefix) {
		t.Skipf("LLM_MODEL=%q — this smoke requires LLM_MODEL=%s", envModel, prefix)
	}
}

func skipIfOllamaOOM(t *testing.T, resp *http.Response) {
	t.Helper()
	if resp.StatusCode != http.StatusBadGateway {
		return
	}
	body := decodeMap(t, resp)
	resp.Body.Close()
	errMsg, _ := body["error"].(string)
	if strings.Contains(errMsg, "system memory") || strings.Contains(errMsg, "memory") {
		t.Skip("Ollama OOM — stop make dev-auth-test to free RAM, then retry ollama-smoke")
	}
}

func TestPhase112_ShowEnrichment(t *testing.T) {
	requireOllamaE2E(t)
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/guardian/models")
	expectStatus(t, resp, 200)
	body := decodeMap(t, resp)
	resp.Body.Close()
	models, _ := body["available_models"].([]any)
	if len(models) == 0 {
		t.Skip("no models in Ollama")
	}
	found := false
	for _, raw := range models {
		m, _ := raw.(map[string]any)
		if cw, ok := m["context_window"].(float64); ok && cw > 0 {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected at least one model with context_window > 0 after /api/show enrichment")
	}
}

func TestPhase112_SessionOverride(t *testing.T) {
	requireOllamaE2E(t)
	tok := smokeJWT(t)

	disc := authGet(t, tok, "/guardian/models")
	expectStatus(t, disc, 200)
	discBody := decodeMap(t, disc)
	disc.Body.Close()
	models, _ := discBody["available_models"].([]any)
	if len(models) == 0 {
		t.Skip("no ollama models")
	}
	serverDefault, _ := discBody["server_default"].(string)
	overrideModel := pickLightChatModel(t, models, serverDefault)

	// Omit farm_id (null) for ungrounded chat — farm_id: 0 is rejected as invalid.
	chatResp := authPost(t, tok, "/v1/chat", map[string]any{
		"message": "Say hello in one word.",
		"model":   overrideModel,
		"stream":  false,
	})
	if chatResp.StatusCode == http.StatusServiceUnavailable {
		chatResp.Body.Close()
		t.Skip("LLM not configured in test server")
	}
	if chatResp.StatusCode == http.StatusBadGateway {
		skipIfOllamaOOM(t, chatResp)
	}
	if chatResp.StatusCode != http.StatusOK && chatResp.StatusCode != http.StatusBadGateway {
		t.Fatalf("chat status %d", chatResp.StatusCode)
	}
	chatBody := decodeMap(t, chatResp)
	chatResp.Body.Close()
	sessionID, _ := chatBody["session_id"].(string)
	if sessionID == "" {
		t.Fatal("missing session_id")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var llmModel string
	err := testPool.QueryRow(ctx, `
SELECT llm_model FROM gr33ncore.conversation_turns
WHERE session_id = $1::uuid
ORDER BY turn_index DESC LIMIT 1`, sessionID).Scan(&llmModel)
	if err != nil {
		t.Fatalf("conversation turn: %v", err)
	}
	if llmModel != overrideModel {
		t.Fatalf("llm_model want %q, got %q", overrideModel, llmModel)
	}
}

func TestPhase112_FarmModelSwitchAudit(t *testing.T) {
	requireOllamaE2E(t)
	tok := smokeJWT(t)

	disc := authGet(t, tok, "/guardian/models")
	expectStatus(t, disc, 200)
	discBody := decodeMap(t, disc)
	disc.Body.Close()
	models, _ := discBody["available_models"].([]any)
	if len(models) == 0 {
		t.Skip("no ollama models")
	}
	first, _ := models[0].(map[string]any)
	modelName, _ := first["name"].(string)

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

func TestPhase112_ContextWindowGuardrail(t *testing.T) {
	requireOllamaE2E(t)
	tok := smokeJWT(t)

	disc := authGet(t, tok, "/guardian/models")
	expectStatus(t, disc, 200)
	discBody := decodeMap(t, disc)
	disc.Body.Close()
	models, _ := discBody["available_models"].([]any)
	var smallModel string
	for _, raw := range models {
		m, _ := raw.(map[string]any)
		name, _ := m["name"].(string)
		cw, _ := m["context_window"].(float64)
		if cw > 0 && cw < 8192 {
			smallModel = name
			break
		}
	}
	if smallModel == "" {
		t.Skip("no model with context_window < 8192 in cache (pull phi3:mini for this test)")
	}

	chatResp := authPost(t, tok, "/v1/chat", map[string]any{
		"message": "What is the EC target for zone 1?",
		"farm_id": 1,
		"model":   smallModel,
		"stream":  false,
	})
	if chatResp.StatusCode != http.StatusBadRequest {
		t.Fatalf("grounded small model want 400, got %d", chatResp.StatusCode)
	}
	chatResp.Body.Close()
}

func TestPhase112_FallbackOnMissingModel(t *testing.T) {
	requireOllamaE2E(t)
	tok := smokeJWT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := testPool.Exec(ctx, `
UPDATE gr33ncore.farms SET guardian_preferred_model = $1 WHERE id = 1`,
		"phase112-missing-model:smoke")
	if err != nil {
		t.Fatalf("seed farm model: %v", err)
	}
	defer func() {
		_, _ = testPool.Exec(context.Background(), `
UPDATE gr33ncore.farms SET guardian_preferred_model = NULL WHERE id = 1`)
	}()

	chatResp := authPost(t, tok, "/v1/chat", map[string]any{
		"message": "Hello",
		"farm_id": 1,
		"stream":  false,
	})
	if chatResp.StatusCode == http.StatusBadGateway {
		skipIfOllamaOOM(t, chatResp)
	}
	body := decodeMap(t, chatResp)
	chatResp.Body.Close()
	if chatResp.StatusCode == http.StatusBadRequest {
		errMsg, _ := body["error"].(string)
		if strings.Contains(errMsg, "below the minimum required for grounded") {
			// Missing farm model fell back to env default (e.g. tinyllama); Phase 118
			// guardrail correctly rejects grounded chat on small context windows.
			return
		}
		t.Fatalf("chat status 400: %s", errMsg)
	}
	if chatResp.StatusCode != http.StatusOK {
		t.Fatalf("chat status %d", chatResp.StatusCode)
	}
	fallback, _ := body["model_fallback"].(bool)
	if !fallback {
		t.Fatal("expected model_fallback true for missing farm model")
	}
}

func TestPhase112_PullThenDiscover(t *testing.T) {
	requireOllamaE2E(t)
	if !farmguardianLocalOllama() {
		t.Skip("LLM_BASE_URL is not local Ollama")
	}
	tok := smokeJWT(t)
	const pullName = "tinyllama"

	pull := authPost(t, tok, "/guardian/models/pull", map[string]any{
		"name":    pullName,
		"farm_id": 1,
	})
	if pull.StatusCode == http.StatusGatewayTimeout {
		pull.Body.Close()
		t.Skip("pull timed out")
	}
	expectStatus(t, pull, 200)
	pull.Body.Close()

	disc := authGet(t, tok, "/guardian/models")
	expectStatus(t, disc, 200)
	body := decodeMap(t, disc)
	disc.Body.Close()
	models, _ := body["available_models"].([]any)
	found := false
	for _, raw := range models {
		m, _ := raw.(map[string]any)
		name, _ := m["name"].(string)
		if strings.HasPrefix(name, pullName) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected %q in available_models after pull", pullName)
	}
}

func farmguardianLocalOllama() bool {
	base := strings.TrimSpace(os.Getenv("LLM_BASE_URL"))
	return strings.Contains(base, "127.0.0.1") || strings.Contains(base, "localhost")
}
