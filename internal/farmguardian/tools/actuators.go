package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	db "gr33n-api/internal/db"
	acthandler "gr33n-api/internal/handler/actuator"
)

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
	command = acthandler.NormalizeCommand(command)
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
	if !acthandler.CommandAllowed(actuator.ActuatorType, command) {
		return nil, fmt.Errorf("command %q is not valid for actuator_type %q; valid: %v",
			command, actuator.ActuatorType, acthandler.ValidCommands(actuator.ActuatorType))
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

	pendingJSON, err := acthandler.BuildPendingCommandJSON(actuatorID, command, "guardian", "")
	if err != nil {
		return nil, err
	}
	var pending map[string]any
	if err := json.Unmarshal(pendingJSON, &pending); err != nil {
		return nil, err
	}
	if reason != nil && *reason != "" {
		pending["reason"] = *reason
	}
	if deps.ProposalID != uuid.Nil {
		pending["proposal_id"] = deps.ProposalID.String()
	}
	pendingJSON, err = json.Marshal(pending)
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
