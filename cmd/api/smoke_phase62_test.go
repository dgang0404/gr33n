// Phase 62 — grow advisor read tool registration.
package main

import (
	"strings"
	"testing"

	"gr33n-api/internal/ai"
	"gr33n-api/internal/farmguardian"
)

func TestPhase62_GrowAdvisorReadToolRegistered(t *testing.T) {
	found := false
	for _, id := range farmguardian.ReadToolIDs() {
		if id == "grow_advisor" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("grow_advisor missing from ReadToolIDs")
	}
	if got := farmguardian.CalcVPDKpa(25, 50); got < 1.5 || got > 1.7 {
		t.Fatalf("CalcVPDKpa sanity check failed: %v", got)
	}
	block := farmguardian.PlatformContextBlock(ai.Config{Enabled: true}, true, farmguardian.ReadToolIDs())
	if !strings.Contains(block, "grow_advisor") {
		t.Fatal("platform context missing grow_advisor persona rule")
	}
}
