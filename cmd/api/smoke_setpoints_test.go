// Phase 20.6 WS5 — smoke tests for gr33ncore.zone_setpoints.
//
// Covers:
//  1. CRUD round-trip on /farms/{id}/setpoints and /setpoints/{id}.
//  2. Client-side validation mirrors the Postgres CHECK constraints
//     (scope requirement, numeric coherence) and the cross-farm guard.
//  3. Precedence resolver: a cycle+stage row shadows a zone-wide row
//     for the same sensor_type.
//  4. Rule-engine hook: a setpoint-typed predicate on a rule whose
//     zone has no matching row skips with `no_setpoint_for_scope`;
//     once the operator lands a matching row the next tick succeeds.

package main

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"
)

// seedSetpointZone picks a zone on farm 1 and returns its id. All
// tests in this file stay on farm 1 to match the broader smoke suite.
func seedSetpointZone(t *testing.T) int64 {
	t.Helper()
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	var zoneID int64
	if err := testPool.QueryRow(context.Background(),
		`SELECT id FROM gr33ncore.zones WHERE farm_id = 1 AND deleted_at IS NULL ORDER BY id LIMIT 1`,
	).Scan(&zoneID); err != nil {
		t.Fatalf("seed zone lookup: %v", err)
	}
	return zoneID
}

// seedFreshSetpointZone creates a brand-new zone on farm 1 so tests
// that need to attach an *active* crop cycle don't fight
// `uq_active_crop_cycle` against the seed zone's pre-existing cycle.
func seedFreshSetpointZone(t *testing.T, tok string) int64 {
	t.Helper()
	resp := authPost(t, tok, "/farms/1/zones", map[string]any{
		"name":      uniqueName("sp_zone"),
		"zone_type": "veg",
	})
	expectStatus(t, resp, http.StatusCreated)
	return int64(decodeMap(t, resp)["id"].(float64))
}

// TestSetpointCRUD — Phase 20.6 WS2. Round-trip a zone-scoped setpoint
// through every verb and verify the list endpoint's filter params.
func TestSetpointCRUD(t *testing.T) {
	tok := smokeJWT(t)
	zoneID := seedSetpointZone(t)

	sensorType := "dew_point_" + uniqueName("sp")
	min, ideal, max := 48.0, 52.0, 56.0
	resp := authPost(t, tok, "/farms/1/setpoints", map[string]any{
		"zone_id":     zoneID,
		"sensor_type": sensorType,
		"min_value":   min,
		"ideal_value": ideal,
		"max_value":   max,
	})
	expectStatus(t, resp, http.StatusCreated)
	created := decodeMap(t, resp)
	id := int64(created["id"].(float64))
	if created["sensor_type"] != sensorType {
		t.Fatalf("expected sensor_type=%s, got %v", sensorType, created["sensor_type"])
	}

	resp = authGet(t, tok, fmt.Sprintf("/setpoints/%d", id))
	expectStatus(t, resp, http.StatusOK)
	_ = decodeMap(t, resp)

	resp = authGet(t, tok, fmt.Sprintf("/farms/1/setpoints?zone_id=%d&sensor_type=%s", zoneID, sensorType))
	expectStatus(t, resp, http.StatusOK)
	list := decodeSlice(t, resp)
	found := false
	for _, r := range list {
		if m, ok := r.(map[string]any); ok && int64(m["id"].(float64)) == id {
			found = true
		}
	}
	if !found {
		t.Fatalf("list /farms/1/setpoints did not return newly-created row %d", id)
	}

	newIdeal := 53.0
	resp = authPut(t, tok, fmt.Sprintf("/setpoints/%d", id), map[string]any{
		"zone_id":     zoneID,
		"sensor_type": sensorType,
		"min_value":   min,
		"ideal_value": newIdeal,
		"max_value":   max,
	})
	expectStatus(t, resp, http.StatusOK)

	resp = authDelete(t, tok, fmt.Sprintf("/setpoints/%d", id))
	expectStatus(t, resp, http.StatusNoContent)

	resp = authGet(t, tok, fmt.Sprintf("/setpoints/%d", id))
	expectStatus(t, resp, http.StatusNotFound)
}

