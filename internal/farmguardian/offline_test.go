package farmguardian

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsLocalInferenceURL(t *testing.T) {
	cases := []struct {
		url  string
		want bool
	}{
		{"http://127.0.0.1:11434/v1", true},
		{"http://localhost:11434/v1", true},
		{"http://192.168.1.50:11434/v1", true},
		{"http://10.0.0.2/v1", true},
		{"https://api.openai.com/v1", false},
		{"", false},
	}
	for _, tc := range cases {
		if got := IsLocalInferenceURL(tc.url); got != tc.want {
			t.Errorf("IsLocalInferenceURL(%q) = %v, want %v", tc.url, got, tc.want)
		}
	}
}

func TestBuildFieldAssistantHealth_FieldMode(t *testing.T) {
	t.Setenv("LLM_BASE_URL", "http://127.0.0.1:11434/v1")
	t.Setenv("EMBEDDING_BASE_URL", "http://127.0.0.1:11434/v1")
	h := BuildFieldAssistantHealth(t.Context(), func(context.Context, string, string) error { return nil }, 3, 10)
	if !h.FieldMode {
		t.Fatal("expected field_mode true for loopback LLM")
	}
	if !h.LLMReachable {
		t.Fatal("expected llm_reachable when probe succeeds")
	}
	if !h.EmbeddingReachable {
		t.Fatal("expected embedding_reachable when probe succeeds")
	}
	if h.FieldGuideChunkCount != 3 || h.PlatformDocChunkCount != 10 {
		t.Fatalf("chunk counts: field=%d platform=%d", h.FieldGuideChunkCount, h.PlatformDocChunkCount)
	}
}

func TestBuildFieldAssistantHealth_SplitHosts(t *testing.T) {
	chat := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/models" || r.URL.Path == "/v1/models" {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.NotFound(w, r)
	}))
	defer chat.Close()
	embed := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/models" || r.URL.Path == "/v1/models" {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.NotFound(w, r)
	}))
	defer embed.Close()
	t.Setenv("LLM_BASE_URL", chat.URL+"/v1")
	t.Setenv("EMBEDDING_BASE_URL", embed.URL+"/v1")
	h := BuildFieldAssistantHealth(t.Context(), nil, 0, 0)
	if !h.SplitInferenceHosts {
		t.Fatal("expected split_inference_hosts")
	}
	if !h.LLMReachable || !h.EmbeddingReachable {
		t.Fatalf("reachability chat=%v embed=%v", h.LLMReachable, h.EmbeddingReachable)
	}
}
