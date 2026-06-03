// Phase 34 — revise/supersede + confirm-safety smoke.
// Exercises the DB-level supersede chain and the Confirm gate without the LLM:
// rev1 → superseded, rev2 (volume delta + operator fact) pending; confirming the
// superseded draft returns 410 + live_proposal_id, confirming the latest applies
// the delta and audits the revision lineage + operator-supplied facts.
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

func insertRevisionProposal(
	t *testing.T,
	toolID string,
	args map[string]any,
	summary, riskTier, sessionID string,
	supersedes string,
	revision int,
	meta map[string]any,
	status string,
) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	uid := uuid.MustParse(smokeDevUserUUID)
	argsRaw, _ := json.Marshal(args)
	metaRaw, _ := json.Marshal(meta)
	if len(metaRaw) == 0 {
		metaRaw = []byte("{}")
	}
	var supersedesArg any
	if supersedes != "" {
		supersedesArg = supersedes
	}

	var proposalID string
	err := testPool.QueryRow(ctx, `
INSERT INTO gr33ncore.guardian_action_proposals
    (user_id, farm_id, session_id, tool_id, args, summary, risk_tier, expires_at,
     meta, supersedes_proposal_id, revision, status)
VALUES ($1, 1, $2, $3, $4::jsonb, $5, $6, NOW() + INTERVAL '10 minutes',
        $7::jsonb, $8, $9, $10::gr33ncore.guardian_proposal_status_enum)
RETURNING proposal_id::text`,
		uid, sessionID, toolID, argsRaw, summary, riskTier,
		metaRaw, supersedesArg, revision, status).Scan(&proposalID)
	if err != nil {
		t.Fatalf("insert revision proposal: %v", err)
	}
	return proposalID
}

func TestPhase34_ReviseSupersedeConfirm(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	zoneName := fmt.Sprintf("Tent A P34 %d", time.Now().UnixNano())
	plantName := fmt.Sprintf("Philodendron P34 %d", time.Now().UnixNano())
	sessionID := uuid.NewString()

	var zoneID int64
	if err := testPool.QueryRow(ctx, `
INSERT INTO gr33ncore.zones (farm_id, name, description, zone_type)
VALUES (1, $1, 'Phase 34 smoke zone', 'indoor')
RETURNING id`, zoneName).Scan(&zoneID); err != nil {
		t.Fatalf("insert zone: %v", err)
	}
	t.Cleanup(func() {
		c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, _ = testPool.Exec(c, `DELETE FROM gr33ncore.tasks WHERE farm_id = 1 AND title LIKE $1`, "Monitor new "+plantName+"%")
		_, _ = testPool.Exec(c, `UPDATE gr33nfertigation.programs SET deleted_at = NOW() WHERE farm_id = 1 AND name = $1`, plantName+" light feed")
		_, _ = testPool.Exec(c, `DELETE FROM gr33nfertigation.crop_cycles WHERE zone_id = $1`, zoneID)
		_, _ = testPool.Exec(c, `UPDATE gr33ncrops.plants SET deleted_at = NOW() WHERE farm_id = 1 AND display_name = $1`, plantName)
		_, _ = testPool.Exec(c, `DELETE FROM gr33ncore.guardian_action_proposals WHERE session_id = $1`, sessionID)
		_, _ = testPool.Exec(c, `DELETE FROM gr33ncore.zones WHERE id = $1`, zoneID)
	})

	// Revision 1 — the original frozen draft (0.5 L), now superseded.
	rev1Args := housePlantSetupPackArgs(zoneID, zoneName, plantName)
	rev1 := insertRevisionProposal(t, "apply_grow_setup_pack", rev1Args,
		"Setup pack rev1", "high", sessionID, "", 1, nil, "superseded")

	// Revision 2 — supersedes rev1, volume corrected to 0.3 L, operator-stated RH.
	rev2Args := housePlantSetupPackArgs(zoneID, zoneName, plantName)
	rev2Args["program"].(map[string]any)["total_volume_liters"] = 0.3
	rev2Meta := map[string]any{
		"operator_provided": []map[string]any{
			{"field": "rh_pct", "value": 60, "basis": "operator_stated",
				"label": "RH 60% (operator-stated, not measured)"},
		},
	}
	rev2 := insertRevisionProposal(t, "apply_grow_setup_pack", rev2Args,
		"Setup pack rev2", "high", sessionID, rev1, 2, rev2Meta, "pending")

	tok := smokeJWT(t)

	// Confirming the superseded draft must 410 and point at the live revision.
	supersededResp := authPost(t, tok, "/v1/chat/confirm", map[string]string{"proposal_id": rev1})
	defer supersededResp.Body.Close()
	expectStatus(t, supersededResp, http.StatusGone)
	var goneBody struct {
		Error          string `json:"error"`
		LiveProposalID string `json:"live_proposal_id"`
		LiveRevision   int    `json:"live_revision"`
	}
	decodeJSON(t, supersededResp.Body, &goneBody)
	if goneBody.LiveProposalID != rev2 {
		t.Fatalf("410 live_proposal_id = %q want rev2 %q", goneBody.LiveProposalID, rev2)
	}
	if goneBody.LiveRevision != 2 {
		t.Fatalf("410 live_revision = %d want 2", goneBody.LiveRevision)
	}

	// Confirming the latest live revision applies the corrected args.
	confirmResp := authPost(t, tok, "/v1/chat/confirm", map[string]string{"proposal_id": rev2})
	defer confirmResp.Body.Close()
	expectStatus(t, confirmResp, http.StatusOK)
	var confirmBody struct {
		Result map[string]any `json:"result"`
	}
	decodeJSON(t, confirmResp.Body, &confirmBody)
	programBlock, _ := confirmBody.Result["program"].(map[string]any)
	if programBlock == nil {
		t.Fatalf("missing program result: %+v", confirmBody.Result)
	}
	programID, _ := programBlock["program_id"].(float64)
	if programID == 0 {
		t.Fatalf("missing program id: %+v", programBlock)
	}

	// The 0.3 L delta from rev2 must be the value persisted, not the 0.5 L original.
	var vol float64
	if err := testPool.QueryRow(ctx,
		`SELECT total_volume_liters FROM gr33nfertigation.programs WHERE id = $1`,
		int64(programID)).Scan(&vol); err != nil {
		t.Fatalf("program row: %v", err)
	}
	if vol != 0.3 {
		t.Fatalf("persisted volume = %v want 0.3 (delta not applied)", vol)
	}

	// Audit must carry revision lineage + operator-supplied facts.
	var details []byte
	if err := testPool.QueryRow(ctx, `
SELECT details FROM gr33ncore.user_activity_log
WHERE action_type = 'guardian_tool_executed'
  AND details->>'proposal_id' = $1
ORDER BY created_at DESC LIMIT 1`, rev2).Scan(&details); err != nil {
		t.Fatalf("audit query: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(details, &parsed); err != nil {
		t.Fatalf("audit details: %v", err)
	}
	if rev, _ := parsed["revision"].(float64); rev != 2 {
		t.Fatalf("audit revision = %v want 2", parsed["revision"])
	}
	if root, _ := parsed["root_proposal_id"].(string); root != rev1 {
		t.Fatalf("audit root_proposal_id = %v want rev1 %q", parsed["root_proposal_id"], rev1)
	}
	if _, ok := parsed["operator_provided"]; !ok {
		t.Fatalf("audit missing operator_provided: %v", parsed)
	}
}
