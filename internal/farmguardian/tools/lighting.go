package tools

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	db "gr33n-api/internal/db"
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

func execCreateLightingProgram(ctx context.Context, deps ExecutorDeps, args map[string]any) (any, error) {
	if deps.FarmID <= 0 {
		return nil, errors.New("farm_id required in proposal scope")
	}
	presetKey, err := stringFromArgs(args, "preset_key")
	if err != nil {
		return nil, err
	}
	zoneID, err := int64FromArgs(args, "zone_id")
	if err != nil {
		return nil, err
	}
	actuatorID, err := int64FromArgs(args, "actuator_id")
	if err != nil {
		return nil, err
	}
	z, err := deps.Q.GetZoneByID(ctx, zoneID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("zone %d not found", zoneID)
		}
		return nil, err
	}
	if err := ensureFarmScope(z.FarmID, deps.FarmID); err != nil {
		return nil, err
	}
	act, err := deps.Q.GetActuatorByID(ctx, actuatorID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("actuator %d not found", actuatorID)
		}
		return nil, err
	}
	if err := ensureFarmScope(act.FarmID, deps.FarmID); err != nil {
		return nil, err
	}
	pool, ok := deps.Pool.(*pgxpool.Pool)
	if !ok || pool == nil {
		return nil, errors.New("create_lighting_program requires database pool")
	}
	baseQ, ok := deps.Q.(*db.Queries)
	if !ok {
		return nil, errors.New("create_lighting_program requires database queries")
	}

	var name *string
	if n, err := optionalStringFromArgs(args, "name"); err != nil {
		return nil, err
	} else if n != nil {
		name = n
	}
	lightsOnAt, _ := optionalStringFromArgs(args, "lights_on_at")
	tz, _ := optionalStringFromArgs(args, "timezone")
	var cropCycleID *int64
	if v, ok := args["crop_cycle_id"]; ok && v != nil {
		cid, err := int64FromArgs(args, "crop_cycle_id")
		if err != nil {
			return nil, err
		}
		cropCycleID = &cid
	}
	in := lightinghandler.FromPresetInput{
		PresetKey:   presetKey,
		Name:        name,
		ZoneID:      zoneID,
		ActuatorID:  actuatorID,
		CropCycleID: cropCycleID,
	}
	if lightsOnAt != nil {
		in.LightsOnAt = *lightsOnAt
	}
	if tz != nil {
		in.Timezone = *tz
	}

	prog, err := lightinghandler.CreateProgramFromPreset(ctx, pool, baseQ, deps.FarmID, in)
	if err != nil {
		return nil, err
	}
	offAt := computeOffTime(prog.LightsOnAt, prog.OnHours)
	return map[string]any{
		"lighting_program_id": prog.ID,
		"name":                prog.Name,
		"zone_id":             prog.ZoneID,
		"actuator_id":         prog.ActuatorID,
		"preset_key":          presetKey,
		"photoperiod":         fmt.Sprintf("%dh/%dh ON/OFF — lights on at %s, off at %s (%s)", prog.OnHours, prog.OffHours, prog.LightsOnAt, offAt, prog.Timezone),
		"schedule_on_id":      prog.ScheduleOnID,
		"schedule_off_id":     prog.ScheduleOffID,
	}, nil
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
