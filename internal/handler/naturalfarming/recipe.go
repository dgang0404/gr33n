package naturalfarming

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/httputil"
	"gr33n-api/internal/reciperevision"
)

// ListRecipes — GET /farms/{id}/naturalfarming/recipes
func (h *Handler) ListRecipes(w http.ResponseWriter, r *http.Request) {
	farmID, err := httputil.PathValueInt64(r, "id")
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, farmID) {
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

// GetRecipe — GET /naturalfarming/recipes/{id}
func (h *Handler) GetRecipe(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.PathValueInt64(r, "id")
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
	if !farmauthz.RequireFarmMember(w, r, h.q, row.FarmID) {
		return
	}
	httputil.WriteJSON(w, http.StatusOK, row)
}

// CreateRecipe — POST /farms/{id}/naturalfarming/recipes
func (h *Handler) CreateRecipe(w http.ResponseWriter, r *http.Request) {
	farmID, err := httputil.PathValueInt64(r, "id")
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, farmID) {
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
	if _, err := reciperevision.Record(r.Context(), h.q, row.ID, "recipe created"); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, row)
}

// UpdateRecipe — PUT /naturalfarming/recipes/{id}
func (h *Handler) UpdateRecipe(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.PathValueInt64(r, "id")
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
	if !farmauthz.RequireFarmOperate(w, r, h.q, rec.FarmID) {
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
	if _, err := reciperevision.Record(r.Context(), h.q, row.ID, "recipe updated"); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, row)
}

// ListRecipeRevisions — GET /naturalfarming/recipes/{id}/revisions
func (h *Handler) ListRecipeRevisions(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.PathValueInt64(r, "id")
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
	rows, err := h.q.ListRecipeRevisions(r.Context(), id)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if rows == nil {
		rows = []db.Gr33nnaturalfarmingApplicationRecipeRevision{}
	}
	httputil.WriteJSON(w, http.StatusOK, rows)
}

// DeleteRecipe — DELETE /naturalfarming/recipes/{id}
func (h *Handler) DeleteRecipe(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.PathValueInt64(r, "id")
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
	if !farmauthz.RequireFarmOperate(w, r, h.q, rec.FarmID) {
		return
	}
	if err := h.q.SoftDeleteRecipe(r.Context(), id); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListRecipeComponents — GET /naturalfarming/recipes/{id}/components
func (h *Handler) ListRecipeComponents(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.PathValueInt64(r, "id")
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

// AddRecipeComponent — POST /naturalfarming/recipes/{id}/components
func (h *Handler) AddRecipeComponent(w http.ResponseWriter, r *http.Request) {
	recipeID, err := httputil.PathValueInt64(r, "id")
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
	if !farmauthz.RequireFarmOperate(w, r, h.q, rec.FarmID) {
		return
	}
	pv, err := httputil.NumericFromFloat64(body.PartValue)
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
	if _, err := reciperevision.Record(r.Context(), h.q, recipeID, "component added or updated"); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// RemoveRecipeComponent — DELETE /naturalfarming/recipes/{id}/components/{iid}
func (h *Handler) RemoveRecipeComponent(w http.ResponseWriter, r *http.Request) {
	recipeID, err := httputil.PathValueInt64(r, "id")
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid recipe id")
		return
	}
	inputID, err := httputil.PathValueInt64(r, "iid")
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
	if !farmauthz.RequireFarmOperate(w, r, h.q, rec.FarmID) {
		return
	}
	if err := h.q.RemoveRecipeComponent(r.Context(), db.RemoveRecipeComponentParams{
		ApplicationRecipeID: recipeID,
		InputDefinitionID:   inputID,
	}); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if _, err := reciperevision.Record(r.Context(), h.q, recipeID, "component removed"); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// RestoreRecipeRevision — POST /naturalfarming/recipes/{id}/revisions/{rid}/restore
func (h *Handler) RestoreRecipeRevision(w http.ResponseWriter, r *http.Request) {
	recipeID, err := httputil.PathValueInt64(r, "id")
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid recipe id")
		return
	}
	revisionID, err := httputil.PathValueInt64(r, "rid")
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid revision id")
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
	if !farmauthz.RequireFarmOperate(w, r, h.q, rec.FarmID) {
		return
	}

	tx, err := h.pool.Begin(r.Context())
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to begin transaction")
		return
	}
	defer tx.Rollback(r.Context())

	created, err := reciperevision.RestoreFromRevision(r.Context(), h.q.WithTx(tx), recipeID, revisionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "revision not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := tx.Commit(r.Context()); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to commit restore")
		return
	}

	updated, err := h.q.GetRecipeByID(r.Context(), recipeID)
	if err != nil {
		httputil.WriteJSON(w, http.StatusOK, map[string]any{"revision": created})
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]any{"recipe": updated, "revision": created})
}