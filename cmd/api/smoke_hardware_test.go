//go:build hardware

// Phase 33 WS4 — @hardware CI lane.
//
// These tests drive REAL GPIO / edge hardware (a Pi wired to the seeded
// demo-veg-relay-01) and are EXCLUDED from the default build. The Makefile and
// CI run `go test -tags dev ./...`, which does not set the `hardware` tag, so
// this file is never compiled there.
//
// Run on a Pi/bench with the API up and PI_API_KEY set:
//
//	GR33N_HARDWARE_TEST=1 go test -tags 'dev hardware' -run Hardware ./cmd/api/ -count=1 -v
package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestHardwareActuatorGPIORoundTrip(t *testing.T) {
	if os.Getenv("GR33N_HARDWARE_TEST") != "1" {
		t.Skip("live GPIO bench — set GR33N_HARDWARE_TEST=1 (needs a Pi + seeded demo-veg-relay-01 + running API)")
	}

	script := filepath.Join("..", "..", "scripts", "run-edge-actuator-smoke.sh")
	if _, err := os.Stat(script); err != nil {
		t.Fatalf("edge actuator smoke script missing (%s): %v", script, err)
	}

	// Drives pending_command → pi_client → actuator_events → clear pending on
	// real hardware. Fails the lane on any non-zero exit.
	cmd := exec.Command("bash", script, "--direct", "--command", "on")
	out, err := cmd.CombinedOutput()
	t.Logf("run-edge-actuator-smoke.sh output:\n%s", out)
	if err != nil {
		t.Fatalf("hardware actuator GPIO round-trip failed: %v", err)
	}
}
