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
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	tok := smokeJWT(t)

	proposalID := insertGuardianProposal(t, "create_plant", map[string]any{
		"crop_key":            "tomato",
		"variety_or_cultivar": "Cherokee Purple",
	}, "Create tomato plant")
	resp := authPost(t, tok, "/v1/chat/confirm", map[string]string{"proposal_id": proposalID})
	expectStatus(t, resp, http.StatusOK)
	var firstBody struct {
		Result map[string]any `json:"result"`
	}
	decodeJSON(t, resp.Body, &firstBody)
	resp.Body.Close()
	firstID := int64(firstBody.Result["plant_id"].(float64))

	proposalID = insertGuardianProposal(t, "create_plant", map[string]any{
		"crop_key":            "tomato",
		"variety_or_cultivar": "Roma",
	}, "Upsert tomato variety")
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
		_, _ = testPool.Exec(ctx, `UPDATE gr33ncrops.plants SET deleted_at = NOW() WHERE farm_id = 1 AND crop_key = 'tomato'`)
	})
}

func TestPhase101_GuardianCreatePlantRejectsDisplayName(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	proposalID := insertGuardianProposal(t, "create_plant", map[string]any{
		"crop_key":     "basil",
		"display_name": "Custom Basil Label",
	}, "Reject client display_name")
	tok := smokeJWT(t)
	resp := authPost(t, tok, "/v1/chat/confirm", map[string]string{"proposal_id": proposalID})
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusBadRequest)
}
