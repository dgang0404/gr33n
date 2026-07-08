package eval

import (
	"testing"

	"gr33n-api/internal/farmguardian"
)

func TestScore_fieldGuideRegressionDriftFails(t *testing.T) {
	t.Parallel()
	res := Score(ScoreInput{
		Question: Question{
			ID:             "fg-citation-format",
			Category:       "field_guide",
			Prompt:         "What EC range does the platform recommend for hydro lettuce?",
			ExpectCitation: true,
		},
		Answer:        `Lettuce EC 1.0–1.3 mS/cm. Endocrine disruptors in Lake Erie affect wildlife.`,
		CitationCount: 2,
	})
	if res.Passed {
		t.Fatalf("field_guide regression drift should fail: %+v", res)
	}
}

func TestScore_phase127FertTriageDriftFails(t *testing.T) {
	t.Parallel()
	res := Score(ScoreInput{
		Question: Question{
			ID:             "p128-fert-triage",
			Category:       "field_guide",
			Prompt:         "Program active but no dose — what to check first?",
			ExpectCitation: true,
		},
		Answer:        "Check reservoir [1].\nSources:\n[1] type=field_guide source_id=23 chunk_id=489",
		CitationCount: 1,
	})
	if res.Passed {
		t.Fatalf("phase127 agronomy drift should fail: %+v", res)
	}
}

func TestScore_applyAnswerCritiqueFailsWhenEnabled(t *testing.T) {
	t.Parallel()
	res := Score(ScoreInput{
		Question: Question{ID: "smoke-ec-ph", Category: "field_guide"},
		Answer:   "Lettuce EC 1.0–1.3 mS/cm with pH 5.8.",
		Critique: farmguardian.AnswerCritique{Enabled: true, Pass: false, Reason: "Answer drifts off-topic"},
	})
	if res.Passed {
		t.Fatalf("expected critique fail: %+v", res)
	}
}
