// Phase 32 WS1 — grow read layer smokes (snapshot plants/programs + read tools).
package main

import (
	"context"
	"strings"
	"testing"
	"time"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmguardian"
)

func TestPhase32WS1_SnapshotProgramsByZone(t *testing.T) {
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
	if len(snap.ProgramsByZone) == 0 {
		t.Fatal("expected seeded active programs in snapshot ProgramsByZone")
	}
	rendered := snap.Render()
	if !strings.Contains(rendered, "Active fertigation programs by zone:") {
		t.Fatalf("snapshot render missing programs block:\n%s", rendered)
	}
	for _, zp := range snap.ProgramsByZone {
		if len(zp.Programs) == 0 {
			t.Fatalf("zone %q has empty program list", zp.ZoneName)
		}
	}
}

func TestPhase32WS1_EnrichSummarizeZoneFertigation(t *testing.T) {
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
	block := farmguardian.EnrichPromptBlock(ctx, q, 1, "what fertigation program runs in Veg Room?", snap, nil)
	if block == "" {
		t.Fatal("expected read-tool enrichment for fertigation question")
	}
	if !strings.Contains(block, "summarize_zone_fertigation") {
		t.Fatalf("block missing summarize_zone_fertigation:\n%s", block)
	}
	if !strings.Contains(block, "Veg Room") {
		t.Fatalf("block missing zone name:\n%s", block)
	}
	if !strings.Contains(block, "Veg Daily JLF Program") {
		t.Fatalf("block missing seeded program name:\n%s", block)
	}
}

func TestPhase32WS1_ListPlantsEmptyFarm(t *testing.T) {
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
	block := farmguardian.EnrichPromptBlock(ctx, q, 1, "list my plants", snap, nil)
	if block == "" {
		t.Fatal("expected list_plants enrichment block")
	}
	if !strings.Contains(block, "list_plants") {
		t.Fatalf("block missing list_plants:\n%s", block)
	}
}
