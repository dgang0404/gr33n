package opstimeline

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "gr33n-api/internal/db"
)

// Event is one row on the crop ops timeline.
type Event struct {
	Kind       string         `json:"kind"`
	OccurredAt time.Time      `json:"occurred_at"`
	ID         int64          `json:"id"`
	Summary    string         `json:"summary,omitempty"`
	Details    map[string]any `json:"details,omitempty"`
}

// Result is the API / Guardian read-tool payload.
type Result struct {
	CropCycleID int64     `json:"crop_cycle_id"`
	FarmID      int64     `json:"farm_id"`
	ZoneID      int64     `json:"zone_id"`
	From        time.Time `json:"from"`
	To          time.Time `json:"to"`
	Events      []Event   `json:"events"`
}

// DefaultRange picks [from, to] from cycle dates when query params are omitted.
func DefaultRange(cycle db.Gr33nfertigationCropCycle, now time.Time) (time.Time, time.Time) {
	from := now.UTC().AddDate(0, -3, 0)
	if cycle.StartedAt.Valid {
		from = cycle.StartedAt.Time.UTC()
	}
	to := now.UTC()
	if cycle.HarvestedAt.Valid {
		end := cycle.HarvestedAt.Time.UTC().Add(24 * time.Hour)
		if end.After(to) {
			to = end
		}
	}
	return from, to
}

// ParseTimeQuery parses RFC3339 or YYYY-MM-DD; zero time means unset.
func ParseTimeQuery(raw string) (time.Time, error) {
	raw = trim(raw)
	if raw == "" {
		return time.Time{}, nil
	}
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t.UTC(), nil
	}
	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return time.Time{}, err
	}
	return t.UTC(), nil
}

func trim(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 {
		c := s[len(s)-1]
		if c != ' ' && c != '\t' {
			break
		}
		s = s[:len(s)-1]
	}
	return s
}

