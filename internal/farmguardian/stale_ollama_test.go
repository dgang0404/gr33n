package farmguardian

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestDetectStaleOllamaCLI_whenPgrepAndEmptyPS(t *testing.T) {
	staleOllamaRunCount = func(ctx context.Context) int { return 2 }
	t.Cleanup(func() { staleOllamaRunCount = defaultStaleOllamaRunCount })

	srv := startMockOllamaPS(t, nil)
	t.Setenv("LLM_BASE_URL", srv.URL+"/v1")
	if !DetectStaleOllamaCLI(t.Context(), srv.URL+"/v1") {
		t.Fatal("expected stale ollama cli")
	}
}

func TestDetectStaleOllamaCLI_falseWhenModelsLoaded(t *testing.T) {
	staleOllamaRunCount = func(ctx context.Context) int { return 3 }
	t.Cleanup(func() { staleOllamaRunCount = defaultStaleOllamaRunCount })

	srv := startMockOllamaPS(t, []map[string]any{
		{"name": "phi3:mini", "size_vram": 0},
	})
	if DetectStaleOllamaCLI(t.Context(), srv.URL+"/v1") {
		t.Fatal("expected false when ps has models")
	}
}

func TestBuildAwakeningHealth_StaleOllamaHint(t *testing.T) {
	staleOllamaRunCount = func(ctx context.Context) int { return 1 }
	t.Cleanup(func() { staleOllamaRunCount = defaultStaleOllamaRunCount })

	srv := startMockOllamaPS(t, nil)
	t.Setenv("LLM_BASE_URL", srv.URL+"/v1")
	field := BuildFieldAssistantHealth(t.Context(), nil, 5, 2)
	h := BuildAwakeningHealth(t.Context(), AwakeningBuildInput{
		AIEnabled:         true,
		Field:             field,
		Mode:              WarmupModeFarmCounsel,
		FieldGuideChunks:  5,
		PlatformDocChunks: 2,
		EnvDefault:        "phi3:mini",
	})
	if !h.StaleOllamaCLI {
		t.Fatal("expected stale_ollama_cli")
	}
	found := false
	for _, m := range h.Messages {
		if m == staleOllamaMessage {
			found = true
		}
	}
	if !found {
		t.Fatalf("messages=%v", h.Messages)
	}
}

func TestBuildAwakeningHealth_BusyWhenChatInFlight(t *testing.T) {
	if !TryAcquireGroundedChat() {
		t.Fatal("acquire")
	}
	t.Cleanup(ReleaseGroundedChat)

	srv := startMockOllamaPS(t, []map[string]any{
		{"name": "phi3:mini", "size_vram": 0},
	})
	t.Setenv("LLM_BASE_URL", srv.URL+"/v1")
	field := BuildFieldAssistantHealth(t.Context(), nil, 5, 2)
	h := BuildAwakeningHealth(t.Context(), AwakeningBuildInput{
		AIEnabled:         true,
		Field:             field,
		Mode:              WarmupModeFarmCounsel,
		FieldGuideChunks:  5,
		PlatformDocChunks: 2,
		EnvDefault:        "phi3:mini",
	})
	if h.State != AwakeningStateBusy {
		t.Fatalf("state=%q", h.State)
	}
}

func TestMaybeUnloadEmbedForChat_unloadsEmbedWhenChatCold(t *testing.T) {
	var unloadCalls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/ps":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"models": []map[string]any{
					{"name": "gte-embed", "size_vram": 0},
				},
			})
		case "/api/generate":
			unloadCalls.Add(1)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"done":true}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	t.Setenv("EMBEDDING_MODEL", "gte-embed")

	MaybeUnloadEmbedForChat(t.Context(), srv.URL+"/v1", "gte-embed", "phi3:mini")
	if unloadCalls.Load() != 1 {
		t.Fatalf("unload calls=%d", unloadCalls.Load())
	}
}

func TestMaybeUnloadEmbedForChat_skipsWhenChatLoaded(t *testing.T) {
	var unloadCalls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/ps":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"models": []map[string]any{
					{"name": "gte-embed", "size_vram": 8192},
					{"name": "phi3:mini", "size_vram": 8192},
				},
			})
		case "/api/generate":
			unloadCalls.Add(1)
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	MaybeUnloadEmbedForChat(t.Context(), srv.URL+"/v1", "gte-embed", "phi3:mini")
	if unloadCalls.Load() != 0 {
		t.Fatalf("expected skip, unload calls=%d", unloadCalls.Load())
	}
}
