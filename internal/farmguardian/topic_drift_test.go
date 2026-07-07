package farmguardian

import (
	"strings"
	"testing"
)

func TestSmokeTopicDriftNote_run3MorningWalkHygiene(t *testing.T) {
	t.Parallel()
	answer := `Check veg EC per [task #5](https://gr33n-docs/phase_40.plan.md#tasks).
I apologize for misunderstanding. Here's an updated answer:`
	note := SmokeTopicDriftNote(SmokeTopicDriftInput{
		QuestionID: "smoke-morning-walk",
		Category:   "farm_state",
		Prompt:     "What should I check first on a morning walkthrough of this farm today?",
		Answer:     answer,
	})
	if note == "" {
		t.Fatal("expected hygiene failure")
	}
}

func TestSmokeTopicDriftNote_citationMisaligned(t *testing.T) {
	t.Parallel()
	answer := `Lettuce EC 1.0–1.3 mS/cm and pH 5.5–6.0.
Endocrine disruptors in Lake Erie wildlife show profound effects.`
	cites := []CitationSummary{
		{Ref: 1, Excerpt: "Lettuce EC targets 0.8–1.3 mS/cm."},
		{Ref: 6, Excerpt: "Endocrine disruptors in Lake Erie ecosystem."},
	}
	note := SmokeTopicDriftNote(SmokeTopicDriftInput{
		QuestionID: "smoke-ec-ph",
		Category:   "field_guide",
		Prompt:     "What EC and pH targets for leafy greens?",
		Answer:     answer,
		Citations:  cites,
	})
	if note != "citation_misaligned" && note != "topic_drift: off-topic from leafy greens EC/pH" {
		t.Fatalf("note=%q", note)
	}
}

func TestSmokeTopicDriftNote_lowRelevance(t *testing.T) {
	t.Parallel()
	note := SmokeTopicDriftNote(SmokeTopicDriftInput{
		QuestionID: "smoke-ec-ph",
		Category:   "field_guide",
		Prompt:     "EC and pH for lettuce",
		Answer:     "On-topic opening only.",
		Relevance: AnswerRelevance{
			QuestionAnswerCosine: 0.12,
			OpeningTailCosine:    0.9,
			LowRelevance:         true,
			MinThreshold:         0.35,
		},
	})
	if note == "" || !strings.Contains(note, "low_relevance") {
		t.Fatalf("note=%q", note)
	}
}

func TestSmokeTopicDriftNote_cleanPasses(t *testing.T) {
	t.Parallel()
	note := SmokeTopicDriftNote(SmokeTopicDriftInput{
		QuestionID: "smoke-unread-alerts",
		Category:   "farm_state",
		Answer:     "You have two humidity alerts in Flower Room and low calcium stock.",
	})
	if note != "" {
		t.Fatalf("unexpected note %q", note)
	}
}
