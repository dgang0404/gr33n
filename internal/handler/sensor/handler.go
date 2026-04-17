package sensor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/httputil"
)

type SSENotifier interface {
	Notify()
}

// FarmAlertPusher sends mobile push for farm-scoped alerts (optional; nil disables).
type FarmAlertPusher interface {
	DispatchFarmAlert(ctx context.Context, alert db.Gr33ncoreAlertsNotification)
}

// batchReadingIngest is the request shape for POST /sensors/readings/batch items.
type batchReadingIngest struct {
	SensorID            int64      `json:"sensor_id"`
	ReadingTime         *time.Time `json:"reading_time"`
	ValueRaw            float64    `json:"value_raw"`
	ValueText           *string    `json:"value_text"`
	BatteryLevelPercent *float64   `json:"battery_level_percent"`
	SignalStrengthDbm   *int32     `json:"signal_strength_dbm"`
	IsValid             *bool      `json:"is_valid"`
}

type Handler struct {
	q    db.Querier
	sse  SSENotifier
	push FarmAlertPusher
	pool *pgxpool.Pool
}

func NewHandler(pool *pgxpool.Pool, sse SSENotifier, push FarmAlertPusher) *Handler {
	return &Handler{q: db.New(pool), sse: sse, push: push, pool: pool}
}

func NewHandlerWithQuerier(q db.Querier, sse SSENotifier, push FarmAlertPusher) *Handler {
	return &Handler{q: q, sse: sse, push: push, pool: nil}
}

func (h *Handler) ListByFarm(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, farmID) {
		return
	}
	rows, err := h.q.ListSensorsByFarm(r.Context(), farmID)
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
	s, err := h.q.GetSensorByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "sensor not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, s.FarmID) {
		return
	}
	httputil.WriteJSON(w, http.StatusOK, s)
}

// Duration + cooldown bounds (mirrored in CHECK constraints on gr33ncore.sensors).
const (
	maxAlertDurationSeconds int32 = 86400  // 24h
	maxAlertCooldownSeconds int32 = 604800 // 7d
)

func clampAlertSeconds(v *int32, max int32, fallback int32) int32 {
	if v == nil {
		return fallback
	}
	if *v < 0 {
		return 0
	}
	if *v > max {
		return max
	}
	return *v
}

