package fieldguides

import (
	"net/http"
	"strings"

	"gr33n-api/internal/farmguardian/procedures"
	"gr33n-api/internal/httputil"
	"gr33n-api/internal/naturalfarmingcatalog"
)

// GetProcessCatalog handles GET /v1/field-guides/process-catalog (Phase 208 WS5).
func (h *Handler) GetProcessCatalog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	cat, err := naturalfarmingcatalog.LoadMaterialCatalog(h.repoRoot())
	if err != nil {
		httputil.WriteError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, cat)
}

// GetProcessMaterial handles GET /v1/field-guides/process-catalog/materials/{id}.
func (h *Handler) GetProcessMaterial(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		httputil.WriteError(w, http.StatusBadRequest, "material id required")
		return
	}
	cat, err := naturalfarmingcatalog.LoadMaterialCatalog(h.repoRoot())
	if err != nil {
		httputil.WriteError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	mat, ok := naturalfarmingcatalog.MaterialByID(cat, id)
	if !ok {
		httputil.WriteError(w, http.StatusNotFound, "material not found")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, mat)
}

// GetRecipeCanon handles GET /v1/field-guides/recipe-canon (Phase 208 WS5).
func (h *Handler) GetRecipeCanon(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	cat, err := naturalfarmingcatalog.LoadRecipeCanon(h.repoRoot())
	if err != nil {
		httputil.WriteError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, cat)
}

func (h *Handler) repoRoot() string {
	if root := strings.TrimSpace(h.RepoRoot); root != "" {
		return root
	}
	return procedures.RepoRoot()
}
