package cropprofile

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	cropprofilepkg "gr33n-api/internal/cropprofile"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/httputil"
	db "gr33n-api/internal/db"
)

// GetEffective — GET /farms/{id}/crop-profiles/effective?crop_key=&variety=
func (h *Handler) GetEffective(w http.ResponseWriter, r *http.Request) {
	farmID, cropKey, ok := parseFarmCropKeyFromQuery(w, r)
	if !ok {
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, farmID) {
		return
	}
	var variety *string
	if v := strings.TrimSpace(r.URL.Query().Get("variety")); v != "" {
		variety = &v
	}
	ep, err := cropprofilepkg.ResolveEffective(r.Context(), h.q, farmID, cropKey, variety)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "crop profile not found")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, ep)
}

// GetGenetics — GET /farms/{id}/crop-profiles/{crop_key}/genetics/{variety_slug}
func (h *Handler) GetGenetics(w http.ResponseWriter, r *http.Request) {
	farmID, cropKey, slug, ok := parseFarmCropGenetics(w, r)
	if !ok {
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, farmID) {
		return
	}
	link, err := h.q.GetGeneticsProfileLink(r.Context(), db.GetGeneticsProfileLinkParams{
		FarmID: farmID, CropKey: cropKey, VarietySlug: slug,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "genetics profile not found")
		return
	}
	ep, err := cropprofilepkg.ResolveEffective(r.Context(), h.q, farmID, cropKey, &link.VarietyLabel)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, ep)
}

// PutGenetics — PUT /farms/{id}/crop-profiles/{crop_key}/genetics/{variety_slug}
func (h *Handler) PutGenetics(w http.ResponseWriter, r *http.Request) {
	farmID, cropKey, slug, ok := parseFarmCropGenetics(w, r)
	if !ok {
		return
	}
	if !farmauthz.RequireFarmAdmin(w, r, h.q, farmID) {
		return
	}
	var body struct {
		VarietyLabel string                          `json:"variety_label"`
		Source       *string                         `json:"source"`
		Stages       []db.Gr33ncropsCropProfileStage `json:"stages"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	label := strings.TrimSpace(body.VarietyLabel)
	if label == "" {
		label = strings.ReplaceAll(slug, "_", " ")
	}
	source := "genetics override (UI)"
	if body.Source != nil && strings.TrimSpace(*body.Source) != "" {
		source = strings.TrimSpace(*body.Source)
	}
	if err := cropprofilepkg.UpsertGeneticsProfileFromStages(r.Context(), h.q, farmID, cropKey, label, source, body.Stages); err != nil {
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "required") {
			httputil.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	ep, err := cropprofilepkg.ResolveEffective(r.Context(), h.q, farmID, cropKey, &label)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, ep)
}

// DeleteGenetics — DELETE /farms/{id}/crop-profiles/{crop_key}/genetics/{variety_slug}
func (h *Handler) DeleteGenetics(w http.ResponseWriter, r *http.Request) {
	farmID, cropKey, slug, ok := parseFarmCropGenetics(w, r)
	if !ok {
		return
	}
	if !farmauthz.RequireFarmAdmin(w, r, h.q, farmID) {
		return
	}
	link, err := h.q.GetGeneticsProfileLink(r.Context(), db.GetGeneticsProfileLinkParams{
		FarmID: farmID, CropKey: cropKey, VarietySlug: slug,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "genetics profile not found")
		return
	}
	if err := cropprofilepkg.DeleteGeneticsProfile(r.Context(), h.q, farmID, cropKey, link.VarietyLabel); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func parseFarmCropKeyFromQuery(w http.ResponseWriter, r *http.Request) (farmID int64, cropKey string, ok bool) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || farmID <= 0 {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return 0, "", false
	}
	cropKey = strings.ToLower(strings.TrimSpace(r.URL.Query().Get("crop_key")))
	if cropKey == "" {
		httputil.WriteError(w, http.StatusBadRequest, "crop_key query required")
		return 0, "", false
	}
	return farmID, cropKey, true
}

func parseFarmCropGenetics(w http.ResponseWriter, r *http.Request) (farmID int64, cropKey, slug string, ok bool) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || farmID <= 0 {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return 0, "", "", false
	}
	cropKey = strings.ToLower(strings.TrimSpace(r.PathValue("crop_key")))
	slug = strings.ToLower(strings.TrimSpace(r.PathValue("variety_slug")))
	if cropKey == "" || slug == "" {
		httputil.WriteError(w, http.StatusBadRequest, "invalid crop_key or variety_slug")
		return 0, "", "", false
	}
	return farmID, cropKey, slug, true
}
