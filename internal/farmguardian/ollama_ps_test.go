package farmguardian

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEnrichModelRuntimeHints(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/ps":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"models": []map[string]any{
					{"name": "tinyllama:latest", "size_vram": 0},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	models := []ModelInfo{
		{Name: "tinyllama:latest"},
		{Name: "phi3:mini"},
	}
	out := EnrichModelRuntimeHints(context.Background(), srv.URL+"/v1", models, srv.Client())
	if !out[0].Loaded || out[0].Processor != "cpu" {
		t.Fatalf("loaded cpu model: %+v", out[0])
	}
	if out[0].RuntimeHint == "" {
		t.Fatal("expected runtime hint on loaded model")
	}
	if out[1].Loaded {
		t.Fatal("phi3:mini should be cold")
	}
	if out[1].RuntimeHint == "" {
		t.Fatal("expected cold hint")
	}
}

func TestFilterChatModels(t *testing.T) {
	all := []ModelInfo{
		{Name: "llama3.2:latest", Capabilities: []string{"completion"}},
		{Name: "nomic-embed-text", Capabilities: []string{"embedding"}},
	}
	chat := filterChatModels(all)
	if len(chat) != 1 || chat[0].Name != "llama3.2:latest" {
		t.Fatalf("got %+v", chat)
	}
}
