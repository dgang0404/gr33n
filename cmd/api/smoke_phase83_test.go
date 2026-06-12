// Phase 83 WS7 — agronomy seed pack + bootstrap readiness smoke tests.

package main

import (
	"testing"
)

func TestPhase83_CropProfileOverridePutDelete(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	tok := smokeJWT(t)
	const cropKey = "cannabis"

	// Ensure clean slate
	del := authDelete(t, tok, "/farms/1/crop-profiles/"+cropKey)
	del.Body.Close()

	get0 := authGet(t, tok, "/farms/1/crop-profiles/"+cropKey)
	expectStatus(t, get0, 200)
	before := decodeMap(t, get0)
	get0.Body.Close()
	stagesBefore, _ := before["stages"].([]any)
	if len(stagesBefore) == 0 {
		t.Fatal("expected cannabis builtin stages")
	}

	// Build PUT body from effective profile with tweaked EC max on first stage
	first, _ := stagesBefore[0].(map[string]any)
	stageName, _ := first["stage"].(string)
	if stageName == "" {
		t.Fatal("missing stage on first row")
	}
	putBody := map[string]any{
		"display_name": before["display_name"],
		"stages": []map[string]any{{
			"stage":   stageName,
			"ec_min":  first["ec_min"],
			"ec_max":  9.99,
			"ec_target": first["ec_target"],
		}},
	}
	put := authPut(t, tok, "/farms/1/crop-profiles/"+cropKey, putBody)
	expectStatus(t, put, 200)
	afterPut := decodeMap(t, put)
	put.Body.Close()
	if afterPut["is_builtin"] == true {
		t.Fatal("PUT should create farm override row")
	}

	get1 := authGet(t, tok, "/farms/1/crop-profiles/"+cropKey)
	expectStatus(t, get1, 200)
	got := decodeMap(t, get1)
	get1.Body.Close()
	stagesGot, _ := got["stages"].([]any)
	if len(stagesGot) != 1 {
		t.Fatalf("override should have 1 stage, got %d", len(stagesGot))
	}
	s0, _ := stagesGot[0].(map[string]any)
	ecMax, _ := s0["ec_max"].(float64)
	if ecMax != 9.99 {
		t.Fatalf("ec_max want 9.99, got %v", s0["ec_max"])
	}

	del2 := authDelete(t, tok, "/farms/1/crop-profiles/"+cropKey)
	expectStatus(t, del2, 204)
	del2.Body.Close()

	get2 := authGet(t, tok, "/farms/1/crop-profiles/"+cropKey)
	expectStatus(t, get2, 200)
	reverted := decodeMap(t, get2)
	get2.Body.Close()
	if reverted["is_builtin"] != true {
		t.Fatal("after DELETE expected builtin profile")
	}
}

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
