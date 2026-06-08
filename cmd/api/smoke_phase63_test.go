// Phase 63 — Guardian session memory helpers.
package main

import (
	"strings"
	"testing"
	"time"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmguardian"
)

func TestPhase63_SessionMemoryTopicsPresent(t *testing.T) {
	topics := farmguardian.InferSessionTopics("What supplies are low stock?", "restock CalMag")
	if len(topics) == 0 {
		t.Fatal("expected inferred topics")
	}
	block := farmguardian.PriorSessionContextBlock(db.Gr33ncoreSessionSummary{
		SummaryText: "Discussed VPD targets.",
		Topics:      []string{"grow"},
		CreatedAt:   time.Now().UTC().Add(-72 * time.Hour),
	}, time.Now().UTC())
	if !strings.Contains(block, "Prior session context") {
		t.Fatal("missing prior session block")
	}
}
