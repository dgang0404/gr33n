//go:build ollama

// Phase 118 — Guardian model capabilities Ollama E2E smokes.
package main

import (
	"context"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func TestPhase118_ChatModelsExcludeEmbeddingOnly(t *testing.T) {
	requireOllamaE2E(t)
	tok := smokeJWT(t)

	resp := authGet(t, tok, "/guardian/models")
	expectStatus(t, resp, 200)
	body := decodeMap(t, resp)
	resp.Body.Close()
	models, _ := body["available_models"].([]any)
	for _, raw := range models {
		m, _ := raw.(map[string]any)
		caps, _ := m["capabilities"].([]any)
		if len(caps) == 1 {
			if c, _ := caps[0].(string); c == "embedding" {
				t.Fatalf("embedding-only model %v should not appear in default list", m["name"])
			}
		}
	}

	allResp := authGet(t, tok, "/guardian/models?all=true")
	expectStatus(t, allResp, 200)
	allBody := decodeMap(t, allResp)
	allResp.Body.Close()
	allModels, _ := allBody["available_models"].([]any)
	if len(allModels) < len(models) {
		t.Fatalf("all=true should be >= chat list: all=%d chat=%d", len(allModels), len(models))
	}
}

func TestPhase118_EnvDefaultTagNormalizationGuardrail(t *testing.T) {
	requireOllamaE2E(t)
	tok := smokeJWT(t)

	envModel := strings.TrimSpace(os.Getenv("LLM_MODEL"))
	if envModel == "" {
		t.Skip("LLM_MODEL not set")
	}

	disc := authGet(t, tok, "/guardian/models")
	expectStatus(t, disc, 200)
	discBody := decodeMap(t, disc)
	disc.Body.Close()
	models, _ := discBody["available_models"].([]any)

	var tinyCtx float64
	var tinyName string
	for _, raw := range models {
		m, _ := raw.(map[string]any)
		name, _ := m["name"].(string)
		cw, _ := m["context_window"].(float64)
		if strings.HasPrefix(name, "tinyllama") && cw > 0 && cw < 8192 {
			tinyCtx = cw
			tinyName = name
			break
		}
	}
	if tinyName == "" {
		t.Skip("tinyllama with context_window < 8192 not in chat list — pull tinyllama for this test")
	}
	_ = tinyCtx

	// Clear session/farm overrides so resolution uses env default (often bare "tinyllama").
	ctx, cancel := contextWithTimeout(t)
	defer cancel()
	_, _ = testPool.Exec(ctx, `UPDATE gr33ncore.farms SET guardian_preferred_model = NULL WHERE id = 1`)

	grounded := authPost(t, tok, "/v1/chat", map[string]any{
		"message": "What is the EC target for zone 1?",
		"farm_id": 1,
		"stream":  false,
	})
	if grounded.StatusCode != http.StatusBadRequest {
		t.Fatalf("grounded env-default tinyllama want 400, got %d", grounded.StatusCode)
	}
	grounded.Body.Close()

	ungrounded := authPost(t, tok, "/v1/chat", map[string]any{
		"message": "Say hello in one word.",
		"farm_id": 0,
		"stream":  false,
	})
	if ungrounded.StatusCode != http.StatusOK && ungrounded.StatusCode != http.StatusBadGateway {
		t.Fatalf("ungrounded tinyllama want 200/502, got %d", ungrounded.StatusCode)
	}
	ungrounded.Body.Close()
}

func TestPhase118_ExplicitAndEnvDefaultSameGuardrail(t *testing.T) {
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
		t.Skip("no small-context chat model available")
	}

	bare := strings.TrimSuffix(smallModel, ":latest")

	respExplicit := authPost(t, tok, "/v1/chat", map[string]any{
		"message": "What is the EC target for zone 1?",
		"farm_id": 1,
		"model":   smallModel,
		"stream":  false,
	})
	explicitCode := respExplicit.StatusCode
	respExplicit.Body.Close()

	respBare := authPost(t, tok, "/v1/chat", map[string]any{
		"message": "What is the EC target for zone 1?",
		"farm_id": 1,
		"model":   bare,
		"stream":  false,
	})
	bareCode := respBare.StatusCode
	respBare.Body.Close()

	if explicitCode != http.StatusBadRequest || bareCode != http.StatusBadRequest {
		t.Fatalf("both spellings should 400 for grounded small model: explicit=%d bare=%d", explicitCode, bareCode)
	}
}

func TestPhase118_ModelsExposeRuntimeHints(t *testing.T) {
	requireOllamaE2E(t)
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/guardian/models")
	expectStatus(t, resp, 200)
	body := decodeMap(t, resp)
	resp.Body.Close()
	models, _ := body["available_models"].([]any)
	if len(models) == 0 {
		t.Skip("no models")
	}
	m, _ := models[0].(map[string]any)
	if _, ok := m["runtime_hint"]; !ok {
		t.Fatalf("expected runtime_hint field: %#v", m)
	}
}

func contextWithTimeout(t *testing.T) (context.Context, context.CancelFunc) {
	t.Helper()
	return context.WithTimeout(context.Background(), 5*time.Second)
}
