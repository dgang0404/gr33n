package chat

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestProbeCacheTTL(t *testing.T) {
	if probeCacheTTL(true) != reachableCacheTTL {
		t.Fatalf("reachable ttl")
	}
	if probeCacheTTL(false) != unreachableCacheTTL {
		t.Fatalf("unreachable ttl")
	}
}

func TestProbeLLMReachable_shortNegativeCacheAllowsRetry(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/ps" {
			http.NotFound(w, r)
			return
		}
		attempts++
		if attempts == 1 {
			http.Error(w, "busy", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer srv.Close()

	// Use /v1 path so /api/ps is not routed to the models handler.
	t.Setenv("LLM_BASE_URL", srv.URL+"/v1")
	t.Setenv("LLM_MODEL", "test")
	ResetLLMReachabilityCache()

	if probeLLMReachable(context.Background()) {
		t.Fatal("expected first probe false")
	}
	time.Sleep(unreachableCacheTTL + 50*time.Millisecond)
	if !probeLLMReachable(context.Background()) {
		t.Fatal("expected second probe true after short negative cache")
	}
	if attempts < 2 {
		t.Fatalf("expected 2 probe attempts, got %d", attempts)
	}
}

func TestProbeLLMReachable_ollamaBusyFallback(t *testing.T) {
	combined := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/models", "/models":
			http.Error(w, "busy", http.StatusServiceUnavailable)
		case "/api/ps":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"models":[{"name":"phi3:mini"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer combined.Close()

	t.Setenv("LLM_BASE_URL", combined.URL+"/v1")
	t.Setenv("LLM_MODEL", "phi3:mini")
	ResetLLMReachabilityCache()

	if !probeLLMReachable(context.Background()) {
		t.Fatal("expected busy-alive fallback to mark reachable")
	}
}

func TestLlmProbeTimeout_localDefault(t *testing.T) {
	if got := llmProbeTimeout("http://127.0.0.1:11434/v1"); got != defaultLocalProbe {
		t.Fatalf("local default = %v want %v", got, defaultLocalProbe)
	}
	if got := llmProbeTimeout("https://api.openai.com/v1"); got != defaultRemoteProbe {
		t.Fatalf("remote default = %v want %v", got, defaultRemoteProbe)
	}
}
