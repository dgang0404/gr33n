// Phase 20.95 WS5 — split out of cmd/api/smoke_test.go with zero behaviour
// change. Shared globals (testPool / testServer / testWorker / testNotifier)
// and helpers live in smoke_helpers_test.go; TestMain stays in smoke_test.go.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestSchedulePreconditionFailsRun(t *testing.T) {
	if testPool == nil || testWorker == nil {
		t.Skip("testPool/testWorker unavailable")
	}
	tok := smokeJWT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var unitID int64
	if err := testPool.QueryRow(ctx, `SELECT id FROM gr33ncore.units LIMIT 1`).Scan(&unitID); err != nil {
		t.Fatalf("find a unit id: %v", err)
	}

	// Sensor on farm 1 — we'll seed a "tank empty" reading below its threshold.
	sensorName := uniqueName("precondition_sensor")
	resp := authPost(t, tok, "/farms/1/sensors", map[string]any{
		"name":        sensorName,
		"sensor_type": "level",
		"unit_id":     unitID,
	})
	expectStatus(t, resp, http.StatusCreated)
	sensorRow := decodeMap(t, resp)
	sid := int64(sensorRow["id"].(float64))

	// Seed a failing reading: level = 2, rule will require >= 50.
	if _, err := testPool.Exec(ctx, `
		INSERT INTO gr33ncore.sensor_readings (reading_time, sensor_id, value_raw, is_valid)
		VALUES (NOW(), $1, 2, TRUE)`, sid); err != nil {
		t.Fatalf("seed failing reading: %v", err)
	}

	// --- Validation: precondition with a sensor from another farm is rejected. ---
	resp = authPost(t, tok, "/farms/1/schedules", map[string]any{
		"name":            uniqueName("precond_invalid"),
		"schedule_type":   "cron",
		"cron_expression": "* * * * *",
		"timezone":        "UTC",
		"is_active":       true,
		"preconditions": []map[string]any{
			{"sensor_id": 999999, "op": "gte", "value": 50.0},
		},
	})
	expectStatus(t, resp, http.StatusBadRequest)
	resp.Body.Close()

	// --- Validation: unknown op is rejected. ---
	resp = authPost(t, tok, "/farms/1/schedules", map[string]any{
		"name":            uniqueName("precond_badop"),
		"schedule_type":   "cron",
		"cron_expression": "* * * * *",
		"timezone":        "UTC",
		"is_active":       true,
		"preconditions": []map[string]any{
			{"sensor_id": sid, "op": "totally-invalid", "value": 50.0},
		},
	})
	expectStatus(t, resp, http.StatusBadRequest)
	resp.Body.Close()

	// Every-minute schedule with a precondition that the current reading (2) will FAIL.
	schedName := uniqueName("interlock_schedule")
	resp = authPost(t, tok, "/farms/1/schedules", map[string]any{
		"name":            schedName,
		"schedule_type":   "cron",
		"cron_expression": "* * * * *",
		"timezone":        "UTC",
		"is_active":       true,
		"preconditions": []map[string]any{
			{"sensor_id": sid, "op": "gte", "value": 50.0},
		},
	})
	expectStatus(t, resp, http.StatusCreated)
	schedRow := decodeMap(t, resp)
	schedID := int64(schedRow["id"].(float64))

	// Remember how many actuator events exist for this farm before the tick — the
	// worker must NOT write any, since no executable actions should run.
	var evBefore int
	if err := testPool.QueryRow(ctx, `
		SELECT COUNT(*) FROM gr33ncore.actuator_events
		WHERE triggered_by_schedule_id = $1`, schedID).Scan(&evBefore); err != nil {
		t.Fatalf("count actuator events before: %v", err)
	}

	// Run a tick — the precondition should fail and the run should be recorded as skipped.
	testWorker.Tick(ctx)

	// --- Assert a skipped run with message='precondition_failed' exists. ---
	var msg string
	var status string
	var detailsJSON []byte
	if err := testPool.QueryRow(ctx, `
		SELECT status, message, details::text FROM gr33ncore.automation_runs
		WHERE schedule_id = $1 AND status = 'skipped' AND message = 'precondition_failed'
		ORDER BY executed_at DESC LIMIT 1`, schedID).Scan(&status, &msg, &detailsJSON); err != nil {
		t.Fatalf("expected a skipped run with precondition_failed: %v", err)
	}
	if status != "skipped" || msg != "precondition_failed" {
		t.Fatalf("expected status=skipped message=precondition_failed, got %s/%s", status, msg)
	}
	// PostgreSQL's JSONB rendering collapses whitespace inconsistently
	// between releases, so parse before asserting.
	var details struct {
		Phase  string `json:"phase"`
		Failed []struct {
			SensorID int64   `json:"sensor_id"`
			Op       string  `json:"op"`
			Expected float64 `json:"expected"`
			Actual   float64 `json:"actual"`
			Reason   string  `json:"reason"`
		} `json:"failed"`
	}
	if err := json.Unmarshal(detailsJSON, &details); err != nil {
		t.Fatalf("parse details json: %v (raw=%s)", err, string(detailsJSON))
	}
	if details.Phase != "preconditions" {
		t.Fatalf("expected details.phase=preconditions, got %q", details.Phase)
	}
	if len(details.Failed) != 1 {
		t.Fatalf("expected 1 failed precondition, got %d", len(details.Failed))
	}
	f := details.Failed[0]
	if f.SensorID != sid || f.Op != "gte" || f.Expected != 50 || f.Actual != 2 || f.Reason != "predicate_failed" {
		t.Fatalf("unexpected failed entry: %+v", f)
	}

	// No actuator events should have been written for this schedule.
	var evAfter int
	if err := testPool.QueryRow(ctx, `
		SELECT COUNT(*) FROM gr33ncore.actuator_events
		WHERE triggered_by_schedule_id = $1`, schedID).Scan(&evAfter); err != nil {
		t.Fatalf("count actuator events after: %v", err)
	}
	if evAfter != evBefore {
		t.Fatalf("expected no actuator events when precondition fails, got %d new", evAfter-evBefore)
	}
	// Last-triggered should remain NULL — the next tick should get another chance.
	var lastTriggered *time.Time
	if err := testPool.QueryRow(ctx,
		`SELECT last_triggered_time FROM gr33ncore.schedules WHERE id = $1`, schedID,
	).Scan(&lastTriggered); err != nil {
		t.Fatalf("read last_triggered_time: %v", err)
	}
	if lastTriggered != nil {
		t.Fatalf("expected last_triggered_time to remain NULL when skipped by precondition, got %v", *lastTriggered)
	}

	// --- Flip the predicate: the reading (2) satisfies op=lte value=50. ---
	resp = authPut(t, tok, fmt.Sprintf("/schedules/%d", schedID), map[string]any{
		"name":            schedName,
		"schedule_type":   "cron",
		"cron_expression": "* * * * *",
		"timezone":        "UTC",
		"is_active":       true,
		"preconditions": []map[string]any{
			{"sensor_id": sid, "op": "lte", "value": 50.0},
		},
	})
	expectStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	testWorker.Tick(ctx)

	// After flipping, preconditions pass and the worker should proceed to
	// executeSchedule. No executable actions are attached, so the run we
	// care about is the post-precondition run — we assert it is NOT a
	// precondition_failed row. Several rows may share executed_at within
	// the same minute, so order by id (monotonic) rather than timestamp.
	var latestStatus, latestMsg string
	if err := testPool.QueryRow(ctx, `
		SELECT status, COALESCE(message, '') FROM gr33ncore.automation_runs
		WHERE schedule_id = $1
		ORDER BY id DESC LIMIT 1`, schedID).Scan(&latestStatus, &latestMsg); err != nil {
		t.Fatalf("read latest run after flip: %v", err)
	}
	if latestMsg == "precondition_failed" {
		t.Fatalf("expected the latest run to pass preconditions after flipping the rule, got message=%s", latestMsg)
	}

	// Double-check that the flip caused a new run to be recorded — counts
	// should reflect at least one non-precondition_failed skipped/success row.
	var nonPrecondCount int
	if err := testPool.QueryRow(ctx, `
		SELECT COUNT(*) FROM gr33ncore.automation_runs
		WHERE schedule_id = $1 AND COALESCE(message, '') <> 'precondition_failed'`, schedID,
	).Scan(&nonPrecondCount); err != nil {
		t.Fatalf("count non-precondition runs: %v", err)
	}
	if nonPrecondCount == 0 {
		t.Fatalf("expected at least one non-precondition_failed run after flip")
	}
}

