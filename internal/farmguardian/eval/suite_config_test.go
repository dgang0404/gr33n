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
	if got := len(SmokeFullFixtures()); got != 16 {
		t.Fatalf("expected 16 smoke-full fixtures, got %d", got)
	}
}

func TestFixturesForSuite_smokeFull(t *testing.T) {
	if len(FixturesForSuite("smoke-full")) != 16 {
		t.Fatal("smoke-full suite size")
	}
}

func TestSmokeAllFixtures_count(t *testing.T) {
	if got := len(SmokeAllFixtures()); got != 24 {
		t.Fatalf("expected 24 smoke-all fixtures, got %d", got)
	}
}

func TestFixturesForSuite_smokeAll(t *testing.T) {
	if len(FixturesForSuite("smoke-all")) != 24 {
		t.Fatal("smoke-all suite size")
	}
}
