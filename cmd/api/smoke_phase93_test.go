// Phase 93 — batch_label vocabulary: API primary field + strain_or_variety alias.

package main

import (
	"context"
	"fmt"
	"net/http"
	"testing"
)

func TestPhase93_CycleBatchLabelPrimaryAndAlias(t *testing.T) {
	tok := smokeJWT(t)
	zoneID, restore := smokeZoneWithoutActiveCycle(t)
	defer restore()

	// "kale" — a crop_key the Phase 124 demo seed doesn't touch — so this
	// stays a clean 201 regardless of what's already seeded on farm 1.
	resp := authPost(t, tok, "/farms/1/plants", map[string]any{
		"crop_key":            "kale",
		"variety_or_cultivar": "Cherokee Purple",
	})
	expectStatus(t, resp, http.StatusCreated)
	plant := decodeMap(t, resp)
	plantID := int64(plant["id"].(float64))
	t.Cleanup(func() {
		_, _ = testPool.Exec(context.Background(),
			`UPDATE gr33ncrops.plants SET deleted_at = NOW() WHERE farm_id = 1 AND crop_key = 'kale'`)
	})

	resp = authPost(t, tok, "/farms/1/crop-cycles", map[string]any{
		"zone_id":       zoneID,
		"plant_id":      plantID,
		"name":          uniqueName("phase93_batch"),
		"batch_label":   "Row A",
		"current_stage": "early_veg",
		"started_at":    "2026-06-01",
		"is_active":     true,
	})
	expectStatus(t, resp, http.StatusCreated)
	created := decodeMap(t, resp)
	if created["batch_label"] != "Row A" {
		t.Fatalf("expected batch_label Row A, got %v", created["batch_label"])
	}
	if created["strain_or_variety"] != "Row A" {
		t.Fatalf("expected strain_or_variety alias Row A, got %v", created["strain_or_variety"])
	}
	cycleID := int64(created["id"].(float64))

	resp = authGet(t, tok, fmt.Sprintf("/crop-cycles/%d", cycleID))
	expectStatus(t, resp, http.StatusOK)
	got := decodeMap(t, resp)
	if got["batch_label"] != "Row A" {
		t.Fatalf("GET expected batch_label, got %v", got["batch_label"])
	}

	resp = authPut(t, tok, fmt.Sprintf("/crop-cycles/%d", cycleID), map[string]any{
		"name":      created["name"],
		"zone_id":   zoneID,
		"is_active": true,
		"batch_label": "Row B",
		"plant_id":  plantID,
	})
	expectStatus(t, resp, http.StatusOK)
	updated := decodeMap(t, resp)
	if updated["batch_label"] != "Row B" {
		t.Fatalf("update batch_label want Row B got %v", updated["batch_label"])
	}
}

func TestPhase93_CycleStrainOrVarietyWriteAlias(t *testing.T) {
	tok := smokeJWT(t)
	zoneID, restore := smokeZoneWithoutActiveCycle(t)
	defer restore()

	// "spinach" — a crop_key the Phase 124 demo seed doesn't touch.
	resp := authPost(t, tok, "/farms/1/plants", map[string]any{"crop_key": "spinach"})
	expectStatus(t, resp, http.StatusCreated)
	plantID := int64(decodeMap(t, resp)["id"].(float64))
	t.Cleanup(func() {
		_, _ = testPool.Exec(context.Background(),
			`UPDATE gr33ncrops.plants SET deleted_at = NOW() WHERE farm_id = 1 AND crop_key = 'spinach'`)
	})

	resp = authPost(t, tok, "/farms/1/crop-cycles", map[string]any{
		"zone_id":           zoneID,
		"plant_id":          plantID,
		"name":              uniqueName("phase93_alias"),
		"strain_or_variety": "Legacy tag",
		"current_stage":     "seedling",
		"started_at":        "2026-06-02",
		"is_active":         true,
	})
	expectStatus(t, resp, http.StatusCreated)
	cycle := decodeMap(t, resp)
	if cycle["batch_label"] != "Legacy tag" {
		t.Fatalf("alias write expected batch_label Legacy tag, got %v", cycle["batch_label"])
	}
}
