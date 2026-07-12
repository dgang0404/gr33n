// Phase 178 — online weather forecast routes and provider config.
package main

import (
	"os"
	"strings"
	"testing"

	wxsvc "gr33n-api/internal/weather"
)

func TestPhase178_WeatherForecastRoutesRegistered(t *testing.T) {
	data, err := os.ReadFile("routes.go")
	if err != nil {
		t.Fatalf("read routes.go: %v", err)
	}
	body := string(data)
	if !strings.Contains(body, "PATCH /farms/{id}/weather/settings") {
		t.Fatal("missing weather settings route")
	}
	if !strings.Contains(body, "weather_forecast_available") {
		t.Fatal("missing weather_forecast_available in capabilities")
	}
}

func TestPhase178_OpenMeteoConfigAvailable(t *testing.T) {
	t.Setenv("WEATHER_PROVIDER", "openmeteo")
	cfg := wxsvc.LoadConfigFromEnv()
	if !cfg.Available() {
		t.Fatal("openmeteo should be available without API key")
	}
	if cfg.Label() == "" {
		t.Fatal("expected provider label")
	}
}
