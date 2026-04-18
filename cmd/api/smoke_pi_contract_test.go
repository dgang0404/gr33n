// Pi ↔ API contract — enqueue pending_command (as the worker does), poll
// GET /farms/{id}/devices with X-API-Key, POST /actuators/{id}/events with
// full provenance (schedule_id, rule_id, program_id), DELETE pending-command.
// Asserts actuator_events rows carry audit fields and invalid cross-farm
// references are rejected.
package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	db "gr33n-api/internal/db"
)

// deviceConfigFromJSONValue decodes gr33ncore.devices.config as returned by
// encoding/json on []byte fields (base64 string) or as an object (tests /
// alternate encoders).
func deviceConfigFromJSONValue(raw any) (map[string]any, error) {
	switch v := raw.(type) {
	case map[string]any:
		return v, nil
	case string:
		b, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			return nil, err
		}
		var m map[string]any
		if err := json.Unmarshal(b, &m); err != nil {
			return nil, err
		}
		return m, nil
	default:
		return map[string]any{}, nil
	}
}

func TestPiContractScheduleAndProgramFeedback(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	q := db.New(testPool)

	uid := fmt.Sprintf("pi_contract_%d", time.Now().UnixNano())
	var schedID int64
	if err := testPool.QueryRow(ctx, `
		INSERT INTO gr33ncore.schedules (farm_id, name, schedule_type, cron_expression, timezone, is_active, preconditions)
		VALUES (1, $1, 'cron', '* * * * *', 'UTC', TRUE, '[]'::jsonb)
		RETURNING id
	`, uid+"_sched").Scan(&schedID); err != nil {
		t.Fatalf("insert schedule: %v", err)
	}
	var progID int64
	if err := testPool.QueryRow(ctx, `
		INSERT INTO gr33nfertigation.programs (farm_id, name, schedule_id, total_volume_liters, is_active)
		VALUES (1, $1, $2, 1.0, TRUE)
		RETURNING id
	`, uid+"_prog", schedID).Scan(&progID); err != nil {
		t.Fatalf("insert program: %v", err)
	}
	var devID int64
	if err := testPool.QueryRow(ctx, `
		INSERT INTO gr33ncore.devices (farm_id, zone_id, name, device_uid, device_type, status, config)
		VALUES (1, 1, $1, $2, 'edge_gateway', 'online', '{}'::jsonb)
		RETURNING id
	`, uid+"_dev", uid+"_uid").Scan(&devID); err != nil {
		t.Fatalf("insert device: %v", err)
	}
	var actID int64
	if err := testPool.QueryRow(ctx, `
		INSERT INTO gr33ncore.actuators (device_id, farm_id, zone_id, name, actuator_type, config, meta_data)
		VALUES ($1, 1, 1, $2, 'relay', '{}'::jsonb, '{}'::jsonb)
		RETURNING id
	`, devID, uid+"_act").Scan(&actID); err != nil {
		t.Fatalf("insert actuator: %v", err)
	}
	t.Cleanup(func() {
		_, _ = testPool.Exec(context.Background(), `DELETE FROM gr33ncore.actuator_events WHERE actuator_id = $1`, actID)
		_, _ = testPool.Exec(context.Background(), `DELETE FROM gr33ncore.actuators WHERE id = $1`, actID)
		_, _ = testPool.Exec(context.Background(), `DELETE FROM gr33ncore.devices WHERE id = $1`, devID)
		_, _ = testPool.Exec(context.Background(), `DELETE FROM gr33nfertigation.programs WHERE id = $1`, progID)
		_, _ = testPool.Exec(context.Background(), `DELETE FROM gr33ncore.schedules WHERE id = $1`, schedID)
	})

	pending, _ := json.Marshal(map[string]any{
		"command":     "on",
		"schedule_id": schedID,
		"program_id":  progID,
	})
	if err := q.SetDevicePendingCommand(ctx, db.SetDevicePendingCommandParams{ID: devID, Column2: pending}); err != nil {
		t.Fatalf("set pending: %v", err)
	}

	// --- Pi poll (config is JSON-marshaled []byte → base64 string in API JSON) ---
	resp := piGet(t, "/farms/1/devices")
	expectStatus(t, resp, http.StatusOK)
	devices := decodeSlice(t, resp)
	var found bool
	for _, row := range devices {
		d := row.(map[string]any)
		if int64(d["id"].(float64)) != devID {
			continue
		}
		found = true
		cfg, err := deviceConfigFromJSONValue(d["config"])
		if err != nil {
			t.Fatalf("decode device config: %v", err)
		}
		pc, ok := cfg["pending_command"].(map[string]any)
		if !ok {
			t.Fatalf("expected pending_command object in config, got %#v", cfg["pending_command"])
		}
		if pc["command"].(string) != "on" || int64(pc["schedule_id"].(float64)) != schedID {
			t.Fatalf("unexpected pending_command: %#v", pc)
		}
		if int64(pc["program_id"].(float64)) != progID {
			t.Fatalf("expected program_id in pending, got %#v", pc["program_id"])
		}
	}
	if !found {
		t.Fatalf("device %d not returned from GET /farms/1/devices", devID)
	}

	// --- Pi feedback ---
	evBody := map[string]any{
		"command_sent":               "on",
		"source":                     "schedule_trigger",
		"execution_status":           "execution_completed_success_on_device",
		"triggered_by_schedule_id":   schedID,
		"program_id":                 progID,
		"parameters_sent":            map[string]any{"echo": "pi_contract"},
		"meta_data":                  map[string]any{"note": "smoke"},
	}
	evResp := piPostJSON(t, fmt.Sprintf("/actuators/%d/events", actID), evBody)
	expectStatus(t, evResp, http.StatusCreated)
	_ = decodeMap(t, evResp)

	cl := piDelete(t, fmt.Sprintf("/devices/%d/pending-command", devID))
	expectStatus(t, cl, http.StatusNoContent)
	cl.Body.Close()

	var trigSched *int64
	var trigRule *int64
	var metaJSON string
	var paramsJSON string
	if err := testPool.QueryRow(ctx, `
		SELECT triggered_by_schedule_id, triggered_by_rule_id, meta_data::text, parameters_sent::text
		FROM gr33ncore.actuator_events
		WHERE actuator_id = $1
		ORDER BY event_time DESC
		LIMIT 1
	`, actID).Scan(&trigSched, &trigRule, &metaJSON, &paramsJSON); err != nil {
		t.Fatalf("load event: %v", err)
	}
	if trigSched == nil || *trigSched != schedID {
		t.Fatalf("expected triggered_by_schedule_id=%d, got %v", schedID, trigSched)
	}
	if trigRule != nil {
		t.Fatalf("expected nil rule_id, got %v", *trigRule)
	}
	var meta map[string]any
	if err := json.Unmarshal([]byte(metaJSON), &meta); err != nil {
		t.Fatal(err)
	}
	if int64(meta["program_id"].(float64)) != progID {
		t.Fatalf("meta program_id: want %d got %v", progID, meta["program_id"])
	}
	if meta["reported_by"].(string) != "pi_client" {
		t.Fatalf("meta reported_by: %v", meta["reported_by"])
	}
	if meta["note"].(string) != "smoke" {
		t.Fatalf("meta note: %v", meta["note"])
	}
	var params map[string]any
	if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
		t.Fatal(err)
	}
	if params["echo"].(string) != "pi_contract" {
		t.Fatalf("parameters_sent: %v", params["echo"])
	}

	// Pending cleared
	var cfg []byte
	if err := testPool.QueryRow(ctx, `SELECT config::text FROM gr33ncore.devices WHERE id = $1`, devID).Scan(&cfg); err != nil {
		t.Fatal(err)
	}
	var cfgMap map[string]any
	_ = json.Unmarshal(cfg, &cfgMap)
	if _, ok := cfgMap["pending_command"]; ok {
		t.Fatalf("expected pending_command removed, config=%s", string(cfg))
	}
}

