package recipe

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

type Handler struct{ q *db.Queries }

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{q: db.New(pool)}
}

func numericFromFloat64(v float64) (pgtype.Numeric, error) {
	var n pgtype.Numeric
	err := n.Scan(strconv.FormatFloat(v, 'f', -1, 64))
	return n, err
}

// List — GET /farms/{id}/naturalfarming/recipes
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	rows, err := h.q.ListRecipesByFarm(r.Context(), farmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if rows == nil {
		rows = []db.Gr33nnaturalfarmingApplicationRecipe{}
	}
	httputil.WriteJSON(w, http.StatusOK, rows)
}

// Get — GET /naturalfarming/recipes/{id}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid recipe id")
		return
	}
	row, err := h.q.GetRecipeByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "recipe not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, row)
}

// Create — POST /farms/{id}/naturalfarming/recipes
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, farmID) {
		return
	}
	var body struct {
		Name                  string  `json:"name"`
		InputDefinitionID     *int64  `json:"input_definition_id"`
		Description           *string `json:"description"`
		TargetApplicationType string  `json:"target_application_type"`
		DilutionRatio         *string `json:"dilution_ratio"`
		Instructions          *string `json:"instructions"`
		FrequencyGuidelines   *string `json:"frequency_guidelines"`
		Notes                 *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	name := strings.TrimSpace(body.Name)
	if name == "" || strings.TrimSpace(body.TargetApplicationType) == "" {
		httputil.WriteError(w, http.StatusBadRequest, "name and target_application_type required")
		return
	}
	row, err := h.q.CreateRecipe(r.Context(), db.CreateRecipeParams{
		FarmID:                farmID,
		Name:                  name,
		InputDefinitionID:     body.InputDefinitionID,
		Description:           body.Description,
		TargetApplicationType: db.Gr33nnaturalfarmingApplicationTargetEnum(body.TargetApplicationType),
		DilutionRatio:         body.DilutionRatio,
		Instructions:          body.Instructions,
		FrequencyGuidelines:   body.FrequencyGuidelines,
		Notes:                 body.Notes,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, row)
}

// Update — PUT /naturalfarming/recipes/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid recipe id")
		return
	}
	var body struct {
		Name                  string  `json:"name"`
		InputDefinitionID     *int64  `json:"input_definition_id"`
		Description           *string `json:"description"`
		TargetApplicationType string  `json:"target_application_type"`
		DilutionRatio         *string `json:"dilution_ratio"`
		Instructions          *string `json:"instructions"`
		FrequencyGuidelines   *string `json:"frequency_guidelines"`
		Notes                 *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	name := strings.TrimSpace(body.Name)
	if name == "" || strings.TrimSpace(body.TargetApplicationType) == "" {
		httputil.WriteError(w, http.StatusBadRequest, "name and target_application_type required")
		return
	}
	rec, err := h.q.GetRecipeByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "recipe not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, rec.FarmID) {
		return
	}
	row, err := h.q.UpdateRecipe(r.Context(), db.UpdateRecipeParams{
		ID:                    id,
		Name:                  name,
		InputDefinitionID:     body.InputDefinitionID,
		Description:           body.Description,
		TargetApplicationType: db.Gr33nnaturalfarmingApplicationTargetEnum(body.TargetApplicationType),
		DilutionRatio:         body.DilutionRatio,
		Instructions:          body.Instructions,
		FrequencyGuidelines:   body.FrequencyGuidelines,
		Notes:                 body.Notes,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "recipe not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, row)
}

// Delete — DELETE /naturalfarming/recipes/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid recipe id")
		return
	}
	rec, err := h.q.GetRecipeByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "recipe not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, rec.FarmID) {
		return
	}
	if err := h.q.SoftDeleteRecipe(r.Context(), id); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListComponents — GET /naturalfarming/recipes/{id}/components
func (h *Handler) ListComponents(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid recipe id")
		return
	}
	if _, err := h.q.GetRecipeByID(r.Context(), id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "recipe not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	rows, err := h.q.ListRecipeComponents(r.Context(), id)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if rows == nil {
		rows = []db.ListRecipeComponentsRow{}
	}
	httputil.WriteJSON(w, http.StatusOK, rows)
}

// AddComponent — POST /naturalfarming/recipes/{id}/components
func (h *Handler) AddComponent(w http.ResponseWriter, r *http.Request) {
	recipeID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid recipe id")
		return
	}
	var body struct {
		InputDefinitionID int64   `json:"input_definition_id"`
		PartValue         float64 `json:"part_value"`
		PartUnitID        *int64  `json:"part_unit_id"`
		Notes             *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if body.InputDefinitionID == 0 {
		httputil.WriteError(w, http.StatusBadRequest, "input_definition_id required")
		return
	}
	rec, err := h.q.GetRecipeByID(r.Context(), recipeID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "recipe not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, rec.FarmID) {
		return
	}
	pv, err := numericFromFloat64(body.PartValue)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid part_value")
		return
	}
	if err := h.q.AddRecipeComponent(r.Context(), db.AddRecipeComponentParams{
		ApplicationRecipeID: recipeID,
		InputDefinitionID:   body.InputDefinitionID,
		PartValue:           pv,
		PartUnitID:          body.PartUnitID,
		Notes:               body.Notes,
	}); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// RemoveComponent — DELETE /naturalfarming/recipes/{id}/components/{iid}
func (h *Handler) RemoveComponent(w http.ResponseWriter, r *http.Request) {
	recipeID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid recipe id")
		return
	}
	inputID, err := strconv.ParseInt(r.PathValue("iid"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid input_definition id")
		return
	}
	rec, err := h.q.GetRecipeByID(r.Context(), recipeID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "recipe not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, rec.FarmID) {
		return
	}
	if err := h.q.RemoveRecipeComponent(r.Context(), db.RemoveRecipeComponentParams{
		ApplicationRecipeID: recipeID,
		InputDefinitionID:   inputID,
	}); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
