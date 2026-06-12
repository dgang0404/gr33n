package cropprofile

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"

	"gr33n-api/internal/agronomyoverrides"
	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/httputil"
)

// GetByCropKey — GET /farms/{id}/crop-profiles/{crop_key}
func (h *Handler) GetByCropKey(w http.ResponseWriter, r *http.Request) {
	farmID, cropKey, ok := parseFarmCropKey(w, r)
	if !ok {
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, farmID) {
		return
	}
	out, err := h.effectiveProfile(r, farmID, cropKey)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "crop profile not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, out)
}

// PutByCropKey — PUT /farms/{id}/crop-profiles/{crop_key}
func (h *Handler) PutByCropKey(w http.ResponseWriter, r *http.Request) {
	farmID, cropKey, ok := parseFarmCropKey(w, r)
	if !ok {
		return
	}
	if !farmauthz.RequireFarmAdmin(w, r, h.q, farmID) {
		return
	}
	var body struct {
		DisplayName string                          `json:"display_name"`
		Source      *string                         `json:"source"`
		Stages      []db.Gr33ncropsCropProfileStage `json:"stages"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	source := "farm override (UI)"
	if body.Source != nil && strings.TrimSpace(*body.Source) != "" {
		source = strings.TrimSpace(*body.Source)
	}
	if err := agronomyoverrides.UpsertFarmProfileFromStages(r.Context(), h.q, farmID, cropKey, body.DisplayName, source, body.Stages); err != nil {
		if strings.Contains(err.Error(), "unsupported") || strings.Contains(err.Error(), "invalid stage") {
			httputil.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	out, err := h.effectiveProfile(r, farmID, cropKey)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, out)
}

// DeleteByCropKey — DELETE /farms/{id}/crop-profiles/{crop_key}
func (h *Handler) DeleteByCropKey(w http.ResponseWriter, r *http.Request) {
	farmID, cropKey, ok := parseFarmCropKey(w, r)
	if !ok {
		return
	}
	if !farmauthz.RequireFarmAdmin(w, r, h.q, farmID) {
		return
	}
	if err := h.q.DeleteFarmCropProfileByKey(r.Context(), db.DeleteFarmCropProfileByKeyParams{
		FarmID:  &farmID,
		CropKey: cropKey,
	}); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func parseFarmCropKey(w http.ResponseWriter, r *http.Request) (farmID int64, cropKey string, ok bool) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || farmID <= 0 {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return 0, "", false
	}
	cropKey = strings.ToLower(strings.TrimSpace(r.PathValue("crop_key")))
	if cropKey == "" {
		httputil.WriteError(w, http.StatusBadRequest, "invalid crop_key")
		return 0, "", false
	}
	return farmID, cropKey, true
}

func (h *Handler) effectiveProfile(r *http.Request, farmID int64, cropKey string) (profileWithStages, error) {
	profile, err := h.q.GetCropProfileByKey(r.Context(), db.GetCropProfileByKeyParams{
		CropKey: cropKey,
		FarmID:  &farmID,
	})
	if err != nil {
		return profileWithStages{}, err
	}
	stages, err := h.q.ListCropProfileStages(r.Context(), profile.ID)
	if err != nil {
		return profileWithStages{}, err
	}
	if stages == nil {
		stages = []db.Gr33ncropsCropProfileStage{}
	}
	return profileWithStages{Gr33ncropsCropProfile: profile, Stages: stages}, nil
}
