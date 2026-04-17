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

	// Create
	name := uniqueName("smoke_plant")
	resp := authPost(t, tok, "/farms/1/plants", map[string]any{
		"display_name":        name,
		"variety_or_cultivar": "Indica",
		"meta":                map[string]any{"photoperiod": "short-day"},
	})
	expectStatus(t, resp, http.StatusCreated)
	created := decodeMap(t, resp)
	plantID := int64(created["id"].(float64))
	if created["display_name"] != name {
		t.Fatalf("expected display_name=%s, got %v", name, created["display_name"])
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
	if got["display_name"] != name {
		t.Fatalf("get: expected display_name=%s, got %v", name, got["display_name"])
	}

	// Update
	updatedName := uniqueName("smoke_plant_upd")
	resp = authPut(t, tok, fmt.Sprintf("/plants/%d", plantID), map[string]any{
		"display_name":        updatedName,
		"variety_or_cultivar": "Sativa",
		"meta":                map[string]any{"photoperiod": "long-day"},
	})
	expectStatus(t, resp, http.StatusOK)
	updated := decodeMap(t, resp)
	if updated["display_name"] != updatedName {
		t.Fatalf("expected updated name=%s, got %v", updatedName, updated["display_name"])
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
