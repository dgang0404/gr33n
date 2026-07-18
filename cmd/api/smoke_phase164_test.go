// Phase 164 — demo farm seed: living sensors, chrysanthemum crops, gravity drip.
package main

import (
	"context"
	"testing"
	"time"
)

func TestPhase164_Farm1NoCannabisPlantRow(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var n int
	err := testPool.QueryRow(ctx, `
SELECT count(*)::int FROM gr33ncrops.plants
WHERE farm_id = 1 AND crop_key = 'cannabis' AND deleted_at IS NULL`).Scan(&n)
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("farm 1 demo seed must not have cannabis plants row, got %d", n)
	}
}

func TestPhase164_Farm1ChrysanthemumDemoCycles(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Seed theming check only — is_active is intentionally not part of the
	// WHERE clause. This suite shares farm 1 with ~150 other smoke tests
	// that create/retire crop_cycles in the same zones (uq_active_crop_cycle
	// permits only one active cycle per zone), so run order can flip the
	// seeded row's is_active flag without touching its name/batch_label.
	var bloomBatch string
	err := testPool.QueryRow(ctx, `
SELECT cc.batch_label
FROM gr33nfertigation.crop_cycles cc
JOIN gr33ncore.zones z ON z.id = cc.zone_id
WHERE cc.farm_id = 1
  AND z.name = 'Flower Room'
  AND cc.name = 'Bloom run (12/12)'
ORDER BY cc.id
LIMIT 1`).Scan(&bloomBatch)
	if err != nil {
		t.Fatalf("Bloom run (12/12) cycle: %v", err)
	}
	if bloomBatch != "Zembla White" {
		t.Fatalf("Bloom run batch_label = %q, want Zembla White", bloomBatch)
	}

	var vegBatch string
	err = testPool.QueryRow(ctx, `
SELECT cc.batch_label
FROM gr33nfertigation.crop_cycles cc
JOIN gr33ncore.zones z ON z.id = cc.zone_id
WHERE cc.farm_id = 1
  AND z.name = 'Veg Room'
  AND cc.name = 'Veg canopy (18/6)'
ORDER BY cc.id
LIMIT 1`).Scan(&vegBatch)
	if err != nil {
		t.Fatalf("Veg canopy cycle: %v", err)
	}
	if vegBatch != "Anastasia Green" {
		t.Fatalf("Veg batch_label = %q, want Anastasia Green", vegBatch)
	}
}

func TestPhase164_Farm1WiredSensorsHaveReadings(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var wired int
	err := testPool.QueryRow(ctx, `
SELECT count(*)::int
FROM gr33ncore.sensors s
WHERE s.farm_id = 1 AND s.deleted_at IS NULL
  AND s.name = 'Air Temp Indoor'
  AND EXISTS (
    SELECT 1 FROM gr33ncore.sensor_readings sr
    WHERE sr.sensor_id = s.id
      AND sr.meta_data @> '{"seed":"phase164_demo"}'::jsonb
  )`).Scan(&wired)
	if err != nil {
		t.Fatal(err)
	}
	if wired != 1 {
		t.Fatalf("Air Temp Indoor should have phase164_demo readings, got %d", wired)
	}

	var unwired int
	err = testPool.QueryRow(ctx, `
SELECT count(*)::int
FROM gr33ncore.sensors s
WHERE s.farm_id = 1 AND s.deleted_at IS NULL
  AND s.name = 'Propagation Dome Temp'
  AND NOT EXISTS (
    SELECT 1 FROM gr33ncore.sensor_readings sr WHERE sr.sensor_id = s.id
  )`).Scan(&unwired)
	if err != nil {
		t.Fatal(err)
	}
	if unwired != 1 {
		t.Fatalf("Propagation Dome Temp should stay unwired (no readings), got %d", unwired)
	}
}

func TestPhase164_Farm1GravityDripProgram(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var irrigationOnly bool
	var runDur int
	err := testPool.QueryRow(ctx, `
SELECT irrigation_only, COALESCE(run_duration_seconds, 0)::int
FROM gr33nfertigation.programs
WHERE farm_id = 1 AND name = 'Herb Room Gravity Drip' AND deleted_at IS NULL
LIMIT 1`).Scan(&irrigationOnly, &runDur)
	if err != nil {
		t.Fatalf("Herb Room Gravity Drip program: %v", err)
	}
	if !irrigationOnly {
		t.Fatal("Herb Room Gravity Drip must be irrigation_only")
	}
	if runDur != 180 {
		t.Fatalf("run_duration_seconds = %d, want 180", runDur)
	}

	var eventCount int
	err = testPool.QueryRow(ctx, `
SELECT count(*)::int FROM gr33nfertigation.fertigation_events
WHERE farm_id = 1 AND notes LIKE '%[seed:herb-gravity-drip-demo]%'`).Scan(&eventCount)
	if err != nil {
		t.Fatal(err)
	}
	if eventCount < 1 {
		t.Fatal("expected seeded gravity-drip fertigation event")
	}
}
