package ingest

import (
	"strconv"
	"strings"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/rag/sanitize"
)

const (
	SourceTypeTask           = "task"
	SourceTypeAutomationRun  = "automation_run"
	metadataModuleCore       = "core"
	metadataModuleAutomation = "automation"
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