func floatToNumeric(v *float64) pgtype.Numeric {
	if v == nil {
		return pgtype.Numeric{}
	}
	var n pgtype.Numeric
	_ = n.Scan(strconv.FormatFloat(*v, 'f', -1, 64))
	return n
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, farmID) {
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
		AlertDurationSeconds   *int32   `json:"alert_duration_seconds"`
		AlertCooldownSeconds   *int32   `json:"alert_cooldown_seconds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	params := db.CreateSensorParams{
		FarmID:                 farmID,
		ZoneID:                 body.ZoneID,
		DeviceID:               body.DeviceID,
		Name:                   body.Name,
		SensorType:             body.SensorType,
		UnitID:                 body.UnitID,
		HardwareIdentifier:     body.HardwareIdentifier,
		ValueMinExpected:       floatToNumeric(body.ValueMinExpected),
		ValueMaxExpected:       floatToNumeric(body.ValueMaxExpected),
		AlertThresholdLow:      floatToNumeric(body.AlertThresholdLow),
		AlertThresholdHigh:     floatToNumeric(body.AlertThresholdHigh),
		ReadingIntervalSeconds: body.ReadingIntervalSeconds,
		AlertDurationSeconds:   clampAlertSeconds(body.AlertDurationSeconds, maxAlertDurationSeconds, 0),
		AlertCooldownSeconds:   clampAlertSeconds(body.AlertCooldownSeconds, maxAlertCooldownSeconds, 300),
		Config:                 []byte("{}"),
		MetaData:               []byte("{}"),
	}
	s, err := h.q.CreateSensor(r.Context(), params)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, s)
}

// Update — PUT /sensors/{id}
// Patch-style: any field omitted from the body leaves the stored value unchanged.
// Callers that want to null out a nullable threshold can pass JSON null explicitly.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid sensor id")
		return
	}
	existing, err := h.q.GetSensorByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "sensor not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, existing.FarmID) {
		return
	}

	// json.RawMessage-free patch: presence of key is enough — we parse into a map
	// and only forward fields that were present. This lets the client pass `null`
	// to clear a threshold and omit the field entirely to leave it untouched.
	var raw map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}

	params := db.UpdateSensorParams{ID: id}

	decodeInt64Ptr := func(key string, dst **int64) error {
		if v, ok := raw[key]; ok {
			if string(v) == "null" {
				*dst = nil
				return nil
			}
			var n int64
			if err := json.Unmarshal(v, &n); err != nil {
				return fmt.Errorf("invalid %s", key)
			}
			*dst = &n
		}
		return nil
	}
	decodeStringPtr := func(key string, dst **string) error {
		if v, ok := raw[key]; ok {
			if string(v) == "null" {
				*dst = nil
				return nil
			}
			var s string
			if err := json.Unmarshal(v, &s); err != nil {
				return fmt.Errorf("invalid %s", key)
			}
			*dst = &s
		}
		return nil
	}
	decodeInt32Ptr := func(key string, dst **int32, max int32) error {
		if v, ok := raw[key]; ok {
			if string(v) == "null" {
				*dst = nil
				return nil
			}
			var n int32
			if err := json.Unmarshal(v, &n); err != nil {
				return fmt.Errorf("invalid %s", key)
			}
			if n < 0 {
				n = 0
			}
			if max > 0 && n > max {
				n = max
			}
			*dst = &n
		}
		return nil
	}
	decodeNumeric := func(key string, dst *pgtype.Numeric) error {
		if v, ok := raw[key]; ok {
			if string(v) == "null" {
				*dst = pgtype.Numeric{} // invalid ⇒ SQL NULL via sqlc.narg
				return nil
			}
			var f float64
			if err := json.Unmarshal(v, &f); err != nil {
				return fmt.Errorf("invalid %s", key)
			}
			var n pgtype.Numeric
			if err := n.Scan(strconv.FormatFloat(f, 'f', -1, 64)); err != nil {
				return fmt.Errorf("invalid %s", key)
			}
			*dst = n
		}
		return nil
	}

	for _, fn := range []func() error{
		func() error { return decodeInt64Ptr("zone_id", &params.ZoneID) },
		func() error { return decodeInt64Ptr("device_id", &params.DeviceID) },
		func() error { return decodeStringPtr("name", &params.Name) },
		func() error { return decodeStringPtr("sensor_type", &params.SensorType) },
		func() error { return decodeInt64Ptr("unit_id", &params.UnitID) },
		func() error { return decodeStringPtr("hardware_identifier", &params.HardwareIdentifier) },
		func() error { return decodeNumeric("value_min_expected", &params.ValueMinExpected) },
		func() error { return decodeNumeric("value_max_expected", &params.ValueMaxExpected) },
		func() error { return decodeNumeric("alert_threshold_low", &params.AlertThresholdLow) },
		func() error { return decodeNumeric("alert_threshold_high", &params.AlertThresholdHigh) },
		func() error { return decodeInt32Ptr("reading_interval_seconds", &params.ReadingIntervalSeconds, 0) },
		func() error {
			return decodeInt32Ptr("alert_duration_seconds", &params.AlertDurationSeconds, maxAlertDurationSeconds)
		},
		func() error {
			return decodeInt32Ptr("alert_cooldown_seconds", &params.AlertCooldownSeconds, maxAlertCooldownSeconds)
		},
	} {
		if err := fn(); err != nil {
			httputil.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	updated, err := h.q.UpdateSensor(r.Context(), params)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, updated)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid sensor id")
		return
	}
	s0, err := h.q.GetSensorByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "sensor not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, s0.FarmID) {
		return
	}
	if err := h.q.SoftDeleteSensor(r.Context(), db.SoftDeleteSensorParams{
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
	s, err := h.q.GetSensorByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "sensor not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, s.FarmID) {
		return
	}
	reading, err := h.q.GetLatestReadingBySensor(r.Context(), id)
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

const (
	defaultReadingListLimit = 500
	maxReadingListLimit     = 5000
	maxBatchReadings        = 64
	defaultReadingRange     = 24 * time.Hour
)

func parseRFC3339OrNano(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, errors.New("empty time")
	}
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t.UTC(), nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}, err
	}
	return t.UTC(), nil
}

func readingTimeRange(r *http.Request) (since, until time.Time, err error) {
	q := r.URL.Query()
	until = time.Now().UTC()
	if u := q.Get("until"); u != "" {
		until, err = parseRFC3339OrNano(u)
		if err != nil {
			return
		}
	}
	since = until.Add(-defaultReadingRange)
	if s := q.Get("since"); s != "" {
		since, err = parseRFC3339OrNano(s)
		if err != nil {
			return
		}
	}
	if since.After(until) {
		err = errors.New("since must be before or equal to until")
	}
	return since, until, err
}

func parseReadingListLimit(r *http.Request) (int, error) {
	ls := r.URL.Query().Get("limit")
	if ls == "" {
		return defaultReadingListLimit, nil
	}
	n, err := strconv.Atoi(ls)
	if err != nil || n < 1 {
		return 0, errors.New("invalid limit")
	}
	if n > maxReadingListLimit {
		n = maxReadingListLimit
	}
	return n, nil
}

func reverseReadingsInPlace(s []db.Gr33ncoreSensorReading) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

// ListReadings — GET /sensors/{id}/readings?since=&until=&limit=
func (h *Handler) ListReadings(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid sensor id")
		return
	}
	s, err := h.q.GetSensorByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "sensor not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, s.FarmID) {
		return
	}
	since, until, err := readingTimeRange(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	limit, err := parseReadingListLimit(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	rows, err := h.q.ListReadingsBySensorAndTimeRange(r.Context(), db.ListReadingsBySensorAndTimeRangeParams{
		SensorID:      id,
		ReadingTime:   since,
		ReadingTime_2: until,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if rows == nil {
		rows = []db.Gr33ncoreSensorReading{}
	}
	if len(rows) > limit {
		rows = rows[:limit]
	}
	reverseReadingsInPlace(rows)
	httputil.WriteJSON(w, http.StatusOK, rows)
}

type sensorReadingStatsResponse struct {
	Count            int64      `json:"count"`
	Avg              float64    `json:"avg"`
	Min              *float64   `json:"min"`
	Max              *float64   `json:"max"`
	FirstReadingTime *time.Time `json:"first_reading_time"`
	LastReadingTime  *time.Time `json:"last_reading_time"`
}

func statsNumericPtr(v interface{}) *float64 {
	if v == nil {
		return nil
	}
	switch x := v.(type) {
	case float64:
		return &x
	case float32:
		f := float64(x)
		return &f
	case []byte:
		f, err := strconv.ParseFloat(string(x), 64)
		if err != nil {
			return nil
		}
		return &f
	case string:
		f, err := strconv.ParseFloat(x, 64)
		if err != nil {
			return nil
		}
		return &f
	default:
		f, err := strconv.ParseFloat(fmt.Sprint(x), 64)
		if err != nil {
			return nil
		}
		return &f
	}
}

func statsTimePtr(v interface{}) *time.Time {
	if v == nil {
		return nil
	}
	switch x := v.(type) {
	case time.Time:
		if x.IsZero() {
			return nil
		}
		u := x.UTC()
		return &u
	case pgtype.Timestamptz:
		if !x.Valid {
			return nil
		}
		u := x.Time.UTC()
		return &u
	default:
		return nil
	}
}

// ReadingStats — GET /sensors/{id}/readings/stats?since=&until=
func (h *Handler) ReadingStats(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid sensor id")
		return
	}
	s, err := h.q.GetSensorByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "sensor not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, s.FarmID) {
		return
	}
	since, until, err := readingTimeRange(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	row, err := h.q.GetSensorReadingStats(r.Context(), db.GetSensorReadingStatsParams{
		SensorID:      id,
		ReadingTime:   since,
		ReadingTime_2: until,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	avg := statsNumericPtr(row.AvgValue)
	avgVal := 0.0
	if avg != nil {
		avgVal = *avg
	}
	out := sensorReadingStatsResponse{
		Count: row.TotalReadings,
		Avg:   avgVal,
		Min:   statsNumericPtr(row.MinValue),
		Max:   statsNumericPtr(row.MaxValue),
	}
	out.FirstReadingTime = statsTimePtr(row.FirstReading)
	out.LastReadingTime = statsTimePtr(row.LastReading)
	httputil.WriteJSON(w, http.StatusOK, out)
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
	reading, err := h.q.InsertSensorReading(r.Context(), params)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if h.sse != nil {
		h.sse.Notify()
	}

	// Run the evaluator in the background with its own context — r.Context() is cancelled
	// as soon as the HTTP handler returns, which was racing with the goroutine's SQL calls.
	go h.evaluateThresholds(context.Background(), id, body.ValueRaw)

	httputil.WriteJSON(w, http.StatusCreated, reading)
}

// PostReadingsBatch — POST /sensors/readings/batch
// Compact multi-reading ingest for MQTT bridges and microcontrollers (one HTTP round-trip).
// Body: JSON array of { "sensor_id", "value_raw", optional "reading_time", "is_valid", ... }.
func (h *Handler) PostReadingsBatch(w http.ResponseWriter, r *http.Request) {
	var items []batchReadingIngest
	if err := json.NewDecoder(r.Body).Decode(&items); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if len(items) == 0 {
		httputil.WriteError(w, http.StatusBadRequest, "empty batch")
		return
	}
	if len(items) > maxBatchReadings {
		httputil.WriteError(w, http.StatusBadRequest,
			fmt.Sprintf("at most %d readings per request", maxBatchReadings))
		return
	}
	for i := range items {
		if items[i].SensorID < 1 {
			httputil.WriteError(w, http.StatusBadRequest, "each item requires sensor_id")
			return
		}
	}

	ctx := r.Context()
	var out []db.Gr33ncoreSensorReading
	var err error
	if h.pool != nil {
		var tx pgx.Tx
		tx, err = h.pool.Begin(ctx)
		if err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
		defer tx.Rollback(ctx)
		qtx := db.New(tx)
		out, err = h.insertSensorReadingBatch(ctx, qtx, items)
		if err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if err := tx.Commit(ctx); err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
	} else {
		out, err = h.insertSensorReadingBatch(ctx, h.q, items)
		if err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	if h.sse != nil {
		h.sse.Notify()
	}
	for _, item := range items {
		go h.evaluateThresholds(context.Background(), item.SensorID, item.ValueRaw)
	}

	httputil.WriteJSON(w, http.StatusCreated, map[string]any{
		"inserted": len(out),
		"readings": out,
	})
}

func (h *Handler) insertSensorReadingBatch(ctx context.Context, q db.Querier, items []batchReadingIngest) ([]db.Gr33ncoreSensorReading, error) {
	out := make([]db.Gr33ncoreSensorReading, 0, len(items))
	for _, item := range items {
		ts := time.Now().UTC()
		if item.ReadingTime != nil {
			ts = *item.ReadingTime
		}
		isValid := true
		if item.IsValid != nil {
			isValid = *item.IsValid
		}
		pIsValid := &isValid

		var valueRaw pgtype.Numeric
		_ = valueRaw.Scan(strconv.FormatFloat(item.ValueRaw, 'f', -1, 64))

		var battery pgtype.Numeric
		if item.BatteryLevelPercent != nil {
			_ = battery.Scan(strconv.FormatFloat(*item.BatteryLevelPercent, 'f', 2, 64))
		}

		row, err := q.InsertSensorReading(ctx, db.InsertSensorReadingParams{
			ReadingTime:         ts,
			SensorID:            item.SensorID,
			ValueRaw:            valueRaw,
			ValueText:           item.ValueText,
			ValueJson:           nil,
			BatteryLevelPercent: battery,
			SignalStrengthDbm:   item.SignalStrengthDbm,
			IsValid:             pIsValid,
			MetaData:            nil,
		})
		if err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, nil
}

func numericToFloat64(n pgtype.Numeric) (float64, bool) {
	if !n.Valid {
		return 0, false
	}
	f, err := n.Float64Value()
	if err != nil || !f.Valid {
		return 0, false
	}
	return f.Float64, true
}

// evaluateThresholds implements the Phase 19 duration + cooldown state machine.
//
// Per sensor:
//   - If the reading is in bounds: clear any existing breach-start timestamp and return.
//   - If the reading is out of bounds:
//   - If alert_breach_started_at is NULL, stamp it with "now" (marks the start of the streak).
//   - If the streak has lasted < alert_duration_seconds, return (still accumulating evidence).
//   - If the last alert for this source fired < alert_cooldown_seconds ago, return (suppressed).
//   - Otherwise, create the alert. Keep alert_breach_started_at set so subsequent readings
//     remain suppressed until the cooldown window elapses or the reading returns to bounds.
func (h *Handler) evaluateThresholds(ctx context.Context, sensorID int64, valueRaw float64) {
	sensor, err := h.q.GetSensorByID(ctx, sensorID)
	if err != nil {
		return
	}

	lo, hasLo := numericToFloat64(sensor.AlertThresholdLow)
	hi, hasHi := numericToFloat64(sensor.AlertThresholdHigh)
	if !hasLo && !hasHi {
		return
	}

	breach := false
	var msg string
	if hasLo && valueRaw < lo {
		breach = true
		msg = fmt.Sprintf("Value %.1f below low threshold %.1f", valueRaw, lo)
	} else if hasHi && valueRaw > hi {
		breach = true
		msg = fmt.Sprintf("Value %.1f exceeds high threshold %.1f", valueRaw, hi)
	}

	now := time.Now().UTC()

	if !breach {
		// Reading returned to bounds — reset the streak so the next excursion starts fresh.
		if sensor.AlertBreachStartedAt.Valid {
			if err := h.q.ClearSensorAlertBreachStart(ctx, sensorID); err != nil {
				log.Printf("alert: failed to clear breach start for sensor %d: %v", sensorID, err)
			}
		}
		return
	}

	// Breach ongoing — ensure we have a streak start timestamp.
	breachStart := now
	if sensor.AlertBreachStartedAt.Valid {
		breachStart = sensor.AlertBreachStartedAt.Time
	} else {
		if err := h.q.SetSensorAlertBreachStart(ctx, db.SetSensorAlertBreachStartParams{
			ID:                   sensorID,
			AlertBreachStartedAt: pgtype.Timestamptz{Time: now, Valid: true},
		}); err != nil {
			log.Printf("alert: failed to set breach start for sensor %d: %v", sensorID, err)
			// Not fatal — we still evaluate duration/cooldown below using breachStart=now.
		}
	}

	// Gate 1: sustained-breach duration.
	if sensor.AlertDurationSeconds > 0 {
		elapsed := now.Sub(breachStart)
		if elapsed < time.Duration(sensor.AlertDurationSeconds)*time.Second {
			return
		}
	}

	srcType := "sensor_reading"

	// Gate 2: per-sensor cooldown since the last alert for this source.
	if sensor.AlertCooldownSeconds > 0 {
		lastCreated, err := h.q.GetLatestAlertCreatedAtForSource(ctx, db.GetLatestAlertCreatedAtForSourceParams{
			FarmID:                    sensor.FarmID,
			TriggeringEventSourceType: &srcType,
			TriggeringEventSourceID:   &sensorID,
		})
		if err == nil {
			cooldown := time.Duration(sensor.AlertCooldownSeconds) * time.Second
			if now.Sub(lastCreated) < cooldown {
				return
			}
		}
	}

	severity := db.Gr33ncoreNotificationPriorityEnumHigh
	if hasLo && hasHi {
		rangeSpan := hi - lo
		if rangeSpan > 0 {
			deviation := math.Max(lo-valueRaw, valueRaw-hi)
			if deviation > rangeSpan*0.5 {
				severity = db.Gr33ncoreNotificationPriorityEnumCritical
			}
		}
	}

	subject := fmt.Sprintf("Sensor '%s' threshold breach", sensor.Name)
	created, err := h.q.CreateAlert(ctx, db.CreateAlertParams{
		FarmID:                    sensor.FarmID,
		TriggeringEventSourceType: &srcType,
		TriggeringEventSourceID:   &sensorID,
		Severity: db.NullGr33ncoreNotificationPriorityEnum{
			Gr33ncoreNotificationPriorityEnum: severity,
			Valid:                             true,
		},
		SubjectRendered:     &subject,
		MessageTextRendered: &msg,
	})
	if err != nil {
		log.Printf("alert: failed to create for sensor %d: %v", sensorID, err)
		return
	}
	if h.push != nil {
		a := created
		go func() {
			c, cancel := context.WithTimeout(context.Background(), 45*time.Second)
			defer cancel()
			h.push.DispatchFarmAlert(c, a)
		}()
	}
}
