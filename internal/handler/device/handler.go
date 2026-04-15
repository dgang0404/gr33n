package device

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/httputil"
	commontypes "gr33n-api/internal/platform/commontypes"
)

type Handler struct {
	q db.Querier
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{q: db.New(pool)}
}

func NewHandlerWithQuerier(q db.Querier) *Handler {
	return &Handler{q: q}
}

// GET /farms/{id}/devices
func (h *Handler) ListByFarm(w http.ResponseWriter, r *http.Request) {
	farmID, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	devices, err := h.q.ListDevicesByFarm(ctx, farmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list devices")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, devices)
}

// GET /devices/{id}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid device id")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	device, err := h.q.GetDeviceByID(ctx, id)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "device not found")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, device)
}

// POST /farms/{id}/devices
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	farmID, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	var params db.CreateDeviceParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	params.FarmID = farmID
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	device, err := h.q.CreateDevice(ctx, params)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create device")
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, device)
}

// PATCH /devices/{id}/status
func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid device id")
		return
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	device, err := h.q.UpdateDeviceStatus(ctx, db.UpdateDeviceStatusParams{
		ID:     id,
		Status: commontypes.DeviceStatusEnum(body.Status),
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update device status")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, device)
}

// DELETE /devices/{id}/pending-command
func (h *Handler) ClearPendingCommand(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid device id")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := h.q.ClearDevicePendingCommand(ctx, id); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to clear pending command")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DELETE /devices/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid device id")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	err = h.q.SoftDeleteDevice(ctx, db.SoftDeleteDeviceParams{ID: id})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to delete device")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
