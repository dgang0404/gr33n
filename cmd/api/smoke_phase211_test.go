// Phase 211 WS6 — natural farming Commons import + switchover pack apply smoke.

package main

import (
	"fmt"
	"net/http"
	"testing"
)

func smokeBlankFarm(t *testing.T, tok string) int64 {
	t.Helper()
	resp := authPost(t, tok, "/farms", map[string]any{
		"name":               uniqueName("ph211_blank"),
		"owner_user_id":      smokeDevUserUUID,
		"timezone":           "UTC",
		"currency":           "USD",
		"operational_status": "active",
		"scale_tier":         "small",
		"bootstrap_template": "none",
	})
	expectStatus(t, resp, http.StatusCreated)
	payload := decodeMap(t, resp)
	farmObj, ok := payload["farm"].(map[string]any)
	if !ok {
		t.Fatalf("expected farm in response, got %#v", payload)
	}
	return int64(farmObj["id"].(float64))
}

func TestPhase211CommonsNaturalFarmingRecipePackImport(t *testing.T) {
	tok := smokeJWT(t)
	farmID := smokeBlankFarm(t, tok)

	resp := authGet(t, tok, "/commons/catalog/jadam-indoor-starter-recipes-v1")
	expectStatus(t, resp, http.StatusOK)
	detail := decodeMap(t, resp)
	if detail["slug"] != "jadam-indoor-starter-recipes-v1" {
		t.Fatalf("unexpected slug %v", detail["slug"])
	}

	resp = authPost(t, tok, fmt.Sprintf("/farms/%d/commons/catalog-imports", farmID), map[string]any{
		"slug": "jadam-indoor-starter-recipes-v1",
		"note": "phase 211 smoke",
	})
	expectStatus(t, resp, http.StatusOK)
	out := decodeMap(t, resp)
	apply, _ := out["apply"].(map[string]any)
	if apply == nil {
		t.Fatalf("expected apply block, got %#v", out)
	}
	if apply["kind"] != "natural_farming_recipe_pack" {
		t.Fatalf("unexpected kind %v", apply["kind"])
	}
	if apply["status"] != "applied" {
		t.Fatalf("expected applied on blank farm, got %#v", apply)
	}
	createdIn, _ := apply["inputs_created"].(float64)
	createdRec, _ := apply["recipes_created"].(float64)
	if createdIn < 1 || createdRec < 1 {
		t.Fatalf("expected inputs and recipes created, got %#v", apply)
	}

	// Re-import is idempotent (skipped rows, components refreshed).
	resp = authPost(t, tok, fmt.Sprintf("/farms/%d/commons/catalog-imports", farmID), map[string]any{
		"slug": "jadam-indoor-starter-recipes-v1",
	})
	expectStatus(t, resp, http.StatusOK)
	out2 := decodeMap(t, resp)
	apply2, _ := out2["apply"].(map[string]any)
	if apply2 == nil {
		t.Fatalf("expected apply on re-import, got %#v", out2)
	}
	skippedIn, _ := apply2["inputs_skipped"].(float64)
	skippedRec, _ := apply2["recipes_skipped"].(float64)
	if skippedIn < 1 || skippedRec < 1 {
		t.Fatalf("expected skipped counts on re-import, got %#v", apply2)
	}
}

func TestPhase211SwitchoverPackApplyIdempotent(t *testing.T) {
	tok := smokeJWT(t)
	farmID := smokeBlankFarm(t, tok)

	resp := authPost(t, tok, fmt.Sprintf("/farms/%d/naturalfarming/apply-pack", farmID), map[string]any{
		"pack_key": "mericle_veg_to_jlf_v1",
	})
	expectStatus(t, resp, http.StatusOK)
	out := decodeMap(t, resp)
	if out["pack_key"] != "mericle_veg_to_jlf_v1" {
		t.Fatalf("unexpected pack_key %v", out["pack_key"])
	}
	if out["status"] != "applied" {
		t.Fatalf("expected applied on blank farm, got %#v", out)
	}
	apply, _ := out["apply"].(map[string]any)
	if apply == nil {
		t.Fatalf("expected nested apply, got %#v", out)
	}
	createdIn, _ := apply["inputs_created"].(float64)
	createdRec, _ := apply["recipes_created"].(float64)
	if createdIn < 2 || createdRec < 2 {
		t.Fatalf("expected veg switchover subset created, got %#v", apply)
	}

	resp = authPost(t, tok, fmt.Sprintf("/farms/%d/naturalfarming/apply-pack", farmID), map[string]any{
		"pack_key": "mericle_veg_to_jlf_v1",
	})
	expectStatus(t, resp, http.StatusOK)
	out2 := decodeMap(t, resp)
	if out2["status"] != "already_applied" {
		t.Fatalf("expected already_applied on second apply, got %#v", out2)
	}
}
