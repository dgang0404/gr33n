package farmguardian

import (
	"strings"
	"testing"
)

func TestCitationAlignmentNote_offTopicCitations(t *testing.T) {
	t.Parallel()
	question := "What does our operational documentation say about EC and pH targets for leafy greens here?"
	answer := "Lettuce EC 1.0–1.3 mS/cm and pH 5.5–6.0.\n\n" +
		"Endocrine disruptors in Lake Erie wildlife show profound hormonal effects."
	cites := []CitationSummary{
		{Ref: 1, SourceType: "field_guide", Excerpt: "Lettuce EC targets 0.8–1.3 mS/cm."},
		{Ref: 6, SourceType: "field_guide", Excerpt: "Endocrine disruptors in aquatic lifeforms and Lake Erie ecosystem."},
	}
	if got := CitationAlignmentNote(question, answer, cites); got == "" {
		t.Fatal("expected misaligned citation note")
	}
}

func TestCitationAlignmentNote_alignedAgronomy(t *testing.T) {
	t.Parallel()
	question := "What EC range does the platform recommend for hydro lettuce?"
	answer := "Hydro lettuce targets EC 0.8–1.3 mS/cm and pH 5.5–6.0 per our field guide [1]."
	cites := []CitationSummary{
		{Ref: 1, SourceType: "field_guide", Excerpt: "Hydro lettuce EC 0.8–1.3 mS/cm; pH 5.5–6.0."},
	}
	if got := CitationAlignmentNote(question, answer, cites); got != "" {
		t.Fatalf("expected pass, got %q", got)
	}
}

func TestCitationAlignmentNote_uncitedTail(t *testing.T) {
	t.Parallel()
	question := "What does our operational documentation say about EC and pH targets for leafy greens here?"
	opening := strings.Repeat("Lettuce EC 1.0–1.3 mS/cm and pH 5.5–6.0 per our field guide. ", 12)
	tail := strings.Repeat("Typha latifolia biosorption wetlands reduce contaminant concentrations substantially. ", 4)
	answer := opening + "\n\n" + tail
	cites := []CitationSummary{
		{Ref: 1, SourceType: "field_guide", Excerpt: "Leafy greens EC and pH targets for lettuce."},
	}
	if got := CitationAlignmentNote(question, answer, cites); got == "" {
		t.Fatal("expected uncited tail or misalignment")
	}
}
