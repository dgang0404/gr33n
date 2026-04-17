// Phase 20.8 smoke coverage — animal_groups CRUD, lifecycle event
// timeline, aquaponics loops, feed-consumption-through-autologger
// (animal_feed input → feed_livestock cost row), and bootstrap
// idempotency for chicken_coop_v1 + small_aquaponics_v1.

package main

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"
)

// seedZone creates a fresh zone for a test so we don't conflict with
// seeded / bootstrap zones. Returns the new zone id.
func seedZoneForAnimal(t *testing.T, tok string) int64 {
	t.Helper()
	resp := authPost(t, tok, "/farms/1/zones", map[string]any{
		"name":        uniqueName("ph208_zone"),
		"description": "ph208 test zone",
		"zone_type":   "indoor",
	})
	expectStatus(t, resp, http.StatusCreated)
	return int64(decodeMap(t, resp)["id"].(float64))
}

// seedFeedInputDefinition is the Phase 20.8 analogue of
// seedPricedInputDefinition — same shape, but the category is
// `animal_feed` so the autologger maps it to the `feed_livestock`
// cost category (WS3).
func seedFeedInputDefinition(t *testing.T, tok string, unitCost float64) int64 {
	t.Helper()
	resp := authPost(t, tok, "/farms/1/naturalfarming/inputs", map[string]any{
		"name":               uniqueName("ph208_feed"),
		"category":           "animal_feed",
		"unit_cost":          unitCost,
		"unit_cost_currency": "USD",
		"unit_cost_unit_id":  1,
	})
	expectStatus(t, resp, http.StatusCreated)
	return int64(decodeMap(t, resp)["id"].(float64))
}

// ── WS2: animal_groups + lifecycle CRUD ─────────────────────────────────────

func TestPhase208AnimalGroupCRUDAndTimeline(t *testing.T) {
	tok := smokeJWT(t)
	zoneID := seedZoneForAnimal(t, tok)

	// Create
	resp := authPost(t, tok, "/farms/1/animal-groups", map[string]any{
		"label":           uniqueName("Layer flock"),
		"species":         "chicken",
		"count":           12,
		"primary_zone_id": zoneID,
	})
	expectStatus(t, resp, http.StatusCreated)
	group := decodeMap(t, resp)
	groupID := int64(group["id"].(float64))
	if species := group["species"].(string); species != "chicken" {
		t.Fatalf("expected species=chicken, got %q", species)
	}
	if count := int(group["count"].(float64)); count != 12 {
		t.Fatalf("expected count=12, got %d", count)
	}
	if active := group["active"].(bool); !active {
		t.Fatal("expected new group to be active")
	}

	// Cross-farm zone validation — a zone from another farm should 400.
	resp = authPost(t, tok, "/farms/1/animal-groups", map[string]any{
		"label":           uniqueName("bad_zone"),
		"primary_zone_id": 999999, // non-existent
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing zone, got %d", resp.StatusCode)
	}

	// Record lifecycle events with signed deltas.
	for _, ev := range []map[string]any{
		{"event_type": "added", "delta_count": 12, "notes": "initial stocking"},
		{"event_type": "born", "delta_count": 3, "notes": "hatch"},
		{"event_type": "died", "delta_count": -1, "notes": "injury"},
		{"event_type": "note", "notes": "vet visit scheduled"},
	} {
		resp = authPost(t, tok, fmt.Sprintf("/animal-groups/%d/lifecycle-events", groupID), ev)
		expectStatus(t, resp, http.StatusCreated)
	}

	// List — ordered by event_time DESC.
	resp = authGet(t, tok, fmt.Sprintf("/animal-groups/%d/lifecycle-events", groupID))
	expectStatus(t, resp, http.StatusOK)
	events := decodeSlice(t, resp)
	if len(events) != 4 {
		t.Fatalf("expected 4 lifecycle events, got %d", len(events))
	}

	// Detail endpoint returns group + delta_total (12 + 3 - 1 = 14).
	resp = authGet(t, tok, fmt.Sprintf("/animal-groups/%d", groupID))
	expectStatus(t, resp, http.StatusOK)
	detail := decodeMap(t, resp)
	if delta := int(detail["delta_total"].(float64)); delta != 14 {
		t.Fatalf("expected delta_total=14, got %d", delta)
	}

	// Archive — group stays visible but active=false.
	resp = authPatch(t, tok, fmt.Sprintf("/animal-groups/%d/archive", groupID), map[string]any{
		"archived_reason": "sold to neighbour",
	})
	expectStatus(t, resp, http.StatusOK)
	archived := decodeMap(t, resp)
	if active := archived["active"].(bool); active {
		t.Fatal("expected active=false after archive")
	}
	if reason, _ := archived["archived_reason"].(string); reason != "sold to neighbour" {
		t.Fatalf("expected archived_reason to round-trip, got %q", reason)
	}

	// Update — relabel + clear count.
	resp = authPut(t, tok, fmt.Sprintf("/animal-groups/%d", groupID), map[string]any{
		"label":           "Archived flock",
		"primary_zone_id": zoneID,
	})
	expectStatus(t, resp, http.StatusOK)

	// Soft delete.
	resp = authDelete(t, tok, fmt.Sprintf("/animal-groups/%d", groupID))
	expectStatus(t, resp, http.StatusNoContent)

	// Get 404s after soft delete.
	resp = authGet(t, tok, fmt.Sprintf("/animal-groups/%d", groupID))
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 after soft delete, got %d", resp.StatusCode)
	}
}

