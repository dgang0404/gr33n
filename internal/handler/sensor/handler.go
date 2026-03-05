package sensor

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/httputil"
)

type Handler struct{ pool *pgxpool.Pool }

func NewHandler(pool *pgxpool.Pool) *Handler { return &Handler{pool: pool} }

func (h *Handler) ListByFarm(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	q := db.New(h.pool)
	rows, err := q.ListSensorsByFarm(r.Context(), farmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if rows == nil {
		rows = []db.Gr33ncoreSensor{}
	}
	httputil.WriteJSON(w, http.StatusOK, rows)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid sensor id")
		return
	}
	q := db.New(h.pool)
	s, err := q.GetSensorByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "sensor not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, s)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	var body struct {
		ZoneID                 *int64   `json:"zone_id"`
		DeviceID               *int64   `json:"device_id"`
		Name                   string   `json:"name"`
		SensorType             string   `json:"sensor_type"`
		UnitID                 int64    `json:"unit_id"`
		HardwareIdentifier     *string  `json:"hardware_identifier"`
		ValueMinExpected       *float64 `json:"value_min_expected"`
		ValueMaxExpected       *float64 `json:"value_max_expected"`
		AlertThresholdLow      *float64 `json:"alert_threshold_low"`
		AlertThresholdHigh     *float64 `json:"alert_threshold_high"`
		ReadingIntervalSeconds *int32   `json:"reading_interval_seconds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	toNum := func(v *float64) pgtype.Numeric {
		if v == nil {
			return pgtype.Numeric{}
		}
		var n pgtype.Numeric
		_ = n.Scan(strconv.FormatFloat(*v, 'f', -1, 64))
		return n
	}
	params := db.CreateSensorParams{
		FarmID:                 farmID,
		ZoneID:                 body.ZoneID,
		DeviceID:               body.DeviceID,
		Name:                   body.Name,
		SensorType:             body.SensorType,
		UnitID:                 body.UnitID,
		HardwareIdentifier:     body.HardwareIdentifier,
		ValueMinExpected:       toNum(body.ValueMinExpected),
		ValueMaxExpected:       toNum(body.ValueMaxExpected),
		AlertThresholdLow:      toNum(body.AlertThresholdLow),
		AlertThresholdHigh:     toNum(body.AlertThresholdHigh),
		ReadingIntervalSeconds: body.ReadingIntervalSeconds,
		Config:                 []byte("{}"),
		MetaData:               []byte("{}"),
	}
	q := db.New(h.pool)
	s, err := q.CreateSensor(r.Context(), params)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, s)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid sensor id")
		return
	}
	q := db.New(h.pool)
	if err := q.SoftDeleteSensor(r.Context(), db.SoftDeleteSensorParams{
		ID:              id,
		UpdatedByUserID: pgtype.UUID{}, // zero value = NULL in DB
	}); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// LatestReading — GET /sensors/{id}/readings/latest
func (h *Handler) LatestReading(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid sensor id")
		return
	}
	q := db.New(h.pool)
	reading, err := q.GetLatestReadingBySensor(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "no readings yet")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, reading)
}

// PostReading — POST /sensors/{id}/readings
// Pi payload: { "value_raw": 22.5, "is_valid": true, "battery_level_percent": 87.0 }
func (h *Handler) PostReading(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid sensor id")
		return
	}
	var body struct {
		ReadingTime         *time.Time `json:"reading_time"`
		ValueRaw            float64    `json:"value_raw"`
		ValueText           *string    `json:"value_text"`
		BatteryLevelPercent *float64   `json:"battery_level_percent"`
		SignalStrengthDbm   *int32     `json:"signal_strength_dbm"`
		IsValid             *bool      `json:"is_valid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	ts := time.Now().UTC()
	if body.ReadingTime != nil {
		ts = *body.ReadingTime
	}
	// Default is_valid to true if not provided
	isValid := true
	if body.IsValid != nil {
		isValid = *body.IsValid
	}
	pIsValid := &isValid // *bool as required by InsertSensorReadingParams

	var valueRaw pgtype.Numeric
	_ = valueRaw.Scan(strconv.FormatFloat(body.ValueRaw, 'f', -1, 64))

	var battery pgtype.Numeric
	if body.BatteryLevelPercent != nil {
		_ = battery.Scan(strconv.FormatFloat(*body.BatteryLevelPercent, 'f', 2, 64))
	}

	params := db.InsertSensorReadingParams{
		ReadingTime:         ts,
		SensorID:            id,
		ValueRaw:            valueRaw,
		ValueText:           body.ValueText,
		ValueJson:           nil,
		BatteryLevelPercent: battery,
		SignalStrengthDbm:   body.SignalStrengthDbm,
		IsValid:             pIsValid,
		MetaData:            nil,
	}
	q := db.New(h.pool)
	reading, err := q.InsertSensorReading(r.Context(), params)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, reading)
}