func TestListSchedules(t *testing.T) {
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/farms/1/schedules")
	expectStatus(t, resp, 200)
	_ = decodeSlice(t, resp)
}

func TestListAutomationRuns(t *testing.T) {
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/farms/1/automation/runs")
	expectStatus(t, resp, 200)
	_ = decodeSlice(t, resp)
}

func TestPhase2095ExecutableActionsProgramID(t *testing.T) {
	tok := smokeJWT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Find (or make) a fertigation program we can attach an action to.
	var programID int64
	if err := testPool.QueryRow(ctx,
		`SELECT id FROM gr33nfertigation.programs WHERE farm_id = 1 ORDER BY id LIMIT 1`,
	).Scan(&programID); err != nil {
		// If none exist in seed data, create one directly.
		row := testPool.QueryRow(ctx, `
			INSERT INTO gr33nfertigation.programs (farm_id, name, is_active)
			VALUES (1, 'ws3_smoke_program', FALSE)
			RETURNING id`)
		if err := row.Scan(&programID); err != nil {
			t.Fatalf("seed program: %v", err)
		}
	}

	// ── 1. program_id round-trip (direct INSERT; UI comes in Phase 20.7) ──
	var actionID int64
	if err := testPool.QueryRow(ctx, `
		INSERT INTO gr33ncore.executable_actions
		    (program_id, action_type, action_parameters)
		VALUES ($1, 'create_task', '{"title":"ws3-smoke"}'::jsonb)
		RETURNING id`, programID).Scan(&actionID); err != nil {
		t.Fatalf("insert program-bound action: %v", err)
	}

	var readBackProgramID *int64
	if err := testPool.QueryRow(ctx,
		`SELECT program_id FROM gr33ncore.executable_actions WHERE id = $1`, actionID,
	).Scan(&readBackProgramID); err != nil {
		t.Fatalf("read action: %v", err)
	}
	if readBackProgramID == nil || *readBackProgramID != programID {
		t.Fatalf("expected program_id=%d, got %v", programID, readBackProgramID)
	}

	// ── 2. DB CHECK rejects two-source rows (schedule_id + program_id) ──
	var scheduleID int64
	if err := testPool.QueryRow(ctx, `
		INSERT INTO gr33ncore.schedules
		    (farm_id, name, schedule_type, cron_expression, timezone, is_active)
		VALUES (1, 'ws3_smoke_schedule', 'cron_job', '0 0 * * *', 'UTC', FALSE)
		RETURNING id`).Scan(&scheduleID); err != nil {
		t.Fatalf("seed schedule: %v", err)
	}
	_, err := testPool.Exec(ctx, `
		INSERT INTO gr33ncore.executable_actions
		    (schedule_id, program_id, action_type, action_parameters)
		VALUES ($1, $2, 'create_task', '{"title":"two-source"}'::jsonb)`,
		scheduleID, programID)
	if err == nil {
		t.Fatalf("expected two-source INSERT to be rejected by chk_executable_source, but it succeeded")
	}
	if !strings.Contains(err.Error(), "chk_executable_source") {
		t.Fatalf("expected chk_executable_source violation, got: %v", err)
	}

	// ── 3. rules_handler rejects two-source writes at the API ─────────
	// Create a rule we can POST an action under.
	resp := authPost(t, tok, "/farms/1/automation/rules", map[string]any{
		"name":           uniqueName("ws3_rule"),
		"is_active":      false,
		"trigger_source": "manual_api_trigger",
	})
	expectStatus(t, resp, 201)
	rule := decodeMap(t, resp)
	ruleID := int64(rule["id"].(float64))

	resp = authPost(t, tok, fmt.Sprintf("/automation/rules/%d/actions", ruleID), map[string]any{
		"action_type":       "create_task",
		"action_parameters": map[string]any{"title": "ws3-two-source-api"},
		"program_id":        programID,
	})
	expectStatus(t, resp, 400)
}

// TestPhase2095CostEnergyColumns — Phase 20.95 WS2.
// Asserts additive column round-trips (input_definitions.unit_cost / currency / unit_id,
// input_batches.low_stock_threshold, cost_transactions.crop_cycle_id, actuators.watts)
// plus full CRUD for the new farm_energy_prices table and the three enum additions.

func TestScheduleCreateUpdateDelete(t *testing.T) {
	tok := smokeJWT(t)
	name := uniqueName("smoke_schedule")
	resp := authPost(t, tok, "/farms/1/schedules", map[string]any{
		"name":            name,
		"schedule_type":   "cron",
		"cron_expression": "0 6 * * *",
		"timezone":        "UTC",
		"is_active":       true,
	})
	expectStatus(t, resp, 201)
	created := decodeMap(t, resp)
	if created["name"] != name {
		t.Fatalf("expected name=%s, got %v", name, created["name"])
	}
	id := int64(created["id"].(float64))

	updatedName := uniqueName("smoke_schedule_upd")
	resp = authPut(t, tok, fmt.Sprintf("/schedules/%d", id), map[string]any{
		"name":            updatedName,
		"schedule_type":   "cron",
		"cron_expression": "0 8 * * *",
		"timezone":        "America/New_York",
		"is_active":       false,
	})
	expectStatus(t, resp, 200)
	updated := decodeMap(t, resp)
	if updated["name"] != updatedName {
		t.Fatalf("expected updated name=%s, got %v", updatedName, updated["name"])
	}
	if updated["is_active"] != false {
		t.Fatal("expected is_active=false after update")
	}

	resp = authDelete(t, tok, fmt.Sprintf("/schedules/%d", id))
	expectStatus(t, resp, 204)

	resp = authGet(t, tok, "/farms/1/schedules")
	expectStatus(t, resp, 200)
	schedList := decodeSlice(t, resp)
	for _, s := range schedList {
		if m, ok := s.(map[string]any); ok && m["name"] == updatedName {
			t.Fatal("deleted schedule still appears in list")
		}
	}
}

// ── Phase 16: Mixing Event Creation ─────────────────────────────────────────

func TestScheduleActiveToggle(t *testing.T) {
	tok := smokeJWT(t)

	name := uniqueName("smoke_toggle_sched")
	resp := authPost(t, tok, "/farms/1/schedules", map[string]any{
		"name":            name,
		"schedule_type":   "cron",
		"cron_expression": "0 12 * * *",
		"timezone":        "UTC",
		"is_active":       true,
	})
	expectStatus(t, resp, http.StatusCreated)
	created := decodeMap(t, resp)
	schedID := int64(created["id"].(float64))

	resp = authPatch(t, tok, fmt.Sprintf("/schedules/%d/active", schedID), map[string]any{
		"is_active": false,
	})
	expectStatus(t, resp, http.StatusOK)
	toggled := decodeMap(t, resp)
	if toggled["is_active"] != false {
		t.Fatal("expected is_active=false after toggle")
	}

	resp = authPatch(t, tok, fmt.Sprintf("/schedules/%d/active", schedID), map[string]any{
		"is_active": true,
	})
	expectStatus(t, resp, http.StatusOK)

	resp = authGet(t, tok, fmt.Sprintf("/schedules/%d/actuator-events", schedID))
	expectStatus(t, resp, http.StatusOK)
	_ = decodeSlice(t, resp)

	resp = authDelete(t, tok, fmt.Sprintf("/schedules/%d", schedID))
	expectStatus(t, resp, http.StatusNoContent)
}

