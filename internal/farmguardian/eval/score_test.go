package eval

import "testing"

func TestScore_citationQuestion(t *testing.T) {
	res := Score(ScoreInput{
		Question: Question{ID: "x", Category: "field_guide", ExpectCitation: true},
		Answer:   "See [1] for EC targets.",
		CitationCount: 1,
	})
	if !res.Passed {
		t.Fatalf("expected pass, got %+v", res)
	}
}

func TestScore_declineQuestion(t *testing.T) {
	res := Score(ScoreInput{
		Question: Question{ID: "x", Category: "out_of_scope", ExpectDecline: true},
		Answer:   "That's outside farm operations — check the Dashboard for gr33n features.",
	})
	if !res.Passed {
		t.Fatalf("expected decline pass, got %+v", res)
	}
}

func TestAggregate_rates(t *testing.T) {
	scores := []ScoreResult{
		{Category: "field_guide", Passed: true},
		{Category: "field_guide", Passed: false},
		{Category: "out_of_scope", Passed: true},
		{Category: "write_intent", Passed: true, LatencyMs: 100},
	}
	cite, dec, prop, lat, _ := Aggregate(scores)
	if cite != 0.5 || dec != 1.0 || prop != 1.0 || lat != 25 {
		t.Fatalf("rates cite=%v dec=%v prop=%v lat=%v", cite, dec, prop, lat)
	}
}

func TestFixtures_count(t *testing.T) {
	if len(Fixtures()) < 18 {
		t.Fatalf("expected ~20 fixtures, got %d", len(Fixtures()))
	}
}
