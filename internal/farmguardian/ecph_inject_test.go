package farmguardian

import (
	"strings"
	"testing"

	db "gr33n-api/internal/db"
)

func TestInjectPHFromChunks(t *testing.T) {
	chunks := []db.SearchRagNearestNeighborsFilteredRow{
		{ContentText: "Leafy greens: maintain pH 5.5–6.5 in reservoir."},
	}
	got, ok := InjectPHFromChunks("EC targets are typically 1.2–1.8 mS/cm for lettuce.", chunks)
	if !ok {
		t.Fatal("expected pH inject")
	}
	if !strings.Contains(strings.ToLower(got), "ph") {
		t.Fatalf("expected ph in answer: %q", got)
	}
	_, ok = InjectPHFromChunks("EC 1.2 and pH 5.8 already present.", chunks)
	if ok {
		t.Fatal("should not inject when pH already present")
	}
}

func TestCitationExcerptsContain(t *testing.T) {
	cites := []CitationSummary{{Excerpt: "Target pH 5.8–6.2"}}
	if !CitationExcerptsContain(cites, "ph") {
		t.Fatal("expected cite hit")
	}
	if CitationExcerptsContain(cites, "mars") {
		t.Fatal("unexpected cite hit")
	}
}
