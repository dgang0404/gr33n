package farmguardian

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func startMockOllamaPS(t *testing.T, models []map[string]any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/ps":
			_ = json.NewEncoder(w).Encode(map[string]any{"models": models})
		case "/v1/models", "/models":
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
}
