// Phase 20.95 WS5 — split out of cmd/api/smoke_test.go with zero behaviour
// change. Shared globals (testPool / testServer / testWorker / testNotifier)
// and helpers live in smoke_helpers_test.go; TestMain stays in smoke_test.go.

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestListZones(t *testing.T) {
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/farms/1/zones")
	expectStatus(t, resp, 200)
	items := decodeSlice(t, resp)
	if len(items) == 0 {
		t.Fatal("expected at least one zone from seed data")
	}
}

func TestListSensors(t *testing.T) {
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/farms/1/sensors")
	expectStatus(t, resp, 200)
	_ = decodeSlice(t, resp)
}

func TestSensorReadingsAndStats(t *testing.T) {
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/farms/1/sensors")
	expectStatus(t, resp, 200)
	items := decodeSlice(t, resp)
	if len(items) == 0 {
		t.Skip("no sensors in seed")
	}
	m := items[0].(map[string]any)
	sid := int64(m["id"].(float64))
	resp = authGet(t, tok, fmt.Sprintf("/sensors/%d/readings?limit=10", sid))
	expectStatus(t, resp, 200)
	_ = decodeSlice(t, resp)

	resp = authGet(t, tok, fmt.Sprintf("/sensors/%d/readings/stats", sid))
	expectStatus(t, resp, 200)
	_ = decodeMap(t, resp)
}

// TestSensorAlertDurationAndCooldown verifies the Phase 19 WS2 state machine:
//  1. A reading that breaches the threshold but hasn't sustained for alert_duration_seconds
//     does NOT create an alert.
//  2. Once the streak has been backdated past the duration, the next breaching reading DOES
//     create an alert.
//  3. Further breaching readings within alert_cooldown_seconds are suppressed (no duplicate).
//  4. A reading that returns to bounds clears alert_breach_started_at.

func TestSensorAlertDurationAndCooldown(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	tok := smokeJWT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Pick any unit id (exact unit doesn't matter for the evaluator).
	var unitID int64
	if err := testPool.QueryRow(ctx,
		`SELECT id FROM gr33ncore.units WHERE name = 'celsius' LIMIT 1`,
	).Scan(&unitID); err != nil {
		if err := testPool.QueryRow(ctx, `SELECT id FROM gr33ncore.units LIMIT 1`).Scan(&unitID); err != nil {
			t.Fatalf("find a unit id: %v", err)
		}
	}

	sensorName := uniqueName("alert_gate_sensor")
	resp := authPost(t, tok, "/farms/1/sensors", map[string]any{
		"name":                   sensorName,
		"sensor_type":            "temperature",
		"unit_id":                unitID,
		"alert_threshold_low":    10.0,
		"alert_threshold_high":   40.0,
		"alert_duration_seconds": 60,
		"alert_cooldown_seconds": 3600,
	})
	expectStatus(t, resp, http.StatusCreated)
	s := decodeMap(t, resp)
	sid := int64(s["id"].(float64))

	countAlerts := func() int {
		t.Helper()
		var n int
		err := testPool.QueryRow(ctx, `
			SELECT COUNT(*) FROM gr33ncore.alerts_notifications
			WHERE farm_id = 1
			  AND triggering_event_source_type = 'sensor_reading'
			  AND triggering_event_source_id = $1`, sid).Scan(&n)
		if err != nil {
			t.Fatalf("count alerts: %v", err)
		}
		return n
	}

	postReading := func(value float64) {
		t.Helper()
		b, _ := json.Marshal(map[string]any{
			"value_raw": value,
			"is_valid":  true,
		})
		req, err := http.NewRequest(http.MethodPost,
			testServer.URL+fmt.Sprintf("/sensors/%d/readings", sid), bytes.NewReader(b))
		if err != nil {
			t.Fatalf("build reading request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", piAPIKey)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("post reading: %v", err)
		}
		expectStatus(t, resp, http.StatusCreated)
		resp.Body.Close()
	}

	waitForAlertCount := func(want int) int {
		t.Helper()
		deadline := time.Now().Add(2 * time.Second)
		var got int
		for time.Now().Before(deadline) {
			got = countAlerts()
			if got == want {
				return got
			}
			time.Sleep(40 * time.Millisecond)
		}
		return got
	}

	// Step 1 — breaching reading, but the streak is brand new: duration gate should suppress.
	postReading(5.0)
	if got := waitForAlertCount(0); got != 0 {
		t.Fatalf("expected 0 alerts while within duration window, got %d", got)
	}

	// The evaluator should have stamped alert_breach_started_at on the sensor.
	var breachStart *time.Time
	if err := testPool.QueryRow(ctx,
		`SELECT alert_breach_started_at FROM gr33ncore.sensors WHERE id = $1`, sid,
	).Scan(&breachStart); err != nil {
		t.Fatalf("read breach start: %v", err)
	}
	if breachStart == nil {
		// Evaluator runs in a goroutine; give it a brief moment.
		time.Sleep(200 * time.Millisecond)
		if err := testPool.QueryRow(ctx,
			`SELECT alert_breach_started_at FROM gr33ncore.sensors WHERE id = $1`, sid,
		).Scan(&breachStart); err != nil {
			t.Fatalf("read breach start (retry): %v", err)
		}
		if breachStart == nil {
			t.Fatal("expected alert_breach_started_at to be set after breaching reading")
		}
	}

	// Step 2 — backdate the streak past alert_duration_seconds, then re-post a breach.
	if _, err := testPool.Exec(ctx,
		`UPDATE gr33ncore.sensors SET alert_breach_started_at = NOW() - INTERVAL '10 minutes' WHERE id = $1`,
		sid,
	); err != nil {
		t.Fatalf("backdate breach start: %v", err)
	}
	postReading(4.5)
	if got := waitForAlertCount(1); got != 1 {
		t.Fatalf("expected exactly 1 alert once duration elapsed, got %d", got)
	}

	// Step 3 — another breaching reading within cooldown must NOT produce a second alert.
	postReading(4.0)
	// Give the goroutine a moment and re-check.
	time.Sleep(200 * time.Millisecond)
	if got := countAlerts(); got != 1 {
		t.Fatalf("expected cooldown to suppress duplicate, still 1 alert, got %d", got)
	}

	// Step 4 — a healthy reading should clear alert_breach_started_at.
	postReading(25.0)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		var current *time.Time
		if err := testPool.QueryRow(ctx,
			`SELECT alert_breach_started_at FROM gr33ncore.sensors WHERE id = $1`, sid,
		).Scan(&current); err == nil && current == nil {
			return
		}
		time.Sleep(40 * time.Millisecond)
	}
	t.Fatal("expected alert_breach_started_at to be cleared after in-bounds reading")
}

// TestAlertToTaskLinkage verifies the Phase 19 WS3 flow:
//  1. A breaching reading creates an alert via the evaluator.
//  2. POST /alerts/{id}/create-task synthesises a task, inherits the sensor's zone,
//     derives a sensible priority from severity, and back-links source_alert_id.
//  3. Overrides in the body (title, priority, due_date) win over the derived defaults.

func TestListActuators(t *testing.T) {
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/farms/1/actuators")
	expectStatus(t, resp, 200)
	_ = decodeSlice(t, resp)
}

func TestWorkerHealth(t *testing.T) {
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/automation/worker/health")
	expectStatus(t, resp, 200)
	body := decodeMap(t, resp)
	if body["simulation_mode"] != true {
		t.Fatal("expected simulation_mode=true")
	}
}

// ── Phase 9 CRUD + authz ─────────────────────────────────────────────────────