// ── Phase 20 WS1: Automation Rule CRUD ──────────────────────────────────────

// TestAutomationRuleCRUD exercises the full CRUD surface for
// automation_rules and rule-bound executable_actions, plus the input
// validation the handler layers in front of the DB constraints:
//   - unknown trigger_source rejected at 400
//   - predicates that reference sensors on another farm rejected at 400
//   - deferred action_type values (http_webhook_call etc.) rejected at 400
//   - cascade-delete on automation_rules cleans up child actions and
//     nulls out tasks.source_rule_id rather than deleting the task.

func TestAutomationRuleCRUD(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	tok := smokeJWT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var unitID int64
	if err := testPool.QueryRow(ctx, `SELECT id FROM gr33ncore.units LIMIT 1`).Scan(&unitID); err != nil {
		t.Fatalf("find a unit id: %v", err)
	}

	// Seed a sensor on farm 1 to use in predicates.
	sensorName := uniqueName("rule_sensor")
	resp := authPost(t, tok, "/farms/1/sensors", map[string]any{
		"name":        sensorName,
		"sensor_type": "moisture",
		"unit_id":     unitID,
	})
	expectStatus(t, resp, http.StatusCreated)
	sensorRow := decodeMap(t, resp)
	sid := int64(sensorRow["id"].(float64))

	// --- Validation: unknown trigger_source is rejected. ---
	resp = authPost(t, tok, "/farms/1/automation/rules", map[string]any{
		"name":                  uniqueName("rule_bad_trigger"),
		"trigger_source":        "totally-bogus",
		"trigger_configuration": map[string]any{},
		"condition_logic":       "ALL",
		"conditions":            []map[string]any{},
	})
	expectStatus(t, resp, http.StatusBadRequest)
	resp.Body.Close()

	// --- Validation: predicate sensor not on this farm is rejected. ---
	resp = authPost(t, tok, "/farms/1/automation/rules", map[string]any{
		"name":            uniqueName("rule_foreign_sensor"),
		"trigger_source":  "manual_api_trigger",
		"condition_logic": "ALL",
		"conditions": []map[string]any{
			{"sensor_id": 99999999, "op": "gte", "value": 1.0},
		},
	})
	expectStatus(t, resp, http.StatusBadRequest)
	resp.Body.Close()

	// --- Validation: unknown precondition op rejected. ---
	resp = authPost(t, tok, "/farms/1/automation/rules", map[string]any{
		"name":            uniqueName("rule_bad_op"),
		"trigger_source":  "manual_api_trigger",
		"condition_logic": "ALL",
		"conditions": []map[string]any{
			{"sensor_id": sid, "op": "nope", "value": 1.0},
		},
	})
	expectStatus(t, resp, http.StatusBadRequest)
	resp.Body.Close()

	// --- Happy path: create a sensor_reading_threshold rule. ---
	ruleName := uniqueName("rule_crud")
	resp = authPost(t, tok, "/farms/1/automation/rules", map[string]any{
		"name":           ruleName,
		"description":    "smoke test rule",
		"is_active":      false,
		"trigger_source": "sensor_reading_threshold",
		"trigger_configuration": map[string]any{
			"sensor_id": sid,
			"op":        "lt",
			"value":     10.0,
		},
		"condition_logic": "ALL",
		"conditions": []map[string]any{
			{"sensor_id": sid, "op": "lt", "value": 10.0},
		},
		"cooldown_period_seconds": 60,
	})
	expectStatus(t, resp, http.StatusCreated)
	created := decodeMap(t, resp)
	ruleID := int64(created["id"].(float64))
	if created["name"] != ruleName {
		t.Fatalf("expected name=%s, got %v", ruleName, created["name"])
	}
	if created["is_active"] != false {
		t.Fatal("expected is_active=false on created rule")
	}

	// GET by id.
	resp = authGet(t, tok, fmt.Sprintf("/automation/rules/%d", ruleID))
	expectStatus(t, resp, http.StatusOK)
	_ = decodeMap(t, resp)

	// List by farm includes it.
	resp = authGet(t, tok, "/farms/1/automation/rules")
	expectStatus(t, resp, http.StatusOK)
	ruleList := decodeSlice(t, resp)
	found := false
	for _, r := range ruleList {
		if m, ok := r.(map[string]any); ok && int64(m["id"].(float64)) == ruleID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("rule %d missing from farm list", ruleID)
	}

	// Toggle active.
	resp = authPatch(t, tok, fmt.Sprintf("/automation/rules/%d/active", ruleID), map[string]any{
		"is_active": true,
	})
	expectStatus(t, resp, http.StatusOK)
	toggled := decodeMap(t, resp)
	if toggled["is_active"] != true {
		t.Fatal("expected is_active=true after toggle")
	}

	// Full update — change cooldown + predicate value.
	resp = authPut(t, tok, fmt.Sprintf("/automation/rules/%d", ruleID), map[string]any{
		"name":           ruleName,
		"is_active":      true,
		"trigger_source": "sensor_reading_threshold",
		"trigger_configuration": map[string]any{
			"sensor_id": sid,
			"op":        "lt",
			"value":     5.0,
		},
		"condition_logic": "ANY",
		"conditions": []map[string]any{
			{"sensor_id": sid, "op": "lt", "value": 5.0},
		},
		"cooldown_period_seconds": 120,
	})
	expectStatus(t, resp, http.StatusOK)
	updated := decodeMap(t, resp)
	if cp, _ := updated["cooldown_period_seconds"].(float64); int(cp) != 120 {
		t.Fatalf("expected cooldown_period_seconds=120, got %v", updated["cooldown_period_seconds"])
	}

	// --- Deferred action types MUST be rejected with 400. ---
	for _, deferred := range []string{
		"trigger_another_automation_rule",
		"http_webhook_call",
		"update_record_in_gr33n",
		"log_custom_event",
	} {
		resp = authPost(t, tok, fmt.Sprintf("/automation/rules/%d/actions", ruleID), map[string]any{
			"execution_order": 0,
			"action_type":     deferred,
		})
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400 for deferred action_type=%s, got %d", deferred, resp.StatusCode)
		}
		resp.Body.Close()
	}

	// --- Happy path: attach a create_task action (no actuator needed). ---
	resp = authPost(t, tok, fmt.Sprintf("/automation/rules/%d/actions", ruleID), map[string]any{
		"execution_order": 1,
		"action_type":     "create_task",
		"action_parameters": map[string]any{
			"title":    "auto-generated smoke task",
			"priority": 1,
		},
	})
	expectStatus(t, resp, http.StatusCreated)
	action := decodeMap(t, resp)
	actionID := int64(action["id"].(float64))

	// Action missing required shape is rejected.
	resp = authPost(t, tok, fmt.Sprintf("/automation/rules/%d/actions", ruleID), map[string]any{
		"execution_order":   2,
		"action_type":       "create_task",
		"action_parameters": map[string]any{}, // empty payload
	})
	expectStatus(t, resp, http.StatusBadRequest)
	resp.Body.Close()

	// List actions on the rule.
	resp = authGet(t, tok, fmt.Sprintf("/automation/rules/%d/actions", ruleID))
	expectStatus(t, resp, http.StatusOK)
	actionList := decodeSlice(t, resp)
	if len(actionList) == 0 {
		t.Fatal("expected at least one action on the rule")
	}

	// Update the action — bump execution_order.
	resp = authPut(t, tok, fmt.Sprintf("/automation/actions/%d", actionID), map[string]any{
		"execution_order": 5,
		"action_type":     "create_task",
		"action_parameters": map[string]any{
			"title":    "auto-generated smoke task (updated)",
			"priority": 2,
		},
	})
	expectStatus(t, resp, http.StatusOK)

	// --- Seed a task with source_rule_id set to the rule and verify ON DELETE
	// SET NULL behavior when the rule goes away. ---
	resp = authPost(t, tok, "/farms/1/tasks", map[string]any{
		"title":          uniqueName("task_from_rule"),
		"priority":       1,
		"source_rule_id": ruleID,
	})
	expectStatus(t, resp, http.StatusCreated)
	linkedTask := decodeMap(t, resp)
	linkedTaskID := int64(linkedTask["id"].(float64))
	if srid, _ := linkedTask["source_rule_id"].(float64); int64(srid) != ruleID {
		t.Fatalf("expected source_rule_id=%d, got %v", ruleID, linkedTask["source_rule_id"])
	}

	// Delete the rule — cascades to actions, nulls source_rule_id on tasks.
	resp = authDelete(t, tok, fmt.Sprintf("/automation/rules/%d", ruleID))
	expectStatus(t, resp, http.StatusNoContent)

	// Rule is gone.
	resp = authGet(t, tok, fmt.Sprintf("/automation/rules/%d", ruleID))
	expectStatus(t, resp, http.StatusNotFound)
	resp.Body.Close()

	// Child action was cascade-deleted.
	var actionCount int
	if err := testPool.QueryRow(ctx,
		`SELECT COUNT(*) FROM gr33ncore.executable_actions WHERE id = $1`, actionID,
	).Scan(&actionCount); err != nil {
		t.Fatalf("count actions after rule delete: %v", err)
	}
	if actionCount != 0 {
		t.Fatalf("expected 0 actions after rule delete, got %d", actionCount)
	}

	// Task still exists but with source_rule_id nulled out.
	var nullCount int
	if err := testPool.QueryRow(ctx,
		`SELECT COUNT(*) FROM gr33ncore.tasks WHERE id = $1 AND source_rule_id IS NULL`,
		linkedTaskID,
	).Scan(&nullCount); err != nil {
		t.Fatalf("check task source_rule_id after rule delete: %v", err)
	}
	if nullCount != 1 {
		t.Fatalf("expected task %d to remain with source_rule_id=NULL, got %d", linkedTaskID, nullCount)
	}
}

