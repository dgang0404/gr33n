// Phase 90 — device taxonomy API contract smoke.
package main

import (
	"net/http"
	"testing"
)

func TestPhase90_DeviceTaxonomyContract(t *testing.T) {
	tok := smokeJWT(t)

	resp := authGet(t, tok, "/platform/device-taxonomy")
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusOK)

	body := decodeMap(t, resp)
	sensors, ok := body["sensors"].([]any)
	if !ok || len(sensors) < 20 {
		t.Fatalf("sensors: want >=20, got %d", len(sensors))
	}
	actuators, ok := body["actuators"].([]any)
	if !ok || len(actuators) < 15 {
		t.Fatalf("actuators: want >=15, got %d", len(actuators))
	}

	foundTempF := false
	for _, raw := range sensors {
		row, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if row["type_key"] == "temp_f" {
			foundTempF = true
			if row["plant_need"] != "air" {
				t.Fatalf("temp_f plant_need: %#v", row["plant_need"])
			}
		}
	}
	if !foundTempF {
		t.Fatal("expected temp_f sensor in registry")
	}

	byNeed, ok := body["by_plant_need"].(map[string]any)
	if !ok {
		t.Fatalf("by_plant_need missing: %#v", body["by_plant_need"])
	}
	water, _ := byNeed["water"].([]any)
	if len(water) < 8 {
		t.Fatalf("by_plant_need.water: want >=8, got %d", len(water))
	}

	wiring, _ := body["wiring_source_options"].([]any)
	if len(wiring) < 4 {
		t.Fatalf("wiring_source_options: want >=4, got %d", len(wiring))
	}
}
