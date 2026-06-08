package farmguardian

import (
	"strings"
	"testing"

	db "gr33n-api/internal/db"
)

func TestInferSessionTopics(t *testing.T) {
	topics := InferSessionTopics("Why is VPD high in the flower room?", "comfort band breach")
	if !containsString(topics, "comfort") && !containsString(topics, "grow") {
		t.Fatalf("expected comfort or grow topics, got %v", topics)
	}
}

func TestTopicsForRoute(t *testing.T) {
	if got := TopicsForRoute("/alerts"); !containsString(got, "alerts") {
		t.Fatalf("alerts route: %v", got)
	}
	if got := TopicsForRoute("/zones/12"); len(got) == 0 {
		t.Fatal("zone route should map to topics")
	}
}

func TestTopicsOverlap(t *testing.T) {
	if !TopicsOverlap([]string{"grow", "comfort"}, []string{"comfort"}) {
		t.Fatal("expected overlap")
	}
	if TopicsOverlap([]string{"alerts"}, []string{"stock"}) {
		t.Fatal("expected no overlap")
	}
}

func TestPriorSessionContextBlock(t *testing.T) {
	block := PriorSessionContextBlock(db.Gr33ncoreSessionSummary{
		SummaryText: "You asked about high VPD in Flower Room.",
		Topics:      []string{"grow", "comfort"},
	}, nowFunc())
	if !strings.Contains(block, "Prior session context") {
		t.Fatalf("missing header: %q", block)
	}
	if !strings.Contains(block, "VPD") {
		t.Fatalf("missing summary: %q", block)
	}
	if !strings.Contains(block, "Do not repeat") {
		t.Fatalf("missing instruction: %q", block)
	}
}

func containsString(ss []string, want string) bool {
	for _, s := range ss {
		if s == want {
			return true
		}
	}
	return false
}
