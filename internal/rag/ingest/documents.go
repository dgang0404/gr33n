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
	SourceTypeSchedule         = "schedule"
	SourceTypeAutomationRule   = "automation_rule"
	SourceTypeExecutableAction = "executable_action"
	SourceTypeCostTransaction    = "cost_transaction"
	SourceTypeInputDefinition    = "input_definition"
	SourceTypeInputBatch         = "input_batch"
	SourceTypeAlertNotification  = "alert_notification"
	metadataModuleCore           = "core"
	metadataModuleAutomation     = "automation"
	metadataModuleFertigation    = "fertigation"
	metadataModuleCost           = "cost"
	metadataModuleInventory      = "inventory"
	metadataModuleAlerts         = "alerts"
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

func writeDateLine(b *strings.Builder, label string, d pgtype.Date) {
	if s := formatPGDate(d); s != "" {
		b.WriteString(label)
		b.WriteString(": ")
		b.WriteString(s)
		b.WriteByte('\n')
	}
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

// ScheduleDocument builds embed text from gr33ncore.schedules (single chunk).
func ScheduleDocument(s db.Gr33ncoreSchedule) string {
	var b strings.Builder
	b.WriteString("schedule: ")
	b.WriteString(strings.TrimSpace(s.Name))
	b.WriteByte('\n')
	if s.Description != nil && strings.TrimSpace(*s.Description) != "" {
		b.WriteString(sanitize.PlainNotes(*s.Description, 12000))
		b.WriteByte('\n')
	}
	b.WriteString("schedule_type: ")
	b.WriteString(strings.TrimSpace(s.ScheduleType))
	b.WriteByte('\n')
	b.WriteString("cron_expression: ")
	b.WriteString(strings.TrimSpace(s.CronExpression))
	b.WriteByte('\n')
	b.WriteString("timezone: ")
	b.WriteString(strings.TrimSpace(s.Timezone))
	b.WriteByte('\n')
	if s.IsActive {
		b.WriteString("active: yes\n")
	} else {
		b.WriteString("active: no\n")
	}
	if txt := sanitize.AutomationDetailsJSON(s.MetaData); txt != "" {
		b.WriteString("meta_data:\n")
		b.WriteString(txt)
		b.WriteByte('\n')
	}
	if txt := sanitize.AutomationDetailsJSON(s.Preconditions); txt != "" {
		b.WriteString("preconditions:\n")
		b.WriteString(txt)
		b.WriteByte('\n')
	}
	return strings.TrimSpace(b.String())
}

// AutomationRuleDocument builds embed text from gr33ncore.automation_rules (single chunk).
func AutomationRuleDocument(r db.Gr33ncoreAutomationRule) string {
	var b strings.Builder
	b.WriteString("automation_rule: ")
	b.WriteString(strings.TrimSpace(r.Name))
	b.WriteByte('\n')
	if r.Description != nil && strings.TrimSpace(*r.Description) != "" {
		b.WriteString(sanitize.PlainNotes(*r.Description, 12000))
		b.WriteByte('\n')
	}
	if r.IsActive {
		b.WriteString("active: yes\n")
	} else {
		b.WriteString("active: no\n")
	}
	b.WriteString("trigger_source: ")
	b.WriteString(string(r.TriggerSource))
	b.WriteByte('\n')
	if r.ConditionLogic != nil && strings.TrimSpace(*r.ConditionLogic) != "" {
		b.WriteString("condition_logic: ")
		b.WriteString(strings.TrimSpace(*r.ConditionLogic))
		b.WriteByte('\n')
	}
	if r.CooldownPeriodSeconds != nil && *r.CooldownPeriodSeconds > 0 {
		b.WriteString("cooldown_period_seconds: ")
		b.WriteString(strconv.FormatInt(int64(*r.CooldownPeriodSeconds), 10))
		b.WriteByte('\n')
	}
	if txt := sanitize.AutomationDetailsJSON(r.TriggerConfiguration); txt != "" {
		b.WriteString("trigger_configuration:\n")
		b.WriteString(txt)
		b.WriteByte('\n')
	}
	if txt := sanitize.AutomationDetailsJSON(r.ConditionsJsonb); txt != "" {
		b.WriteString("conditions:\n")
		b.WriteString(txt)
		b.WriteByte('\n')
	}
	return strings.TrimSpace(b.String())
}

// ExecutableActionDocument builds embed text from executable_actions (labels + scrubbed JSON only).
func ExecutableActionDocument(a db.Gr33ncoreExecutableAction) string {
	var b strings.Builder
	b.WriteString("executable_action\n")
	switch {
	case a.ScheduleID != nil:
		fmt.Fprintf(&b, "parent: schedule_id %d\n", *a.ScheduleID)
	case a.RuleID != nil:
		fmt.Fprintf(&b, "parent: automation_rule_id %d\n", *a.RuleID)
	case a.ProgramID != nil:
		fmt.Fprintf(&b, "parent: fertigation_program_id %d\n", *a.ProgramID)
	}
	fmt.Fprintf(&b, "execution_order: %d\n", a.ExecutionOrder)
	b.WriteString("action_type: ")
	b.WriteString(string(a.ActionType))
	b.WriteByte('\n')
	writeOptionalID(&b, "target_actuator_id", a.TargetActuatorID)
	writeOptionalID(&b, "target_automation_rule_id", a.TargetAutomationRuleID)
	writeOptionalID(&b, "target_notification_template_id", a.TargetNotificationTemplateID)
	if a.ActionCommand != nil && strings.TrimSpace(*a.ActionCommand) != "" {
		b.WriteString("action_command: ")
		b.WriteString(strings.TrimSpace(*a.ActionCommand))
		b.WriteByte('\n')
	}
	if a.DelayBeforeExecutionSeconds != nil && *a.DelayBeforeExecutionSeconds > 0 {
		b.WriteString("delay_before_execution_seconds: ")
		b.WriteString(strconv.FormatInt(int64(*a.DelayBeforeExecutionSeconds), 10))
		b.WriteByte('\n')
	}
	if txt := sanitize.AutomationDetailsJSON(a.ActionParameters); txt != "" {
		b.WriteString("action_parameters:\n")
		b.WriteString(txt)
		b.WriteByte('\n')
	}
	return strings.TrimSpace(b.String())
}

// CostTransactionDocument builds embed text from cost_transactions without monetary amounts or currency
// (§4.2 commercial sensitivity — narrative retrieval only).
func CostTransactionDocument(ct db.Gr33ncoreCostTransaction) string {
	var b strings.Builder
	b.WriteString("cost_transaction\n")
	if d := formatPGDate(ct.TransactionDate); d != "" {
		b.WriteString("transaction_date: ")
		b.WriteString(d)
		b.WriteByte('\n')
	}
	b.WriteString("category: ")
	b.WriteString(string(ct.Category))
	b.WriteByte('\n')
	if ct.Subcategory != nil && strings.TrimSpace(*ct.Subcategory) != "" {
		b.WriteString("subcategory: ")
		b.WriteString(strings.TrimSpace(*ct.Subcategory))
		b.WriteByte('\n')
	}
	if ct.IsIncome {
		b.WriteString("direction: income\n")
	} else {
		b.WriteString("direction: expense\n")
	}
	if ct.Description != nil && strings.TrimSpace(*ct.Description) != "" {
		b.WriteString("description: ")
		b.WriteString(sanitize.PlainNotes(*ct.Description, 8000))
		b.WriteByte('\n')
	}
	if ct.DocumentType != nil && strings.TrimSpace(*ct.DocumentType) != "" {
		b.WriteString("document_type: ")
		b.WriteString(strings.TrimSpace(*ct.DocumentType))
		b.WriteByte('\n')
	}
	if ct.DocumentReference != nil && strings.TrimSpace(*ct.DocumentReference) != "" {
		b.WriteString("document_reference: ")
		b.WriteString(sanitize.PlainNotes(*ct.DocumentReference, 2000))
		b.WriteByte('\n')
	}
	if ct.Counterparty != nil && strings.TrimSpace(*ct.Counterparty) != "" {
		b.WriteString("counterparty: ")
		b.WriteString(sanitize.PlainNotes(*ct.Counterparty, 500))
		b.WriteByte('\n')
	}
	if ct.RelatedModuleSchema != nil && strings.TrimSpace(*ct.RelatedModuleSchema) != "" {
		b.WriteString("related: ")
		b.WriteString(strings.TrimSpace(*ct.RelatedModuleSchema))
		if ct.RelatedTableName != nil {
			b.WriteByte('.')
			b.WriteString(strings.TrimSpace(*ct.RelatedTableName))
		}
		if ct.RelatedRecordID != nil {
			b.WriteByte('#')
			b.WriteString(strconv.FormatInt(*ct.RelatedRecordID, 10))
		}
		b.WriteByte('\n')
	}
	if ct.CropCycleID != nil && *ct.CropCycleID > 0 {
		b.WriteString("crop_cycle_id: ")
		b.WriteString(strconv.FormatInt(*ct.CropCycleID, 10))
		b.WriteByte('\n')
	}
	return strings.TrimSpace(b.String())
}

// InputDefinitionDocument embeds catalog text without unit cost / currency (§4.2).
func InputDefinitionDocument(d db.Gr33nnaturalfarmingInputDefinition) string {
	var b strings.Builder
	b.WriteString("input_definition: ")
	b.WriteString(strings.TrimSpace(d.Name))
	b.WriteByte('\n')
	b.WriteString("category: ")
	b.WriteString(string(d.Category))
	b.WriteByte('\n')
	if d.Description != nil && strings.TrimSpace(*d.Description) != "" {
		b.WriteString(sanitize.PlainNotes(*d.Description, 12000))
		b.WriteByte('\n')
	}
	if d.TypicalIngredients != nil && strings.TrimSpace(*d.TypicalIngredients) != "" {
		b.WriteString("typical_ingredients: ")
		b.WriteString(sanitize.PlainNotes(*d.TypicalIngredients, 8000))
		b.WriteByte('\n')
	}
	if d.PreparationSummary != nil && strings.TrimSpace(*d.PreparationSummary) != "" {
		b.WriteString("preparation_summary: ")
		b.WriteString(sanitize.PlainNotes(*d.PreparationSummary, 8000))
		b.WriteByte('\n')
	}
	if d.StorageGuidelines != nil && strings.TrimSpace(*d.StorageGuidelines) != "" {
		b.WriteString("storage_guidelines: ")
		b.WriteString(sanitize.PlainNotes(*d.StorageGuidelines, 8000))
		b.WriteByte('\n')
	}
	if d.SafetyPrecautions != nil && strings.TrimSpace(*d.SafetyPrecautions) != "" {
		b.WriteString("safety_precautions: ")
		b.WriteString(sanitize.PlainNotes(*d.SafetyPrecautions, 8000))
		b.WriteByte('\n')
	}
	if d.ReferenceSource != nil && strings.TrimSpace(*d.ReferenceSource) != "" {
		b.WriteString("reference_source: ")
		b.WriteString(sanitize.PlainNotes(*d.ReferenceSource, 4000))
		b.WriteByte('\n')
	}
	return strings.TrimSpace(b.String())
}

// InputBatchDocument embeds batch narrative without quantity or commercial numerics (§4.2).
func InputBatchDocument(batch db.Gr33nnaturalfarmingInputBatch) string {
	var b strings.Builder
	b.WriteString("input_batch\n")
	b.WriteString("input_definition_id: ")
	b.WriteString(strconv.FormatInt(batch.InputDefinitionID, 10))
	b.WriteByte('\n')
	if batch.BatchIdentifier != nil && strings.TrimSpace(*batch.BatchIdentifier) != "" {
		b.WriteString("batch_identifier: ")
		b.WriteString(strings.TrimSpace(*batch.BatchIdentifier))
		b.WriteByte('\n')
	}
	writeDateLine(&b, "creation_start_date", batch.CreationStartDate)
	writeDateLine(&b, "creation_end_date", batch.CreationEndDate)
	writeDateLine(&b, "expected_ready_date", batch.ExpectedReadyDate)
	writeDateLine(&b, "actual_ready_date", batch.ActualReadyDate)
	b.WriteString("status: ")
	b.WriteString(string(batch.Status))
	b.WriteByte('\n')
	if batch.StorageLocation != nil && strings.TrimSpace(*batch.StorageLocation) != "" {
		b.WriteString("storage_location: ")
		b.WriteString(strings.TrimSpace(*batch.StorageLocation))
		b.WriteByte('\n')
	}
	if batch.ShelfLifeDays != nil && *batch.ShelfLifeDays > 0 {
		b.WriteString("shelf_life_days: ")
		b.WriteString(strconv.FormatInt(int64(*batch.ShelfLifeDays), 10))
		b.WriteByte('\n')
	}
	if batch.TemperatureDuringMaking != nil && strings.TrimSpace(*batch.TemperatureDuringMaking) != "" {
		b.WriteString("temperature_during_making: ")
		b.WriteString(strings.TrimSpace(*batch.TemperatureDuringMaking))
		b.WriteByte('\n')
	}
	if batch.IngredientsUsed != nil && strings.TrimSpace(*batch.IngredientsUsed) != "" {
		b.WriteString("ingredients_used: ")
		b.WriteString(sanitize.PlainNotes(*batch.IngredientsUsed, 8000))
		b.WriteByte('\n')
	}
	if batch.ProcedureFollowed != nil && strings.TrimSpace(*batch.ProcedureFollowed) != "" {
		b.WriteString("procedure_followed: ")
		b.WriteString(sanitize.PlainNotes(*batch.ProcedureFollowed, 8000))
		b.WriteByte('\n')
	}
	if batch.ObservationsNotes != nil && strings.TrimSpace(*batch.ObservationsNotes) != "" {
		b.WriteString("observations_notes: ")
		b.WriteString(sanitize.PlainNotes(*batch.ObservationsNotes, 8000))
		b.WriteByte('\n')
	}
	if batch.RelatedTaskID != nil && *batch.RelatedTaskID > 0 {
		b.WriteString("related_task_id: ")
		b.WriteString(strconv.FormatInt(*batch.RelatedTaskID, 10))
		b.WriteByte('\n')
	}
	return strings.TrimSpace(b.String())
}

// AlertNotificationDocument embeds rendered alert text without recipient identity (§4.2).
func AlertNotificationDocument(a db.Gr33ncoreAlertsNotification) string {
	var b strings.Builder
	b.WriteString("alert_notification\n")
	if a.Severity.Valid {
		b.WriteString("severity: ")
		b.WriteString(string(a.Severity.Gr33ncoreNotificationPriorityEnum))
		b.WriteByte('\n')
	}
	if a.Status.Valid {
		b.WriteString("status: ")
		b.WriteString(string(a.Status.Gr33ncoreNotificationStatusEnum))
		b.WriteByte('\n')
	}
	if a.TriggeringEventSourceType != nil && strings.TrimSpace(*a.TriggeringEventSourceType) != "" {
		b.WriteString("triggering_event_source_type: ")
		b.WriteString(strings.TrimSpace(*a.TriggeringEventSourceType))
		b.WriteByte('\n')
	}
	if a.TriggeringEventSourceID != nil && *a.TriggeringEventSourceID > 0 {
		b.WriteString("triggering_event_source_id: ")
		b.WriteString(strconv.FormatInt(*a.TriggeringEventSourceID, 10))
		b.WriteByte('\n')
	}
	if a.SubjectRendered != nil && strings.TrimSpace(*a.SubjectRendered) != "" {
		b.WriteString("subject: ")
		b.WriteString(sanitize.PlainNotes(*a.SubjectRendered, 4000))
		b.WriteByte('\n')
	}
	if a.MessageTextRendered != nil && strings.TrimSpace(*a.MessageTextRendered) != "" {
		b.WriteString("message: ")
		b.WriteString(sanitize.PlainNotes(*a.MessageTextRendered, 12000))
		b.WriteByte('\n')
	}
	return strings.TrimSpace(b.String())
}
