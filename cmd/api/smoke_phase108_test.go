// Phase 108 — commons recipe pack crop_key metadata on import.
package main

import (
	"fmt"
	"net/http"
	"testing"
)

func TestPhase108_RecipePackCropTags(t *testing.T) {
	tok := smokeJWT(t)

	resp := authGet(t, tok, "/commons/catalog/gr33n-recipe-pack-v7-lettuce-veg")
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusOK)
	catalog := decodeMap(t, resp)
	body, _ := catalog["body"].(map[string]any)
	if body == nil {
		t.Fatal("catalog body missing")
	}
	programs, _ := body["programs"].([]any)
	if len(programs) < 3 {
		t.Fatalf("expected >=3 tagged programs in catalog body, got %d", len(programs))
	}

	cannabisName := "Recipe Pack v7 — Cannabis Flower Standard"
	var packProgram map[string]any
	for _, raw := range programs {
		row, _ := raw.(map[string]any)
		if row["name"] == cannabisName {
			packProgram = row
			keys, _ := row["recommended_crop_keys"].([]any)
			if len(keys) == 0 {
				t.Fatalf("catalog program missing recommended_crop_keys: %#v", row)
			}
		}
	}
	if packProgram == nil {
		t.Fatal("catalog pack missing cannabis flower program — run phase 108 migration")
	}

	resp = authGet(t, tok, "/farms/1/fertigation/programs")
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusOK)
	existing := decodeSlice(t, resp)
	var cannabisID int64
	for _, raw := range existing {
		row, _ := raw.(map[string]any)
		if row["name"] == cannabisName {
			cannabisID = int64(row["id"].(float64))
			break
		}
	}

	metaPatch := map[string]any{
		"recommended_crop_keys": packProgram["recommended_crop_keys"],
		"recommended_stages":    packProgram["recommended_stages"],
		"profile_ec_source":     packProgram["profile_ec_source"],
		"ec_band_mscm":          packProgram["ec_band_mscm"],
	}

	if cannabisID == 0 {
		payload := map[string]any{
			"name":                cannabisName,
			"description":         packProgram["description"],
			"total_volume_liters": packProgram["total_volume_liters"],
			"ec_trigger_low":      packProgram["ec_trigger_low"],
			"ph_trigger_low":      packProgram["ph_trigger_low"],
			"ph_trigger_high":     packProgram["ph_trigger_high"],
			"is_active":           false,
		}
		for k, v := range metaPatch {
			payload[k] = v
		}
		resp = authPost(t, tok, "/farms/1/fertigation/programs", payload)
		expectStatus(t, resp, http.StatusCreated)
		cannabisID = int64(decodeMap(t, resp)["id"].(float64))
	} else {
		resp = authPatch(t, tok, fmt.Sprintf("/fertigation/programs/%d/metadata", cannabisID), metaPatch)
		expectStatus(t, resp, http.StatusOK)
	}

	resp = authGet(t, tok, "/farms/1/fertigation/programs?crop_key=cannabis&stage=early_flower")
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusOK)
	filtered := decodeSlice(t, resp)
	found := false
	for _, raw := range filtered {
		row, _ := raw.(map[string]any)
		if int64(row["id"].(float64)) == cannabisID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("program id=%d not in cannabis+early_flower filter", cannabisID)
	}

	resp = authGet(t, tok, "/farms/1/fertigation/programs?crop_key=lettuce&stage=early_veg")
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusOK)
	for _, raw := range decodeSlice(t, resp) {
		row, _ := raw.(map[string]any)
		if int64(row["id"].(float64)) == cannabisID {
			t.Fatal("cannabis program should not match lettuce filter")
		}
	}
}
