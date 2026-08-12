package device

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gr33n-api/internal/authctx"
	db "gr33n-api/internal/db"
	commontypes "gr33n-api/internal/platform/commontypes"
)

var errNotFound = errors.New("not found")

type mockQuerier struct {
	db.Querier
	updateStatusFn       func(ctx context.Context, arg db.UpdateDeviceStatusParams) (db.Gr33ncoreDevice, error)
	clearPendingCmdFn    func(ctx context.Context, id int64) error
	listDevicesByFarmFn  func(ctx context.Context, farmID int64) ([]db.Gr33ncoreDevice, error)
	getDeviceByIDFn      func(ctx context.Context, id int64) (db.Gr33ncoreDevice, error)
	listDeviceIPEventsFn func(ctx context.Context, arg db.ListDeviceIPEventsByDeviceParams) ([]db.Gr33ncoreDeviceIpEvent, error)
}

func (m *mockQuerier) GetDeviceByID(ctx context.Context, id int64) (db.Gr33ncoreDevice, error) {
	if m.getDeviceByIDFn != nil {
		return m.getDeviceByIDFn(ctx, id)
	}
	return db.Gr33ncoreDevice{ID: id}, nil
}

func (m *mockQuerier) ListDeviceIPEventsByDevice(ctx context.Context, arg db.ListDeviceIPEventsByDeviceParams) ([]db.Gr33ncoreDeviceIpEvent, error) {
	if m.listDeviceIPEventsFn != nil {
		return m.listDeviceIPEventsFn(ctx, arg)
	}
	return []db.Gr33ncoreDeviceIpEvent{}, nil
}

func (m *mockQuerier) ListDevicesByFarm(ctx context.Context, farmID int64) ([]db.Gr33ncoreDevice, error) {
	if m.listDevicesByFarmFn != nil {
		return m.listDevicesByFarmFn(ctx, farmID)
	}
	return []db.Gr33ncoreDevice{}, nil
}

func (m *mockQuerier) UpdateDeviceStatus(ctx context.Context, arg db.UpdateDeviceStatusParams) (db.Gr33ncoreDevice, error) {
	return m.updateStatusFn(ctx, arg)
}

func (m *mockQuerier) ClearDevicePendingCommand(ctx context.Context, id int64) error {
	return m.clearPendingCmdFn(ctx, id)
}

func (m *mockQuerier) InsertDeviceIPEvent(ctx context.Context, arg db.InsertDeviceIPEventParams) error {
	return nil
}

func (m *mockQuerier) UpdateDeviceIPAddress(ctx context.Context, arg db.UpdateDeviceIPAddressParams) error {
	return nil
}