// ── Phase 20 WS2: Rule evaluator ────────────────────────────────────────────

// seedRuleSensorWithReading creates a sensor on farm 1 and seeds a single
// reading. Returns the sensor id. Test helper for WS2 rule tick tests.

func seedRuleSensorWithReading(t *testing.T, tok string, unitID int64, value float64) int64 {
	t.Helper()
	resp := authPost(t, tok, "/farms/1/sensors", map[string]any{
		"name":        uniqueName("rule_tick_sensor"),
		"sensor_type": "moisture",
		"unit_id":     unitID,
	})
	expectStatus(t, resp, http.StatusCreated)
	row := decodeMap(t, resp)
	sid := int64(row["id"].(float64))
	if _, err := testPool.Exec(context.Background(), `
		INSERT INTO gr33ncore.sensor_readings (reading_time, sensor_id, value_raw, is_valid)
		VALUES (NOW(), $1, $2, TRUE)`, sid, value); err != nil {
		t.Fatalf("seed reading for sensor %d: %v", sid, err)
	}
	return sid
}

// TestAutomationRuleTickALLvsANY verifies the rule evaluator honors
// condition_logic. One predicate passes (sensor reads 5, predicate lt 10)
// and one fails (other sensor reads 50, predicate lt 10). Under ALL the
// rule must skip with message=conditions_not_met; under ANY the rule must
// fire (success). last_evaluated_time must be stamped in both cases.

func TestAutomationRuleTickALLvsANY(t *testing.T) {
	if testPool == nil || testWorker == nil {
		t.Skip("testPool/testWorker unavailable")
	}
	tok := smokeJWT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var unitID int64
	if err := testPool.QueryRow(ctx, `SELECT id FROM gr33ncore.units LIMIT 1`).Scan(&unitID); err != nil {
		t.Fatalf("find a unit id: %v", err)
	}

	// Sensor A reads 5 (predicate lt 10 → passes).
	// Sensor B reads 50 (predicate lt 10 → fails).
	sidA := seedRuleSensorWithReading(t, tok, unitID, 5)
	sidB := seedRuleSensorWithReading(t, tok, unitID, 50)

	resp := authPost(t, tok, "/farms/1/automation/rules", map[string]any{
		"name":            uniqueName("rule_all_any"),
		"is_active":       true,
		"trigger_source":  "manual_api_trigger",
		"condition_logic": "ALL",
		"conditions": []map[string]any{
			{"sensor_id": sidA, "op": "lt", "value": 10.0},
			{"sensor_id": sidB, "op": "lt", "value": 10.0},
		},
	})
	expectStatus(t, resp, http.StatusCreated)
	created := decodeMap(t, resp)
	ruleID := int64(created["id"].(float64))

	// --- ALL: must skip with conditions_not_met. ---
	testWorker.TickRules(ctx)

	var status, msg string
	var detailsJSON []byte
	if err := testPool.QueryRow(ctx, `
		SELECT status, COALESCE(message, ''), details::text
		FROM gr33ncore.automation_runs
		WHERE rule_id = $1
		ORDER BY id DESC LIMIT 1`, ruleID).Scan(&status, &msg, &detailsJSON); err != nil {
		t.Fatalf("read latest rule run (ALL): %v", err)
	}
	if status != "skipped" || msg != "conditions_not_met" {
		t.Fatalf("expected ALL tick to skip with conditions_not_met, got status=%s msg=%s", status, msg)
	}
	var details struct {
		Phase         string `json:"phase"`
		Logic         string `json:"logic"`
		ConditionsMet bool   `json:"conditions_met"`
		Failed        []struct {
			SensorID int64   `json:"sensor_id"`
			Reason   string  `json:"reason"`
			Expected float64 `json:"expected"`
		} `json:"failed"`
	}
	if err := json.Unmarshal(detailsJSON, &details); err != nil {
		t.Fatalf("parse details: %v (raw=%s)", err, detailsJSON)
	}
	if details.Phase != "conditions" || details.Logic != "ALL" || details.ConditionsMet {
		t.Fatalf("unexpected details on ALL skip: %+v", details)
	}
	if len(details.Failed) != 1 || details.Failed[0].SensorID != sidB {
		t.Fatalf("expected exactly sensor B (%d) to be in failed list, got %+v", sidB, details.Failed)
	}

	// last_evaluated_time must be stamped even on skip.
	var lastEval *time.Time
	if err := testPool.QueryRow(ctx,
		`SELECT last_evaluated_time FROM gr33ncore.automation_rules WHERE id = $1`, ruleID,
	).Scan(&lastEval); err != nil {
		t.Fatalf("read last_evaluated_time: %v", err)
	}
	if lastEval == nil {
		t.Fatal("expected last_evaluated_time to be set after a skip tick")
	}
	// Must NOT fire: last_triggered_time stays NULL.
	var lastTrig *time.Time
	if err := testPool.QueryRow(ctx,
		`SELECT last_triggered_time FROM gr33ncore.automation_rules WHERE id = $1`, ruleID,
	).Scan(&lastTrig); err != nil {
		t.Fatalf("read last_triggered_time: %v", err)
	}
	if lastTrig != nil {
		t.Fatalf("expected last_triggered_time to stay NULL after ALL skip, got %v", *lastTrig)
	}

	// --- Flip to ANY: must fire. ---
	resp = authPut(t, tok, fmt.Sprintf("/automation/rules/%d", ruleID), map[string]any{
		"name":            created["name"],
		"is_active":       true,
		"trigger_source":  "manual_api_trigger",
		"condition_logic": "ANY",
		"conditions": []map[string]any{
			{"sensor_id": sidA, "op": "lt", "value": 10.0},
			{"sensor_id": sidB, "op": "lt", "value": 10.0},
		},
	})
	expectStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	testWorker.TickRules(ctx)

	if err := testPool.QueryRow(ctx, `
		SELECT status, COALESCE(message, '')
		FROM gr33ncore.automation_runs
		WHERE rule_id = $1
		ORDER BY id DESC LIMIT 1`, ruleID).Scan(&status, &msg); err != nil {
		t.Fatalf("read latest rule run (ANY): %v", err)
	}
	// No actions attached → the evaluator records a skipped run with
	// message="rule has no executable actions" after conditions met.
	// Either way the fire path ran: last_triggered_time MUST be set.
	if err := testPool.QueryRow(ctx,
		`SELECT last_triggered_time FROM gr33ncore.automation_rules WHERE id = $1`, ruleID,
	).Scan(&lastTrig); err != nil {
		t.Fatalf("read last_triggered_time after ANY tick: %v", err)
	}
	if lastTrig == nil {
		t.Fatal("expected last_triggered_time to be stamped after ANY tick with conditions met")
	}
	if status == "skipped" && msg == "conditions_not_met" {
		t.Fatalf("ANY tick incorrectly reported conditions_not_met")
	}
}

