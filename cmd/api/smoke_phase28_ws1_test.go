// Phase 28 WS1 — crop cycle analytics smoke tests.
//
// Drives GET /crop-cycles/{id}/summary and GET /farms/{id}/crop-cycles/compare
// (+ their .csv variants) against a real Postgres so the fertigation /
// cost rollup SQL is exercised end-to-end. The smoke harness already seeds
// farm 1 + zone(s) in TestMain.

package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"gr33n-api/internal/farmguardian" // pulled in so smoke uses the same module
)

// seedCropCycleForAnalytics creates a crop cycle plus a small fertigation
// history so the summary endpoint has something to roll up. Returns the
// crop_cycle_id.
func seedCropCycleForAnalytics(t *testing.T, tok string, name string, harvestedAt *string, yieldGrams *float64) int64 {
	t.Helper()

	resp := authPost(t, tok, "/farms/1/zones", map[string]any{
		"name":               uniqueName("ws1_zone"),
		"farm_id":            1,
		"floorplan_x":        0,
		"floorplan_y":        0,
		"floorplan_width":    1,
		"floorplan_height":   1,
		"zone_type":          "indoor",
		"description":        "ws1 analytics zone",
		"environment_type":   "soil",
		"crop_type":          "vegetable",
		"current_grow_stage": "seedling",
	})
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusCreated)
	zone := decodeMap(t, resp)
	zoneID := int64(zone["id"].(float64))

	// Active cycles require a catalog plant_id (Phase 86). "lettuce" is part
	// of the Phase 124 demo seed — smokeEnsureCatalogPlant finds-or-creates,
	// never mutating it, so this is safe regardless of seed state.
	plantID := smokeEnsureCatalogPlant(t, tok, 1, "lettuce")
	createBody := map[string]any{
		"zone_id":           zoneID,
		"plant_id":          plantID,
		"name":              name,
		"strain_or_variety": "OG Kush",
		"current_stage":     "early_veg",
		"started_at":        "2026-03-01",
	}
	resp = authPost(t, tok, "/farms/1/crop-cycles", createBody)
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusCreated)
	cycle := decodeMap(t, resp)
	cycleID := int64(cycle["id"].(float64))

	// Backfill three fertigation events with deterministic numbers we can
	// assert against (10 L total, EC 1.0 / 1.5 / 2.0 → avg 1.5 / min 1.0 /
	// max 2.0; pH avg 6.0).
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for _, ev := range []struct {
		liters   float64
		ecAfter  float64
		phBefore float64
		phAfter  float64
	}{
		{2.5, 1.0, 6.1, 5.9},
		{2.5, 1.5, 6.0, 6.0},
		{5.0, 2.0, 6.0, 6.0},
	} {
		if _, err := testPool.Exec(ctx, `
INSERT INTO gr33nfertigation.fertigation_events
    (farm_id, zone_id, crop_cycle_id, applied_at,
     volume_applied_liters, ec_after_mscm, ph_before, ph_after)
VALUES (1, $1, $2, NOW() - INTERVAL '5 days', $3, $4, $5, $6)`,
			zoneID, cycleID, ev.liters, ev.ecAfter, ev.phBefore, ev.phAfter); err != nil {
			t.Fatalf("seed fertigation event: %v", err)
		}
	}

	// One cost row tagged to the cycle so cost_per_gram has something to
	// chew on when yield is provided.
	if _, err := testPool.Exec(ctx, `
INSERT INTO gr33ncore.cost_transactions
    (farm_id, transaction_date, category, amount, currency, is_income,
     description, crop_cycle_id, created_by_user_id)
VALUES (1, CURRENT_DATE - INTERVAL '2 days', 'fertilizers_soil_amendments', 50.00, 'USD', false,
        'ws1 seed cost', $1, $2)`,
		cycleID, uuid.MustParse(smokeDevUserUUID)); err != nil {
		t.Fatalf("seed cost transaction: %v", err)
	}

	// Optionally flip the cycle to harvested + record yield so yield math
	// has a non-nil grams value.
	if harvestedAt != nil && yieldGrams != nil {
		upd := map[string]any{
			"name":              name,
			"strain_or_variety": "OG Kush",
			"zone_id":           zoneID,
			"is_active":         false,
			"harvested_at":      *harvestedAt,
			"yield_grams":       *yieldGrams,
		}
		resp := authPut(t, tok, "/crop-cycles/"+intStr(cycleID), upd)
		defer resp.Body.Close()
		expectStatus(t, resp, http.StatusOK)
	}

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = testPool.Exec(ctx, `DELETE FROM gr33ncore.cost_transactions WHERE crop_cycle_id = $1`, cycleID)
		_, _ = testPool.Exec(ctx, `DELETE FROM gr33nfertigation.fertigation_events WHERE crop_cycle_id = $1`, cycleID)
		_, _ = testPool.Exec(ctx, `DELETE FROM gr33nfertigation.crop_cycles WHERE id = $1`, cycleID)
		_, _ = testPool.Exec(ctx, `DELETE FROM gr33ncore.zones WHERE id = $1`, zoneID)
	})

	// Touch farmguardian package so the import stays alive in case future
	// edits remove the reference. Cheap — this is just a no-op call.
	_ = farmguardian.RAGTopK
	return cycleID
}

