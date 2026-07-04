package farmguardian

import "testing"

func TestParseOriginalContextLength_phi3(t *testing.T) {
	info := map[string]any{
		"phi3.context_length":                        float64(131072),
		"phi3.rope.scaling.original_context_length": float64(4096),
	}
	if got := parseOriginalContextLength(info); got != 4096 {
		t.Fatalf("want 4096, got %d", got)
	}
}

func TestResolveEffectiveContextWindow_phi3Builtin(t *testing.T) {
	got := ResolveEffectiveContextWindow("phi3:mini", 131072, map[string]any{
		"phi3.context_length":                        float64(131072),
		"phi3.rope.scaling.original_context_length": float64(4096),
	})
	if got != 4096 {
		t.Fatalf("want builtin 4096, got %d", got)
	}
}

func TestResolveEffectiveContextWindow_ropeWithoutBuiltin(t *testing.T) {
	got := ResolveEffectiveContextWindow("llama3.2:latest", 131072, map[string]any{
		"llama.context_length":                        float64(131072),
		"llama.rope.scaling.original_context_length": float64(8192),
	})
	if got != 8192 {
		t.Fatalf("want 8192, got %d", got)
	}
}

func TestParseEffectiveContextOverrides(t *testing.T) {
	got := parseEffectiveContextOverrides("phi3:mini=8192, tinyllama=1024")
	if got["phi3:mini"] != 8192 || got["tinyllama"] != 1024 {
		t.Fatalf("got %+v", got)
	}
}

func TestComputePromptBudget_phi3Effective4096(t *testing.T) {
	budget, log := ComputePromptBudget(4096, 20)
	if budget.RAGTopK != 5 {
		t.Fatalf("RAGTopK want 5, got %d", budget.RAGTopK)
	}
	if budget.MaxHistoryTurns != 8 {
		t.Fatalf("MaxHistoryTurns want 8, got %d", budget.MaxHistoryTurns)
	}
	if len(log) == 0 {
		t.Fatal("expected trim log for effective 4096")
	}
}

func TestPromptBudgetContextWindow_prefersEffective(t *testing.T) {
	if got := PromptBudgetContextWindow(ModelInfo{ContextWindow: 131072, EffectiveContextWindow: 4096}); got != 4096 {
		t.Fatalf("got %d", got)
	}
}
