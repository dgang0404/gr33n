// Package devicecmd implements the Phase 39 WS1 device command queue API.
//
// Routes (registered in cmd/api/routes.go):
//
//	POST   /devices/{id}/commands           — enqueue (JWT operator or Pi-key; requires farm operate)
//	GET    /devices/{id}/commands/next      — Pi-key: dequeue head (marks in_progress)
//	POST   /devices/{id}/commands/{cid}/ack — Pi-key: complete or fail
//	GET    /devices/{id}/commands           — JWT operator: list with optional ?status= filter
package devicecmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	db "gr33n-api/internal/db"
	acthandler "gr33n-api/internal/handler/actuator"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/httputil"
)

// Handler holds the query layer.
type Handler struct {
	q db.Querier
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{q: db.New(pool)}
}

func NewHandlerWithQuerier(q db.Querier) *Handler {
	return &Handler{q: q}
}

// POST /devices/{id}/commands
// Body: { "command_type": "pulse"|"actuator", "actuator_id": N, "command": "on",
//         "duration_seconds": N?, "reason": "..." }
//
// Also accepts "command_type": "mix_batch" with a full MixPlan payload (Phase 39 WS3).
func (h *Handler) Enqueue(w http.ResponseWriter, r *http.Request) {
	deviceID, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid device id")
		return
	}

	ctx := r.Context()
	device, err := h.q.GetDeviceByID(ctx, deviceID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "device not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load device")
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, device.FarmID) {
		return
	}

	var body struct {
		CommandType     string          `json:"command_type"`
		ActuatorID      *int64          `json:"actuator_id"`
		Command         string          `json:"command"`
		DurationSeconds *int            `json:"duration_seconds"`
		Reason          string          `json:"reason"`
		Payload         json.RawMessage `json:"payload"` // mix_batch: pass the full MixPlan here
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmdType := strings.TrimSpace(body.CommandType)
	if cmdType == "" {
		cmdType = "actuator"
	}
	switch cmdType {
	case "actuator", "pulse", "mix_batch":
	default:
		httputil.WriteError(w, http.StatusBadRequest,
			fmt.Sprintf("command_type must be actuator, pulse, or mix_batch; got %q", cmdType))
		return
	}

	// For actuator / pulse commands, build the payload from the simple fields.
	// For mix_batch the caller supplies the full payload JSON.
	var payload json.RawMessage
	var actuatorIDPtr *int64

	switch cmdType {
	case "actuator", "pulse":
		if body.ActuatorID == nil {
			httputil.WriteError(w, http.StatusBadRequest, "actuator_id is required for command_type actuator or pulse")
			return
		}
		command := acthandler.NormalizeCommand(body.Command)
		if command == "" {
			httputil.WriteError(w, http.StatusBadRequest, "command is required")
			return
		}
		actuator, err := h.q.GetActuatorByID(ctx, *body.ActuatorID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				httputil.WriteError(w, http.StatusBadRequest, "actuator not found")
				return
			}
			httputil.WriteError(w, http.StatusInternalServerError, "failed to load actuator")
			return
		}
		if actuator.DeviceID == nil || *actuator.DeviceID != deviceID {
			httputil.WriteError(w, http.StatusBadRequest, "actuator is not bound to this device")
			return
		}
		if !acthandler.CommandAllowed(actuator.ActuatorType, command) {
			httputil.WriteError(w, http.StatusBadRequest,
				fmt.Sprintf("command %q is not valid for actuator_type %q; valid: %v",
					command, actuator.ActuatorType, acthandler.ValidCommands(actuator.ActuatorType)))
			return
		}
		if cmdType == "pulse" {
			if err := acthandler.ValidatePulseDuration(actuator.ActuatorType, body.DurationSeconds); err != nil {
				httputil.WriteError(w, http.StatusBadRequest, err.Error())
				return
			}
		}
		m := map[string]any{
			"actuator_id": *body.ActuatorID,
			"command":     command,
			"source":      "operator",
		}
		if body.Reason != "" {
			m["reason"] = body.Reason
		}
		if body.DurationSeconds != nil && *body.DurationSeconds > 0 {
			m["duration_seconds"] = *body.DurationSeconds
		}
		payload, _ = json.Marshal(m)
		actuatorIDPtr = body.ActuatorID

	case "mix_batch":
		if len(body.Payload) == 0 || string(body.Payload) == "null" {
			httputil.WriteError(w, http.StatusBadRequest, "payload is required for command_type mix_batch")
			return
		}
		payload = body.Payload
	}

	cmd, err := h.q.EnqueueDeviceCommand(ctx, db.EnqueueDeviceCommandParams{
		DeviceID:    deviceID,
		FarmID:      device.FarmID,
		CommandType: cmdType,
		Payload:     payload,
		Source:      "operator",
		ActuatorID:  actuatorIDPtr,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to enqueue command")
		return
	}

	// Backward compat: mirror head payload on devices.config.pending_command.
	// Old Pi clients (pre-39) still pick this up. Remove in a future release.
	_ = h.q.SetDevicePendingCommand(ctx, db.SetDevicePendingCommandParams{
		ID:      deviceID,
		Column2: payload,
	})

	httputil.WriteJSON(w, http.StatusAccepted, cmd)
}

