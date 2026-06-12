// Phase 70 — relay-HAT export + config_version on assign smoke.
package main

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestPhase70_RelayHATExportAndAssignVersionBump(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var deviceID, actuatorID int64
	var versionBefore int32
	var hwBefore *string
	err := testPool.QueryRow(ctx, `
SELECT d.id, a.id, d.config_version, a.hardware_identifier
FROM gr33ncore.devices d
JOIN gr33ncore.actuators a ON a.device_id = d.id AND a.deleted_at IS NULL
WHERE d.farm_id = 1 AND d.device_uid = 'demo-veg-relay-01'
  AND (a.config->'wiring'->>'gpio_pin') IS NULL
  AND a.hardware_identifier IS NOT NULL
LIMIT 1`).Scan(&deviceID, &actuatorID, &versionBefore, &hwBefore)
	if err != nil || deviceID == 0 {
		t.Skip("demo-veg-relay-01 HAT-only actuator missing")
	}

	tok := smokeJWT(t)
	newChannel := fmt.Sprintf("phase70_%d", time.Now().UnixNano()%50)
	resp := authPatch(t, tok, fmt.Sprintf("/actuators/%d/assign", actuatorID), map[string]any{
		"hardware_identifier": newChannel,
	})
	expectStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	var versionAfter int32
	if err := testPool.QueryRow(ctx, `
SELECT config_version FROM gr33ncore.devices WHERE id = $1`, deviceID).Scan(&versionAfter); err != nil {
		t.Fatalf("read config_version: %v", err)
	}
	if versionAfter <= versionBefore {
		t.Fatalf("config_version = %d want > %d after assign", versionAfter, versionBefore)
	}

	cfgResp := piGet(t, "/devices/by-uid/demo-veg-relay-01/config")
	expectStatus(t, cfgResp, http.StatusOK)
	cfg := decodeMap(t, cfgResp)
	cfgResp.Body.Close()

	actuators, _ := cfg["actuators"].([]any)
	var found bool
	for _, raw := range actuators {
		row, _ := raw.(map[string]any)
		if int64(row["actuator_id"].(float64)) != actuatorID {
			continue
		}
		if row["driver"] != "relay_hat" {
			t.Fatalf("expected relay_hat driver: %#v", row)
		}
		if row["channel"] == nil {
			t.Fatalf("expected channel on HAT actuator: %#v", row)
		}
		found = true
		break
	}
	if !found {
		t.Fatalf("actuator %d not in pi runtime config: %#v", actuatorID, actuators)
	}

	t.Cleanup(func() {
		if hwBefore != nil {
			resp := authPatch(t, tok, fmt.Sprintf("/actuators/%d/assign", actuatorID), map[string]any{
				"hardware_identifier": *hwBefore,
			})
			resp.Body.Close()
		}
	})
}
