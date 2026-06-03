// Phase 36 WS4-prep — operator POST /actuators/{id}/command → pending_command.
package main

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestPhase36WS4Prep_OperatorActuatorCommand(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var deviceID, actuatorID int64
	var actuatorType string
	err := testPool.QueryRow(ctx, `
SELECT d.id, a.id, a.actuator_type
FROM gr33ncore.devices d
JOIN gr33ncore.actuators a ON a.device_id = d.id AND a.deleted_at IS NULL
WHERE d.farm_id = 1 AND a.actuator_type IN ('shade_screen', 'relay', 'light')
LIMIT 1`).Scan(&deviceID, &actuatorID, &actuatorType)
	if err != nil || deviceID == 0 {
		t.Skip("no device-bound actuator on farm 1")
	}

	command := "on"
	if actuatorType == "shade_screen" {
		command = "deploy"
	}

	_, _ = testPool.Exec(ctx, `UPDATE gr33ncore.devices SET config = '{}'::jsonb WHERE id = $1`, deviceID)
	t.Cleanup(func() {
		_, _ = testPool.Exec(context.Background(),
			`UPDATE gr33ncore.devices SET config = config - 'pending_command' WHERE id = $1`, deviceID)
	})

	tok := smokeJWT(t)
	resp := authPost(t, tok, fmt.Sprintf("/actuators/%d/command", actuatorID), map[string]any{
		"command": command,
		"reason":  "Phase 36 WS4-prep smoke",
	})
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusAccepted)

	body := decodeMap(t, resp)
	if body["command"].(string) != command {
		t.Fatalf("command %#v", body["command"])
	}

	listResp := authGet(t, tok, "/farms/1/devices")
	defer listResp.Body.Close()
	expectStatus(t, listResp, http.StatusOK)
	devices := decodeSlice(t, listResp)
	for _, row := range devices {
		d := row.(map[string]any)
		if int64(d["id"].(float64)) != deviceID {
			continue
		}
		cfg, err := deviceConfigFromJSONValue(d["config"])
		if err != nil {
			t.Fatalf("decode config: %v", err)
		}
		pc, ok := cfg["pending_command"].(map[string]any)
		if !ok {
			t.Fatalf("expected pending_command, config=%#v", cfg)
		}
		if pc["command"].(string) != command {
			t.Fatalf("pending command %#v", pc["command"])
		}
		if pc["source"].(string) != "operator" {
			t.Fatalf("source %#v", pc["source"])
		}
		return
	}
	t.Fatal("device not found in list")
}

func TestPhase36WS4Prep_ListActuatorsIncludesValidCommands(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/farms/1/actuators")
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusOK)
	rows := decodeSlice(t, resp)
	if len(rows) == 0 {
		t.Skip("no actuators on farm 1")
	}
	found := false
	for _, row := range rows {
		m := row.(map[string]any)
		vc, ok := m["valid_commands"].([]any)
		if ok && len(vc) > 0 {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected valid_commands on farm actuator list items")
	}
}
