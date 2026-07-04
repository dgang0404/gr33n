// Phase 32 WS2 — grow create tools propose→confirm smokes.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func insertGuardianProposal(t *testing.T, toolID string, args map[string]any, summary string) string {
	return insertGuardianProposalWithRisk(t, toolID, args, summary, "medium")
}

func insertGuardianProposalWithRisk(t *testing.T, toolID string, args map[string]any, summary, riskTier string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	uid := uuid.MustParse(smokeDevUserUUID)
	raw, err := json.Marshal(args)
	if err != nil {
		t.Fatalf("marshal args: %v", err)
	}
	var proposalID string
	err = testPool.QueryRow(ctx, `
INSERT INTO gr33ncore.guardian_action_proposals
    (user_id, farm_id, tool_id, args, summary, risk_tier, expires_at)
VALUES ($1, 1, $2, $3::jsonb, $4, $5, NOW() + INTERVAL '10 minutes')
RETURNING proposal_id::text`, uid, toolID, raw, summary, riskTier).Scan(&proposalID)
	if err != nil {
		t.Fatalf("insert proposal: %v", err)
	}
	return proposalID
}

func TestPhase32WS2_CreatePlantConfirm(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// "mint" — a crop_key the Phase 124 demo seed doesn't touch — so this
	// test's cleanup never soft-deletes a permanently-seeded plant.
	proposalID := insertGuardianProposal(t, "create_plant", map[string]any{
		"crop_key":            "mint",
		"variety_or_cultivar": "Spearmint",
	}, "Create catalog plant: mint")
	t.Cleanup(func() {
		_, _ = testPool.Exec(context.Background(),
			`UPDATE gr33ncrops.plants SET deleted_at = NOW() WHERE farm_id = 1 AND crop_key = 'mint'`)
	})

	tok := smokeJWT(t)
	confirmResp := authPost(t, tok, "/v1/chat/confirm", map[string]string{"proposal_id": proposalID})
	defer confirmResp.Body.Close()
	expectStatus(t, confirmResp, http.StatusOK)

	var confirmBody struct {
		Result map[string]any `json:"result"`
	}
	decodeJSON(t, confirmResp.Body, &confirmBody)
	plantID, _ := confirmBody.Result["plant_id"].(float64)
	if plantID == 0 {
		t.Fatalf("missing plant_id: %+v", confirmBody.Result)
	}
	if confirmBody.Result["crop_key"] != "mint" {
		t.Fatalf("expected crop_key mint, got %v", confirmBody.Result["crop_key"])
	}

	var gotKey string
	err := testPool.QueryRow(ctx, `
SELECT crop_key FROM gr33ncrops.plants WHERE id = $1 AND farm_id = 1 AND deleted_at IS NULL`,
		int64(plantID)).Scan(&gotKey)
	if err != nil {
		t.Fatalf("plant row: %v", err)
	}
	if gotKey != "mint" {
		t.Fatalf("crop_key %q want mint", gotKey)
	}

	assertGuardianToolAudit(t, ctx, "create_plant", proposalID)
}

func TestPhase32WS2_CreateCropCycleConfirm(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	var zoneID int64
	err := testPool.QueryRow(ctx, `
INSERT INTO gr33ncore.zones (farm_id, name, description, zone_type)
VALUES (1, $1, 'Phase 32 WS2 smoke zone', 'indoor')
RETURNING id`, fmt.Sprintf("WS2 smoke zone %d", time.Now().UnixNano())).Scan(&zoneID)
	if err != nil {
		t.Fatalf("insert zone: %v", err)
	}
	t.Cleanup(func() {
		_, _ = testPool.Exec(context.Background(), `DELETE FROM gr33nfertigation.crop_cycles WHERE zone_id = $1`, zoneID)
		_, _ = testPool.Exec(context.Background(), `DELETE FROM gr33ncore.zones WHERE id = $1`, zoneID)
	})

	plantProposal := insertGuardianProposal(t, "create_plant", map[string]any{
		"crop_key": "mint",
	}, "Create plant for cycle smoke")
	t.Cleanup(func() {
		_, _ = testPool.Exec(context.Background(),
			`UPDATE gr33ncrops.plants SET deleted_at = NOW() WHERE farm_id = 1 AND crop_key = 'mint'`)
	})
	tok := smokeJWT(t)
	plantConfirm := authPost(t, tok, "/v1/chat/confirm", map[string]string{"proposal_id": plantProposal})
	expectStatus(t, plantConfirm, http.StatusOK)
	var plantBody struct {
		Result map[string]any `json:"result"`
	}
	decodeJSON(t, plantConfirm.Body, &plantBody)
	plantConfirm.Body.Close()
	plantID := int64(plantBody.Result["plant_id"].(float64))

	proposalID := insertGuardianProposal(t, "create_crop_cycle", map[string]any{
		"zone_id":       zoneID,
		"plant_id":      plantID,
		"name":          "Basil — smoke",
		"batch_label":   "heartleaf",
		"current_stage": "vegetative",
		"started_at":    time.Now().UTC().Format("2006-01-02"),
	}, "Start crop cycle in smoke zone")
	confirmResp := authPost(t, tok, "/v1/chat/confirm", map[string]string{"proposal_id": proposalID})
	defer confirmResp.Body.Close()
	expectStatus(t, confirmResp, http.StatusOK)

	var confirmBody struct {
		Result map[string]any `json:"result"`
	}
	decodeJSON(t, confirmResp.Body, &confirmBody)
	cycleID, _ := confirmBody.Result["crop_cycle_id"].(float64)
	if cycleID == 0 {
		t.Fatalf("missing crop_cycle_id: %+v", confirmBody.Result)
	}

	var active bool
	err = testPool.QueryRow(ctx, `
SELECT is_active FROM gr33nfertigation.crop_cycles WHERE id = $1 AND farm_id = 1`, int64(cycleID)).Scan(&active)
	if err != nil {
		t.Fatalf("cycle row: %v", err)
	}
	if !active {
		t.Fatal("expected active cycle")
	}

	assertGuardianToolAudit(t, ctx, "create_crop_cycle", proposalID)
}

