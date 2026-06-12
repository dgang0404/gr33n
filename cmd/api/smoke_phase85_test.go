// Phase 85 — catalog-bound plants: crop_key upsert + unsupported rejection.

package main

import (
	"fmt"
	"net/http"
	"testing"
)

func TestPhase85CatalogBoundPlants(t *testing.T) {
	tok := smokeJWT(t)

	// First tomato create → 201
	resp := authPost(t, tok, "/farms/1/plants", map[string]any{
		"crop_key":            "tomato",
		"variety_or_cultivar": "Cherokee Purple",
	})
	expectStatus(t, resp, http.StatusCreated)
	first := decodeMap(t, resp)
	firstID := int64(first["id"].(float64))
	if first["crop_key"] != "tomato" {
		t.Fatalf("expected crop_key=tomato, got %v", first["crop_key"])
	}
	if first["display_name"] == nil || first["display_name"] == "" {
		t.Fatal("expected server-set display_name from catalog")
	}

	// Duplicate tomato → 200 same id (upsert)
	resp = authPost(t, tok, "/farms/1/plants", map[string]any{
		"crop_key":            "tomato",
		"variety_or_cultivar": "Roma",
	})
	expectStatus(t, resp, http.StatusOK)
	second := decodeMap(t, resp)
	secondID := int64(second["id"].(float64))
	if secondID != firstID {
		t.Fatalf("duplicate crop_key should return same plant id: %d vs %d", secondID, firstID)
	}
	if second["variety_or_cultivar"] != "Roma" {
		t.Fatalf("expected variety update on upsert, got %v", second["variety_or_cultivar"])
	}

	// Unsupported catalog crop → 400
	resp = authPost(t, tok, "/farms/1/plants", map[string]any{
		"crop_key": "ramps",
	})
	expectStatus(t, resp, http.StatusBadRequest)
	body := decodeMap(t, resp)
	errMsg, _ := body["error"].(string)
	if errMsg == "" {
		t.Fatalf("expected error message for ramps, got %#v", body)
	}

	// Cleanup
	resp = authDelete(t, tok, fmt.Sprintf("/plants/%d", firstID))
	expectStatus(t, resp, http.StatusNoContent)
}
