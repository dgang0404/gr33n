// Phase 51 — Pi config platform sync smokes (edge config by uid + version bump).
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

func TestPhase51_ConfigByUIDAndVersion(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var deviceID int64
	var configVersion int32
	err := testPool.QueryRow(ctx, `
SELECT id, config_version
FROM gr33ncore.devices
WHERE farm_id = 1 AND device_uid = 'demo-veg-relay-01' AND deleted_at IS NULL
LIMIT 1`).Scan(&deviceID, &configVersion)
	if err != nil || deviceID == 0 {
		t.Skip("demo-veg-relay-01 missing — run migrations + master_seed.sql")
	}

	verResp := piGet(t, "/devices/by-uid/demo-veg-relay-01/config/version")
	expectStatus(t, verResp, http.StatusOK)
	verBody := decodeMap(t, verResp)
	verResp.Body.Close()
	if got := int32(verBody["config_version"].(float64)); got != configVersion {
		t.Fatalf("config_version = %d want %d from DB", got, configVersion)
	}

	cfgResp := piGet(t, "/devices/by-uid/demo-veg-relay-01/config")
	expectStatus(t, cfgResp, http.StatusOK)
	cfg := decodeMap(t, cfgResp)
	cfgResp.Body.Close()

	if cfg["device_uid"] != "demo-veg-relay-01" {
		t.Fatalf("device_uid = %v", cfg["device_uid"])
	}
	if int64(cfg["device_id"].(float64)) != deviceID {
		t.Fatalf("device_id = %v want %d", cfg["device_id"], deviceID)
	}
	if int32(cfg["config_version"].(float64)) != configVersion {
		t.Fatalf("config_version = %v want %d", cfg["config_version"], configVersion)
	}

	sensors, ok := cfg["sensors"].([]any)
	if !ok || len(sensors) == 0 {
		t.Fatalf("sensors = %#v", cfg["sensors"])
	}
	s0 := sensors[0].(map[string]any)
	if _, bad := s0["gpio_pin"]; bad {
		t.Fatalf("sensor must use pin not gpio_pin: %#v", s0)
	}

	actuators, ok := cfg["actuators"].([]any)
	if !ok || len(actuators) == 0 {
		t.Fatalf("actuators = %#v", cfg["actuators"])
	}

	badResp := piGet(t, "/devices/by-uid/no-such-pi-uid/config")
	expectStatus(t, badResp, http.StatusNotFound)
	badResp.Body.Close()
}

func TestPhase51_ConfigVersionBumpsOnWiringPatch(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var deviceID, sensorID int64
	var versionBefore int32
	var deviceIDWiring int64
	err := testPool.QueryRow(ctx, `
SELECT d.id, d.config_version, s.id,
       (s.config->'wiring'->>'device_id')::bigint
FROM gr33ncore.devices d
JOIN gr33ncore.sensors s ON (s.config->'wiring'->>'device_id')::bigint = d.id
WHERE d.farm_id = 1 AND d.device_uid = 'demo-veg-relay-01'
  AND d.deleted_at IS NULL AND s.deleted_at IS NULL
  AND s.name = 'Air Temp Indoor'
LIMIT 1`).Scan(&deviceID, &versionBefore, &sensorID, &deviceIDWiring)
	if err != nil || deviceID == 0 {
		t.Skip("demo wired Air Temp Indoor missing — run Phase 50/51 migrations + seed")
	}

	tok := smokeJWT(t)
	bumpNote := fmt.Sprintf("phase51 smoke bump %d", time.Now().UnixNano())
	resp := authPatch(t, tok, fmt.Sprintf("/sensors/%d/wiring", sensorID), map[string]any{
		"wiring": map[string]any{
			"source":    "dht22",
			"gpio_pin":  4,
			"device_id": deviceIDWiring,
			"notes":     bumpNote,
		},
	})
	expectStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	var versionAfter int32
	if err := testPool.QueryRow(ctx, `
SELECT config_version FROM gr33ncore.devices WHERE id = $1`, deviceID).Scan(&versionAfter); err != nil {
		t.Fatalf("read config_version: %v", err)
	}
	if versionAfter <= versionBefore {
		t.Fatalf("config_version = %d want > %d after wiring PATCH", versionAfter, versionBefore)
	}

	verResp := piGet(t, "/devices/by-uid/demo-veg-relay-01/config/version")
	expectStatus(t, verResp, http.StatusOK)
	verBody := decodeMap(t, verResp)
	verResp.Body.Close()
	if int32(verBody["config_version"].(float64)) != versionAfter {
		t.Fatalf("version endpoint = %v want %d", verBody["config_version"], versionAfter)
	}

	t.Cleanup(func() {
		_, _ = testPool.Exec(context.Background(), `
UPDATE gr33ncore.sensors
SET config = jsonb_set(
  config,
  '{wiring,notes}',
  '"Air Temp Indoor — DHT22 data line"'::jsonb
)
WHERE id = $1`, sensorID)
	})
}

func TestPhase51_StatusPatchStoresLastConfigFetchAt(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var deviceID int64
	err := testPool.QueryRow(ctx, `
SELECT id FROM gr33ncore.devices
WHERE farm_id = 1 AND device_uid = 'demo-veg-relay-01' AND deleted_at IS NULL
LIMIT 1`).Scan(&deviceID)
	if err != nil || deviceID == 0 {
		t.Skip("demo-veg-relay-01 missing")
	}

	ts := time.Now().UTC().Format(time.RFC3339)
	patchBody, _ := json.Marshal(map[string]string{
		"status":               "online",
		"last_config_fetch_at": ts,
	})
	req, _ := http.NewRequest(http.MethodPatch, testServer.URL+fmt.Sprintf("/devices/%d/status", deviceID), bytes.NewReader(patchBody))
	req.Header.Set("X-API-Key", piAPIKey)
	req.Header.Set("Content-Type", "application/json")
	patchResp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	expectStatus(t, patchResp, http.StatusOK)
	patchResp.Body.Close()

	var cfgText string
	if err := testPool.QueryRow(ctx, `SELECT config::text FROM gr33ncore.devices WHERE id = $1`, deviceID).Scan(&cfgText); err != nil {
		t.Fatalf("read config: %v", err)
	}
	if !bytes.Contains([]byte(cfgText), []byte("last_config_fetch_at")) {
		t.Fatalf("config missing last_config_fetch_at: %s", cfgText)
	}
}
