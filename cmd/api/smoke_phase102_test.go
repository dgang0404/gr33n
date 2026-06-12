// Phase 102 — fertigation program catalog metadata contract smoke.
package main

import (
	"fmt"
	"net/http"
	"testing"
)

func TestPhase102_ProgramMetadataAndFit(t *testing.T) {
	tok := smokeJWT(t)

	resp := authGet(t, tok, "/farms/1/fertigation/programs")
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusOK)
	programs := decodeSlice(t, resp)

	var vegID int64
	var vegMeta map[string]any
	for _, raw := range programs {
		row, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		name, _ := row["name"].(string)
		if name == "Veg Daily JLF Program" {
			vegID = int64(row["id"].(float64))
			vegMeta, _ = row["metadata"].(map[string]any)
		}
	}
	if vegID == 0 {
		t.Fatal("Veg Daily JLF Program not found — run phase 102 migration")
	}
	stages, _ := vegMeta["recommended_stages"].([]any)
	if len(stages) < 2 {
		t.Fatalf("veg program metadata: %#v", vegMeta)
	}

	resp = authGet(t, tok, "/farms/1/fertigation/programs?crop_key=cannabis&stage=early_flower")
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusOK)
	filtered := decodeSlice(t, resp)
	for _, raw := range filtered {
		row, _ := raw.(map[string]any)
		if row["name"] == "Veg Daily JLF Program" {
			t.Fatal("veg program should not match early_flower filter")
		}
	}

	zoneID := seedSetpointZone(t)
	plantID := smokeEnsureCatalogPlant(t, tok, 1, "cannabis")

	resp = authPost(t, tok, "/farms/1/crop-cycles", map[string]any{
		"zone_id":            zoneID,
		"name":               fmt.Sprintf("Phase102 fit %d", vegID),
		"current_stage":      "early_flower",
		"started_at":         "2026-06-01",
		"plant_id":           plantID,
		"primary_program_id": vegID,
		"is_active":          false,
	})
	expectStatus(t, resp, http.StatusCreated)
	body := decodeMap(t, resp)
	warnings, _ := body["program_fit_warnings"].([]any)
	if len(warnings) == 0 {
		t.Fatalf("expected program_fit_warnings, got %#v", body)
	}
	cycleID := int64(body["id"].(float64))

	resp = authDelete(t, tok, fmt.Sprintf("/crop-cycles/%d", cycleID))
	expectStatus(t, resp, http.StatusNoContent)
}

func smokeEnsureCatalogPlant(t *testing.T, tok string, farmID int64, cropKey string) int64 {
	t.Helper()
	resp := authGet(t, tok, fmt.Sprintf("/farms/%d/plants", farmID))
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusOK)
	for _, raw := range decodeSlice(t, resp) {
		row, _ := raw.(map[string]any)
		if row["crop_key"] == cropKey {
			return int64(row["id"].(float64))
		}
	}
	resp = authPost(t, tok, fmt.Sprintf("/farms/%d/plants", farmID), map[string]any{
		"crop_key": cropKey,
	})
	expectStatus(t, resp, http.StatusCreated)
	return int64(decodeMap(t, resp)["id"].(float64))
}
