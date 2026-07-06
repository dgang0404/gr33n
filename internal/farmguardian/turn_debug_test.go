package farmguardian

import (
	"testing"

	db "gr33n-api/internal/db"
)

func TestCountRAGChunksBySource_groupsAndUnknown(t *testing.T) {
	chunks := []db.SearchRagNearestNeighborsFilteredRow{
		{SourceType: "field_guide"},
		{SourceType: "field_guide"},
		{SourceType: "platform_doc"},
		{SourceType: ""},
	}
	got := CountRAGChunksBySource(chunks)
	if got["field_guide"] != 2 || got["platform_doc"] != 1 || got["unknown"] != 1 {
		t.Fatalf("counts: %#v", got)
	}
}

func TestBuildTurnDebug_includesToolsAndTrim(t *testing.T) {
	trim := &TrimSummary{HistoryTurns: "20→8", RAGTopK: "8→4"}
	dbg := BuildTurnDebug(
		"req-1",
		ToolPlan{ToolIDs: []string{"walk_farm", "summarize_unread_alerts"}},
		[]db.SearchRagNearestNeighborsFilteredRow{{SourceType: "field_guide"}},
		trim,
		"phi3:mini",
		4096,
		8192,
		DefaultPromptBudget(20),
	)
	if dbg == nil {
		t.Fatal("nil debug")
	}
	if dbg.RequestID != "req-1" {
		t.Fatalf("request_id %q", dbg.RequestID)
	}
	if len(dbg.ToolsPlanned) != 2 || dbg.RAGChunkTotal != 1 {
		t.Fatalf("tools=%v rag_total=%d", dbg.ToolsPlanned, dbg.RAGChunkTotal)
	}
	if dbg.TrimSummary == nil || dbg.TrimSummary.HistoryTurns != "20→8" {
		t.Fatalf("trim %#v", dbg.TrimSummary)
	}
	if dbg.EffectiveContextWindow != 4096 || dbg.Model != "phi3:mini" {
		t.Fatalf("model/window %q %d", dbg.Model, dbg.EffectiveContextWindow)
	}
}
