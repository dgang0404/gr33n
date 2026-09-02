// Phase 216 WS8 — dosing session state machine smoke tests.
//
// Exercises the WS2 worker (dose -> circulate -> settle -> sample -> evaluate)
// against a real DB via testWorker (constructed with simulation=true in
// smoke_test.go, so actuator pulses flip state directly instead of requiring
// a real Pi). Rather than waiting out circulate_seconds/settle_seconds in
// real time, each test forces next_action_at into the past before calling
// TickDosingSessions again — same effect as an injectable clock, without
// having to thread one through the worker.
package main

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// forceDosingSessionDue makes a session immediately eligible for the next
// tick, standing in for "time has passed" without a real sleep.
func forceDosingSessionDue(t *testing.T, ctx context.Context, sessionID int64) {
	t.Helper()
	if _, err := testPool.Exec(ctx, `
		UPDATE gr33nfertigation.dosing_sessions
		SET next_action_at = NOW() - INTERVAL '1 second'
		WHERE id = $1
	`, sessionID); err != nil {
		t.Fatalf("force dosing session due: %v", err)
	}
}

func dosingSessionStatus(t *testing.T, ctx context.Context, sessionID int64) string {
	t.Helper()
	var status string
	if err := testPool.QueryRow(ctx, `
		SELECT status FROM gr33nfertigation.dosing_sessions WHERE id = $1
	`, sessionID).Scan(&status); err != nil {
		t.Fatalf("read dosing session status: %v", err)
	}
	return status
}

// setupDosingFixtures creates a reservoir (with delivery + circulation pumps
// and an EC sensor), an ec_target, and a mixing_event bound to that target.
// Returns (reservoirID, mixingEventID, ecSensorID); cleans up via t.Cleanup.
func setupDosingFixtures(t *testing.T, ctx context.Context, ecMin, ecMax float64) (int64, int64, int64) {
	t.Helper()
	uid := fmt.Sprintf("ph216_%d", time.Now().UnixNano())

	var unitID int64
	if err := testPool.QueryRow(ctx, `SELECT id FROM gr33ncore.units LIMIT 1`).Scan(&unitID); err != nil {
		t.Fatalf("load a unit for the ec sensor: %v", err)
	}

	var deliveryActuatorID, circActuatorID, ecSensorID, reservoirID, ecTargetID, mixingEventID int64
	if err := testPool.QueryRow(ctx, `
		INSERT INTO gr33ncore.actuators (farm_id, zone_id, name, actuator_type, config, meta_data)
		VALUES (1, 1, $1, 'pump', '{"ml_per_second": 1.0}'::jsonb, '{}'::jsonb)
		RETURNING id
	`, uid+"_delivery").Scan(&deliveryActuatorID); err != nil {
		t.Fatalf("insert delivery actuator: %v", err)
	}
	if err := testPool.QueryRow(ctx, `
		INSERT INTO gr33ncore.actuators (farm_id, zone_id, name, actuator_type, config, meta_data)
		VALUES (1, 1, $1, 'circulation_pump', '{}'::jsonb, '{}'::jsonb)
		RETURNING id
	`, uid+"_circ").Scan(&circActuatorID); err != nil {
		t.Fatalf("insert circulation actuator: %v", err)
	}
	if err := testPool.QueryRow(ctx, `
		INSERT INTO gr33ncore.sensors (farm_id, zone_id, name, sensor_type, unit_id)
		VALUES (1, 1, $1, 'ec', $2)
		RETURNING id
	`, uid+"_ec", unitID).Scan(&ecSensorID); err != nil {
		t.Fatalf("insert ec sensor: %v", err)
	}
	if err := testPool.QueryRow(ctx, `
		INSERT INTO gr33nfertigation.reservoirs (
			farm_id, name, capacity_liters, current_volume_liters, status,
			ec_sensor_id, delivery_actuator_id, circulation_actuator_id
		) VALUES (1, $1, 100, 50, 'ready', $2, $3, $4)
		RETURNING id
	`, uid+"_res", ecSensorID, deliveryActuatorID, circActuatorID).Scan(&reservoirID); err != nil {
		t.Fatalf("insert reservoir: %v", err)
	}
	if err := testPool.QueryRow(ctx, `
		INSERT INTO gr33nfertigation.ec_targets (farm_id, zone_id, growth_stage, ec_min_mscm, ec_max_mscm)
		VALUES (1, NULL, 'seedling', $1, $2)
		RETURNING id
	`, ecMin, ecMax).Scan(&ecTargetID); err != nil {
		t.Fatalf("insert ec target: %v", err)
	}
	if err := testPool.QueryRow(ctx, `
		INSERT INTO gr33nfertigation.mixing_events (farm_id, reservoir_id, water_volume_liters, ec_target_id)
		VALUES (1, $1, 50, $2)
		RETURNING id
	`, reservoirID, ecTargetID).Scan(&mixingEventID); err != nil {
		t.Fatalf("insert mixing event: %v", err)
	}

	t.Cleanup(func() {
		bg := context.Background()
		_, _ = testPool.Exec(bg, `DELETE FROM gr33nfertigation.dosing_sessions WHERE mixing_event_id = $1`, mixingEventID)
		_, _ = testPool.Exec(bg, `DELETE FROM gr33nfertigation.mixing_events WHERE id = $1`, mixingEventID)
		_, _ = testPool.Exec(bg, `DELETE FROM gr33nfertigation.ec_targets WHERE id = $1`, ecTargetID)
		_, _ = testPool.Exec(bg, `DELETE FROM gr33nfertigation.reservoirs WHERE id = $1`, reservoirID)
		_, _ = testPool.Exec(bg, `DELETE FROM gr33ncore.sensor_readings WHERE sensor_id = $1`, ecSensorID)
		_, _ = testPool.Exec(bg, `DELETE FROM gr33ncore.sensors WHERE id = $1`, ecSensorID)
		_, _ = testPool.Exec(bg, `DELETE FROM gr33ncore.actuators WHERE id IN ($1, $2)`, deliveryActuatorID, circActuatorID)
	})

	return reservoirID, mixingEventID, ecSensorID
}

