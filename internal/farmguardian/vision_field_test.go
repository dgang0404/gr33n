package farmguardian

import (
	"strings"
	"testing"
)

func TestVisionContextBlock_Phase67CropGrounding(t *testing.T) {
	block := VisionContextBlock()
	for _, frag := range []string{"crop profile", "hypotheses", "Phase 67"} {
		if !strings.Contains(block, frag) {
			t.Fatalf("vision block missing %q", frag)
		}
	}
}
