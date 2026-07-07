package farmguardian

import "testing"

func TestResolveWarmupModel_requestModelOverridesEnv(t *testing.T) {
	t.Parallel()
	cache := NewModelCache()
	cache.Set([]ModelInfo{
		{Name: "tinyllama:latest", ContextWindow: 2048, Capabilities: []string{"completion"}},
		{Name: "phi3:mini", ContextWindow: 131072, EffectiveContextWindow: 4096, Capabilities: []string{"completion"}},
	}, "tinyllama:latest")

	model, grounded, reject := ResolveWarmupModel(cache, WarmupModeFarmCounsel, "", nil, nil, "tinyllama:latest")
	if reject == "" {
		t.Fatal("expected reject when env default is tinyllama for farm_counsel")
	}

	model, grounded, reject = ResolveWarmupModel(cache, WarmupModeFarmCounsel, "phi3:mini", nil, nil, "tinyllama:latest")
	if reject != "" {
		t.Fatalf("eval override should not reject: %q", reject)
	}
	if model != "phi3:mini" || !grounded {
		t.Fatalf("model=%q grounded=%v", model, grounded)
	}
}
