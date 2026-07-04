// Phase 85 — catalog-bound plants: crop_key upsert + unsupported rejection.

package main

import (
	"fmt"
	"net/http"
	"testing"
)

func TestPhase85CatalogBoundPlants(t *testing.T) {
	tok := smokeJWT(t)

	// First create → 201. Uses "cucumber" — a crop_key the Phase 124 demo
	// seed doesn't touch — so this test's own delete at the end never
	// removes a permanently-seeded farm plant.
	resp := authPost(t, tok, "/farms/1/plants", map[string]any{
		"crop_key":            "cucumber",
		"variety_or_cultivar": "Marketmore",
	})
	expectStatus(t, resp, http.StatusCreated)
	first := decodeMap(t, resp)
	firstID := int64(first["id"].(float64))
	if first["crop_key"] != "cucumber" {
		t.Fatalf("expected crop_key=cucumber, got %v", first["crop_key"])
	}
	if first["display_name"] == nil || first["display_name"] == "" {
		t.Fatal("expected server-set display_name from catalog")
	}

	// Duplicate → 200 same id (upsert)
	resp = authPost(t, tok, "/farms/1/plants", map[string]any{
		"crop_key":            "cucumber",
		"variety_or_cultivar": "Persian",
	})
	expectStatus(t, resp, http.StatusOK)
	second := decodeMap(t, resp)
	secondID := int64(second["id"].(float64))
	if secondID != firstID {
		t.Fatalf("duplicate crop_key should return same plant id: %d vs %d", secondID, firstID)
	}
	if second["variety_or_cultivar"] != "Persian" {
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
