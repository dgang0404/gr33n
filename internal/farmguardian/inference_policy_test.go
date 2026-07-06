package farmguardian

import (
	"testing"

	db "gr33n-api/internal/db"
)

func TestFarmCounselModel_fallbackPreferred(t *testing.T) {
	pref := "phi3:mini"
	f := db.Gr33ncoreFarm{GuardianPreferredModel: &pref}
	got := FarmCounselModel(&f)
	if got == nil || *got != "phi3:mini" {
		t.Fatalf("got %v", got)
	}
}

func TestFarmCounselModel_prefersCounselColumn(t *testing.T) {
	pref := "tinyllama"
	counsel := "phi3:mini"
	f := db.Gr33ncoreFarm{GuardianPreferredModel: &pref, GuardianCounselModel: &counsel}
	got := FarmCounselModel(&f)
	if got == nil || *got != "phi3:mini" {
		t.Fatalf("got %v", got)
	}
}

func TestInferenceHostsSplit(t *testing.T) {
	t.Setenv("LLM_BASE_URL", "http://chat:11434/v1")
	t.Setenv("EMBEDDING_BASE_URL", "http://embed:11434/v1")
	if !InferenceHostsSplit() {
		t.Fatal("expected split hosts")
	}
	t.Setenv("EMBEDDING_BASE_URL", "http://chat:11434/v1")
	if InferenceHostsSplit() {
		t.Fatal("expected single host")
	}
}

func TestResolveChatModel_counselVsQuickFarmPolicy(t *testing.T) {
	cache := NewModelCache()
	cache.Set([]ModelInfo{
		{Name: "phi3:mini", ContextWindow: 8192},
		{Name: "tinyllama", ContextWindow: 2048},
	}, "tinyllama")
	counsel := "phi3:mini"
	quick := "tinyllama"
	outCounsel := ResolveChatModel(cache, "", &counsel, "tinyllama", true)
	if outCounsel.ModelName != "phi3:mini" {
		t.Fatalf("counsel=%q", outCounsel.ModelName)
	}
	outQuick := ResolveChatModel(cache, "", &quick, "phi3:mini", false)
	if outQuick.ModelName != "tinyllama" {
		t.Fatalf("quick=%q", outQuick.ModelName)
	}
}
