package commonscatalog

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/fertigation/programmeta"
	"gr33n-api/internal/httputil"
)

// ApplyPack runs kind-specific import side effects after audit row is recorded.
func ApplyPack(ctx context.Context, q db.Querier, farmID int64, raw json.RawMessage) (ApplyResult, error) {
	body, err := ParsePackBody(raw)
	if err != nil {
		return ApplyResult{Status: "failed", Message: err.Error()}, err
	}
	switch body.Kind {
	case KindFertigationRecipePack:
		return applyRecipePack(ctx, q, farmID, body)
	case KindAgronomySeedPack:
		return applyAgronomySeedPack(ctx, q, body)
	case KindDocumentationPack:
		return ApplyResult{
			Kind:    KindDocumentationPack,
			Status:  "noop",
			Message: "Documentation pack recorded — no farm data changes.",
		}, nil
	default:
		return ApplyResult{
			Kind:    body.Kind,
			Status:  "skipped",
			Message: fmt.Sprintf("Unknown pack kind %q — import audit only.", body.Kind),
		}, nil
	}
}

func applyRecipePack(ctx context.Context, q db.Querier, farmID int64, body PackBody) (ApplyResult, error) {
	crops, err := q.ListCropCatalogEntries(ctx)
	if err != nil {
		return ApplyResult{Kind: KindFertigationRecipePack, Status: "failed"}, err
	}
	if err := ValidateRecipeCropKeys(body.Programs, crops); err != nil {
		return ApplyResult{Kind: KindFertigationRecipePack, Status: "failed", Message: err.Error()}, err
	}

	existing, err := q.ListProgramsByFarm(ctx, farmID)
	if err != nil {
		return ApplyResult{Kind: KindFertigationRecipePack, Status: "failed"}, err
	}
	byName := map[string]db.Gr33nfertigationProgram{}
	for _, p := range existing {
		byName[p.Name] = p
	}

	res := ApplyResult{
		Kind:    KindFertigationRecipePack,
		Status:  "applied",
		Message: "Fertigation programs imported. Review in Zones → Water — programs start inactive.",
		NextSteps: []string{
			"Open Zones → Water (or Feeding) and review imported programs.",
			"Enable is_active only after you confirm EC/pH targets match your setup.",
		},
	}

	for _, spec := range body.Programs {
		meta := programMetaFromRecipe(spec)
		if cur, ok := byName[spec.Name]; ok {
			res.ProgramsSkipped++
			if meta.HasCatalogTags() || meta.ProfileECSource != nil || meta.ECBandMSCM != nil {
				merged, mErr := programmeta.MergeMetadata(cur.Metadata, meta)
				if mErr != nil {
					return res, mErr
				}
				if _, uErr := q.UpdateProgramMetadata(ctx, db.UpdateProgramMetadataParams{
					ID:       cur.ID,
					Metadata: merged,
				}); uErr != nil {
					return res, uErr
				}
				res.ProgramsUpdated++
				res.Details = append(res.Details, fmt.Sprintf("Updated metadata for existing program %q (id %d)", spec.Name, cur.ID))
			} else {
				res.Details = append(res.Details, fmt.Sprintf("Skipped existing program %q (id %d)", spec.Name, cur.ID))
			}
			continue
		}

		totalVol, err := httputil.NumericFromFloat64(spec.TotalVolumeLiters)
		if err != nil {
			return res, fmt.Errorf("program %q: invalid total_volume_liters", spec.Name)
		}
		ecLow, err := httputil.NumericFromFloat64(spec.EcTriggerLow)
		if err != nil {
			return res, fmt.Errorf("program %q: invalid ec_trigger_low", spec.Name)
		}
		phLow, err := httputil.NumericFromFloat64(spec.PhTriggerLow)
		if err != nil {
			return res, fmt.Errorf("program %q: invalid ph_trigger_low", spec.Name)
		}
		phHigh, err := httputil.NumericFromFloat64(spec.PhTriggerHigh)
		if err != nil {
			return res, fmt.Errorf("program %q: invalid ph_trigger_high", spec.Name)
		}

		row, err := q.CreateProgram(ctx, db.CreateProgramParams{
			FarmID:            farmID,
			Name:              spec.Name,
			Description:       spec.Description,
			TotalVolumeLiters: totalVol,
			EcTriggerLow:      ecLow,
			PhTriggerLow:      phLow,
			PhTriggerHigh:     phHigh,
			IsActive:          spec.IsActive,
		})
		if err != nil {
			return res, err
		}
		if meta.HasCatalogTags() || meta.ProfileECSource != nil || meta.ECBandMSCM != nil {
			merged, mErr := programmeta.MergeMetadata(row.Metadata, meta)
			if mErr != nil {
				return res, mErr
			}
			row, err = q.UpdateProgramMetadata(ctx, db.UpdateProgramMetadataParams{
				ID:       row.ID,
				Metadata: merged,
			})
			if err != nil {
				return res, err
			}
		}
		byName[spec.Name] = row
		res.ProgramsCreated++
		res.Details = append(res.Details, fmt.Sprintf("Created program %q (id %d)", spec.Name, row.ID))
	}
	return res, nil
}

