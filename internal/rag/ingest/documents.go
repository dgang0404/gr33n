package ingest

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/rag/sanitize"
)

const (
	SourceTypeTask                = "task"
	SourceTypeAutomationRun       = "automation_run"
	SourceTypeCropCycle           = "crop_cycle"
	SourceTypeFertigationProgram  = "fertigation_program"
	metadataModuleCore           = "core"
	metadataModuleAutomation     = "automation"
	metadataModuleFertigation    = "fertigation"
)

// TaskDocument builds deterministic embed text from a task row (single chunk).
func TaskDocument(t db.Gr33ncoreTask) string {
	var b strings.Builder
	b.WriteString("task: ")
	b.WriteString(strings.TrimSpace(t.Title))
	b.WriteByte('\n')
	b.WriteString("status: ")
	b.WriteString(string(t.Status))
	b.WriteByte('\n')
	if t.Description != nil && strings.TrimSpace(*t.Description) != "" {
		b.WriteString(sanitize.PlainNotes(*t.Description, 12000))
		b.WriteByte('\n')
	}
	if t.TaskType != nil && *t.TaskType != "" {
		b.WriteString("task_type: ")
		b.WriteString(strings.TrimSpace(*t.TaskType))
		b.WriteByte('\n')
	}
	if t.RelatedModuleSchema != nil && *t.RelatedModuleSchema != "" {
		b.WriteString("related: ")
		b.WriteString(strings.TrimSpace(*t.RelatedModuleSchema))
		if t.RelatedTableName != nil {
			b.WriteByte('.')
			b.WriteString(strings.TrimSpace(*t.RelatedTableName))
		}
		if t.RelatedRecordID != nil {
			b.WriteByte('#')
			b.WriteString(strconv.FormatInt(*t.RelatedRecordID, 10))
		}
		b.WriteByte('\n')
	}
	s := strings.TrimSpace(b.String())
	return s
}

// AutomationRunDocument combines status, message, and sanitized details JSON.
func AutomationRunDocument(run db.Gr33ncoreAutomationRun) string {
	var b strings.Builder
	b.WriteString("automation_run ")
	b.WriteString(run.Status)
	b.WriteByte('\n')
	if run.Message != nil && strings.TrimSpace(*run.Message) != "" {
		b.WriteString(strings.TrimSpace(*run.Message))
		b.WriteByte('\n')
	}
	if txt := sanitize.AutomationDetailsJSON(run.Details); txt != "" {
		b.WriteString(txt)
	}
	return strings.TrimSpace(b.String())
}

// CropCycleDocument builds deterministic embed text from a crop_cycles row (single chunk).
func CropCycleDocument(c db.Gr33nfertigationCropCycle) string {
	var b strings.Builder
	b.WriteString("crop_cycle: ")
	b.WriteString(strings.TrimSpace(c.Name))
	b.WriteByte('\n')
	b.WriteString("zone_id: ")
	b.WriteString(strconv.FormatInt(c.ZoneID, 10))
	b.WriteByte('\n')
	if c.StrainOrVariety != nil && strings.TrimSpace(*c.StrainOrVariety) != "" {
		b.WriteString("strain_or_variety: ")
		b.WriteString(strings.TrimSpace(*c.StrainOrVariety))
		b.WriteByte('\n')
	}
	if c.CurrentStage.Valid {
		b.WriteString("stage: ")
		b.WriteString(string(c.CurrentStage.Gr33nfertigationGrowthStageEnum))
		b.WriteByte('\n')
	}
	if c.IsActive {
		b.WriteString("active: yes\n")
	} else {
		b.WriteString("active: no\n")
	}
	if s := formatPGDate(c.StartedAt); s != "" {
		b.WriteString("started_at: ")
		b.WriteString(s)
		b.WriteByte('\n')
	}
	if s := formatPGDate(c.HarvestedAt); s != "" {
		b.WriteString("harvested_at: ")
		b.WriteString(s)
		b.WriteByte('\n')
	}
	if f, err := c.YieldGrams.Float64Value(); err == nil && f.Valid {
		fmt.Fprintf(&b, "yield_grams: %g\n", f.Float64)
	}
	if c.YieldNotes != nil && strings.TrimSpace(*c.YieldNotes) != "" {
		b.WriteString("yield_notes: ")
		b.WriteString(sanitize.PlainNotes(*c.YieldNotes, 8000))
		b.WriteByte('\n')
	}
	if c.CycleNotes != nil && strings.TrimSpace(*c.CycleNotes) != "" {
		b.WriteString("cycle_notes: ")
		b.WriteString(sanitize.PlainNotes(*c.CycleNotes, 8000))
		b.WriteByte('\n')
	}
	if c.PrimaryProgramID != nil && *c.PrimaryProgramID > 0 {
		b.WriteString("primary_program_id: ")
		b.WriteString(strconv.FormatInt(*c.PrimaryProgramID, 10))
		b.WriteByte('\n')
	}
	return strings.TrimSpace(b.String())
}

