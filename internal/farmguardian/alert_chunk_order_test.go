package farmguardian

import (
	"testing"

	db "gr33n-api/internal/db"
)

func chunkRow(id int64, sourceType, content string) db.SearchRagNearestNeighborsFilteredRow {
	return db.SearchRagNearestNeighborsFilteredRow{ID: id, SourceType: sourceType, ContentText: content}
}

func TestPrioritizeAlertChunks_sortsBySeverityDesc(t *testing.T) {
	t.Parallel()
	chunks := []db.SearchRagNearestNeighborsFilteredRow{
		chunkRow(1, "platform_doc", "workflow-guide"),
		chunkRow(2, "alert_notification", "alert_notification\nseverity: low\nsubject: Light schedule change"),
		chunkRow(3, "alert_notification", "alert_notification\nseverity: medium\nsubject: OHN batch below minimum"),
		chunkRow(4, "alert_notification", "alert_notification\nseverity: high\nsubject: Humidity high — Flower Room"),
	}
	out := PrioritizeAlertChunks(chunks)
	if len(out) != len(chunks) {
		t.Fatalf("expected %d chunks, got %d", len(chunks), len(out))
	}
	// High severity alert (id=4) should now be first, then medium (id=3), then low (id=2).
	if out[0].ID != 4 || out[1].ID != 3 || out[2].ID != 2 {
		t.Fatalf("expected severity-desc order [4,3,2,...], got ids: %d,%d,%d,%d", out[0].ID, out[1].ID, out[2].ID, out[3].ID)
	}
	// Non-alert chunk should follow the alerts.
	if out[3].ID != 1 {
		t.Fatalf("expected platform_doc chunk last, got id=%d", out[3].ID)
	}
}

func TestPrioritizeAlertChunks_leavesNonAlertChunksUntouched(t *testing.T) {
	t.Parallel()
	chunks := []db.SearchRagNearestNeighborsFilteredRow{
		chunkRow(1, "field_guide", "lettuce nutrition"),
		chunkRow(2, "platform_doc", "workflow guide"),
	}
	out := PrioritizeAlertChunks(chunks)
	if out[0].ID != 1 || out[1].ID != 2 {
		t.Fatalf("expected unchanged order, got ids: %d,%d", out[0].ID, out[1].ID)
	}
}

func TestPrioritizeAlertChunks_singleAlertUnchanged(t *testing.T) {
	t.Parallel()
	chunks := []db.SearchRagNearestNeighborsFilteredRow{
		chunkRow(1, "platform_doc", "workflow guide"),
		chunkRow(2, "alert_notification", "alert_notification\nseverity: low\nsubject: Light schedule change"),
	}
	out := PrioritizeAlertChunks(chunks)
	if out[0].ID != 1 || out[1].ID != 2 {
		t.Fatalf("single alert should not be reordered, got ids: %d,%d", out[0].ID, out[1].ID)
	}
}
