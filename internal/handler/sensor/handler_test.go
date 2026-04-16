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
	h := NewHandlerWithQuerier(mq, noopSSE{}, nil)

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
	h := NewHandlerWithQuerier(&mockQuerier{}, noopSSE{}, nil)

	req := httptest.NewRequest(http.MethodPost, "/sensors/1/readings", bytes.NewReader([]byte("not json")))
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	h.PostReading(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestPostReading_InvalidSensorID_400(t *testing.T) {
	h := NewHandlerWithQuerier(&mockQuerier{}, noopSSE{}, nil)

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
	h := NewHandlerWithQuerier(mq, sse, nil)

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

func TestPostReadingsBatch_Valid_201(t *testing.T) {
	var calls int
	mq := &mockQuerier{
		insertReadingFn: func(_ context.Context, arg db.InsertSensorReadingParams) (db.Gr33ncoreSensorReading, error) {
			calls++
			return db.Gr33ncoreSensorReading{
				SensorID:    arg.SensorID,
				ReadingTime: arg.ReadingTime,
				ValueRaw:    arg.ValueRaw,
			}, nil
		},
	}
	h := NewHandlerWithQuerier(mq, noopSSE{}, nil)

	body, _ := json.Marshal([]map[string]any{
		{"sensor_id": 1, "value_raw": 21.0},
		{"sensor_id": 2, "value_raw": 55.2},
	})
	req := httptest.NewRequest(http.MethodPost, "/sensors/readings/batch", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.PostReadingsBatch(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	if calls != 2 {
		t.Fatalf("expected 2 inserts, got %d", calls)
	}
	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if int(resp["inserted"].(float64)) != 2 {
		t.Fatalf("expected inserted=2, got %v", resp["inserted"])
	}
}

func TestPostReadingsBatch_Empty_400(t *testing.T) {
	h := NewHandlerWithQuerier(&mockQuerier{}, noopSSE{}, nil)
	body, _ := json.Marshal([]map[string]any{})
	req := httptest.NewRequest(http.MethodPost, "/sensors/readings/batch", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.PostReadingsBatch(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestPostReadingsBatch_MissingSensorID_400(t *testing.T) {
	h := NewHandlerWithQuerier(&mockQuerier{}, noopSSE{}, nil)
	body, _ := json.Marshal([]map[string]any{{"value_raw": 1.0}})
	req := httptest.NewRequest(http.MethodPost, "/sensors/readings/batch", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.PostReadingsBatch(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestPostReadingsBatch_SSENotified(t *testing.T) {
	notified := false
	sse := &mockSSE{fn: func() { notified = true }}
	mq := &mockQuerier{
		insertReadingFn: func(_ context.Context, arg db.InsertSensorReadingParams) (db.Gr33ncoreSensorReading, error) {
			return db.Gr33ncoreSensorReading{
				SensorID:    arg.SensorID,
				ReadingTime: arg.ReadingTime,
				ValueRaw:    arg.ValueRaw,
			}, nil
		},
	}
	h := NewHandlerWithQuerier(mq, sse, nil)
	body, _ := json.Marshal([]map[string]any{{"sensor_id": 3, "value_raw": 7.0}})
	req := httptest.NewRequest(http.MethodPost, "/sensors/readings/batch", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.PostReadingsBatch(rec, req)
	if !notified {
		t.Fatal("expected SSE notify once for batch")
	}
}
