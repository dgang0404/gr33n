package weather

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseOpenMeteo(t *testing.T) {
	body := []byte(`{
		"current": {
			"time": "2026-07-12T18:30",
			"temperature_2m": 24.2,
			"relative_humidity_2m": 58,
			"cloud_cover": 35,
			"precipitation": 0,
			"wind_speed_10m": 2.1
		},
		"daily": {
			"time": ["2026-07-12", "2026-07-13"],
			"temperature_2m_min": [1.5, 10.2],
			"temperature_2m_max": [28.1, 27.5],
			"precipitation_sum": [0, 2.1]
		}
	}`)
	snap, err := parseOpenMeteo(body)
	if err != nil {
		t.Fatal(err)
	}
	if snap.TemperatureC == nil || *snap.TemperatureC != 24.2 {
		t.Fatalf("temp got %v", snap.TemperatureC)
	}
	if snap.CloudCoverPercent == nil || *snap.CloudCoverPercent != 35 {
		t.Fatalf("cloud got %v", snap.CloudCoverPercent)
	}
	if snap.TonightLowC == nil || *snap.TonightLowC != 1.5 {
		t.Fatalf("tonight low got %v", snap.TonightLowC)
	}
	if !snap.FrostRisk {
		t.Fatal("expected frost risk")
	}
}

func TestFetchOpenMeteo_HTTP(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/forecast" {
			t.Fatalf("path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"current": {"time":"2026-07-12T12:00","temperature_2m":20,"relative_humidity_2m":50,"cloud_cover":10,"precipitation":0,"wind_speed_10m":1},
			"daily": {"time":["2026-07-12"],"temperature_2m_min":[8],"temperature_2m_max":[22],"precipitation_sum":[0]}
		}`))
	}))
	defer srv.Close()

	old := openMeteoBaseURL
	openMeteoBaseURL = srv.URL + "/v1/forecast"
	defer func() { openMeteoBaseURL = old }()

	snap, err := FetchOpenMeteo(context.Background(), 40.89, -81.41, 2e9)
	if err != nil {
		t.Fatal(err)
	}
	if snap.TemperatureC == nil || *snap.TemperatureC != 20 {
		t.Fatalf("got %v", snap.TemperatureC)
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	t.Setenv("WEATHER_PROVIDER", "openmeteo")
	t.Setenv("WEATHER_CACHE_MINUTES", "15")
	cfg := LoadConfigFromEnv()
	if cfg.Provider != ProviderOpenMeteo {
		t.Fatalf("provider %s", cfg.Provider)
	}
	if !cfg.Available() {
		t.Fatal("expected available")
	}
	if cfg.CacheTTL != 15*60*1e9 {
		t.Fatalf("cache %v", cfg.CacheTTL)
	}
}

func TestFarmForecastOptedIn(t *testing.T) {
	if FarmForecastOptedIn(nil) {
		t.Fatal("nil meta")
	}
	raw, _ := json.Marshal(map[string]any{"weather_forecast_enabled": true})
	if !FarmForecastOptedIn(raw) {
		t.Fatal("expected opted in")
	}
}
