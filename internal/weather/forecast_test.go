package weather

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/platform/commontypes"
)

type forecastStoreStub struct {
	cacheRow  db.Gr33ncoreWeatherDatum
	cacheErr  error
	insertErr error
}

func (s *forecastStoreStub) GetLatestAPIWeatherForFarm(_ context.Context, _ int64) (db.Gr33ncoreWeatherDatum, error) {
	return s.cacheRow, s.cacheErr
}

func (s *forecastStoreStub) InsertWeatherData(_ context.Context, _ db.InsertWeatherDataParams) (db.Gr33ncoreWeatherDatum, error) {
	if s.insertErr != nil {
		return db.Gr33ncoreWeatherDatum{}, s.insertErr
	}
	return db.Gr33ncoreWeatherDatum{}, nil
}

func TestResolveOnlineForecast_cacheReadErrorDegrades(t *testing.T) {
	old := openMeteoBaseURL
	openMeteoBaseURL = "http://127.0.0.1:1/unreachable"
	defer func() { openMeteoBaseURL = old }()

	cfg := Config{Provider: ProviderOpenMeteo, CacheTTL: time.Hour, FetchTimeout: 50 * time.Millisecond}
	stub := &forecastStoreStub{cacheErr: errors.New("db timeout")}
	out, row, err := resolveOnlineForecast(context.Background(), stub, cfg, 1, 40.89, -81.41, true, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if row != nil {
		t.Fatal("expected no row")
	}
	if out.Status != StatusOffline {
		t.Fatalf("status %s want offline", out.Status)
	}
}

func TestResolveOnlineForecast_insertFailureServesLiveFetch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"current": {"time":"2026-07-12T12:00","temperature_2m":20,"relative_humidity_2m":50,"cloud_cover":10,"precipitation":0,"wind_speed_10m":1},
			"daily": {"time":["2026-07-12"],"temperature_2m_min":[8],"temperature_2m_max":[22],"precipitation_sum":[0]}
		}`))
	}))
	defer srv.Close()

	old := openMeteoBaseURL
	openMeteoBaseURL = srv.URL
	defer func() { openMeteoBaseURL = old }()

	cfg := Config{Provider: ProviderOpenMeteo, CacheTTL: time.Hour, FetchTimeout: 2 * time.Second}
	stub := &forecastStoreStub{cacheErr: pgx.ErrNoRows, insertErr: errors.New("disk full")}
	out, row, err := resolveOnlineForecast(context.Background(), stub, cfg, 1, 40.89, -81.41, true, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if row != nil {
		t.Fatal("expected no persisted row")
	}
	if out.Status != StatusConnected {
		t.Fatalf("status %s want connected", out.Status)
	}
	if out.Current == nil || out.Current["temperature_celsius"] == nil {
		t.Fatalf("expected current conditions: %#v", out.Current)
	}
}

func TestResolveOnlineForecast_usesCachedWhenFetchFails(t *testing.T) {
	old := openMeteoBaseURL
	openMeteoBaseURL = "http://127.0.0.1:1/unreachable"
	defer func() { openMeteoBaseURL = old }()

	cfg := Config{Provider: ProviderOpenMeteo, CacheTTL: time.Hour, FetchTimeout: 50 * time.Millisecond}
	cached := db.Gr33ncoreWeatherDatum{
		RecordedAt:         time.Now().Add(-2 * time.Hour),
		DataSource:         commontypes.WeatherDataSourceAPIOpenMeteo,
		TemperatureCelsius: mustNumeric(22.5),
	}
	stub := &forecastStoreStub{cacheRow: cached}
	out, _, err := resolveOnlineForecast(context.Background(), stub, cfg, 1, 40.89, -81.41, true, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Status != StatusCachedStale {
		t.Fatalf("status %s want cached_stale", out.Status)
	}
	if out.Stale {
		// ok
	} else {
		t.Fatal("expected stale=true")
	}
}

func mustNumeric(f float64) pgtype.Numeric {
	var n pgtype.Numeric
	_ = n.Scan(f)
	return n
}
