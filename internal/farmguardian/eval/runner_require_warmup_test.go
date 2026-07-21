package eval

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRunSuite_requireWarmupFailsWhenNotReady(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/guardian/warmup":
			w.WriteHeader(http.StatusAccepted)
		case "/v1/chat/health":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"awakening":{"state":"stirring"}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	client := &APIClient{BaseURL: srv.URL, FarmID: 1, HTTP: srv.Client()}
	fixtures := []Question{{ID: "smoke-nf-jlf-doc", Category: "natural_farming", Grounded: true, Prompt: "test"}}
	_, err := RunSuite(context.Background(), client, "phi3:mini", fixtures, RunSuiteOptions{
		WarmupGrounded: true,
		RequireWarmup:  true,
		WarmupTimeout:  200 * time.Millisecond,
	})
	if err == nil {
		t.Fatal("expected warmup failure")
	}
}
