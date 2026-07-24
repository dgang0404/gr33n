// Phase 211.03 WS9 — farm permissions smoke tests.

package main

import (
	"context"
	"fmt"
	"net/http"
	"testing"
)

func TestPhase211_03_MeCaps(t *testing.T) {
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/farms/1/me/caps")
	expectStatus(t, resp, http.StatusOK)
	body := decodeMap(t, resp)
	scopes, ok := body["scopes"].([]any)
	if !ok || len(scopes) == 0 {
		t.Fatalf("expected scopes array, got %v", body["scopes"])
	}
}

func TestPhase211_03_ViewerCannotDeleteInput(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	_, viewerTok := seedSmokeViewerUser(t, context.Background())
	tok := smokeJWT(t)
	name := uniqueName("smoke_nf_input")
	resp := authPost(t, tok, "/farms/1/naturalfarming/inputs", map[string]any{
		"name":     name,
		"category": "fermented_plant_juice",
	})
	expectStatus(t, resp, http.StatusCreated)
	id := int64(decodeMap(t, resp)["id"].(float64))

	resp = authDelete(t, viewerTok, fmt.Sprintf("/naturalfarming/inputs/%d", id))
	expectStatus(t, resp, http.StatusForbidden)
}

func TestPhase211_03_FinanceCannotDeleteRecipe(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	_, financeTok := seedSmokeFinanceUser(t, context.Background())
	tok := smokeJWT(t)
	resp := authPost(t, tok, "/farms/1/naturalfarming/recipes", map[string]any{
		"name":                    uniqueName("smoke_recipe_cap"),
		"target_application_type": "foliar_spray",
	})
	expectStatus(t, resp, http.StatusCreated)
	id := int64(decodeMap(t, resp)["id"].(float64))

	resp = authDelete(t, financeTok, fmt.Sprintf("/naturalfarming/recipes/%d", id))
	expectStatus(t, resp, http.StatusForbidden)
}
