package synthesis

import (
	"testing"

	db "gr33n-api/internal/db"
)

func TestRefNumbersInAnswer(t *testing.T) {
	s := "The pump ran [2] and later [1] per [2]."
	got := RefNumbersInAnswer(s)
	if len(got) != 2 || got[0] != 1 || got[1] != 2 {
		t.Fatalf("got %v", got)
	}
}

func TestBuildCitations(t *testing.T) {
	chunks := []db.SearchRagNearestNeighborsFilteredRow{
		{ID: 10, SourceType: "task", SourceID: 1, ContentText: "hello world"},
		{ID: 11, SourceType: "automation_run", SourceID: 2, ContentText: "done"},
	}
	ans := "Tasks show [1] and runs [2]. Bad [99]."
	cites := BuildCitations(ans, chunks)
	if len(cites) != 2 {
		t.Fatalf("want 2 cites, got %d", len(cites))
	}
	if cites[0].Ref != 1 || cites[0].ChunkID != 10 {
		t.Fatal(cites)
	}
	if cites[1].Ref != 2 || cites[1].ChunkID != 11 {
		t.Fatal(cites)
	}
}
