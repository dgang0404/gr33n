// Phase 127 — snapshot device/fertigation posture + field guide smokes.
package main

import (
	"context"
	"strings"
	"testing"
	"time"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmguardian"
)

func TestPhase127_SnapshotDevicesAndFertigationSchedule(t *testing.T) {
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
	if snap.Devices.Total < 1 {
		t.Fatalf("expected seeded edge devices, got %+v", snap.Devices)
	}
	if snap.FertigationSchedule.ScheduledActive < 1 {
		t.Fatalf("expected scheduled active programs, got %+v", snap.FertigationSchedule)
	}
	rendered := snap.Render()
	for _, want := range []string{
		"Edge devices:",
		"Fertigation programs:",
		"on schedule",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("snapshot render missing %q:\n%s", want, rendered)
		}
	}
}

func TestPhase127_EnrichDeviceHealthDemoFarm(t *testing.T) {
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
	block := farmguardian.EnrichPromptBlock(ctx, q, 1, "are any Pis offline on this farm?", snap, nil)
	if block == "" {
		t.Fatal("expected summarize_device_health enrichment")
	}
	if !strings.Contains(block, "summarize_device_health") {
		t.Fatalf("block missing summarize_device_health:\n%s", block)
	}
}

func TestPhase127_EnrichFertigationTroubleshooting(t *testing.T) {
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
	block := farmguardian.EnrichPromptBlock(ctx, q, 1, "why didn't the veg fertigation program run?", snap, nil)
	if block == "" {
		t.Fatal("expected fertigation read-tool enrichment")
	}
	if !strings.Contains(block, "summarize_zone_fertigation") && !strings.Contains(block, "lookup_crop_targets") {
		t.Fatalf("block missing fertigation tools:\n%s", block)
	}
}
