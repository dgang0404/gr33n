// Package tools executes confirmed Farm Guardian actions in-process (Phase 29 WS2).
package tools

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	db "gr33n-api/internal/db"
)

// Tool describes one operator-confirmed action exposed to Guardian.
type Tool struct {
	ID              string
	Description     string
	RequiresOperate bool
	RequiresAdmin   bool // farm admin (e.g. bootstrap template)
	Execute         func(ctx context.Context, deps ExecutorDeps, args map[string]any) (any, error)
}

// ExecutorDeps carries auth + persistence handles for tool execution.
type ExecutorDeps struct {
	Q       db.Querier
	UserID  uuid.UUID
	HasUser bool
	FarmID  int64 // proposal farm scope
	Request *http.Request
}

// registry is the Guardian tool catalog (Phase 29–30).
var registry = map[string]Tool{
	"mark_alert_read": {
		ID:              "mark_alert_read",
		Description:     "Mark an alert as read (PATCH /alerts/{id}/read)",
		RequiresOperate: true,
		Execute:         execMarkAlertRead,
	},
	"ack_alert": {
		ID:              "ack_alert",
		Description:     "Acknowledge an alert (PATCH /alerts/{id}/acknowledge)",
		RequiresOperate: true,
		Execute:         execAckAlert,
	},
	"create_task_from_alert": {
		ID:              "create_task_from_alert",
		Description:     "Create a task from an alert (POST /alerts/{id}/create-task)",
		RequiresOperate: true,
		Execute:         execCreateTaskFromAlert,
	},
	"create_task": {
		ID:              "create_task",
		Description:     "Create a farm task (POST /farms/{id}/tasks)",
		RequiresOperate: true,
		Execute:         execCreateTask,
	},
	"update_cycle_stage": {
		ID:              "update_cycle_stage",
		Description:     "Update crop cycle growth stage (PATCH /crop-cycles/{id}/stage)",
		RequiresOperate: true,
		Execute:         execUpdateCycleStage,
	},
	"patch_schedule": {
		ID:              "patch_schedule",
		Description:     "Patch schedule name, cron, or active flag (PUT /schedules/{id})",
		RequiresOperate: true,
		Execute:         execPatchSchedule,
	},
	"patch_fertigation_program": {
		ID:              "patch_fertigation_program",
		Description:     "Patch fertigation program EC target, volume, or active flag",
		RequiresOperate: true,
		Execute:         execPatchFertigationProgram,
	},
	"patch_rule": {
		ID:              "patch_rule",
		Description:     "Patch automation rule active flag or first threshold predicate",
		RequiresOperate: true,
		Execute:         execPatchRule,
	},
	"apply_bootstrap_template": {
		ID:              "apply_bootstrap_template",
		Description:     "Apply a farm bootstrap template (POST /farms/{id}/bootstrap-template)",
		RequiresOperate: false,
		RequiresAdmin:   true,
		Execute:         execApplyBootstrapTemplate,
	},
}

// Lookup returns a registered tool or an error.
func Lookup(id string) (Tool, error) {
	t, ok := registry[id]
	if !ok {
		return Tool{}, fmt.Errorf("unknown tool %q", id)
	}
	return t, nil
}

// IDs returns tool IDs for system-prompt listing.
func IDs() []string {
	out := make([]string, 0, len(registry))
	for id := range registry {
		out = append(out, id)
	}
	return out
}

// Execute runs a registered tool with validated args.
func Execute(ctx context.Context, toolID string, args map[string]any, deps ExecutorDeps) (any, error) {
	t, err := Lookup(toolID)
	if err != nil {
		return nil, err
	}
	if !deps.HasUser {
		return nil, errors.New("authentication required")
	}
	return t.Execute(ctx, deps, args)
}
