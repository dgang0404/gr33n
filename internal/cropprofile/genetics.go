package cropprofile

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"gr33n-api/internal/croplibrary"
	db "gr33n-api/internal/db"
)

// UpsertGeneticsProfileFromStages replaces a per-variety EC override for one crop on a farm.
func UpsertGeneticsProfileFromStages(ctx context.Context, q db.Querier, farmID int64, cropKey, varietyLabel, sourceNote string, stages []db.Gr33ncropsCropProfileStage) error {
	cropKey = strings.ToLower(strings.TrimSpace(cropKey))
	slug := SlugifyVariety(varietyLabel)
	if farmID <= 0 || cropKey == "" || slug == "" {
		return fmt.Errorf("farm_id, crop_key, and variety required")
	}
	if len(stages) == 0 {
		return fmt.Errorf("at least one stage required")
	}
	for _, st := range stages {
		if _, ok := croplibrary.ValidGrowthStages[string(st.Stage)]; !ok {
			return fmt.Errorf("invalid stage %q", st.Stage)
		}
	}

	if link, err := q.GetGeneticsProfileLink(ctx, db.GetGeneticsProfileLinkParams{
		FarmID: farmID, CropKey: cropKey, VarietySlug: slug,
	}); err == nil {
		_ = q.DeleteCropProfileByID(ctx, link.CropProfileID)
		_ = q.DeleteGeneticsProfileLink(ctx, db.DeleteGeneticsProfileLinkParams{
			FarmID: farmID, CropKey: cropKey, VarietySlug: slug,
		})
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	genKey := GeneticsCropKey(cropKey, slug)
	displayName := strings.TrimSpace(varietyLabel)
	if displayName == "" {
		displayName = slug
	}
	sourceNote = strings.TrimSpace(sourceNote)
	if sourceNote == "" {
		sourceNote = "genetics override (variety)"
	}

	builtin, err := q.GetBuiltinCropProfileByKey(ctx, cropKey)
	if err != nil {
		return fmt.Errorf("no builtin profile for %q: %w", cropKey, err)
	}

	farmIDPtr := farmID
	meta := map[string]any{
		"genetics":      true,
		"parent_key":    cropKey,
		"variety_slug":  slug,
		"variety_label": displayName,
	}
	metaBytes, _ := json.Marshal(meta)

	created, err := q.CreateCropProfile(ctx, db.CreateCropProfileParams{
		FarmID:      &farmIDPtr,
		CropKey:     genKey,
		DisplayName: builtin.DisplayName + " (" + displayName + ")",
		Category:    builtin.Category,
		Source:      &sourceNote,
		Version:     1,
		IsBuiltin:   false,
		Meta:        metaBytes,
	})
	if err != nil {
		return fmt.Errorf("create genetics profile: %w", err)
	}

	for _, st := range stages {
		st.CropProfileID = created.ID
		if _, err := q.CreateCropProfileStage(ctx, db.CreateCropProfileStageParams{
			CropProfileID: st.CropProfileID,
			Stage:         st.Stage,
			EcMin:         st.EcMin,
			EcTarget:      st.EcTarget,
			EcMax:         st.EcMax,
			PhMin:         st.PhMin,
			PhMax:         st.PhMax,
			VpdMinKpa:     st.VpdMinKpa,
			VpdMaxKpa:     st.VpdMaxKpa,
			TempMinC:      st.TempMinC,
			TempMaxC:      st.TempMaxC,
			RhMinPct:      st.RhMinPct,
			RhMaxPct:      st.RhMaxPct,
			DliTarget:     st.DliTarget,
			PhotoperiodHrs: st.PhotoperiodHrs,
			Notes:         st.Notes,
		}); err != nil {
			return fmt.Errorf("create genetics stage: %w", err)
		}
	}

	_, err = q.InsertGeneticsProfileLink(ctx, db.InsertGeneticsProfileLinkParams{
		FarmID:         farmID,
		CropKey:        cropKey,
		VarietySlug:    slug,
		VarietyLabel:   displayName,
		CropProfileID:  created.ID,
	})
	return err
}

// DeleteGeneticsProfile removes a variety-specific override.
func DeleteGeneticsProfile(ctx context.Context, q db.Querier, farmID int64, cropKey, varietyLabel string) error {
	slug := SlugifyVariety(varietyLabel)
	if slug == "" {
		return fmt.Errorf("variety required")
	}
	link, err := q.GetGeneticsProfileLink(ctx, db.GetGeneticsProfileLinkParams{
		FarmID: farmID, CropKey: cropKey, VarietySlug: slug,
	})
	if err != nil {
		return err
	}
	if err := q.DeleteGeneticsProfileLink(ctx, db.DeleteGeneticsProfileLinkParams{
		FarmID: farmID, CropKey: cropKey, VarietySlug: slug,
	}); err != nil {
		return err
	}
	return q.DeleteCropProfileByID(ctx, link.CropProfileID)
}
