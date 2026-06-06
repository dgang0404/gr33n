package farmguardian

import (
	"strings"
	"testing"

	"gr33n-api/internal/ai"
)

func TestSystemPrompt_OperationsVocabulary(t *testing.T) {
	p := SystemPrompt()
	for _, want := range []string{
		"Supplies",
		"Feeding (details)",
		"Money",
		"cannot change stock quantities",
		"input_batches",
	} {
		if !strings.Contains(p, want) {
			t.Fatalf("persona missing %q:\n%s", want, p)
		}
	}
}

func TestPlatformContextBlock_OperationsHub(t *testing.T) {
	block := PlatformContextBlock(ai.Config{Enabled: true}, true, nil)
	for _, want := range []string{
		"Supplies",
		"Feeding (details)",
		"Money",
		"over Inventory",
	} {
		if !strings.Contains(block, want) {
			t.Fatalf("platform block missing %q:\n%s", want, block)
		}
	}
}
