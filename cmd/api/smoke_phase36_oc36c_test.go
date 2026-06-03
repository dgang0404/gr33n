// Phase 36 OC-36C — greenhouse template interlocks, rule fire + cooldown, manual shade command.
package main

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func seedPhase36GHZone(t *testing.T, tok string) int64 {
	t.Helper()
	resp := authPost(t, tok, "/farms/1/zones", map[string]any{
		"name":        uniqueName("gh_oc36c"),
		"description": "Phase 36 greenhouse smoke zone",
		"zone_type":   "greenhouse",
	})
	expectStatus(t, resp, http.StatusCreated)
	return int64(decodeMap(t, resp)["id"].(float64))
}

func seedPhase36ShadeActuator(t *testing.T, ctx context.Context, zoneID int64) int64 {
	t.Helper()
	var id int64
	err := testPool.QueryRow(ctx, `
INSERT INTO gr33ncore.actuators (farm_id, zone_id, name, actuator_type)
VALUES (1, $1, $2, 'shade_screen') RETURNING id`, zoneID, uniqueName("gh_shade")).Scan(&id)
	if err != nil {
		t.Fatalf("shade actuator: %v", err)
	}
	return id
}

func seedPhase36LuxSensor(t *testing.T, tok string, ctx context.Context, zoneID int64) int64 {
	t.Helper()
	var unitID int64
	if err := testPool.QueryRow(ctx, `SELECT id FROM gr33ncore.units LIMIT 1`).Scan(&unitID); err != nil {
		t.Fatalf("unit: %v", err)
	}
	resp := authPost(t, tok, "/farms/1/sensors", map[string]any{
		"name":        uniqueName("gh_lux"),
		"sensor_type": "lux",
		"unit_id":     unitID,
		"zone_id":     zoneID,
	})
	expectStatus(t, resp, http.StatusCreated)
	return int64(decodeMap(t, resp)["id"].(float64))
}

func TestPhase36OC36C_TemplateRequiresLuxOrOverride(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	tok := smokeJWT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	zoneID := seedPhase36GHZone(t, tok)
	shadeID := seedPhase36ShadeActuator(t, ctx, zoneID)

	resp := authPost(t, tok, "/farms/1/automation/rule-templates/greenhouse", map[string]any{
		"zone_id":           zoneID,
		"shade_actuator_id": shadeID,
	})
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusBadRequest)

	resp = authPost(t, tok, "/farms/1/automation/rule-templates/greenhouse", map[string]any{
		"zone_id":                  zoneID,
		"shade_actuator_id":        shadeID,
		"allow_missing_lux_sensor": true,
	})
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusCreated)
	body := decodeMap(t, resp)
	skipped, ok := body["skipped_rule_families"].([]any)
	if !ok || len(skipped) == 0 {
		t.Fatalf("expected skipped_rule_families, got %#v", body["skipped_rule_families"])
	}
}

func TestPhase36OC36C_HighLuxRuleFireAndCooldown(t *testing.T) {
	if testPool == nil || testWorker == nil {
		t.Skip("testPool/testWorker unavailable")
	}
	tok := smokeJWT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	zoneID := seedPhase36GHZone(t, tok)
	luxID := seedPhase36LuxSensor(t, tok, ctx, zoneID)
	shadeID := seedPhase36ShadeActuator(t, ctx, zoneID)

	resp := authPost(t, tok, "/farms/1/automation/rule-templates/greenhouse", map[string]any{
		"zone_id":           zoneID,
		"shade_actuator_id": shadeID,
		"lux_sensor_id":     luxID,
	})
	expectStatus(t, resp, http.StatusCreated)
	resp.Body.Close()

	var ruleID int64
	ruleName := fmt.Sprintf("GH — High lux: deploy shade (zone %d)", zoneID)
	if err := testPool.QueryRow(ctx, `
SELECT id FROM gr33ncore.automation_rules
WHERE farm_id = 1 AND name = $1`, ruleName).Scan(&ruleID); err != nil {
		t.Fatalf("find GH lux rule: %v", err)
	}

	if _, err := testPool.Exec(ctx, `
UPDATE gr33ncore.automation_rules SET cooldown_period_seconds = 120 WHERE id = $1`, ruleID); err != nil {
		t.Fatalf("shorten cooldown: %v", err)
	}

	resp = authPatch(t, tok, fmt.Sprintf("/automation/rules/%d/active", ruleID), map[string]any{"is_active": true})
	expectStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	if _, err := testPool.Exec(ctx, `
INSERT INTO gr33ncore.sensor_readings (reading_time, sensor_id, value_raw, is_valid)
VALUES (NOW(), $1, 90000, TRUE)`, luxID); err != nil {
		t.Fatalf("seed lux reading: %v", err)
	}

	testWorker.TickRules(ctx)

	var status string
	if err := testPool.QueryRow(ctx, `
SELECT status FROM gr33ncore.automation_runs WHERE rule_id = $1 ORDER BY id DESC LIMIT 1`, ruleID).Scan(&status); err != nil {
		t.Fatalf("rule run: %v", err)
	}
	if status != "success" {
		t.Fatalf("expected success run, got %s", status)
	}

	var cmd string
	if err := testPool.QueryRow(ctx, `
SELECT command_sent FROM gr33ncore.actuator_events
WHERE triggered_by_rule_id = $1 ORDER BY event_time DESC LIMIT 1`, ruleID).Scan(&cmd); err != nil {
		t.Fatalf("actuator event: %v", err)
	}
	if cmd != "deploy" {
		t.Fatalf("expected deploy command, got %q", cmd)
	}

	testWorker.TickRules(ctx)
	var msg string
	if err := testPool.QueryRow(ctx, `
SELECT COALESCE(message, '') FROM gr33ncore.automation_runs
WHERE rule_id = $1 ORDER BY id DESC LIMIT 1`, ruleID).Scan(&msg); err != nil {
		t.Fatalf("second run: %v", err)
	}
	if msg != "cooldown" {
		t.Fatalf("expected cooldown on second tick, got message=%q", msg)
	}
}

func TestPhase36OC36C_ActivateHighLuxWithoutSensorRejected(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	tok := smokeJWT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	zoneID := seedPhase36GHZone(t, tok)
	luxID := seedPhase36LuxSensor(t, tok, ctx, zoneID)
	shadeID := seedPhase36ShadeActuator(t, ctx, zoneID)

	resp := authPost(t, tok, "/farms/1/automation/rule-templates/greenhouse", map[string]any{
		"zone_id":           zoneID,
		"shade_actuator_id": shadeID,
		"lux_sensor_id":     luxID,
	})
	expectStatus(t, resp, http.StatusCreated)
	resp.Body.Close()

	ruleName := fmt.Sprintf("GH — High lux: deploy shade (zone %d)", zoneID)
	var ruleID int64
	if err := testPool.QueryRow(ctx, `
SELECT id FROM gr33ncore.automation_rules WHERE farm_id = 1 AND name = $1`, ruleName).Scan(&ruleID); err != nil {
		t.Fatalf("find GH lux rule: %v", err)
	}

	// Simulate a broken rule row (no sensor_id) — activation must be rejected.
	if _, err := testPool.Exec(ctx, `
UPDATE gr33ncore.automation_rules
SET is_active = FALSE,
    trigger_configuration = '{"op":"gt","value":80000}'::jsonb
WHERE id = $1`, ruleID); err != nil {
		t.Fatalf("prep rule: %v", err)
	}

	resp = authPatch(t, tok, fmt.Sprintf("/automation/rules/%d/active", ruleID), map[string]any{"is_active": true})
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusBadRequest)
}
