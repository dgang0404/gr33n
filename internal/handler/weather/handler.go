// Package weather — Phase 66 site solar + manual weather ingestion.
package weather

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/platform/commontypes"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/httputil"
	"gr33n-api/internal/solar"
)

type Handler struct {
	q db.Querier
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{q: db.New(pool)}
}

// GET /farms/{id}/site-weather?date=YYYY-MM-DD
func (h *Handler) GetSiteWeather(w http.ResponseWriter, r *http.Request) {
	farmID, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, farmID) {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := buildSiteWeatherResponse(ctx, h.q, farmID, r.URL.Query().Get("date"))
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load site weather")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

// POST /farms/{id}/weather/manual
func (h *Handler) PostManual(w http.ResponseWriter, r *http.Request) {
	farmID, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmAdmin(w, r, h.q, farmID) {
		return
	}
	var req struct {
		TemperatureCelsius *float64 `json:"temperature_celsius"`
		HumidityPercent    *float64 `json:"humidity_percent"`
		CloudCoverPercent  *float64 `json:"cloud_cover_percent"`
		Notes              *string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	raw := map[string]any{}
	if req.Notes != nil {
		raw["notes"] = strings.TrimSpace(*req.Notes)
	}
	rawJSON, _ := json.Marshal(raw)

	row, err := h.q.InsertWeatherData(ctx, db.InsertWeatherDataParams{
		FarmID:     farmID,
		RecordedAt: time.Now().UTC(),
		DataSource: commontypes.WeatherDataSourceManual,
		TemperatureCelsius: numericFromFloat(req.TemperatureCelsius),
		HumidityPercent:    numericFromFloat(req.HumidityPercent),
		CloudCoverPercent:  numericFromFloat(req.CloudCoverPercent),
		RawData:            rawJSON,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to save weather entry")
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, row)
}

func buildSiteWeatherResponse(ctx context.Context, q db.Querier, farmID int64, dateStr string) (map[string]any, error) {
	site, err := q.GetFarmSiteCoords(ctx, farmID)
	if err != nil {
		return nil, err
	}
	tzName := strings.TrimSpace(site.Timezone)
	if tzName == "" {
		tzName = "UTC"
	}
	tz, err := time.LoadLocation(tzName)
	if err != nil {
		tz = time.UTC
	}
	date := time.Now().In(tz)
	if dateStr != "" {
		if parsed, perr := time.ParseInLocation("2006-01-02", dateStr, tz); perr == nil {
			date = parsed
		}
	}

	out := map[string]any{
		"farm_id":  farmID,
		"timezone": tzName,
		"tiers":    []string{},
	}

	lat, latOK := ifaceFloat(site.Latitude)
	lng, lngOK := ifaceFloat(site.Longitude)
	if latOK && lngOK {
		day := solar.SolarForDate(lat, lng, tz, date)
		out["coordinates"] = map[string]any{
			"latitude":    lat,
			"longitude":   lng,
			"elevation_m": ifaceFloatOrNil(site.ElevationM),
		}
		out["solar"] = map[string]any{
			"date":                  date.Format("2006-01-02"),
			"sunrise":               day.Sunrise.Format(time.RFC3339),
			"sunset":                day.Sunset.Format(time.RFC3339),
			"solar_noon":            day.SolarNoon.Format(time.RFC3339),
			"daylength_hours":       round2(day.DaylengthHours),
			"clear_sky_dli":         round2(day.ClearSkyDLI),
			"max_sun_elevation_deg": round2(day.MaxSunElevationDeg),
			"tier":                  "solar_math",
		}
		tiers := []string{"solar_math"}
		out["tiers"] = tiers
	} else {
		out["solar"] = nil
		out["coordinates"] = nil
	}

	latest, err := q.GetLatestWeatherForFarm(ctx, farmID)
	if err == nil {
		src := string(latest.DataSource)
		entry := map[string]any{
			"recorded_at":          latest.RecordedAt,
			"data_source":          src,
			"temperature_celsius":  numericToFloat(latest.TemperatureCelsius),
			"humidity_percent":     numericToFloat(latest.HumidityPercent),
			"cloud_cover_percent":  numericToFloat(latest.CloudCoverPercent),
			"solar_radiation_wm2":  numericToFloat(latest.SolarRadiationWm2),
		}
		out["latest_reading"] = entry
		tiers, _ := out["tiers"].([]string)
		if src == "manual_entry" || src == "iot_sensor_reading" || src == "farm_weather_station" {
			tiers = append(tiers, "local_sensor")
		} else if strings.HasPrefix(src, "api_") {
			tiers = append(tiers, "online_forecast")
		}
		out["tiers"] = tiers
	} else if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	return out, nil
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

func ifaceFloatOrNil(v any) any {
	f, ok := ifaceFloat(v)
	if !ok {
		return nil
	}
	return f
}

func round2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}
