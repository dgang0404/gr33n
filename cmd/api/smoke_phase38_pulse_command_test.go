// Phase 38 — pending_command duration_seconds on actuator enqueue API.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	acthandler "gr33n-api/internal/handler/actuator"
)

func TestPhase38PulsePendingCommandShape(t *testing.T) {
	d := 2
	raw, err := acthandler.BuildPendingCommandJSONFull(acthandler.PendingCommandInput{
		ActuatorID: 1, Command: "on", Source: "operator", DurationSeconds: &d,
	})
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	if int(m["duration_seconds"].(float64)) != 2 {
		t.Fatalf("duration: %#v", m)
	}
	if !acthandler.PulseDurationAllowed("pump") {
		t.Fatal("pump should allow pulse")
	}
	if acthandler.ValidatePulseDuration("grow_light", &d) == nil {
		t.Fatal("grow_light should reject pulse duration")
	}
}

func TestPhase38EnqueueCommandWithDurationHTTP(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	uid := fmt.Sprintf("ph38_%d", time.Now().UnixNano())
	var devID, actID int64
	if err := testPool.QueryRow(ctx, `
		INSERT INTO gr33ncore.devices (farm_id, zone_id, name, device_uid, device_type, status, config)
		VALUES (1, 1, $1, $2, 'edge_gateway', 'online', '{}'::jsonb)
		RETURNING id
	`, uid+"_dev", uid+"_uid").Scan(&devID); err != nil {
		t.Fatalf("device: %v", err)
	}
	if err := testPool.QueryRow(ctx, `
		INSERT INTO gr33ncore.actuators (device_id, farm_id, zone_id, name, actuator_type, config, meta_data)
		VALUES ($1, 1, 1, $2, 'pump', '{}'::jsonb, '{}'::jsonb)
		RETURNING id
	`, devID, uid+"_pump").Scan(&actID); err != nil {
		t.Fatalf("actuator: %v", err)
	}
	t.Cleanup(func() {
		_, _ = testPool.Exec(context.Background(), `DELETE FROM gr33ncore.actuators WHERE id = $1`, actID)
		_, _ = testPool.Exec(context.Background(), `DELETE FROM gr33ncore.devices WHERE id = $1`, devID)
	})

	resp := authPost(t, smokeJWT(t), fmt.Sprintf("/actuators/%d/command", actID), map[string]any{
		"command": "on", "duration_seconds": 3,
	})
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusAccepted)
	var out map[string]any
	decodeJSON(t, resp.Body, &out)
	pc, ok := out["pending_command"].(map[string]any)
	if !ok {
		t.Fatalf("pending_command: %#v", out["pending_command"])
	}
	if int(pc["duration_seconds"].(float64)) != 3 {
		t.Fatalf("pending duration: %#v", pc)
	}
	if out["pulse_supported"] != true {
		t.Fatalf("pulse_supported: %#v", out["pulse_supported"])
	}
}
