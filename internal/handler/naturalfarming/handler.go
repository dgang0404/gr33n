package naturalfarming

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5/pgtype"
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

// GET /farms/{id}/naturalfarming/inputs
func (h *Handler) ListInputs(w http.ResponseWriter, r *http.Request) {
	farmID, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	rows, err := h.q.ListInputDefinitionsByFarm(r.Context(), farmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list input definitions")
		return
	}
	if rows == nil {
		rows = []db.Gr33nnaturalfarmingInputDefinition{}
	}
	httputil.WriteJSON(w, http.StatusOK, rows)
}

// GET /farms/{id}/naturalfarming/batches
func (h *Handler) ListBatches(w http.ResponseWriter, r *http.Request) {
	farmID, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	rows, err := h.q.ListInputBatchesByFarm(r.Context(), farmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list input batches")
		return
	}
	if rows == nil {
		rows = []db.Gr33nnaturalfarmingInputBatch{}
	}
	httputil.WriteJSON(w, http.StatusOK, rows)
}

func farmIDFromPath(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}

func resourceIDFromPath(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}

// POST /farms/{id}/naturalfarming/inputs
func (h *Handler) CreateInputDefinition(w http.ResponseWriter, r *http.Request) {
	farmID, err := farmIDFromPath(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	var req struct {
		Name               string  `json:"name"`
		Category           string  `json:"category"`
		Description        *string `json:"description"`
		TypicalIngredients *string `json:"typical_ingredients"`
		PreparationSummary *string `json:"preparation_summary"`
		StorageGuidelines  *string `json:"storage_guidelines"`
		SafetyPrecautions  *string `json:"safety_precautions"`
		ReferenceSource    *string `json:"reference_source"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	row, err := h.q.CreateInputDefinition(r.Context(), db.CreateInputDefinitionParams{
		FarmID:             farmID,
		Name:               req.Name,
		Category:           db.Gr33nnaturalfarmingInputCategoryEnum(req.Category),
		Description:        req.Description,
		TypicalIngredients: req.TypicalIngredients,
		PreparationSummary: req.PreparationSummary,
		StorageGuidelines:  req.StorageGuidelines,
		SafetyPrecautions:  req.SafetyPrecautions,
		ReferenceSource:    req.ReferenceSource,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, row)
}

// PUT /naturalfarming/inputs/{id}
func (h *Handler) UpdateInputDefinition(w http.ResponseWriter, r *http.Request) {
	id, err := resourceIDFromPath(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid input definition id")
		return
	}
	var req struct {
		Name               string  `json:"name"`
		Category           string  `json:"category"`
		Description        *string `json:"description"`
		TypicalIngredients *string `json:"typical_ingredients"`
		PreparationSummary *string `json:"preparation_summary"`
		StorageGuidelines  *string `json:"storage_guidelines"`
		SafetyPrecautions  *string `json:"safety_precautions"`
		ReferenceSource    *string `json:"reference_source"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	row, err := h.q.UpdateInputDefinition(r.Context(), db.UpdateInputDefinitionParams{
		ID:                 id,
		Name:               req.Name,
		Category:           db.Gr33nnaturalfarmingInputCategoryEnum(req.Category),
		Description:        req.Description,
		TypicalIngredients: req.TypicalIngredients,
		PreparationSummary: req.PreparationSummary,
		StorageGuidelines:  req.StorageGuidelines,
		SafetyPrecautions:  req.SafetyPrecautions,
		ReferenceSource:    req.ReferenceSource,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, row)
}

// DELETE /naturalfarming/inputs/{id}
func (h *Handler) DeleteInputDefinition(w http.ResponseWriter, r *http.Request) {
	id, err := resourceIDFromPath(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid input definition id")
		return
	}
	if err := h.q.SoftDeleteInputDefinition(r.Context(), db.SoftDeleteInputDefinitionParams{
		ID:              id,
		UpdatedByUserID: pgtype.UUID{},
	}); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// POST /farms/{id}/naturalfarming/batches
func (h *Handler) CreateInputBatch(w http.ResponseWriter, r *http.Request) {
	farmID, err := farmIDFromPath(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	var params db.CreateInputBatchParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	params.FarmID = farmID
	row, err := h.q.CreateInputBatch(r.Context(), params)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, row)
}

// PUT /naturalfarming/batches/{id}
func (h *Handler) UpdateInputBatch(w http.ResponseWriter, r *http.Request) {
	id, err := resourceIDFromPath(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid batch id")
		return
	}
	var params db.UpdateInputBatchParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	params.ID = id
	row, err := h.q.UpdateInputBatch(r.Context(), params)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, row)
}

// DELETE /naturalfarming/batches/{id}
func (h *Handler) DeleteInputBatch(w http.ResponseWriter, r *http.Request) {
	id, err := resourceIDFromPath(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid batch id")
		return
	}
	if err := h.q.SoftDeleteInputBatch(r.Context(), db.SoftDeleteInputBatchParams{
		ID:              id,
		UpdatedByUserID: pgtype.UUID{},
	}); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
