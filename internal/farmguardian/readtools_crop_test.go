package farmguardian

import (
	"strings"
	"testing"

	"gr33n-api/internal/ai"
)

func TestShouldRunLookupCropTargetsReadIntent(t *testing.T) {
	if !shouldRunLookupCropTargetsReadIntent("what is my EC target for flower", nil) {
		t.Fatal("expected EC target intent")
	}
	if !shouldRunLookupCropTargetsReadIntent("", &ContextRef{Type: "zone", ID: 2}) {
		t.Fatal("expected zone context to trigger")
	}
	if shouldRunLookupCropTargetsReadIntent("hello", nil) {
		t.Fatal("expected no match for generic greeting")
	}
}

func TestReadToolIDs_IncludesLookupCropTargets(t *testing.T) {
	for _, id := range ReadToolIDs() {
		if id == "lookup_crop_targets" {
			return
		}
	}
	t.Fatal("lookup_crop_targets missing from ReadToolIDs")
}

func TestPlatformContextBlock_IncludesCropTargetsRule(t *testing.T) {
	block := PlatformContextBlock(ai.Config{Enabled: true}, true, ReadToolIDs())
	if !strings.Contains(block, "NEVER state an EC") || !strings.Contains(block, "lookup_crop_targets") {
		t.Fatalf("platform context missing crop rule")
	}
}