func TestPhase32WS2_CreateCropCycleRejectsActiveZoneConflict(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var zoneID int64
	err := testPool.QueryRow(ctx, `
SELECT z.id FROM gr33ncore.zones z
JOIN gr33nfertigation.crop_cycles cc ON cc.zone_id = z.id AND cc.is_active = TRUE
WHERE z.farm_id = 1
LIMIT 1`).Scan(&zoneID)
	if err != nil {
		t.Skip("no seeded active cycle zone")
	}

	proposalID := insertGuardianProposal(t, "create_crop_cycle", map[string]any{
		"zone_id":           zoneID,
		"name":              "Duplicate cycle",
		"strain_or_variety": "test",
		"current_stage":     "vegetative",
		"started_at":        time.Now().UTC().Format("2006-01-02"),
	}, "Should fail — zone already has active cycle")

	tok := smokeJWT(t)
	confirmResp := authPost(t, tok, "/v1/chat/confirm", map[string]string{"proposal_id": proposalID})
	defer confirmResp.Body.Close()
	if confirmResp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 conflict, got %d: %s", confirmResp.StatusCode, readBodyPreview(confirmResp))
	}
}

func TestPhase32WS2_CreateFertigationProgramConfirm(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var zoneID int64
	err := testPool.QueryRow(ctx, `SELECT id FROM gr33ncore.zones WHERE farm_id = 1 AND name = 'Veg Room'`).Scan(&zoneID)
	if err != nil {
		t.Fatalf("veg room zone: %v", err)
	}

	progName := fmt.Sprintf("Guardian smoke program %d", time.Now().UnixNano())
	proposalID := insertGuardianProposal(t, "create_fertigation_program", map[string]any{
		"name":                progName,
		"target_zone_id":      zoneID,
		"total_volume_liters": 0.5,
		"ec_trigger_low":      0.8,
		"ph_trigger_low":      5.8,
		"ph_trigger_high":     6.5,
		"is_active":           true,
	}, "Create fertigation program: "+progName)
	t.Cleanup(func() {
		_, _ = testPool.Exec(context.Background(),
			`UPDATE gr33nfertigation.programs SET deleted_at = NOW() WHERE farm_id = 1 AND name = $1`, progName)
	})

	tok := smokeJWT(t)
	confirmResp := authPost(t, tok, "/v1/chat/confirm", map[string]string{"proposal_id": proposalID})
	defer confirmResp.Body.Close()
	expectStatus(t, confirmResp, http.StatusOK)

	var confirmBody struct {
		Result map[string]any `json:"result"`
	}
	decodeJSON(t, confirmResp.Body, &confirmBody)
	programID, _ := confirmBody.Result["program_id"].(float64)
	if programID == 0 {
		t.Fatalf("missing program_id: %+v", confirmBody.Result)
	}

	var gotName string
	err = testPool.QueryRow(ctx, `
SELECT name FROM gr33nfertigation.programs WHERE id = $1 AND farm_id = 1 AND deleted_at IS NULL`,
		int64(programID)).Scan(&gotName)
	if err != nil {
		t.Fatalf("program row: %v", err)
	}
	if gotName != progName {
		t.Fatalf("name %q want %q", gotName, progName)
	}

	assertGuardianToolAudit(t, ctx, "create_fertigation_program", proposalID)
}

func assertGuardianToolAudit(t *testing.T, ctx context.Context, toolID, proposalID string) {
	t.Helper()
	var auditCount int
	err := testPool.QueryRow(ctx, `
SELECT COUNT(*) FROM gr33ncore.user_activity_log
WHERE action_type = 'guardian_tool_executed'
  AND details->>'tool_id' = $1
  AND details->>'proposal_id' = $2`, toolID, proposalID).Scan(&auditCount)
	if err != nil {
		t.Fatalf("audit query: %v", err)
	}
	if auditCount < 1 {
		t.Fatalf("expected guardian_tool_executed audit for %s", toolID)
	}
}