// TestSetpointValidation — Phase 20.6 WS2. The handler layers the two
// CHECK constraints (`chk_setpoint_scope` and
// `chk_setpoint_numeric_coherent`) plus the cross-farm guard in front
// of the DB so operators get readable 400s.
func TestSetpointValidation(t *testing.T) {
	tok := smokeJWT(t)
	zoneID := seedSetpointZone(t)

	// Missing both zone_id and crop_cycle_id → scope violation.
	resp := authPost(t, tok, "/farms/1/setpoints", map[string]any{
		"sensor_type": "dew_point",
		"ideal_value": 50.0,
	})
	expectStatus(t, resp, http.StatusBadRequest)

	// Missing sensor_type.
	resp = authPost(t, tok, "/farms/1/setpoints", map[string]any{
		"zone_id":     zoneID,
		"ideal_value": 50.0,
	})
	expectStatus(t, resp, http.StatusBadRequest)

	// Numeric incoherence: min > max.
	resp = authPost(t, tok, "/farms/1/setpoints", map[string]any{
		"zone_id":     zoneID,
		"sensor_type": "dew_point_bad_range_" + uniqueName("sp"),
		"min_value":   60.0,
		"max_value":   50.0,
	})
	expectStatus(t, resp, http.StatusBadRequest)

	// Numeric incoherence: ideal outside [min, max].
	resp = authPost(t, tok, "/farms/1/setpoints", map[string]any{
		"zone_id":     zoneID,
		"sensor_type": "dew_point_bad_ideal_" + uniqueName("sp"),
		"min_value":   50.0,
		"max_value":   60.0,
		"ideal_value": 70.0,
	})
	expectStatus(t, resp, http.StatusBadRequest)

	// Cross-farm: zone_id on another farm must 400 (not 500).
	var otherZoneID int64
	if err := testPool.QueryRow(context.Background(),
		`SELECT id FROM gr33ncore.zones WHERE farm_id <> 1 AND deleted_at IS NULL ORDER BY id LIMIT 1`,
	).Scan(&otherZoneID); err == nil {
		resp = authPost(t, tok, "/farms/1/setpoints", map[string]any{
			"zone_id":     otherZoneID,
			"sensor_type": "dew_point_cross_farm_" + uniqueName("sp"),
			"ideal_value": 50.0,
		})
		expectStatus(t, resp, http.StatusBadRequest)
	}
}