func TestPiContractRuleFeedback(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	q := db.New(testPool)

	uid := fmt.Sprintf("pi_rule_%d", time.Now().UnixNano())
	var ruleID int64
	if err := testPool.QueryRow(ctx, `
		INSERT INTO gr33ncore.automation_rules (
			farm_id, name, trigger_source, trigger_configuration,
			condition_logic, conditions_jsonb, is_active
		) VALUES (1, $1, 'sensor_reading_threshold', '{}'::jsonb, 'ALL', '[]'::jsonb, TRUE)
		RETURNING id
	`, uid+"_rule").Scan(&ruleID); err != nil {
		t.Fatalf("insert rule: %v", err)
	}
	var devID int64
	if err := testPool.QueryRow(ctx, `
		INSERT INTO gr33ncore.devices (farm_id, zone_id, name, device_uid, device_type, status, config)
		VALUES (1, 1, $1, $2, 'edge_gateway', 'online', '{}'::jsonb)
		RETURNING id
	`, uid+"_dev", uid+"_uid").Scan(&devID); err != nil {
		t.Fatalf("insert device: %v", err)
	}
	var actID int64
	if err := testPool.QueryRow(ctx, `
		INSERT INTO gr33ncore.actuators (device_id, farm_id, zone_id, name, actuator_type, config, meta_data)
		VALUES ($1, 1, 1, $2, 'relay', '{}'::jsonb, '{}'::jsonb)
		RETURNING id
	`, devID, uid+"_act").Scan(&actID); err != nil {
		t.Fatalf("insert actuator: %v", err)
	}
	t.Cleanup(func() {
		_, _ = testPool.Exec(context.Background(), `DELETE FROM gr33ncore.actuator_events WHERE actuator_id = $1`, actID)
		_, _ = testPool.Exec(context.Background(), `DELETE FROM gr33ncore.actuators WHERE id = $1`, actID)
		_, _ = testPool.Exec(context.Background(), `DELETE FROM gr33ncore.devices WHERE id = $1`, devID)
		_, _ = testPool.Exec(context.Background(), `DELETE FROM gr33ncore.automation_rules WHERE id = $1`, ruleID)
	})

	pending, _ := json.Marshal(map[string]any{"command": "off", "rule_id": ruleID})
	if err := q.SetDevicePendingCommand(ctx, db.SetDevicePendingCommandParams{ID: devID, Column2: pending}); err != nil {
		t.Fatalf("set pending: %v", err)
	}

	evResp := piPostJSON(t, fmt.Sprintf("/actuators/%d/events", actID), map[string]any{
		"command_sent":           "off",
		"source":                 "automation_rule_trigger",
		"execution_status":       "execution_completed_success_on_device",
		"triggered_by_rule_id":   ruleID,
	})
	expectStatus(t, evResp, http.StatusCreated)
	_ = decodeMap(t, evResp)

	cl := piDelete(t, fmt.Sprintf("/devices/%d/pending-command", devID))
	expectStatus(t, cl, http.StatusNoContent)
	cl.Body.Close()

	var trigRule *int64
	if err := testPool.QueryRow(ctx, `
		SELECT triggered_by_rule_id FROM gr33ncore.actuator_events
		WHERE actuator_id = $1 ORDER BY event_time DESC LIMIT 1
	`, actID).Scan(&trigRule); err != nil {
		t.Fatalf("load event: %v", err)
	}
	if trigRule == nil || *trigRule != ruleID {
		t.Fatalf("expected triggered_by_rule_id=%d, got %v", ruleID, trigRule)
	}
}

