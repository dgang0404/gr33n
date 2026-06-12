// Phase 98 — enterprise catalog promotion: farm-local override ≠ other farms.
package main

import (
	"fmt"
	"net/http"
	"testing"
)

func TestPhase98_FarmOverrideDoesNotPromoteToOtherFarm(t *testing.T) {
	tok := smokeJWT(t)
	const cropKey = "cannabis"
	const overrideTarget = 3.88

	resp := authPost(t, tok, "/farms", map[string]any{
		"name":               uniqueName("phase98_farm_b"),
		"owner_user_id":      smokeDevUserUUID,
		"timezone":           "UTC",
		"currency":           "USD",
		"operational_status": "active",
		"scale_tier":         "small",
		"bootstrap_template": "none",
	})
	expectStatus(t, resp, http.StatusCreated)
	payload := decodeMap(t, resp)
	farmB, ok := payload["farm"].(map[string]any)
	if !ok {
		t.Fatalf("expected farm object, got %#v", payload)
	}
	farmBID := int64(farmB["id"].(float64))

	t.Cleanup(func() {
		resp := authDelete(t, tok, fmt.Sprintf("/farms/%d/crop-profiles/%s", 1, cropKey))
		if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
			t.Logf("farm A override cleanup: %d", resp.StatusCode)
		}
		resp = authDelete(t, tok, fmt.Sprintf("/farms/%d", farmBID))
		if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
			t.Logf("farm B cleanup: %d", resp.StatusCode)
		}
	})

	resp = authPut(t, tok, fmt.Sprintf("/farms/%d/crop-profiles/%s", 1, cropKey), map[string]any{
		"display_name": "Cannabis (Farm A override)",
		"source":       "phase98 promotion smoke",
		"stages": []map[string]any{
			{
				"stage":     "early_flower",
				"ec_min":    3.5,
				"ec_target": overrideTarget,
				"ec_max":    4.0,
			},
		},
	})
	expectStatus(t, resp, http.StatusOK)
	farmA := decodeMap(t, resp)
	if farmA["is_builtin"] == true {
		t.Fatal("farm A should have farm override row")
	}

	resp = authGet(t, tok, fmt.Sprintf("/farms/%d/crop-profiles/%s", farmBID, cropKey))
	expectStatus(t, resp, http.StatusOK)
	other := decodeMap(t, resp)
	if other["is_builtin"] != true {
		t.Fatalf("farm B should still use builtin profile, got is_builtin=%v", other["is_builtin"])
	}
	targetB := stageEcTarget(t, other)
	if targetB >= overrideTarget {
		t.Fatalf("farm B ec_target %.2f should be below farm A override %.2f", targetB, overrideTarget)
	}

	resp = authGet(t, tok, fmt.Sprintf("/farms/%d/crop-profiles/effective?crop_key=%s", 1, cropKey))
	expectStatus(t, resp, http.StatusOK)
	effectiveA := decodeMap(t, resp)
	targetA := stageEcTarget(t, effectiveA)
	if targetA < overrideTarget-0.01 {
		t.Fatalf("farm A effective ec_target want ~%.2f got %.2f", overrideTarget, targetA)
	}
}
