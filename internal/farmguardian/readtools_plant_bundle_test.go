package farmguardian

import (
	"strings"
	"testing"
)

func TestShouldRunPlantContextBundleIntent_CropCycleRef(t *testing.T) {
	if !ShouldRunPlantContextBundleIntent("hello", &ContextRef{CropCycleID: 9}) {
		t.Fatal("crop_cycle_id ref should always trigger bundle")
	}
}

func TestShouldRunPlantContextBundleIntent_VegGrowQuestion(t *testing.T) {
	if !ShouldRunPlantContextBundleIntent("What stage is my veg grow?", nil) {
		t.Fatal("expected veg grow question to trigger bundle")
	}
}

func TestShouldRunPlantContextBundleIntent_QuickChatNoFarm(t *testing.T) {
	if ShouldRunPlantContextBundleIntent("", nil) {
		t.Fatal("empty question without ref should not trigger")
	}
}

func TestShouldRunPlantContextBundleIntent_ZoneRef(t *testing.T) {
	if !ShouldRunPlantContextBundleIntent("How is this grow doing?", &ContextRef{Type: "zone", ID: 2, CropCycleID: 5}) {
		t.Fatal("zone+cycle ref with grow question should trigger")
	}
}

func TestTrimPlantContextBundle_DropsLowPriority(t *testing.T) {
	sections := []plantBundleSection{
		{priority: 0, text: "header"},
		{priority: 1, text: stringsRepeat("targets ", 200)},
		{priority: 5, text: stringsRepeat("lighting ", 200)},
	}
	out := trimPlantContextBundle(sections, 400)
	if out == "" {
		t.Fatal("expected trimmed output")
	}
	if strings.Contains(out, "lighting") {
		t.Fatalf("expected lighting section dropped first, got len=%d", len(out))
	}
	if !strings.Contains(out, "header") {
		t.Fatal("header should remain")
	}
}

func stringsRepeat(s string, n int) string {
	var b string
	for i := 0; i < n; i++ {
		b += s
	}
	return b
}

func TestLightingOffAt(t *testing.T) {
	if got := lightingOffAt("06:00", 18); got != "00:00" {
		t.Fatalf("off at = %q want 00:00", got)
	}
}

func TestBundleCoversReadTool(t *testing.T) {
	if !bundleCoversReadTool(true, "lookup_crop_targets") {
		t.Fatal("bundle should cover lookup_crop_targets")
	}
	if bundleCoversReadTool(true, "walk_farm") {
		t.Fatal("walk_farm should not be covered")
	}
}

func TestPlanReadTools_IncludesPlantBundle(t *testing.T) {
	plan := PlanReadTools("What stage is my veg grow?", &ContextRef{CropCycleID: 3}, Snapshot{})
	if !planContains(plan, "plant_context_bundle") {
		t.Fatalf("plan missing plant_context_bundle: %+v", plan.ToolIDs)
	}
}
