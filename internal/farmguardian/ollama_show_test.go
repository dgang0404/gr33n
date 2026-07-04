package farmguardian

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseContextLength(t *testing.T) {
	info := map[string]any{
		"llama.context_length":  float64(8192),
		"general.architecture":  "llama",
		"gemma.context_length":  float64(4096),
	}
	if got := parseContextLength(info); got != 8192 {
		t.Fatalf("want 8192, got %d", got)
	}
	if got := parseContextLength(nil); got != 0 {
		t.Fatalf("want 0, got %d", got)
	}
}

func TestEnrichModelContextWindows(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/show" {
			http.NotFound(w, r)
			return
		}
		var req struct {
			Name string `json:"name"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"model_info": map[string]any{
				"llama.context_length": float64(8192),
			},
			"capabilities": []string{"completion"},
		})
	}))
	defer srv.Close()

	models := []ModelInfo{{Name: "llama3.2:latest"}}
	enriched := EnrichModelContextWindows(context.Background(), srv.URL+"/v1", models, srv.Client(), 2)
	if len(enriched) != 1 || enriched[0].ContextWindow != 8192 {
		t.Fatalf("got %+v", enriched)
	}
	if len(enriched[0].Capabilities) != 1 || enriched[0].Capabilities[0] != "completion" {
		t.Fatalf("capabilities: %+v", enriched[0].Capabilities)
	}
}

func TestEnrichModelContextWindows_phi3Effective(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/show" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"model_info": map[string]any{
				"phi3.context_length":                        float64(131072),
				"phi3.rope.scaling.original_context_length": float64(4096),
			},
			"capabilities": []string{"completion"},
		})
	}))
	defer srv.Close()

	models := []ModelInfo{{Name: "phi3:mini"}}
	enriched := EnrichModelContextWindows(context.Background(), srv.URL+"/v1", models, srv.Client(), 2)
	if len(enriched) != 1 {
		t.Fatalf("got %+v", enriched)
	}
	if enriched[0].ContextWindow != 131072 {
		t.Fatalf("advertised want 131072, got %d", enriched[0].ContextWindow)
	}
	if enriched[0].EffectiveContextWindow != 4096 {
		t.Fatalf("effective want 4096, got %d", enriched[0].EffectiveContextWindow)
	}
}

func TestPullOllamaModel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/pull" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	}))
	defer srv.Close()

	if err := PullOllamaModel(context.Background(), srv.URL, "tinyllama", srv.Client()); err != nil {
		t.Fatal(err)
	}
}
