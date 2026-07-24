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

func TestRecipeRevisionHistory(t *testing.T) {
	tok := smokeJWT(t)

	recipesResp := authGet(t, tok, "/farms/1/naturalfarming/recipes")
	expectStatus(t, recipesResp, http.StatusOK)
	recipes := decodeSlice(t, recipesResp)
	if len(recipes) == 0 {
		t.Fatal("expected seeded recipes")
	}
	recipeID := int64(recipes[0].(map[string]any)["id"].(float64))

	revsResp := authGet(t, tok, fmt.Sprintf("/naturalfarming/recipes/%d/revisions", recipeID))
	expectStatus(t, revsResp, http.StatusOK)
	revs := decodeSlice(t, revsResp)
	if len(revs) == 0 {
		t.Fatal("expected bootstrap revision for seeded recipe")
	}
	first := revs[len(revs)-1].(map[string]any)
	if int(first["revision_number"].(float64)) != 1 {
		t.Fatalf("expected revision_number 1, got %v", first["revision_number"])
	}

	name := uniqueName("smoke_rev_recipe")
	resp := authPost(t, tok, "/farms/1/naturalfarming/recipes", map[string]any{
		"name":                    name,
		"description":             "revision smoke",
		"target_application_type": "soil_drench",
		"dilution_ratio":          "1:10",
	})
	expectStatus(t, resp, http.StatusCreated)
	created := decodeMap(t, resp)
	newID := int64(created["id"].(float64))

	revsResp = authGet(t, tok, fmt.Sprintf("/naturalfarming/recipes/%d/revisions", newID))
	expectStatus(t, revsResp, http.StatusOK)
	revs = decodeSlice(t, revsResp)
	if len(revs) != 1 {
		t.Fatalf("expected 1 revision after create, got %d", len(revs))
	}

	resp = authPut(t, tok, fmt.Sprintf("/naturalfarming/recipes/%d", newID), map[string]any{
		"name":                    name,
		"description":             "revision smoke updated",
		"target_application_type": "soil_drench",
		"dilution_ratio":          "1:20",
	})
	expectStatus(t, resp, http.StatusOK)

	revsResp = authGet(t, tok, fmt.Sprintf("/naturalfarming/recipes/%d/revisions", newID))
	expectStatus(t, revsResp, http.StatusOK)
	revs = decodeSlice(t, revsResp)
	if len(revs) != 2 {
		t.Fatalf("expected 2 revisions after update, got %d", len(revs))
	}
	latest := revs[0].(map[string]any)
	if latest["snapshot"] == nil {
		t.Fatal("expected snapshot on latest revision")
	}
	oldest := revs[1].(map[string]any)
	if oldest["snapshot"] == nil {
		t.Fatal("expected snapshot on first revision")
	}
	if int(latest["revision_number"].(float64)) != 2 {
		t.Fatalf("expected latest revision_number 2, got %v", latest["revision_number"])
	}

	resp = authDelete(t, tok, fmt.Sprintf("/naturalfarming/recipes/%d", newID))
	expectStatus(t, resp, http.StatusNoContent)
}

func TestRecipeRestoreRevision(t *testing.T) {
	tok := smokeJWT(t)

	name := uniqueName("smoke_restore")
	resp := authPost(t, tok, "/farms/1/naturalfarming/recipes", map[string]any{
		"name":                    name,
		"target_application_type": "soil_drench",
		"dilution_ratio":          "1:10",
	})
	expectStatus(t, resp, http.StatusCreated)
	created := decodeMap(t, resp)
	recipeID := int64(created["id"].(float64))

	resp = authPut(t, tok, fmt.Sprintf("/naturalfarming/recipes/%d", recipeID), map[string]any{
		"name":                    name,
		"target_application_type": "soil_drench",
		"dilution_ratio":          "1:20",
	})
	expectStatus(t, resp, http.StatusOK)

	revsResp := authGet(t, tok, fmt.Sprintf("/naturalfarming/recipes/%d/revisions", recipeID))
	expectStatus(t, revsResp, http.StatusOK)
	revs := decodeSlice(t, revsResp)
	if len(revs) < 2 {
		t.Fatalf("expected at least 2 revisions, got %d", len(revs))
	}
	restoreFrom := revs[len(revs)-1].(map[string]any)
	revID := int64(restoreFrom["id"].(float64))

	resp = authPost(t, tok, fmt.Sprintf("/naturalfarming/recipes/%d/revisions/%d/restore", recipeID, revID), nil)
	expectStatus(t, resp, http.StatusOK)
	body := decodeMap(t, resp)
	restored := body["recipe"].(map[string]any)
	if restored["dilution_ratio"] != "1:10" {
		t.Fatalf("expected dilution 1:10 after restore, got %v", restored["dilution_ratio"])
	}
	newRev, ok := body["revision"].(map[string]any)
	if !ok {
		t.Fatal("expected revision in restore response")
	}
	if int(newRev["revision_number"].(float64)) != 3 {
		t.Fatalf("expected revision 3 after restore, got %v", newRev["revision_number"])
	}

	resp = authDelete(t, tok, fmt.Sprintf("/naturalfarming/recipes/%d", recipeID))
	expectStatus(t, resp, http.StatusNoContent)
}