// ── WS2: aquaponics loops CRUD ──────────────────────────────────────────────

func TestPhase208AquaponicsLoopCRUD(t *testing.T) {
	tok := smokeJWT(t)
	tank := seedZoneForAnimal(t, tok)
	bed := seedZoneForAnimal(t, tok)

	resp := authPost(t, tok, "/farms/1/aquaponics-loops", map[string]any{
		"label":             uniqueName("ph208_loop"),
		"fish_tank_zone_id": tank,
		"grow_bed_zone_id":  bed,
	})
	expectStatus(t, resp, http.StatusCreated)
	loop := decodeMap(t, resp)
	loopID := int64(loop["id"].(float64))
	if fid := int64(loop["fish_tank_zone_id"].(float64)); fid != tank {
		t.Fatalf("expected fish_tank_zone_id=%d, got %d", tank, fid)
	}

	// Update toggles active off + reassigns.
	newBed := seedZoneForAnimal(t, tok)
	resp = authPut(t, tok, fmt.Sprintf("/aquaponics-loops/%d", loopID), map[string]any{
		"label":             "renamed loop",
		"fish_tank_zone_id": tank,
		"grow_bed_zone_id":  newBed,
		"active":            false,
	})
	expectStatus(t, resp, http.StatusOK)
	updated := decodeMap(t, resp)
	if updated["active"].(bool) {
		t.Fatal("expected active=false after update")
	}
	if int64(updated["grow_bed_zone_id"].(float64)) != newBed {
		t.Fatal("expected grow_bed_zone_id to update")
	}

	// Soft delete.
	resp = authDelete(t, tok, fmt.Sprintf("/aquaponics-loops/%d", loopID))
	expectStatus(t, resp, http.StatusNoContent)
}

// ── WS3: animal_feed input → feed_livestock cost category ───────────────────

