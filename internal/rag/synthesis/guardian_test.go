package synthesis

import (
	"testing"

	db "gr33n-api/internal/db"
)

func TestHasPlatformDocChunks(t *testing.T) {
	if HasPlatformDocChunks(nil) {
		t.Fatal("expected false for nil")
	}
	if HasPlatformDocChunks([]db.SearchRagNearestNeighborsFilteredRow{
		{SourceType: "task"},
	}) {
		t.Fatal("expected false for task only")
	}
	if !HasPlatformDocChunks([]db.SearchRagNearestNeighborsFilteredRow{
		{SourceType: "platform_doc"},
	}) {
		t.Fatal("expected true for platform_doc")
	}
}

func TestGuardianRAGInstructionsIncludesPlatformDocHint(t *testing.T) {
	base := GuardianRAGInstructions(nil)
	withDoc := GuardianRAGInstructions([]db.SearchRagNearestNeighborsFilteredRow{
		{SourceType: "platform_doc"},
	})
	if withDoc == base {
		t.Fatal("expected extra platform_doc guidance")
	}
	if len(withDoc) <= len(base) {
		t.Fatal("expected longer instructions with platform_doc chunks")
	}
}