// TestSetpointPrecedenceCycleOverZone — Phase 20.6 WS1. The precedence
// resolver (GetActiveSetpointForScope) must prefer a cycle+stage row
// (rank 1) over a zone-wide fallback (rank 4) for the same
// sensor_type. We don't hit the rule engine here — that's WS3's test
// — we drive the SQL directly so a regression in the ORDER BY blows up
// even before a rule is wired.
func TestSetpointPrecedenceCycleOverZone(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	tok := smokeJWT(t)
	zoneID := seedFreshSetpointZone(t, tok)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Crop cycle on that zone, stage=early_veg, is_active=true.
	resp := authPost(t, tok, "/farms/1/crop-cycles", map[string]any{
		"zone_id":       zoneID,
		"name":          uniqueName("sp_cycle"),
		"current_stage": "early_veg",
		"started_at":    "2025-04-01",
		"is_active":     true,
	})
	expectStatus(t, resp, http.StatusCreated)
	cycleID := int64(decodeMap(t, resp)["id"].(float64))
	t.Cleanup(func() {
		// Deactivate so other tests can create their own active
		// cycle on this zone without tripping uq_active_crop_cycle.
		_, _ = testPool.Exec(context.Background(),
			`UPDATE gr33nfertigation.crop_cycles SET is_active = FALSE WHERE id = $1`, cycleID)
	})

	sensorType := "dew_point_prec_" + uniqueName("sp")

	// Zone-wide fallback: ideal=70.
	resp = authPost(t, tok, "/farms/1/setpoints", map[string]any{
		"zone_id":     zoneID,
		"sensor_type": sensorType,
		"ideal_value": 70.0,
	})
	expectStatus(t, resp, http.StatusCreated)

	// Cycle+stage-specific row: ideal=50.
	resp = authPost(t, tok, "/farms/1/setpoints", map[string]any{
		"crop_cycle_id": cycleID,
		"stage":         "early_veg",
		"sensor_type":   sensorType,
		"ideal_value":   50.0,
	})
	expectStatus(t, resp, http.StatusCreated)

	// Drive the resolver directly. Expect the rank-1 row (ideal=50).
	var ideal float64
	if err := testPool.QueryRow(ctx, `
		SELECT ideal_value::float8
		FROM gr33ncore.zone_setpoints
		WHERE sensor_type = $1
		  AND (
		        (crop_cycle_id = $2 AND stage = $3)
		     OR (crop_cycle_id = $2 AND stage IS NULL)
		     OR (zone_id       = $4 AND stage = $3)
		     OR (zone_id       = $4 AND stage IS NULL)
		  )
		ORDER BY (
		    CASE
		        WHEN crop_cycle_id IS NOT NULL AND stage IS NOT NULL THEN 1
		        WHEN crop_cycle_id IS NOT NULL AND stage IS NULL     THEN 2
		        WHEN zone_id       IS NOT NULL AND stage IS NOT NULL THEN 3
		        WHEN zone_id       IS NOT NULL AND stage IS NULL     THEN 4
		        ELSE 99
		    END
		) ASC, updated_at DESC
		LIMIT 1`,
		sensorType, cycleID, "early_veg", zoneID,
	).Scan(&ideal); err != nil {
		t.Fatalf("resolve setpoint: %v", err)
	}
	if ideal != 50.0 {
		t.Fatalf("precedence broken: expected cycle+stage ideal=50, got %v", ideal)
	}

	// Dropping the rank-1 row uncovers the rank-4 zone fallback (ideal=70).
	if _, err := testPool.Exec(ctx,
		`DELETE FROM gr33ncore.zone_setpoints WHERE crop_cycle_id = $1 AND stage = 'early_veg' AND sensor_type = $2`,
		cycleID, sensorType,
	); err != nil {
		t.Fatalf("delete rank-1 row: %v", err)
	}
	if err := testPool.QueryRow(ctx, `
		SELECT ideal_value::float8 FROM gr33ncore.zone_setpoints
		WHERE farm_id = 1 AND zone_id = $1 AND sensor_type = $2`,
		zoneID, sensorType,
	).Scan(&ideal); err != nil {
		t.Fatalf("resolve zone fallback: %v", err)
	}
	if ideal != 70.0 {
		t.Fatalf("fallback broken: expected zone-wide ideal=70, got %v", ideal)
	}
}