func formatPGDate(d pgtype.Date) string {
	if !d.Valid {
		return ""
	}
	return d.Time.Format("2006-01-02")
}

// FertigationProgramDocument builds embed text from gr33nfertigation.programs (single chunk).
func FertigationProgramDocument(p db.Gr33nfertigationProgram) string {
	var b strings.Builder
	b.WriteString("fertigation_program: ")
	b.WriteString(strings.TrimSpace(p.Name))
	b.WriteByte('\n')
	if p.Description != nil && strings.TrimSpace(*p.Description) != "" {
		b.WriteString(sanitize.PlainNotes(*p.Description, 12000))
		b.WriteByte('\n')
	}
	if p.IsActive {
		b.WriteString("active: yes\n")
	} else {
		b.WriteString("active: no\n")
	}
	writeOptionalID(&b, "application_recipe_id", p.ApplicationRecipeID)
	writeOptionalID(&b, "reservoir_id", p.ReservoirID)
	writeOptionalID(&b, "target_zone_id", p.TargetZoneID)
	writeOptionalID(&b, "schedule_id", p.ScheduleID)
	writeOptionalID(&b, "ec_target_id", p.EcTargetID)
	writeNumericLine(&b, "volume_liters_per_sqm", p.VolumeLitersPerSqm)
	writeNumericLine(&b, "total_volume_liters", p.TotalVolumeLiters)
	if p.DilutionRatio != nil && strings.TrimSpace(*p.DilutionRatio) != "" {
		b.WriteString("dilution_ratio: ")
		b.WriteString(strings.TrimSpace(*p.DilutionRatio))
		b.WriteByte('\n')
	}
	if p.RunDurationSeconds != nil && *p.RunDurationSeconds > 0 {
		b.WriteString("run_duration_seconds: ")
		b.WriteString(strconv.FormatInt(int64(*p.RunDurationSeconds), 10))
		b.WriteByte('\n')
	}
	writeNumericLine(&b, "ec_trigger_low", p.EcTriggerLow)
	writeNumericLine(&b, "ph_trigger_low", p.PhTriggerLow)
	writeNumericLine(&b, "ph_trigger_high", p.PhTriggerHigh)
	if txt := sanitize.FertigationProgramMetadataForEmbed(p.Metadata); txt != "" {
		b.WriteString("metadata:\n")
		b.WriteString(txt)
		b.WriteByte('\n')
	}
	return strings.TrimSpace(b.String())
}

func writeOptionalID(b *strings.Builder, label string, id *int64) {
	if id == nil || *id <= 0 {
		return
	}
	b.WriteString(label)
	b.WriteString(": ")
	b.WriteString(strconv.FormatInt(*id, 10))
	b.WriteByte('\n')
}

func writeNumericLine(b *strings.Builder, label string, n pgtype.Numeric) {
	if !n.Valid {
		return
	}
	f, err := n.Float64Value()
	if err != nil || !f.Valid {
		return
	}
	fmt.Fprintf(b, "%s: %g\n", label, f.Float64)
}
