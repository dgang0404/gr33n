package mixplan_test

import (
	"errors"
	"math"
	"testing"

	"gr33n-api/internal/fertigation/mixplan"
)

// vegInput returns a representative "Veg program" Input mirroring the
// acceptance criterion in the Phase 39 WS2 plan:
// base 0.2 mS/cm, target 1.6 mS/cm, 95 L batch at 1:500 dilution.
func vegInput() mixplan.Input {
	return mixplan.Input{
		ReservoirID:       3,
		WaterVolumeLiters: 95,
		BaseEcMscm:        0.2,
		TargetEcMscm:      1.6,
		DilutionRatioStr:  "1:500",
		Components: []mixplan.ComponentInput{
			{InputDefinitionID: 12, InputName: "JLF", PartValue: 1},
			{InputDefinitionID: 14, InputName: "JLW", PartValue: 1},
		},
		PumpFlowRateMlPerSec: 10.0,
	}
}

// TestCalculateAcceptanceCriteria is the plan's stated acceptance criterion.
func TestCalculateAcceptanceCriteria(t *testing.T) {
	plan, err := mixplan.Calculate(vegInput())
	if err != nil {
		t.Fatalf("Calculate: %v", err)
	}
	if len(plan.Steps) == 0 {
		t.Fatal("expected non-empty steps")
	}
	for i, step := range plan.Steps {
		if step.RunSeconds <= 0 {
			t.Fatalf("step %d: RunSeconds=%d, want >0", i+1, step.RunSeconds)
		}
	}
}

// TestCalculateVolumeDistribution verifies the ml math for a 1:500 batch.
//
//	95 L × (1/500) × 1000 ml/L = 190 ml total concentrate
//	equal parts (JLF:JLW = 1:1) → 95 ml each
//	at 10 ml/s → 10 s each (ceil(95/10) = 10)
func TestCalculateVolumeDistribution(t *testing.T) {
	plan, err := mixplan.Calculate(vegInput())
	if err != nil {
		t.Fatal(err)
	}
	const wantTotalML = 190.0
	total := 0.0
	for _, s := range plan.Steps {
		total += s.VolumeMl
	}
	if math.Abs(total-wantTotalML) > 1.0 {
		t.Fatalf("total concentrate ml = %.1f, want %.1f", total, wantTotalML)
	}
	for _, s := range plan.Steps {
		if math.Abs(s.VolumeMl-95.0) > 1.0 {
			t.Fatalf("step %d VolumeMl = %.1f, want ~95", s.Step, s.VolumeMl)
		}
		if s.RunSeconds != 10 {
			t.Fatalf("step %d RunSeconds = %d, want 10", s.Step, s.RunSeconds)
		}
	}
}

// TestCalculateChannelAssignment verifies channel indices are 1-indexed and sequential.
func TestCalculateChannelAssignment(t *testing.T) {
	plan, err := mixplan.Calculate(vegInput())
	if err != nil {
		t.Fatal(err)
	}
	for i, s := range plan.Steps {
		if s.ChannelIndex != i+1 {
			t.Fatalf("step %d: channel=%d, want %d", i+1, s.ChannelIndex, i+1)
		}
	}
}

// TestCalculateUnequalParts verifies proportional distribution when part_values differ.
// JLF:JLW = 2:1 → JLF gets 2/3 of concentrate, JLW gets 1/3.
func TestCalculateUnequalParts(t *testing.T) {
	in := vegInput()
	in.Components[0].PartValue = 2 // JLF double dose
	in.Components[1].PartValue = 1

	plan, err := mixplan.Calculate(in)
	if err != nil {
		t.Fatal(err)
	}
	jlfML := plan.Steps[0].VolumeMl
	jlwML := plan.Steps[1].VolumeMl
	if math.Abs(jlfML/jlwML-2.0) > 0.05 {
		t.Fatalf("expected JLF:JLW ratio ~2:1, got %.1f:%.1f", jlfML, jlwML)
	}
}

