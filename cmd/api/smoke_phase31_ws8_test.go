// Phase 31 WS8 — edge loop smokes: WS1 reading ingest path + WS3 pending_command round-trip.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	db "gr33n-api/internal/db"
)

func TestPhase31WS8_EdgePostReadingLatestNot404(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var sensorID int64
	err := testPool.QueryRow(ctx, `
SELECT MIN(id) FROM gr33ncore.sensors
WHERE farm_id = 1 AND deleted_at IS NULL AND name = 'Air Temp Indoor'
`).Scan(&sensorID)
	if err != nil || sensorID == 0 {
		t.Skip("Air Temp Indoor sensor missing — run master_seed.sql")
	}

	unique := float64(time.Now().UnixNano()%10000)/100.0 + 20.0
	now := time.Now().UTC().Format(time.RFC3339)
	body := map[string]any{
		"sensor_id":    sensorID,
		"value_raw":    unique,
		"reading_time": now,
		"is_valid":     true,
	}
	resp := piPostJSON(t, fmt.Sprintf("/sensors/%d/readings", sensorID), body)
	expectStatus(t, resp, http.StatusCreated)
	resp.Body.Close()

	tok := smokeJWT(t)
	latest := authGet(t, tok, fmt.Sprintf("/sensors/%d/readings/latest", sensorID))
	defer latest.Body.Close()
	if latest.StatusCode == http.StatusNotFound {
		t.Fatalf("GET /sensors/%d/readings/latest returned 404 after Pi post — WS1 edge loop broken", sensorID)
	}
	expectStatus(t, latest, http.StatusOK)
	row := decodeMap(t, latest)
	if got, ok := row["value_raw"].(float64); !ok || got != unique {
		t.Fatalf("latest value_raw = %v want %v", row["value_raw"], unique)
	}
}

func TestPhase31WS8_EdgePendingCommandPiRoundTrip(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	q := db.New(testPool)

	var deviceID, actuatorID int64
	err := testPool.QueryRow(ctx, `
SELECT d.id, a.id
FROM gr33ncore.devices d
JOIN gr33ncore.actuators a ON a.device_id = d.id AND a.deleted_at IS NULL
WHERE d.farm_id = 1 AND d.device_uid = 'demo-veg-relay-01'
LIMIT 1`).Scan(&deviceID, &actuatorID)
	if err != nil || deviceID == 0 {
		t.Skip("demo-veg-relay-01 missing — run master_seed.sql")
	}

	t.Cleanup(func() {
		_, _ = testPool.Exec(context.Background(),
			`UPDATE gr33ncore.devices SET config = config - 'pending_command' WHERE id = $1`, deviceID)
	})

	pending, _ := json.Marshal(map[string]any{
		"command":     "on",
		"actuator_id": actuatorID,
		"source":      "bench",
		"reason":      "Phase 31 WS8 smoke",
	})
	if err := q.SetDevicePendingCommand(ctx, db.SetDevicePendingCommandParams{
		ID: deviceID, Column2: pending,
	}); err != nil {
		t.Fatalf("set pending: %v", err)
	}

	devResp := piGet(t, "/farms/1/devices")
	expectStatus(t, devResp, http.StatusOK)
	devices := decodeSlice(t, devResp)
	devResp.Body.Close()

	var found bool
	for _, row := range devices {
		d := row.(map[string]any)
		if int64(d["id"].(float64)) != deviceID {
			continue
		}
		found = true
		cfg, err := deviceConfigFromJSONValue(d["config"])
		if err != nil {
			t.Fatalf("decode config: %v", err)
		}
		pc, ok := cfg["pending_command"].(map[string]any)
		if !ok || pc["command"].(string) != "on" {
			t.Fatalf("expected pending_command on device %d, config=%#v", deviceID, cfg)
		}
	}
	if !found {
		t.Fatalf("device %d not in Pi GET /farms/1/devices", deviceID)
	}

	evBody := map[string]any{
		"command_sent":     "on",
		"source":           "manual_api_call",
		"execution_status": "command_sent_to_device",
		"meta_data":        map[string]any{"note": "phase31_ws8"},
	}
	evResp := piPostJSON(t, fmt.Sprintf("/actuators/%d/events", actuatorID), evBody)
	expectStatus(t, evResp, http.StatusCreated)
	evResp.Body.Close()

	cl := piDelete(t, fmt.Sprintf("/devices/%d/pending-command", deviceID))
	expectStatus(t, cl, http.StatusNoContent)
	cl.Body.Close()

	var stillPending string
	_ = testPool.QueryRow(ctx, `
SELECT CASE WHEN config ? 'pending_command' THEN 'yes' ELSE 'no' END
FROM gr33ncore.devices WHERE id = $1`, deviceID).Scan(&stillPending)
	if stillPending != "no" {
		t.Fatalf("pending_command still set after Pi clear, got %q", stillPending)
	}
}

// Live GPIO E2E moved to the @hardware lane (build tag `hardware`):
// cmd/api/smoke_hardware_test.go. Default `make test` / CI (`-tags dev`) never
// compile it. See Phase 33 WS4.
