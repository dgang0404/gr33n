// Phase 24 WS6 — RAG smoke tests (auth + degraded config without cloud keys).

package main

import (
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
