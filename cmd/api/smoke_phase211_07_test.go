// Phase 211.07 — seed-pending proposal for browser Pending Confirm/Dismiss (no LLM).
package main

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestPhase21107_SeedPendingConfirmAndDismiss(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	tok := smokeJWT(t)
	stamp := time.Now().UnixNano()
	confirmTitle := fmt.Sprintf("E2E confirm %d", stamp)
	dismissTitle := fmt.Sprintf("E2E dismiss %d", stamp)

	confirmResp := authPost(t, tok, "/v1/chat/proposals/seed-pending", map[string]any{
		"farm_id": 1,
		"title":   confirmTitle,
	})
	defer confirmResp.Body.Close()
	expectStatus(t, confirmResp, http.StatusCreated)
	confirmProp := decodeMap(t, confirmResp)
	confirmID, _ := confirmProp["proposal_id"].(string)
	if confirmID == "" || confirmProp["tool"] != "create_task" {
		t.Fatalf("confirm seed: %#v", confirmProp)
	}

	dismissResp := authPost(t, tok, "/v1/chat/proposals/seed-pending", map[string]any{
		"farm_id": 1,
		"title":   dismissTitle,
	})
	defer dismissResp.Body.Close()
	expectStatus(t, dismissResp, http.StatusCreated)
	dismissProp := decodeMap(t, dismissResp)
	dismissID, _ := dismissProp["proposal_id"].(string)
	if dismissID == "" {
		t.Fatalf("dismiss seed: %#v", dismissProp)
	}

	applyResp := authPost(t, tok, "/v1/chat/confirm", map[string]string{"proposal_id": confirmID})
	defer applyResp.Body.Close()
	expectStatus(t, applyResp, http.StatusOK)

	goneResp := authPost(t, tok, fmt.Sprintf("/v1/chat/proposals/%s/dismiss", dismissID), nil)
	defer goneResp.Body.Close()
	expectStatus(t, goneResp, http.StatusOK)
	dismissed := decodeMap(t, goneResp)
	if dismissed["status"] != "dismissed" {
		t.Fatalf("status %v want dismissed", dismissed["status"])
	}
}
