// Package cropcycle provides shared crop-cycle helpers (Phase 86).
package cropcycle

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	db "gr33n-api/internal/db"
)

// ValidatePlantForActiveGrow ensures plant belongs to farm and has a supported catalog crop_key.
func ValidatePlantForActiveGrow(ctx context.Context, q db.Querier, farmID int64, plantID int64) error {
	if plantID <= 0 {
		return errors.New("plant_id required for active crop cycle — pick a catalog plant in Zone → Plants or Start grow")
	}
	p, err := q.GetPlant(ctx, plantID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("plant_id not found")
		}
		return err
	}
	if p.FarmID != farmID {
		return errors.New("plant_id does not belong to this farm")
	}
	cropKey := ""
	if p.CropKey != nil {
		cropKey = strings.TrimSpace(*p.CropKey)
	}
	if cropKey == "" {
		return errors.New("plant must be catalog-bound (crop_key) before starting an active grow")
	}
	cat, err := q.GetCropCatalogEntry(ctx, cropKey)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("unknown crop_key %q on plant", cropKey)
		}
		return err
	}
	if !cat.Supported {
		reason := strings.TrimSpace(derefStr(cat.UnsupportedReason))
		if reason != "" {
			return fmt.Errorf("plant crop %q is not supported: %s", cropKey, reason)
		}
		return fmt.Errorf("plant crop %q is not supported for structured grows", cropKey)
	}
	return nil
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