func intStr(n int64) string {
	return strings.TrimSpace(formatInt(n))
}

// formatInt is a tiny shim to avoid importing strconv just for tests.
func formatInt(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func TestPhase28_CropCycleSummary_JSON(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	tok := smokeJWT(t)
	harvested := "2026-05-01"
	yield := 200.0
	cycleID := seedCropCycleForAnalytics(t, tok, uniqueName("ws1_cycle"), &harvested, &yield)

	resp := authGet(t, tok, "/crop-cycles/"+intStr(cycleID)+"/summary")
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusOK)
	body := decodeMap(t, resp)

	cycle, ok := body["cycle"].(map[string]any)
	if !ok || int64(cycle["id"].(float64)) != cycleID {
		t.Fatalf("cycle block missing or wrong id: %+v", body["cycle"])
	}
	if body["duration_days"].(float64) <= 0 {
		t.Fatalf("duration_days should be > 0, got %v", body["duration_days"])
	}

	fert := body["fertigation"].(map[string]any)
	if fert["event_count"].(float64) != 3 {
		t.Fatalf("event_count = %v, want 3", fert["event_count"])
	}
	if fert["total_liters"].(float64) != 10.0 {
		t.Fatalf("total_liters = %v, want 10", fert["total_liters"])
	}
	if fert["min_ec_mscm"].(float64) != 1.0 {
		t.Fatalf("min_ec_mscm = %v, want 1.0", fert["min_ec_mscm"])
	}
	if fert["max_ec_mscm"].(float64) != 2.0 {
		t.Fatalf("max_ec_mscm = %v, want 2.0", fert["max_ec_mscm"])
	}

	cost := body["cost"].(map[string]any)
	totals := cost["totals"].([]any)
	if len(totals) != 1 {
		t.Fatalf("expected exactly one currency total, got %d (%+v)", len(totals), totals)
	}
	total := totals[0].(map[string]any)
	if total["currency"].(string) != "USD" {
		t.Fatalf("currency = %v, want USD", total["currency"])
	}
	if total["total_expenses"].(float64) != 50.0 {
		t.Fatalf("total_expenses = %v, want 50", total["total_expenses"])
	}

	yieldBlock := body["yield"].(map[string]any)
	if yieldBlock["grams"].(float64) != 200.0 {
		t.Fatalf("yield.grams = %v, want 200", yieldBlock["grams"])
	}
	cpg, ok := yieldBlock["cost_per_gram"].(float64)
	if !ok || cpg <= 0 {
		t.Fatalf("cost_per_gram should be set + > 0, got %v", yieldBlock["cost_per_gram"])
	}

	stages := body["stages"].([]any)
	if len(stages) < 1 {
		t.Fatalf("stages length = %d, want at least one timeline row", len(stages))
	}
	if !body["stage_history_supported"].(bool) {
		t.Fatalf("stage_history_supported must be true when crop_cycle_stage_events has rows (Phase 56)")
	}
}