// TestCalculateNoDilutionRatioError verifies ErrNoDilutionRatio is returned
// when the dilution ratio is missing or unparseable.
func TestCalculateNoDilutionRatioError(t *testing.T) {
	in := vegInput()
	in.DilutionRatioStr = ""
	_, err := mixplan.Calculate(in)
	if !errors.Is(err, mixplan.ErrNoDilutionRatio) {
		t.Fatalf("expected ErrNoDilutionRatio, got %v", err)
	}

	in.DilutionRatioStr = "not-a-ratio"
	_, err = mixplan.Calculate(in)
	if !errors.Is(err, mixplan.ErrNoDilutionRatio) {
		t.Fatalf("expected ErrNoDilutionRatio for garbage input, got %v", err)
	}
}

// TestCalculateNoComponentsError verifies ErrNoComponents is returned for an
// empty component list.
func TestCalculateNoComponentsError(t *testing.T) {
	in := vegInput()
	in.Components = nil
	_, err := mixplan.Calculate(in)
	if !errors.Is(err, mixplan.ErrNoComponents) {
		t.Fatalf("expected ErrNoComponents, got %v", err)
	}
}

// TestCalculateZeroBaseECWarning verifies that a zero base EC emits a warning
// rather than an error (caller decides whether to block or allow).
func TestCalculateZeroBaseECWarning(t *testing.T) {
	in := vegInput()
	in.BaseEcMscm = 0
	plan, err := mixplan.Calculate(in)
	if err != nil {
		t.Fatalf("unexpected error for zero base EC: %v", err)
	}
	if len(plan.Warnings) == 0 {
		t.Fatal("expected at least one warning for zero base EC")
	}
	found := false
	for _, w := range plan.Warnings {
		if len(w) > 10 {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("no meaningful warnings: %v", plan.Warnings)
	}
}

// TestCalculateSlashDilutionRatio verifies "1/500" is parsed the same as "1:500".
func TestCalculateSlashDilutionRatio(t *testing.T) {
	in := vegInput()
	in.DilutionRatioStr = "1/500"
	plan, err := mixplan.Calculate(in)
	if err != nil {
		t.Fatalf("Calculate: %v", err)
	}
	if len(plan.Steps) == 0 {
		t.Fatal("expected non-empty steps")
	}
}

// TestCalculateSingleComponent verifies a single-component recipe works.
func TestCalculateSingleComponent(t *testing.T) {
	in := mixplan.Input{
		ReservoirID:          1,
		WaterVolumeLiters:    100,
		BaseEcMscm:           0.3,
		TargetEcMscm:         1.8,
		DilutionRatioStr:     "1:1000",
		Components:           []mixplan.ComponentInput{{InputDefinitionID: 5, InputName: "IMO-3", PartValue: 0}},
		PumpFlowRateMlPerSec: mixplan.DefaultPumpFlowRateMLPerSec,
	}
	plan, err := mixplan.Calculate(in)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Steps) != 1 {
		t.Fatalf("want 1 step, got %d", len(plan.Steps))
	}
	// 100 L × 1/1000 × 1000 ml/L = 100 ml → ceil(100/10) = 10 s
	if plan.Steps[0].RunSeconds != 10 {
		t.Fatalf("RunSeconds = %d, want 10", plan.Steps[0].RunSeconds)
	}
}

// TestCalculateDefaultFlowRate verifies the default 10 ml/s is used when
// PumpFlowRateMlPerSec is zero.
func TestCalculateDefaultFlowRate(t *testing.T) {
	in := vegInput()
	in.PumpFlowRateMlPerSec = 0 // force default

	planDefault, err := mixplan.Calculate(in)
	if err != nil {
		t.Fatal(err)
	}
	in.PumpFlowRateMlPerSec = mixplan.DefaultPumpFlowRateMLPerSec
	planExplicit, err := mixplan.Calculate(in)
	if err != nil {
		t.Fatal(err)
	}
	for i := range planDefault.Steps {
		if planDefault.Steps[i].RunSeconds != planExplicit.Steps[i].RunSeconds {
			t.Fatalf("step %d: default RunSeconds=%d != explicit RunSeconds=%d",
				i+1, planDefault.Steps[i].RunSeconds, planExplicit.Steps[i].RunSeconds)
		}
	}
}
