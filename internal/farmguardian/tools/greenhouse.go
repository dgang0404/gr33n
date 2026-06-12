package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gr33n-api/internal/db"
	"gr33n-api/internal/platform/domainenums"
)

// execSummarizeZoneGreenhouseClimate — read tool for Phase 36 WS7.
// Returns the greenhouse_climate profile from zone meta_data, current
// actuator states for shade/fan/vent, active automation rules for the
// zone, and the most recent shade/fan actuator events.
func execSummarizeZoneGreenhouseClimate(ctx context.Context, deps ExecutorDeps, args map[string]any) (any, error) {
	zoneID, err := int64FromArgs(args, "zone_id")
	if err != nil {
		return nil, err
	}

	zone, err := deps.Q.GetZoneByID(ctx, zoneID)
	if err != nil {
		return nil, fmt.Errorf("zone %d not found", zoneID)
	}
	if zone.FarmID != deps.FarmID {
		return nil, fmt.Errorf("zone %d does not belong to this farm", zoneID)
	}

	zoneType := ""
	if zone.ZoneType != nil {
		zoneType = *zone.ZoneType
	}

	interlocks, _ := zoneSensorInterlocksForGuardian(ctx, deps.Q, zoneID)

	// Parse greenhouse_climate profile from meta_data.
	var profile map[string]any
	profileMissing := false
	if len(zone.MetaData) > 0 {
		var meta map[string]json.RawMessage
		if err := json.Unmarshal(zone.MetaData, &meta); err == nil {
			if gcRaw, ok := meta["greenhouse_climate"]; ok && len(gcRaw) > 0 {
				_ = json.Unmarshal(gcRaw, &profile)
			}
		}
	}
	if profile == nil {
		profileMissing = true
		profile = map[string]any{}
	}

	// Collect actuator IDs referenced by the profile.
	var trackedActuatorIDs []int64
	if v, ok := profile["shade_actuator_id"]; ok {
		if id, ok := jsonInt64(v); ok {
			trackedActuatorIDs = append(trackedActuatorIDs, id)
		}
	}
	if v, ok := profile["vent_actuator_id"]; ok {
		if id, ok := jsonInt64(v); ok {
			trackedActuatorIDs = append(trackedActuatorIDs, id)
		}
	}
	if fanIDs, ok := profile["fan_actuator_ids"]; ok {
		if ids, ok := fanIDs.([]any); ok {
			for _, raw := range ids {
				if id, ok := jsonInt64(raw); ok {
					trackedActuatorIDs = append(trackedActuatorIDs, id)
				}
			}
		}
	}

	// Load actuator states for tracked actuators.
	type actuatorSummary struct {
		ID           int64   `json:"id"`
		Name         string  `json:"name"`
		ActuatorType string  `json:"actuator_type"`
		StateText    *string `json:"state_text,omitempty"`
		Role         string  `json:"role"`
	}
	actuators := make([]actuatorSummary, 0, len(trackedActuatorIDs))

	shadeID := int64(-1)
	if v, ok := profile["shade_actuator_id"]; ok {
		if id, ok := jsonInt64(v); ok {
			shadeID = id
		}
	}
	ventID := int64(-1)
	if v, ok := profile["vent_actuator_id"]; ok {
		if id, ok := jsonInt64(v); ok {
			ventID = id
		}
	}
	fanIDSet := map[int64]struct{}{}
	if fanIDs, ok := profile["fan_actuator_ids"]; ok {
		if ids, ok := fanIDs.([]any); ok {
			for _, raw := range ids {
				if id, ok := jsonInt64(raw); ok {
					fanIDSet[id] = struct{}{}
				}
			}
		}
	}

	for _, aid := range trackedActuatorIDs {
		a, err := deps.Q.GetActuatorByID(ctx, aid)
		if err != nil {
			continue
		}
		role := "fan"
		if a.ID == shadeID {
			role = "shade"
		} else if a.ID == ventID {
			role = "vent"
		}
		actuators = append(actuators, actuatorSummary{
			ID:           a.ID,
			Name:         a.Name,
			ActuatorType: a.ActuatorType,
			StateText:    a.CurrentStateText,
			Role:         role,
		})
	}

	// Recent shade / fan actuator events (last 48 h across tracked actuators).
	type recentEvent struct {
		EventTime   time.Time `json:"event_time"`
		ActuatorID  int64     `json:"actuator_id"`
		CommandSent string    `json:"command_sent"`
		Source      string    `json:"source"`
	}
	recentEvents := make([]recentEvent, 0)
	since := time.Now().UTC().Add(-48 * time.Hour)
	for _, aid := range trackedActuatorIDs {
		evts, err := deps.Q.ListActuatorEventsByActuator(ctx, db.ListActuatorEventsByActuatorParams{
			ActuatorID: aid,
			EventTime:  since,
			Limit:      5,
		})
		if err != nil {
			continue
		}
		for _, e := range evts {
			cmd := ""
			if e.CommandSent != nil {
				cmd = *e.CommandSent
			}
			recentEvents = append(recentEvents, recentEvent{
				EventTime:   e.EventTime,
				ActuatorID:  e.ActuatorID,
				CommandSent: cmd,
				Source:      string(e.Source),
			})
		}
	}

	// Active rules for this farm that mention tracked actuators.
	allRules, err := deps.Q.ListAutomationRulesByFarm(ctx, deps.FarmID)
	if err != nil {
		allRules = nil
	}
	type ruleSummary struct {
		ID       int64  `json:"id"`
		Name     string `json:"name"`
		IsActive bool   `json:"is_active"`
	}
	ghRules := make([]ruleSummary, 0)
	for _, rule := range allRules {
		if strings.HasPrefix(rule.Name, "GH —") {
			ghRules = append(ghRules, ruleSummary{
				ID:       rule.ID,
				Name:     rule.Name,
				IsActive: rule.IsActive,
			})
		}
	}

	// Build human summary text.
	var lines []string
	lines = append(lines, fmt.Sprintf("Zone %q (id %d, type=%q)", zone.Name, zone.ID, zoneType))
	if profileMissing {
		lines = append(lines, "  ⚠ No greenhouse_climate profile in meta_data. Set cover_type, shade/fan actuator refs, and automation_policy via PUT /zones/{id}.")
	} else {
		ct, _ := profile["cover_type"].(string)
		pol, _ := profile["automation_policy"].(string)
		notes, _ := profile["notes"].(string)
		lines = append(lines, fmt.Sprintf("  Cover type: %s | Automation policy: %s",
			orNA(domainenums.GreenhouseCoverTypeLabel(ct)), orNA(domainenums.GreenhouseAutomationPolicyLabel(pol))))
		if notes != "" {
			lines = append(lines, "  Notes: "+notes)
		}
		if pol == "auto" && !interlocks.HasLux {
			lines = append(lines, "  ⚠ No lux/PAR sensor in zone — do not propose high-lux auto-shade rules unless the operator states they have no lux meter. Use manual policy or operator override.")
		}
		if pol == "auto" && !interlocks.HasTemp {
			lines = append(lines, "  ⚠ No temperature sensor in zone — high-temp fan and night-retract rules need temp_sensor_id when applying templates.")
		}
	}
	if !interlocks.HasLux {
		lines = append(lines, "  Sensor interlock: lux/PAR absent in zone.")
	}
	if len(actuators) == 0 {
		lines = append(lines, "  No linked actuators in profile.")
	} else {
		lines = append(lines, "  Actuators:")
		for _, a := range actuators {
			state := "unknown"
			if a.StateText != nil {
				state = *a.StateText
			}
			lines = append(lines, fmt.Sprintf("    [%s] %s (id %d, %s) — state: %s", a.Role, a.Name, a.ID, a.ActuatorType, state))
		}
	}
	if len(ghRules) == 0 {
		lines = append(lines, "  No GH automation rules found.")
	} else {
		activeCount := 0
		for _, r := range ghRules {
			if r.IsActive {
				activeCount++
			}
		}
		lines = append(lines, fmt.Sprintf("  Rules: %d GH rules (%d active)", len(ghRules), activeCount))
	}
	if len(recentEvents) > 0 {
		lines = append(lines, fmt.Sprintf("  Last %d shade/fan events (48 h):", len(recentEvents)))
		for _, e := range recentEvents {
			lines = append(lines, fmt.Sprintf("    %s — cmd=%q source=%s (actuator %d)",
				e.EventTime.Format("2006-01-02 15:04"), e.CommandSent, e.Source, e.ActuatorID))
		}
	} else {
		lines = append(lines, "  No shade/fan events in last 48 h.")
	}

	return map[string]any{
		"zone": map[string]any{
			"id":        zone.ID,
			"name":      zone.Name,
			"zone_type": zoneType,
		},
		"sensor_interlocks": interlocks,
		"profile":           profile,
		"profile_missing": profileMissing,
		"actuators":       actuators,
		"rules":           ghRules,
		"recent_events":   recentEvents,
		"summary":         strings.Join(lines, "\n"),
	}, nil
}

