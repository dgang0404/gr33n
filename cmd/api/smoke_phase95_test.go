// Phase 95 — catalog integrator ops cadence smoke.
package main

import (
	"net/http"
	"testing"

	"gr33n-api/internal/croplibrary"
)

func TestPhase95_CatalogVersionMatchesPicker(t *testing.T) {
	tok := smokeJWT(t)
	root, err := croplibrary.FindRepoRoot()
	if err != nil {
		t.Fatal(err)
	}
	cat, err := croplibrary.LoadCatalog(root, croplibrary.DefaultCatalogPath)
	if err != nil {
		t.Fatal(err)
	}
	if cat.Version < 1 {
		t.Fatalf("yaml version: want >= 1, got %d", cat.Version)
	}

	resp := authGet(t, tok, "/farms/1/crop-library/picker")
	expectStatus(t, resp, http.StatusOK)
	picker := decodeMap(t, resp)
	pv, _ := picker["version"].(float64)
	if int(pv) != cat.Version {
		t.Fatalf("picker version: yaml=%d picker=%v", cat.Version, picker["version"])
	}
}

func TestPhase95_NewCropInPickerAndCommons(t *testing.T) {
	tok := smokeJWT(t)
	const cropKey = "san_pedro"

	resp := authGet(t, tok, "/commons/crop-catalog/"+cropKey)
	expectStatus(t, resp, http.StatusOK)
	entry := decodeMap(t, resp)
	if entry["crop_key"] != cropKey {
		t.Fatalf("commons crop_key: %#v", entry["crop_key"])
	}
	if entry["supported"] != true {
		t.Fatalf("san_pedro should be supported: %#v", entry["supported"])
	}

	resp = authGet(t, tok, "/farms/1/crop-library/picker")
	expectStatus(t, resp, http.StatusOK)
	picker := decodeMap(t, resp)
	groups, _ := picker["groups"].([]any)
	var found bool
	for _, g := range groups {
		grp, _ := g.(map[string]any)
		items, _ := grp["items"].([]any)
		for _, raw := range items {
			item, _ := raw.(map[string]any)
			if item["crop_key"] == cropKey {
				found = true
				break
			}
		}
	}
	if !found {
		t.Fatalf("%s not found in crop-library picker", cropKey)
	}
}
