package farmguardian

import (
	"testing"
)

func TestBuildAwakeningHealth_UnavailableWhenAIDisabled(t *testing.T) {
	h := BuildAwakeningHealth(t.Context(), AwakeningBuildInput{AIEnabled: false})
	if h.State != AwakeningStateUnavailable {
		t.Fatalf("state=%q", h.State)
	}
	if h.Profile != AwakeningProfileLite {
		t.Fatalf("profile=%q", h.Profile)
	}
}

func TestBuildAwakeningHealth_SleepingWhenModelCold(t *testing.T) {
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
	if h.State != AwakeningStateSleeping {
		t.Fatalf("state=%q", h.State)
	}
	if !h.RagCorpusOK {
		t.Fatal("expected rag_corpus_ok")
	}
}

func TestNormalizeWarmupMode(t *testing.T) {
	if got := normalizeWarmupMode("quick"); got != WarmupModeQuick {
		t.Fatalf("got %q", got)
	}
	if got := normalizeWarmupMode(""); got != WarmupModeFarmCounsel {
		t.Fatalf("got %q", got)
	}
}

func TestPlanReadTools_MorningWalkthrough(t *testing.T) {
	plan := PlanReadTools("hello", &ContextRef{GuardianMode: "morning_walkthrough"}, Snapshot{})
	if !planContains(plan, "walk_farm") {
		t.Fatalf("plan=%v", plan.ToolIDs)
	}
	if !planContains(plan, "summarize_device_health") {
		t.Fatalf("plan=%v", plan.ToolIDs)
	}
}

func TestPlanReadTools_UnreadAlerts(t *testing.T) {
	plan := PlanReadTools("hi", nil, Snapshot{UnreadAlerts: 2})
	if !planContains(plan, "list_unread_alerts") {
		t.Fatalf("plan=%v", plan.ToolIDs)
	}
}
