package farmguardian

import "testing"

func TestBuildTrimSummary_reducedHistory(t *testing.T) {
	full := DefaultPromptBudget(20)
	applied := full
	applied.MaxHistoryTurns = 8
	applied.RAGTopK = 5
	log := []string{"history turns 20→8", "RAG topK 8→5", "snapshot caps reduced (context_window=4096)"}
	s := BuildTrimSummary(full, applied, log, 4096)
	if s == nil {
		t.Fatal("expected summary")
	}
	if s.HistoryTurns != "20→8" {
		t.Fatalf("history %q", s.HistoryTurns)
	}
	if s.RAGTopK != "8→5" {
		t.Fatalf("rag %q", s.RAGTopK)
	}
	if !s.SnapshotReduced {
		t.Fatal("expected snapshot_reduced")
	}
	if s.EffectiveContextWindow != 4096 {
		t.Fatalf("window %d", s.EffectiveContextWindow)
	}
}

func TestBuildTrimSummary_nilWhenNoTrim(t *testing.T) {
	full := DefaultPromptBudget(20)
	if BuildTrimSummary(full, full, nil, 8192) != nil {
		t.Fatal("expected nil")
	}
}

func TestGroundedHonestyPromptBlock(t *testing.T) {
	b := GroundedHonestyPromptBlock()
	if b == "" || !containsSubstr(b, "LIVE FARM DATA") {
		t.Fatalf("block %q", b)
	}
}

func containsSubstr(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && stringIndexHonesty(s, sub) >= 0)
}

func stringIndexHonesty(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
