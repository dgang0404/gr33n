// Phase 20.95 WS5 — split out of cmd/api/smoke_test.go with zero behaviour
// change. Shared globals (testPool / testServer / testWorker / testNotifier)
// and helpers live in smoke_helpers_test.go; TestMain stays in smoke_test.go.

package main

import (
	"fmt"
	"net/http"
	"testing"
)

func TestRecipeList(t *testing.T) {
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/farms/1/naturalfarming/recipes")
	expectStatus(t, resp, 200)
	items := decodeSlice(t, resp)
	if len(items) == 0 {
		t.Fatal("expected seeded recipes")
	}
}

func TestNfInputDefinitionCRUD(t *testing.T) {
	tok := smokeJWT(t)

	name := uniqueName("smoke_nf_input")
	resp := authPost(t, tok, "/farms/1/naturalfarming/inputs", map[string]any{
		"name":        name,
		"category":    "fermented_plant_juice",
		"description": "smoke test input",
	})
	expectStatus(t, resp, http.StatusCreated)
	created := decodeMap(t, resp)
	inputID := int64(created["id"].(float64))

	updName := uniqueName("smoke_nf_upd")
	resp = authPut(t, tok, fmt.Sprintf("/naturalfarming/inputs/%d", inputID), map[string]any{
		"name":        updName,
		"category":    "fermented_plant_juice",
		"description": "updated",
	})
	expectStatus(t, resp, http.StatusOK)
	updated := decodeMap(t, resp)
	if updated["name"] != updName {
		t.Fatalf("expected updated name=%s, got %v", updName, updated["name"])
	}

	resp = authDelete(t, tok, fmt.Sprintf("/naturalfarming/inputs/%d", inputID))
	expectStatus(t, resp, http.StatusNoContent)
}

func TestNfBatchCRUD(t *testing.T) {
	tok := smokeJWT(t)

	inputsResp := authGet(t, tok, "/farms/1/naturalfarming/inputs")
	expectStatus(t, inputsResp, http.StatusOK)
	inputs := decodeSlice(t, inputsResp)
	if len(inputs) == 0 {
		t.Skip("no NF inputs to create batch against")
	}
	inputID := int64(inputs[0].(map[string]any)["id"].(float64))

	code := uniqueName("batch")
	resp := authPost(t, tok, "/farms/1/naturalfarming/batches", map[string]any{
		"input_definition_id": inputID,
		"batch_identifier":    code,
		"status":              "fermenting_brewing",
		"creation_start_date": "2025-06-01",
	})
	expectStatus(t, resp, http.StatusCreated)
	created := decodeMap(t, resp)
	batchID := int64(created["id"].(float64))

	resp = authPut(t, tok, fmt.Sprintf("/naturalfarming/batches/%d", batchID), map[string]any{
		"input_definition_id": inputID,
		"batch_identifier":    code,
		"status":              "ready_for_use",
	})
	expectStatus(t, resp, http.StatusOK)

	resp = authDelete(t, tok, fmt.Sprintf("/naturalfarming/batches/%d", batchID))
	expectStatus(t, resp, http.StatusNoContent)
}

func TestRecipeFullCRUD(t *testing.T) {
	tok := smokeJWT(t)

	name := uniqueName("smoke_recipe")
	resp := authPost(t, tok, "/farms/1/naturalfarming/recipes", map[string]any{
		"name":                    name,
		"description":             "smoke recipe",
		"target_application_type": "soil_drench",
	})
	expectStatus(t, resp, http.StatusCreated)
	created := decodeMap(t, resp)
	recipeID := int64(created["id"].(float64))

	resp = authGet(t, tok, fmt.Sprintf("/naturalfarming/recipes/%d", recipeID))
	expectStatus(t, resp, http.StatusOK)

	inputsResp := authGet(t, tok, "/farms/1/naturalfarming/inputs")
	expectStatus(t, inputsResp, http.StatusOK)
	inputs := decodeSlice(t, inputsResp)
	if len(inputs) > 0 {
		inputID := int64(inputs[0].(map[string]any)["id"].(float64))

		resp = authPost(t, tok, fmt.Sprintf("/naturalfarming/recipes/%d/components", recipeID), map[string]any{
			"input_definition_id": inputID,
			"volume_ml":           20.0,
			"dilution_ratio":      "1:500",
		})
		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
			t.Fatalf("add component: expected 2xx, got %d", resp.StatusCode)
		}

		resp = authGet(t, tok, fmt.Sprintf("/naturalfarming/recipes/%d/components", recipeID))
		expectStatus(t, resp, http.StatusOK)
		comps := decodeSlice(t, resp)
		if len(comps) == 0 {
			t.Fatal("expected at least one recipe component")
		}

		resp = authDelete(t, tok, fmt.Sprintf("/naturalfarming/recipes/%d/components/%d", recipeID, inputID))
		expectStatus(t, resp, http.StatusNoContent)
	}

	updName := uniqueName("smoke_recipe_upd")
	resp = authPut(t, tok, fmt.Sprintf("/naturalfarming/recipes/%d", recipeID), map[string]any{
		"name":                    updName,
		"description":             "updated smoke recipe",
		"target_application_type": "foliar_spray",
	})
	expectStatus(t, resp, http.StatusOK)

	resp = authDelete(t, tok, fmt.Sprintf("/naturalfarming/recipes/%d", recipeID))
	expectStatus(t, resp, http.StatusNoContent)
}
