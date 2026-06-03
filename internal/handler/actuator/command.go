package actuator

import (
	"encoding/json"
	"fmt"
	"strings"
)

// NormalizeCommand lowercases and trims an operator or automation command string.
func NormalizeCommand(command string) string {
	return strings.ToLower(strings.TrimSpace(command))
}

// CommandAllowed reports whether command is valid for the given actuator_type.
func CommandAllowed(actuatorType, command string) bool {
	cmd := NormalizeCommand(command)
	for _, c := range ValidCommands(actuatorType) {
		if c == cmd {
			return true
		}
	}
	return false
}

// PulseDurationAllowed reports whether timed on/off pulses are supported for this actuator type.
func PulseDurationAllowed(actuatorType string) bool {
	switch actuatorType {
	case "pump", "relay", "water_valve", "return_pump", "air_pump", "feeder_hopper":
		return true
	default:
		return false
	}
}

// ValidatePulseDuration returns an error when duration is set but invalid for the actuator type.
func ValidatePulseDuration(actuatorType string, durationSeconds *int) error {
	if durationSeconds == nil {
		return nil
	}
	d := *durationSeconds
	if d <= 0 {
		return fmt.Errorf("duration_seconds must be positive")
	}
	if d > 3600 {
		return fmt.Errorf("duration_seconds must be at most 3600")
	}
	if !PulseDurationAllowed(actuatorType) {
		return fmt.Errorf("actuator_type %q does not support timed pulses; omit duration_seconds", actuatorType)
	}
	return nil
}

// PendingCommandInput is the full pending_command payload for Pi poll and automation.
type PendingCommandInput struct {
	ActuatorID      int64
	Command         string
	Source          string
	Reason          string
	DurationSeconds *int
	ScheduleID      *int64
	RuleID          *int64
	ProgramID       *int64
}

// BuildPendingCommandJSON returns the devices.config.pending_command payload
// written for Pi poll (same shape as automation worker and Guardian enqueue).
func BuildPendingCommandJSON(actuatorID int64, command, source, reason string) ([]byte, error) {
	return BuildPendingCommandJSONFull(PendingCommandInput{
		ActuatorID: actuatorID,
		Command:    command,
		Source:     source,
		Reason:     reason,
	})
}

// BuildPendingCommandJSONFull builds pending_command with optional duration and provenance IDs.
func BuildPendingCommandJSONFull(in PendingCommandInput) ([]byte, error) {
	cmd := NormalizeCommand(in.Command)
	if cmd == "" {
		return nil, fmt.Errorf("command is required")
	}
	source := in.Source
	if source == "" {
		source = "operator"
	}
	pending := map[string]any{
		"command":     cmd,
		"actuator_id": in.ActuatorID,
		"source":      source,
	}
	if in.Reason != "" {
		pending["reason"] = in.Reason
	}
	if in.DurationSeconds != nil && *in.DurationSeconds > 0 {
		pending["duration_seconds"] = *in.DurationSeconds
	}
	if in.ScheduleID != nil {
		pending["schedule_id"] = *in.ScheduleID
	}
	if in.RuleID != nil {
		pending["rule_id"] = *in.RuleID
	}
	if in.ProgramID != nil {
		pending["program_id"] = *in.ProgramID
	}
	return json.Marshal(pending)
}
