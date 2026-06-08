package farmguardian

import (
	"math"
	"strings"
	"testing"

	"gr33n-api/internal/ai"
)

func TestCalcVPDKpa_KnownPair(t *testing.T) {
	got := CalcVPDKpa(25.0, 50.0)
	if math.Abs(got-1.585) > 0.01 {
		t.Fatalf("VPD(25°C, 50%% RH) = %v, want ~1.585", got)
	}
}

func TestCalcVPDKpa_SaturatedAir(t *testing.T) {
	got := CalcVPDKpa(22.0, 100.0)
	if math.Abs(got) > 0.001 {
		t.Fatalf("VPD at 100%% RH = %v, want ~0", got)
	}
}

func TestShouldRunGrowAdvisorReadIntent(t *testing.T) {
	if !shouldRunGrowAdvisorReadIntent("is my vpd on target", nil) {
		t.Fatal("expected VPD intent")
	}
	if !shouldRunGrowAdvisorReadIntent("", &ContextRef{Type: "zone", ID: 2, CropCycleID: 9}) {
		t.Fatal("expected zone+cycle context to trigger")
	}
	if !shouldRunGrowAdvisorReadIntent("how many days to flip", nil) {
		t.Fatal("expected flip intent")
	}
	if shouldRunGrowAdvisorReadIntent("hello", nil) {
		t.Fatal("expected no match for generic greeting")
	}
}

func TestReadToolIDs_IncludesGrowAdvisor(t *testing.T) {
	for _, id := range ReadToolIDs() {
		if id == "grow_advisor" {
			return
		}
	}
	t.Fatal("grow_advisor missing from ReadToolIDs")
}

func TestPlatformContextBlock_IncludesGrowAdvisorRule(t *testing.T) {
	block := PlatformContextBlock(ai.Config{Enabled: true}, true, ReadToolIDs())
	if !strings.Contains(block, "grow_advisor") || !strings.Contains(block, "flip") {
		t.Fatalf("platform context missing grow advisor rule: %s", block)
	}
}

func TestEstimateDLI(t *testing.T) {
	got := estimateDLI(500, 12)
	if math.Abs(got-21.6) > 0.1 {
		t.Fatalf("DLI estimate = %v, want ~21.6", got)
	}
}

func TestIsVegAndLateFlowerStages(t *testing.T) {
	if !isVegStage("early_veg") || isVegStage("early_flower") {
		t.Fatal("veg stage classification wrong")
	}
	if !isLateFlowerStage("late_flower") || isLateFlowerStage("mid_flower") {
		t.Fatal("late flower stage classification wrong")
	}
}
