// Phase 96 — grow feeding program validation smoke (Phase 102 metadata path).
package main

import (
	"fmt"
	"net/http"
	"testing"
)

func TestPhase96_ProgramFitWarningOnCropCycleAttach(t *testing.T) {
	tok := smokeJWT(t)

	resp := authGet(t, tok, "/farms/1/fertigation/programs")
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusOK)
	programs := decodeSlice(t, resp)

	var vegID int64
	for _, raw := range programs {
		row, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if name, _ := row["name"].(string); name == "Veg Daily JLF Program" {
			vegID = int64(row["id"].(float64))
		}
	}
	if vegID == 0 {
		t.Fatal("Veg Daily JLF Program not found — run phase 102 migration")
	}

	zoneID := seedSetpointZone(t)
	plantID := smokeEnsureCatalogPlant(t, tok, 1, "cannabis")

	resp = authPost(t, tok, "/farms/1/crop-cycles", map[string]any{
		"zone_id":            zoneID,
		"name":               fmt.Sprintf("Phase96 fit %d", vegID),
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
		t.Fatalf("expected program_fit_warnings on attach, got %#v", body)
	}
	cycleID := int64(body["id"].(float64))

	resp = authDelete(t, tok, fmt.Sprintf("/crop-cycles/%d", cycleID))
	expectStatus(t, resp, http.StatusNoContent)
}