func insertEcReading(t *testing.T, ctx context.Context, sensorID int64, value float64) {
	t.Helper()
	if _, err := testPool.Exec(ctx, `
		INSERT INTO gr33ncore.sensor_readings (reading_time, sensor_id, value_raw, is_valid)
		VALUES (NOW(), $1, $2, TRUE)
	`, sensorID, value); err != nil {
		t.Fatalf("insert ec reading: %v", err)
	}
}

func createDosingSession(t *testing.T, ctx context.Context, reservoirID, mixingEventID, circActuatorID int64, doseIncrementMl, maxTotalDoseMl float64) int64 {
	t.Helper()
	var sessionID int64
	if err := testPool.QueryRow(ctx, `
		INSERT INTO gr33nfertigation.dosing_sessions (
			mixing_event_id, farm_id, reservoir_id, circulation_actuator_id,
			dose_increment_ml, max_total_dose_ml, circulate_seconds, settle_seconds,
			samples_per_cycle, max_no_improvement_cycles, next_action_at
		) VALUES ($1, 1, $2, $3, $4, $5, 1, 1, 2, 2, NOW())
		RETURNING id
	`, mixingEventID, reservoirID, circActuatorID, doseIncrementMl, maxTotalDoseMl).Scan(&sessionID); err != nil {
		t.Fatalf("insert dosing session: %v", err)
	}
	return sessionID
}

