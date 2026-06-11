// Phase 64 — crop knowledge base (builtin profiles + list API).
package main

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestPhase64_CropProfilesListAndCannabisFlowerEC(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var ecMin, ecMax float64
	err := testPool.QueryRow(ctx, `
SELECT s.ec_min::float8, s.ec_max::float8
FROM gr33ncrops.crop_profiles p
JOIN gr33ncrops.crop_profile_stages s ON s.crop_profile_id = p.id
WHERE p.is_builtin = TRUE AND p.crop_key = 'cannabis' AND s.stage = 'early_flower'
LIMIT 1`).Scan(&ecMin, &ecMax)
	if err != nil {
		t.Skip("phase 64 seed missing — run migration 20260610_phase64_crop_knowledge_base.sql")
	}
	if ecMin < 1.0 || ecMax > 2.5 {
		t.Fatalf("cannabis early_flower EC range unexpected: %v–%v", ecMin, ecMax)
	}

	tok := smokeJWT(t)
	resp := authGet(t, tok, "/farms/1/crop-profiles")
	expectStatus(t, resp, http.StatusOK)
	body := decodeSlice(t, resp)
	resp.Body.Close()
	if len(body) < 13 {
		t.Fatalf("want >= 13 profiles, got %d", len(body))
	}
}

func TestPhase82_CropLibraryPicker(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/farms/1/crop-library/picker")
	expectStatus(t, resp, http.StatusOK)
	body := decodeMap(t, resp)
	resp.Body.Close()
	counts, _ := body["counts"].(map[string]any)
	if counts == nil {
		t.Fatal("picker response missing counts")
	}
	withTargets, _ := counts["with_targets"].(float64)
	if withTargets < 13 {
		t.Fatalf("want >= 13 with_targets, got %v", withTargets)
	}
}
