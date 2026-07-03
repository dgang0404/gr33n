// Phase 114 — Pi/edge integrity smokes
package main

import (
	"context"
	"fmt"
	"net/http"
	"testing"
)

func TestPhase114_StaleDeviceMarkedOffline(t *testing.T) {
	if testPool == nil || testWorker == nil {
		t.Skip("testPool/testWorker unavailable")
	}
	ctx := context.Background()
	var devID int64
	if err := testPool.QueryRow(ctx, `
		SELECT id FROM gr33ncore.devices
		WHERE farm_id = 1 AND deleted_at IS NULL
		ORDER BY id LIMIT 1`).Scan(&devID); err != nil {
		t.Fatalf("load device: %v", err)
	}
	_, err := testPool.Exec(ctx, `
		UPDATE gr33ncore.devices
		SET status = 'online', last_heartbeat = NOW() - INTERVAL '2 hours'
		WHERE id = $1`, devID)
	if err != nil {
		t.Fatalf("seed stale heartbeat: %v", err)
	}
	defer func() {
		_, _ = testPool.Exec(ctx, `
			UPDATE gr33ncore.devices SET status = 'online', last_heartbeat = NOW() WHERE id = $1`, devID)
	}()

	testWorker.TickDeviceHealth(ctx)

	var status string
	if err := testPool.QueryRow(ctx, `SELECT status FROM gr33ncore.devices WHERE id = $1`, devID).Scan(&status); err != nil {
		t.Fatalf("read status: %v", err)
	}
	if status != "offline" {
		t.Fatalf("expected offline, got %q", status)
	}

	var alertCount int
	if err := testPool.QueryRow(ctx, `
		SELECT COUNT(*) FROM gr33ncore.alerts_notifications
		WHERE farm_id = 1 AND triggering_event_source_type = 'device_offline'
		  AND triggering_event_source_id = $1
		  AND created_at > NOW() - INTERVAL '5 minutes'`, devID).Scan(&alertCount); err != nil {
		t.Fatalf("count alerts: %v", err)
	}
	if alertCount < 1 {
		t.Fatal("expected device_offline alert")
	}
}

func TestPhase114_DeviceTelemetryPatch(t *testing.T) {
	var devID int64
	if err := testPool.QueryRow(context.Background(), `
		SELECT id FROM gr33ncore.devices WHERE farm_id = 1 AND deleted_at IS NULL ORDER BY id LIMIT 1`).Scan(&devID); err != nil {
		t.Fatalf("load device: %v", err)
	}
	resp := piPatchJSON(t, fmt.Sprintf("/devices/%d/status", devID), map[string]any{
		"status":           "online",
		"client_version":   "phase114-smoke",
		"firmware_version": "phase114-smoke",
		"uptime_seconds":   42,
	})
	expectStatus(t, resp, http.StatusOK)
	body := decodeMap(t, resp)
	cfg, ok := body["config"].(map[string]any)
	if !ok {
		t.Fatalf("config missing: %#v", body["config"])
	}
	if cfg["client_version"] != "phase114-smoke" {
		t.Fatalf("client_version: %#v", cfg["client_version"])
	}
	if cfg["client_uptime_seconds"].(float64) != 42 {
		t.Fatalf("uptime: %#v", cfg["client_uptime_seconds"])
	}
}

