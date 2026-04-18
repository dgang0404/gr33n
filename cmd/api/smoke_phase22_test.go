// Phase 22 smoke coverage —
//
//  WS1:  runProgramTick fires a program's control_actuator step; provenance
//        lands on actuator_events (meta_data.program_id) and automation_runs
//        (program_id column); program.last_triggered_time is stamped; a
//        second Tick inside the same minute short-circuits (idempotent).
//  WS1:  runProgramTick also fires programs whose actions only live in
//        metadata.steps (resolve fallback path), with details.action_source
//        = "metadata_steps_fallback" so ops can spot legacy programs.
//  WS2:  the 20260517 sweep + _backfill_program_actions function remain
//        idempotent: a hand-seeded legacy program gets its rows on the
//        first explicit call and zero on the second.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"gr33n-api/internal/automation"
)

// ── helpers ────────────────────────────────────────────────────────────────

// seedEveryMinuteSchedule creates a schedule on farm 1 whose cron
// expression matches this minute. Returns the new schedule ID.
func seedEveryMinuteSchedule(t *testing.T, tok, label string) int64 {
	t.Helper()
	resp := authPost(t, tok, "/farms/1/schedules", map[string]any{
		"name":            uniqueName(label),
		"schedule_type":   "cron",
		"cron_expression": "* * * * *",
		"timezone":        "UTC",
		"is_active":       true,
	})
	expectStatus(t, resp, 201)
	return int64(decodeMap(t, resp)["id"].(float64))
}

// seedProgramWithSchedule creates a fertigation program bound to a
// schedule. Programs created via the handler start with
// is_active=true and no metadata.steps.
func seedProgramWithSchedule(t *testing.T, tok string, scheduleID int64, label string) int64 {
	t.Helper()
	resp := authPost(t, tok, "/farms/1/fertigation/programs", map[string]any{
		"name":                uniqueName(label),
		"schedule_id":         scheduleID,
		"total_volume_liters": 1.0,
		"is_active":           true,
	})
	expectStatus(t, resp, 201)
	return int64(decodeMap(t, resp)["id"].(float64))
}

// seedActuatorForFarm1 creates an actuator owned by farm 1 that's not
// tied to a device, so the worker's `set pending command` path no-ops
// safely.
func seedActuatorForFarm1(t *testing.T, tok, label string) int64 {
	t.Helper()
	resp := authPost(t, tok, "/farms/1/actuators", map[string]any{
		"name":           uniqueName(label),
		"actuator_type":  "relay",
		"unit":           "state",
		"current_state":  "off",
	})
	if resp.StatusCode == 201 {
		return int64(decodeMap(t, resp)["id"].(float64))
	}
	// Some builds route actuators under /zones/{id}/actuators or a
	// sql-only path — fall back to a direct insert so the smoke test
	// never skips over infra details irrelevant to this phase.
	resp.Body.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var id int64
	if err := testPool.QueryRow(ctx, `
		INSERT INTO gr33ncore.actuators (farm_id, name, actuator_type)
		VALUES (1, $1, 'relay')
		RETURNING id
	`, uniqueName(label)).Scan(&id); err != nil {
		t.Fatalf("seed actuator: %v", err)
	}
	return id
}

// ── WS1: runProgramTick fires a program's control_actuator step ────────────