func programMetaFromRecipe(spec RecipeProgram) programmeta.Meta {
	meta := programmeta.Meta{
		RecommendedCropKeys: spec.RecommendedCropKeys,
		RecommendedStages:   spec.RecommendedStages,
	}
	if spec.ProfileECSource != nil {
		meta.ProfileECSource = &programmeta.ProfileECSource{
			CropKey: spec.ProfileECSource.CropKey,
			Stage:   spec.ProfileECSource.Stage,
		}
	}
	if spec.ECBandMSCM != nil {
		meta.ECBandMSCM = &programmeta.ECBand{
			Min: spec.ECBandMSCM.Min,
			Max: spec.ECBandMSCM.Max,
		}
	}
	return meta
}

func applyAgronomySeedPack(ctx context.Context, q db.Querier, body PackBody) (ApplyResult, error) {
	res := ApplyResult{
		Kind:   KindAgronomySeedPack,
		Status: "verified",
		NextSteps: []string{
			"Settings → Field memories → Re-ingest (or run make guardian-bootstrap-farm FARM_ID=N).",
			"Optional: Settings → Crops & targets for farm-specific EC overrides.",
		},
	}
	maxVer, err := q.GetMaxCropCatalogVersion(ctx)
	if err != nil {
		return ApplyResult{Kind: KindAgronomySeedPack, Status: "failed"}, err
	}
	if int(maxVer) < body.PlatformCatalogVersion {
		msg := fmt.Sprintf("platform catalog version %d is below pack minimum %d — run make migrate",
			maxVer, body.PlatformCatalogVersion)
		return ApplyResult{Kind: KindAgronomySeedPack, Status: "failed", Message: msg},
			fmt.Errorf("%s", msg)
	}
	res.Details = append(res.Details, fmt.Sprintf("platform_catalog_version OK (DB %d, pack >= %d)", maxVer, body.PlatformCatalogVersion))

	checks := []struct {
		key string
		fn  func(context.Context, db.Querier) (int64, error)
	}{
		{"crop_catalog_entries", func(ctx context.Context, q db.Querier) (int64, error) {
			c, err := q.CountCropCatalogEntries(ctx)
			return c, err
		}},
		{"supported_crops", func(ctx context.Context, q db.Querier) (int64, error) {
			return q.CountSupportedCropCatalogEntries(ctx)
		}},
		{"unsupported_crops", func(ctx context.Context, q db.Querier) (int64, error) {
			return q.CountUnsupportedCropCatalogEntries(ctx)
		}},
		{"field_guides_published", func(ctx context.Context, q db.Querier) (int64, error) {
			return q.CountAgronomyFieldGuides(ctx)
		}},
		{"builtin_profiles", func(ctx context.Context, q db.Querier) (int64, error) {
			return q.CountBuiltinCropProfiles(ctx)
		}},
	}
	for _, c := range checks {
		got, err := c.fn(ctx, q)
		if err != nil {
			return ApplyResult{Kind: KindAgronomySeedPack, Status: "failed"}, err
		}
		want := body.ExpectedCounts[c.key]
		if want > 0 && got < int64(want) {
			msg := fmt.Sprintf("%s: got %d want >= %d", c.key, got, want)
			return ApplyResult{Kind: KindAgronomySeedPack, Status: "failed", Message: msg},
				fmt.Errorf("%s", msg)
		}
		if want > 0 {
			res.Details = append(res.Details, fmt.Sprintf("%s OK (%d >= %d)", c.key, got, want))
		}
	}
	res.Message = "Platform agronomy catalog verified. Run Guardian field-memory bootstrap to embed guides for this farm."
	return res, nil
}

// BuildRecipePackBody exports farm fertigation programs into a publishable pack body.
func BuildRecipePackBody(programs []db.Gr33nfertigationProgram, readme string) (json.RawMessage, error) {
	pack := PackBody{
		CatalogVersion: CatalogVersion,
		Kind:           KindFertigationRecipePack,
		ReadmeMD:       readme,
		Programs:       make([]RecipeProgram, 0, len(programs)),
	}
	for _, p := range programs {
		meta := programmeta.Parse(p.Metadata)
		spec := RecipeProgram{
			Name:              p.Name,
			Description:       p.Description,
			TotalVolumeLiters: numericFloat(p.TotalVolumeLiters),
			EcTriggerLow:      numericFloat(p.EcTriggerLow),
			PhTriggerLow:      numericFloat(p.PhTriggerLow),
			PhTriggerHigh:     numericFloat(p.PhTriggerHigh),
			IsActive:          false, // always export inactive for safety
			RecommendedCropKeys: meta.RecommendedCropKeys,
			RecommendedStages:   meta.RecommendedStages,
		}
		if meta.ProfileECSource != nil {
			spec.ProfileECSource = &profileECSource{
				CropKey: meta.ProfileECSource.CropKey,
				Stage:   meta.ProfileECSource.Stage,
			}
		}
		if meta.ECBandMSCM != nil {
			spec.ECBandMSCM = &ecBand{Min: meta.ECBandMSCM.Min, Max: meta.ECBandMSCM.Max}
		}
		pack.Programs = append(pack.Programs, spec)
	}
	return json.Marshal(pack)
}

func numericFloat(n pgtype.Numeric) float64 {
	f, err := n.Float64Value()
	if err != nil || !f.Valid {
		return 0
	}
	return f.Float64
}
