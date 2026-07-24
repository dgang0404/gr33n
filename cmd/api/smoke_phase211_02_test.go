// Phase 211.02 WS5 — crop ops timeline API smoke.

package main

import (
	"fmt"
	"net/http"
	"testing"
)

func TestPhase211_02_CropOpsTimeline(t *testing.T) {
	tok := smokeJWT(t)
	name := uniqueName("smoke_ops_tl")
	resp := authPost(t, tok, "/farms/1/crop-cycles", map[string]any{
		"zone_id":       1,
		"name":          name,
		"current_stage": "early_veg",
		"started_at":    "2025-01-01",
		"is_active":     false,
	})
	expectStatus(t, resp, http.StatusCreated)
	created := decodeMap(t, resp)
	cycleID := int64(created["id"].(float64))

	resp = authPatch(t, tok, fmt.Sprintf("/crop-cycles/%d/stage", cycleID), map[string]any{
		"current_stage": "late_veg",
	})
	expectStatus(t, resp, http.StatusOK)

	resp = authGet(t, tok, fmt.Sprintf("/farms/1/crop-cycles/%d/ops-timeline?from=2024-01-01&to=2030-01-01", cycleID))
	expectStatus(t, resp, http.StatusOK)
	body := decodeMap(t, resp)
	if int64(body["crop_cycle_id"].(float64)) != cycleID {
		t.Fatalf("crop_cycle_id = %v", body["crop_cycle_id"])
	}
	events, ok := body["events"].([]any)
	if !ok {
		t.Fatalf("events type = %T", body["events"])
	}
	if len(events) < 2 {
		t.Fatalf("expected at least 2 stage events, got %d", len(events))
	}
	hasStage := false
	for _, raw := range events {
		ev := raw.(map[string]any)
		if ev["kind"] == "stage" {
			hasStage = true
			break
		}
	}
	if !hasStage {
		t.Fatal("expected at least one stage event in ops timeline")
	}
}