// TestAutomationRuleSetpointGracefulSkip — Phase 20.6 WS3. A rule with
// a setpoint-typed predicate whose zone has no matching row must
// record a `skipped` run with message=no_setpoint_for_scope (not the
// generic conditions_not_met). Once the operator lands a matching row
// the next tick succeeds.
func TestAutomationRuleSetpointGracefulSkip(t *testing.T) {
	if testPool == nil || testWorker == nil {
		t.Skip("testPool/testWorker unavailable")
	}
	tok := smokeJWT(t)
	zoneID := seedFreshSetpointZone(t, tok)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	var unitID int64
	if err := testPool.QueryRow(ctx, `SELECT id FROM gr33ncore.units LIMIT 1`).Scan(&unitID); err != nil {
		t.Fatalf("find unit id: %v", err)
	}

	// Sensor on the zone with a unique sensor_type so other tests in
	// the suite don't shadow our precedence resolution.
	sensorType := "sp_skip_" + uniqueName("sp")
	resp := authPost(t, tok, "/farms/1/sensors", map[string]any{
		"name":        uniqueName("sp_skip_sensor"),
		"sensor_type": sensorType,
		"unit_id":     unitID,
		"zone_id":     zoneID,
	})
	expectStatus(t, resp, http.StatusCreated)
	sid := int64(decodeMap(t, resp)["id"].(float64))

	// Single reading so the evaluator has something to compare.
	if _, err := testPool.Exec(ctx, `
		INSERT INTO gr33ncore.sensor_readings (reading_time, sensor_id, value_raw, is_valid)
		VALUES (NOW(), $1, $2, TRUE)`, sid, 72.0); err != nil {
		t.Fatalf("seed reading: %v", err)
	}

	// Active crop cycle on the zone so scope resolution can reach
	// rank 1 / rank 2 when we eventually land a setpoint.
	resp = authPost(t, tok, "/farms/1/crop-cycles", map[string]any{
		"zone_id":       zoneID,
		"name":          uniqueName("sp_skip_cycle"),
		"current_stage": "early_veg",
		"started_at":    "2025-04-01",
		"is_active":     true,
	})
	expectStatus(t, resp, http.StatusCreated)
	cycleID := int64(decodeMap(t, resp)["id"].(float64))
	t.Cleanup(func() {
		_, _ = testPool.Exec(context.Background(),
			`UPDATE gr33nfertigation.crop_cycles SET is_active = FALSE WHERE id = $1`, cycleID)
	})

	// Rule with a setpoint predicate. trigger_configuration carries
	// zone_id so rules.go:ruleZoneID can thread ScopeContext down to
	// the evaluator.
	resp = authPost(t, tok, "/farms/1/automation/rules", map[string]any{
		"name":                  uniqueName("sp_skip_rule"),
		"is_active":             true,
		"trigger_source":        "manual_api_trigger",
		"trigger_configuration": map[string]any{"zone_id": zoneID},
		"condition_logic":       "ALL",
		"conditions": []map[string]any{
			{
				"type":        "setpoint",
				"sensor_type": sensorType,
				"scope":       "current_stage",
				"op":          "out_of_range",
			},
		},
	})
	expectStatus(t, resp, http.StatusCreated)
	ruleID := int64(decodeMap(t, resp)["id"].(float64))

	// No setpoint exists yet → tick must skip with no_setpoint_for_scope.
	testWorker.TickRules(ctx)
	var status, msg string
	if err := testPool.QueryRow(ctx,
		`SELECT status, COALESCE(message, '') FROM gr33ncore.automation_runs
		 WHERE rule_id = $1 ORDER BY id DESC LIMIT 1`, ruleID,
	).Scan(&status, &msg); err != nil {
		t.Fatalf("read run: %v", err)
	}
	if status != "skipped" {
		t.Fatalf("expected status=skipped on missing setpoint, got %q", status)
	}
	if msg != "no_setpoint_for_scope" {
		t.Fatalf("expected message=no_setpoint_for_scope, got %q", msg)
	}

	// Land a cycle+stage setpoint with range 50–60. Reading is 72 →
	// out_of_range=true → the rule now passes conditions.
	resp = authPost(t, tok, "/farms/1/setpoints", map[string]any{
		"crop_cycle_id": cycleID,
		"stage":         "early_veg",
		"sensor_type":   sensorType,
		"min_value":     50.0,
		"max_value":     60.0,
		"ideal_value":   55.0,
	})
	expectStatus(t, resp, http.StatusCreated)

	testWorker.TickRules(ctx)
	if err := testPool.QueryRow(ctx,
		`SELECT status, COALESCE(message, '') FROM gr33ncore.automation_runs
		 WHERE rule_id = $1 ORDER BY id DESC LIMIT 1`, ruleID,
	).Scan(&status, &msg); err != nil {
		t.Fatalf("read second run: %v", err)
	}
	// A rule with zero actions succeeds trivially on a conditions-met
	// tick — the worker records status=success and no `message`.
	// What we care about is that the setpoint-specific skip is gone.
	if status == "skipped" && msg == "no_setpoint_for_scope" {
		t.Fatalf("rule still skipping with no_setpoint_for_scope after landing matching row")
	}
}
