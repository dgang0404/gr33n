// Product backlog B1 — program run-now API + idempotency.
package main

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestProductBacklogProgramRunNow(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	tok := smokeJWT(t)
	actID := seedActuatorForFarm1(t, tok, "backlog_run_act")

	resp := authPost(t, tok, "/farms/1/fertigation/programs", map[string]any{
		"name":                uniqueName("backlog_run_prog"),
		"total_volume_liters": 2.0,
		"is_active":           true,
	})
	expectStatus(t, resp, 201)
	progID := int64(decodeMap(t, resp)["id"].(float64))

	resp = authPost(t, tok, fmt.Sprintf("/fertigation/programs/%d/actions", progID), map[string]any{
		"execution_order":    0,
		"action_type":        "control_actuator",
		"target_actuator_id": actID,
		"action_command":     "on",
	})
	expectStatus(t, resp, 201)

	runPath := fmt.Sprintf("/farms/1/fertigation/programs/%d/run-now", progID)
	first := authPost(t, tok, runPath, map[string]any{})
	expectStatus(t, first, 202)
	body := decodeMap(t, first)
	if body["duplicate"] == true {
		t.Fatalf("first run-now should not be duplicate: %v", body)
	}

	second := authPost(t, tok, runPath, map[string]any{})
	expectStatus(t, second, 200)
	body2 := decodeMap(t, second)
	if body2["duplicate"] != true {
		t.Fatalf("second run-now in same minute should be duplicate: %v", body2)
	}

	var runCount int
	if err := testPool.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM gr33ncore.automation_runs
		WHERE program_id = $1 AND executed_at >= date_trunc('minute', NOW() AT TIME ZONE 'UTC')
	`, progID).Scan(&runCount); err != nil {
		t.Fatalf("count automation_runs: %v", err)
	}
	if runCount != 1 {
		t.Fatalf("expected 1 automation_runs row this minute, got %d", runCount)
	}
}
