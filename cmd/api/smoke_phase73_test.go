// Phase 73 — Guardian PR discoverability: empty-zone propose, server dismiss, read-tool widening.
package main

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestPhase73_SuggestEmptyZoneAndDismiss(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	tok := smokeJWT(t)
	zoneName := uniqueName("phase73_empty")

	createResp := authPost(t, tok, "/farms/1/zones", map[string]any{
		"name":             zoneName,
		"farm_id":          1,
		"zone_type":        "indoor",
		"description":      "Phase 73 empty-zone smoke",
		"environment_type": "soil",
	})
	defer createResp.Body.Close()
	expectStatus(t, createResp, http.StatusCreated)
	zone := decodeMap(t, createResp)
	zoneID := int64(zone["id"].(float64))

	suggestResp := authPost(t, tok, "/v1/chat/proposals/suggest-empty-zone", map[string]any{
		"farm_id": 1,
		"zone_id": zoneID,
	})
	defer suggestResp.Body.Close()
	expectStatus(t, suggestResp, http.StatusCreated)
	prop := decodeMap(t, suggestResp)
	proposalID, _ := prop["proposal_id"].(string)
	if proposalID == "" {
		t.Fatalf("missing proposal_id: %#v", prop)
	}
	if prop["tool"] != "apply_grow_setup_pack" {
		t.Fatalf("tool %v want apply_grow_setup_pack", prop["tool"])
	}

	listResp := authGet(t, tok, fmt.Sprintf("/v1/chat/proposals?farm_id=1&status=pending&limit=50"))
	defer listResp.Body.Close()
	expectStatus(t, listResp, http.StatusOK)
	listBody := decodeMap(t, listResp)
	found := false
	for _, raw := range listBody["proposals"].([]any) {
		row, _ := raw.(map[string]any)
		if row["proposal_id"] == proposalID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("proposal %s not in pending list", proposalID)
	}

	dismissResp := authPost(t, tok, fmt.Sprintf("/v1/chat/proposals/%s/dismiss", proposalID), nil)
	defer dismissResp.Body.Close()
	expectStatus(t, dismissResp, http.StatusOK)
	dismissed := decodeMap(t, dismissResp)
	if dismissed["status"] != "dismissed" {
		t.Fatalf("status %v want dismissed", dismissed["status"])
	}

	listAfter := authGet(t, tok, fmt.Sprintf("/v1/chat/proposals?farm_id=1&status=pending&limit=50"))
	defer listAfter.Body.Close()
	afterBody := decodeMap(t, listAfter)
	for _, raw := range afterBody["proposals"].([]any) {
		row, _ := raw.(map[string]any)
		if row["proposal_id"] == proposalID {
			t.Fatalf("dismissed proposal %s still in pending list", proposalID)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var status string
	err := testPool.QueryRow(ctx, `
SELECT status FROM gr33ncore.guardian_action_proposals WHERE proposal_id = $1`, proposalID).Scan(&status)
	if err != nil {
		t.Fatalf("read proposal status: %v", err)
	}
	if status != "dismissed" {
		t.Fatalf("db status %q want dismissed", status)
	}
}

func TestPhase73_SiteWeatherIntentBroadened(t *testing.T) {
	tok := smokeJWT(t)
	chatResp := authPost(t, tok, "/v1/chat", map[string]any{
		"message": "do I need supplemental light today?",
		"farm_id": 1,
		"stream":  false,
	})
	defer chatResp.Body.Close()
	if chatResp.StatusCode == http.StatusServiceUnavailable {
		t.Skip("LLM not configured — set LLM_BASE_URL and LLM_MODEL for full E2E")
	}
	if chatResp.StatusCode != http.StatusOK {
		t.Fatalf("chat status %d: %s", chatResp.StatusCode, readBodyPreview(chatResp))
	}
	body := readBodyPreview(chatResp)
	lower := body
	if !(containsAny(lower, "light", "DLI", "daylight", "supplemental", "Settings", "coordinates", "latitude")) {
		t.Fatalf("expected grounded weather/light answer, got: %s", truncate(body, 400))
	}
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if sub != "" && stringContainsFold(s, sub) {
			return true
		}
	}
	return false
}

func stringContainsFold(s, sub string) bool {
	return len(sub) == 0 || len(s) >= len(sub) && (func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if equalFoldASCII(s[i:i+len(sub)], sub) {
				return true
			}
		}
		return false
	})()
}

func equalFoldASCII(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
