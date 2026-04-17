package naturalfarming

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
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
	if !farmauthz.RequireFarmMember(w, r, h.q, farmID) {
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
	if !farmauthz.RequireFarmMember(w, r, h.q, farmID) {
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
	if !farmauthz.RequireFarmOperate(w, r, h.q, farmID) {
		return
	}
	var req struct {
		Name               string   `json:"name"`
		Category           string   `json:"category"`
		Description        *string  `json:"description"`
		TypicalIngredients *string  `json:"typical_ingredients"`
		PreparationSummary *string  `json:"preparation_summary"`
		StorageGuidelines  *string  `json:"storage_guidelines"`
		SafetyPrecautions  *string  `json:"safety_precautions"`
		ReferenceSource    *string  `json:"reference_source"`
		UnitCost           *float64 `json:"unit_cost"`
		UnitCostCurrency   *string  `json:"unit_cost_currency"`
		UnitCostUnitID     *int64   `json:"unit_cost_unit_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	unitCost, unitCostCurrency, err := parseUnitCost(req.UnitCost, req.UnitCostCurrency)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
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
		UnitCost:           unitCost,
		UnitCostCurrency:   unitCostCurrency,
		UnitCostUnitID:     req.UnitCostUnitID,
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
		Name               string   `json:"name"`
		Category           string   `json:"category"`
		Description        *string  `json:"description"`
		TypicalIngredients *string  `json:"typical_ingredients"`
		PreparationSummary *string  `json:"preparation_summary"`
		StorageGuidelines  *string  `json:"storage_guidelines"`
		SafetyPrecautions  *string  `json:"safety_precautions"`
		ReferenceSource    *string  `json:"reference_source"`
		UnitCost           *float64 `json:"unit_cost"`
		UnitCostCurrency   *string  `json:"unit_cost_currency"`
		UnitCostUnitID     *int64   `json:"unit_cost_unit_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	def, err := h.q.GetInputDefinitionByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "input definition not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, def.FarmID) {
		return
	}
	unitCost, unitCostCurrency, err := parseUnitCost(req.UnitCost, req.UnitCostCurrency)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
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
		UnitCost:           unitCost,
		UnitCostCurrency:   unitCostCurrency,
		UnitCostUnitID:     req.UnitCostUnitID,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, row)
}

// parseUnitCost validates and converts optional unit_cost / currency inputs.
// Returns zeroed pgtype.Numeric (Valid=false) and nil *string when both are
// absent; otherwise a valid Numeric and a normalised uppercase currency.
func parseUnitCost(rawCost *float64, rawCurrency *string) (pgtype.Numeric, *string, error) {
	var num pgtype.Numeric
	if rawCost != nil {
		if err := num.Scan(strconv.FormatFloat(*rawCost, 'f', -1, 64)); err != nil {
			return pgtype.Numeric{}, nil, errors.New("invalid unit_cost")
		}
	}
	var currencyOut *string
	if rawCurrency != nil {
		trimmed := strings.ToUpper(strings.TrimSpace(*rawCurrency))
		if trimmed == "" {
			currencyOut = nil
		} else {
			if len(trimmed) != 3 {
				return pgtype.Numeric{}, nil, errors.New("unit_cost_currency must be ISO 4217 (3 uppercase letters)")
			}
			currencyOut = &trimmed
		}
	}
	return num, currencyOut, nil
}

// DELETE /naturalfarming/inputs/{id}
func (h *Handler) DeleteInputDefinition(w http.ResponseWriter, r *http.Request) {
	id, err := resourceIDFromPath(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid input definition id")
		return
	}
	def, err := h.q.GetInputDefinitionByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "input definition not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, def.FarmID) {
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
	if !farmauthz.RequireFarmOperate(w, r, h.q, farmID) {
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
	b0, err := h.q.GetInputBatchByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "batch not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, b0.FarmID) {
		return
	}
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
	b0, err := h.q.GetInputBatchByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "batch not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, b0.FarmID) {
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
