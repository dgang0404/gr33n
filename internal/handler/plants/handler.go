package plants

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/httputil"
)

type Handler struct{ q *db.Queries }

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{q: db.New(pool)}
}

// List — GET /farms/{id}/plants
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, farmID) {
		return
	}
	rows, err := h.q.ListPlantsByFarm(r.Context(), farmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if rows == nil {
		rows = []db.Gr33ncropsPlant{}
	}
	httputil.WriteJSON(w, http.StatusOK, rows)
}

// Get — GET /plants/{id}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid plant id")
		return
	}
	row, err := h.q.GetPlant(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "plant not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, row.FarmID) {
		return
	}
	httputil.WriteJSON(w, http.StatusOK, row)
}

// Create — POST /farms/{id}/plants
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, farmID) {
		return
	}
	var body struct {
		DisplayName       string          `json:"display_name"`
		VarietyOrCultivar *string         `json:"variety_or_cultivar"`
		Meta              json.RawMessage `json:"meta"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	name := strings.TrimSpace(body.DisplayName)
	if name == "" {
		httputil.WriteError(w, http.StatusBadRequest, "display_name required")
		return
	}
	meta := []byte("{}")
	if len(body.Meta) > 0 {
		if !json.Valid(body.Meta) {
			httputil.WriteError(w, http.StatusBadRequest, "meta must be valid JSON")
			return
		}
		meta = body.Meta
	}
	row, err := h.q.CreatePlant(r.Context(), db.CreatePlantParams{
		FarmID:            farmID,
		DisplayName:       name,
		VarietyOrCultivar: body.VarietyOrCultivar,
		Meta:              meta,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, row)
}

// Update — PUT /plants/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid plant id")
		return
	}
	existing, err := h.q.GetPlant(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "plant not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, existing.FarmID) {
		return
	}
	var body struct {
		DisplayName       string          `json:"display_name"`
		VarietyOrCultivar *string         `json:"variety_or_cultivar"`
		Meta              json.RawMessage `json:"meta"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	name := strings.TrimSpace(body.DisplayName)
	if name == "" {
		httputil.WriteError(w, http.StatusBadRequest, "display_name required")
		return
	}
	meta := existing.Meta
	if len(body.Meta) > 0 {
		if !json.Valid(body.Meta) {
			httputil.WriteError(w, http.StatusBadRequest, "meta must be valid JSON")
			return
		}
		meta = body.Meta
	}
	row, err := h.q.UpdatePlant(r.Context(), db.UpdatePlantParams{
		ID:                id,
		DisplayName:       name,
		VarietyOrCultivar: body.VarietyOrCultivar,
		Meta:              meta,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, row)
}

// Delete — DELETE /plants/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid plant id")
		return
	}
	existing, err := h.q.GetPlant(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "plant not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, existing.FarmID) {
		return
	}
	if err := h.q.SoftDeletePlant(r.Context(), id); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
