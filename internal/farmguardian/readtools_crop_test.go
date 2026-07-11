package farmguardian

import (
	"strings"
	"testing"

	"gr33n-api/internal/ai"
	"gr33n-api/internal/croplibrary"
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
	if !shouldRunLookupCropTargetsReadIntent("how should I feed and light my cannabis and orchid", nil) {
		t.Fatal("expected feed/light compare intent")
	}
	if !shouldRunLookupCropTargetsReadIntent("apple tree watering in the greenhouse", nil) {
		t.Fatal("expected crop mention to trigger")
	}
}

func TestQuestionMentionsCrop(t *testing.T) {
	if !questionMentionsCrop("Compare cucumber vs tomato feed targets") {
		t.Fatal("expected cucumber and tomato mentions")
	}
	m, ok := defaultCropRegistry()
	if ok != nil {
		t.Fatal(ok)
	}
	mentions := m.FindMentions("aubergine EC")
	var foundEggplant bool
	for _, mention := range mentions {
		if mention.Key == "eggplant" {
			foundEggplant = true
		}
	}
	if !foundEggplant {
		t.Fatalf("expected eggplant from aubergine, got %+v", mentions)
	}
}

func TestFormatUnsupportedCropBlock_NoTargets(t *testing.T) {
	block := formatUnsupportedCropBlock(croplibrary.ResolvedMention{
		DisplayName: "ramps",
		Reason:      "Woodland spring ephemeral",
	})
	if strings.Contains(block, "photoperiod") && strings.Contains(block, "EC target:") {
		t.Fatalf("unsupported block must not include fake targets: %s", block)
	}
	if !strings.Contains(block, "not supported") || !strings.Contains(block, "Do not state EC") {
		t.Fatalf("unsupported block: %s", block)
	}
}

func TestSplitMentions(t *testing.T) {
	reg, err := defaultCropRegistry()
	if err != nil {
		t.Fatal(err)
	}
	mentions := reg.FindMentions("cucumber vs tomato and wild_leek")
	crops, unsup := splitMentions(mentions)
	if len(crops) < 2 {
		t.Fatalf("want 2 crops, got %d (%+v)", len(crops), crops)
	}
	if len(unsup) < 1 || unsup[0].Key != "ramps" {
		t.Fatalf("want ramps unsupported, got %+v", unsup)
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
