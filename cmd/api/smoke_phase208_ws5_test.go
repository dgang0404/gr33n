// Phase 208 WS5 — natural farming read API smokes.
package main

import (
	"net/http"
	"testing"
)

func TestPhase208WS5_ProcessCatalog(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/v1/field-guides/process-catalog")
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusOK)
	body := decodeMap(t, resp)
	if int(body["version"].(float64)) < 1 {
		t.Fatalf("version: %v", body["version"])
	}
	materials, ok := body["materials"].([]any)
	if !ok || len(materials) < 14 {
		t.Fatalf("materials: %T len=%d", body["materials"], len(materials))
	}
}

func TestPhase208WS5_ProcessMaterialGoldenrod(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/v1/field-guides/process-catalog/materials/goldenrod")
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusOK)
	body := decodeMap(t, resp)
	if body["id"] != "goldenrod" {
		t.Fatalf("id: %v", body["id"])
	}
	if body["source_tier"] != "extension_method" {
		t.Fatalf("source_tier: %v", body["source_tier"])
	}

	resp404 := authGet(t, tok, "/v1/field-guides/process-catalog/materials/no-such-plant")
	defer resp404.Body.Close()
	expectStatus(t, resp404, http.StatusNotFound)
}

func TestPhase208WS5_RecipeCanon(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/v1/field-guides/recipe-canon")
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusOK)
	body := decodeMap(t, resp)
	inputs, ok := body["inputs"].([]any)
	if !ok || len(inputs) < 16 {
		t.Fatalf("inputs: %T len=%d", body["inputs"], len(inputs))
	}
	recipes, ok := body["application_recipes"].([]any)
	if !ok || len(recipes) < 14 {
		t.Fatalf("application_recipes: %T len=%d", body["application_recipes"], len(recipes))
	}
}
