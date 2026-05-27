package farmguardian

import (
	"strings"
	"testing"
)

func TestVisionContextBlock_HypothesisDisclaimer(t *testing.T) {
	got := VisionContextBlock()
	for _, want := range []string{"hypotheses", "create_task", "Confirm"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in block", want)
		}
	}
}
