// Phase 20.95 WS5 — split out of cmd/api/smoke_test.go with zero behaviour
// change. Shared globals (testPool / testServer / testWorker / testNotifier)
// and helpers live in smoke_helpers_test.go; TestMain stays in smoke_test.go.

package main

import (
	"fmt"
	"net/http"
	"testing"
)

func TestCropCycleCreateAndStage(t *testing.T) {
	tok := smokeJWT(t)
	name := uniqueName("smoke_cycle")
	resp := authPost(t, tok, "/farms/1/crop-cycles", map[string]any{
		"zone_id":       1,
		"name":          name,
		"current_stage": "early_veg",
		"started_at":    "2025-01-01",
		"is_active":     false,
	})
	expectStatus(t, resp, 201)
	created := decodeMap(t, resp)
	id := int64(created["id"].(float64))

	resp = authPatch(t, tok, fmt.Sprintf("/crop-cycles/%d/stage", id), map[string]any{
		"current_stage": "late_veg",
	})
	expectStatus(t, resp, 200)
}

func TestCropCycleFullCRUD(t *testing.T) {
	tok := smokeJWT(t)

	name := uniqueName("smoke_cc_crud")
	resp := authPost(t, tok, "/farms/1/crop-cycles", map[string]any{
		"zone_id":       1,
		"name":          name,
		"current_stage": "seedling",
		"started_at":    "2025-03-01",
		"is_active":     false,
	})
	expectStatus(t, resp, http.StatusCreated)
	created := decodeMap(t, resp)
	ccID := int64(created["id"].(float64))

	resp = authGet(t, tok, fmt.Sprintf("/crop-cycles/%d", ccID))
	expectStatus(t, resp, http.StatusOK)
	got := decodeMap(t, resp)
	if got["name"] != name {
		t.Fatalf("GET crop cycle: expected name=%s, got %v", name, got["name"])
	}

	updName := uniqueName("smoke_cc_upd")
	resp = authPut(t, tok, fmt.Sprintf("/crop-cycles/%d", ccID), map[string]any{
		"name":      updName,
		"zone_id":   1,
		"is_active": false,
	})
	expectStatus(t, resp, http.StatusOK)
	updated := decodeMap(t, resp)
	if updated["name"] != updName {
		t.Fatalf("PUT crop cycle: expected name=%s, got %v", updName, updated["name"])
	}

	resp = authDelete(t, tok, fmt.Sprintf("/crop-cycles/%d", ccID))
	expectStatus(t, resp, http.StatusNoContent)

	resp = authGet(t, tok, "/farms/1/crop-cycles")
	expectStatus(t, resp, http.StatusOK)
	cycles := decodeSlice(t, resp)
	for _, c := range cycles {
		m := c.(map[string]any)
		if int64(m["id"].(float64)) == ccID && m["is_active"] == true {
			t.Fatal("deleted crop cycle still active in list")
		}
	}
}