func TestPiContractRejectRuleAndProgramTogether(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	uid := fmt.Sprintf("pi_bad_%d", time.Now().UnixNano())
	var schedID int64
	_ = testPool.QueryRow(ctx, `
		INSERT INTO gr33ncore.schedules (farm_id, name, schedule_type, cron_expression, timezone, is_active, preconditions)
		VALUES (1, $1, 'cron', '0 0 * * *', 'UTC', TRUE, '[]'::jsonb)
		RETURNING id
	`, uid+"_s").Scan(&schedID)
	var ruleID int64
	_ = testPool.QueryRow(ctx, `
		INSERT INTO gr33ncore.automation_rules (farm_id, name, trigger_source, trigger_configuration, is_active)
		VALUES (1, $1, 'sensor_reading_threshold', '{}'::jsonb, TRUE)
		RETURNING id
	`, uid+"_r").Scan(&ruleID)
	var progID int64
	_ = testPool.QueryRow(ctx, `
		INSERT INTO gr33nfertigation.programs (farm_id, name, schedule_id, total_volume_liters, is_active)
		VALUES (1, $1, $2, 1.0, TRUE)
		RETURNING id
	`, uid+"_p", schedID).Scan(&progID)
	var devID int64
	_ = testPool.QueryRow(ctx, `
		INSERT INTO gr33ncore.devices (farm_id, zone_id, name, device_uid, device_type, status, config)
		VALUES (1, 1, $1, $2, 'edge_gateway', 'online', '{}'::jsonb)
		RETURNING id
	`, uid+"_d", uid+"_u").Scan(&devID)
	var actID int64
	_ = testPool.QueryRow(ctx, `
		INSERT INTO gr33ncore.actuators (device_id, farm_id, zone_id, name, actuator_type, config, meta_data)
		VALUES ($1, 1, 1, $2, 'relay', '{}'::jsonb, '{}'::jsonb)
		RETURNING id
	`, devID, uid+"_a").Scan(&actID)
	t.Cleanup(func() {
		_, _ = testPool.Exec(context.Background(), `DELETE FROM gr33ncore.actuator_events WHERE actuator_id = $1`, actID)
		_, _ = testPool.Exec(context.Background(), `DELETE FROM gr33ncore.actuators WHERE id = $1`, actID)
		_, _ = testPool.Exec(context.Background(), `DELETE FROM gr33ncore.devices WHERE id = $1`, devID)
		_, _ = testPool.Exec(context.Background(), `DELETE FROM gr33nfertigation.programs WHERE id = $1`, progID)
		_, _ = testPool.Exec(context.Background(), `DELETE FROM gr33ncore.automation_rules WHERE id = $1`, ruleID)
		_, _ = testPool.Exec(context.Background(), `DELETE FROM gr33ncore.schedules WHERE id = $1`, schedID)
	})

	evResp := piPostJSON(t, fmt.Sprintf("/actuators/%d/events", actID), map[string]any{
		"command_sent":         "on",
		"source":               "automation_rule_trigger",
		"triggered_by_rule_id": ruleID,
		"program_id":           progID,
	})
	expectStatus(t, evResp, http.StatusBadRequest)
	evResp.Body.Close()
}

