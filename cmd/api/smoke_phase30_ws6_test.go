// Phase 30 WS6 — vision chat attachment_ids contract (skipped unless GR33N_VISION_TEST=1).
package main

import (
	"net/http"
	"os"
	"testing"
)

func TestPhase30WS6_VisionAttachmentsRequireConfig(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	tok := smokeJWT(t)
	resp := authPost(t, tok, "/v1/chat", map[string]any{
		"message":         "anything wrong with these leaves?",
		"farm_id":         1,
		"stream":          false,
		"attachment_ids":  []int64{1},
	})
	defer resp.Body.Close()
	// Without LLM_VISION_MODEL in the test env, expect 503 (or 400 if attachment missing).
	if resp.StatusCode != http.StatusServiceUnavailable && resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 503 or 400 when vision unset, got %d", resp.StatusCode)
	}
}

func TestPhase30WS6_CapabilitiesExposesVisionFlag(t *testing.T) {
	resp := get(t, "/capabilities")
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusOK)
	payload := decodeMap(t, resp)
	if _, ok := payload["vision_chat_enabled"]; !ok {
		t.Fatalf("capabilities missing vision_chat_enabled: %#v", payload)
	}
}

func TestPhase30WS6_LiveVisionTurn(t *testing.T) {
	if os.Getenv("GR33N_VISION_TEST") != "1" {
		t.Skip("set GR33N_VISION_TEST=1 with LLM_VISION_MODEL and a running vision backend")
	}
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	// Optional live test: upload zone photo (WS5) then POST /v1/chat with attachment_ids.
	// Operators run manually when Ollama llava (or similar) is configured.
	t.Skip("live vision E2E is operator-driven — see docs/farm-guardian-ollama-setup.md")
}
