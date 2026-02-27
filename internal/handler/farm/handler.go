package farm

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
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

// GET /farms?user_id=<uuid>
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		httputil.WriteError(w, http.StatusBadRequest, "user_id query param required")
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid user_id")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	farms, err := h.q.ListFarmsForUser(ctx, userID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list farms")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, farms)
}

// GET /farms/{id}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	farm, err := h.q.GetFarmByID(ctx, id)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "farm not found")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, farm)
}

// POST /farms
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var params db.CreateFarmParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	farm, err := h.q.CreateFarm(ctx, params)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create farm")
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, farm)
}

// PUT /farms/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	var params db.UpdateFarmParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	params.ID = id
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	farm, err := h.q.UpdateFarm(ctx, params)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update farm")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, farm)
}

// DELETE /farms/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	var body struct {
		UpdatedByUserID uuid.UUID `json:"updated_by_user_id"`
	}
	json.NewDecoder(r.Body).Decode(&body)
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	err = h.q.SoftDeleteFarm(ctx, db.SoftDeleteFarmParams{
		ID:              id,
		UpdatedByUserID: &body.UpdatedByUserID,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to delete farm")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
