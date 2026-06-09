// Phase 66 — site weather routes and solar engine.
package main

import (
	"os"
	"strings"
	"testing"
	"time"

	"gr33n-api/internal/farmguardian"
	"gr33n-api/internal/solar"
)

func TestPhase66_SiteWeatherRouteRegistered(t *testing.T) {
	data, err := os.ReadFile("routes.go")
	if err != nil {
		t.Fatalf("read routes.go: %v", err)
	}
	if !strings.Contains(string(data), "GET /farms/{id}/site-weather") {
		t.Fatal("missing site-weather route")
	}
}

func TestPhase66_SolarEngineAndReadTool(t *testing.T) {
	tz, _ := time.LoadLocation("America/Los_Angeles")
	day := solar.SolarForDate(45.52, -122.68, tz, time.Date(2026, 6, 8, 0, 0, 0, 0, tz))
	if day.DaylengthHours < 10 {
		t.Fatalf("unexpected daylength %.2f", day.DaylengthHours)
	}
	found := false
	for _, id := range farmguardian.ReadToolIDs() {
		if id == "site_weather" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("site_weather missing from ReadToolIDs")
	}
}