func TestPhase28_CropCycleSummary_NotFound404(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/crop-cycles/99999999/summary")
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusNotFound)
}

func TestPhase28_CropCycleSummary_CSV(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	tok := smokeJWT(t)
	harvested := "2026-04-15"
	yield := 100.0
	cycleID := seedCropCycleForAnalytics(t, tok, uniqueName("ws1_csv"), &harvested, &yield)

	resp := authGet(t, tok, "/crop-cycles/"+intStr(cycleID)+"/summary.csv")
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusOK)
	if !strings.HasPrefix(resp.Header.Get("Content-Type"), "text/csv") {
		t.Fatalf("Content-Type = %q, want text/csv*", resp.Header.Get("Content-Type"))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(body)), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected header + 1 data row, got %d lines:\n%s", len(lines), string(body))
	}
	header := lines[0]
	for _, want := range []string{"cycle_id", "total_liters", "min_ec_mscm", "max_ec_mscm", "yield_grams", "cost_per_gram", "currency"} {
		if !strings.Contains(header, want) {
			t.Fatalf("CSV header missing %q: %s", want, header)
		}
	}
	if !strings.Contains(lines[1], "USD") {
		t.Fatalf("CSV row should include the USD currency, got %s", lines[1])
	}
}

func TestPhase28_CropCycleCompare_JSON(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	tok := smokeJWT(t)
	yield1 := 150.0
	harvested1 := "2026-04-20"
	id1 := seedCropCycleForAnalytics(t, tok, uniqueName("ws1_cmp_a"), &harvested1, &yield1)

	yield2 := 220.0
	harvested2 := "2026-05-05"
	id2 := seedCropCycleForAnalytics(t, tok, uniqueName("ws1_cmp_b"), &harvested2, &yield2)

	url := "/farms/1/crop-cycles/compare?ids=" + intStr(id1) + "," + intStr(id2)
	resp := authGet(t, tok, url)
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusOK)
	body := decodeMap(t, resp)
	cycles, ok := body["cycles"].([]any)
	if !ok || len(cycles) != 2 {
		t.Fatalf("expected 2 cycles, got %d (%+v)", len(cycles), body)
	}
	for _, c := range cycles {
		cc := c.(map[string]any)["cycle"].(map[string]any)
		gotID := int64(cc["id"].(float64))
		if gotID != id1 && gotID != id2 {
			t.Fatalf("compare returned unexpected cycle id %d", gotID)
		}
	}
}

func TestPhase28_CropCycleCompare_RejectsForeignFarm(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	tok := smokeJWT(t)
	id := seedCropCycleForAnalytics(t, tok, uniqueName("ws1_foreign"), nil, nil)
	resp := authGet(t, tok, "/farms/99999/crop-cycles/compare?ids="+intStr(id))
	defer resp.Body.Close()
	// Either the user is not a member of the foreign farm (403) or the
	// farm doesn't exist (404). Both are acceptable rejections; what we
	// care about is "this cycle does not get leaked across the farm
	// boundary".
	if resp.StatusCode == http.StatusOK {
		var raw map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&raw)
		t.Fatalf("compare leaked a cycle from a foreign farm: %+v", raw)
	}
}

func TestPhase28_CropCycleCompare_TooMany(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/farms/1/crop-cycles/compare?ids=1,2,3,4,5,6")
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusBadRequest)
}

func TestPhase28_CropCycleCompare_MissingIDs(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/farms/1/crop-cycles/compare")
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusBadRequest)
}
