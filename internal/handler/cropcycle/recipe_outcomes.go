// Phase 211.05 WS3 — recipe outcome analytics API.

package cropcycle

import (
	"net/http"
	"strconv"
	"strings"

	"gr33n-api/internal/authctx"
	"gr33n-api/internal/cropcycle/recipeoutcomes"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/httputil"
)

// RecipeOutcomes — GET /farms/{id}/crop-analytics/recipe-outcomes?crop_key=&recipe_id=
func (h *Handler) RecipeOutcomes(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, farmID) {
		return
	}

	opt := recipeoutcomes.Options{}
	if ck := strings.TrimSpace(r.URL.Query().Get("crop_key")); ck != "" {
		opt.CropKey = &ck
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("recipe_id")); raw != "" {
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			httputil.WriteError(w, http.StatusBadRequest, "invalid recipe_id")
			return
		}
		opt.ApplicationRecipeID = &id
	}
	includeCosts := true
	if uid, ok := authctx.UserID(r.Context()); ok {
		has, err := farmauthz.HasFarmScope(r.Context(), h.q, uid, farmID, farmauthz.ScopeMoneyCostsRead)
		if err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "failed to resolve farm scopes")
			return
		}
		includeCosts = has
	} else if !authctx.FarmAuthzSkip(r.Context()) {
		includeCosts = false
	}
	opt.IncludeCosts = includeCosts

	result, err := recipeoutcomes.Build(r.Context(), h.q, farmID, opt)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, result)
}
