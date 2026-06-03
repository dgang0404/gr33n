package farmguardian

import (
	"regexp"
	"strings"
)

// SafetyStopMessage is returned when the operator asks for prohibited instructions (Phase 37 WS4).
const SafetyStopMessage = `I have to stop here. Wiring mains AC (120V/240V line voltage) or working on pressurized or potable water lines is not something I can walk you through step by step.

Please use a licensed electrician for line-voltage work and a licensed plumber for pressurized or potable plumbing. I can still help with low-voltage Pi control wiring (GPIO, relay IN/VCC/GND) and general checks — say "start procedure wire-pi-relay-light" or ask about a specific symptom.`

var (
	mainsWiringRe = regexp.MustCompile(`(?i)(120\s*v|240\s*v|110\s*v|mains|line[- ]?voltage|wall\s+outlet|breaker\s+panel|wire\s+the\s+.*\s+to\s+the\s+relay|hot\s+wire|neutral\s+wire)`)
	pressurizedPlumbingRe = regexp.MustCompile(`(?i)(pressurized|potable|municipal\s+water|backflow|house\s+pressure|solenoid\s+on\s+the\s+main)`)
)

// UnsafeInstructionRequest reports whether the operator message asks for prohibited step-by-step work.
func UnsafeInstructionRequest(message string) bool {
	msg := strings.TrimSpace(message)
	if msg == "" {
		return false
	}
	if mainsWiringRe.FindStringIndex(msg) != nil {
		return true
	}
	if pressurizedPlumbingRe.FindStringIndex(msg) != nil {
		return true
	}
	return false
}

// SafetyStopForTier returns escalation copy for a procedure step tier.
func SafetyStopForTier(tier string) string {
	switch strings.TrimSpace(strings.ToLower(tier)) {
	case SafetyTierQualifiedPersonRequired:
		return `This step is **line-voltage or high-risk physical work**. I cannot guide you through it in chat.

Stop here and bring in a **licensed electrician** (mains / grow-light load side) or **licensed plumber** (pressurized or potable lines). You can continue other low-voltage Pi wiring steps later, or print the checklist for reference.`
	default:
		return ""
	}
}
