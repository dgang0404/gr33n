// Phase 30 WS4 — enqueue_actuator_command propose→confirm → pending_command on device.
package main

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPhase30WS4_EnqueueActuatorConfirmSetsPendingCommand(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	var deviceID, actuatorID int64
	err := testPool.QueryRow(ctx, `
SELECT d.id, a.id
FROM gr33ncore.devices d
JOIN gr33ncore.actuators a ON a.device_id = d.id AND a.deleted_at IS NULL
WHERE d.farm_id = 1 AND d.device_uid = 'demo-veg-relay-01'
LIMIT 1`).Scan(&deviceID, &actuatorID)
	if err != nil || deviceID == 0 {
		t.Skip("demo veg relay device/actuator missing — run master_seed.sql")
	}

	_, _ = testPool.Exec(ctx, `UPDATE gr33ncore.devices SET config = '{}'::jsonb WHERE id = $1`, deviceID)

	uid := uuid.MustParse(smokeDevUserUUID)
	args, _ := json.Marshal(map[string]any{
		"device_id":   deviceID,
		"actuator_id": actuatorID,
		"command":     "on",
		"reason":      "Guardian smoke: operator requested lights on for inspection",
	})
	var proposalID string
	err = testPool.QueryRow(ctx, `
INSERT INTO gr33ncore.guardian_action_proposals
    (user_id, farm_id, tool_id, args, summary, risk_tier, expires_at)
VALUES ($1, 1, 'enqueue_actuator_command', $2::jsonb, 'Turn on Veg Room Grow Light', 'high', NOW() + INTERVAL '5 minutes')
RETURNING proposal_id::text`, uid, args).Scan(&proposalID)
	if err != nil {
		t.Fatalf("insert proposal: %v", err)
	}
	t.Cleanup(func() {
		_, _ = testPool.Exec(context.Background(), `UPDATE gr33ncore.devices SET config = config - 'pending_command' WHERE id = $1`, deviceID)
	})

	tok := smokeJWT(t)
	confirmResp := authPost(t, tok, "/v1/chat/confirm", map[string]string{"proposal_id": proposalID})
	defer confirmResp.Body.Close()
	expectStatus(t, confirmResp, http.StatusOK)

	listResp := authGet(t, tok, "/farms/1/devices")
	defer listResp.Body.Close()
	expectStatus(t, listResp, http.StatusOK)

	devices := decodeSlice(t, listResp)
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
		if !ok {
			t.Fatalf("expected pending_command, config=%#v", cfg)
		}
		if pc["command"].(string) != "on" {
			t.Fatalf("command %#v", pc["command"])
		}
		if int64(pc["actuator_id"].(float64)) != actuatorID {
			t.Fatalf("actuator_id %#v", pc["actuator_id"])
		}
		if pc["proposal_id"].(string) != proposalID {
			t.Fatalf("proposal_id %#v want %s", pc["proposal_id"], proposalID)
		}
		if pc["source"].(string) != "guardian" {
			t.Fatalf("source %#v", pc["source"])
		}
	}
	if !found {
		t.Fatalf("device %d not in GET /farms/1/devices", deviceID)
	}

	var auditCount int
	_ = testPool.QueryRow(ctx, `
SELECT COUNT(*) FROM gr33ncore.user_activity_log
WHERE action_type = 'guardian_tool_executed'
  AND details->>'tool_id' = 'enqueue_actuator_command'
  AND details->>'proposal_id' = $1`, proposalID).Scan(&auditCount)
	if auditCount < 1 {
		t.Fatal("expected guardian_tool_executed audit row")
	}
}
