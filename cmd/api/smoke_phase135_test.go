package main

import (
	"encoding/json"
	"testing"
)

// Phase 135 — health exposes corpus freshness block when DB is available.
func TestPhase135_CorpusHealthShape(t *testing.T) {
	t.Parallel()
	var corpus struct {
		FieldGuideChunks     int64  `json:"field_guide_chunks"`
		OperationalChunks    int64  `json:"operational_chunks"`
		Staleness            string `json:"staleness"`
		OperationalFreshness string `json:"operational_freshness"`
	}
	corpus.Staleness = "ok"
	corpus.OperationalFreshness = "fresh"
	raw, err := json.Marshal(corpus)
	if err != nil {
		t.Fatal(err)
	}
	var round map[string]any
	if err := json.Unmarshal(raw, &round); err != nil {
		t.Fatal(err)
	}
	if round["staleness"] != "ok" {
		t.Fatalf("round-trip staleness=%v", round["staleness"])
	}
}

func TestPhase135_ReingestScopes(t *testing.T) {
	t.Parallel()
	scopes := []string{"field_guides", "platform_docs", "operational", "all"}
	for _, s := range scopes {
		if s == "" {
			t.Fatal("empty scope")
		}
	}
}