// jsonInt64 coerces an any (from JSON unmarshal) to int64.
// JSON numbers unmarshal as float64 by default.
func jsonInt64(v any) (int64, bool) {
	switch x := v.(type) {
	case float64:
		return int64(x), true
	case int64:
		return x, true
	case int:
		return int64(x), true
	}
	return 0, false
}

func orNA(s string) string {
	if s == "" {
		return "not set"
	}
	return s
}

type guardianZoneInterlocks struct {
	HasLux      bool `json:"has_lux_or_par"`
	HasTemp     bool `json:"has_temperature"`
	HasHumidity bool `json:"has_humidity"`
}

func zoneSensorInterlocksForGuardian(ctx context.Context, q db.Querier, zoneID int64) (guardianZoneInterlocks, error) {
	zid := zoneID
	sensors, err := q.ListSensorsByZone(ctx, &zid)
	if err != nil {
		return guardianZoneInterlocks{}, err
	}
	var st guardianZoneInterlocks
	for _, s := range sensors {
		if s.DeletedAt.Valid {
			continue
		}
		t := strings.ToLower(s.SensorType)
		if t == "lux" || strings.Contains(t, "lux") || t == "par" || strings.Contains(t, "par") {
			st.HasLux = true
		}
		if strings.Contains(t, "temp") {
			st.HasTemp = true
		}
		if strings.Contains(t, "humid") || t == "rh" {
			st.HasHumidity = true
		}
	}
	return st, nil
}