// TestPhase208FeedConsumptionAutologgerCategory is the end-to-end
// check that WS3's category mapping actually stamps the right
// cost_category on the auto-logged row. Pathway:
//   animal_feed input  →  task_input_consumption  →  cost row
// and we assert category = 'feed_livestock'.
func TestPhase208FeedConsumptionAutologgerCategory(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	tok := smokeJWT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	defID := seedFeedInputDefinition(t, tok, 2.0) // $2 / unit of feed
	batchID := seedBatchWithStock(t, tok, defID, 100.0, nil)

	// Task to attach the consumption to.
	resp := authPost(t, tok, "/farms/1/tasks", map[string]any{
		"title": uniqueName("ph208_feed_task"),
	})
	expectStatus(t, resp, http.StatusCreated)
	taskID := int64(decodeMap(t, resp)["id"].(float64))

	// Consumption: 5 units × $2 = $10 feed cost.
	resp = authPost(t, tok, fmt.Sprintf("/tasks/%d/consumptions", taskID), map[string]any{
		"input_batch_id": batchID,
		"quantity":       5.0,
		"unit_id":        1,
		"notes":          "morning feed — layer flock",
	})
	expectStatus(t, resp, http.StatusCreated)
	cons := decodeMap(t, resp)
	consID := int64(cons["id"].(float64))

	var category string
	var amount float64
	if err := testPool.QueryRow(ctx, `
		SELECT category::text, amount::float8
		FROM gr33ncore.cost_transactions
		WHERE related_module_schema = 'gr33ncore'
		  AND related_table_name = 'task_input_consumptions'
		  AND related_record_id = $1`, consID,
	).Scan(&category, &amount); err != nil {
		t.Fatalf("fetch feed cost row: %v", err)
	}
	if category != "feed_livestock" {
		t.Fatalf("expected category=feed_livestock, got %q", category)
	}
	if amount != 10.0 {
		t.Fatalf("expected amount=10.0, got %.2f", amount)
	}
}

// ── WS5: bootstrap idempotency ──────────────────────────────────────────────

// TestPhase208ChickenCoopBootstrapSeedsAnimalGroup asserts the
// patched chicken_coop_v1 bootstrap seeds an animal_groups row and
// that a second apply is a no-op (dispatcher short-circuits on
// farm_bootstrap_applications). Also validates the internal inner
// function is idempotent — we run it directly a second time against
// the same farm and confirm the groups row count does not increase.
func TestPhase208ChickenCoopBootstrapSeedsAnimalGroup(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	tok := smokeJWT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Isolate on a brand-new farm whose bootstrap is chicken_coop_v1
	// — the endpoint applies it transactionally on create.
	resp := authPost(t, tok, "/farms", map[string]any{
		"name":               uniqueName("ph208_coop_farm"),
		"owner_user_id":      smokeDevUserUUID,
		"timezone":           "UTC",
		"currency":           "USD",
		"operational_status": "active",
		"scale_tier":         "small",
		"bootstrap_template": "chicken_coop_v1",
	})
	expectStatus(t, resp, http.StatusCreated)
	payload := decodeMap(t, resp)
	farmObj := payload["farm"].(map[string]any)
	farmID := int64(farmObj["id"].(float64))

	// Exactly one Layer flock row exists for this farm.
	countGroups := func() int {
		t.Helper()
		var n int
		if err := testPool.QueryRow(ctx, `
			SELECT COUNT(*) FROM gr33nanimals.animal_groups
			WHERE farm_id = $1 AND label = 'Layer flock' AND deleted_at IS NULL`,
			farmID,
		).Scan(&n); err != nil {
			t.Fatalf("count animal_groups: %v", err)
		}
		return n
	}
	if n := countGroups(); n != 1 {
		t.Fatalf("expected 1 Layer flock row after bootstrap, got %d", n)
	}

	// Second apply via dispatcher: short-circuits at farm_bootstrap_applications.
	var appliedJSON string
	if err := testPool.QueryRow(ctx,
		`SELECT gr33ncore.apply_farm_bootstrap_template($1, 'chicken_coop_v1')::text`,
		farmID,
	).Scan(&appliedJSON); err != nil {
		t.Fatalf("second apply: %v", err)
	}
	if n := countGroups(); n != 1 {
		t.Fatalf("expected still 1 Layer flock row after re-dispatch, got %d", n)
	}

	// Call the inner function directly — this bypasses the outer guard
	// and proves the NOT EXISTS guard inside the function also holds.
	if _, err := testPool.Exec(ctx,
		`SELECT gr33ncore._bootstrap_chicken_coop_v1($1, 'UTC')`, farmID,
	); err != nil {
		t.Fatalf("direct inner call: %v", err)
	}
	if n := countGroups(); n != 1 {
		t.Fatalf("expected still 1 Layer flock row after direct inner re-run, got %d", n)
	}

	// The seeded group has the expected shape.
	var species string
	var count int
	var primaryZoneID *int64
	if err := testPool.QueryRow(ctx, `
		SELECT species, count, primary_zone_id
		FROM gr33nanimals.animal_groups
		WHERE farm_id = $1 AND label = 'Layer flock' AND deleted_at IS NULL`,
		farmID,
	).Scan(&species, &count, &primaryZoneID); err != nil {
		t.Fatalf("inspect seeded group: %v", err)
	}
	if species != "chicken" {
		t.Fatalf("expected species=chicken, got %q", species)
	}
	if count != 12 {
		t.Fatalf("expected count=12, got %d", count)
	}
	if primaryZoneID == nil {
		t.Fatal("expected primary_zone_id to be populated (coop zone)")
	}
}

