package sensor

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/httputil"
)

type Handler struct {
	q *db.Queries
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{q: db.New(pool)}
}

// GET /farms/{id}/sensors
func (h *Handler) ListByFarm(w http.ResponseWriter, r *http.Request) {
	farmID, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	sensors, err := h.q.ListSensorsByFarm(ctx, farmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list sensors")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, sensors)
}

// GET /sensors/{id}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid sensor id")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	sensor, err := h.q.GetSensorByID(ctx, id)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "sensor not found")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, sensor)
}

// POST /farms/{id}/sensors
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	farmID, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	var params db.CreateSensorParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	params.FarmID = farmID
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	sensor, err := h.q.CreateSensor(ctx, params)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create sensor")
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, sensor)
}

// DELETE /sensors/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid sensor id")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	err = h.q.SoftDeleteSensor(ctx, id)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to delete sensor")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
