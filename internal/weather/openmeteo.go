package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

var openMeteoBaseURL = "https://api.open-meteo.com/v1/forecast"

// Snapshot is a normalized provider reading for storage and API responses.
type Snapshot struct {
	Provider          Provider
	FetchedAt         time.Time
	TemperatureC      *float64
	HumidityPercent   *float64
	CloudCoverPercent *float64
	PrecipitationMm   *float64
	WindSpeedMs       *float64
	TonightLowC       *float64
	FrostRisk         bool
	ForecastJSON      []byte
	RawJSON           []byte
}

// FetchOpenMeteo loads current conditions and tonight's low for coordinates.
func FetchOpenMeteo(ctx context.Context, lat, lng float64, timeout time.Duration) (*Snapshot, error) {
	if timeout <= 0 {
		timeout = 8 * time.Second
	}
	u, err := url.Parse(openMeteoBaseURL)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("latitude", fmt.Sprintf("%.6f", lat))
	q.Set("longitude", fmt.Sprintf("%.6f", lng))
	q.Set("current", "temperature_2m,relative_humidity_2m,cloud_cover,precipitation,wind_speed_10m")
	q.Set("daily", "temperature_2m_min,temperature_2m_max,precipitation_sum")
	q.Set("timezone", "auto")
	q.Set("forecast_days", "2")
	q.Set("wind_speed_unit", "ms")
	u.RawQuery = q.Encode()

	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("open-meteo: HTTP %d", resp.StatusCode)
	}
	return parseOpenMeteo(body)
}

func parseOpenMeteo(body []byte) (*Snapshot, error) {
	var payload struct {
		Current struct {
			Time                 string   `json:"time"`
			Temperature2m        *float64 `json:"temperature_2m"`
			RelativeHumidity2m   *float64 `json:"relative_humidity_2m"`
			CloudCover           *float64 `json:"cloud_cover"`
			Precipitation        *float64 `json:"precipitation"`
			WindSpeed10m         *float64 `json:"wind_speed_10m"`
		} `json:"current"`
		Daily struct {
			Time               []string   `json:"time"`
			Temperature2mMin  []float64  `json:"temperature_2m_min"`
			Temperature2mMax  []float64  `json:"temperature_2m_max"`
			PrecipitationSum   []float64  `json:"precipitation_sum"`
		} `json:"daily"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	fetchedAt := time.Now().UTC()
	if payload.Current.Time != "" {
		if t, err := time.Parse("2006-01-02T15:04", payload.Current.Time); err == nil {
			fetchedAt = t.UTC()
		}
	}

	var tonightLow *float64
	if len(payload.Daily.Temperature2mMin) > 0 {
		v := payload.Daily.Temperature2mMin[0]
		tonightLow = &v
	}

	snap := &Snapshot{
		Provider:          ProviderOpenMeteo,
		FetchedAt:         fetchedAt,
		TemperatureC:      payload.Current.Temperature2m,
		HumidityPercent:   payload.Current.RelativeHumidity2m,
		CloudCoverPercent: payload.Current.CloudCover,
		PrecipitationMm:   payload.Current.Precipitation,
		WindSpeedMs:       payload.Current.WindSpeed10m,
		TonightLowC:       tonightLow,
		ForecastJSON:      body,
		RawJSON:           body,
	}
	if tonightLow != nil && *tonightLow < FrostRiskCelsius {
		snap.FrostRisk = true
	}
	return snap, nil
}
