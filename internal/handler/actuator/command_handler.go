package actuator

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/httputil"
)

// actuatorWithCommands is the list/get response shape including valid_commands.
type actuatorWithCommands struct {
	db.Gr33ncoreActuator
	ValidCommands []string `json:"valid_commands"`
}

func wrapActuatorWithCommands(row db.Gr33ncoreActuator) actuatorWithCommands {
	return actuatorWithCommands{
		Gr33ncoreActuator: row,
		ValidCommands:     ValidCommands(row.ActuatorType),
	}
}

// POST /actuators/{id}/command
// Queues devices.config.pending_command for the Pi client (operator manual control).
func (h *Handler) EnqueueCommand(w http.ResponseWriter, r *http.Request) {
	actuatorID, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid actuator id")
		return
	}

	var body struct {
		Command         string `json:"command"`
		Reason          string `json:"reason"`
		DurationSeconds *int   `json:"duration_seconds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	command := NormalizeCommand(body.Command)
	if command == "" {
		httputil.WriteError(w, http.StatusBadRequest, "command is required")
		return
	}

	ctx := r.Context()
	actuator, err := h.q.GetActuatorByID(ctx, actuatorID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "actuator not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load actuator")
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, actuator.FarmID) {
		return
	}
	if !CommandAllowed(actuator.ActuatorType, command) {
		httputil.WriteError(w, http.StatusBadRequest,
			fmt.Sprintf("command %q is not valid for actuator_type %q; valid commands: %v",
				command, actuator.ActuatorType, ValidCommands(actuator.ActuatorType)))
		return
	}
	if err := ValidatePulseDuration(actuator.ActuatorType, body.DurationSeconds); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if actuator.DeviceID == nil {
		httputil.WriteError(w, http.StatusBadRequest, "actuator is not bound to a device; link device_id before sending commands")
		return
	}

	device, err := h.q.GetDeviceByID(ctx, *actuator.DeviceID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusBadRequest, "actuator device not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load device")
		return
	}
	if device.FarmID != actuator.FarmID {
		httputil.WriteError(w, http.StatusBadRequest, "actuator device farm mismatch")
		return
	}

	cmdType := "actuator"
	if body.DurationSeconds != nil && *body.DurationSeconds > 0 {
		cmdType = "pulse"
	}

	pendingJSON, err := BuildPendingCommandJSONFull(PendingCommandInput{
		ActuatorID:      actuatorID,
		Command:         command,
		Source:          "operator",
		Reason:          body.Reason,
		DurationSeconds: body.DurationSeconds,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Phase 39 WS1: enqueue into the FIFO device_commands queue.
	queued, qErr := h.q.EnqueueDeviceCommand(ctx, db.EnqueueDeviceCommandParams{
		DeviceID:    device.ID,
		FarmID:      device.FarmID,
		CommandType: cmdType,
		Payload:     pendingJSON,
		Source:      "operator",
		ActuatorID:  &actuatorID,
	})
	if qErr != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to enqueue command")
		return
	}

	// Backward compat: mirror head payload on devices.config.pending_command
	// so pre-39 Pi clients still pick it up.
	_ = h.q.SetDevicePendingCommand(ctx, db.SetDevicePendingCommandParams{
		ID:      device.ID,
		Column2: pendingJSON,
	})

	var pending map[string]any
	_ = json.Unmarshal(pendingJSON, &pending)

	resp := map[string]any{
		"command_id":      queued.ID,
		"device_id":       device.ID,
		"actuator_id":     actuatorID,
		"command":         command,
		"command_type":    cmdType,
		"pending_command": pending,
		"actuator_name":   actuator.Name,
		"device_name":     device.Name,
		"valid_commands":  ValidCommands(actuator.ActuatorType),
		"pulse_supported": PulseDurationAllowed(actuator.ActuatorType),
	}
	if body.DurationSeconds != nil {
		resp["duration_seconds"] = *body.DurationSeconds
	}
	httputil.WriteJSON(w, http.StatusAccepted, resp)
}
