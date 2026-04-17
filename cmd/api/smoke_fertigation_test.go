// Phase 20.95 WS5 — split out of cmd/api/smoke_test.go with zero behaviour
// change. Shared globals (testPool / testServer / testWorker / testNotifier)
// and helpers live in smoke_helpers_test.go; TestMain stays in smoke_test.go.

package main

import (
	"fmt"
	"net/http"
	"testing"
)

func TestFertigationReservoirRoundtrip(t *testing.T) {
	tok := smokeJWT(t)
	name := uniqueName("smoke_reservoir")
	payload := map[string]any{
		"name":                  name,
		"status":                "ready",
		"capacity_liters":       100.0,
		"current_volume_liters": 50.0,
	}
	resp := authPost(t, tok, "/farms/1/fertigation/reservoirs", payload)
	expectStatus(t, resp, 201)
	created := decodeMap(t, resp)
	if created["name"] != name {
		t.Fatalf("expected name=%s, got %v", name, created["name"])
	}

	resp = authGet(t, tok, "/farms/1/fertigation/reservoirs")
	expectStatus(t, resp, 200)
	items := decodeSlice(t, resp)
	found := false
	for _, item := range items {
		if m, ok := item.(map[string]any); ok && m["name"] == name {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("created reservoir not found in list")
	}
}

func TestFertigationEcTargetRoundtrip(t *testing.T) {
	tok := smokeJWT(t)
	payload := map[string]any{
		"growth_stage": "early_veg",
		"ec_min_mscm":  1.0,
		"ec_max_mscm":  2.5,
		"ph_min":       5.5,
		"ph_max":       6.5,
		"notes":        "smoke test",
	}
	resp := authPost(t, tok, "/farms/1/fertigation/ec-targets", payload)
	expectStatus(t, resp, 201)

	resp = authGet(t, tok, "/farms/1/fertigation/ec-targets")
	expectStatus(t, resp, 200)
	_ = decodeSlice(t, resp)
}

func TestFertigationProgramRoundtrip(t *testing.T) {
	tok := smokeJWT(t)
	payload := map[string]any{
		"name":                uniqueName("smoke_program"),
		"total_volume_liters": 5.0,
		"is_active":           false,
		"ec_trigger_low":      0.0,
		"ph_trigger_low":      0.0,
		"ph_trigger_high":     0.0,
	}
	resp := authPost(t, tok, "/farms/1/fertigation/programs", payload)
	expectStatus(t, resp, 201)

	resp = authGet(t, tok, "/farms/1/fertigation/programs")
	expectStatus(t, resp, 200)
	_ = decodeSlice(t, resp)
}

func TestFertigationEventRoundtripWithCropCycle(t *testing.T) {
	tok := smokeJWT(t)
	name := uniqueName("smoke_cc_fert")
	resp := authPost(t, tok, "/farms/1/crop-cycles", map[string]any{
		"zone_id":       1,
		"name":          name,
		"current_stage": "early_veg",
		"started_at":    "2025-02-01",
		"is_active":     false,
	})
	expectStatus(t, resp, 201)
	cc := decodeMap(t, resp)
	ccID := int64(cc["id"].(float64))

	payload := map[string]any{
		"zone_id":               1,
		"crop_cycle_id":         ccID,
		"volume_applied_liters": 2.5,
		"ec_before_mscm":        1.2,
		"ec_after_mscm":         1.8,
		"ph_before":             6.0,
		"ph_after":              6.2,
		"trigger_source":        "manual",
	}
	resp = authPost(t, tok, "/farms/1/fertigation/events", payload)
	expectStatus(t, resp, 201)

	resp = authGet(t, tok, fmt.Sprintf("/farms/1/fertigation/events?crop_cycle_id=%d", ccID))
	expectStatus(t, resp, 200)
	items := decodeSlice(t, resp)
	if len(items) == 0 {
		t.Fatal("expected filtered fertigation events")
	}
}

// ── Phase 16: Schedule CRUD ─────────────────────────────────────────────────

func TestMixingEventCreateWithComponents(t *testing.T) {
	tok := smokeJWT(t)

	resName := uniqueName("smoke_mix_res")
	resp := authPost(t, tok, "/farms/1/fertigation/reservoirs", map[string]any{
		"name":                  resName,
		"status":                "ready",
		"capacity_liters":       50.0,
		"current_volume_liters": 40.0,
	})
	expectStatus(t, resp, 201)
	res := decodeMap(t, resp)
	resID := int64(res["id"].(float64))

	inputsResp := authGet(t, tok, "/farms/1/naturalfarming/inputs")
	expectStatus(t, inputsResp, 200)
	inputs := decodeSlice(t, inputsResp)
	if len(inputs) == 0 {
		t.Skip("no NF inputs in seed data")
	}
	inputDef := inputs[0].(map[string]any)
	inputDefID := int64(inputDef["id"].(float64))

	resp = authPost(t, tok, "/farms/1/fertigation/mixing-events", map[string]any{
		"reservoir_id":        resID,
		"water_volume_liters": 20.0,
		"water_source":        "municipal",
		"water_ec_mscm":       0.3,
		"water_ph":            7.0,
		"final_ec_mscm":       1.5,
		"final_ph":            6.2,
		"notes":               "smoke test mix",
		"components": []map[string]any{
			{
				"input_definition_id": inputDefID,
				"volume_added_ml":     40.0,
				"dilution_ratio":      "1:500",
			},
		},
	})
	expectStatus(t, resp, 201)
	result := decodeMap(t, resp)
	if result["event"] == nil {
		t.Fatal("expected event in response")
	}
	comps, ok := result["components"].([]any)
	if !ok || len(comps) != 1 {
		t.Fatalf("expected 1 component, got %v", result["components"])
	}
}

// ── Phase 16: Task Update + Delete ──────────────────────────────────────────

func TestFertigationReservoirUpdateDelete(t *testing.T) {
	tok := smokeJWT(t)

	name := uniqueName("smoke_res_ud")
	resp := authPost(t, tok, "/farms/1/fertigation/reservoirs", map[string]any{
		"name":                  name,
		"status":                "ready",
		"capacity_liters":       80.0,
		"current_volume_liters": 40.0,
	})
	expectStatus(t, resp, http.StatusCreated)
	created := decodeMap(t, resp)
	resID := int64(created["id"].(float64))

	updName := uniqueName("smoke_res_upd")
	resp = authPatch(t, tok, fmt.Sprintf("/fertigation/reservoirs/%d", resID), map[string]any{
		"name":                  updName,
		"status":                "mixing",
		"capacity_liters":       80.0,
		"current_volume_liters": 35.0,
	})
	expectStatus(t, resp, http.StatusOK)
	updated := decodeMap(t, resp)
	if updated["name"] != updName {
		t.Fatalf("expected updated name=%s, got %v", updName, updated["name"])
	}

	resp = authDelete(t, tok, fmt.Sprintf("/fertigation/reservoirs/%d", resID))
	expectStatus(t, resp, http.StatusNoContent)
}

func TestFertigationProgramUpdateDelete(t *testing.T) {
	tok := smokeJWT(t)

	name := uniqueName("smoke_prog_ud")
	resp := authPost(t, tok, "/farms/1/fertigation/programs", map[string]any{
		"name":                name,
		"total_volume_liters": 10.0,
		"is_active":           false,
		"ec_trigger_low":      0.0,
		"ph_trigger_low":      0.0,
		"ph_trigger_high":     0.0,
	})
	expectStatus(t, resp, http.StatusCreated)
	created := decodeMap(t, resp)
	progID := int64(created["id"].(float64))

	updName := uniqueName("smoke_prog_upd")
	resp = authPatch(t, tok, fmt.Sprintf("/fertigation/programs/%d", progID), map[string]any{
		"name":      updName,
		"is_active": true,
	})
	expectStatus(t, resp, http.StatusOK)
	updated := decodeMap(t, resp)
	if updated["name"] != updName {
		t.Fatalf("expected updated name=%s, got %v", updName, updated["name"])
	}

	resp = authDelete(t, tok, fmt.Sprintf("/fertigation/programs/%d", progID))
	expectStatus(t, resp, http.StatusNoContent)
}
