package farmguardian

import (
	"strings"
	"testing"
)

func TestEcphCropDriftNote_detectsBlueberryTail(t *testing.T) {
	t.Parallel()
	prompt := "What does our operational documentation say about EC and pH targets for leafy greens here?"
	answer := "Lettuce EC 1.0–1.3 mS/cm and pH 5.5–6.0 per our field guide [1].\n\n" +
		"What about blueberry pH targets for acidic fruiting crops?"
	if note := EcphCropDriftNote(prompt, answer); note == "" {
		t.Fatal("expected blueberry crop drift")
	}
}

func TestEcphCropDriftNote_passesWhenCropInPrompt(t *testing.T) {
	t.Parallel()
	prompt := "What EC and pH should we run for blueberry in zone B?"
	answer := "Blueberry needs pH 4.5–5.5 and lower EC than lettuce [1]."
	if note := EcphCropDriftNote(prompt, answer); note != "" {
		t.Fatalf("unexpected note %q", note)
	}
}

func TestTrimUncitedTail_removesBlueberryAppendix(t *testing.T) {
	t.Parallel()
	prompt := "What does our operational documentation say about EC and pH targets for leafy greens here?"
	answer := "Lettuce EC 1.0–1.3 mS/cm and pH 5.5–6.0 per our field guide [1].\n\n" +
		"What about blueberry pH targets for acidic fruiting crops?"
	cites := []CitationSummary{
		{Ref: 1, Excerpt: "Leafy greens EC and pH targets for lettuce and kale."},
	}
	got, meta := TrimUncitedTail(answer, prompt, cites)
	if !meta.Trimmed {
		t.Fatal("expected trim")
	}
	if strings.Contains(strings.ToLower(got), "blueberry") {
		t.Fatalf("blueberry tail should be removed: %q", got)
	}
	if !strings.Contains(got, "Lettuce EC") {
		t.Fatalf("opening should remain: %q", got)
	}
}

func TestTrimUncitedTail_noOpOnCleanAnswer(t *testing.T) {
	t.Parallel()
	prompt := "What EC range does the platform recommend for hydro lettuce?"
	answer := "Hydro lettuce targets EC 0.8–1.3 mS/cm and pH 5.5–6.0 per our field guide [1]."
	cites := []CitationSummary{
		{Ref: 1, Excerpt: "Hydro lettuce EC 0.8–1.3 mS/cm; pH 5.5–6.0."},
	}
	got, meta := TrimUncitedTail(answer, prompt, cites)
	if meta.Trimmed || got != answer {
		t.Fatalf("unexpected trim meta=%+v got=%q", meta, got)
	}
}

func TestSmokeTopicDriftNote_ecphBlueberryDrift(t *testing.T) {
	t.Parallel()
	prompt := "What does our operational documentation say about EC and pH targets for leafy greens here?"
	answer := "Lettuce EC 1.0–1.3 mS/cm and pH 5.5–6.0.\n\nWhat blueberry pH should we use?"
	note := SmokeTopicDriftNote(SmokeTopicDriftInput{
		QuestionID: "smoke-ec-ph",
		Category:   "field_guide",
		Prompt:     prompt,
		Answer:     answer,
	})
	if note == "" {
		t.Fatal("expected crop drift note")
	}
}
