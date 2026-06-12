// Phase 107 — crop catalog photos in picker + commons API.
package main

import (
	"strings"
	"testing"
)

func TestPhase107_CropCatalogImageURL(t *testing.T) {
	tok := smokeJWT(t)

	resp := authGet(t, tok, "/commons/crop-catalog/san_pedro")
	expectStatus(t, resp, 200)
	sanPedro := decodeMap(t, resp)
	url, _ := sanPedro["image_url"].(string)
	if !strings.HasPrefix(url, "/assets/crops/") {
		t.Fatalf("san_pedro image_url: %#v", sanPedro["image_url"])
	}

	resp = authGet(t, tok, "/commons/crop-catalog/tomato")
	expectStatus(t, resp, 200)
	tomato := decodeMap(t, resp)
	if tomato["image_url"] != nil {
		t.Fatalf("tomato should have null image_url, got %#v", tomato["image_url"])
	}

	resp = authGet(t, tok, "/farms/1/crop-library/picker")
	expectStatus(t, resp, 200)
	picker := decodeMap(t, resp)
	groups, _ := picker["groups"].([]any)
	var foundThumb bool
	for _, g := range groups {
		grp, _ := g.(map[string]any)
		items, _ := grp["items"].([]any)
		for _, raw := range items {
			item, _ := raw.(map[string]any)
			if item["crop_key"] == "san_pedro" {
				if item["image_url"] == nil || item["image_url"] == "" {
					t.Fatalf("picker san_pedro missing image_url: %#v", item)
				}
				foundThumb = true
			}
			if item["crop_key"] == "tomato" && item["image_url"] != nil {
				t.Fatalf("tomato should omit image_url when unset: %#v", item["image_url"])
			}
		}
	}
	if !foundThumb {
		t.Fatal("san_pedro not found in picker groups")
	}
}
