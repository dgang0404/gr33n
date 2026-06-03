// Package fieldguides serves static field procedure catalogs (Phase 37 WS3/WS6).
package fieldguides

import (
	"net/http"
	"strings"

	"gr33n-api/internal/farmguardian/procedures"
	"gr33n-api/internal/httputil"
)

// Handler exposes field guide procedure HTTP routes.
type Handler struct {
	RepoRoot string
}

// NewHandler returns a handler that reads procedures from repoRoot.
func NewHandler(repoRoot string) *Handler {
	if strings.TrimSpace(repoRoot) == "" {
		repoRoot = procedures.RepoRoot()
	}
	return &Handler{RepoRoot: repoRoot}
}

// ListProcedures handles GET /v1/field-guides/procedures.
func (h *Handler) ListProcedures(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	all, err := procedures.List(h.RepoRoot)
	if err != nil {
		httputil.WriteError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	type summary struct {
		ID        string `json:"id"`
		Title     string `json:"title"`
		Domain    string `json:"domain"`
		OfflineOK bool   `json:"offline_ok"`
		StepCount int    `json:"step_count"`
	}
	out := make([]summary, 0, len(all))
	for _, p := range all {
		out = append(out, summary{
			ID: p.ID, Title: p.Title, Domain: p.Domain, OfflineOK: p.OfflineOK, StepCount: len(p.Steps),
		})
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]any{"procedures": out})
}

// GetProcedure handles GET /v1/field-guides/procedures/{id}.
func (h *Handler) GetProcedure(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		httputil.WriteError(w, http.StatusBadRequest, "procedure id required")
		return
	}
	p, err := procedures.Get(h.RepoRoot, id)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, p)
}

// PrintProcedure handles GET /v1/field-guides/procedures/{id}/print — static markdown (WS6).
func (h *Handler) PrintProcedure(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		httputil.WriteError(w, http.StatusBadRequest, "procedure id required")
		return
	}
	p, err := procedures.Get(h.RepoRoot, id)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, err.Error())
		return
	}
	md := procedures.PrintMarkdown(p)
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(md))
}
