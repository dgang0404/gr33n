package sensor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "gr33n-api/internal/db"
)

type mockQuerier struct {
	db.Querier
	insertReadingFn func(ctx context.Context, arg db.InsertSensorReadingParams) (db.Gr33ncoreSensorReading, error)
}

func (m *mockQuerier) InsertSensorReading(ctx context.Context, arg db.InsertSensorReadingParams) (db.Gr33ncoreSensorReading, error) {
	return m.insertReadingFn(ctx, arg)
}

func (m *mockQuerier) GetSensorByID(_ context.Context, _ int64) (db.Gr33ncoreSensor, error) {
	return db.Gr33ncoreSensor{}, fmt.Errorf("not found")
}

func (m *mockQuerier) GetRecentUnacknowledgedAlertForSource(_ context.Context, _ db.GetRecentUnacknowledgedAlertForSourceParams) (int64, error) {
	return 0, fmt.Errorf("not found")
}

func (m *mockQuerier) CreateAlert(_ context.Context, _ db.CreateAlertParams) (db.Gr33ncoreAlertsNotification, error) {
	return db.Gr33ncoreAlertsNotification{}, nil
}

type noopSSE struct{}

func (noopSSE) Notify() {}

func TestPostReading_ValidBody_201(t *testing.T) {
	mq := &mockQuerier{
		insertReadingFn: func(_ context.Context, arg db.InsertSensorReadingParams) (db.Gr33ncoreSensorReading, error) {
			return db.Gr33ncoreSensorReading{
				SensorID:    arg.SensorID,
				ReadingTime: time.Now().UTC(),
				ValueRaw:    arg.ValueRaw,
			}, nil
		},
	}
	h := NewHandlerWithQuerier(mq, noopSSE{})

	body, _ := json.Marshal(map[string]any{"value_raw": 22.5, "is_valid": true})
	req := httptest.NewRequest(http.MethodPost, "/sensors/1/readings", bytes.NewReader(body))
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	h.PostReading(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["sensor_id"] == nil {
		t.Fatal("response missing sensor_id")
	}
}

func TestPostReading_InvalidJSON_400(t *testing.T) {
	h := NewHandlerWithQuerier(&mockQuerier{}, noopSSE{})

	req := httptest.NewRequest(http.MethodPost, "/sensors/1/readings", bytes.NewReader([]byte("not json")))
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	h.PostReading(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestPostReading_InvalidSensorID_400(t *testing.T) {
	h := NewHandlerWithQuerier(&mockQuerier{}, noopSSE{})

	body, _ := json.Marshal(map[string]any{"value_raw": 22.5})
	req := httptest.NewRequest(http.MethodPost, "/sensors/abc/readings", bytes.NewReader(body))
	req.SetPathValue("id", "abc")
	rec := httptest.NewRecorder()

	h.PostReading(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestPostReading_SSENotified(t *testing.T) {
	notified := false
	sse := &mockSSE{fn: func() { notified = true }}
	mq := &mockQuerier{
		insertReadingFn: func(_ context.Context, arg db.InsertSensorReadingParams) (db.Gr33ncoreSensorReading, error) {
			return db.Gr33ncoreSensorReading{
				SensorID:    arg.SensorID,
				ReadingTime: time.Now().UTC(),
				ValueRaw:    pgtype.Numeric{},
			}, nil
		},
	}
	h := NewHandlerWithQuerier(mq, sse)

	body, _ := json.Marshal(map[string]any{"value_raw": 22.5})
	req := httptest.NewRequest(http.MethodPost, "/sensors/1/readings", bytes.NewReader(body))
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	h.PostReading(rec, req)

	if !notified {
		t.Fatal("expected SSE.Notify() to be called")
	}
}

type mockSSE struct {
	fn func()
}

func (m *mockSSE) Notify() { m.fn() }
