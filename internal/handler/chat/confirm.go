package chat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"gr33n-api/internal/auditlog"
	"gr33n-api/internal/authctx"
	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/farmguardian"
	"gr33n-api/internal/farmguardian/tools"
	"gr33n-api/internal/httputil"
)

type confirmBody struct {
	ProposalID string `json:"proposal_id"`
}

type confirmResponse struct {
	Result  any    `json:"result,omitempty"`
	Summary string `json:"summary"`
}

// PostConfirm handles POST /v1/chat/confirm — executes a frozen proposal (Phase 29 WS3).
func (h *Handler) PostConfirm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if h.q == nil {
		httputil.WriteError(w, http.StatusInternalServerError, "database unavailable")
		return
	}
	userID, hasUser := authctx.UserID(r.Context())
	if !hasUser {
		httputil.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<16))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	var cb confirmBody
	if err := json.Unmarshal(body, &cb); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	pid, err := uuid.Parse(strings.TrimSpace(cb.ProposalID))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid proposal_id")
		return
	}

	ctx := r.Context()
	_ = h.q.ExpireStaleGuardianProposals(ctx)

	prop, err := h.q.GetGuardianProposalByID(ctx, pid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "proposal not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if prop.UserID != userID {
		httputil.WriteError(w, http.StatusForbidden, "proposal belongs to another user")
		return
	}

	if prop.Status == db.Gr33ncoreGuardianProposalStatusEnumConfirmed {
		var cached any
		if len(prop.Result) > 0 {
			_ = json.Unmarshal(prop.Result, &cached)
		}
		httputil.WriteJSON(w, http.StatusOK, confirmResponse{Result: cached, Summary: prop.Summary})
		return
	}
	if prop.Status != db.Gr33ncoreGuardianProposalStatusEnumPending {
		httputil.WriteError(w, http.StatusGone, "proposal is no longer confirmable")
		return
	}
	if time.Now().UTC().After(prop.ExpiresAt) {
		httputil.WriteError(w, http.StatusGone, "proposal expired")
		return
	}

	tool, err := tools.Lookup(prop.ToolID)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if tool.RequiresAdmin && !farmauthz.RequireFarmAdmin(w, r, h.q, prop.FarmID) {
		return
	}
	if tool.RequiresOperate && !farmauthz.RequireFarmOperate(w, r, h.q, prop.FarmID) {
		return
	}
	if !h.checkCostBudget(ctx, w, userID, hasUser, prop.FarmID) {
		return
	}

	var args map[string]any
	if len(prop.Args) > 0 {
		_ = json.Unmarshal(prop.Args, &args)
	}
	result, execErr := tools.Execute(ctx, prop.ToolID, args, tools.ExecutorDeps{
		Q:          h.q,
		Pool:       h.pool,
		UserID:     userID,
		HasUser:    true,
		FarmID:     prop.FarmID,
		ProposalID: pid,
		Request:    r,
	})
	if execErr != nil {
		h.auditToolFailure(r, prop, args, execErr.Error())
		httputil.WriteError(w, http.StatusBadRequest, execErr.Error())
		return
	}

	resultJSON, _ := json.Marshal(result)
	confirmed, err := h.q.ConfirmGuardianProposal(ctx, db.ConfirmGuardianProposalParams{
		ProposalID: pid,
		Result:     resultJSON,
		UserID:     userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusGone, "proposal expired or already handled")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.auditToolExecution(r, confirmed, result)

	summary := successSummary(prop.ToolID, result)
	httputil.WriteJSON(w, http.StatusOK, confirmResponse{Result: result, Summary: summary})
}

func (h *Handler) auditToolExecution(r *http.Request, prop db.Gr33ncoreGuardianActionProposal, result any) {
	var args map[string]any
	_ = json.Unmarshal(prop.Args, &args)
	details := map[string]any{
		"kind":        "guardian_tool_executed",
		"tool_id":     prop.ToolID,
		"proposal_id": prop.ProposalID.String(),
		"args":        args,
		"result":      result,
	}
	table, recID := auditTargetForTool(prop.ToolID, args, result)
	auditlog.Submit(r.Context(), h.q, r, auditlog.Event{
		FarmID:         auditlog.FarmIDPtr(prop.FarmID),
		Action:         db.Gr33ncoreUserActionTypeEnumGuardianToolExecuted,
		TargetSchema:   strPtr("gr33ncore"),
		TargetTable:    table,
		TargetRecordID: recID,
		Status:         "success",
		Details:        details,
	})
}

func (h *Handler) auditToolFailure(r *http.Request, prop db.Gr33ncoreGuardianActionProposal, args map[string]any, reason string) {
	details := map[string]any{
		"kind":        "guardian_tool_executed",
		"tool_id":     prop.ToolID,
		"proposal_id": prop.ProposalID.String(),
		"args":        args,
	}
	table, recID := auditTargetForTool(prop.ToolID, args, nil)
	st := "failure"
	auditlog.Submit(r.Context(), h.q, r, auditlog.Event{
		FarmID:         auditlog.FarmIDPtr(prop.FarmID),
		Action:         db.Gr33ncoreUserActionTypeEnumGuardianToolExecuted,
		TargetSchema:   strPtr("gr33ncore"),
		TargetTable:    table,
		TargetRecordID: recID,
		Status:         st,
		FailureReason:  &reason,
		Details:        details,
	})
}

