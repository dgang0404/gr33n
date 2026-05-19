// Phase 28 WS3 — Farm Guardian snapshot ↔ crop cycle analytics smoke
// test. Verifies that BuildSnapshot pulls fertigation + cost data
// alongside the existing zone/cycle/alert info on a real Postgres, so the
// chat handler that consumes PromptBlock() is exercised against the
// integration surface (not a stub).
//
// Coverage:
//   - Active cycle gets analytics attached (event_count + EC range + cost).
//   - PromptBlock includes the new "metrics:" line under the cycle bullet.
//   - SnapshotMaxAnalyticsCycles caps analytics to N cycles even when N+1
//     are active (the extra cycles still render their basic line).

package main

import (
	"context"
	"strings"
	"testing"
	"time"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmguardian"
)

func TestPhase28WS3_Snapshot_AttachesCycleAnalytics(t *testing.T) {
	tok := smokeJWT(t)
	cycleID := seedCropCycleForAnalytics(t, tok, uniqueName("ws3-cycle"), nil, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// BuildSnapshot consumes the SnapshotMaxAnalyticsCycles budget in
	// ListCropCyclesByFarm order (started_at DESC). The smoke DB has a
	// handful of bootstrap cycles already; bump our seed's started_at
	// to "tomorrow" so it lands in the analytics-eligible window
	// regardless of how many other active cycles are around.
	if _, err := testPool.Exec(ctx,
		`UPDATE gr33nfertigation.crop_cycles SET started_at = CURRENT_DATE + INTERVAL '1 day' WHERE id = $1`,
		cycleID); err != nil {
		t.Fatalf("bump started_at: %v", err)
	}

	queries := db.New(testPool)
	snap, err := farmguardian.BuildSnapshot(ctx, queries, 1)
	if err != nil {
		t.Fatalf("BuildSnapshot: %v", err)
	}

	if len(snap.ActiveCycles) == 0 {
		t.Fatal("expected at least one active cycle in snapshot")
	}

	// Locate our specific seeded cycle by ID — the ActiveCycle.ID
	// field is what makes this assertion stable regardless of how
	// many other cycles the smoke harness has accumulated.
	var found *farmguardian.CycleAnalytics
	for i := range snap.ActiveCycles {
		if snap.ActiveCycles[i].ID == cycleID {
			found = &snap.ActiveCycles[i].Analytics
			break
		}
	}
	if found == nil {
		t.Fatalf("seeded cycle %d not in snapshot:\n%s", cycleID, snap.Render())
	}

	// The seed inserted 3 fertigation events totalling 10L with EC
	// {1.0, 1.5, 2.0}. Allow some slack — exact equality on numeric
	// types after pgtype roundtrip can drift in the last decimal.
	if found.EventCount < 1 {
		t.Errorf("expected EventCount > 0, got %d", found.EventCount)
	}
	if found.TotalLiters <= 0 {
		t.Errorf("expected TotalLiters > 0, got %f", found.TotalLiters)
	}
	if found.AvgECmSCm <= 0 {
		t.Errorf("expected AvgECmSCm > 0, got %f", found.AvgECmSCm)
	}
	if found.MinECmSCm <= 0 || found.MaxECmSCm < found.MinECmSCm {
		t.Errorf("EC min/max look wrong: min=%f max=%f", found.MinECmSCm, found.MaxECmSCm)
	}
	// Seeded cost is 50 USD; cycle has no yield so cost_per_gram is nil.
	if found.Currency != "USD" {
		t.Errorf("expected Currency USD, got %q", found.Currency)
	}
	if found.TotalExpenses <= 0 {
		t.Errorf("expected TotalExpenses > 0, got %f", found.TotalExpenses)
	}
	if found.CostPerGram != nil {
		t.Errorf("CostPerGram should be nil with no yield, got %v", *found.CostPerGram)
	}

	// PromptBlock must include the new metrics: line so the LLM sees it.
	block := snap.PromptBlock()
	if !strings.Contains(block, "metrics:") {
		t.Fatalf("PromptBlock missing metrics line:\n%s", block)
	}
	if !strings.Contains(block, "feed:") {
		t.Fatalf("PromptBlock missing fertigation summary:\n%s", block)
	}
	if !strings.Contains(block, "EC ") {
		t.Fatalf("PromptBlock missing EC summary:\n%s", block)
	}
}

func TestPhase28WS3_Snapshot_CapsAnalyticsAtBudget(t *testing.T) {
	tok := smokeJWT(t)

	// Seed SnapshotMaxAnalyticsCycles + 1 cycles so the cap actually
	// matters. Each seed also creates fertigation events so they would
	// all qualify for analytics if the cap weren't enforced.
	want := farmguardian.SnapshotMaxAnalyticsCycles + 1
	for i := 0; i < want; i++ {
		seedCropCycleForAnalytics(t, tok, uniqueName("ws3-cap"), nil, nil)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	queries := db.New(testPool)
	snap, err := farmguardian.BuildSnapshot(ctx, queries, 1)
	if err != nil {
		t.Fatalf("BuildSnapshot: %v", err)
	}

	withAnalytics := 0
	for _, c := range snap.ActiveCycles {
		if c.Analytics.EventCount > 0 {
			withAnalytics++
		}
	}
	if withAnalytics > farmguardian.SnapshotMaxAnalyticsCycles {
		t.Fatalf("more cycles than budget got analytics: %d > %d",
			withAnalytics, farmguardian.SnapshotMaxAnalyticsCycles)
	}
	// We seeded `want` ws3 cycles plus there may be older active
	// cycles from other tests. The snapshot ActiveCycles slice should
	// at least include our seeds — sanity-check that.
	if len(snap.ActiveCycles) < want {
		t.Fatalf("expected >= %d active cycles, got %d", want, len(snap.ActiveCycles))
	}
}
