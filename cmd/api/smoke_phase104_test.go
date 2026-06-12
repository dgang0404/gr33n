// Phase 104 — harvest analytics grouped by catalog crop_key.
package main

import (
	"fmt"
	"net/http"
	"testing"
)

func TestPhase104_CompareAndAnalyticsByCropKey(t *testing.T) {
	tok := smokeJWT(t)
	plantID := smokeEnsureCatalogPlant(t, tok, 1, "cannabis")

	zoneA := seedSetpointZone(t)
	zoneB := seedSetpointZone(t)

	harvested := "2026-05-01"
	yield1 := 180.0
	resp := authPost(t, tok, "/farms/1/crop-cycles", map[string]any{
		"zone_id":       zoneA,
		"name":          fmt.Sprintf("Phase104 cannabis A %d", zoneA),
		"current_stage": "late_flower",
		"started_at":    "2026-03-01",
		"plant_id":      plantID,
		"is_active":     false,
		"harvested_at":  harvested,
		"yield_grams":   yield1,
	})
	expectStatus(t, resp, http.StatusCreated)
	cycleA := decodeMap(t, resp)
	idA := int64(cycleA["id"].(float64))

	harvested2 := "2026-06-01"
	yield2 := 210.0
	resp = authPost(t, tok, "/farms/1/crop-cycles", map[string]any{
		"zone_id":       zoneB,
		"name":          fmt.Sprintf("Phase104 cannabis B %d", zoneB),
		"current_stage": "late_flower",
		"started_at":    "2026-04-01",
		"plant_id":      plantID,
		"is_active":     false,
		"harvested_at":  harvested2,
		"yield_grams":   yield2,
	})
	expectStatus(t, resp, http.StatusCreated)
	cycleB := decodeMap(t, resp)
	idB := int64(cycleB["id"].(float64))

	if cycleA["crop_key"] != "cannabis" || cycleB["crop_key"] != "cannabis" {
		t.Fatalf("list create should expose crop_key: A=%#v B=%#v", cycleA["crop_key"], cycleB["crop_key"])
	}

	resp = authGet(t, tok, fmt.Sprintf("/farms/1/crop-cycles/compare?ids=%d,%d", idA, idB))
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusOK)
	compare := decodeMap(t, resp)
	cycles, _ := compare["cycles"].([]any)
	if len(cycles) != 2 {
		t.Fatalf("expected 2 compare summaries, got %#v", compare)
	}
	for _, raw := range cycles {
		row, _ := raw.(map[string]any)
		if row["crop_key"] != "cannabis" {
			t.Fatalf("compare summary missing crop_key=cannabis: %#v", row)
		}
		if row["catalog_display_name"] == nil || row["catalog_display_name"] == "" {
			t.Fatalf("compare summary missing catalog_display_name: %#v", row)
		}
	}

	resp = authGet(t, tok, "/farms/1/crop-analytics")
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusOK)
	rollup := decodeMap(t, resp)
	crops, _ := rollup["crops"].([]any)
	var cannabisRow map[string]any
	for _, raw := range crops {
		row, _ := raw.(map[string]any)
		if row["crop_key"] == "cannabis" {
			cannabisRow = row
			break
		}
	}
	if cannabisRow == nil {
		t.Fatalf("crop-analytics missing cannabis bucket: %#v", rollup)
	}
	if int64(cannabisRow["cycle_count"].(float64)) < 2 {
		t.Fatalf("expected at least 2 cannabis cycles in rollup: %#v", cannabisRow)
	}

	resp = authGet(t, tok, "/farms/1/crop-cycles?crop_key=cannabis")
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusOK)
	filtered := decodeSlice(t, resp)
	if len(filtered) < 2 {
		t.Fatalf("crop_key filter expected >=2 cycles, got %d", len(filtered))
	}

	resp = authDelete(t, tok, fmt.Sprintf("/crop-cycles/%d", idA))
	expectStatus(t, resp, http.StatusNoContent)
	resp = authDelete(t, tok, fmt.Sprintf("/crop-cycles/%d", idB))
	expectStatus(t, resp, http.StatusNoContent)
}