func successSummary(toolID string, result any) string {
	switch toolID {
	case "ack_alert":
		if m, ok := result.(map[string]any); ok {
			if id, ok := m["alert_id"]; ok {
				return "Alert acknowledged (#" + formatAnyInt(id) + ")."
			}
		}
		return "Alert acknowledged."
	case "mark_alert_read":
		if m, ok := result.(map[string]any); ok {
			if id, ok := m["alert_id"]; ok {
				return "Alert marked read (#" + formatAnyInt(id) + ")."
			}
		}
		return "Alert marked read."
	case "create_task", "create_task_from_alert":
		if m, ok := result.(map[string]any); ok {
			if id, ok := m["task_id"]; ok {
				return "Task created (#" + formatAnyInt(id) + ")."
			}
		}
		return "Task created."
	case "update_cycle_stage":
		if m, ok := result.(map[string]any); ok {
			if stage, ok := m["current_stage"].(string); ok && stage != "" {
				return "Crop cycle stage updated to " + stage + "."
			}
		}
		return "Crop cycle stage updated."
	case "create_plant":
		if m, ok := result.(map[string]any); ok {
			if name, ok := m["display_name"].(string); ok && name != "" {
				return "Plant created: " + name + "."
			}
		}
		return "Plant created."
	case "create_crop_cycle":
		if m, ok := result.(map[string]any); ok {
			if name, ok := m["name"].(string); ok && name != "" {
				return "Crop cycle started: " + name + "."
			}
		}
		return "Crop cycle created."
	case "create_fertigation_program":
		if m, ok := result.(map[string]any); ok {
			if name, ok := m["name"].(string); ok && name != "" {
				return "Fertigation program created: " + name + "."
			}
		}
		return "Fertigation program created."
	case "apply_grow_setup_pack":
		return "Grow setup pack applied — plant, crop cycle, and fertigation program created."
	case "patch_schedule", "patch_fertigation_program", "patch_rule":
		return "Configuration updated."
	case "apply_bootstrap_template":
		return "Bootstrap template applied."
	case "enqueue_actuator_command":
		if m, ok := result.(map[string]any); ok {
			cmd, _ := m["command"].(string)
			name, _ := m["actuator_name"].(string)
			if name != "" && cmd != "" {
				return fmt.Sprintf("Queued %q for %s — Pi will run on next poll.", cmd, name)
			}
			if id, ok := m["actuator_id"]; ok {
				return "Actuator command queued (#" + formatAnyInt(id) + ")."
			}
		}
		return "Actuator command queued for Pi pickup."
	default:
		return "Action completed."
	}
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func auditTargetForTool(toolID string, args map[string]any, result any) (*string, *string) {
	switch toolID {
	case "enqueue_actuator_command":
		table := strPtr("devices")
		recID := strPtr("")
		if id, ok := args["device_id"]; ok {
			recID = strPtr(formatAnyInt(id))
		} else if result != nil {
			if m, ok := result.(map[string]any); ok {
				if id, ok := m["device_id"]; ok {
					recID = strPtr(formatAnyInt(id))
				}
			}
		}
		return table, recID
	case "create_task", "create_task_from_alert":
		table := strPtr("tasks")
		recID := strPtr("")
		if result != nil {
			if m, ok := result.(map[string]any); ok {
				if id, ok := m["task_id"]; ok {
					recID = strPtr(formatAnyInt(id))
				}
			}
		}
		return table, recID
	case "create_plant":
		table := strPtr("plants")
		recID := strPtr("")
		if result != nil {
			if m, ok := result.(map[string]any); ok {
				if id, ok := m["plant_id"]; ok {
					recID = strPtr(formatAnyInt(id))
				}
			}
		}
		return table, recID
	case "create_crop_cycle":
		table := strPtr("crop_cycles")
		recID := strPtr("")
		if result != nil {
			if m, ok := result.(map[string]any); ok {
				if id, ok := m["crop_cycle_id"]; ok {
					recID = strPtr(formatAnyInt(id))
				}
			}
		}
		return table, recID
	case "create_fertigation_program":
		table := strPtr("programs")
		recID := strPtr("")
		if result != nil {
			if m, ok := result.(map[string]any); ok {
				if id, ok := m["program_id"]; ok {
					recID = strPtr(formatAnyInt(id))
				}
			}
		}
		return table, recID
	case "apply_grow_setup_pack":
		table := strPtr("crop_cycles")
		recID := strPtr("")
		if result != nil {
			if m, ok := result.(map[string]any); ok {
				if cycle, ok := m["cycle"].(map[string]any); ok {
					if id, ok := cycle["crop_cycle_id"]; ok {
						recID = strPtr(formatAnyInt(id))
					}
				}
			}
		}
		return table, recID
	default:
		table := strPtr("alerts_notifications")
		recID := strPtr("")
		if id, ok := args["alert_id"]; ok {
			recID = strPtr(formatAnyInt(id))
		}
		return table, recID
	}
}

func formatAnyInt(v any) string {
	switch n := v.(type) {
	case float64:
		return formatInt(int64(n))
	case int64:
		return formatInt(n)
	case int:
		return formatInt(int64(n))
	default:
		return ""
	}
}

func formatInt(n int64) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

// attachProposals adds rule-assisted proposals to the chat response when grounded.
func (h *Handler) attachProposals(
	ctx context.Context,
	farmID int64,
	hasUser bool,
	userID uuid.UUID,
	sessionID uuid.UUID,
	question string,
	snap farmguardian.Snapshot,
	resp *postResponse,
) {
	if !hasUser || farmID <= 0 || h.q == nil {
		return
	}
	props, err := farmguardian.BuildRuleAssistedProposals(ctx, h.q, userID, farmID, sessionID, question, snap)
	if err != nil {
		return
	}
	resp.Proposals = props
}
