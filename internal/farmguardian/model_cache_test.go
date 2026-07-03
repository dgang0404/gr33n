package farmguardian

import "testing"

func TestParseParameterCount(t *testing.T) {
	cases := []struct {
		in   string
		want int64
	}{
		{"7B", 7},
		{"70B", 70},
		{"3.2B", 3},
		{"", 0},
		{"unknown", 0},
	}
	for _, tc := range cases {
		if got := parseParameterCount(tc.in); got != tc.want {
			t.Errorf("parseParameterCount(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

func TestClassifySpeedClass(t *testing.T) {
	if got := classifySpeedClass("deepseek-r1:latest", 7); got != "reasoning" {
		t.Fatalf("want reasoning, got %q", got)
	}
	if got := classifySpeedClass("phi3:mini", 3); got != "fast" {
		t.Fatalf("want fast, got %q", got)
	}
}

func TestResolveChatModel(t *testing.T) {
	cache := NewModelCache()
	cache.Set([]ModelInfo{
		{Name: "llama3.1:8b", ContextWindow: 8192, Capabilities: []string{"completion"}},
		{Name: "phi3:mini", ContextWindow: 4096, Capabilities: []string{"completion"}},
	}, "llama3.1:8b")

	out := ResolveChatModel(cache, "llama3.1:8b", nil, "llama3.1:8b", true)
	if out.ModelName != "llama3.1:8b" || out.Fallback {
		t.Fatalf("direct hit: %+v", out)
	}

	out = ResolveChatModel(cache, "phi3:mini", nil, "llama3.1:8b", true)
	if out.RejectReason == "" {
		t.Fatal("expected grounded context reject")
	}

	out = ResolveChatModel(cache, "phi3:mini", nil, "llama3.1:8b", false)
	if out.ModelName != "phi3:mini" {
		t.Fatalf("non-grounded should allow small model: %+v", out)
	}

	farm := "not-in-ollama:latest"
	out = ResolveChatModel(cache, "", &farm, "llama3.1:8b", true)
	if !out.Fallback || out.ModelName != "llama3.1:8b" {
		t.Fatalf("missing farm model should fallback: %+v", out)
	}
}

func TestResolveChatModel_NameNormalization(t *testing.T) {
	cache := NewModelCache()
	cache.Set([]ModelInfo{
		{Name: "tinyllama:latest", ContextWindow: 2048, Capabilities: []string{"completion"}},
	}, "tinyllama")

	out := ResolveChatModel(cache, "", nil, "tinyllama", true)
	if out.RejectReason == "" {
		t.Fatal("grounded tinyllama via env default should reject on context window")
	}
	if out.ModelName != "" {
		t.Fatalf("reject should not set model name: %+v", out)
	}

	out = ResolveChatModel(cache, "", nil, "tinyllama", false)
	if out.ModelName != "tinyllama:latest" {
		t.Fatalf("ungrounded should resolve canonical name: %+v", out)
	}

	out = ResolveChatModel(cache, "tinyllama", nil, "tinyllama", true)
	if out.RejectReason == "" {
		t.Fatal("explicit bare tinyllama should hit same guardrail")
	}
}

func TestResolveChatModel_EnvDefaultGuardrailNotBypassed(t *testing.T) {
	cache := NewModelCache()
	cache.Set([]ModelInfo{
		{Name: "tinyllama:latest", ContextWindow: 2048, Capabilities: []string{"completion"}},
	}, "tinyllama")

	// Simulates session model == env default where first lookup used to bypass guardrail.
	out := ResolveChatModel(cache, "tinyllama", nil, "tinyllama", true)
	if out.RejectReason == "" {
		t.Fatal("env-default path must not bypass grounded context check")
	}
}

func TestModelCache_GetNormalization(t *testing.T) {
	cache := NewModelCache()
	cache.Set([]ModelInfo{
		{Name: "tinyllama:latest", ContextWindow: 2048, Capabilities: []string{"completion"}},
	}, "tinyllama")
	if !cache.Contains("tinyllama") {
		t.Fatal("Contains should match bare name")
	}
	info, ok := cache.Get("tinyllama")
	if !ok || info.Name != "tinyllama:latest" {
		t.Fatalf("Get: %+v ok=%v", info, ok)
	}
}

func TestModelCache_SnapshotAll(t *testing.T) {
	cache := NewModelCache()
	cache.Set([]ModelInfo{
		{Name: "llama3.2:latest", Capabilities: []string{"completion"}},
		{Name: "nomic-embed-text", Capabilities: []string{"embedding"}},
	}, "llama3.2:latest")

	chat, _ := cache.Snapshot(false)
	if len(chat) != 1 {
		t.Fatalf("chat snapshot len=%d", len(chat))
	}
	all, _ := cache.Snapshot(true)
	if len(all) != 2 {
		t.Fatalf("all snapshot len=%d", len(all))
	}
}

func TestOllamaNativeBase(t *testing.T) {
	if got := OllamaNativeBase("http://127.0.0.1:11434/v1"); got != "http://127.0.0.1:11434" {
		t.Fatalf("got %q", got)
	}
}
