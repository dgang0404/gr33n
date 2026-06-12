// Phase 103 — legacy plant dedupe: typo rows merge to one crop_key; cycles keep plant_id.

package main

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestPhase103_LegacyPlantMerge(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	tag := fmt.Sprintf("phase103_%d", time.Now().UnixNano())
	var farmID int64 = 1

	// Two typo tomato rows + cycles linked to each duplicate.
	var idA, idB int64
	err := testPool.QueryRow(ctx, `
INSERT INTO gr33ncrops.plants (farm_id, display_name, variety_or_cultivar, meta)
VALUES ($1, 'Tomato', 'typo A', '{}'::jsonb) RETURNING id`, farmID).Scan(&idA)
	if err != nil {
		t.Fatalf("insert plant A: %v", err)
	}
	err = testPool.QueryRow(ctx, `
INSERT INTO gr33ncrops.plants (farm_id, display_name, variety_or_cultivar, meta)
VALUES ($1, 'tomato', 'typo B', '{}'::jsonb) RETURNING id`, farmID).Scan(&idB)
	if err != nil {
		t.Fatalf("insert plant B: %v", err)
	}

	var zoneID int64
	err = testPool.QueryRow(ctx, `
INSERT INTO gr33ncore.zones (farm_id, name, description, zone_type)
VALUES ($1, $2, 'phase 103 smoke', 'indoor') RETURNING id`, farmID, tag+" zone").Scan(&zoneID)
	if err != nil {
		t.Fatalf("insert zone: %v", err)
	}

	var cycleA, cycleB int64
	err = testPool.QueryRow(ctx, `
INSERT INTO gr33nfertigation.crop_cycles (farm_id, zone_id, name, batch_label, current_stage, is_active, started_at, plant_id)
VALUES ($1, $2, $3, 'Batch A', 'early_veg', FALSE, CURRENT_DATE, $4) RETURNING id`,
		farmID, zoneID, tag+" A", idA).Scan(&cycleA)
	if err != nil {
		t.Fatalf("insert cycle A: %v", err)
	}
	err = testPool.QueryRow(ctx, `
INSERT INTO gr33nfertigation.crop_cycles (farm_id, zone_id, name, batch_label, current_stage, is_active, started_at, plant_id)
VALUES ($1, $2, $3, 'Batch B', 'early_veg', FALSE, CURRENT_DATE, $4) RETURNING id`,
		farmID, zoneID, tag+" B", idB).Scan(&cycleB)
	if err != nil {
		t.Fatalf("insert cycle B: %v", err)
	}

	t.Cleanup(func() {
		c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, _ = testPool.Exec(c, `DELETE FROM gr33nfertigation.crop_cycles WHERE id IN ($1, $2)`, cycleA, cycleB)
		_, _ = testPool.Exec(c, `DELETE FROM gr33ncore.zones WHERE id = $1`, zoneID)
		_, _ = testPool.Exec(c, `UPDATE gr33ncrops.plants SET deleted_at = NOW() WHERE id IN ($1, $2)`, idA, idB)
	})

	var merged int
	err = testPool.QueryRow(ctx, `
SELECT plants_merged FROM gr33ncrops.merge_legacy_plants()`).Scan(&merged)
	if err != nil {
		t.Fatalf("merge_legacy_plants: %v", err)
	}
	if merged < 1 {
		t.Fatalf("expected at least one duplicate merged, got %d", merged)
	}

	var activeTomato int
	err = testPool.QueryRow(ctx, `
SELECT count(*) FROM gr33ncrops.plants
WHERE farm_id = $1 AND deleted_at IS NULL AND crop_key = 'tomato'`, farmID).Scan(&activeTomato)
	if err != nil {
		t.Fatalf("count tomato plants: %v", err)
	}
	if activeTomato != 1 {
		t.Fatalf("expected exactly one active tomato plant, got %d", activeTomato)
	}

	var keeperID int64
	err = testPool.QueryRow(ctx, `
SELECT id FROM gr33ncrops.plants
WHERE farm_id = $1 AND deleted_at IS NULL AND crop_key = 'tomato'`, farmID).Scan(&keeperID)
	if err != nil {
		t.Fatalf("keeper plant: %v", err)
	}

	for _, cid := range []int64{cycleA, cycleB} {
		var linked int64
		err = testPool.QueryRow(ctx, `
SELECT plant_id FROM gr33nfertigation.crop_cycles WHERE id = $1`, cid).Scan(&linked)
		if err != nil {
			t.Fatalf("cycle %d: %v", cid, err)
		}
		if linked != keeperID {
			t.Fatalf("cycle %d plant_id=%d want keeper %d", cid, linked, keeperID)
		}
	}

	var batchA, batchB string
	_ = testPool.QueryRow(ctx, `SELECT batch_label FROM gr33nfertigation.crop_cycles WHERE id = $1`, cycleA).Scan(&batchA)
	_ = testPool.QueryRow(ctx, `SELECT batch_label FROM gr33nfertigation.crop_cycles WHERE id = $1`, cycleB).Scan(&batchB)
	if batchA != "Batch A" || batchB != "Batch B" {
		t.Fatalf("batch labels changed: %q %q", batchA, batchB)
	}
}

func TestPhase103_AuditNoDuplicateCropKeyAfterMigrate(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var dupGroups int
	err := testPool.QueryRow(ctx, `
SELECT count(*) FROM (
  SELECT farm_id, crop_key FROM gr33ncrops.plants
  WHERE deleted_at IS NULL AND crop_key IS NOT NULL
  GROUP BY farm_id, crop_key HAVING count(*) > 1
) t`).Scan(&dupGroups)
	if err != nil {
		t.Fatalf("dup query: %v", err)
	}
	if dupGroups > 0 {
		t.Fatalf("expected zero duplicate crop_key groups on demo DB, got %d — run ./scripts/merge-legacy-plants.sh --apply", dupGroups)
	}
}
