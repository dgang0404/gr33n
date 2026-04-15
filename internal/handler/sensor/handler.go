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

type Handler struct {
	q   db.Querier
	sse SSENotifier
}

func NewHandler(pool *pgxpool.Pool, sse SSENotifier) *Handler {
	return &Handler{q: db.New(pool), sse: sse}
}

func NewHandlerWithQuerier(q db.Querier, sse SSENotifier) *Handler {
	return &Handler{q: q, sse: sse}
}

func (h *Handler) ListByFarm(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
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
	httputil.WriteJSON(w, http.StatusOK, s)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, farmID) {
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
	s, err := h.q.CreateSensor(r.Context(), params)
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
	s0, err := h.q.GetSensorByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "sensor not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, s0.FarmID) {
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
	if _, err := h.q.GetSensorByID(r.Context(), id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "sensor not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
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
	if _, err := h.q.GetSensorByID(r.Context(), id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "sensor not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
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

	go h.evaluateThresholds(r.Context(), id, body.ValueRaw)

	httputil.WriteJSON(w, http.StatusCreated, reading)
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
	if !breach {
		return
	}

	srcType := "sensor_reading"
	_, err = h.q.GetRecentUnacknowledgedAlertForSource(ctx, db.GetRecentUnacknowledgedAlertForSourceParams{
		FarmID:                    sensor.FarmID,
		TriggeringEventSourceType: &srcType,
		TriggeringEventSourceID:   &sensorID,
	})
	if err == nil {
		return // recent unacknowledged alert exists, skip
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
	_, err = h.q.CreateAlert(ctx, db.CreateAlertParams{
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
	}
}
