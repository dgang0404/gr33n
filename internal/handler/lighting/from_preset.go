package lighting

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	db "gr33n-api/internal/db"
)

// FromPresetInput is the shared payload for preset-based lighting program creation.
type FromPresetInput struct {
	PresetKey   string
	Name        *string
	ZoneID      int64
	ActuatorID  int64
	LightsOnAt  string
	Timezone    string
	CropCycleID *int64
}

// LookupPreset returns the preset definition or false.
func LookupPreset(key string) (presetDef, bool) {
	p, ok := presets[key]
	return p, ok
}

// PresetKeys returns valid preset_key values.
func PresetKeys() []string {
	keys := make([]string, 0, len(presets))
	for k := range presets {
		keys = append(keys, k)
	}
	return keys
}

// CreateProgramFromPreset creates a lighting program with ON/OFF schedules and actions.
func CreateProgramFromPreset(ctx context.Context, pool *pgxpool.Pool, q *db.Queries, farmID int64, in FromPresetInput) (db.Gr33ncoreLightingProgram, error) {
	p, ok := presets[in.PresetKey]
	if !ok {
		return db.Gr33ncoreLightingProgram{}, fmt.Errorf("unknown preset_key %q; available: %s",
			in.PresetKey, strings.Join(PresetKeys(), ", "))
	}
	name := p.Name
	if in.Name != nil && strings.TrimSpace(*in.Name) != "" {
		name = strings.TrimSpace(*in.Name)
	}
	lightsOnAt := in.LightsOnAt
	if lightsOnAt == "" {
		lightsOnAt = "06:00"
	}
	tz := in.Timezone
	if tz == "" {
		if farm, ferr := q.GetFarmByID(ctx, farmID); ferr == nil && farm.Timezone != "" {
			tz = farm.Timezone
		} else {
			tz = "UTC"
		}
	}
	if _, err := time.LoadLocation(tz); err != nil {
		return db.Gr33ncoreLightingProgram{}, fmt.Errorf("invalid timezone %q", tz)
	}

	meta, _ := json.Marshal(map[string]string{"preset_key": in.PresetKey})
	req := createProgramRequest{
		Name:        name,
		ZoneID:      in.ZoneID,
		ActuatorID:  in.ActuatorID,
		OnHours:     p.OnHours,
		OffHours:    p.OffHours,
		LightsOnAt:  lightsOnAt,
		Timezone:    tz,
		CropCycleID: in.CropCycleID,
		IsActive:    true,
		Metadata:    json.RawMessage(meta),
	}
	if err := req.validate(); err != nil {
		return db.Gr33ncoreLightingProgram{}, err
	}

	onCron, offCron, err := buildCronExpressions(req.LightsOnAt, req.OnHours)
	if err != nil {
		return db.Gr33ncoreLightingProgram{}, err
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return db.Gr33ncoreLightingProgram{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	qtx := q.WithTx(tx)
	prog, err := qtx.CreateLightingProgram(ctx, db.CreateLightingProgramParams{
		FarmID:        farmID,
		ZoneID:        req.ZoneID,
		ActuatorID:    req.ActuatorID,
		Name:          req.Name,
		Description:   req.Description,
		OnHours:       req.OnHours,
		OffHours:      req.OffHours,
		LightsOnAt:    req.LightsOnAt,
		Timezone:      req.Timezone,
		CropCycleID:   req.CropCycleID,
		IsActive:      req.IsActive,
		Metadata:      req.Metadata,
		ScheduleOnID:  nil,
		ScheduleOffID: nil,
	})
	if err != nil {
		return db.Gr33ncoreLightingProgram{}, fmt.Errorf("create lighting program: %w", err)
	}
	onID, offID, err := materializeSchedules(ctx, qtx, prog, onCron, offCron)
	if err != nil {
		return db.Gr33ncoreLightingProgram{}, err
	}
	if err := createScheduleActions(ctx, qtx, req.ActuatorID, onID, offID); err != nil {
		return db.Gr33ncoreLightingProgram{}, err
	}
	prog, err = qtx.UpdateLightingProgramSchedules(ctx, db.UpdateLightingProgramSchedulesParams{
		ID:            prog.ID,
		ScheduleOnID:  &onID,
		ScheduleOffID: &offID,
	})
	if err != nil {
		return db.Gr33ncoreLightingProgram{}, fmt.Errorf("link schedules: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return db.Gr33ncoreLightingProgram{}, fmt.Errorf("commit: %w", err)
	}
	return prog, nil
}
