// Phase 94 — genetics EC profile precedence smoke.
package main

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
)

func TestPhase94_GeneticsProfilePrecedence(t *testing.T) {
	tok := smokeJWT(t)
	farmID := int64(1)
	cropKey := "cannabis"
	variety := "Blue Dream Test"
	slug := "blue_dream_test"

	t.Cleanup(func() {
		resp := authDelete(t, tok, fmt.Sprintf("/farms/%d/crop-profiles/%s/genetics/%s", farmID, cropKey, slug))
		if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
			t.Logf("genetics cleanup: %d", resp.StatusCode)
		}
		resp = authDelete(t, tok, fmt.Sprintf("/farms/%d/crop-profiles/%s", farmID, cropKey))
		if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
			t.Logf("farm override cleanup: %d", resp.StatusCode)
		}
	})

	resp := authPut(t, tok, fmt.Sprintf("/farms/%d/crop-profiles/%s", farmID, cropKey), map[string]any{
		"display_name": "Cannabis (farm)",
		"source":       "phase94 farm smoke",
		"stages": []map[string]any{
			{
				"stage":     "early_flower",
				"ec_min":    2.0,
				"ec_target": 2.2,
				"ec_max":    2.4,
			},
		},
	})
	expectStatus(t, resp, http.StatusOK)

	resp = authPut(t, tok, fmt.Sprintf("/farms/%d/crop-profiles/%s/genetics/%s", farmID, cropKey, slug), map[string]any{
		"variety_label": variety,
		"source":        "phase94 genetics smoke",
		"stages": []map[string]any{
			{
				"stage":     "early_flower",
				"ec_min":    2.6,
				"ec_target": 2.8,
				"ec_max":    3.0,
			},
		},
	})
	expectStatus(t, resp, http.StatusOK)

	resp = authGet(t, tok, fmt.Sprintf("/farms/%d/crop-profiles/effective?crop_key=%s&variety=%s", farmID, cropKey, url.QueryEscape(variety)))
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusOK)
	body := decodeMap(t, resp)
	if body["source"] != "genetics" {
		t.Fatalf("source: want genetics got %#v", body["source"])
	}
	genTarget := stageEcTarget(t, body)
	if genTarget < 2.7 {
		t.Fatalf("genetics ec_target: want >=2.7 got %v", genTarget)
	}

	resp = authGet(t, tok, fmt.Sprintf("/farms/%d/crop-profiles/effective?crop_key=%s", farmID, cropKey))
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusOK)
	farmBody := decodeMap(t, resp)
	if farmBody["source"] != "farm" {
		t.Fatalf("farm source: want farm got %#v", farmBody["source"])
	}
	farmTarget := stageEcTarget(t, farmBody)
	if farmTarget >= 2.7 {
		t.Fatalf("farm ec_target should be below genetics: %v", farmTarget)
	}
	if farmTarget >= genTarget {
		t.Fatalf("farm target %v should be less than genetics %v", farmTarget, genTarget)
	}
}

func stageEcTarget(t *testing.T, body map[string]any) float64 {
	t.Helper()
	stages, _ := body["stages"].([]any)
	if len(stages) == 0 {
		t.Fatal("expected stages")
	}
	row, _ := stages[0].(map[string]any)
	target, _ := row["ec_target"].(float64)
	return target
}
