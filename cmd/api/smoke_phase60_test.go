// Phase 60 — morning walkthrough read tool registration.
package main

import (
	"strings"
	"testing"

	"gr33n-api/internal/ai"
	"gr33n-api/internal/farmguardian"
)

func TestPhase60_WalkFarmReadToolRegistered(t *testing.T) {
	found := false
	for _, id := range farmguardian.ReadToolIDs() {
		if id == "walk_farm" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("walk_farm missing from ReadToolIDs")
	}
	block := farmguardian.PlatformContextBlock(ai.Config{Enabled: true}, true, farmguardian.ReadToolIDs())
	if !strings.Contains(block, "walk_farm") {
		t.Fatal("platform context missing morning walkthrough rule")
	}
}
