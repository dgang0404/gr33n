package eval

import "testing"

func TestSuiteNeedsWarmup_smokeNaturalFarming(t *testing.T) {
	if !SuiteNeedsWarmup("smoke-natural-farming") {
		t.Fatal("expected warmup for smoke-natural-farming")
	}
}

func TestSuiteRequiresWarmup_smokeNaturalFarming(t *testing.T) {
	if !SuiteRequiresWarmup("smoke-nf") {
		t.Fatal("expected required warmup for smoke-nf")
	}
}

func TestSuiteRequiresWarmup_regressionFalse(t *testing.T) {
	if SuiteRequiresWarmup("regression") {
		t.Fatal("regression should not require warmup")
	}
}

func TestSmokeFullFixtures_count(t *testing.T) {
	if got := len(SmokeFullFixtures()); got != 15 {
		t.Fatalf("expected 15 smoke-full fixtures, got %d", got)
	}
}

func TestFixturesForSuite_smokeFull(t *testing.T) {
	if len(FixturesForSuite("smoke-full")) != 15 {
		t.Fatal("smoke-full suite size")
	}
}
