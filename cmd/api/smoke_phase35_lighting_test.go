// Phase 35 WS8 / OC-35C — lighting program preset apply, list, schedule actions, deactivate.
package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

func skipIfLightingProgramsMissing(t *testing.T) {
	t.Helper()
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var ok bool
	err := testPool.QueryRow(ctx, `
SELECT EXISTS (
  SELECT 1 FROM information_schema.tables
  WHERE table_schema = 'gr33ncore' AND table_name = 'lighting_programs'
)`).Scan(&ok)
	if err != nil || !ok {
		t.Skip("lighting_programs table missing — apply Phase 35 migration")
	}
}

func seedLightingZone(t *testing.T, tok string) int64 {
	t.Helper()
	resp := authPost(t, tok, "/farms/1/zones", map[string]any{
		"name":        uniqueName("ph35_zone"),
		"description": "Phase 35 lighting smoke zone",
		"zone_type":   "indoor",
	})
	expectStatus(t, resp, http.StatusCreated)
	return int64(decodeMap(t, resp)["id"].(float64))
}

func seedLightActuator(t *testing.T, tok string, zoneID int64) int64 {
	t.Helper()
	resp := authPost(t, tok, "/farms/1/actuators", map[string]any{
		"name":          uniqueName("ph35_light"),
		"actuator_type": "light",
		"zone_id":       zoneID,
	})
	if resp.StatusCode == http.StatusCreated {
		return int64(decodeMap(t, resp)["id"].(float64))
	}
	resp.Body.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var id int64
	if err := testPool.QueryRow(ctx, `
INSERT INTO gr33ncore.actuators (farm_id, zone_id, name, actuator_type)
VALUES (1, $1, $2, 'light')
RETURNING id`, zoneID, uniqueName("ph35_light_sql")).Scan(&id); err != nil {
		t.Fatalf("seed light actuator: %v", err)
	}
	return id
}

func TestPhase35WS8_LightingPresetsContract(t *testing.T) {
	skipIfLightingProgramsMissing(t)
	tok := smokeJWT(t)

	resp := authGet(t, tok, "/lighting-programs/presets")
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusOK)
	presets := decodeSlice(t, resp)
	if len(presets) < 4 {
		t.Fatalf("expected >=4 presets, got %d", len(presets))
	}
	found := false
	for _, raw := range presets {
		p, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if p["key"] == "veg_18_6" {
			found = true
			if int(p["on_hours"].(float64)) != 18 {
				t.Fatalf("veg_18_6 on_hours: %#v", p["on_hours"])
			}
		}
	}
	if !found {
		t.Fatal("expected veg_18_6 preset in list")
	}
}

func TestPhase35WS8_PresetApplyListScheduleActionsDeactivate(t *testing.T) {
	skipIfLightingProgramsMissing(t)
	tok := smokeJWT(t)
	zoneID := seedLightingZone(t, tok)
	actuatorID := seedLightActuator(t, tok, zoneID)
	progName := uniqueName("ph35_veg_program")

	t.Cleanup(func() {
		if testPool == nil {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = testPool.Exec(ctx, `
DELETE FROM gr33ncore.lighting_programs WHERE farm_id = 1 AND name = $1`, progName)
		_, _ = testPool.Exec(ctx, `DELETE FROM gr33ncore.zones WHERE id = $1`, zoneID)
	})

	createResp := authPost(t, tok, "/farms/1/lighting-programs/from-preset", map[string]any{
		"preset_key":  "veg_18_6",
		"name":        progName,
		"zone_id":     zoneID,
		"actuator_id": actuatorID,
		"lights_on_at": "06:00",
		"timezone":    "America/New_York",
	})
	defer createResp.Body.Close()
	if createResp.StatusCode == http.StatusInternalServerError {
		body := readBodyPreview(createResp)
		if strings.Contains(body, "lighting_programs") {
			t.Skip("lighting_programs not migrated")
		}
	}
	expectStatus(t, createResp, http.StatusCreated)
	prog := decodeMap(t, createResp)
	progID := int64(prog["id"].(float64))
	if int(prog["on_hours"].(float64)) != 18 {
		t.Fatalf("on_hours: %#v", prog["on_hours"])
	}
	if int(prog["off_hours"].(float64)) != 6 {
		t.Fatalf("off_hours: %#v", prog["off_hours"])
	}
	if prog["timezone"].(string) != "America/New_York" {
		t.Fatalf("timezone: %#v", prog["timezone"])
	}
	schOnID := int64(prog["schedule_on_id"].(float64))
	schOffID := int64(prog["schedule_off_id"].(float64))
	if schOnID <= 0 || schOffID <= 0 {
		t.Fatalf("expected linked schedules, got on=%d off=%d", schOnID, schOffID)
	}

	listResp := authGet(t, tok, "/farms/1/lighting-programs")
	defer listResp.Body.Close()
	expectStatus(t, listResp, http.StatusOK)
	rows := decodeSlice(t, listResp)
	found := false
	for _, raw := range rows {
		row, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if int64(row["id"].(float64)) == progID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("program %d not in farm list", progID)
	}

	actionsResp := authGet(t, tok, "/schedules/"+strconv.FormatInt(schOnID, 10)+"/actions")
	defer actionsResp.Body.Close()
	expectStatus(t, actionsResp, http.StatusOK)
	actions := decodeSlice(t, actionsResp)
	if len(actions) != 1 {
		t.Fatalf("expected 1 ON schedule action, got %d", len(actions))
	}
	act := actions[0].(map[string]any)
	if act["action_type"].(string) != "control_actuator" {
		t.Fatalf("action_type: %#v", act["action_type"])
	}
	if act["action_command"].(string) != "on" {
		t.Fatalf("action_command: %#v", act["action_command"])
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var onCron, onTZ string
	err := testPool.QueryRow(ctx, `
SELECT cron_expression, timezone FROM gr33ncore.schedules WHERE id = $1`, schOnID).
		Scan(&onCron, &onTZ)
	if err != nil {
		t.Fatalf("load ON schedule: %v", err)
	}
	if onCron != "0 6 * * *" {
		t.Fatalf("ON cron = %q, want 0 6 * * *", onCron)
	}
	if onTZ != "America/New_York" {
		t.Fatalf("ON timezone = %q, want America/New_York", onTZ)
	}

	deactResp := authPost(t, tok, fmt.Sprintf("/lighting-programs/%d/deactivate", progID), map[string]any{})
	defer deactResp.Body.Close()
	expectStatus(t, deactResp, http.StatusOK)
	deact := decodeMap(t, deactResp)
	if deact["is_active"].(bool) {
		t.Fatal("expected program inactive after deactivate")
	}

	var progActive bool
	err = testPool.QueryRow(ctx, `SELECT is_active FROM gr33ncore.lighting_programs WHERE id = $1`, progID).Scan(&progActive)
	if err != nil {
		t.Fatalf("reload program: %v", err)
	}
	if progActive {
		t.Fatal("DB program still active")
	}

	delResp := authDelete(t, tok, fmt.Sprintf("/lighting-programs/%d", progID))
	defer delResp.Body.Close()
	expectStatus(t, delResp, http.StatusNoContent)
}
