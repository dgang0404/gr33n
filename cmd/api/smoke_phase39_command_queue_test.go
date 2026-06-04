// Phase 39 WS1 — device command queue: enqueue, dequeue (next), ack.
// Tests the full operator-enqueue → Pi-next → Pi-ack cycle against a real DB.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

// TestPhase39CommandQueueEnqueueNextAck covers the full happy path:
//  1. Operator enqueues a pulse command via POST /devices/{id}/commands
//  2. Pi polls GET /devices/{id}/commands/next — receives the command (in_progress)
//  3. Pi acks POST /devices/{id}/commands/{cid}/ack (status=completed)
//  4. GET /devices/{id}/commands shows the completed row
//
// Also asserts backward-compat: devices.config.pending_command is mirrored
// after enqueue and cleared after ack.
func TestPhase39CommandQueueEnqueueNextAck(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	uid := fmt.Sprintf("ph39_%d", time.Now().UnixNano())
	var devID, actID int64

	if err := testPool.QueryRow(ctx, `
		INSERT INTO gr33ncore.devices (farm_id, zone_id, name, device_uid, device_type, status, config)
		VALUES (1, 1, $1, $2, 'edge_gateway', 'online', '{}'::jsonb)
		RETURNING id
	`, uid+"_dev", uid+"_uid").Scan(&devID); err != nil {
		t.Fatalf("insert device: %v", err)
	}
	if err := testPool.QueryRow(ctx, `
		INSERT INTO gr33ncore.actuators (device_id, farm_id, zone_id, name, actuator_type, config, meta_data)
		VALUES ($1, 1, 1, $2, 'pump', '{}'::jsonb, '{}'::jsonb)
		RETURNING id
	`, devID, uid+"_pump").Scan(&actID); err != nil {
		t.Fatalf("insert actuator: %v", err)
	}
	t.Cleanup(func() {
		bg := context.Background()
		_, _ = testPool.Exec(bg, `DELETE FROM gr33ncore.device_commands WHERE device_id = $1`, devID)
		_, _ = testPool.Exec(bg, `DELETE FROM gr33ncore.actuators WHERE id = $1`, actID)
		_, _ = testPool.Exec(bg, `DELETE FROM gr33ncore.devices WHERE id = $1`, devID)
	})

	jwt := smokeJWT(t)

	// ── Step 1: enqueue via operator API ────────────────────────────────────
	enqResp := authPost(t, jwt, fmt.Sprintf("/devices/%d/commands", devID), map[string]any{
		"command_type":     "pulse",
		"actuator_id":      actID,
		"command":          "on",
		"duration_seconds": 2,
	})
	defer enqResp.Body.Close()
	expectStatus(t, enqResp, http.StatusAccepted)

	var enqOut map[string]any
	if err := json.NewDecoder(enqResp.Body).Decode(&enqOut); err != nil {
		t.Fatalf("decode enqueue response: %v", err)
	}
	cmdIDFloat, ok := enqOut["id"].(float64)
	if !ok || cmdIDFloat == 0 {
		t.Fatalf("expected command id in response; got %#v", enqOut)
	}
	cmdID := int64(cmdIDFloat)
	if enqOut["status"] != "pending" {
		t.Fatalf("expected status=pending; got %v", enqOut["status"])
	}
	if enqOut["command_type"] != "pulse" {
		t.Fatalf("expected command_type=pulse; got %v", enqOut["command_type"])
	}

	// Verify backward-compat mirror: config.pending_command should be set.
	var pendingRaw []byte
	if err := testPool.QueryRow(ctx,
		`SELECT config->'pending_command' FROM gr33ncore.devices WHERE id = $1`, devID,
	).Scan(&pendingRaw); err != nil {
		t.Fatalf("read pending_command: %v", err)
	}
	if len(pendingRaw) == 0 || string(pendingRaw) == "null" {
		t.Fatal("expected pending_command to be mirrored on device config after enqueue")
	}

	// ── Step 2: Pi polls next ────────────────────────────────────────────────
	nextResp := piGet(t, fmt.Sprintf("/devices/%d/commands/next", devID))
	defer nextResp.Body.Close()
	expectStatus(t, nextResp, http.StatusOK)

	var nextOut map[string]any
	if err := json.NewDecoder(nextResp.Body).Decode(&nextOut); err != nil {
		t.Fatalf("decode next response: %v", err)
	}
	if int64(nextOut["id"].(float64)) != cmdID {
		t.Fatalf("next returned id %v, want %v", nextOut["id"], cmdID)
	}
	if nextOut["status"] != "in_progress" {
		t.Fatalf("expected status=in_progress after next; got %v", nextOut["status"])
	}

	// Payload should contain the pulse command JSON.
	payload := extractPayload(t, nextOut)
	if payload["command"] != "on" {
		t.Fatalf("payload.command: %v", payload)
	}
	if int(payload["duration_seconds"].(float64)) != 2 {
		t.Fatalf("payload.duration_seconds: %v", payload)
	}

	// ── Step 3: Pi acks completed ────────────────────────────────────────────
	ackResp := piPostJSON(t, fmt.Sprintf("/devices/%d/commands/%d/ack", devID, cmdID),
		map[string]any{"status": "completed", "result": map[string]any{"ec_measured": 1.6}},
	)
	defer ackResp.Body.Close()
	expectStatus(t, ackResp, http.StatusOK)

	var ackOut map[string]any
	if err := json.NewDecoder(ackResp.Body).Decode(&ackOut); err != nil {
		t.Fatalf("decode ack response: %v", err)
	}
	if ackOut["status"] != "completed" {
		t.Fatalf("expected status=completed after ack; got %v", ackOut["status"])
	}

	// pending_command should be cleared after completed ack.
	var pendingAfter []byte
	if err := testPool.QueryRow(ctx,
		`SELECT config->'pending_command' FROM gr33ncore.devices WHERE id = $1`, devID,
	).Scan(&pendingAfter); err != nil {
		t.Fatalf("read pending_command after ack: %v", err)
	}
	if string(pendingAfter) != "null" && len(pendingAfter) != 0 {
		t.Fatalf("expected pending_command cleared after ack; got %s", pendingAfter)
	}

	// ── Step 4: operator lists completed commands ────────────────────────────
	listResp := authGet(t, jwt, fmt.Sprintf("/devices/%d/commands?status=completed", devID))
	defer listResp.Body.Close()
	expectStatus(t, listResp, http.StatusOK)
	var listOut map[string]any
	if err := json.NewDecoder(listResp.Body).Decode(&listOut); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	cmds, _ := listOut["commands"].([]any)
	if len(cmds) == 0 {
		t.Fatal("expected at least one completed command in list")
	}
}

