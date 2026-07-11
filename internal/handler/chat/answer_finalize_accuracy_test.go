// Phase 152 WS1 — live accuracy-note wiring. Confirms the response-shaped
// citations correctly feed farmguardian.AnswerAccuracyNote and that a clean
// answer stays silent (no false-positive banner on every grounded turn).

package chat

import (
	"testing"

	"gr33n-api/internal/rag/synthesis"
)

func TestApplyAnswerAccuracyNote_flagsGarbledTruncation(t *testing.T) {
	t.Parallel()
	answer := "Lights are consistent and ade0:"
	if note := applyAnswerAccuracyNote(answer, nil); note == "" {
		t.Fatal("expected accuracy note for garbled truncation")
	}
}

func TestApplyAnswerAccuracyNote_cleanAnswerSilent(t *testing.T) {
	t.Parallel()
	answer := "Humidity is high in the Flower Room [1] and OHN batch is low [2]; reorder soon."
	cites := []synthesis.Citation{
		{Ref: 1, SourceType: "alert_notification", Excerpt: "severity: high\nsubject: Humidity high — Flower Room."},
		{Ref: 2, SourceType: "alert_notification", Excerpt: "OHN batch below minimum — reorder or brew soon."},
	}
	if note := applyAnswerAccuracyNote(answer, cites); note != "" {
		t.Fatalf("unexpected note %q on clean grounded answer", note)
	}
}

func TestApplyAnswerAccuracyNote_flagsInventedAssumptionMath(t *testing.T) {
	t.Parallel()
	answer := "That's about ~1.2 mL per plant if we assume an average yield density for your cultivar."
	if note := applyAnswerAccuracyNote(answer, nil); note == "" {
		t.Fatal("expected accuracy note for invented assumption math")
	}
}
