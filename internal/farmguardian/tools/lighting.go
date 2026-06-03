package tools

import (
	"context"
	"fmt"
	"strings"

	lightinghandler "gr33n-api/internal/handler/lighting"
)

// execSummarizeZoneLighting — read tool that returns active lighting programs
// for a farm (optionally filtered by zone), including photoperiod, next
// trigger hint, and actuator reference.
func execSummarizeZoneLighting(ctx context.Context, deps ExecutorDeps, args map[string]any) (any, error) {
	programs, err := deps.Q.ListLightingProgramsByFarm(ctx, deps.FarmID)
	if err != nil {
		return nil, fmt.Errorf("listing lighting programs: %w", err)
	}

	// Optional zone filter.
	var zoneID *int64
	if v, ok := args["zone_id"]; ok && v != nil {
		zid, err := int64FromArgs(args, "zone_id")
		if err != nil {
			return nil, err
		}
		zoneID = &zid
	}

	type programSummary struct {
		ID          int64   `json:"id"`
		Name        string  `json:"name"`
		ZoneID      int64   `json:"zone_id"`
		ActuatorID  int64   `json:"actuator_id"`
		OnHours     int32   `json:"on_hours"`
		OffHours    int32   `json:"off_hours"`
		LightsOnAt  string  `json:"lights_on_at"`
		LightsOffAt string  `json:"lights_off_at"`
		Timezone    string  `json:"timezone"`
		IsActive    bool    `json:"is_active"`
		Photoperiod string  `json:"photoperiod"`
		ScheduleOn  *int64  `json:"schedule_on_id,omitempty"`
		ScheduleOff *int64  `json:"schedule_off_id,omitempty"`
	}

	out := make([]programSummary, 0, len(programs))
	for _, p := range programs {
		if zoneID != nil && p.ZoneID != *zoneID {
			continue
		}
		offAt := computeOffTime(p.LightsOnAt, p.OnHours)
		out = append(out, programSummary{
			ID:          p.ID,
			Name:        p.Name,
			ZoneID:      p.ZoneID,
			ActuatorID:  p.ActuatorID,
			OnHours:     p.OnHours,
			OffHours:    p.OffHours,
			LightsOnAt:  p.LightsOnAt,
			LightsOffAt: offAt,
			Timezone:    p.Timezone,
			IsActive:    p.IsActive,
			Photoperiod: fmt.Sprintf("%dh/%dh ON/OFF — lights on at %s, off at %s (%s)", p.OnHours, p.OffHours, p.LightsOnAt, offAt, p.Timezone),
			ScheduleOn:  p.ScheduleOnID,
			ScheduleOff: p.ScheduleOffID,
		})
	}

	if len(out) == 0 {
		return map[string]any{
			"programs": []any{},
			"summary":  "No lighting programs configured for this farm.",
		}, nil
	}

	active := 0
	for _, p := range out {
		if p.IsActive {
			active++
		}
	}

	lines := make([]string, 0, len(out))
	for _, p := range out {
		status := "inactive"
		if p.IsActive {
			status = "active"
		}
		lines = append(lines, fmt.Sprintf("• %s (zone %d): %s [%s]", p.Name, p.ZoneID, p.Photoperiod, status))
	}

	return map[string]any{
		"programs": out,
		"count":    len(out),
		"active":   active,
		"summary":  fmt.Sprintf("%d lighting program(s), %d active.\n%s", len(out), active, strings.Join(lines, "\n")),
		"presets":  lightinghandler.PresetList(),
	}, nil
}

// computeOffTime derives the lights-off time string given a lights-on anchor and on_hours.
func computeOffTime(lightsOnAt string, onHours int32) string {
	parts := strings.SplitN(lightsOnAt, ":", 2)
	if len(parts) != 2 {
		return "unknown"
	}
	h, m := parseInt(parts[0]), parseInt(parts[1])
	totalMins := (h*60 + m + int(onHours)*60) % (24 * 60)
	return fmt.Sprintf("%02d:%02d", totalMins/60, totalMins%60)
}

func parseInt(s string) int {
	v := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return v
		}
		v = v*10 + int(c-'0')
	}
	return v
}
