package farmguardian

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAutoDormantMinutesFromEnv_disabledByDefault(t *testing.T) {
	t.Setenv("GUARDIAN_AUTO_DORMANT_MINUTES", "")
	if got := AutoDormantMinutesFromEnv(); got != 0 {
		t.Fatalf("got %v want 0", got)
	}
}

func TestAutoDormantMinutesFromEnv_parsesMinutes(t *testing.T) {
	t.Setenv("GUARDIAN_AUTO_DORMANT_MINUTES", "45")
	if got := AutoDormantMinutesFromEnv(); got != 45*time.Minute {
		t.Fatalf("got %v", got)
	}
}

func TestNoteGuardianActivity_clearsDormant(t *testing.T) {
	t.Cleanup(ClearDormantFlag)
	srv := startMockOllamaGenerate(t, nil)
	defer srv.Close()
	if err := RequestDormant(t.Context(), srv.URL+"/v1", "phi3:mini", "", false); err != nil {
		t.Fatal(err)
	}
	NoteGuardianActivity("phi3:mini")
	requested, _, _ := snapshotDormantState()
	if requested {
		t.Fatal("expected dormant cleared after activity")
	}
}

func TestMaybeAutoDormant_skipsWhenDisabled(t *testing.T) {
	t.Cleanup(ClearDormantFlag)
	t.Setenv("GUARDIAN_AUTO_DORMANT_MINUTES", "")
	NoteGuardianActivity("phi3:mini")
	ok, err := MaybeAutoDormant(t.Context())
	if err != nil || ok {
		t.Fatalf("got ok=%v err=%v", ok, err)
	}
}

func TestMaybeAutoDormant_firesAfterIdle(t *testing.T) {
	t.Cleanup(func() {
		ClearDormantFlag()
		activityMu.Lock()
		lastActivityAt = time.Time{}
		lastActivityChatModel = ""
		activityMu.Unlock()
	})
	t.Setenv("GUARDIAN_AUTO_DORMANT_MINUTES", "30")

	var unloadModel string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/generate":
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			if m, ok := body["model"].(string); ok {
				unloadModel = m
			}
			w.WriteHeader(http.StatusOK)
		case "/api/ps":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"models": []map[string]any{{"name": "phi3:mini", "size_vram": 0}},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	t.Setenv("LLM_BASE_URL", srv.URL+"/v1")

	activityMu.Lock()
	lastActivityAt = time.Now().Add(-31 * time.Minute)
	lastActivityChatModel = "phi3:mini"
	activityMu.Unlock()

	ok, err := MaybeAutoDormant(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected auto-dormant to fire")
	}
	if unloadModel != "phi3:mini" {
		t.Fatalf("unload model=%q", unloadModel)
	}
	requested, auto, _ := snapshotDormantState()
	if !requested || !auto {
		t.Fatalf("requested=%v auto=%v", requested, auto)
	}
}

func TestAutoDormantIdleRemaining_countsDown(t *testing.T) {
	t.Cleanup(ClearDormantFlag)
	t.Setenv("GUARDIAN_AUTO_DORMANT_MINUTES", "60")
	NoteGuardianActivity("phi3:mini")
	enabled, remaining := AutoDormantIdleRemaining()
	if !enabled || remaining <= 0 || remaining > 60*time.Minute {
		t.Fatalf("enabled=%v remaining=%v", enabled, remaining)
	}
}
