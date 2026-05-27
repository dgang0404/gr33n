// Phase 30 WS8 — OpenAPI contract + Guardian PR queue acceptance smokes.
package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

// TestPhase30WS8_CapabilitiesAndProposalListContract verifies public capability
// flags and the inbox list shape (risk_tier on each item when proposals exist).
func TestPhase30WS8_CapabilitiesAndProposalListContract(t *testing.T) {
	capResp := get(t, "/capabilities")
	defer capResp.Body.Close()
	expectStatus(t, capResp, http.StatusOK)
	cap := decodeMap(t, capResp)
	if _, ok := cap["ai_enabled"]; !ok {
		t.Fatalf("capabilities missing ai_enabled: %#v", cap)
	}
	if _, ok := cap["vision_chat_enabled"]; !ok {
		t.Fatalf("capabilities missing vision_chat_enabled: %#v", cap)
	}

	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	tok := smokeJWT(t)
	listResp := authGet(t, tok, "/v1/chat/proposals?farm_id=1&status=pending&limit=10")
	defer listResp.Body.Close()
	if listResp.StatusCode == http.StatusInternalServerError {
		body := readBodyPreview(listResp)
		if strings.Contains(body, "guardian_action_proposals") {
			t.Skip("guardian_action_proposals table missing — apply Phase 29/30 migrations")
		}
	}
	expectStatus(t, listResp, http.StatusOK)
	payload := decodeMap(t, listResp)
	for _, key := range []string{"proposals", "total", "limit", "offset"} {
		if _, ok := payload[key]; !ok {
			t.Fatalf("list response missing %q: %#v", key, payload)
		}
	}
	proposals, _ := payload["proposals"].([]any)
	for i, raw := range proposals {
		p, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("proposal[%d] not object", i)
		}
		if _, ok := p["risk_tier"]; !ok {
			t.Fatalf("proposal[%d] missing risk_tier: %#v", i, p)
		}
		switch p["risk_tier"].(string) {
		case "low", "medium", "high":
		default:
			t.Fatalf("proposal[%d] invalid risk_tier %#v", i, p["risk_tier"])
		}
	}
}

// TestPhase30WS8_ZonePhotosRoutesDocumented is a light contract check that zone
// photo routes respond (upload/list exercised in WS5 smoke).
func TestPhase30WS8_ZonePhotosRoutesRespond(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	tok := smokeJWT(t)
	zonesResp := authGet(t, tok, "/farms/1/zones")
	defer zonesResp.Body.Close()
	expectStatus(t, zonesResp, http.StatusOK)
	zones := decodeSlice(t, zonesResp)
	if len(zones) == 0 {
		t.Skip("no zones on farm 1")
	}
	zoneID := int64(zones[0].(map[string]any)["id"].(float64))
	listResp := authGet(t, tok, "/zones/"+strconv.FormatInt(zoneID, 10)+"/photos")
	defer listResp.Body.Close()
	expectStatus(t, listResp, http.StatusOK)
	var listBody map[string]any
	_ = json.NewDecoder(listResp.Body).Decode(&listBody)
	if _, ok := listBody["photos"]; !ok {
		t.Fatalf("expected photos key: %#v", listBody)
	}
}