// TestPhase39CommandQueueEmptyReturns204 verifies that polling an empty
// queue returns 204 No Content so the Pi knows it can sleep until next tick.
func TestPhase39CommandQueueEmptyReturns204(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	uid := fmt.Sprintf("ph39_empty_%d", time.Now().UnixNano())
	var devID int64
	if err := testPool.QueryRow(ctx, `
		INSERT INTO gr33ncore.devices (farm_id, zone_id, name, device_uid, device_type, status, config)
		VALUES (1, 1, $1, $2, 'edge_gateway', 'online', '{}'::jsonb)
		RETURNING id
	`, uid+"_dev", uid+"_uid").Scan(&devID); err != nil {
		t.Fatalf("insert device: %v", err)
	}
	t.Cleanup(func() {
		_, _ = testPool.Exec(context.Background(), `DELETE FROM gr33ncore.devices WHERE id = $1`, devID)
	})

	resp := piGet(t, fmt.Sprintf("/devices/%d/commands/next", devID))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204 for empty queue; got %d", resp.StatusCode)
	}
}

// TestPhase39CommandQueueOrderFIFO verifies that multiple enqueued commands
// are returned oldest-first (FIFO).
func TestPhase39CommandQueueOrderFIFO(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	uid := fmt.Sprintf("ph39_fifo_%d", time.Now().UnixNano())
	var devID, actID int64

	if err := testPool.QueryRow(ctx, `
		INSERT INTO gr33ncore.devices (farm_id, zone_id, name, device_uid, device_type, status, config)
		VALUES (1, 1, $1, $2, 'edge_gateway', 'online', '{}'::jsonb)
		RETURNING id
	`, uid+"_dev", uid+"_uid").Scan(&devID); err != nil {
		t.Fatalf("insert device: %v", err)
	}
	if err := testPool.QueryRow(ctx, `
		INSERT INTO gr33ncore.actuators (device_id, farm_id, zone_id, name, actuator_type, config, meta_data)
		VALUES ($1, 1, 1, $2, 'pump', '{}'::jsonb, '{}'::jsonb)
		RETURNING id
	`, devID, uid+"_pump").Scan(&actID); err != nil {
		t.Fatalf("insert actuator: %v", err)
	}
	t.Cleanup(func() {
		bg := context.Background()
		_, _ = testPool.Exec(bg, `DELETE FROM gr33ncore.device_commands WHERE device_id = $1`, devID)
		_, _ = testPool.Exec(bg, `DELETE FROM gr33ncore.actuators WHERE id = $1`, actID)
		_, _ = testPool.Exec(bg, `DELETE FROM gr33ncore.devices WHERE id = $1`, devID)
	})

	jwt := smokeJWT(t)

	// Enqueue "on" then "off" with a small sleep so created_at differs.
	r1 := authPost(t, jwt, fmt.Sprintf("/devices/%d/commands", devID), map[string]any{
		"command_type": "actuator", "actuator_id": actID, "command": "on",
	})
	r1.Body.Close()
	expectStatus(t, r1, http.StatusAccepted)
	time.Sleep(10 * time.Millisecond)

	r2 := authPost(t, jwt, fmt.Sprintf("/devices/%d/commands", devID), map[string]any{
		"command_type": "actuator", "actuator_id": actID, "command": "off",
	})
	r2.Body.Close()
	expectStatus(t, r2, http.StatusAccepted)

	// Pi should get "on" first.
	next1 := piGet(t, fmt.Sprintf("/devices/%d/commands/next", devID))
	defer next1.Body.Close()
	expectStatus(t, next1, http.StatusOK)
	var out1 map[string]any
	if err := json.NewDecoder(next1.Body).Decode(&out1); err != nil {
		t.Fatalf("decode next1: %v", err)
	}
	if extractPayload(t, out1)["command"] != "on" {
		t.Fatalf("FIFO order: expected first=on, got %v", extractPayload(t, out1)["command"])
	}

	// Ack the first command, then drain the second.
	id1 := int64(out1["id"].(float64))
	a1 := piPostJSON(t, fmt.Sprintf("/devices/%d/commands/%d/ack", devID, id1),
		map[string]any{"status": "completed"},
	)
	a1.Body.Close()
	expectStatus(t, a1, http.StatusOK)

	next2 := piGet(t, fmt.Sprintf("/devices/%d/commands/next", devID))
	defer next2.Body.Close()
	expectStatus(t, next2, http.StatusOK)
	var out2 map[string]any
	if err := json.NewDecoder(next2.Body).Decode(&out2); err != nil {
		t.Fatalf("decode next2: %v", err)
	}
	if extractPayload(t, out2)["command"] != "off" {
		t.Fatalf("FIFO order: expected second=off, got %v", extractPayload(t, out2)["command"])
	}
}

// extractPayload unpacks device_commands.payload from JSON (can be map or base64 string).
func extractPayload(t *testing.T, row map[string]any) map[string]any {
	t.Helper()
	raw := row["payload"]
	switch v := raw.(type) {
	case map[string]any:
		return v
	case string:
		var m map[string]any
		if err := json.Unmarshal([]byte(v), &m); err != nil {
			t.Fatalf("extractPayload: %v", err)
		}
		return m
	default:
		t.Fatalf("unexpected payload type %T: %v", raw, raw)
		return nil
	}
}
