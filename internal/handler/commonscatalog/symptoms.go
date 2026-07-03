package commonscatalog

import (
	"net/http"
	"strings"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/httputil"
)

// GET /commons/agronomy-symptoms
func (h *Handler) ListAgronomySymptoms(w http.ResponseWriter, r *http.Request) {
	cropKey := strings.TrimSpace(r.URL.Query().Get("crop_key"))
	category := strings.TrimSpace(r.URL.Query().Get("category"))
	ctx := r.Context()

	if cropKey == "" && category == "" {
		rows, err := h.q.ListAgronomySymptomEntries(ctx)
		if err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "failed to list symptoms")
			return
		}
		httputil.WriteJSON(w, http.StatusOK, map[string]any{"symptoms": rows})
		return
	}

	var catPtr *string
	if category != "" {
		catPtr = &category
	}
	rows, err := h.q.ListAgronomySymptomsForCrop(ctx, db.ListAgronomySymptomsForCropParams{
		CropKey:  cropKey,
		Category: catPtr,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list symptoms")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]any{"symptoms": rows})
}
