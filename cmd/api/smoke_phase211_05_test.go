// Phase 211.05 WS3 — recipe outcomes API smoke.

package main

import (
	"net/http"
	"testing"
)

func TestPhase211_05_RecipeOutcomesAPI(t *testing.T) {
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/farms/1/crop-analytics/recipe-outcomes")
	expectStatus(t, resp, http.StatusOK)
	body := decodeMap(t, resp)
	if int64(body["farm_id"].(float64)) != 1 {
		t.Fatalf("farm_id = %v", body["farm_id"])
	}
	if _, ok := body["outcomes"]; !ok {
		t.Fatalf("missing outcomes key: %#v", body)
	}
	if _, ok := body["min_sample_size"]; !ok {
		t.Fatalf("missing min_sample_size: %#v", body)
	}
}
