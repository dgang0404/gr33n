// Phase 39 WS8 — mix plan API: preview, enqueue mix_batch, base-water gate, Pi dequeue.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

// TestPhase39MixPlanPreviewAndEnqueue covers:
//  1. PATCH base-water on reservoir
//  2. POST mix-jobs preview_only → MixPlan with steps
//  3. POST mix-jobs enqueue → command_id; Pi GET next receives mix_batch
//  4. mix-preview and water-status read-only endpoints
func TestPhase39MixPlanPreviewAndEnqueue(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	uid := fmt.Sprintf("ph39mix_%d", time.Now().UnixNano())
	jwt := smokeJWT(t)

	var devID, actID, resID, recipeID, progID int64
	var inputA, inputB int64

	if err := testPool.QueryRow(ctx, `
		INSERT INTO gr33ncore.devices (farm_id, zone_id, name, device_uid, device_type, status, config)
		VALUES (1, 1, $1, $2, 'edge_gateway', 'online', '{}'::jsonb)
		RETURNING id
	`, uid+"_dev", uid+"_uid").Scan(&devID); err != nil {
		t.Fatalf("insert device: %v", err)
	}
	if err := testPool.QueryRow(ctx, `
		INSERT INTO gr33ncore.actuators (device_id, farm_id, zone_id, name, actuator_type, config, meta_data)
		VALUES ($1, 1, 1, $2, 'pump', '{}'::jsonb, '{}'::jsonb)
		RETURNING id
	`, devID, uid+"_pump").Scan(&actID); err != nil {
		t.Fatalf("insert actuator: %v", err)
	}
	if err := testPool.QueryRow(ctx, `
		INSERT INTO gr33nfertigation.reservoirs
		    (farm_id, zone_id, name, capacity_liters, current_volume_liters, status, delivery_actuator_id)
		VALUES (1, 1, $1, 500, 320, 'ready', $2)
		RETURNING id
	`, uid+"_res", actID).Scan(&resID); err != nil {
		t.Fatalf("insert reservoir: %v", err)
	}
	if err := testPool.QueryRow(ctx, `
		SELECT id FROM gr33nnaturalfarming.input_definitions
		WHERE farm_id = 1 AND deleted_at IS NULL
		ORDER BY id LIMIT 1
	`).Scan(&inputA); err != nil {
		t.Fatalf("input def A: %v", err)
	}
	if err := testPool.QueryRow(ctx, `
		SELECT id FROM gr33nnaturalfarming.input_definitions
		WHERE farm_id = 1 AND deleted_at IS NULL AND id <> $1
		ORDER BY id LIMIT 1
	`, inputA).Scan(&inputB); err != nil {
		t.Fatalf("input def B: %v", err)
	}
	if err := testPool.QueryRow(ctx, `
		INSERT INTO gr33nnaturalfarming.application_recipes
		    (farm_id, name, description, target_application_type, dilution_ratio)
		VALUES (1, $1, 'Phase 39 smoke recipe', 'soil_drench', '1:500')
		RETURNING id
	`, uid+"_recipe").Scan(&recipeID); err != nil {
		t.Fatalf("insert recipe: %v", err)
	}
	if _, err := testPool.Exec(ctx, `
		INSERT INTO gr33nnaturalfarming.recipe_input_components
		    (application_recipe_id, input_definition_id, part_value, notes)
		VALUES ($1, $2, 1.0, 'A'), ($1, $3, 1.0, 'B')
	`, recipeID, inputA, inputB); err != nil {
		t.Fatalf("insert components: %v", err)
	}
	if err := testPool.QueryRow(ctx, `
		INSERT INTO gr33nfertigation.programs
		    (farm_id, name, application_recipe_id, reservoir_id, target_zone_id,
		     total_volume_liters, run_duration_seconds, ec_trigger_low, ph_trigger_low, ph_trigger_high, is_active)
		VALUES (1, $1, $2, $3, 1, 95, 30, 1.0, 5.8, 6.8, TRUE)
		RETURNING id
	`, uid+"_prog", recipeID, resID).Scan(&progID); err != nil {
		t.Fatalf("insert program: %v", err)
	}

	t.Cleanup(func() {
		bg := context.Background()
		_, _ = testPool.Exec(bg, `DELETE FROM gr33ncore.device_commands WHERE device_id = $1`, devID)
		_, _ = testPool.Exec(bg, `DELETE FROM gr33nfertigation.programs WHERE id = $1`, progID)
		_, _ = testPool.Exec(bg, `DELETE FROM gr33nnaturalfarming.recipe_input_components WHERE application_recipe_id = $1`, recipeID)
		_, _ = testPool.Exec(bg, `DELETE FROM gr33nnaturalfarming.application_recipes WHERE id = $1`, recipeID)
		_, _ = testPool.Exec(bg, `DELETE FROM gr33nfertigation.reservoirs WHERE id = $1`, resID)
		_, _ = testPool.Exec(bg, `DELETE FROM gr33ncore.actuators WHERE id = $1`, actID)
		_, _ = testPool.Exec(bg, `DELETE FROM gr33ncore.devices WHERE id = $1`, devID)
	})

	// Without base EC, mix-jobs must refuse.
	badMix := authPost(t, jwt, "/farms/1/fertigation/mix-jobs", map[string]any{
		"program_id": progID,
	})
	defer badMix.Body.Close()
	expectStatus(t, badMix, http.StatusBadRequest)

	// Set base water EC (WS6).
	baseResp := authPatch(t, jwt, fmt.Sprintf("/fertigation/reservoirs/%d/base-water", resID), map[string]any{
		"ec_mscm": 0.2,
		"ph":      7.0,
	})
	defer baseResp.Body.Close()
	expectStatus(t, baseResp, http.StatusOK)

	// Preview only.
	prevResp := authPost(t, jwt, "/farms/1/fertigation/mix-jobs", map[string]any{
		"program_id":   progID,
		"preview_only": true,
	})
	defer prevResp.Body.Close()
	expectStatus(t, prevResp, http.StatusOK)
	var prevOut map[string]any
	if err := json.NewDecoder(prevResp.Body).Decode(&prevOut); err != nil {
		t.Fatalf("decode preview: %v", err)
	}
	if prevOut["preview_only"] != true {
		t.Fatalf("expected preview_only=true; got %v", prevOut["preview_only"])
	}
	plan, ok := prevOut["mix_plan"].(map[string]any)
	if !ok {
		t.Fatalf("expected mix_plan object; got %#v", prevOut)
	}
	steps, _ := plan["steps"].([]any)
	if len(steps) < 2 {
		t.Fatalf("expected >=2 mix steps; got %d", len(steps))
	}

	// mix-preview GET
	mixPrev := authGet(t, jwt, fmt.Sprintf("/fertigation/programs/%d/mix-preview", progID))
	defer mixPrev.Body.Close()
	expectStatus(t, mixPrev, http.StatusOK)

	// water-status GET
	waterStat := authGet(t, jwt, fmt.Sprintf("/fertigation/programs/%d/water-status", progID))
	defer waterStat.Body.Close()
	expectStatus(t, waterStat, http.StatusOK)
	var ws map[string]any
	if err := json.NewDecoder(waterStat.Body).Decode(&ws); err != nil {
		t.Fatalf("decode water-status: %v", err)
	}
	if ws["mix_required"] != true {
		t.Fatalf("expected mix_required=true; got %v", ws["mix_required"])
	}

	// Enqueue mix_batch.
	enqResp := authPost(t, jwt, "/farms/1/fertigation/mix-jobs", map[string]any{
		"program_id": progID,
	})
	defer enqResp.Body.Close()
	expectStatus(t, enqResp, http.StatusAccepted)
	var enqOut map[string]any
	if err := json.NewDecoder(enqResp.Body).Decode(&enqOut); err != nil {
		t.Fatalf("decode enqueue: %v", err)
	}
	cmdID := int64(enqOut["command_id"].(float64))
	if cmdID == 0 {
		t.Fatalf("expected command_id; got %#v", enqOut)
	}

	nextResp := piGet(t, fmt.Sprintf("/devices/%d/commands/next", devID))
	defer nextResp.Body.Close()
	expectStatus(t, nextResp, http.StatusOK)
	var nextOut map[string]any
	if err := json.NewDecoder(nextResp.Body).Decode(&nextOut); err != nil {
		t.Fatalf("decode next: %v", err)
	}
	if nextOut["command_type"] != "mix_batch" {
		t.Fatalf("expected command_type=mix_batch; got %v", nextOut["command_type"])
	}
	payload := extractPayload(t, nextOut)
	if payload["command_type"] != "mix_batch" {
		t.Fatalf("payload command_type: %v", payload["command_type"])
	}
	mixPlan, ok := payload["mix_plan"].(map[string]any)
	if !ok || len(mixPlan) == 0 {
		t.Fatalf("expected mix_plan in payload; got %#v", payload)
	}

	ackResp := piPostJSON(t, fmt.Sprintf("/devices/%d/commands/%d/ack", devID, cmdID),
		map[string]any{"status": "completed", "result": map[string]any{"steps_ran": 2}},
	)
	defer ackResp.Body.Close()
	expectStatus(t, ackResp, http.StatusOK)
}
