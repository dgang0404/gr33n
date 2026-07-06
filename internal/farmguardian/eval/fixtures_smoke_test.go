package eval

import "testing"

func TestSmokeFixtures_count(t *testing.T) {
	if len(SmokeFixtures()) != 4 {
		t.Fatalf("expected 4 smoke fixtures, got %d", len(SmokeFixtures()))
	}
}

func TestFixturesForSuite_smoke(t *testing.T) {
	if len(FixturesForSuite("smoke")) != 4 {
		t.Fatal("smoke suite size")
	}
}

func TestFixturesForSuite_regression(t *testing.T) {
	if len(FixturesForSuite("regression")) != len(Fixtures()) {
		t.Fatal("regression should match Fixtures()")
	}
}

func TestScore_smokeCherry(t *testing.T) {
	res := Score(ScoreInput{
		Question: Question{ID: "smoke-cherry-forest", Category: "ungrounded"},
		Answer:   "Your cherry tree understory can keep blackberries if you manage goldenrod for dyes separately.",
	})
	if !res.Passed {
		t.Fatalf("expected pass, got %+v", res)
	}
}
