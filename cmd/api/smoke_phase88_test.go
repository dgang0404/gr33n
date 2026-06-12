// Phase 88 — domain enums API contract smoke.
package main

import (
	"fmt"
	"net/http"
	"testing"
)

func TestPhase88_DomainEnumsContract(t *testing.T) {
	tok := smokeJWT(t)

	resp := authGet(t, tok, "/platform/domain-enums")
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusOK)

	body := decodeMap(t, resp)
	stages, ok := body["growth_stages"].([]any)
	if !ok || len(stages) != 11 {
		t.Fatalf("growth_stages: want 11 options, got %#v", body["growth_stages"])
	}
	foundTransition := false
	foundFlush := false
	for _, raw := range stages {
		row, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		v, _ := row["value"].(string)
		if v == "transition" {
			foundTransition = true
		}
		if v == "flush" {
			foundFlush = true
		}
	}
	if !foundTransition || !foundFlush {
		t.Fatalf("expected transition and flush in growth_stages, got %#v", stages)
	}

	costs, ok := body["cost_categories"].([]any)
	if !ok || len(costs) < 10 {
		t.Fatalf("cost_categories: want >=10, got %d", len(costs))
	}
	reservoirs, ok := body["reservoir_statuses"].([]any)
	if !ok || len(reservoirs) < 5 {
		t.Fatalf("reservoir_statuses: want >=5, got %d", len(reservoirs))
	}
}

func TestPhase88_SetpointTransitionStagePersists(t *testing.T) {
	tok := smokeJWT(t)
	zoneID := seedSetpointZone(t)
	sensorType := uniqueName("ph88_transition")

	resp := authPost(t, tok, "/farms/1/setpoints", map[string]any{
		"zone_id":     zoneID,
		"stage":       "transition",
		"sensor_type": sensorType,
		"min_value":   20.0,
		"max_value":   28.0,
		"ideal_value": 24.0,
	})
	expectStatus(t, resp, http.StatusCreated)
	id := int64(decodeMap(t, resp)["id"].(float64))

	resp = authGet(t, tok, fmt.Sprintf("/setpoints/%d", id))
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusOK)
	got := decodeMap(t, resp)
	stage := got["stage"]
	stageStr, _ := stage.(string)
	if stageStr == "" {
		if m, ok := stage.(map[string]any); ok {
			stageStr, _ = m["gr33nfertigation_growth_stage_enum"].(string)
		}
	}
	if stageStr != "transition" {
		t.Fatalf("expected stage transition, got %#v", got["stage"])
	}

	resp = authDelete(t, tok, fmt.Sprintf("/setpoints/%d", id))
	expectStatus(t, resp, http.StatusNoContent)
}
