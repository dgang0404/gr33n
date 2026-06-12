package cropcycle

import (
	"strings"

	"github.com/jackc/pgx/v5/pgtype"

	db "gr33n-api/internal/db"
)

// ResolveBatchLabel accepts batch_label (primary) or deprecated strain_or_variety.
func ResolveBatchLabel(primary, deprecated *string) *string {
	if primary != nil {
		v := strings.TrimSpace(*primary)
		if v != "" {
			return &v
		}
	}
	if deprecated != nil {
		v := strings.TrimSpace(*deprecated)
		if v != "" {
			return &v
		}
	}
	return nil
}

// CycleJSON returns API JSON with batch_label primary and strain_or_variety alias.
func CycleJSON(c db.Gr33nfertigationCropCycle) map[string]any {
	m := map[string]any{
		"id":           c.ID,
		"farm_id":      c.FarmID,
		"zone_id":      c.ZoneID,
		"name":         c.Name,
		"current_stage": stageString(c.CurrentStage),
		"is_active":    c.IsActive,
		"started_at":   formatDate(c.StartedAt),
		"created_at":   c.CreatedAt,
		"updated_at":   c.UpdatedAt,
	}
	if c.BatchLabel != nil {
		m["batch_label"] = *c.BatchLabel
		m["strain_or_variety"] = *c.BatchLabel
	}
	if c.HarvestedAt.Valid {
		m["harvested_at"] = c.HarvestedAt.Time.Format("2006-01-02")
	}
	if c.YieldGrams.Valid {
		m["yield_grams"] = numericFloat(c.YieldGrams)
	}
	if c.YieldNotes != nil {
		m["yield_notes"] = *c.YieldNotes
	}
	if c.CycleNotes != nil {
		m["cycle_notes"] = *c.CycleNotes
	}
	if c.PrimaryProgramID != nil {
		m["primary_program_id"] = *c.PrimaryProgramID
	}
	if c.PlantID != nil {
		m["plant_id"] = *c.PlantID
	}
	return m
}

func stageString(s *db.Gr33nfertigationGrowthStageEnum) any {
	if s == nil {
		return nil
	}
	return string(*s)
}

func formatDate(d pgtype.Date) string {
	if !d.Valid {
		return ""
	}
	return d.Time.Format("2006-01-02")
}

func numericFloat(n pgtype.Numeric) float64 {
	if !n.Valid {
		return 0
	}
	f, err := n.Float64Value()
	if err != nil || !f.Valid {
		return 0
	}
	return f.Float64
}

// BatchLabelFromArgs resolves batch_label or strain_or_variety from tool/proposal maps.
func BatchLabelFromArgs(args map[string]any) *string {
	if v, ok := args["batch_label"]; ok && v != nil {
		if s, ok := v.(string); ok {
			return ResolveBatchLabel(&s, nil)
		}
	}
	if v, ok := args["strain_or_variety"]; ok && v != nil {
		if s, ok := v.(string); ok {
			return ResolveBatchLabel(nil, &s)
		}
	}
	return nil
}
