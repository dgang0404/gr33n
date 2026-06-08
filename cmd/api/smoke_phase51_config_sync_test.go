// Phase 51 WS1 — Pi runtime config sync by device_uid (X-API-Key).
package main

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestPhase51WS1_ConfigByUIDAndVersion(t *testing.T) {
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
	if int64(cfg["farm_id"].(float64)) != 1 {
		t.Fatalf("farm_id = %v", cfg["farm_id"])
	}
	if int32(cfg["config_version"].(float64)) != configVersion {
		t.Fatalf("config_version = %v want %d", cfg["config_version"], configVersion)
	}

	sensors, ok := cfg["sensors"].([]any)
	if !ok {
		t.Fatalf("sensors type %T", cfg["sensors"])
	}
	if len(sensors) == 0 {
		t.Fatal("expected at least one wired sensor for demo-veg-relay-01")
	}
	s0 := sensors[0].(map[string]any)
	if _, hasPin := s0["pin"]; !hasPin {
		if _, hasCh := s0["channel"]; !hasCh {
			t.Fatalf("sensor entry needs pin or channel: %#v", s0)
		}
	}
	if _, bad := s0["gpio_pin"]; bad {
		t.Fatalf("sensor must expose pin not gpio_pin: %#v", s0)
	}

	actuators, ok := cfg["actuators"].([]any)
	if !ok || len(actuators) == 0 {
		t.Fatalf("actuators = %#v", cfg["actuators"])
	}
	a0 := actuators[0].(map[string]any)
	if _, ok := a0["gpio_pin"]; !ok {
		t.Fatalf("actuator missing gpio_pin: %#v", a0)
	}

	if int(cfg["schedule_poll_interval_seconds"].(float64)) != 30 {
		t.Fatalf("schedule_poll_interval_seconds = %v", cfg["schedule_poll_interval_seconds"])
	}
	if cfg["offline_queue_path"] != "/var/lib/gr33n/queue.db" {
		t.Fatalf("offline_queue_path = %v", cfg["offline_queue_path"])
	}

	badResp := piGet(t, "/devices/by-uid/no-such-pi-uid/config")
	expectStatus(t, badResp, http.StatusNotFound)
	badResp.Body.Close()

	noKeyReq, _ := http.NewRequest(http.MethodGet, testServer.URL+"/devices/by-uid/demo-veg-relay-01/config", nil)
	noKeyResp, err := http.DefaultClient.Do(noKeyReq)
	if err != nil {
		t.Fatal(err)
	}
	if noKeyResp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("missing API key: got %d want 401", noKeyResp.StatusCode)
	}
	noKeyResp.Body.Close()
}
