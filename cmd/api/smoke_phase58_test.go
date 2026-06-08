// Phase 58 — farm task-consumptions list endpoint.
package main

import (
	"os"
	"strings"
	"testing"
)

func TestPhase58_TaskConsumptionRouteRegistered(t *testing.T) {
	data, err := os.ReadFile("routes.go")
	if err != nil {
		t.Fatalf("read routes.go: %v", err)
	}
	if !strings.Contains(string(data), "GET /farms/{id}/task-consumptions") {
		t.Fatal("missing GET /farms/{id}/task-consumptions route")
	}
}
