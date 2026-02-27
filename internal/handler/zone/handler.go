package zone

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

// GET /farms/{id}/zones
func (h *Handler) ListByFarm(w http.ResponseWriter, r *http.Request) {
	farmID, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	zones, err := h.q.ListZonesByFarm(ctx, farmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list zones")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, zones)
}

// GET /zones/{id}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid zone id")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	zone, err := h.q.GetZoneByID(ctx, id)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "zone not found")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, zone)
}

// POST /farms/{id}/zones
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	farmID, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	var params db.CreateZoneParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	params.FarmID = farmID
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	zone, err := h.q.CreateZone(ctx, params)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create zone")
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, zone)
}

// PUT /zones/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid zone id")
		return
	}
	var params db.UpdateZoneParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	params.ID = id
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	zone, err := h.q.UpdateZone(ctx, params)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update zone")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, zone)
}

// DELETE /zones/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid zone id")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	err = h.q.SoftDeleteZone(ctx, id)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to delete zone")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
