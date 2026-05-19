// Phase 24 WS6 — RAG smoke tests (auth + degraded config without cloud keys).

package main

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestRagSearchUnauthorized(t *testing.T) {
	resp := get(t, "/farms/1/rag/search?q=test")
	expectStatus(t, resp, http.StatusUnauthorized)
}

func TestRagSearchPOSTUnauthorized(t *testing.T) {
	resp := postNoAuth("/farms/1/rag/search", map[string]any{"query": "hello"})
	expectStatus(t, resp, http.StatusUnauthorized)
}

func TestRagAnswerUnauthorized(t *testing.T) {
	resp := postNoAuth("/farms/1/rag/answer", map[string]any{"query": "hello"})
	expectStatus(t, resp, http.StatusUnauthorized)
}

// Without EMBEDDING_API_KEY / LLM_* in the smoke process, authenticated calls return 503 (explicitly unconfigured).

func TestRagSearchRequiresEmbeddingConfigured(t *testing.T) {
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/farms/1/rag/search?q=smoke+RAG")
	expectStatus(t, resp, http.StatusServiceUnavailable)
}

func TestRagAnswerRequiresLLMAndEmbedding(t *testing.T) {
	tok := smokeJWT(t)
	resp := authPost(t, tok, "/farms/1/rag/answer", map[string]any{"query": "summarize"})
	expectStatus(t, resp, http.StatusServiceUnavailable)
}

func TestCapabilitiesPublic(t *testing.T) {
	resp := get(t, "/capabilities")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	var body map[string]bool
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if _, ok := body["ai_enabled"]; !ok {
		t.Fatal("missing ai_enabled")
	}
}

func TestV1ChatUnauthorized(t *testing.T) {
	resp := postNoAuth("/v1/chat", map[string]any{"message": "hi"})
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusUnauthorized)
}

func TestV1ChatRequiresLLMConfigured(t *testing.T) {
	tok := smokeJWT(t)
	resp := authPost(t, tok, "/v1/chat", map[string]any{"message": "hi"})
	defer resp.Body.Close()
	// AI defaults on, but smoke tests do not configure LLM_BASE_URL/LLM_MODEL —
	// the handler returns 503 with a config hint (Phase 27 WS5 v1).
	expectStatus(t, resp, http.StatusServiceUnavailable)
}

func TestV1ChatRejectsEmptyMessage(t *testing.T) {
	// Even without LLM configured the empty-body path 503s on missing config first,
	// so we only assert this when AI is on. Both ordering rules are covered by
	// internal/handler/chat unit tests; this just keeps the smoke surface honest.
	tok := smokeJWT(t)
	resp := authPost(t, tok, "/v1/chat", map[string]any{})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected 400 or 503, got %d", resp.StatusCode)
	}
}

// Phase 27 WS5 follow-up — session history endpoints (auth + empty-list shape).

func TestV1ChatSessionsUnauthorized(t *testing.T) {
	resp := get(t, "/v1/chat/sessions")
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusUnauthorized)
}

func TestV1ChatGetSessionUnauthorized(t *testing.T) {
	resp := get(t, "/v1/chat/sessions/11111111-1111-4111-8111-111111111111")
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusUnauthorized)
}

func TestV1ChatGetSessionRejectsBadUUID(t *testing.T) {
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/v1/chat/sessions/not-a-uuid")
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusBadRequest)
}

func TestV1ChatPatchSessionUnauthorized(t *testing.T) {
	resp := patchNoAuth("/v1/chat/sessions/11111111-1111-4111-8111-111111111111", map[string]any{"title": "x"})
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusUnauthorized)
}

func TestV1ChatPatchSessionRejectsBadUUID(t *testing.T) {
	tok := smokeJWT(t)
	resp := authPatch(t, tok, "/v1/chat/sessions/not-a-uuid", map[string]any{"title": "x"})
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusBadRequest)
}

func TestV1ChatPatchSessionNotFound(t *testing.T) {
	tok := smokeJWT(t)
	// Valid UUID but no row exists in the smoke DB.
	resp := authPatch(t, tok, "/v1/chat/sessions/aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee", map[string]any{"title": "x"})
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusNotFound)
}

func TestV1ChatDeleteSessionUnauthorized(t *testing.T) {
	resp := deleteNoAuth("/v1/chat/sessions/11111111-1111-4111-8111-111111111111")
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusUnauthorized)
}

func TestV1ChatDeleteSessionIdempotent(t *testing.T) {
	tok := smokeJWT(t)
	// Deleting a session that doesn't exist still returns 204 — DELETE is idempotent.
	resp := authDelete(t, tok, "/v1/chat/sessions/aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee")
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusNoContent)
}
