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
	"strings"
	"testing"
	"time"
)

func TestAlertToTaskLinkage(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	tok := smokeJWT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Unit (exact type doesn't matter for the evaluator).
	var unitID int64
	if err := testPool.QueryRow(ctx, `SELECT id FROM gr33ncore.units LIMIT 1`).Scan(&unitID); err != nil {
		t.Fatalf("find a unit id: %v", err)
	}

	// Pick any zone on farm 1 so we can verify zone-carry-over.
	var zoneID int64
	if err := testPool.QueryRow(ctx,
		`SELECT id FROM gr33ncore.zones WHERE farm_id = 1 AND deleted_at IS NULL ORDER BY id LIMIT 1`,
	).Scan(&zoneID); err != nil {
		t.Fatalf("find a zone id on farm 1: %v", err)
	}

	// Sensor with duration=0 so the first breaching reading fires immediately.
	sensorName := uniqueName("alert_to_task_sensor")
	resp := authPost(t, tok, "/farms/1/sensors", map[string]any{
		"name":                   sensorName,
		"sensor_type":            "temperature",
		"unit_id":                unitID,
		"zone_id":                zoneID,
		"alert_threshold_low":    10.0,
		"alert_threshold_high":   40.0,
		"alert_duration_seconds": 0,
		"alert_cooldown_seconds": 3600,
	})
	expectStatus(t, resp, http.StatusCreated)
	s := decodeMap(t, resp)
	sid := int64(s["id"].(float64))

	// Post a breaching reading via the Pi API-key path.
	b, _ := json.Marshal(map[string]any{"value_raw": 5.0, "is_valid": true})
	req, err := http.NewRequest(http.MethodPost,
		testServer.URL+fmt.Sprintf("/sensors/%d/readings", sid), bytes.NewReader(b))
	if err != nil {
		t.Fatalf("build reading request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", piAPIKey)
	rresp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("post reading: %v", err)
	}
	expectStatus(t, rresp, http.StatusCreated)
	rresp.Body.Close()

	// Wait for the evaluator goroutine to persist the alert.
	var alertID int64
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if err := testPool.QueryRow(ctx, `
			SELECT id FROM gr33ncore.alerts_notifications
			WHERE farm_id = 1
			  AND triggering_event_source_type = 'sensor_reading'
			  AND triggering_event_source_id = $1
			ORDER BY id DESC LIMIT 1`, sid).Scan(&alertID); err == nil {
			break
		}
		time.Sleep(40 * time.Millisecond)
	}
	if alertID == 0 {
		t.Fatal("expected an alert to be created for the breaching reading")
	}

	// --- Case A: empty body → server-derived title/priority/zone. ---
	resp = authPost(t, tok, fmt.Sprintf("/alerts/%d/create-task", alertID), map[string]any{})
	expectStatus(t, resp, http.StatusCreated)
	task := decodeMap(t, resp)

	if int64(task["farm_id"].(float64)) != 1 {
		t.Fatalf("expected task.farm_id = 1, got %v", task["farm_id"])
	}
	if int64(task["source_alert_id"].(float64)) != alertID {
		t.Fatalf("expected task.source_alert_id = %d, got %v", alertID, task["source_alert_id"])
	}
	if task["zone_id"] == nil {
		t.Fatalf("expected task.zone_id to be derived from the sensor, got nil")
	}
	if int64(task["zone_id"].(float64)) != zoneID {
		t.Fatalf("expected task.zone_id = %d (sensor zone), got %v", zoneID, task["zone_id"])
	}
	if title, _ := task["title"].(string); strings.TrimSpace(title) == "" {
		t.Fatal("expected a non-empty title derived from the alert")
	}
	// Default task_type from alert-create path.
	if tt, _ := task["task_type"].(string); tt != "alert_follow_up" {
		t.Fatalf("expected task_type=alert_follow_up, got %q", tt)
	}

	// --- Case B: overrides win. ---
	resp = authPost(t, tok, fmt.Sprintf("/alerts/%d/create-task", alertID), map[string]any{
		"title":    "custom follow-up",
		"priority": 3,
		"due_date": "2030-01-15",
	})
	expectStatus(t, resp, http.StatusCreated)
	task2 := decodeMap(t, resp)
	if got, _ := task2["title"].(string); got != "custom follow-up" {
		t.Fatalf("expected override title, got %q", got)
	}
	if int64(task2["priority"].(float64)) != 3 {
		t.Fatalf("expected override priority=3, got %v", task2["priority"])
	}
	if int64(task2["source_alert_id"].(float64)) != alertID {
		t.Fatalf("expected task2.source_alert_id = %d, got %v", alertID, task2["source_alert_id"])
	}

	// Both tasks should land in ListTasksByFarm with source_alert_id set.
	resp = authGet(t, tok, "/farms/1/tasks")
	expectStatus(t, resp, http.StatusOK)
	list := decodeSlice(t, resp)
	linked := 0
	for _, row := range list {
		m, ok := row.(map[string]any)
		if !ok {
			continue
		}
		if sa, ok := m["source_alert_id"].(float64); ok && int64(sa) == alertID {
			linked++
		}
	}
	if linked < 2 {
		t.Fatalf("expected at least 2 tasks with source_alert_id=%d in farm list, got %d", alertID, linked)
	}

	// --- Case C: bogus alert id returns 404. ---
	resp = authPost(t, tok, "/alerts/99999999/create-task", map[string]any{})
	expectStatus(t, resp, http.StatusNotFound)
	resp.Body.Close()
}

// TestSchedulePreconditionFailsRun verifies Phase 19 WS4 interlock-lite:
//  1. Creating a schedule with an invalid precondition (bogus sensor id) is rejected.
//  2. When the latest reading for a sensor fails the predicate, the worker's
//     Tick() records an automation_runs row with status='skipped' and
//     message='precondition_failed' and does NOT fire executable actions.
//  3. When the reading satisfies the predicate, the worker proceeds as usual
//     (no interlock skip).

func TestAlertLifecycle(t *testing.T) {
	tok := smokeJWT(t)

	resp := authGet(t, tok, "/farms/1/alerts")
	expectStatus(t, resp, http.StatusOK)
	alerts := decodeSlice(t, resp)

	resp = authGet(t, tok, "/farms/1/alerts/unread-count")
	expectStatus(t, resp, http.StatusOK)
	countMap := decodeMap(t, resp)
	if _, ok := countMap["unread_count"]; !ok {
		t.Fatalf("expected unread_count field in response, got %#v", countMap)
	}

	if len(alerts) == 0 {
		t.Skip("no alerts in seed data to test read/acknowledge")
	}

	first := alerts[0].(map[string]any)
	alertID := int64(first["id"].(float64))

	resp = authPatch(t, tok, fmt.Sprintf("/alerts/%d/read", alertID), map[string]any{})
	expectStatus(t, resp, http.StatusOK)

	resp = authPatch(t, tok, fmt.Sprintf("/alerts/%d/acknowledge", alertID), map[string]any{})
	expectStatus(t, resp, http.StatusOK)
}
