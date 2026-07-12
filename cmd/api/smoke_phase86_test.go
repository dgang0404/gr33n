// Phase 86 — grow ops catalog chain: active cycle requires plant_id; Guardian EC parity.

package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"gr33n-api/internal/farmguardian"
	db "gr33n-api/internal/db"
)

func TestPhase86_ActiveCycleRequiresCatalogPlant(t *testing.T) {
	tok := smokeJWT(t)
	zoneID, restore := smokeZoneWithoutActiveCycle(t)
	defer restore()

	resp := authPost(t, tok, "/farms/1/crop-cycles", map[string]any{
		"zone_id":       zoneID,
		"name":          uniqueName("phase86_no_plant"),
		"current_stage": "early_flower",
		"started_at":    "2026-06-01",
		"is_active":     true,
	})
	expectStatus(t, resp, http.StatusBadRequest)
	body := decodeMap(t, resp)
	if !strings.Contains(fmt.Sprint(body["error"]), "plant_id") {
		t.Fatalf("expected plant_id error, got %#v", body)
	}
}

func TestPhase86_GuardianECMatchesCropProfileStage(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	tok := smokeJWT(t)
	zoneID, restore := smokeZoneWithoutActiveCycle(t)
	defer restore()

	// Catalog-bound cannabis plant for built-in crop-profile EC lookup below.
	// Demo seed (Phase 164) uses chrysanthemum on farm 1 — this test creates its
	// own temporary cannabis plant and only deletes it when this test created it.
	resp := authPost(t, tok, "/farms/1/plants", map[string]any{
		"crop_key":            "cannabis",
		"variety_or_cultivar": "Phase86 smoke variety",
	})
	expectStatusOneOf(t, resp, http.StatusCreated, http.StatusOK)
	plantCreated := resp.StatusCode == http.StatusCreated
	plant := decodeMap(t, resp)
	plantID := int64(plant["id"].(float64))

	cycleName := uniqueName("phase86_flower")
	resp = authPost(t, tok, "/farms/1/crop-cycles", map[string]any{
		"zone_id":           zoneID,
		"plant_id":          plantID,
		"name":              cycleName,
		"batch_label":       "Batch A",
		"current_stage":     "early_flower",
		"started_at":        "2026-06-01",
		"is_active":         true,
	})
	expectStatus(t, resp, http.StatusCreated)
	cycle := decodeMap(t, resp)
	cycleID := int64(cycle["id"].(float64))
	if cycle["batch_label"] != "Batch A" {
		t.Fatalf("expected batch_label Batch A, got %v", cycle["batch_label"])
	}
	if cycle["strain_or_variety"] != "Batch A" {
		t.Fatalf("expected strain_or_variety alias Batch A, got %v", cycle["strain_or_variety"])
	}

	var ecMin, ecMax float64
	err := testPool.QueryRow(ctx, `
SELECT s.ec_min::float8, s.ec_max::float8
FROM gr33ncrops.crop_profiles p
JOIN gr33ncrops.crop_profile_stages s ON s.crop_profile_id = p.id
WHERE p.is_builtin = TRUE AND p.crop_key = 'cannabis' AND s.stage = 'early_flower'
LIMIT 1`).Scan(&ecMin, &ecMax)
	if err != nil {
		t.Fatalf("cannabis early_flower seed: %v", err)
	}

	q := db.New(testPool)
	block, err := farmguardian.LookupCropTargets(ctx, q, 1, "Is my EC on target for early flower?", &farmguardian.ContextRef{
		CropCycleID: cycleID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(block, "lookup_crop_targets") {
		t.Fatalf("unexpected block: %s", block)
	}
	if !strings.Contains(block, "mS/cm") {
		t.Fatalf("expected mS/cm EC in block: %s", block)
	}
	ecSnippet := fmt.Sprintf("%.0f", ecMin)
	if !strings.Contains(block, ecSnippet) {
		t.Fatalf("Guardian EC %s not in block (DB min %.2f): %s", ecSnippet, ecMin, block)
	}

	// Unsupported crop mention
	block, err = farmguardian.LookupCropTargets(ctx, q, 1, "What EC for ramps?", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(block, "not supported") {
		t.Fatalf("expected unsupported ramps block: %s", block)
	}
	if strings.Contains(block, "EC target:") {
		t.Fatalf("ramps must not include EC targets: %s", block)
	}

	// Cleanup — deactivate cycle; restore() reactivates any borrowed zone cycle
	resp = authPut(t, tok, fmt.Sprintf("/crop-cycles/%d", cycleID), map[string]any{
		"name":      cycleName,
		"zone_id":   zoneID,
		"is_active": false,
		"plant_id":  plantID,
	})
	expectStatus(t, resp, http.StatusOK)
	if plantCreated {
		resp = authDelete(t, tok, fmt.Sprintf("/plants/%d", plantID))
		expectStatus(t, resp, http.StatusNoContent)
	}
}

func smokeZoneWithoutActiveCycle(t *testing.T) (zoneID int64, restore func()) {
	t.Helper()
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx := context.Background()
	var id int64
	err := testPool.QueryRow(ctx, `
SELECT z.id
FROM gr33ncore.zones z
WHERE z.farm_id = 1 AND z.deleted_at IS NULL
ORDER BY z.id
LIMIT 1`).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}

	var activeCycleID *int64
	_ = testPool.QueryRow(ctx, `
SELECT id FROM gr33nfertigation.crop_cycles
WHERE zone_id = $1 AND is_active = TRUE
LIMIT 1`, id).Scan(&activeCycleID)

	if activeCycleID == nil {
		return id, func() {}
	}

	cycleID := *activeCycleID
	if _, err := testPool.Exec(ctx, `
UPDATE gr33nfertigation.crop_cycles SET is_active = FALSE WHERE id = $1`, cycleID); err != nil {
		t.Fatal(err)
	}
	return id, func() {
		_, _ = testPool.Exec(context.Background(), `
UPDATE gr33nfertigation.crop_cycles SET is_active = TRUE WHERE id = $1`, cycleID)
	}
}
