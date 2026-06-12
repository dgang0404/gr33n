// Phase 84 WS-J — commons crop catalog + field guide read API smoke tests.

package main

import (
	"testing"
)

func TestCommonsCropCatalogAndFieldGuides(t *testing.T) {
	tok := smokeJWT(t)

	resp := authGet(t, tok, "/commons/crop-catalog")
	expectStatus(t, resp, 200)
	list := decodeMap(t, resp)
	if list["entries"] == nil || list["aliases"] == nil {
		t.Fatalf("expected entries and aliases, got %#v", list)
	}
	entries, ok := list["entries"].([]any)
	if !ok || len(entries) < 40 {
		t.Fatalf("expected many catalog entries, got %d", len(entries))
	}

	resp = authGet(t, tok, "/commons/crop-catalog/tomato")
	expectStatus(t, resp, 200)
	tomato := decodeMap(t, resp)
	if tomato["crop_key"] != "tomato" || tomato["supported"] != true {
		t.Fatalf("unexpected tomato entry: %#v", tomato)
	}
	if tomato["crop_profile_id"] == nil {
		t.Fatal("expected crop_profile_id for supported tomato")
	}

	resp = authGet(t, tok, "/commons/crop-catalog/ramps")
	expectStatus(t, resp, 200)
	ramps := decodeMap(t, resp)
	if ramps["supported"] != false || ramps["crop_profile_id"] != nil {
		t.Fatalf("unsupported ramps should have no profile id: %#v", ramps)
	}

	resp = authGet(t, tok, "/commons/agronomy-field-guides?crop_key=tomato")
	expectStatus(t, resp, 200)
	guides := decodeSlice(t, resp)
	if len(guides) < 1 {
		t.Fatal("expected at least one tomato field guide")
	}

	resp = authGet(t, tok, "/commons/agronomy-field-guides/crop-tomato-nutrition")
	expectStatus(t, resp, 200)
	guide := decodeMap(t, resp)
	if guide["body_md"] == nil || guide["slug"] != "crop-tomato-nutrition" {
		t.Fatalf("expected full guide body: %#v", guide)
	}
}
