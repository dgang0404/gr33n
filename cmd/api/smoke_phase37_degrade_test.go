// Phase 37 WS1 — field degrade health flags.
package main

import (
	"net/http"
	"testing"
)

func TestPhase37_ChatHealthFieldDegradeReady(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/v1/chat/health?farm_id=1")
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusOK)
	body := decodeMap(t, resp)
	if body["procedures_available"] != true {
		t.Fatalf("procedures_available: %v", body["procedures_available"])
	}
}
