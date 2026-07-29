package eval

import (
	"strings"
	"testing"

	"gr33n-api/internal/farmguardian"
)

func TestScore_writeIntentSeparatesAnswerRelevance(t *testing.T) {
	res := Score(ScoreInput{
		Question:      Question{ID: "write-ack", Category: "write_intent", ExpectProposal: true, Grounded: true},
		Answer:        "As an AI I cannot help with that unrelated essay.",
		ProposalCount: 1,
		ProposalIDs:   []string{"p1"},
		Relevance: farmguardian.AnswerRelevance{
			QuestionAnswerCosine: 0.1,
			LowRelevance:         true,
		},
	})
	if !res.Passed {
		t.Fatalf("proposal presence should still pass, notes=%q", res.Notes)
	}
	if res.AnswerRelevant == nil || *res.AnswerRelevant {
		t.Fatal("expected AnswerRelevant=false")
	}
	if !strings.Contains(res.Notes, "answer_low_relevance") {
		t.Fatalf("expected low-relevance note, got %q", res.Notes)
	}
}

func TestScore_smokeECPH_citeExcerptSoftens(t *testing.T) {
	res := Score(ScoreInput{
		Question:      Question{ID: "smoke-ec-ph", Category: "field_guide", Grounded: true},
		Answer:        "EC for leafy greens is about 1.2–1.8 mS/cm.",
		CitationCount: 1,
		Citations:     []farmguardian.CitationSummary{{Excerpt: "Target pH 5.5–6.5"}},
	})
	if !res.Passed {
		t.Fatalf("expected pass when cite has pH, notes=%q", res.Notes)
	}
}