// TestPhase216DosingSessionReachesTarget drives one full cycle — idle (dose)
// -> dosing -> circulating -> settling -> sampling x2 -> evaluating — and
// asserts it lands on 'done' once the averaged EC clears ec_min_mscm.
func TestPhase216DosingSessionReachesTarget(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, mixingEventID, ecSensorID := setupDosingFixtures(t, ctx, 2.0, 3.0)
	var reservoirID, circActuatorID int64
	if err := testPool.QueryRow(ctx, `SELECT reservoir_id FROM gr33nfertigation.mixing_events WHERE id = $1`, mixingEventID).Scan(&reservoirID); err != nil {
		t.Fatalf("load reservoir_id: %v", err)
	}
	if err := testPool.QueryRow(ctx, `SELECT circulation_actuator_id FROM gr33nfertigation.reservoirs WHERE id = $1`, reservoirID).Scan(&circActuatorID); err != nil {
		t.Fatalf("load circulation_actuator_id: %v", err)
	}
	sessionID := createDosingSession(t, ctx, reservoirID, mixingEventID, circActuatorID, 0.5, 5.0)
	insertEcReading(t, ctx, ecSensorID, 2.5) // already within [2.0, 3.0] once evaluated

	// idle -> dosing
	testWorker.TickDosingSessions(ctx)
	if got := dosingSessionStatus(t, ctx, sessionID); got != "dosing" {
		t.Fatalf("after idle tick: status = %q, want dosing", got)
	}

	// dosing -> circulating
	forceDosingSessionDue(t, ctx, sessionID)
	testWorker.TickDosingSessions(ctx)
	if got := dosingSessionStatus(t, ctx, sessionID); got != "circulating" {
		t.Fatalf("after dosing tick: status = %q, want circulating", got)
	}

	// circulating -> settling
	forceDosingSessionDue(t, ctx, sessionID)
	testWorker.TickDosingSessions(ctx)
	if got := dosingSessionStatus(t, ctx, sessionID); got != "settling" {
		t.Fatalf("after circulating tick: status = %q, want settling", got)
	}

	// settling -> sampling
	forceDosingSessionDue(t, ctx, sessionID)
	testWorker.TickDosingSessions(ctx)
	if got := dosingSessionStatus(t, ctx, sessionID); got != "sampling" {
		t.Fatalf("after settling tick: status = %q, want sampling", got)
	}

	// sampling (1 of 2) -> still sampling
	forceDosingSessionDue(t, ctx, sessionID)
	testWorker.TickDosingSessions(ctx)
	if got := dosingSessionStatus(t, ctx, sessionID); got != "sampling" {
		t.Fatalf("after first sample: status = %q, want sampling", got)
	}

	// sampling (2 of 2) -> evaluating
	forceDosingSessionDue(t, ctx, sessionID)
	testWorker.TickDosingSessions(ctx)
	if got := dosingSessionStatus(t, ctx, sessionID); got != "evaluating" {
		t.Fatalf("after second sample: status = %q, want evaluating", got)
	}

	// evaluating -> done (avg 2.5 is within [2.0, 3.0])
	forceDosingSessionDue(t, ctx, sessionID)
	testWorker.TickDosingSessions(ctx)
	var status string
	var ecTargetMet *bool
	if err := testPool.QueryRow(ctx, `SELECT status FROM gr33nfertigation.dosing_sessions WHERE id = $1`, sessionID).Scan(&status); err != nil {
		t.Fatalf("read final status: %v", err)
	}
	if status != "done" {
		t.Fatalf("final status = %q, want done", status)
	}
	if err := testPool.QueryRow(ctx, `SELECT ec_target_met FROM gr33nfertigation.mixing_events WHERE id = $1`, mixingEventID).Scan(&ecTargetMet); err != nil {
		t.Fatalf("read mixing_event.ec_target_met: %v", err)
	}
	if ecTargetMet == nil || !*ecTargetMet {
		t.Fatalf("mixing_event.ec_target_met = %v, want true", ecTargetMet)
	}
}

// TestPhase216DosingSessionAbortsWhenAlreadyAboveTarget covers the WS5 safety
// guard: dosing can only raise EC, so a reading already above ec_max_mscm
// must abort rather than keep firing the pump.
func TestPhase216DosingSessionAbortsWhenAlreadyAboveTarget(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, mixingEventID, ecSensorID := setupDosingFixtures(t, ctx, 1.0, 2.0)
	var reservoirID, circActuatorID int64
	if err := testPool.QueryRow(ctx, `SELECT reservoir_id FROM gr33nfertigation.mixing_events WHERE id = $1`, mixingEventID).Scan(&reservoirID); err != nil {
		t.Fatalf("load reservoir_id: %v", err)
	}
	if err := testPool.QueryRow(ctx, `SELECT circulation_actuator_id FROM gr33nfertigation.reservoirs WHERE id = $1`, reservoirID).Scan(&circActuatorID); err != nil {
		t.Fatalf("load circulation_actuator_id: %v", err)
	}
	sessionID := createDosingSession(t, ctx, reservoirID, mixingEventID, circActuatorID, 0.5, 5.0)
	insertEcReading(t, ctx, ecSensorID, 4.0) // already above ec_max_mscm=2.0

	// Drive idle -> dosing -> circulating -> settling -> sampling x2 -> evaluating (6 ticks),
	// then one more tick to actually run the evaluate-and-decide step.
	for i := 0; i < 6; i++ {
		forceDosingSessionDue(t, ctx, sessionID)
		testWorker.TickDosingSessions(ctx)
	}
	forceDosingSessionDue(t, ctx, sessionID)
	testWorker.TickDosingSessions(ctx)

	var status string
	var abortedReason *string
	if err := testPool.QueryRow(ctx, `
		SELECT status, aborted_reason FROM gr33nfertigation.dosing_sessions WHERE id = $1
	`, sessionID).Scan(&status, &abortedReason); err != nil {
		t.Fatalf("read session: %v", err)
	}
	if status != "aborted" {
		t.Fatalf("status = %q, want aborted", status)
	}
	if abortedReason == nil || *abortedReason == "" {
		t.Fatalf("expected a non-empty aborted_reason")
	}
}
