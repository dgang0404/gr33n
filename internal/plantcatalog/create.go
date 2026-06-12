// Package plantcatalog implements catalog-bound plant create/upsert (Phase 85).
package plantcatalog

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	db "gr33n-api/internal/db"
)

// CreateInput is the catalog-bound plant create request.
type CreateInput struct {
	CropKey           string
	CropProfileID     *int64
	DisplayName       string
	VarietyOrCultivar *string
	Meta              []byte
}

// CreateResult is the plant row plus whether it was newly created.
type CreateResult struct {
	Plant   db.Gr33ncropsPlant
	Created bool
}

// UnsupportedCropError indicates catalog row exists but supported=false.
type UnsupportedCropError struct {
	CropKey string
	Reason  string
}

func (e *UnsupportedCropError) Error() string {
	if e.Reason != "" {
		return e.Reason
	}
	return fmt.Sprintf("crop %q is not supported for structured fertigation targets", e.CropKey)
}

// CreateOrGet upserts a farm plant slot by crop_key.
func CreateOrGet(ctx context.Context, q db.Querier, farmID int64, in CreateInput) (CreateResult, error) {
	cropKey := strings.TrimSpace(in.CropKey)
	if cropKey == "" && in.CropProfileID != nil && *in.CropProfileID > 0 {
		prof, err := q.GetCropProfile(ctx, *in.CropProfileID)
		if err != nil {
			return CreateResult{}, fmt.Errorf("crop profile: %w", err)
		}
		cropKey = strings.TrimSpace(prof.CropKey)
	}
	if cropKey == "" {
		return CreateResult{}, errors.New("crop_key required")
	}

	cat, err := q.GetCropCatalogEntry(ctx, cropKey)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return CreateResult{}, fmt.Errorf("unknown crop_key %q", cropKey)
		}
		return CreateResult{}, err
	}
	if !cat.Supported {
		reason := ""
		if cat.UnsupportedReason != nil {
			reason = strings.TrimSpace(*cat.UnsupportedReason)
		}
		return CreateResult{}, &UnsupportedCropError{CropKey: cropKey, Reason: reason}
	}

	if existing, err := q.GetPlantByFarmCropKey(ctx, db.GetPlantByFarmCropKeyParams{
		FarmID:  farmID,
		CropKey: &cropKey,
	}); err == nil {
		if in.VarietyOrCultivar != nil && strings.TrimSpace(*in.VarietyOrCultivar) != "" {
			updated, uerr := q.UpdatePlantVariety(ctx, db.UpdatePlantVarietyParams{
				ID:                existing.ID,
				VarietyOrCultivar: in.VarietyOrCultivar,
			})
			if uerr != nil {
				return CreateResult{}, uerr
			}
			return CreateResult{Plant: updated, Created: false}, nil
		}
		return CreateResult{Plant: existing, Created: false}, nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return CreateResult{}, err
	}

	prof, err := q.GetCropProfileByKey(ctx, db.GetCropProfileByKeyParams{
		CropKey: cropKey,
		FarmID:  &farmID,
	})
	if err != nil {
		return CreateResult{}, fmt.Errorf("effective crop profile for %q: %w", cropKey, err)
	}

	displayName := strings.TrimSpace(cat.DisplayName)
	if displayName == "" {
		displayName = strings.TrimSpace(prof.DisplayName)
	}

	meta := in.Meta
	if len(meta) == 0 {
		meta = []byte("{}")
	}

	profileID := prof.ID
	row, err := q.CreatePlant(ctx, db.CreatePlantParams{
		FarmID:            farmID,
		DisplayName:       displayName,
		VarietyOrCultivar: in.VarietyOrCultivar,
		CropProfileID:     &profileID,
		CropKey:           &cropKey,
		Meta:              meta,
	})
	if err != nil {
		return CreateResult{}, err
	}
	return CreateResult{Plant: row, Created: true}, nil
}
