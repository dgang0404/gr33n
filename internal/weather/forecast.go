package weather

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/platform/commontypes"
)

// ForecastStatus is the stable connection state for UI badges.
type ForecastStatus string

const (
	StatusDisabled       ForecastStatus = "disabled"
	StatusNoCoords       ForecastStatus = "no_coords"
	StatusConnected      ForecastStatus = "connected"
	StatusCached         ForecastStatus = "cached"
	StatusCachedStale    ForecastStatus = "cached_stale"
	StatusOffline        ForecastStatus = "offline"
	StatusMisconfigured  ForecastStatus = "misconfigured"
)

// OnlineForecast is the site-weather online_forecast block.
type OnlineForecast struct {
	Status        ForecastStatus `json:"status"`
	Provider      string         `json:"provider"`
	ProviderLabel string         `json:"provider_label"`
	Enabled       bool           `json:"enabled"`
	OptedIn       bool           `json:"opted_in"`
	FetchedAt     *time.Time     `json:"fetched_at,omitempty"`
	Stale         bool           `json:"stale"`
	Message       string         `json:"message"`
	Current       map[string]any `json:"current,omitempty"`
	TonightLowC   *float64       `json:"tonight_low_celsius,omitempty"`
	FrostRisk     bool           `json:"frost_risk"`
}

// FarmForecastOptedIn reads meta_data.weather_forecast_enabled.
func FarmForecastOptedIn(meta json.RawMessage) bool {
	if len(meta) == 0 {
		return false
	}
	var m map[string]any
	if err := json.Unmarshal(meta, &m); err != nil {
		return false
	}
	v, ok := m["weather_forecast_enabled"].(bool)
	return ok && v
}

// ResolveOnlineForecast fetches or serves cached API weather for a farm.
func ResolveOnlineForecast(ctx context.Context, q db.Querier, cfg Config, farmID int64, lat, lng float64, coordsOK bool, optedIn bool) (OnlineForecast, *db.Gr33ncoreWeatherDatum, error) {
	out := OnlineForecast{
		Status:        StatusDisabled,
		Provider:      string(cfg.Provider),
		ProviderLabel: cfg.Label(),
		Enabled:       cfg.Available(),
		OptedIn:       optedIn,
		Message:       "Forecast off",
	}

	if cfg.Provider == ProviderOff || !cfg.Available() {
		if cfg.Misconfigured() {
			out.Status = StatusMisconfigured
			out.Message = "Forecast misconfigured — check API keys on server"
		}
		return out, nil, nil
	}
	if !optedIn {
		out.Message = "Enable live forecast in Settings"
		return out, nil, nil
	}
	if !coordsOK {
		out.Status = StatusNoCoords
		out.Message = "Set farm location for forecast"
		return out, nil, nil
	}

	cached, cacheErr := q.GetLatestAPIWeatherForFarm(ctx, farmID)
	hasCache := cacheErr == nil
	if cacheErr != nil && !errors.Is(cacheErr, pgx.ErrNoRows) {
		return out, nil, cacheErr
	}

	needsFetch := !hasCache || time.Since(cached.RecordedAt) >= cfg.CacheTTL
	if needsFetch {
		snap, fetchErr := fetchForProvider(ctx, cfg, lat, lng)
		if fetchErr == nil && snap != nil {
			row, insErr := insertSnapshot(ctx, q, farmID, snap)
			if insErr != nil {
				return out, nil, insErr
			}
			return forecastFromRow(cfg, row, StatusConnected, "Live forecast", false), &row, nil
		}
		if hasCache {
			st := StatusCachedStale
			msg := "Cached forecast (offline)"
			return forecastFromRow(cfg, cached, st, msg, true), &cached, nil
		}
		out.Status = StatusOffline
		out.Message = "Forecast offline"
		return out, nil, nil
	}

	st := StatusCached
	msg := "Forecast cached"
	if time.Since(cached.RecordedAt) < 2*time.Minute {
		st = StatusConnected
		msg = "Live forecast"
	}
	return forecastFromRow(cfg, cached, st, msg, false), &cached, nil
}

