// Phase 136 — plant context bundle smokes.
package main

import (
	"context"
	"strings"
	"testing"
	"time"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmguardian"
)

func TestPhase136_PlantContextBundleIntent(t *testing.T) {
	t.Parallel()
	if !farmguardian.ShouldRunPlantContextBundleIntent("What stage is my veg grow?", nil) {
		t.Fatal("veg grow question should trigger bundle intent")
	}
	plan := farmguardian.PlanReadTools("How is my veg canopy doing?", &farmguardian.ContextRef{CropCycleID: 1}, farmguardian.Snapshot{})
	found := false
	for _, id := range plan.ToolIDs {
		if id == "plant_context_bundle" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("router plan missing plant_context_bundle: %v", plan.ToolIDs)
	}
}

func TestPhase136_EnrichVegGrowBundle(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	q := db.New(testPool)
	snap, err := farmguardian.BuildSnapshot(ctx, q, 1)
	if err != nil {
		t.Fatalf("BuildSnapshot: %v", err)
	}

	zones, err := q.ListZonesByFarm(ctx, 1)
	if err != nil {
		t.Fatalf("ListZones: %v", err)
	}
	var vegZoneID int64
	for _, z := range zones {
		if strings.EqualFold(z.Name, "Veg Room") {
			vegZoneID = z.ID
			break
		}
	}
	if vegZoneID == 0 {
		t.Skip("Veg Room zone not in seed")
	}

	cycle, err := q.GetActiveCropCycleForZone(ctx, vegZoneID)
	if err != nil {
		t.Skip("no active veg cycle in seed")
	}

	ref := &farmguardian.ContextRef{
		Type:        "zone",
		ID:          vegZoneID,
		Name:        "Veg Room",
		CropCycleID: cycle.ID,
	}
	block := farmguardian.EnrichPromptBlock(ctx, q, 1, "What stage is my veg grow?", snap, ref)
	if block == "" {
		t.Fatal("expected plant context bundle enrichment")
	}
	if !strings.Contains(block, "plant_context_bundle") {
		t.Fatalf("block missing plant_context_bundle:\n%s", block)
	}
	for _, want := range []string{"lookup_crop_targets", "grow_advisor"} {
		if !strings.Contains(block, want) {
			t.Fatalf("bundle missing %q:\n%s", want, block)
		}
	}
	// Bundle should dedupe standalone duplicate tools.
	if strings.Count(block, "lookup_crop_targets —") > 2 {
		t.Fatalf("expected deduped lookup_crop_targets, got %d sections", strings.Count(block, "lookup_crop_targets —"))
	}
}
