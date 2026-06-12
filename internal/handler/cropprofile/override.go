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
	"gr33n-api/internal/auditlog"
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
	entry, err := h.q.GetCropCatalogEntry(r.Context(), cropKey)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusBadRequest, "crop_key not in catalog")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
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
	submitCropOverrideAudit(r, h.q, farmID, cropKey, out.DisplayName, source, entry.CatalogVersion, len(body.Stages),
		"crop_profile_override_upsert", db.Gr33ncoreUserActionTypeEnumUpdateRecord)
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
	entry, entryErr := h.q.GetCropCatalogEntry(r.Context(), cropKey)
	if entryErr != nil && !errors.Is(entryErr, pgx.ErrNoRows) {
		httputil.WriteError(w, http.StatusInternalServerError, entryErr.Error())
		return
	}
	displayName := cropKey
	if profile, err := h.q.GetCropProfileByKey(r.Context(), db.GetCropProfileByKeyParams{
		CropKey: cropKey,
		FarmID:  &farmID,
	}); err == nil {
		displayName = profile.DisplayName
	}
	if err := h.q.DeleteFarmCropProfileByKey(r.Context(), db.DeleteFarmCropProfileByKeyParams{
		FarmID:  &farmID,
		CropKey: cropKey,
	}); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	catalogVersion := int32(0)
	if entryErr == nil {
		catalogVersion = entry.CatalogVersion
	}
	submitCropOverrideAudit(r, h.q, farmID, cropKey, displayName, "", catalogVersion, 0,
		"crop_profile_override_deleted", db.Gr33ncoreUserActionTypeEnumDeleteRecord)
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

func submitCropOverrideAudit(r *http.Request, q db.Querier, farmID int64, cropKey, displayName, source string, catalogVersion int32, stageCount int, kind string, action db.Gr33ncoreUserActionTypeEnum) {
	mod := "gr33ncrops"
	tbl := "crop_profiles"
	rid := cropKey
	desc := strings.TrimSpace(displayName)
	if desc == "" {
		desc = cropKey
	}
	details := map[string]any{
		"kind":            kind,
		"crop_key":        cropKey,
		"catalog_version": catalogVersion,
	}
	if source != "" {
		details["source"] = source
	}
	if stageCount > 0 {
		details["stage_count"] = stageCount
	}
	auditlog.Submit(r.Context(), q, r, auditlog.Event{
		FarmID:         auditlog.FarmIDPtr(farmID),
		Action:         action,
		TargetSchema:   &mod,
		TargetTable:    &tbl,
		TargetRecordID: &rid,
		TargetDesc:     &desc,
		Details:        details,
	})
}
