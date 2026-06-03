package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	db "gr33n-api/internal/db"
)

// allowedActuatorCommands are the commands Guardian may enqueue via pending_command.
// Phase 36 WS7 extends the set to include greenhouse motor verbs (deploy/retract
// for shade_screen, open/close for ridge_vent and glazing_panel).
// The Pi client maps these onto the same GPIO on/off logic with the configured polarity.
var allowedActuatorCommands = map[string]struct{}{
	"on":      {},
	"off":     {},
	"deploy":  {},
	"retract": {},
	"open":    {},
	"close":   {},
	"stop":    {},
}

func execEnqueueActuatorCommand(ctx context.Context, deps ExecutorDeps, args map[string]any) (any, error) {
	deviceID, err := int64FromArgs(args, "device_id")
	if err != nil {
		return nil, err
	}
	actuatorID, err := int64FromArgs(args, "actuator_id")
	if err != nil {
		return nil, err
	}
	command, err := stringFromArgs(args, "command")
	if err != nil {
		return nil, err
	}
	command = strings.ToLower(strings.TrimSpace(command))
	if _, ok := allowedActuatorCommands[command]; !ok {
		return nil, fmt.Errorf("command must be one of: on, off, deploy, retract, open, close, stop")
	}
	reason, _ := optionalStringFromArgs(args, "reason")

	actuator, err := deps.Q.GetActuatorByID(ctx, actuatorID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("actuator %d not found", actuatorID)
		}
		return nil, err
	}
	if err := ensureFarmScope(actuator.FarmID, deps.FarmID); err != nil {
		return nil, err
	}
	if actuator.DeviceID == nil || *actuator.DeviceID != deviceID {
		return nil, errors.New("actuator is not attached to the given device_id")
	}

	device, err := deps.Q.GetDeviceByID(ctx, deviceID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("device %d not found", deviceID)
		}
		return nil, err
	}
	if err := ensureFarmScope(device.FarmID, deps.FarmID); err != nil {
		return nil, err
	}

	pending := map[string]any{
		"command":     command,
		"actuator_id": actuatorID,
		"source":      "guardian",
	}
	if reason != nil && *reason != "" {
		pending["reason"] = *reason
	}
	if deps.ProposalID != uuid.Nil {
		pending["proposal_id"] = deps.ProposalID.String()
	}
	pendingJSON, err := json.Marshal(pending)
	if err != nil {
		return nil, err
	}
	if err := deps.Q.SetDevicePendingCommand(ctx, db.SetDevicePendingCommandParams{
		ID:      deviceID,
		Column2: pendingJSON,
	}); err != nil {
		return nil, err
	}

	return map[string]any{
		"device_id":        deviceID,
		"actuator_id":      actuatorID,
		"command":          command,
		"pending_command":  pending,
		"actuator_name":    actuator.Name,
		"device_name":      device.Name,
	}, nil
}
