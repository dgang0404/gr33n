package plantcatalog

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/httputil"
)

// CreateFromRequest handles POST /farms/{id}/plants with catalog-bound semantics.
func CreateFromRequest(ctx context.Context, q *db.Queries, farmID int64, body json.RawMessage) (plant db.Gr33ncropsPlant, created bool, status int, errMsg string) {
	var req struct {
		CropKey           string          `json:"crop_key"`
		CropProfileID     *int64          `json:"crop_profile_id"`
		DisplayName       string          `json:"display_name"`
		VarietyOrCultivar *string         `json:"variety_or_cultivar"`
		Meta              json.RawMessage `json:"meta"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		return plant, false, http.StatusBadRequest, "invalid body"
	}

	meta := []byte("{}")
	if len(req.Meta) > 0 {
		if !json.Valid(req.Meta) {
			return plant, false, http.StatusBadRequest, "meta must be valid JSON"
		}
		meta = req.Meta
	}

	var variety *string
	if req.VarietyOrCultivar != nil {
		v := strings.TrimSpace(*req.VarietyOrCultivar)
		if v != "" {
			variety = &v
		}
	}

	res, err := CreateOrGet(ctx, q, farmID, CreateInput{
		CropKey:           strings.TrimSpace(req.CropKey),
		CropProfileID:     req.CropProfileID,
		DisplayName:       strings.TrimSpace(req.DisplayName),
		VarietyOrCultivar: variety,
		Meta:              meta,
	})
	if err != nil {
		var unsup *UnsupportedCropError
		if errors.As(err, &unsup) {
			return plant, false, http.StatusBadRequest, unsup.Error()
		}
		if strings.Contains(err.Error(), "unknown crop_key") {
			return plant, false, http.StatusBadRequest, err.Error()
		}
		if strings.Contains(err.Error(), "crop_key required") {
			return plant, false, http.StatusBadRequest, err.Error()
		}
		return plant, false, http.StatusInternalServerError, err.Error()
	}

	status = http.StatusCreated
	if !res.Created {
		status = http.StatusOK
	}
	return res.Plant, res.Created, status, ""
}

// WriteCreateResponse writes the plant JSON with appropriate status.
func WriteCreateResponse(w http.ResponseWriter, plant db.Gr33ncropsPlant, status int) {
	httputil.WriteJSON(w, status, plant)
}

// ResolveCropKeyFromProfile returns crop_key for a profile id (legacy clients).
func ResolveCropKeyFromProfile(ctx context.Context, q *db.Queries, profileID int64) (string, error) {
	prof, err := q.GetCropProfile(ctx, profileID)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(prof.CropKey), nil
}