// TestAutomationRuleTickCooldown verifies cooldown_period_seconds:
//  1. First tick satisfies conditions → fires (last_triggered_time set).
//  2. Second tick within the cooldown window → skipped with message=cooldown
//     and last_triggered_time NOT advanced.
//  3. After rolling last_triggered_time back past the cooldown window, the
//     next tick fires again.

func TestAutomationRuleTickCooldown(t *testing.T) {
	if testPool == nil || testWorker == nil {
		t.Skip("testPool/testWorker unavailable")
	}
	tok := smokeJWT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var unitID int64
	if err := testPool.QueryRow(ctx, `SELECT id FROM gr33ncore.units LIMIT 1`).Scan(&unitID); err != nil {
		t.Fatalf("find a unit id: %v", err)
	}
	sid := seedRuleSensorWithReading(t, tok, unitID, 5)

	resp := authPost(t, tok, "/farms/1/automation/rules", map[string]any{
		"name":            uniqueName("rule_cooldown"),
		"is_active":       true,
		"trigger_source":  "manual_api_trigger",
		"condition_logic": "ALL",
		"conditions": []map[string]any{
			{"sensor_id": sid, "op": "lt", "value": 10.0},
		},
		"cooldown_period_seconds": 300,
	})
	expectStatus(t, resp, http.StatusCreated)
	created := decodeMap(t, resp)
	ruleID := int64(created["id"].(float64))

	// Tick 1: fires.
	testWorker.TickRules(ctx)
	var firstTrig *time.Time
	if err := testPool.QueryRow(ctx,
		`SELECT last_triggered_time FROM gr33ncore.automation_rules WHERE id = $1`, ruleID,
	).Scan(&firstTrig); err != nil {
		t.Fatalf("read last_triggered_time after fire: %v", err)
	}
	if firstTrig == nil {
		t.Fatal("expected last_triggered_time to be set after first tick")
	}

	// Tick 2: in cooldown, must skip with message=cooldown.
	testWorker.TickRules(ctx)
	var status, msg string
	var detailsJSON []byte
	if err := testPool.QueryRow(ctx, `
		SELECT status, COALESCE(message, ''), details::text
		FROM gr33ncore.automation_runs
		WHERE rule_id = $1
		ORDER BY id DESC LIMIT 1`, ruleID).Scan(&status, &msg, &detailsJSON); err != nil {
		t.Fatalf("read latest rule run after second tick: %v", err)
	}
	if status != "skipped" || msg != "cooldown" {
		t.Fatalf("expected second tick to skip with cooldown, got status=%s msg=%s", status, msg)
	}
	var cdDetails struct {
		Phase            string `json:"phase"`
		CooldownSeconds  int    `json:"cooldown_seconds"`
		RemainingSeconds int    `json:"remaining_seconds"`
	}
	if err := json.Unmarshal(detailsJSON, &cdDetails); err != nil {
		t.Fatalf("parse cooldown details: %v", err)
	}
	if cdDetails.Phase != "cooldown" || cdDetails.CooldownSeconds != 300 {
		t.Fatalf("unexpected cooldown details: %+v", cdDetails)
	}

	// last_triggered_time must not have moved forward on a cooldown skip.
	var secondTrig *time.Time
	if err := testPool.QueryRow(ctx,
		`SELECT last_triggered_time FROM gr33ncore.automation_rules WHERE id = $1`, ruleID,
	).Scan(&secondTrig); err != nil {
		t.Fatalf("re-read last_triggered_time: %v", err)
	}
	if secondTrig == nil || !secondTrig.Equal(*firstTrig) {
		t.Fatalf("expected last_triggered_time to stay %v after cooldown skip, got %v", firstTrig, secondTrig)
	}

	// Roll last_triggered_time back past the cooldown window. After this the
	// next tick MUST fire again.
	if _, err := testPool.Exec(ctx,
		`UPDATE gr33ncore.automation_rules SET last_triggered_time = NOW() - INTERVAL '10 minutes' WHERE id = $1`,
		ruleID,
	); err != nil {
		t.Fatalf("rewind last_triggered_time: %v", err)
	}

	testWorker.TickRules(ctx)
	var thirdTrig *time.Time
	if err := testPool.QueryRow(ctx,
		`SELECT last_triggered_time FROM gr33ncore.automation_rules WHERE id = $1`, ruleID,
	).Scan(&thirdTrig); err != nil {
		t.Fatalf("read last_triggered_time after third tick: %v", err)
	}
	if thirdTrig == nil {
		t.Fatal("expected last_triggered_time to be set after third tick")
	}
	if !thirdTrig.After(*firstTrig) {
		t.Fatalf("expected third-tick last_triggered_time (%v) to advance past first-tick (%v) after cooldown window elapses", thirdTrig, firstTrig)
	}
}

// TestAutomationRuleTickInactiveRuleSkipped verifies ListActiveAutomationRules
// does not return rules with is_active=false, so the evaluator never touches
// them. Negative bookkeeping test — ensures last_evaluated_time STAYS null
// for an inactive rule even after several ticks.

