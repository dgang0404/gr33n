// Phase 67 — field assistant routes and vision grounding.
package main

import (
	"os"
	"strings"
	"testing"

	"gr33n-api/internal/farmguardian"
)

func TestPhase67_FieldAssistantRoutesAndVision(t *testing.T) {
	data, err := os.ReadFile("routes.go")
	if err != nil {
		t.Fatalf("read routes.go: %v", err)
	}
	body := string(data)
	if !strings.Contains(body, "POST /v1/chat/stt") {
		t.Fatal("missing POST /v1/chat/stt")
	}
	if !strings.Contains(body, "stt_local_enabled") {
		t.Fatal("capabilities should expose stt_local_enabled")
	}
	block := farmguardian.VisionContextBlock()
	if !strings.Contains(block, "Phase 67") {
		t.Fatal("vision block should mention Phase 67 field assistant")
	}
}
