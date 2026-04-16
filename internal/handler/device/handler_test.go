package device

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gr33n-api/internal/authctx"
	db "gr33n-api/internal/db"
	commontypes "gr33n-api/internal/platform/commontypes"
)

type mockQuerier struct {
	db.Querier
	updateStatusFn      func(ctx context.Context, arg db.UpdateDeviceStatusParams) (db.Gr33ncoreDevice, error)
	clearPendingCmdFn   func(ctx context.Context, id int64) error
	listDevicesByFarmFn func(ctx context.Context, farmID int64) ([]db.Gr33ncoreDevice, error)
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
