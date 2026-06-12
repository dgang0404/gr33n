// Phase 97 — structured truth beats stale RAG narrative EC.
package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"gr33n-api/internal/farmguardian"
	"gr33n-api/internal/rag/synthesis"
	db "gr33n-api/internal/db"
)

func TestPhase97_FarmOverrideECInLookupCropTargets(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	tok := smokeJWT(t)
	farmID := int64(1)
	const overrideTarget = 2.99

	t.Cleanup(func() {
		resp := authDelete(t, tok, fmt.Sprintf("/farms/%d/crop-profiles/cannabis", farmID))
		if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
			t.Logf("override cleanup: %d", resp.StatusCode)
		}
	})

	resp := authPut(t, tok, fmt.Sprintf("/farms/%d/crop-profiles/cannabis", farmID), map[string]any{
		"display_name": "Cannabis (phase97)",
		"source":       "phase97 structured truth smoke",
		"stages": []map[string]any{
			{
				"stage":     "early_flower",
				"ec_min":    2.8,
				"ec_target": overrideTarget,
				"ec_max":    3.1,
			},
		},
	})
	expectStatus(t, resp, http.StatusOK)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	q := db.New(testPool)
	block, err := farmguardian.LookupCropTargets(ctx, q, farmID, "What EC target for cannabis early flower?", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(block, fmt.Sprintf("%.2f", overrideTarget)) {
		t.Fatalf("override ec_target %.2f not in block: %s", overrideTarget, block)
	}
}

func TestPhase97_StripStaleRAGNutrientNumbers(t *testing.T) {
	stale := "Targets ramp to ~1.6–2.0 mS/cm in mid-flower."
	stripped := synthesis.StripNutrientNumbersFromChunks([]db.SearchRagNearestNeighborsFilteredRow{
		{SourceType: "field_guide", ContentText: stale},
	})[0].ContentText
	if strings.Contains(stripped, "1.6") || strings.Contains(stripped, "2.0") {
		t.Fatalf("stale EC should be stripped: %q", stripped)
	}
	if !strings.Contains(stripped, "lookup_crop_targets") {
		t.Fatal("expected structured-truth hint")
	}
}