func TestAutomationRuleTickInactiveRuleSkipped(t *testing.T) {
	if testPool == nil || testWorker == nil {
		t.Skip("testPool/testWorker unavailable")
	}
	tok := smokeJWT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var unitID int64
	if err := testPool.QueryRow(ctx, `SELECT id FROM gr33ncore.units LIMIT 1`).Scan(&unitID); err != nil {
		t.Fatalf("find a unit id: %v", err)
	}
	sid := seedRuleSensorWithReading(t, tok, unitID, 5)

	resp := authPost(t, tok, "/farms/1/automation/rules", map[string]any{
		"name":            uniqueName("rule_inactive"),
		"is_active":       false,
		"trigger_source":  "manual_api_trigger",
		"condition_logic": "ALL",
		"conditions": []map[string]any{
			{"sensor_id": sid, "op": "lt", "value": 10.0},
		},
	})
	expectStatus(t, resp, http.StatusCreated)
	created := decodeMap(t, resp)
	ruleID := int64(created["id"].(float64))

	testWorker.TickRules(ctx)
	testWorker.TickRules(ctx)

	var lastEval, lastTrig *time.Time
	if err := testPool.QueryRow(ctx,
		`SELECT last_evaluated_time, last_triggered_time FROM gr33ncore.automation_rules WHERE id = $1`,
		ruleID,
	).Scan(&lastEval, &lastTrig); err != nil {
		t.Fatalf("read rule times: %v", err)
	}
	if lastEval != nil || lastTrig != nil {
		t.Fatalf("expected last_evaluated_time and last_triggered_time to stay NULL for inactive rule, got eval=%v trig=%v", lastEval, lastTrig)
	}
	var runCount int
	if err := testPool.QueryRow(ctx,
		`SELECT COUNT(*) FROM gr33ncore.automation_runs WHERE rule_id = $1`, ruleID,
	).Scan(&runCount); err != nil {
		t.Fatalf("count runs: %v", err)
	}
	if runCount != 0 {
		t.Fatalf("expected 0 runs for inactive rule, got %d", runCount)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Phase 20 / WS3 — rule action dispatchers
//
// These tests drive the worker through TickRules and assert the DB side
// effects of each supported `executable_action_type`. The worker runs in
// simulation mode in the test bootstrap, so `control_actuator` dispatches
// go through the "simulated" leg (status = execution_completed_success_on_device,
// no device pending command).
// ─────────────────────────────────────────────────────────────────────────────

// seedRuleActuator inserts a bare actuator row on farm 1 and returns its id.
// Smoke tests don't have a POST /actuators endpoint, so we fabricate one via
// direct SQL. Device/zone are both nullable on the schema, so this minimal row
// is enough for the worker to dispatch against.

func seedRuleActuator(t *testing.T, name string) int64 {
	t.Helper()
	var id int64
	if err := testPool.QueryRow(context.Background(), `
		INSERT INTO gr33ncore.actuators (farm_id, name, actuator_type)
		VALUES (1, $1, 'relay')
		RETURNING id`, name).Scan(&id); err != nil {
		t.Fatalf("seed actuator: %v", err)
	}
	return id
}

// seedRuleNotificationTemplate inserts a per-farm notification template and
// returns its id. The rule evaluator resolves the template by id and uses
// its subject/body for the rendered alerts_notifications row.

func seedRuleNotificationTemplate(t *testing.T, key, subject, body string, priority string) int64 {
	t.Helper()
	var id int64
	if err := testPool.QueryRow(context.Background(), `
		INSERT INTO gr33ncore.notification_templates
		  (farm_id, template_key, subject_template, body_template_text, default_priority)
		VALUES (1, $1, $2, $3, $4::gr33ncore.notification_priority_enum)
		RETURNING id`, key, subject, body, priority).Scan(&id); err != nil {
		t.Fatalf("seed notification template: %v", err)
	}
	return id
}

// TestAutomationRuleDispatchControlActuator verifies the control_actuator
// dispatcher:
//  1. One tick with conditions met writes a gr33ncore.actuator_events row
//     whose `triggered_by_rule_id` is the rule id and `source` is
//     'automation_rule_trigger'.
//  2. The run is recorded as status=success with actions_total=actions_success=1.
//  3. Rule `last_triggered_time` advances.

func TestAutomationRuleDispatchControlActuator(t *testing.T) {
	if testPool == nil || testWorker == nil {
		t.Skip("testPool/testWorker unavailable")
	}
	tok := smokeJWT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var unitID int64
	if err := testPool.QueryRow(ctx, `SELECT id FROM gr33ncore.units LIMIT 1`).Scan(&unitID); err != nil {
		t.Fatalf("find a unit id: %v", err)
	}

	sid := seedRuleSensorWithReading(t, tok, unitID, 5)
	actID := seedRuleActuator(t, uniqueName("ws3_actuator"))

	resp := authPost(t, tok, "/farms/1/automation/rules", map[string]any{
		"name":            uniqueName("rule_ws3_actuator"),
		"is_active":       true,
		"trigger_source":  "manual_api_trigger",
		"condition_logic": "ALL",
		"conditions": []map[string]any{
			{"sensor_id": sid, "op": "lt", "value": 10.0},
		},
	})
	expectStatus(t, resp, http.StatusCreated)
	ruleID := int64(decodeMap(t, resp)["id"].(float64))

	resp = authPost(t, tok, fmt.Sprintf("/automation/rules/%d/actions", ruleID), map[string]any{
		"execution_order":    0,
		"action_type":        "control_actuator",
		"target_actuator_id": actID,
		"action_command":     "on",
	})
	expectStatus(t, resp, http.StatusCreated)
	resp.Body.Close()

	testWorker.TickRules(ctx)

	// Run bookkeeping: status=success, actions_success=1.
	var status, msg string
	var detailsJSON []byte
	if err := testPool.QueryRow(ctx, `
		SELECT status, COALESCE(message, ''), details::text
		FROM gr33ncore.automation_runs
		WHERE rule_id = $1
		ORDER BY id DESC LIMIT 1`, ruleID).Scan(&status, &msg, &detailsJSON); err != nil {
		t.Fatalf("read latest rule run: %v", err)
	}
	if status != "success" {
		t.Fatalf("expected success, got status=%s msg=%s details=%s", status, msg, detailsJSON)
	}
	var details struct {
		Phase          string `json:"phase"`
		ActionsTotal   int    `json:"actions_total"`
		ActionsSuccess int    `json:"actions_success"`
	}
	if err := json.Unmarshal(detailsJSON, &details); err != nil {
		t.Fatalf("parse details: %v", err)
	}
	if details.Phase != "actions" || details.ActionsTotal != 1 || details.ActionsSuccess != 1 {
		t.Fatalf("unexpected details: %+v (raw=%s)", details, detailsJSON)
	}

	// Side effect: one actuator_events row stamped with this rule.
	var eventCount int
	var eventSource, commandSent string
	if err := testPool.QueryRow(ctx, `
		SELECT COUNT(*), COALESCE(MAX(source::text), ''), COALESCE(MAX(command_sent), '')
		FROM gr33ncore.actuator_events
		WHERE triggered_by_rule_id = $1 AND actuator_id = $2`,
		ruleID, actID,
	).Scan(&eventCount, &eventSource, &commandSent); err != nil {
		t.Fatalf("count actuator events: %v", err)
	}
	if eventCount != 1 {
		t.Fatalf("expected exactly 1 actuator_events row for rule %d, got %d", ruleID, eventCount)
	}
	if eventSource != "automation_rule_trigger" {
		t.Fatalf("expected source=automation_rule_trigger, got %s", eventSource)
	}
	if commandSent != "on" {
		t.Fatalf("expected command_sent=on, got %s", commandSent)
	}

	// last_triggered_time advanced.
	var lastTrig *time.Time
	if err := testPool.QueryRow(ctx,
		`SELECT last_triggered_time FROM gr33ncore.automation_rules WHERE id = $1`, ruleID,
	).Scan(&lastTrig); err != nil {
		t.Fatalf("read last_triggered_time: %v", err)
	}
	if lastTrig == nil {
		t.Fatal("expected last_triggered_time to be stamped after successful dispatch")
	}
}

// TestAutomationRuleDispatchCreateTask verifies the create_task dispatcher:
//  1. A ticked rule inserts a task with source_rule_id pointing back at the rule.
//  2. action_parameters.{title,priority,due_in_days} are honored.
//  3. The run is recorded success with actions_success=1.

func TestAutomationRuleDispatchCreateTask(t *testing.T) {
	if testPool == nil || testWorker == nil {
		t.Skip("testPool/testWorker unavailable")
	}
	tok := smokeJWT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var unitID int64
	if err := testPool.QueryRow(ctx, `SELECT id FROM gr33ncore.units LIMIT 1`).Scan(&unitID); err != nil {
		t.Fatalf("find a unit id: %v", err)
	}

	sid := seedRuleSensorWithReading(t, tok, unitID, 5)

	resp := authPost(t, tok, "/farms/1/automation/rules", map[string]any{
		"name":            uniqueName("rule_ws3_task"),
		"is_active":       true,
		"trigger_source":  "manual_api_trigger",
		"condition_logic": "ALL",
		"conditions": []map[string]any{
			{"sensor_id": sid, "op": "lt", "value": 10.0},
		},
	})
	expectStatus(t, resp, http.StatusCreated)
	ruleID := int64(decodeMap(t, resp)["id"].(float64))

	taskTitle := uniqueName("ws3_task_title")
	resp = authPost(t, tok, fmt.Sprintf("/automation/rules/%d/actions", ruleID), map[string]any{
		"execution_order": 0,
		"action_type":     "create_task",
		"action_parameters": map[string]any{
			"title":       taskTitle,
			"priority":    2,
			"due_in_days": 1,
			"task_type":   "inspection",
		},
	})
	expectStatus(t, resp, http.StatusCreated)
	resp.Body.Close()

	testWorker.TickRules(ctx)

	var status string
	if err := testPool.QueryRow(ctx,
		`SELECT status FROM gr33ncore.automation_runs WHERE rule_id = $1 ORDER BY id DESC LIMIT 1`, ruleID,
	).Scan(&status); err != nil {
		t.Fatalf("read run: %v", err)
	}
	if status != "success" {
		t.Fatalf("expected run status=success, got %s", status)
	}

	// The generated task carries source_rule_id and our parameters.
	var gotTitle, gotType string
	var gotPriority int32
	var gotDue *time.Time
	var gotSourceRuleID *int64
	if err := testPool.QueryRow(ctx, `
		SELECT title, COALESCE(task_type, ''), COALESCE(priority, 0), due_date, source_rule_id
		FROM gr33ncore.tasks
		WHERE source_rule_id = $1
		ORDER BY id DESC LIMIT 1`, ruleID,
	).Scan(&gotTitle, &gotType, &gotPriority, &gotDue, &gotSourceRuleID); err != nil {
		t.Fatalf("read generated task: %v", err)
	}
	if gotTitle != taskTitle {
		t.Fatalf("expected task title %q, got %q", taskTitle, gotTitle)
	}
	if gotType != "inspection" {
		t.Fatalf("expected task_type=inspection, got %q", gotType)
	}
	if gotPriority != 2 {
		t.Fatalf("expected priority=2, got %d", gotPriority)
	}
	if gotDue == nil {
		t.Fatal("expected due_date to be set from due_in_days=1")
	}
	if gotSourceRuleID == nil || *gotSourceRuleID != ruleID {
		t.Fatalf("expected source_rule_id=%d, got %v", ruleID, gotSourceRuleID)
	}
}

// TestAutomationRuleDispatchSendNotification verifies the send_notification
// dispatcher:
//  1. The template's subject/body are rendered into alerts_notifications.
//  2. notification_template_id and triggering_event_source_type='automation_rule'
//     are set on the inserted alert.
//  3. Severity defaults to the template's default_priority.

func TestAutomationRuleDispatchSendNotification(t *testing.T) {
	if testPool == nil || testWorker == nil {
		t.Skip("testPool/testWorker unavailable")
	}
	tok := smokeJWT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var unitID int64
	if err := testPool.QueryRow(ctx, `SELECT id FROM gr33ncore.units LIMIT 1`).Scan(&unitID); err != nil {
		t.Fatalf("find a unit id: %v", err)
	}
	sid := seedRuleSensorWithReading(t, tok, unitID, 5)

	tmplKey := uniqueName("ws3_tmpl")
	tmplID := seedRuleNotificationTemplate(t,
		tmplKey,
		"Alert from rule {{rule_name}}",
		"Sensor reading triggered rule {{rule_id}}",
		"high",
	)

	resp := authPost(t, tok, "/farms/1/automation/rules", map[string]any{
		"name":            uniqueName("rule_ws3_notify"),
		"is_active":       true,
		"trigger_source":  "manual_api_trigger",
		"condition_logic": "ALL",
		"conditions": []map[string]any{
			{"sensor_id": sid, "op": "lt", "value": 10.0},
		},
	})
	expectStatus(t, resp, http.StatusCreated)
	created := decodeMap(t, resp)
	ruleID := int64(created["id"].(float64))
	ruleName := created["name"].(string)

	resp = authPost(t, tok, fmt.Sprintf("/automation/rules/%d/actions", ruleID), map[string]any{
		"execution_order":                 0,
		"action_type":                     "send_notification",
		"target_notification_template_id": tmplID,
	})
	expectStatus(t, resp, http.StatusCreated)
	resp.Body.Close()

	testWorker.TickRules(ctx)

	var status string
	if err := testPool.QueryRow(ctx,
		`SELECT status FROM gr33ncore.automation_runs WHERE rule_id = $1 ORDER BY id DESC LIMIT 1`, ruleID,
	).Scan(&status); err != nil {
		t.Fatalf("read run: %v", err)
	}
	if status != "success" {
		t.Fatalf("expected run status=success, got %s", status)
	}

	var subject, body, srcType, severity string
	var gotTmpl, gotSrcID *int64
	if err := testPool.QueryRow(ctx, `
		SELECT COALESCE(subject_rendered, ''), COALESCE(message_text_rendered, ''),
		       COALESCE(triggering_event_source_type, ''), severity::text,
		       notification_template_id, triggering_event_source_id
		FROM gr33ncore.alerts_notifications
		WHERE notification_template_id = $1 AND triggering_event_source_id = $2
		ORDER BY id DESC LIMIT 1`, tmplID, ruleID,
	).Scan(&subject, &body, &srcType, &severity, &gotTmpl, &gotSrcID); err != nil {
		t.Fatalf("read alert: %v", err)
	}
	expectedSubject := "Alert from rule " + ruleName
	if subject != expectedSubject {
		t.Fatalf("expected subject %q, got %q", expectedSubject, subject)
	}
	expectedBody := fmt.Sprintf("Sensor reading triggered rule %d", ruleID)
	if body != expectedBody {
		t.Fatalf("expected body %q, got %q", expectedBody, body)
	}
	if srcType != "automation_rule" {
		t.Fatalf("expected triggering_event_source_type=automation_rule, got %s", srcType)
	}
	if severity != "high" {
		t.Fatalf("expected severity=high (from template default_priority), got %s", severity)
	}
	if gotTmpl == nil || *gotTmpl != tmplID {
		t.Fatalf("expected notification_template_id=%d, got %v", tmplID, gotTmpl)
	}
	if gotSrcID == nil || *gotSrcID != ruleID {
		t.Fatalf("expected triggering_event_source_id=%d, got %v", ruleID, gotSrcID)
	}

	// The worker also fans the alert through the push pipeline. The
	// test wires a recording PushNotifier that captures every dispatched
	// alert — one rule fire should produce exactly one push dispatch
	// stamped with this rule's id.
	if got := testNotifier.countForRule(ruleID); got != 1 {
		t.Fatalf("expected push notifier to receive 1 alert for rule %d, got %d", ruleID, got)
	}
}

// TestAutomationRuleDispatchPartialSuccess verifies that when a rule has
// multiple actions and one of them fails at dispatch time, the run is
// recorded as `partial_success` with details.errors[] populated, and
// the successful action's side effect still lands.
//
// We fabricate the failure by direct-inserting a deferred-type action
// (log_custom_event) into executable_actions — the API CRUD validator
// rejects these, but the DB CHECK constraint permits them, so this is
// the realistic "row written by a newer binary, read by a worker that
// doesn't know that action type" path. The sibling `create_task`
// action still runs and records its side effect.

func TestAutomationRuleDispatchPartialSuccess(t *testing.T) {
	if testPool == nil || testWorker == nil {
		t.Skip("testPool/testWorker unavailable")
	}
	tok := smokeJWT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var unitID int64
	if err := testPool.QueryRow(ctx, `SELECT id FROM gr33ncore.units LIMIT 1`).Scan(&unitID); err != nil {
		t.Fatalf("find a unit id: %v", err)
	}
	sid := seedRuleSensorWithReading(t, tok, unitID, 5)

	resp := authPost(t, tok, "/farms/1/automation/rules", map[string]any{
		"name":            uniqueName("rule_ws3_partial"),
		"is_active":       true,
		"trigger_source":  "manual_api_trigger",
		"condition_logic": "ALL",
		"conditions": []map[string]any{
			{"sensor_id": sid, "op": "lt", "value": 10.0},
		},
	})
	expectStatus(t, resp, http.StatusCreated)
	ruleID := int64(decodeMap(t, resp)["id"].(float64))

	// Direct-insert a deferred-type action the worker doesn't know about.
	// Bypasses the CRUD validator (which would 400) but respects the DB
	// CHECK constraint that log_custom_event needs action_parameters.
	var brokenActionID int64
	if err := testPool.QueryRow(ctx, `
		INSERT INTO gr33ncore.executable_actions
		  (rule_id, execution_order, action_type, action_parameters)
		VALUES ($1, 0, 'log_custom_event', '{"note":"forced failure"}'::jsonb)
		RETURNING id`, ruleID,
	).Scan(&brokenActionID); err != nil {
		t.Fatalf("seed deferred action: %v", err)
	}

	taskTitle := uniqueName("ws3_partial_task")
	resp = authPost(t, tok, fmt.Sprintf("/automation/rules/%d/actions", ruleID), map[string]any{
		"execution_order":   1,
		"action_type":       "create_task",
		"action_parameters": map[string]any{"title": taskTitle},
	})
	expectStatus(t, resp, http.StatusCreated)
	resp.Body.Close()

	testWorker.TickRules(ctx)

	var status, msg string
	var detailsJSON []byte
	if err := testPool.QueryRow(ctx, `
		SELECT status, COALESCE(message, ''), details::text
		FROM gr33ncore.automation_runs
		WHERE rule_id = $1
		ORDER BY id DESC LIMIT 1`, ruleID).Scan(&status, &msg, &detailsJSON); err != nil {
		t.Fatalf("read run: %v", err)
	}
	if status != "partial_success" {
		t.Fatalf("expected partial_success, got status=%s msg=%s details=%s", status, msg, detailsJSON)
	}
	var details struct {
		ActionsTotal   int `json:"actions_total"`
		ActionsSuccess int `json:"actions_success"`
		Errors         []struct {
			ActionID int64  `json:"action_id"`
			Message  string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(detailsJSON, &details); err != nil {
		t.Fatalf("parse details: %v", err)
	}
	if details.ActionsTotal != 2 || details.ActionsSuccess != 1 {
		t.Fatalf("expected 2 total / 1 success, got %+v", details)
	}
	if len(details.Errors) != 1 || details.Errors[0].ActionID != brokenActionID {
		t.Fatalf("expected single error for action %d, got %+v", brokenActionID, details.Errors)
	}

	// The create_task action still landed its task.
	var gotTitle string
	if err := testPool.QueryRow(ctx,
		`SELECT title FROM gr33ncore.tasks WHERE source_rule_id = $1 ORDER BY id DESC LIMIT 1`, ruleID,
	).Scan(&gotTitle); err != nil {
		t.Fatalf("read generated task: %v", err)
	}
	if gotTitle != taskTitle {
		t.Fatalf("expected task %q from successful sibling action, got %q", taskTitle, gotTitle)
	}
}

// TestAutomationRuleDeleteNullsTaskSourceRuleID verifies the task-provenance
// invariant from Phase 20 WS1: deleting a rule that previously generated
// tasks leaves those tasks in place but nulls out `source_rule_id`, so the
// audit trail is preserved even when the originating rule is gone.
//
// Rule of thumb: "the task was real work, even if the rule that spawned it
// no longer exists." The FK uses ON DELETE SET NULL, not CASCADE.

func TestAutomationRuleDeleteNullsTaskSourceRuleID(t *testing.T) {
	if testPool == nil || testWorker == nil {
		t.Skip("testPool/testWorker unavailable")
	}
	tok := smokeJWT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var unitID int64
	if err := testPool.QueryRow(ctx, `SELECT id FROM gr33ncore.units LIMIT 1`).Scan(&unitID); err != nil {
		t.Fatalf("find a unit id: %v", err)
	}
	sid := seedRuleSensorWithReading(t, tok, unitID, 5)

	resp := authPost(t, tok, "/farms/1/automation/rules", map[string]any{
		"name":            uniqueName("rule_ws5_cascade"),
		"is_active":       true,
		"trigger_source":  "manual_api_trigger",
		"condition_logic": "ALL",
		"conditions": []map[string]any{
			{"sensor_id": sid, "op": "lt", "value": 10.0},
		},
	})
	expectStatus(t, resp, http.StatusCreated)
	ruleID := int64(decodeMap(t, resp)["id"].(float64))

	taskTitle := uniqueName("ws5_cascade_task")
	resp = authPost(t, tok, fmt.Sprintf("/automation/rules/%d/actions", ruleID), map[string]any{
		"execution_order": 0,
		"action_type":     "create_task",
		"action_parameters": map[string]any{
			"title":     taskTitle,
			"task_type": "inspection",
		},
	})
	expectStatus(t, resp, http.StatusCreated)
	resp.Body.Close()

	testWorker.TickRules(ctx)

	// Capture the task id while the rule is still alive so we can re-check
	// the same row after the delete — we care that the row survives with
	// source_rule_id NULLed, not that it was replaced.
	var taskID int64
	if err := testPool.QueryRow(ctx,
		`SELECT id FROM gr33ncore.tasks WHERE source_rule_id = $1 ORDER BY id DESC LIMIT 1`, ruleID,
	).Scan(&taskID); err != nil {
		t.Fatalf("locate generated task: %v", err)
	}

	resp = authDelete(t, tok, fmt.Sprintf("/automation/rules/%d", ruleID))
	expectStatus(t, resp, http.StatusNoContent)
	resp.Body.Close()

	// The task must still exist, but with source_rule_id cleared.
	var postDeleteSourceRuleID *int64
	var postDeleteTitle string
	if err := testPool.QueryRow(ctx,
		`SELECT title, source_rule_id FROM gr33ncore.tasks WHERE id = $1`, taskID,
	).Scan(&postDeleteTitle, &postDeleteSourceRuleID); err != nil {
		t.Fatalf("re-read task %d after rule delete: %v", taskID, err)
	}
	if postDeleteTitle != taskTitle {
		t.Fatalf("expected task to survive with title %q, got %q", taskTitle, postDeleteTitle)
	}
	if postDeleteSourceRuleID != nil {
		t.Fatalf("expected source_rule_id to be NULL after parent rule delete, got %d", *postDeleteSourceRuleID)
	}

	// Sanity: the executable_actions row is gone (CASCADE), so the rule
	// really was torn down, not just "soft-hidden".
	var actionsLeft int
	if err := testPool.QueryRow(ctx,
		`SELECT COUNT(*) FROM gr33ncore.executable_actions WHERE rule_id = $1`, ruleID,
	).Scan(&actionsLeft); err != nil {
		t.Fatalf("count actions: %v", err)
	}
	if actionsLeft != 0 {
		t.Fatalf("expected rule's actions to be cascaded away, got %d left", actionsLeft)
	}
}
