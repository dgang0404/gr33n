package farmguardian

import "testing"

func TestNormalizeModelName(t *testing.T) {
	cases := map[string]string{
		"tinyllama:latest": "tinyllama",
		"tinyllama":        "tinyllama",
		"phi3:mini":        "phi3:mini",
		"llama3.2:latest": "llama3.2",
	}
	for in, want := range cases {
		if got := NormalizeModelName(in); got != want {
			t.Fatalf("NormalizeModelName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestModelLookupKeys(t *testing.T) {
	keys := modelLookupKeys("tinyllama")
	if len(keys) < 2 {
		t.Fatalf("expected alias keys, got %v", keys)
	}
	seen := map[string]bool{}
	for _, k := range keys {
		seen[k] = true
	}
	if !seen["tinyllama"] || !seen["tinyllama:latest"] {
		t.Fatalf("missing aliases: %v", keys)
	}
}
