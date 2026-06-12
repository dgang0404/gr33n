package agronomyoverrides

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"gr33n-api/internal/croplibrary"
	db "gr33n-api/internal/db"
)

// UpsertFarmProfileFromStages replaces a farm crop override (same crop_key as builtin).
func UpsertFarmProfileFromStages(ctx context.Context, q db.Querier, farmID int64, cropKey, displayName, sourceNote string, stages []db.Gr33ncropsCropProfileStage) error {
	cropKey = strings.ToLower(strings.TrimSpace(cropKey))
	if farmID <= 0 || cropKey == "" {
		return fmt.Errorf("farm_id and crop_key required")
	}
	if len(stages) == 0 {
		return fmt.Errorf("at least one stage required")
	}

	entry, err := q.GetCropCatalogEntry(ctx, cropKey)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("crop_key %q not in catalog", cropKey)
		}
		return err
	}
	if !entry.Supported {
		return fmt.Errorf("cannot override unsupported crop %q", cropKey)
	}

	builtin, err := q.GetBuiltinCropProfileByKey(ctx, cropKey)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("no builtin profile for %q", cropKey)
		}
		return err
	}

	for _, st := range stages {
		if _, ok := croplibrary.ValidGrowthStages[string(st.Stage)]; !ok {
			return fmt.Errorf("invalid stage %q", st.Stage)
		}
	}

	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		displayName = builtin.DisplayName
	}
	sourceNote = strings.TrimSpace(sourceNote)
	if sourceNote == "" {
		sourceNote = "farm override (UI)"
	}

	if err := q.DeleteFarmCropProfileByKey(ctx, db.DeleteFarmCropProfileByKeyParams{
		FarmID:  &farmID,
		CropKey: cropKey,
	}); err != nil {
		return fmt.Errorf("delete existing override: %w", err)
	}

	farmIDPtr := farmID
	created, err := q.CreateCropProfile(ctx, db.CreateCropProfileParams{
		FarmID:      &farmIDPtr,
		CropKey:     cropKey,
		DisplayName: displayName,
		Category:    builtin.Category,
		Source:      &sourceNote,
		Version:     1,
		IsBuiltin:   false,
		Meta:        builtin.Meta,
	})
	if err != nil {
		return fmt.Errorf("create farm profile: %w", err)
	}

	for _, st := range stages {
		notes := st.Notes
		if _, err := q.CreateCropProfileStage(ctx, db.CreateCropProfileStageParams{
			CropProfileID:  created.ID,
			Stage:          st.Stage,
			EcMin:          st.EcMin,
			EcTarget:       st.EcTarget,
			EcMax:          st.EcMax,
			PhMin:          st.PhMin,
			PhMax:          st.PhMax,
			VpdMinKpa:      st.VpdMinKpa,
			VpdMaxKpa:      st.VpdMaxKpa,
			TempMinC:       st.TempMinC,
			TempMaxC:       st.TempMaxC,
			RhMinPct:       st.RhMinPct,
			RhMaxPct:       st.RhMaxPct,
			DliTarget:      st.DliTarget,
			PhotoperiodHrs: st.PhotoperiodHrs,
			Notes:          notes,
		}); err != nil {
			return fmt.Errorf("create stage %s: %w", st.Stage, err)
		}
	}
	return nil
}