// GET /devices/{id}/commands/next
// Pi-key: atomically claims the oldest pending command (sets in_progress).
// Returns 204 when the queue is empty.
func (h *Handler) Next(w http.ResponseWriter, r *http.Request) {
	deviceID, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid device id")
		return
	}

	ctx := r.Context()
	device, err := h.q.GetDeviceByID(ctx, deviceID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "device not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load device")
		return
	}
	// Pi-key is farm-scoped; ensure the key belongs to this farm.
	if !farmauthz.RequireFarmMemberOrPiEdge(w, r, h.q, device.FarmID) {
		return
	}

	cmd, err := h.q.GetNextDeviceCommand(ctx, deviceID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Queue is empty — 204 No Content so Pi knows to stop polling until next tick.
			w.WriteHeader(http.StatusNoContent)
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to dequeue command")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, cmd)
}

// POST /devices/{id}/commands/{cid}/ack
// Pi-key: mark command completed or failed, with optional result payload.
// Body: { "status": "completed"|"failed", "result": {...}? }
func (h *Handler) Ack(w http.ResponseWriter, r *http.Request) {
	deviceID, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid device id")
		return
	}
	cmdID, err := httputil.PathID(r.URL.Path, 4)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid command id")
		return
	}

	var body struct {
		Status string          `json:"status"`
		Result json.RawMessage `json:"result"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Status != "completed" && body.Status != "failed" {
		httputil.WriteError(w, http.StatusBadRequest, "status must be completed or failed")
		return
	}

	ctx := r.Context()
	device, err := h.q.GetDeviceByID(ctx, deviceID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "device not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load device")
		return
	}
	if !farmauthz.RequireFarmMemberOrPiEdge(w, r, h.q, device.FarmID) {
		return
	}

	result := body.Result
	if len(result) == 0 {
		result = json.RawMessage(`{}`)
	}

	cmd, err := h.q.AckDeviceCommand(ctx, db.AckDeviceCommandParams{
		ID:     cmdID,
		Status: body.Status,
		Result: result,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "command not found or already acked")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to ack command")
		return
	}

	// When a command completes, clear the legacy pending_command mirror so
	// old Pi clients don't re-execute it.
	if body.Status == "completed" {
		_ = h.q.ClearDevicePendingCommand(ctx, deviceID)
	}

	httputil.WriteJSON(w, http.StatusOK, cmd)
}

// GET /devices/{id}/commands
// JWT operator: list commands for a device; optional ?status= filter.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	deviceID, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid device id")
		return
	}

	ctx := r.Context()
	device, err := h.q.GetDeviceByID(ctx, deviceID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "device not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load device")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, device.FarmID) {
		return
	}

	var statusFilter *string
	if s := r.URL.Query().Get("status"); s != "" {
		statusFilter = &s
	}
	statusVal := ""
	if statusFilter != nil {
		statusVal = *statusFilter
	}
	cmds, err := h.q.ListDeviceCommands(ctx, db.ListDeviceCommandsParams{
		DeviceID: deviceID,
		Column2:  statusVal,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list commands")
		return
	}
	if cmds == nil {
		cmds = []db.Gr33ncoreDeviceCommand{}
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"device_id": deviceID,
		"commands":  cmds,
	})
}

// pathSegment returns the Nth segment (0-indexed) of a URL path split by '/'.
// e.g. "/devices/5/commands/3/ack" → segments [devices 5 commands 3 ack]
func pathSegment(path string, n int) (string, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if n >= len(parts) {
		return "", fmt.Errorf("path segment %d out of range", n)
	}
	return parts[n], nil
}

func idSegment(path string, n int) (int64, error) {
	s, err := pathSegment(path, n)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(s, 10, 64)
}
