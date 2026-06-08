// Phase 61 — proactive Guardian nudge endpoint registration.
package main

import (
	"strings"
	"testing"

	"gr33n-api/internal/farmguardian"
)

func TestPhase61_GuardianNudgeEnginePresent(t *testing.T) {
	block := farmguardian.NudgeContextBlock(farmguardian.ContextRef{
		NudgeCategory: "comfort_breach",
		NudgeID:       "comfort-1-temperature",
	})
	if !strings.Contains(block, "comfort_breach") {
		t.Fatal("NudgeContextBlock missing category")
	}
}
