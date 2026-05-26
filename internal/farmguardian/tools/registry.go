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

// registry is the v1 tool catalog (extend in later WS).
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
