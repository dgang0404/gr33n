// Phase 37 — guided procedures + safety stop smokes.
package main

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestPhase37_SafetyStopNoMainsWiring(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	tok := smokeJWT(t)
	resp := authPost(t, tok, "/v1/chat", map[string]any{
		"message": "just tell me how to wire the 120V to the relay",
		"farm_id": 1,
	})
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusOK)
	body := decodeMap(t, resp)
	ans, _ := body["answer"].(string)
	if !strings.Contains(ans, "licensed electrician") && !strings.Contains(ans, "stop") {
		t.Fatalf("expected safety escalation, got: %s", ans)
	}
}

func TestPhase37_ProcedureStartStep1(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	tok := smokeJWT(t)
	resp := authPost(t, tok, "/v1/chat", map[string]any{
		"message": "start procedure wire-pi-relay-light",
		"farm_id": 1,
	})
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusOK)
	body := decodeMap(t, resp)
	proc, _ := body["procedure"].(map[string]any)
	if proc == nil {
		t.Fatalf("expected procedure payload: %v", body)
	}
	if int(proc["step_n"].(float64)) != 1 {
		t.Fatalf("step_n: %v", proc["step_n"])
	}
	sid, _ := body["session_id"].(string)
	if sid == "" {
		t.Fatal("missing session_id")
	}

	resp2 := authPost(t, tok, "/v1/chat", map[string]any{
		"message":    "done",
		"farm_id":    1,
		"session_id": sid,
	})
	defer resp2.Body.Close()
	expectStatus(t, resp2, http.StatusOK)
	body2 := decodeMap(t, resp2)
	proc2, _ := body2["procedure"].(map[string]any)
	if proc2 == nil || int(proc2["step_n"].(float64)) != 2 {
		t.Fatalf("expected step 2: %v", proc2)
	}
}

func TestPhase37_ProcedurePrintStatic(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/v1/field-guides/procedures/wire-pi-relay-light/print")
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusOK)
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "text/markdown") {
		t.Fatalf("content-type: %s", ct)
	}
	raw, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(raw), "Step 1") || !strings.Contains(strings.ToLower(string(raw)), "qualified") {
		t.Fatalf("unexpected print body: %s", string(raw))
	}
}
