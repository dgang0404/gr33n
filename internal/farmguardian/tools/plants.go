package tools

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/plantcatalog"
)

func execCreatePlant(ctx context.Context, deps ExecutorDeps, args map[string]any) (any, error) {
	if deps.FarmID <= 0 {
		return nil, errors.New("farm_id required in proposal scope")
	}
	if displayName, err := optionalStringFromArgs(args, "display_name"); err != nil {
		return nil, err
	} else if displayName != nil && strings.TrimSpace(*displayName) != "" {
		return nil, errors.New("display_name is server-set from catalog; use crop_key")
	}

	cropKey, err := optionalStringFromArgs(args, "crop_key")
	if err != nil {
		return nil, err
	}
	profileID, err := optionalInt64FromArgs(args, "crop_profile_id")
	if err != nil {
		return nil, err
	}
	if (cropKey == nil || strings.TrimSpace(*cropKey) == "") && (profileID == nil || *profileID <= 0) {
		return nil, errors.New("crop_key required")
	}

	variety, err := optionalStringFromArgs(args, "variety_or_cultivar")
	if err != nil {
		return nil, err
	}
	meta, err := optionalMetaJSONFromArgs(args, "meta")
	if err != nil {
		return nil, err
	}

	in := plantcatalog.CreateInput{
		VarietyOrCultivar: variety,
		Meta:              meta,
		CropProfileID:     profileID,
	}
	if cropKey != nil {
		in.CropKey = strings.TrimSpace(*cropKey)
	}

	res, err := plantcatalog.CreateOrGet(ctx, deps.Q, deps.FarmID, in)
	if err != nil {
		var unsup *plantcatalog.UnsupportedCropError
		if errors.As(err, &unsup) {
			return nil, errors.New(unsup.Error())
		}
		if strings.Contains(err.Error(), "unknown crop_key") || strings.Contains(err.Error(), "crop_key required") {
			return nil, err
		}
		return nil, err
	}
	row := res.Plant
	out := map[string]any{
		"plant_id":     row.ID,
		"display_name": row.DisplayName,
		"created":      res.Created,
	}
	if row.CropKey != nil {
		out["crop_key"] = *row.CropKey
	}
	if row.CropProfileID != nil {
		out["crop_profile_id"] = *row.CropProfileID
	}
	if row.VarietyOrCultivar != nil {
		out["variety_or_cultivar"] = *row.VarietyOrCultivar
	}
	return out, nil
}

func plantExistsForCropKey(ctx context.Context, q db.Querier, farmID int64, cropKey string) (bool, error) {
	cropKey = strings.TrimSpace(cropKey)
	if cropKey == "" {
		return false, nil
	}
	_, err := q.GetPlantByFarmCropKey(ctx, db.GetPlantByFarmCropKeyParams{
		FarmID:  farmID,
		CropKey: &cropKey,
	})
	if err == nil {
		return true, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return false, err
}

func catalogLabelForCropKey(cropKey string) string {
	cropKey = strings.TrimSpace(cropKey)
	if cropKey == "" {
		return "grow setup"
	}
	parts := strings.Fields(strings.ReplaceAll(cropKey, "_", " "))
	for i, p := range parts {
		if len(p) == 0 {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
	}
	return strings.Join(parts, " ")
}