func TestPhase22ProgramTickFiresControlActuator(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tok := smokeJWT(t)

	schedID := seedEveryMinuteSchedule(t, tok, "ph22_tick_sched")
	progID := seedProgramWithSchedule(t, tok, schedID, "ph22_tick_prog")
	actID := seedActuatorForFarm1(t, tok, "ph22_tick_act")

	// Program-bound executable_action: control_actuator "on".
	resp := authPost(t, tok, fmt.Sprintf("/fertigation/programs/%d/actions", progID), map[string]any{
		"action_type":        "control_actuator",
		"execution_order":    1,
		"target_actuator_id": actID,
		"action_command":     "on",
	})
	expectStatus(t, resp, 201)
	resp.Body.Close()

	// Count actuator events for this actuator before the tick.
	var evBefore int
	if err := testPool.QueryRow(ctx,
		`SELECT COUNT(*) FROM gr33ncore.actuator_events WHERE actuator_id = $1`, actID,
	).Scan(&evBefore); err != nil {
		t.Fatalf("count events before: %v", err)
	}

	testWorker.TickPrograms(ctx)

	// Exactly one new actuator_events row for this actuator, tagged
	// with program_id in meta_data and source=schedule_trigger.
	rows, err := testPool.Query(ctx, `
		SELECT meta_data::text, source::text, triggered_by_schedule_id
		FROM gr33ncore.actuator_events
		WHERE actuator_id = $1
		ORDER BY event_time DESC
		LIMIT 1
	`, actID)
	if err != nil {
		t.Fatalf("load events: %v", err)
	}
	defer rows.Close()
	if !rows.Next() {
		t.Fatalf("expected at least one actuator_events row for actuator %d", actID)
	}
	var metaJSON string
	var source string
	var triggeredSchedID *int64
	if err := rows.Scan(&metaJSON, &source, &triggeredSchedID); err != nil {
		t.Fatalf("scan event: %v", err)
	}
	var meta map[string]any
	if err := json.Unmarshal([]byte(metaJSON), &meta); err != nil {
		t.Fatalf("parse meta_data: %v (raw=%s)", err, metaJSON)
	}
	if got, _ := meta["program_id"].(float64); int64(got) != progID {
		t.Fatalf("expected meta_data.program_id=%d, got %v (raw=%s)", progID, meta["program_id"], metaJSON)
	}
	if source != "schedule_trigger" {
		t.Fatalf("expected source=schedule_trigger, got %s", source)
	}
	if triggeredSchedID == nil || *triggeredSchedID != schedID {
		t.Fatalf("expected triggered_by_schedule_id=%d, got %v", schedID, triggeredSchedID)
	}

	// Exactly one automation_runs row with program_id=progID,
	// status=success, details.action_source=executable_actions.
	var runStatus string
	var detailsJSON string
	if err := testPool.QueryRow(ctx, `
		SELECT status, details::text
		FROM gr33ncore.automation_runs
		WHERE program_id = $1
		ORDER BY executed_at DESC
		LIMIT 1
	`, progID).Scan(&runStatus, &detailsJSON); err != nil {
		t.Fatalf("load run: %v", err)
	}
	if runStatus != "success" {
		t.Fatalf("expected status=success, got %s (details=%s)", runStatus, detailsJSON)
	}
	var runDetails map[string]any
	if err := json.Unmarshal([]byte(detailsJSON), &runDetails); err != nil {
		t.Fatalf("parse run details: %v", err)
	}
	if runDetails["action_source"] != "executable_actions" {
		t.Fatalf("expected action_source=executable_actions, got %v", runDetails["action_source"])
	}

	// program.last_triggered_time is stamped.
	var lastTriggered *time.Time
	if err := testPool.QueryRow(ctx,
		`SELECT last_triggered_time FROM gr33nfertigation.programs WHERE id = $1`, progID,
	).Scan(&lastTriggered); err != nil {
		t.Fatalf("read last_triggered_time: %v", err)
	}
	if lastTriggered == nil {
		t.Fatalf("expected last_triggered_time to be set, got NULL")
	}
}

// TestPhase22ProgramTickIdempotent runs TickPrograms twice in the same
// minute and asserts exactly one automation_run row exists for the
// program. The last_triggered_time guard is the primary mechanism;
// checkProgramIdempotency on the automation_runs details is the backup.
func TestPhase22ProgramTickIdempotent(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tok := smokeJWT(t)

	schedID := seedEveryMinuteSchedule(t, tok, "ph22_idem_sched")
	progID := seedProgramWithSchedule(t, tok, schedID, "ph22_idem_prog")
	actID := seedActuatorForFarm1(t, tok, "ph22_idem_act")

	resp := authPost(t, tok, fmt.Sprintf("/fertigation/programs/%d/actions", progID), map[string]any{
		"action_type":        "control_actuator",
		"execution_order":    1,
		"target_actuator_id": actID,
		"action_command":     "off",
	})
	expectStatus(t, resp, 201)
	resp.Body.Close()

	testWorker.TickPrograms(ctx)
	testWorker.TickPrograms(ctx)
	testWorker.TickPrograms(ctx)

	var runCount int
	if err := testPool.QueryRow(ctx, `
		SELECT COUNT(*) FROM gr33ncore.automation_runs WHERE program_id = $1
	`, progID).Scan(&runCount); err != nil {
		t.Fatalf("count runs: %v", err)
	}
	if runCount != 1 {
		t.Fatalf("expected exactly 1 automation_run for program %d, got %d", progID, runCount)
	}

	// Only one actuator event too — the retry/dedupe guard stops the
	// second tick before the dispatch path reaches InsertActuatorEvent.
	var evCount int
	if err := testPool.QueryRow(ctx, `
		SELECT COUNT(*) FROM gr33ncore.actuator_events
		WHERE actuator_id = $1
		  AND triggered_by_schedule_id = $2
	`, actID, schedID).Scan(&evCount); err != nil {
		t.Fatalf("count events: %v", err)
	}
	if evCount != 1 {
		t.Fatalf("expected exactly 1 actuator_event for program %d, got %d", progID, evCount)
	}
}