func TestPiContractRejectUnknownSchedule(t *testing.T) {
	var devID int64
	uid := fmt.Sprintf("pi_xf_%d", time.Now().UnixNano())
	if err := testPool.QueryRow(context.Background(), `
		INSERT INTO gr33ncore.devices (farm_id, zone_id, name, device_uid, device_type, status, config)
		VALUES (1, 1, $1, $2, 'edge_gateway', 'online', '{}'::jsonb)
		RETURNING id
	`, uid+"_d", uid+"_u").Scan(&devID); err != nil {
		t.Fatalf("insert device: %v", err)
	}
	var actID int64
	if err := testPool.QueryRow(context.Background(), `
		INSERT INTO gr33ncore.actuators (device_id, farm_id, zone_id, name, actuator_type, config, meta_data)
		VALUES ($1, 1, 1, $2, 'relay', '{}'::jsonb, '{}'::jsonb)
		RETURNING id
	`, devID, uid+"_a").Scan(&actID); err != nil {
		t.Fatalf("insert actuator: %v", err)
	}
	t.Cleanup(func() {
		_, _ = testPool.Exec(context.Background(), `DELETE FROM gr33ncore.actuators WHERE id = $1`, actID)
		_, _ = testPool.Exec(context.Background(), `DELETE FROM gr33ncore.devices WHERE id = $1`, devID)
	})

	unknownScheduleID := int64(999999999)
	evResp := piPostJSON(t, fmt.Sprintf("/actuators/%d/events", actID), map[string]any{
		"command_sent":             "on",
		"source":                   "schedule_trigger",
		"triggered_by_schedule_id": unknownScheduleID,
	})
	expectStatus(t, evResp, http.StatusBadRequest)
	evResp.Body.Close()
}
