// Phase 92 — zone and greenhouse vocabulary on domain-enums contract smoke.
package main

import (
	"net/http"
	"testing"
)

func TestPhase92_ZoneGreenhouseVocabularyContract(t *testing.T) {
	tok := smokeJWT(t)

	resp := authGet(t, tok, "/platform/domain-enums")
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusOK)

	body := decodeMap(t, resp)
	zoneTypes, ok := body["zone_types"].([]any)
	if !ok || len(zoneTypes) != 8 {
		t.Fatalf("zone_types: want 8, got %#v", body["zone_types"])
	}

	wizardVisible := 0
	foundFilm := false
	for _, raw := range zoneTypes {
		row, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if row["wizard_visible"] == true {
			wizardVisible++
		}
	}

	covers, ok := body["greenhouse_cover_types"].([]any)
	if !ok || len(covers) != 3 {
		t.Fatalf("greenhouse_cover_types: want 3, got %#v", body["greenhouse_cover_types"])
	}
	for _, raw := range covers {
		row, _ := raw.(map[string]any)
		if row["value"] == "film" {
			foundFilm = true
			if row["label"] != "Film / poly" {
				t.Fatalf("film label: %#v", row["label"])
			}
		}
	}
	if !foundFilm {
		t.Fatal("expected film cover type")
	}

	policies, ok := body["greenhouse_automation_policies"].([]any)
	if !ok || len(policies) != 3 {
		t.Fatalf("greenhouse_automation_policies: want 3, got %#v", body["greenhouse_automation_policies"])
	}
	if wizardVisible != 3 {
		t.Fatalf("wizard_visible zone types: want 3, got %d", wizardVisible)
	}
}
