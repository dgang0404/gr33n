// Package weather — Phase 178 online forecast providers (Open-Meteo default).
package weather

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Provider identifies the configured online weather backend.
type Provider string

const (
	ProviderOff             Provider = "off"
	ProviderOpenMeteo       Provider = "openmeteo"
	ProviderOpenWeather     Provider = "openweather"
	ProviderVisualCrossing  Provider = "visualcrossing"
)

// Config holds API-level weather provider settings (env only).
type Config struct {
	Provider             Provider
	OpenWeatherAPIKey    string
	VisualCrossingAPIKey string
	CacheTTL             time.Duration
	FetchTimeout         time.Duration
}

// LoadConfigFromEnv reads WEATHER_* environment variables.
func LoadConfigFromEnv() Config {
	cfg := Config{
		Provider:     Provider(strings.ToLower(strings.TrimSpace(os.Getenv("WEATHER_PROVIDER")))),
		OpenWeatherAPIKey:    strings.TrimSpace(os.Getenv("OPENWEATHER_API_KEY")),
		VisualCrossingAPIKey: strings.TrimSpace(os.Getenv("VISUALCROSSING_API_KEY")),
		CacheTTL:     30 * time.Minute,
		FetchTimeout: 8 * time.Second,
	}
	if cfg.Provider == "" {
		cfg.Provider = ProviderOff
	}
	if v := strings.TrimSpace(os.Getenv("WEATHER_CACHE_MINUTES")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.CacheTTL = time.Duration(n) * time.Minute
		}
	}
	if v := strings.TrimSpace(os.Getenv("WEATHER_FETCH_TIMEOUT_SEC")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.FetchTimeout = time.Duration(n) * time.Second
		}
	}
	return cfg
}

// Available reports whether the API can attempt online forecast fetches.
func (c Config) Available() bool {
	switch c.Provider {
	case ProviderOpenMeteo:
		return true
	case ProviderOpenWeather:
		return c.OpenWeatherAPIKey != ""
	case ProviderVisualCrossing:
		return c.VisualCrossingAPIKey != ""
	default:
		return false
	}
}

// Label returns operator-facing provider name.
func (c Config) Label() string {
	switch c.Provider {
	case ProviderOpenMeteo:
		return "Open-Meteo (free)"
	case ProviderOpenWeather:
		return "OpenWeather"
	case ProviderVisualCrossing:
		return "Visual Crossing"
	default:
		return ""
	}
}

// Misconfigured is true when a provider is set but required credentials are missing.
func (c Config) Misconfigured() bool {
	switch c.Provider {
	case ProviderOpenWeather:
		return c.OpenWeatherAPIKey == ""
	case ProviderVisualCrossing:
		return c.VisualCrossingAPIKey == ""
	default:
		return false
	}
}

// FrostRiskCelsius — overnight low below this triggers frost flag.
const FrostRiskCelsius = 2.0