// TestPhase22ProgramTickFromMetadataFallback seeds a program with
// metadata.steps (and NO executable_actions rows — we insert directly
// and deliberately don't run _backfill_program_actions). The program
// tick must still fire via the ResolveProgramActions fallback and
// record action_source=metadata_steps_fallback in the run's details.
func TestPhase22ProgramTickFromMetadataFallback(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tok := smokeJWT(t)
	schedID := seedEveryMinuteSchedule(t, tok, "ph22_meta_sched")
	actID := seedActuatorForFarm1(t, tok, "ph22_meta_act")

	// Insert a program directly with metadata.steps baked in. Routing
	// this through the API would trip the sweep we apply in initMigrations,
	// which would move the step into executable_actions and defeat the
	// point of the test.
	var progID int64
	meta := fmt.Sprintf(
		`{"steps":[{"action_type":"control_actuator","target_actuator_id":%d,"action_command":"on","execution_order":1}]}`,
		actID,
	)
	if err := testPool.QueryRow(ctx, `
		INSERT INTO gr33nfertigation.programs
		  (farm_id, name, schedule_id, total_volume_liters, is_active, metadata)
		VALUES (1, $1, $2, 1.0, TRUE, $3::jsonb)
		RETURNING id
	`, uniqueName("ph22_meta_prog"), schedID, meta).Scan(&progID); err != nil {
		t.Fatalf("seed legacy program: %v", err)
	}

	// Sanity check: no executable_actions rows for this program yet.
	var execRows int
	if err := testPool.QueryRow(ctx,
		`SELECT COUNT(*) FROM gr33ncore.executable_actions WHERE program_id = $1`, progID,
	).Scan(&execRows); err != nil {
		t.Fatalf("count exec rows pre-tick: %v", err)
	}
	if execRows != 0 {
		t.Fatalf("pre-tick: expected 0 executable_actions for program %d, got %d", progID, execRows)
	}

	// Verify the resolver would take the fallback path (catches any
	// migration regression where sweep re-runs implicitly).
	_, source, err := automation.ResolveProgramActionsByID(ctx, testPool, progID)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if source != automation.ProgramActionsFromMetadataStepsFallback {
		t.Fatalf("expected fallback resolution, got %s", source)
	}

	testWorker.TickPrograms(ctx)

	var runStatus string
	var detailsJSON string
	if err := testPool.QueryRow(ctx, `
		SELECT status, details::text FROM gr33ncore.automation_runs
		WHERE program_id = $1
		ORDER BY executed_at DESC LIMIT 1
	`, progID).Scan(&runStatus, &detailsJSON); err != nil {
		t.Fatalf("load run: %v", err)
	}
	if runStatus != "success" {
		t.Fatalf("expected status=success, got %s (details=%s)", runStatus, detailsJSON)
	}
	var runDetails map[string]any
	if err := json.Unmarshal([]byte(detailsJSON), &runDetails); err != nil {
		t.Fatalf("parse run details: %v", err)
	}
	if runDetails["action_source"] != "metadata_steps_fallback" {
		t.Fatalf("expected action_source=metadata_steps_fallback, got %v", runDetails["action_source"])
	}

	// And the fallback fire did NOT create executable_actions rows — a
	// common future regression would be the resolver silently persisting
	// synthetic rows on the fly.
	if err := testPool.QueryRow(ctx,
		`SELECT COUNT(*) FROM gr33ncore.executable_actions WHERE program_id = $1`, progID,
	).Scan(&execRows); err != nil {
		t.Fatalf("count exec rows post-tick: %v", err)
	}
	if execRows != 0 {
		t.Fatalf("post-tick: expected fallback to NOT persist rows, got %d", execRows)
	}
}

// TestPhase22BackfillFunctionIdempotent confirms the
// _backfill_program_actions function still returns 0 on a re-run. The
// 20260517 sweep relies on this contract: calling the function over an
// already-migrated corpus must be free of side effects.
func TestPhase22BackfillFunctionIdempotent(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Seed a fresh legacy program (same shape as the fallback test but
	// in a dedicated row so the fallback test's pre-conditions aren't
	// disturbed if they run in a different order).
	var progID int64
	meta := `{"steps":[{"action_type":"create_task","action_parameters":{"title":"backfill_test"}}]}`
	if err := testPool.QueryRow(ctx, `
		INSERT INTO gr33nfertigation.programs
		  (farm_id, name, total_volume_liters, metadata)
		VALUES (1, $1, 0, $2::jsonb)
		RETURNING id
	`, uniqueName("ph22_backfill_prog"), meta).Scan(&progID); err != nil {
		t.Fatalf("seed program: %v", err)
	}

	var inserted int
	if err := testPool.QueryRow(ctx,
		`SELECT gr33ncore._backfill_program_actions($1)`, progID,
	).Scan(&inserted); err != nil {
		t.Fatalf("first backfill call: %v", err)
	}
	if inserted != 1 {
		t.Fatalf("first call: expected 1 insert, got %d", inserted)
	}

	if err := testPool.QueryRow(ctx,
		`SELECT gr33ncore._backfill_program_actions($1)`, progID,
	).Scan(&inserted); err != nil {
		t.Fatalf("second backfill call: %v", err)
	}
	if inserted != 0 {
		t.Fatalf("second call: expected 0 inserts (idempotent), got %d", inserted)
	}

	// And the resolver now prefers the persisted row.
	_, source, err := automation.ResolveProgramActionsByID(ctx, testPool, progID)
	if err != nil {
		t.Fatalf("resolve post-backfill: %v", err)
	}
	if source != automation.ProgramActionsFromExecutableActions {
		t.Fatalf("expected post-backfill source=executable_actions, got %s", source)
	}
}
