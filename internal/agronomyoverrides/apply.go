package agronomyoverrides

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"gr33n-api/internal/croplibrary"
	db "gr33n-api/internal/db"
)

// ApplyPack creates or replaces farm crop profile overrides from a WS2 pack.
func ApplyPack(ctx context.Context, q db.Querier, farmID int64, pack *croplibrary.OverridePack) (int, error) {
	if q == nil || pack == nil {
		return 0, fmt.Errorf("querier and pack required")
	}
	if farmID <= 0 {
		return 0, fmt.Errorf("farm_id must be > 0")
	}
	applied := 0
	for _, ov := range pack.Overrides {
		if err := applyOne(ctx, q, farmID, pack.Source, ov); err != nil {
			return applied, err
		}
		applied++
	}
	return applied, nil
}

func applyOne(ctx context.Context, q db.Querier, farmID int64, packSource string, ov croplibrary.CropOverride) error {
	cropKey := strings.ToLower(strings.TrimSpace(ov.CropKey))
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
			return fmt.Errorf("no builtin profile for %q — migrate crop profiles first", cropKey)
		}
		return err
	}
	stages, err := q.ListCropProfileStages(ctx, builtin.ID)
	if err != nil {
		return err
	}
	if len(stages) == 0 {
		return fmt.Errorf("builtin profile %q has no stages", cropKey)
	}

	overrideByStage := make(map[string]croplibrary.StageOverride, len(ov.Stages))
	for _, st := range ov.Stages {
		overrideByStage[strings.TrimSpace(st.Stage)] = st
	}

	if err := q.DeleteFarmCropProfileByKey(ctx, db.DeleteFarmCropProfileByKeyParams{
		FarmID:  &farmID,
		CropKey: cropKey,
	}); err != nil {
		return fmt.Errorf("delete existing override: %w", err)
	}

	displayName := strings.TrimSpace(ov.DisplayName)
	if displayName == "" {
		displayName = builtin.DisplayName
	}
	sourceNote := strings.TrimSpace(packSource)
	if sourceNote == "" {
		sourceNote = "farm agronomy override pack"
	} else {
		sourceNote = "farm override — " + sourceNote
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
		return fmt.Errorf("create farm profile %q: %w", cropKey, err)
	}

	for _, st := range stages {
		merged := mergeStage(st, overrideByStage[string(st.Stage)])
		notes := merged.Notes
		var notesPtr *string
		if strings.TrimSpace(notes) != "" {
			n := strings.TrimSpace(notes)
			notesPtr = &n
		}
		if _, err := q.CreateCropProfileStage(ctx, db.CreateCropProfileStageParams{
			CropProfileID:  created.ID,
			Stage:          st.Stage,
			EcMin:          merged.EcMin,
			EcTarget:       merged.EcTarget,
			EcMax:          merged.EcMax,
			PhMin:          merged.PhMin,
			PhMax:          merged.PhMax,
			VpdMinKpa:      merged.VpdMinKpa,
			VpdMaxKpa:      merged.VpdMaxKpa,
			TempMinC:       merged.TempMinC,
			TempMaxC:       merged.TempMaxC,
			RhMinPct:       merged.RhMinPct,
			RhMaxPct:       merged.RhMaxPct,
			DliTarget:      merged.DliTarget,
			PhotoperiodHrs: merged.PhotoperiodHrs,
			Notes:          notesPtr,
		}); err != nil {
			return fmt.Errorf("create stage %s for %q: %w", st.Stage, cropKey, err)
		}
	}
	return nil
}

type mergedStage struct {
	EcMin, EcTarget, EcMax                   pgtype.Numeric
	PhMin, PhMax                             pgtype.Numeric
	VpdMinKpa, VpdMaxKpa                     pgtype.Numeric
	TempMinC, TempMaxC                       pgtype.Numeric
	RhMinPct, RhMaxPct                       pgtype.Numeric
	DliTarget, PhotoperiodHrs                pgtype.Numeric
	Notes                                    string
}

func mergeStage(base db.Gr33ncropsCropProfileStage, ov croplibrary.StageOverride) mergedStage {
	out := mergedStage{
		EcMin: base.EcMin, EcTarget: base.EcTarget, EcMax: base.EcMax,
		PhMin: base.PhMin, PhMax: base.PhMax,
		VpdMinKpa: base.VpdMinKpa, VpdMaxKpa: base.VpdMaxKpa,
		TempMinC: base.TempMinC, TempMaxC: base.TempMaxC,
		RhMinPct: base.RhMinPct, RhMaxPct: base.RhMaxPct,
		DliTarget: base.DliTarget, PhotoperiodHrs: base.PhotoperiodHrs,
	}
	if base.Notes != nil {
		out.Notes = *base.Notes
	}
	if ov.ECMin != nil {
		out.EcMin = numericFromFloat(*ov.ECMin)
	}
	if ov.ECTarget != nil {
		out.EcTarget = numericFromFloat(*ov.ECTarget)
	}
	if ov.ECMax != nil {
		out.EcMax = numericFromFloat(*ov.ECMax)
	}
	if ov.PHMin != nil {
		out.PhMin = numericFromFloat(*ov.PHMin)
	}
	if ov.PHMax != nil {
		out.PhMax = numericFromFloat(*ov.PHMax)
	}
	if ov.VPDMinKPa != nil {
		out.VpdMinKpa = numericFromFloat(*ov.VPDMinKPa)
	}
	if ov.VPDMaxKPa != nil {
		out.VpdMaxKpa = numericFromFloat(*ov.VPDMaxKPa)
	}
	if ov.TempMinC != nil {
		out.TempMinC = numericFromFloat(*ov.TempMinC)
	}
	if ov.TempMaxC != nil {
		out.TempMaxC = numericFromFloat(*ov.TempMaxC)
	}
	if ov.RHMinPct != nil {
		out.RhMinPct = numericFromFloat(*ov.RHMinPct)
	}
	if ov.RHMaxPct != nil {
		out.RhMaxPct = numericFromFloat(*ov.RHMaxPct)
	}
	if ov.DLITarget != nil {
		out.DliTarget = numericFromFloat(*ov.DLITarget)
	}
	if ov.PhotoperiodHrs != nil {
		out.PhotoperiodHrs = numericFromFloat(*ov.PhotoperiodHrs)
	}
	if strings.TrimSpace(ov.Notes) != "" {
		out.Notes = strings.TrimSpace(ov.Notes)
	}
	return out
}

func numericFromFloat(v float64) pgtype.Numeric {
	var n pgtype.Numeric
	_ = n.Scan(fmt.Sprintf("%g", v))
	return n
}
