// Phase 33 WS2 — context_ref zone dedup: when Ask Guardian is opened from a zone
// card, the enriched focus block carries readings and summarize_zone is skipped
// for that same zone (one zone block, not two).
package main

import (
	"context"
	"strings"
	"testing"
	"time"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmguardian"
)

// resolveZoneIDByName finds a seeded zone id by name for farm 1.
func resolveZoneIDByName(ctx context.Context, q *db.Queries, name string) (int64, bool) {
	zones, err := q.ListZonesByFarm(ctx, 1)
	if err != nil {
		return 0, false
	}
	for _, z := range zones {
		if z.Name == name {
			return z.ID, true
		}
	}
	return 0, false
}

func TestPhase33WS2_ZoneContextRefSkipsSummarizeZone(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	q := db.New(testPool)
	snap, err := farmguardian.BuildSnapshot(ctx, q, 1)
	if err != nil {
		t.Fatalf("BuildSnapshot: %v", err)
	}

	zoneID, ok := resolveZoneIDByName(ctx, q, "Flower Room")
	if !ok {
		t.Skip("seeded zone 'Flower Room' not found")
	}

	question := "what's the humidity in Flower Room?"

	// Without a context_ref, summarize_zone enriches as usual (WS1 behavior).
	plain := farmguardian.EnrichPromptBlock(ctx, q, 1, question, snap, nil)
	if !strings.Contains(plain, "summarize_zone") {
		t.Fatalf("expected summarize_zone without context_ref:\n%s", plain)
	}

	// With a matching zone context_ref, summarize_zone is skipped (the focus
	// block injected by the handler carries the readings instead).
	ref := &farmguardian.ContextRef{Type: "zone", ID: zoneID, Name: "Flower Room"}
	deduped := farmguardian.EnrichPromptBlock(ctx, q, 1, question, snap, ref)
	if strings.Contains(deduped, "summarize_zone") {
		t.Fatalf("zone context_ref must skip summarize_zone read tool:\n%s", deduped)
	}

	// The enriched focus block still carries the zone's readings.
	focus := farmguardian.ContextRefPromptBlock(ctx, q, 1, *ref)
	if !strings.Contains(focus, "Latest sensor readings") && !strings.Contains(focus, "none configured") {
		t.Fatalf("zone focus block should carry readings (WS2 enrichment):\n%s", focus)
	}
}

func TestPhase33WS2_NonMatchingZoneRefKeepsSummarizeZone(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	q := db.New(testPool)
	snap, err := farmguardian.BuildSnapshot(ctx, q, 1)
	if err != nil {
		t.Fatalf("BuildSnapshot: %v", err)
	}

	// Context ref points at a different anchor type — summarize_zone still runs.
	ref := &farmguardian.ContextRef{Type: "alert", ID: 999999}
	block := farmguardian.EnrichPromptBlock(ctx, q, 1, "what's the humidity in Flower Room?", snap, ref)
	if !strings.Contains(block, "summarize_zone") {
		t.Fatalf("non-zone context_ref must not skip summarize_zone:\n%s", block)
	}
}