// TestPhase208AquaponicsBootstrapSetsTypedFKs asserts the patched
// small_aquaponics_v1 bootstrap writes the loops row with
// fish_tank_zone_id + grow_bed_zone_id populated, and that the
// UPDATE …COALESCE() back-patch works for a loop row that was
// seeded by the old Phase 20.5 bootstrap (nulled FKs).
func TestPhase208AquaponicsBootstrapSetsTypedFKs(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	tok := smokeJWT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	resp := authPost(t, tok, "/farms", map[string]any{
		"name":               uniqueName("ph208_aqua_farm"),
		"owner_user_id":      smokeDevUserUUID,
		"timezone":           "UTC",
		"currency":           "USD",
		"operational_status": "active",
		"scale_tier":         "small",
		"bootstrap_template": "small_aquaponics_v1",
	})
	expectStatus(t, resp, http.StatusCreated)
	payload := decodeMap(t, resp)
	farmObj := payload["farm"].(map[string]any)
	farmID := int64(farmObj["id"].(float64))

	var fishZoneID, bedZoneID *int64
	if err := testPool.QueryRow(ctx, `
		SELECT fish_tank_zone_id, grow_bed_zone_id
		FROM gr33naquaponics.loops
		WHERE farm_id = $1 AND label = 'Main aquaponics loop' AND deleted_at IS NULL`,
		farmID,
	).Scan(&fishZoneID, &bedZoneID); err != nil {
		t.Fatalf("inspect loops: %v", err)
	}
	if fishZoneID == nil || bedZoneID == nil {
		t.Fatal("expected both typed FKs to be populated after bootstrap")
	}

	// Simulate a legacy row (pre-patch shape) — null the FKs, then
	// re-run the inner function to prove the back-patch path.
	if _, err := testPool.Exec(ctx, `
		UPDATE gr33naquaponics.loops SET fish_tank_zone_id = NULL, grow_bed_zone_id = NULL
		WHERE farm_id = $1`, farmID,
	); err != nil {
		t.Fatalf("null FKs for back-patch test: %v", err)
	}
	if _, err := testPool.Exec(ctx,
		`SELECT gr33ncore._bootstrap_small_aquaponics_v1($1, 'UTC')`, farmID,
	); err != nil {
		t.Fatalf("direct inner call for back-patch: %v", err)
	}
	var fishAfter, bedAfter *int64
	if err := testPool.QueryRow(ctx, `
		SELECT fish_tank_zone_id, grow_bed_zone_id
		FROM gr33naquaponics.loops
		WHERE farm_id = $1 AND label = 'Main aquaponics loop' AND deleted_at IS NULL`,
		farmID,
	).Scan(&fishAfter, &bedAfter); err != nil {
		t.Fatalf("read after back-patch: %v", err)
	}
	if fishAfter == nil || bedAfter == nil {
		t.Fatal("expected back-patch to re-populate FKs")
	}
}
