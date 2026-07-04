// Phase 32 WS7 — end-to-end setup pack: rule-assisted propose → confirm → DB rows.
package main

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmguardian"
)

func TestPhase32WS7_SetupPackIntentToConfirm(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	q := db.New(testPool)
	suffix := time.Now().UnixNano()
	zoneName := fmt.Sprintf("Living Room WS7 %d", suffix)

	var zoneID int64
	err := testPool.QueryRow(ctx, `
INSERT INTO gr33ncore.zones (farm_id, name, description, zone_type)
VALUES (1, $1, 'Phase 32 WS7 E2E zone', 'indoor')
RETURNING id`, zoneName).Scan(&zoneID)
	if err != nil {
		t.Fatalf("insert zone: %v", err)
	}
	// "kale" — a crop_key the Phase 124 demo seed doesn't touch — so this
	// test's cleanup never soft-deletes a permanently-seeded plant.
	t.Cleanup(func() {
		c, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		_, _ = testPool.Exec(c, `DELETE FROM gr33ncore.tasks WHERE farm_id = 1 AND title LIKE $1`, "Monitor new Kale%")
		_, _ = testPool.Exec(c, `UPDATE gr33nfertigation.programs SET deleted_at = NOW() WHERE farm_id = 1 AND name LIKE $1`, "Kale% light feed")
		_, _ = testPool.Exec(c, `DELETE FROM gr33nfertigation.crop_cycles WHERE zone_id = $1`, zoneID)
		_, _ = testPool.Exec(c, `UPDATE gr33ncrops.plants SET deleted_at = NOW() WHERE farm_id = 1 AND crop_key = 'kale'`)
		_, _ = testPool.Exec(c, `DELETE FROM gr33ncore.zones WHERE id = $1`, zoneID)
	})

	snap, err := farmguardian.BuildSnapshot(ctx, q, 1)
	if err != nil {
		t.Fatalf("BuildSnapshot: %v", err)
	}

	question := fmt.Sprintf("add kale to %s with a light fertigation program", zoneName)
	uid := uuid.MustParse(smokeDevUserUUID)
	props, err := farmguardian.BuildRuleAssistedProposals(ctx, q, uid, 1, uuid.Nil, question, snap)
	if err != nil {
		t.Fatalf("BuildRuleAssistedProposals: %v", err)
	}
	if len(props) != 1 {
		t.Fatalf("expected 1 proposal, got %+v", props)
	}
	prop := props[0]
	if prop.Tool != "apply_grow_setup_pack" {
		t.Fatalf("tool %q want apply_grow_setup_pack", prop.Tool)
	}
	t.Cleanup(func() {
		_, _ = testPool.Exec(context.Background(),
			`DELETE FROM gr33ncore.guardian_action_proposals WHERE proposal_id = $1`, prop.ProposalID)
	})

	tok := smokeJWT(t)
	confirmResp := authPost(t, tok, "/v1/chat/confirm", map[string]string{"proposal_id": prop.ProposalID})
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

	var cropKey string
	err = testPool.QueryRow(ctx, `
SELECT crop_key FROM gr33ncrops.plants
WHERE id = $1 AND farm_id = 1 AND deleted_at IS NULL`, int64(plantID)).Scan(&cropKey)
	if err != nil {
		t.Fatalf("plant row: %v", err)
	}
	if cropKey != "kale" {
		t.Fatalf("expected crop_key kale, got %q", cropKey)
	}

	var cycleActive bool
	err = testPool.QueryRow(ctx, `
SELECT is_active FROM gr33nfertigation.crop_cycles
WHERE id = $1 AND farm_id = 1 AND zone_id = $2`, int64(cycleID), zoneID).Scan(&cycleActive)
	if err != nil {
		t.Fatalf("cycle row: %v", err)
	}
	if !cycleActive {
		t.Fatal("expected active crop cycle")
	}

	var linkedProgramID *int64
	err = testPool.QueryRow(ctx, `
SELECT primary_program_id FROM gr33nfertigation.crop_cycles WHERE id = $1`, int64(cycleID)).Scan(&linkedProgramID)
	if err != nil {
		t.Fatalf("cycle primary_program_id: %v", err)
	}
	if linkedProgramID == nil || *linkedProgramID != int64(programID) {
		t.Fatalf("primary_program_id = %v want %d", linkedProgramID, int64(programID))
	}

	var programName string
	err = testPool.QueryRow(ctx, `
SELECT name FROM gr33nfertigation.programs
WHERE id = $1 AND farm_id = 1 AND deleted_at IS NULL`, int64(programID)).Scan(&programName)
	if err != nil {
		t.Fatalf("program row: %v", err)
	}
	if programName == "" {
		t.Fatal("expected program name")
	}

	assertGuardianToolAudit(t, ctx, "apply_grow_setup_pack", prop.ProposalID)
}

func TestPhase32WS7_SetupPackIntentSkipsBusyZone(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
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

	q := db.New(testPool)
	snap, err := farmguardian.BuildSnapshot(ctx, q, 1)
	if err != nil {
		t.Fatalf("BuildSnapshot: %v", err)
	}

	question := fmt.Sprintf("add my basil to %s with a light fertigation program", zoneName)
	uid := uuid.MustParse(smokeDevUserUUID)
	props, err := farmguardian.BuildRuleAssistedProposals(ctx, q, uid, 1, uuid.Nil, question, snap)
	if err != nil {
		t.Fatalf("BuildRuleAssistedProposals: %v", err)
	}
	for _, p := range props {
		if p.Tool == "apply_grow_setup_pack" {
			t.Fatalf("unexpected setup pack for busy zone %q: %+v", zoneName, p)
		}
	}
}