func fetchForProvider(ctx context.Context, cfg Config, lat, lng float64) (*Snapshot, error) {
	switch cfg.Provider {
	case ProviderOpenMeteo:
		return FetchOpenMeteo(ctx, lat, lng, cfg.FetchTimeout)
	default:
		return nil, errors.New("weather provider not implemented")
	}
}

func insertSnapshot(ctx context.Context, q db.Querier, farmID int64, snap *Snapshot) (db.Gr33ncoreWeatherDatum, error) {
	src := commontypes.WeatherDataSourceAPIOpenMeteo
	switch snap.Provider {
	case ProviderOpenWeather:
		src = commontypes.WeatherDataSourceAPIOpenWeather
	case ProviderVisualCrossing:
		src = commontypes.WeatherDataSourceAPIVisualCross
	}
	return q.InsertWeatherData(ctx, db.InsertWeatherDataParams{
		FarmID:             farmID,
		RecordedAt:         snap.FetchedAt.UTC(),
		DataSource:         src,
		TemperatureCelsius: numericFromFloat(snap.TemperatureC),
		HumidityPercent:    numericFromFloat(snap.HumidityPercent),
		PrecipitationMm:    numericFromFloat(snap.PrecipitationMm),
		WindSpeedMs:        numericFromFloat(snap.WindSpeedMs),
		CloudCoverPercent:  numericFromFloat(snap.CloudCoverPercent),
		ForecastData:       snap.ForecastJSON,
		RawData:            snap.RawJSON,
	})
}

func forecastFromRow(cfg Config, row db.Gr33ncoreWeatherDatum, status ForecastStatus, message string, stale bool) OnlineForecast {
	out := OnlineForecast{
		Status:        status,
		Provider:      string(cfg.Provider),
		ProviderLabel: cfg.Label(),
		Enabled:       true,
		OptedIn:       true,
		Stale:         stale,
		Message:       message,
	}
	t := row.RecordedAt
	out.FetchedAt = &t

	current := map[string]any{}
	if v := numericToFloat(row.TemperatureCelsius); v != nil {
		current["temperature_celsius"] = *v
	}
	if v := numericToFloat(row.HumidityPercent); v != nil {
		current["humidity_percent"] = *v
	}
	if v := numericToFloat(row.CloudCoverPercent); v != nil {
		current["cloud_cover_percent"] = *v
	}
	if v := numericToFloat(row.WindSpeedMs); v != nil {
		current["wind_speed_ms"] = *v
	}
	if len(current) > 0 {
		out.Current = current
	}

	if len(row.ForecastData) > 0 {
		var payload struct {
			Daily struct {
				Temperature2mMin []float64 `json:"temperature_2m_min"`
			} `json:"daily"`
		}
		if err := json.Unmarshal(row.ForecastData, &payload); err == nil && len(payload.Daily.Temperature2mMin) > 0 {
			low := payload.Daily.Temperature2mMin[0]
			out.TonightLowC = &low
			out.FrostRisk = low < FrostRiskCelsius
		}
	}
	return out
}

func numericFromFloat(v *float64) pgtype.Numeric {
	if v == nil {
		return pgtype.Numeric{Valid: false}
	}
	var n pgtype.Numeric
	_ = n.Scan(strconv.FormatFloat(*v, 'f', -1, 64))
	return n
}

func numericToFloat(n pgtype.Numeric) *float64 {
	if !n.Valid {
		return nil
	}
	f, err := n.Float64Value()
	if err != nil || !f.Valid {
		return nil
	}
	v := f.Float64
	return &v
}

// AppendForecastTier adds online_forecast to tiers when applicable.
func AppendForecastTier(tiers []string, status ForecastStatus) []string {
	switch status {
	case StatusConnected, StatusCached, StatusCachedStale:
		for _, t := range tiers {
			if t == "online_forecast" {
				return tiers
			}
		}
		return append(tiers, "online_forecast")
	default:
		return tiers
	}
}