func TestIPHistory_ReturnsEvents_200(t *testing.T) {
	now := time.Now().UTC()
	mq := &mockQuerier{
		getDeviceByIDFn: func(_ context.Context, id int64) (db.Gr33ncoreDevice, error) {
			return db.Gr33ncoreDevice{ID: id, FarmID: 7}, nil
		},
		listDeviceIPEventsFn: func(_ context.Context, arg db.ListDeviceIPEventsByDeviceParams) ([]db.Gr33ncoreDeviceIpEvent, error) {
			if arg.DeviceID != 1 {
				t.Fatalf("expected device id 1, got %d", arg.DeviceID)
			}
			if arg.Limit != 20 {
				t.Fatalf("expected default limit 20, got %d", arg.Limit)
			}
			return []db.Gr33ncoreDeviceIpEvent{
				{DeviceID: 1, FarmID: 7, NewIp: net.ParseIP("192.168.1.101"), ObservedAt: now},
			}, nil
		},
	}
	h := NewHandlerWithQuerier(mq)

	req := httptest.NewRequest(http.MethodGet, "/devices/1/ip-history", nil)
	req = req.WithContext(authctx.WithFarmAuthzSkip(context.Background(), true))
	rec := httptest.NewRecorder()

	h.IPHistory(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var events []db.Gr33ncoreDeviceIpEvent
	if err := json.Unmarshal(rec.Body.Bytes(), &events); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
}

func TestIPHistory_LimitQueryParam_Respected(t *testing.T) {
	mq := &mockQuerier{
		getDeviceByIDFn: func(_ context.Context, id int64) (db.Gr33ncoreDevice, error) {
			return db.Gr33ncoreDevice{ID: id, FarmID: 7}, nil
		},
		listDeviceIPEventsFn: func(_ context.Context, arg db.ListDeviceIPEventsByDeviceParams) ([]db.Gr33ncoreDeviceIpEvent, error) {
			if arg.Limit != 5 {
				t.Fatalf("expected limit 5, got %d", arg.Limit)
			}
			return []db.Gr33ncoreDeviceIpEvent{}, nil
		},
	}
	h := NewHandlerWithQuerier(mq)

	req := httptest.NewRequest(http.MethodGet, "/devices/1/ip-history?limit=5", nil)
	req = req.WithContext(authctx.WithFarmAuthzSkip(context.Background(), true))
	rec := httptest.NewRecorder()

	h.IPHistory(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestIPHistory_DeviceNotFound_404(t *testing.T) {
	mq := &mockQuerier{
		getDeviceByIDFn: func(_ context.Context, id int64) (db.Gr33ncoreDevice, error) {
			return db.Gr33ncoreDevice{}, errNotFound
		},
	}
	h := NewHandlerWithQuerier(mq)

	req := httptest.NewRequest(http.MethodGet, "/devices/999/ip-history", nil)
	req = req.WithContext(authctx.WithFarmAuthzSkip(context.Background(), true))
	rec := httptest.NewRecorder()

	h.IPHistory(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateStatus_WithLastConfigFetchAt_200(t *testing.T) {
	var gotFetch *string
	mq := &mockQuerier{
		updateStatusFn: func(_ context.Context, arg db.UpdateDeviceStatusParams) (db.Gr33ncoreDevice, error) {
			gotFetch = &arg.Column3
			return db.Gr33ncoreDevice{
				ID:        arg.ID,
				Status:    arg.Status,
				Config:    []byte(`{"last_config_fetch_at":"2026-06-08T12:00:00Z"}`),
				MetaData:  []byte("{}"),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		},
	}
	h := NewHandlerWithQuerier(mq)

	body, _ := json.Marshal(map[string]string{
		"status":               "online",
		"last_config_fetch_at": "2026-06-08T12:00:00Z",
	})
	req := httptest.NewRequest(http.MethodPatch, "/devices/1/status", bytes.NewReader(body))
	req = req.WithContext(authctx.WithPiEdgeAuth(context.Background()))
	rec := httptest.NewRecorder()

	h.UpdateStatus(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if gotFetch == nil || *gotFetch != "2026-06-08T12:00:00Z" {
		t.Fatalf("expected last_config_fetch_at passthrough, got %v", gotFetch)
	}
}

func TestUpdateStatus_ValidBody_200(t *testing.T) {
	mq := &mockQuerier{
		updateStatusFn: func(_ context.Context, arg db.UpdateDeviceStatusParams) (db.Gr33ncoreDevice, error) {
			return db.Gr33ncoreDevice{
				ID:        arg.ID,
				Name:      "test-device",
				Status:    arg.Status,
				Config:    []byte("{}"),
				MetaData:  []byte("{}"),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		},
	}
	h := NewHandlerWithQuerier(mq)

	body, _ := json.Marshal(map[string]string{"status": "online"})
	req := httptest.NewRequest(http.MethodPatch, "/devices/1/status", bytes.NewReader(body))
	req = req.WithContext(authctx.WithPiEdgeAuth(context.Background()))
	rec := httptest.NewRecorder()

	h.UpdateStatus(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["status"] != string(commontypes.DeviceStatusOnline) {
		t.Fatalf("expected status=online, got %v", resp["status"])
	}
}

func TestUpdateStatus_InvalidBody_400(t *testing.T) {
	h := NewHandlerWithQuerier(&mockQuerier{})

	req := httptest.NewRequest(http.MethodPatch, "/devices/1/status", bytes.NewReader([]byte("bad")))
	req = req.WithContext(authctx.WithPiEdgeAuth(context.Background()))
	rec := httptest.NewRecorder()

	h.UpdateStatus(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestUpdateStatus_InvalidID_400(t *testing.T) {
	h := NewHandlerWithQuerier(&mockQuerier{})

	body, _ := json.Marshal(map[string]string{"status": "online"})
	req := httptest.NewRequest(http.MethodPatch, "/devices/abc/status", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.UpdateStatus(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestClearPendingCommand_204(t *testing.T) {
	mq := &mockQuerier{
		clearPendingCmdFn: func(_ context.Context, id int64) error {
			return nil
		},
	}
	h := NewHandlerWithQuerier(mq)

	req := httptest.NewRequest(http.MethodDelete, "/devices/1/pending-command", nil)
	req = req.WithContext(authctx.WithPiEdgeAuth(context.Background()))
	rec := httptest.NewRecorder()

	h.ClearPendingCommand(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestClearPendingCommand_InvalidID_400(t *testing.T) {
	h := NewHandlerWithQuerier(&mockQuerier{})

	req := httptest.NewRequest(http.MethodDelete, "/devices/xyz/pending-command", nil)
	rec := httptest.NewRecorder()

	h.ClearPendingCommand(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestListByFarm_PiEdgeAuth_200(t *testing.T) {
	mq := &mockQuerier{
		listDevicesByFarmFn: func(_ context.Context, farmID int64) ([]db.Gr33ncoreDevice, error) {
			if farmID != 9 {
				t.Fatalf("unexpected farm id %d", farmID)
			}
			return []db.Gr33ncoreDevice{
				{ID: 1, Name: "edge-gateway", FarmID: 9, Config: []byte(`{"pending_command":{"command":"on"}}`)},
			}, nil
		},
	}
	h := NewHandlerWithQuerier(mq)
	req := httptest.NewRequest(http.MethodGet, "/farms/9/devices", nil)
	req = req.WithContext(authctx.WithPiEdgeAuth(context.Background()))
	rec := httptest.NewRecorder()
	h.ListByFarm(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}
