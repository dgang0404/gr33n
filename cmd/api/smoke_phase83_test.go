// Phase 83 WS7 — agronomy seed pack + bootstrap readiness smoke tests.

package main

import (
	"testing"
)

func TestPhase83CultivatorSeedPackPublished(t *testing.T) {
	tok := smokeJWT(t)

	resp := authGet(t, tok, "/commons/catalog/gr33n-cultivator-seed-pack-v1")
	expectStatus(t, resp, 200)
	detail := decodeMap(t, resp)
	if detail["slug"] != "gr33n-cultivator-seed-pack-v1" {
		t.Fatalf("unexpected slug %v", detail["slug"])
	}
	body, ok := detail["body"].(map[string]any)
	if !ok {
		t.Fatalf("expected body object, got %#v", detail["body"])
	}
	if body["kind"] != "agronomy_seed_pack" {
		t.Fatalf("kind: %#v", body["kind"])
	}
	ver, _ := body["platform_catalog_version"].(float64)
	if ver < 4 {
		t.Fatalf("platform_catalog_version want >= 4, got %v", body["platform_catalog_version"])
	}
	counts, ok := body["expected_counts"].(map[string]any)
	if !ok {
		t.Fatal("expected_counts missing")
	}
	if counts["supported_crops"].(float64) < 46 {
		t.Fatalf("expected_counts: %#v", counts)
	}
}
