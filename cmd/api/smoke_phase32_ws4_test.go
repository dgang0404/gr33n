// Phase 32 WS4 — setup pack intent → proposal generation smoke.
package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmguardian"
)

func TestPhase32WS4_BuildRuleAssistedSetupPackProposal(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	q := db.New(testPool)
	zoneName := fmt.Sprintf("Living Room WS4 %d", time.Now().UnixNano())
	var zoneID int64
	err := testPool.QueryRow(ctx, `
INSERT INTO gr33ncore.zones (farm_id, name, description, zone_type)
VALUES (1, $1, 'Phase 32 WS4 intent smoke', 'indoor')
RETURNING id`, zoneName).Scan(&zoneID)
	if err != nil {
		t.Fatalf("insert zone: %v", err)
	}

	snap, err := farmguardian.BuildSnapshot(ctx, q, 1)
	if err != nil {
		t.Fatalf("BuildSnapshot: %v", err)
	}

	question := "add my philodendron to " + zoneName + " with a light fertigation program"
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
	if prop.RiskTier != "high" {
		t.Fatalf("risk %q want high", prop.RiskTier)
	}
	plant, _ := prop.Args["plant"].(map[string]any)
	if plant["display_name"] != "Philodendron" {
		t.Fatalf("plant %#v", plant["display_name"])
	}

	t.Cleanup(func() {
		c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, _ = testPool.Exec(c, `DELETE FROM gr33ncore.guardian_action_proposals WHERE proposal_id = $1`, prop.ProposalID)
		_, _ = testPool.Exec(c, `DELETE FROM gr33ncore.zones WHERE id = $1`, zoneID)
	})
}

func TestPhase32WS4_NonsenseZoneNoProposal(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	q := db.New(testPool)
	snap, err := farmguardian.BuildSnapshot(ctx, q, 1)
	if err != nil {
		t.Fatalf("BuildSnapshot: %v", err)
	}
	uid := uuid.MustParse(smokeDevUserUUID)
	props, err := farmguardian.BuildRuleAssistedProposals(ctx, q, uid, 1, uuid.Nil,
		"add my philodendron to Narnia with a light fertigation program", snap)
	if err != nil {
		t.Fatalf("BuildRuleAssistedProposals: %v", err)
	}
	for _, p := range props {
		if p.Tool == "apply_grow_setup_pack" {
			t.Fatalf("unexpected setup pack for nonsense zone: %+v", p)
		}
	}
}
