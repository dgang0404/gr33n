// Phase 101 — Guardian write tools require catalog crop_key (same rules as UI Phase 85).

package main

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestPhase101_GuardianCreatePlantRequiresCropKey(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	proposalID := insertGuardianProposal(t, "create_plant", map[string]any{
		"display_name": "Mystery Crop",
	}, "Create plant without crop_key")
	tok := smokeJWT(t)
	resp := authPost(t, tok, "/v1/chat/confirm", map[string]string{"proposal_id": proposalID})
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusBadRequest)
}

func TestPhase101_GuardianCreatePlantUpsertAndUnsupported(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}

	tok := smokeJWT(t)

	// "cucumber" — a crop_key the Phase 124 demo seed doesn't touch — so this
	// test's cleanup below never soft-deletes a permanently-seeded plant.
	proposalID := insertGuardianProposal(t, "create_plant", map[string]any{
		"crop_key":            "cucumber",
		"variety_or_cultivar": "Marketmore",
	}, "Create cucumber plant")
	resp := authPost(t, tok, "/v1/chat/confirm", map[string]string{"proposal_id": proposalID})
	expectStatus(t, resp, http.StatusOK)
	var firstBody struct {
		Result map[string]any `json:"result"`
	}
	decodeJSON(t, resp.Body, &firstBody)
	resp.Body.Close()
	firstID := int64(firstBody.Result["plant_id"].(float64))

	proposalID = insertGuardianProposal(t, "create_plant", map[string]any{
		"crop_key":            "cucumber",
		"variety_or_cultivar": "Persian",
	}, "Upsert cucumber variety")
	resp = authPost(t, tok, "/v1/chat/confirm", map[string]string{"proposal_id": proposalID})
	expectStatus(t, resp, http.StatusOK)
	var secondBody struct {
		Result map[string]any `json:"result"`
	}
	decodeJSON(t, resp.Body, &secondBody)
	resp.Body.Close()
	secondID := int64(secondBody.Result["plant_id"].(float64))
	if firstID != secondID {
		t.Fatalf("duplicate crop_key should upsert same plant: %d vs %d", secondID, firstID)
	}

	proposalID = insertGuardianProposal(t, "create_plant", map[string]any{
		"crop_key": "ramps",
	}, "Unsupported crop")
	resp = authPost(t, tok, "/v1/chat/confirm", map[string]string{"proposal_id": proposalID})
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusBadRequest)

	t.Cleanup(func() {
		// Use a fresh context — the outer ctx's `defer cancel()` has already
		// fired by the time t.Cleanup callbacks run (they run after the test
		// function body, including its own defers, returns), so reusing ctx
		// here made this cleanup silently no-op on a cancelled context.
		cCtx, cCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cCancel()
		_, _ = testPool.Exec(cCtx, `UPDATE gr33ncrops.plants SET deleted_at = NOW() WHERE farm_id = 1 AND crop_key = 'cucumber'`)
	})
}

func TestPhase101_GuardianCreatePlantRejectsDisplayName(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	proposalID := insertGuardianProposal(t, "create_plant", map[string]any{
		"crop_key":     "spinach",
		"display_name": "Custom Spinach Label",
	}, "Reject client display_name")
	tok := smokeJWT(t)
	resp := authPost(t, tok, "/v1/chat/confirm", map[string]string{"proposal_id": proposalID})
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusBadRequest)
}