func TestPhase114_CancelPendingCommand(t *testing.T) {
	tok := smokeJWT(t)
	var devID, actID int64
	ctx := context.Background()
	if err := testPool.QueryRow(ctx, `
		SELECT d.id, a.id FROM gr33ncore.devices d
		JOIN gr33ncore.actuators a ON a.device_id = d.id AND a.deleted_at IS NULL
		WHERE d.farm_id = 1 AND d.deleted_at IS NULL
		ORDER BY d.id LIMIT 1`).Scan(&devID, &actID); err != nil {
		t.Fatalf("load device/actuator: %v", err)
	}

	enq := authPost(t, tok, fmt.Sprintf("/devices/%d/commands", devID), map[string]any{
		"command_type": "actuator",
		"actuator_id":  actID,
		"command":      "on",
	})
	expectStatus(t, enq, http.StatusAccepted)
	cmd := decodeMap(t, enq)
	cmdID := int64(cmd["id"].(float64))

	cancel := authPost(t, tok, fmt.Sprintf("/devices/%d/commands/%d/cancel", devID, cmdID), map[string]any{})
	expectStatus(t, cancel, http.StatusOK)
	cancelled := decodeMap(t, cancel)
	if cancelled["status"] != "cancelled" {
		t.Fatalf("expected cancelled, got %#v", cancelled["status"])
	}
}

func TestPhase114_PiMixingEventWithDeviceKey(t *testing.T) {
	var reservoirID int64
	if err := testPool.QueryRow(context.Background(), `
		SELECT id FROM gr33nfertigation.reservoirs
		WHERE farm_id = 1 ORDER BY id LIMIT 1`).Scan(&reservoirID); err != nil {
		t.Fatalf("load reservoir: %v", err)
	}
	resp := piPostJSON(t, "/farms/1/fertigation/mixing-events", map[string]any{
		"reservoir_id":        reservoirID,
		"water_volume_liters": 10.0,
		"water_source":        "automated",
		"water_ec_mscm":       0.2,
	})
	expectStatus(t, resp, http.StatusCreated)
	body := decodeMap(t, resp)
	ev, _ := body["event"].(map[string]any)
	if ev == nil || ev["id"] == nil {
		t.Fatalf("expected mixing event id: %#v", body)
	}
}

func TestPhase114_SensorCalibration(t *testing.T) {
	tok := smokeJWT(t)
	var sensorID int64
	if err := testPool.QueryRow(context.Background(), `
		SELECT id FROM gr33ncore.sensors
		WHERE farm_id = 1 AND sensor_type IN ('ec','ph') AND deleted_at IS NULL
		ORDER BY id LIMIT 1`).Scan(&sensorID); err != nil {
		t.Skip("no ec/ph sensor in seed")
	}
	resp := authPatch(t, tok, fmt.Sprintf("/sensors/%d/calibration", sensorID), map[string]any{
		"point_a": map[string]any{"raw": 2.5, "reference": 7.0},
		"point_b": map[string]any{"raw": 2.8, "reference": 4.0},
	})
	expectStatus(t, resp, http.StatusOK)
	body := decodeMap(t, resp)
	if body["is_calibrated"] != true {
		t.Fatalf("expected is_calibrated true: %#v", body["is_calibrated"])
	}
	if body["calibration_data"] == nil {
		t.Fatal("expected calibration_data")
	}
}

func TestPhase114_ActuatorEventResultingState(t *testing.T) {
	var actID int64
	if err := testPool.QueryRow(context.Background(), `
		SELECT id FROM gr33ncore.actuators WHERE farm_id = 1 AND deleted_at IS NULL ORDER BY id LIMIT 1`).Scan(&actID); err != nil {
		t.Fatalf("load actuator: %v", err)
	}
	resp := piPostJSON(t, fmt.Sprintf("/actuators/%d/events", actID), map[string]any{
		"command_sent":            "on",
		"source":                  "manual_api_call",
		"execution_status":        "execution_completed_success_on_device",
		"resulting_state_text":    "on",
		"resulting_state_numeric": 1.0,
	})
	expectStatus(t, resp, http.StatusCreated)

	var text *string
	if err := testPool.QueryRow(context.Background(), `
		SELECT resulting_state_text_actual FROM gr33ncore.actuator_events
		WHERE actuator_id = $1 ORDER BY event_time DESC LIMIT 1`, actID).Scan(&text); err != nil {
		t.Fatalf("load event: %v", err)
	}
	if text == nil || *text != "on" {
		t.Fatalf("resulting_state_text: %v", text)
	}
}
