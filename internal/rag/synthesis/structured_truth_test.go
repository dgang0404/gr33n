package synthesis

import (
	"strings"
	"testing"

	db "gr33n-api/internal/db"
)

func TestStripNutrientNumbersFromFieldGuideChunk(t *testing.T) {
	chunks := []db.SearchRagNearestNeighborsFilteredRow{
		{
			SourceType:  "field_guide",
			ContentText: "field_guide\ndoc_path: crop-cannabis-nutrition.md\n\nTargets ramp to ~1.6–2.0 mS/cm in mid-flower.",
		},
		{
			SourceType:  "platform_doc",
			ContentText: "Settings EC override applies at 2.2 mS/cm immediately.",
		},
	}
	out := StripNutrientNumbersFromChunks(chunks)
	if strings.Contains(out[0].ContentText, "1.6") || strings.Contains(out[0].ContentText, "2.0") {
		t.Fatalf("field_guide numbers not stripped: %q", out[0].ContentText)
	}
	if !strings.Contains(out[0].ContentText, "lookup_crop_targets") {
		t.Fatal("expected structured-truth hint on stripped line")
	}
	if !strings.Contains(out[1].ContentText, "2.2") {
		t.Fatal("platform_doc chunk should be unchanged")
	}
}

func TestStripNutrientNumbersNoOpWithoutMetrics(t *testing.T) {
	raw := "field_guide\n\nWater when the slab feels light — no numbers here."
	out := StripNutrientNumbersFromChunks([]db.SearchRagNearestNeighborsFilteredRow{
		{SourceType: "field_guide", ContentText: raw},
	})
	if out[0].ContentText != raw {
		t.Fatalf("unexpected mutation: %q", out[0].ContentText)
	}
}
