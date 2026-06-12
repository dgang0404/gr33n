// Phase 82 — Guardian plant intelligence closure smokes (OC-82).
package main

import (
	"net/http"
	"strings"
	"testing"

	"gr33n-api/internal/rag/synthesis"
)

func TestPhase82_CatalogPickerAndZeroChunkPolicy(t *testing.T) {
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/farms/1/crop-library/picker")
	expectStatus(t, resp, http.StatusOK)
	body := decodeMap(t, resp)
	resp.Body.Close()
	counts, _ := body["counts"].(map[string]any)
	withTargets, _ := counts["with_targets"].(float64)
	if withTargets < 46 {
		t.Fatalf("want >= 46 with_targets, got %v", withTargets)
	}
	groups, _ := body["groups"].([]any)
	if len(groups) < 3 {
		t.Fatalf("want grouped picker, got %d groups", len(groups))
	}

	stripped := synthesis.StripOrphanCitationRefs("Feed at [1] with 2% EC [2].", 0)
	if strings.Contains(stripped, "[1]") {
		t.Fatalf("zero-chunk policy should strip orphan refs: %q", stripped)
	}
	if !strings.Contains(synthesis.ZeroChunkGuardBlock(), "lookup_crop_targets") {
		t.Fatal("ZeroChunkGuardBlock missing read-tool guidance")
	}
}

func TestPhase82_UnsupportedCropInCatalog(t *testing.T) {
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/commons/crop-catalog/ramps")
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusOK)
	body := decodeMap(t, resp)
	if body["supported"] != false {
		t.Fatalf("ramps should be unsupported: %#v", body)
	}
}
