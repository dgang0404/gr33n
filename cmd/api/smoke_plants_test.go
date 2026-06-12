// Phase 20.95 WS5 — split out of cmd/api/smoke_test.go with zero behaviour
// change. Shared globals (testPool / testServer / testWorker / testNotifier)
// and helpers live in smoke_helpers_test.go; TestMain stays in smoke_test.go.

package main

import (
	"fmt"
	"net/http"
	"testing"
)

func TestPlantCRUD(t *testing.T) {
	tok := smokeJWT(t)

	// Create (Phase 85 — catalog-bound by crop_key)
	resp := authPost(t, tok, "/farms/1/plants", map[string]any{
		"crop_key":            "basil",
		"variety_or_cultivar": "Genovese",
		"meta":                map[string]any{"photoperiod": "long-day"},
	})
	expectStatus(t, resp, http.StatusCreated)
	created := decodeMap(t, resp)
	plantID := int64(created["id"].(float64))
	if created["crop_key"] != "basil" {
		t.Fatalf("expected crop_key=basil, got %v", created["crop_key"])
	}
	if created["display_name"] == nil || created["display_name"] == "" {
		t.Fatal("expected server-set display_name from catalog")
	}

	// List
	resp = authGet(t, tok, "/farms/1/plants")
	expectStatus(t, resp, http.StatusOK)
	plants := decodeSlice(t, resp)
	found := false
	for _, item := range plants {
		if m, ok := item.(map[string]any); ok {
			if int64(m["id"].(float64)) == plantID {
				found = true
				break
			}
		}
	}
	if !found {
		t.Fatal("created plant not found in list")
	}

	// Get
	resp = authGet(t, tok, fmt.Sprintf("/plants/%d", plantID))
	expectStatus(t, resp, http.StatusOK)
	got := decodeMap(t, resp)
	if got["crop_key"] != "basil" {
		t.Fatalf("get: expected crop_key=basil, got %v", got["crop_key"])
	}

	// Update (variety + meta only for catalog-bound plants)
	resp = authPut(t, tok, fmt.Sprintf("/plants/%d", plantID), map[string]any{
		"variety_or_cultivar": "Thai",
		"meta":                map[string]any{"photoperiod": "long-day", "notes": "bench A"},
	})
	expectStatus(t, resp, http.StatusOK)
	updated := decodeMap(t, resp)
	if updated["variety_or_cultivar"] != "Thai" {
		t.Fatalf("expected updated variety=Thai, got %v", updated["variety_or_cultivar"])
	}

	// Soft delete
	resp = authDelete(t, tok, fmt.Sprintf("/plants/%d", plantID))
	expectStatus(t, resp, http.StatusNoContent)

	// Verify gone from list
	resp = authGet(t, tok, "/farms/1/plants")
	expectStatus(t, resp, http.StatusOK)
	plantsAfter := decodeSlice(t, resp)
	for _, item := range plantsAfter {
		if m, ok := item.(map[string]any); ok {
			if int64(m["id"].(float64)) == plantID {
				t.Fatal("soft-deleted plant still appears in list")
			}
		}
	}
}

// ── Phase 18: Smoke Test Gap Fill ────────────────────────────────────────────
