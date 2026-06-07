package sensor

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gr33n-api/internal/authctx"
	db "gr33n-api/internal/db"
	"gr33n-api/internal/hardware"
)

func reqWithAuthSkip(req *http.Request) *http.Request {
	return req.WithContext(authctx.WithFarmAuthzSkip(context.Background(), true))
}

type wiringMockQuerier struct {
	db.Querier
	sensor   db.Gr33ncoreSensor
	updated  json.RawMessage
	updateFn func(ctx context.Context, id int64, config json.RawMessage) (db.Gr33ncoreSensor, error)
}

func (m *wiringMockQuerier) GetSensorByID(_ context.Context, _ int64) (db.Gr33ncoreSensor, error) {
	return m.sensor, nil
}

func (m *wiringMockQuerier) UpdateSensorConfig(ctx context.Context, id int64, config json.RawMessage) (db.Gr33ncoreSensor, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, config)
	}
	m.updated = config
	row := m.sensor
	row.Config = config
	return row, nil
}

func TestPatchWiringRejectsInvalidSource(t *testing.T) {
	mq := &wiringMockQuerier{sensor: db.Gr33ncoreSensor{ID: 1, FarmID: 1, Config: json.RawMessage(`{}`)}}
	h := NewHandlerWithQuerier(mq, nil, nil)
	body := `{"wiring":{"source":"unknown","gpio_pin":4}}`
	req := httptest.NewRequest(http.MethodPatch, "/sensors/1/wiring", bytes.NewBufferString(body))
	req.SetPathValue("id", "1")
	req = reqWithAuthSkip(req)
	w := httptest.NewRecorder()
	h.PatchWiring(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status %d, body %s", w.Code, w.Body.String())
	}
}

func TestPatchWiringMergesConfig(t *testing.T) {
	pin := 4
	dev := int64(2)
	wiring := hardware.Wiring{Source: "dht22", GPIOPin: &pin, DeviceID: &dev}
	cfg, _ := hardware.MergeWiring(json.RawMessage(`{"notes":"grow"}`), &wiring)
	mq := &wiringMockQuerier{
		sensor: db.Gr33ncoreSensor{ID: 3, FarmID: 1, Config: json.RawMessage(`{"notes":"grow"}`)},
	}
	h := NewHandlerWithQuerier(mq, nil, nil)
	payload, _ := json.Marshal(map[string]any{
		"wiring": map[string]any{"source": "dht22", "gpio_pin": 4, "device_id": 2},
	})
	req := httptest.NewRequest(http.MethodPatch, "/sensors/3/wiring", bytes.NewReader(payload))
	req.SetPathValue("id", "3")
	req = reqWithAuthSkip(req)
	w := httptest.NewRecorder()
	h.PatchWiring(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d, body %s", w.Code, w.Body.String())
	}
	got, _ := hardware.ExtractWiring(mq.updated)
	if got == nil || got.Source != "dht22" || got.GPIOPin == nil || *got.GPIOPin != 4 {
		t.Fatalf("updated config %+v", mq.updated)
	}
	_ = cfg
}
