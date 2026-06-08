// Phase 65 — Pi diagnostics read tool registration.
package main

import (
	"strings"
	"testing"

	"gr33n-api/internal/ai"
	"gr33n-api/internal/farmguardian"
)

func TestPhase65_SummarizeDeviceHealthReadToolRegistered(t *testing.T) {
	found := false
	for _, id := range farmguardian.ReadToolIDs() {
		if id == "summarize_device_health" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("summarize_device_health missing from ReadToolIDs")
	}
	block := farmguardian.PlatformContextBlock(ai.Config{Enabled: true}, true, farmguardian.ReadToolIDs())
	if !strings.Contains(block, "summarize_device_health") {
		t.Fatal("platform context missing device health rule")
	}
}
