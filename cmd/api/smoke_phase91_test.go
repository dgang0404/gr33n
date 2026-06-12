// Phase 91 — bootstrap template catalog API contract smoke.
package main

import (
	"net/http"
	"testing"
)

func TestPhase91_BootstrapTemplatesContract(t *testing.T) {
	tok := smokeJWT(t)

	resp := authGet(t, tok, "/platform/bootstrap-templates")
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusOK)

	body := decodeMap(t, resp)
	raw, ok := body["templates"].([]any)
	if !ok || len(raw) < 5 {
		t.Fatalf("templates: want >=5, got %d", len(raw))
	}

	foundJadam := false
	for _, item := range raw {
		row, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if row["template_key"] == "jadam_indoor_photoperiod_v1" {
			foundJadam = true
			if row["recommended"] != true {
				t.Fatalf("jadam recommended: %#v", row["recommended"])
			}
			bullets, _ := row["summary_bullets"].([]any)
			if len(bullets) < 4 {
				t.Fatalf("jadam summary_bullets: want >=4, got %d", len(bullets))
			}
		}
	}
	if !foundJadam {
		t.Fatal("expected jadam_indoor_photoperiod_v1 in catalog")
	}
}
