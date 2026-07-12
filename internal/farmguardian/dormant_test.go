package farmguardian

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func startMockOllamaGenerate(t *testing.T, onGenerate func(body map[string]any)) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/generate":
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			if onGenerate != nil {
				onGenerate(body)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"done":true}`))
		case "/api/ps":
			_ = json.NewEncoder(w).Encode(map[string]any{"models": []any{}})
		default:
			http.NotFound(w, r)
		}
	}))
}

func TestRequestDormant_unloadsChatModelAndSetsFlag(t *testing.T) {
	t.Cleanup(ClearDormantFlag)
	var gotKeepAlive any
	srv := startMockOllamaGenerate(t, func(body map[string]any) {
		gotKeepAlive = body["keep_alive"]
	})
	defer srv.Close()

	if err := RequestDormant(t.Context(), srv.URL+"/v1", "phi3:mini", "", false); err != nil {
		t.Fatal(err)
	}
	if gotKeepAlive != float64(0) {
		t.Fatalf("keep_alive=%v want 0", gotKeepAlive)
	}
	requested, _, at := snapshotDormantState()
	if !requested {
		t.Fatal("expected dormantRequested=true")
	}
	if at.IsZero() {
		t.Fatal("expected dormantAt set")
	}
}

func TestRequestDormant_alsoUnloadsVisionModel(t *testing.T) {
	t.Cleanup(ClearDormantFlag)
	seen := map[string]bool{}
	srv := startMockOllamaGenerate(t, func(body map[string]any) {
		if name, ok := body["model"].(string); ok {
			seen[name] = true
		}
	})
	defer srv.Close()

	if err := RequestDormant(t.Context(), srv.URL+"/v1", "phi3:mini", "llava:latest", false); err != nil {
		t.Fatal(err)
	}
	if !seen["phi3:mini"] || !seen["llava:latest"] {
		t.Fatalf("expected both models unloaded, got %v", seen)
	}
}

func TestRequestDormant_emptyChatModelErrors(t *testing.T) {
	if err := RequestDormant(t.Context(), "http://127.0.0.1:11434/v1", "", "", false); err == nil {
		t.Fatal("expected error for empty chat model")
	}
}

func TestClearDormantFlag_resetsState(t *testing.T) {
	t.Cleanup(ClearDormantFlag)
	srv := startMockOllamaGenerate(t, nil)
	defer srv.Close()

	if err := RequestDormant(t.Context(), srv.URL+"/v1", "phi3:mini", "", false); err != nil {
		t.Fatal(err)
	}
	ClearDormantFlag()
	requested, _, _ := snapshotDormantState()
	if requested {
		t.Fatal("expected dormantRequested=false after ClearDormantFlag")
	}
}

func TestBuildAwakeningHealth_DormantWhenRequested(t *testing.T) {
	t.Cleanup(ClearDormantFlag)
	srv := startMockOllamaPS(t, nil)
	t.Setenv("LLM_BASE_URL", srv.URL+"/v1")

	genSrv := startMockOllamaGenerate(t, nil)
	defer genSrv.Close()
	if err := RequestDormant(t.Context(), genSrv.URL+"/v1", "phi3:mini", "", false); err != nil {
		t.Fatal(err)
	}

	field := BuildFieldAssistantHealth(t.Context(), nil, 1, 1)
	h := BuildAwakeningHealth(t.Context(), AwakeningBuildInput{
		AIEnabled:  true,
		Field:      field,
		Mode:       WarmupModeFarmCounsel,
		EnvDefault: "phi3:mini",
	})
	if h.State != AwakeningStateDormant {
		t.Fatalf("state=%q want dormant", h.State)
	}
}

func TestStartWarmup_clearsDormantFlag(t *testing.T) {
	t.Cleanup(ClearDormantFlag)
	genSrv := startMockOllamaGenerate(t, nil)
	defer genSrv.Close()
	if err := RequestDormant(t.Context(), genSrv.URL+"/v1", "phi3:mini", "", false); err != nil {
		t.Fatal(err)
	}
	requested, _, _ := snapshotDormantState()
	if !requested {
		t.Fatal("setup: expected dormant flag set before warmup")
	}

	srv := startMockOllamaPS(t, nil)
	StartWarmup(t.Context(), srv.URL+"/v1", "farm_counsel", "phi3:mini", nil, nil, "phi3:mini", nil, false)

	requested, _, _ = snapshotDormantState()
	if requested {
		t.Fatal("expected StartWarmup to clear dormant flag")
	}
}
