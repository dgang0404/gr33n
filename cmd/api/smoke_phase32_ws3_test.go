// Phase 32 WS3 — apply_grow_setup_pack transactional confirm smoke.
package main

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func housePlantSetupPackArgs(zoneID int64, zoneName, plantName string) map[string]any {
	return map[string]any{
		"profile":   "house_plant",
		"zone_id":   zoneID,
		"zone_name": zoneName,
		"plant": map[string]any{
			"display_name":        plantName,
			"variety_or_cultivar": "heartleaf",
			"notes":               "RO water only",
		},
		"cycle": map[string]any{
			"name":          plantName + " — " + zoneName,
			"current_stage": "vegetative",
			"started_at":    time.Now().UTC().Format("2006-01-02"),
		},
		"program": map[string]any{
			"name":                plantName + " light feed",
			"total_volume_liters": 0.5,
			"ec_trigger_low":      0.8,
			"ph_trigger_low":      5.8,
			"ph_trigger_high":     6.5,
			"is_active":           true,
		},
		"optional_task": map[string]any{
			"title": "Monitor new " + plantName + " — first two weeks",
		},
	}
}

func TestPhase32WS3_ApplyGrowSetupPackConfirm(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	zoneName := fmt.Sprintf("Living Room WS3 %d", time.Now().UnixNano())
	plantName := fmt.Sprintf("Philodendron %d", time.Now().UnixNano())

	var zoneID int64
	err := testPool.QueryRow(ctx, `
INSERT INTO gr33ncore.zones (farm_id, name, description, zone_type)
VALUES (1, $1, 'Phase 32 WS3 smoke zone', 'indoor')
RETURNING id`, zoneName).Scan(&zoneID)
	if err != nil {
		t.Fatalf("insert zone: %v", err)
	}
	t.Cleanup(func() {
		c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, _ = testPool.Exec(c, `DELETE FROM gr33ncore.tasks WHERE farm_id = 1 AND title LIKE $1`, "Monitor new "+plantName+"%")
		_, _ = testPool.Exec(c, `UPDATE gr33nfertigation.programs SET deleted_at = NOW() WHERE farm_id = 1 AND name = $1`, plantName+" light feed")
		_, _ = testPool.Exec(c, `DELETE FROM gr33nfertigation.crop_cycles WHERE zone_id = $1`, zoneID)
		_, _ = testPool.Exec(c, `UPDATE gr33ncrops.plants SET deleted_at = NOW() WHERE farm_id = 1 AND display_name = $1`, plantName)
		_, _ = testPool.Exec(c, `DELETE FROM gr33ncore.zones WHERE id = $1`, zoneID)
	})

	packArgs := housePlantSetupPackArgs(zoneID, zoneName, plantName)
	proposalID := insertGuardianProposalWithRisk(t, "apply_grow_setup_pack", packArgs,
		fmt.Sprintf("Setup pack: %s in %s", plantName, zoneName), "high")
	t.Cleanup(func() {
		_, _ = testPool.Exec(context.Background(),
			`DELETE FROM gr33ncore.guardian_action_proposals WHERE proposal_id = $1`, proposalID)
	})

	tok := smokeJWT(t)
	confirmResp := authPost(t, tok, "/v1/chat/confirm", map[string]string{"proposal_id": proposalID})
	defer confirmResp.Body.Close()
	expectStatus(t, confirmResp, http.StatusOK)

	var confirmBody struct {
		Result map[string]any `json:"result"`
	}
	decodeJSON(t, confirmResp.Body, &confirmBody)

	plantBlock, _ := confirmBody.Result["plant"].(map[string]any)
	cycleBlock, _ := confirmBody.Result["cycle"].(map[string]any)
	programBlock, _ := confirmBody.Result["program"].(map[string]any)
	if plantBlock == nil || cycleBlock == nil || programBlock == nil {
		t.Fatalf("missing setup pack result sections: %+v", confirmBody.Result)
	}
	if confirmBody.Result["primary_program_linked"] != true {
		t.Fatal("expected primary_program_linked true")
	}

	plantID, _ := plantBlock["plant_id"].(float64)
	cycleID, _ := cycleBlock["crop_cycle_id"].(float64)
	programID, _ := programBlock["program_id"].(float64)
	if plantID == 0 || cycleID == 0 || programID == 0 {
		t.Fatalf("missing ids plant=%v cycle=%v program=%v", plantID, cycleID, programID)
	}

	var linkedProgramID *int64
	err = testPool.QueryRow(ctx, `
SELECT primary_program_id FROM gr33nfertigation.crop_cycles WHERE id = $1`, int64(cycleID)).Scan(&linkedProgramID)
	if err != nil {
		t.Fatalf("cycle row: %v", err)
	}
	if linkedProgramID == nil || *linkedProgramID != int64(programID) {
		t.Fatalf("primary_program_id = %v want %d", linkedProgramID, int64(programID))
	}

	var taskCount int
	_ = testPool.QueryRow(ctx, `
SELECT COUNT(*) FROM gr33ncore.tasks
WHERE farm_id = 1 AND title LIKE $1`, "Monitor new "+plantName+"%").Scan(&taskCount)
	if taskCount < 1 {
		t.Fatal("expected optional monitor task")
	}

	assertGuardianToolAudit(t, ctx, "apply_grow_setup_pack", proposalID)
}

func TestPhase32WS3_SetupPackRejectsActiveZone(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var zoneID int64
	var zoneName string
	err := testPool.QueryRow(ctx, `
SELECT z.id, z.name FROM gr33ncore.zones z
JOIN gr33nfertigation.crop_cycles cc ON cc.zone_id = z.id AND cc.is_active = TRUE
WHERE z.farm_id = 1
LIMIT 1`).Scan(&zoneID, &zoneName)
	if err != nil {
		t.Skip("no seeded active cycle zone")
	}

	packArgs := housePlantSetupPackArgs(zoneID, zoneName, "Duplicate setup")
	proposalID := insertGuardianProposalWithRisk(t, "apply_grow_setup_pack", packArgs, "Should fail", "high")
	t.Cleanup(func() {
		_, _ = testPool.Exec(context.Background(),
			`DELETE FROM gr33ncore.guardian_action_proposals WHERE proposal_id = $1`, proposalID)
	})

	tok := smokeJWT(t)
	confirmResp := authPost(t, tok, "/v1/chat/confirm", map[string]string{"proposal_id": proposalID})
	defer confirmResp.Body.Close()
	if confirmResp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", confirmResp.StatusCode, readBodyPreview(confirmResp))
	}
}
