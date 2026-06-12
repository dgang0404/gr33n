package plants

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/httputil"
	"gr33n-api/internal/plantcatalog"
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
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	plant, _, status, errMsg := plantcatalog.CreateFromRequest(r.Context(), h.q, farmID, body)
	if errMsg != "" {
		httputil.WriteError(w, status, errMsg)
		return
	}
	plantcatalog.WriteCreateResponse(w, plant, status)
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
		CropProfileID     *int64          `json:"crop_profile_id"`
		CropKey           *string         `json:"crop_key"`
		Meta              json.RawMessage `json:"meta"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if body.CropKey != nil || body.CropProfileID != nil {
		httputil.WriteError(w, http.StatusBadRequest, "crop_key and crop_profile_id cannot be changed; create uses catalog slot")
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

	displayName := existing.DisplayName
	if existing.CropKey != nil && strings.TrimSpace(*existing.CropKey) != "" {
		displayName = existing.DisplayName
	} else if strings.TrimSpace(body.DisplayName) != "" {
		displayName = strings.TrimSpace(body.DisplayName)
	} else if strings.TrimSpace(body.DisplayName) == "" && existing.CropKey == nil {
		httputil.WriteError(w, http.StatusBadRequest, "display_name required")
		return
	}

	variety := existing.VarietyOrCultivar
	if body.VarietyOrCultivar != nil {
		v := strings.TrimSpace(*body.VarietyOrCultivar)
		if v == "" {
			variety = nil
		} else {
			variety = &v
		}
	}

	row, err := h.q.UpdatePlantVariety(r.Context(), db.UpdatePlantVarietyParams{
		ID:                id,
		VarietyOrCultivar: variety,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !bytes.Equal(meta, existing.Meta) {
		row, err = h.q.UpdatePlant(r.Context(), db.UpdatePlantParams{
			ID:                id,
			DisplayName:       displayName,
			VarietyOrCultivar: variety,
			CropProfileID:     existing.CropProfileID,
			CropKey:           existing.CropKey,
			Meta:              meta,
		})
		if err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
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