// Build loads and merges timeline events for a crop cycle within [from, to].
func Build(ctx context.Context, q db.Querier, cycle db.Gr33nfertigationCropCycle, from, to time.Time) (Result, error) {
	out := Result{
		CropCycleID: cycle.ID,
		FarmID:      cycle.FarmID,
		ZoneID:      cycle.ZoneID,
		From:        from,
		To:          to,
		Events:      []Event{},
	}

	stages, err := q.ListCropCycleStageEventsInRange(ctx, db.ListCropCycleStageEventsInRangeParams{
		CropCycleID: cycle.ID,
		FromTs:      from,
		ToTs:        to,
	})
	if err != nil {
		return out, err
	}
	for _, s := range stages {
		out.Events = append(out.Events, Event{
			Kind:       "stage",
			OccurredAt: s.EnteredAt.UTC(),
			ID:         s.ID,
			Summary:    string(s.GrowthStage),
			Details: map[string]any{
				"growth_stage": string(s.GrowthStage),
			},
		})
	}

	cycleID := cycle.ID
	applies, err := q.ListFertigationEventsForCropCycleInRange(ctx, db.ListFertigationEventsForCropCycleInRangeParams{
		FarmID:      cycle.FarmID,
		CropCycleID: &cycleID,
		FromTs:      from,
		ToTs:        to,
	})
	if err != nil {
		return out, err
	}
	for _, e := range applies {
		details := map[string]any{
			"zone_id": e.ZoneID,
		}
		if vol, ok := numericFloat(e.VolumeAppliedLiters); ok {
			details["volume_liters"] = vol
		}
		if ec, ok := numericFloat(e.EcAfterMscm); ok {
			details["ec_after_mscm"] = ec
		}
		if ph, ok := numericFloat(e.PhAfter); ok {
			details["ph_after"] = ph
		}
		if e.ProgramID != nil {
			details["program_id"] = *e.ProgramID
		}
		out.Events = append(out.Events, Event{
			Kind:       "apply",
			OccurredAt: e.AppliedAt.UTC(),
			ID:         e.ID,
			Summary:    "Fertigation apply",
			Details:    details,
		})
	}

	runs, err := q.ListProgramAutomationRunsForZoneInRange(ctx, db.ListProgramAutomationRunsForZoneInRangeParams{
		FarmID:  cycle.FarmID,
		ZoneID:  &cycle.ZoneID,
		FromTs:  from,
		ToTs:    to,
	})
	if err != nil {
		return out, err
	}
	for _, r := range runs {
		details := parseJSONDetails(r.Details)
		if r.ProgramName != "" {
			details["program_name"] = r.ProgramName
		}
		if r.ProgramID != nil {
			details["program_id"] = *r.ProgramID
		}
		details["status"] = r.Status
		out.Events = append(out.Events, Event{
			Kind:       "program_run",
			OccurredAt: r.ExecutedAt.UTC(),
			ID:         r.ID,
			Summary:    r.ProgramName,
			Details:    details,
		})
	}

	mixes, err := q.ListMixingEventsForZoneInRange(ctx, db.ListMixingEventsForZoneInRangeParams{
		FarmID: cycle.FarmID,
		ZoneID: &cycle.ZoneID,
		FromTs: from,
		ToTs:   to,
	})
	if err != nil {
		return out, err
	}
	for _, m := range mixes {
		components, err := q.ListMixingEventComponents(ctx, m.ID)
		if err != nil {
			return out, err
		}
		compRows := make([]map[string]any, 0, len(components))
		for _, c := range components {
			row := map[string]any{"input_definition_id": c.InputDefinitionID}
			if c.InputBatchID != nil {
				row["input_batch_id"] = *c.InputBatchID
			}
			if vol, ok := numericFloat(c.VolumeAddedMl); ok {
				row["volume_ml"] = vol
			}
			if c.DilutionRatio != nil {
				row["dilution_ratio"] = *c.DilutionRatio
			}
			compRows = append(compRows, row)
		}
		details := map[string]any{
			"reservoir_id": m.ReservoirID,
			"components":   compRows,
		}
		if m.ProgramID != nil {
			details["program_id"] = *m.ProgramID
		}
		if m.ProgramName != "" {
			details["program_name"] = m.ProgramName
		}
		if len(m.Metadata) > 0 {
			meta := parseJSONDetails(m.Metadata)
			for k, v := range meta {
				details[k] = v
			}
		}
		if vol, ok := numericFloat(m.WaterVolumeLiters); ok {
			details["water_volume_liters"] = vol
		}
		out.Events = append(out.Events, Event{
			Kind:       "mix",
			OccurredAt: m.MixedAt.UTC(),
			ID:         m.ID,
			Summary:    m.ProgramName,
			Details:    details,
		})
	}

	lights, err := q.ListLightingAutomationRunsForCropCycleInRange(ctx, db.ListLightingAutomationRunsForCropCycleInRangeParams{
		FarmID:      cycle.FarmID,
		CropCycleID: &cycleID,
		FromTs:      from,
		ToTs:        to,
	})
	if err != nil {
		return out, err
	}
	for _, l := range lights {
		details := parseJSONDetails(l.Details)
		details["lighting_program_id"] = l.LightingProgramID
		details["lighting_program_name"] = l.LightingProgramName
		details["on_hours"] = l.OnHours
		details["off_hours"] = l.OffHours
		details["lights_on_at"] = l.LightsOnAt
		details["status"] = l.Status
		if l.ScheduleID != nil {
			details["schedule_id"] = *l.ScheduleID
		}
		out.Events = append(out.Events, Event{
			Kind:       "light",
			OccurredAt: l.ExecutedAt.UTC(),
			ID:         l.ID,
			Summary:    l.LightingProgramName,
			Details:    details,
		})
	}

	sort.Slice(out.Events, func(i, j int) bool {
		if out.Events[i].OccurredAt.Equal(out.Events[j].OccurredAt) {
			if out.Events[i].Kind == out.Events[j].Kind {
				return out.Events[i].ID < out.Events[j].ID
			}
			return out.Events[i].Kind < out.Events[j].Kind
		}
		return out.Events[i].OccurredAt.Before(out.Events[j].OccurredAt)
	})
	return out, nil
}

func parseJSONDetails(raw json.RawMessage) map[string]any {
	if len(raw) == 0 {
		return map[string]any{}
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil || m == nil {
		return map[string]any{}
	}
	return m
}

func numericFloat(n pgtype.Numeric) (float64, bool) {
	f, err := n.Float64Value()
	if err != nil || !f.Valid {
		return 0, false
	}
	return f.Float64, true
}

