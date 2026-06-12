// Package cropprofile resolves effective EC targets (Phase 94 genetics + farm overrides).
package cropprofile

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"

	db "gr33n-api/internal/db"
)

var slugNonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

// SlugifyVariety normalizes variety_or_cultivar for genetics profile keys.
func SlugifyVariety(label string) string {
	s := strings.ToLower(strings.TrimSpace(label))
	s = slugNonAlnum.ReplaceAllString(s, "_")
	s = strings.Trim(s, "_")
	return s
}

// GeneticsCropKey is the internal crop_profiles.crop_key for a variety override.
func GeneticsCropKey(cropKey, varietySlug string) string {
	return "genetics:" + strings.ToLower(strings.TrimSpace(cropKey)) + ":" + varietySlug
}

// ProfileSource describes which layer supplied the effective profile.
type ProfileSource string

const (
	SourceGenetics ProfileSource = "genetics"
	SourceFarm     ProfileSource = "farm"
	SourceBuiltin  ProfileSource = "builtin"
)

// EffectiveProfile is the resolved crop profile + stages for a grow context.
type EffectiveProfile struct {
	ProfileWithStages
	Source       ProfileSource `json:"source"`
	VarietySlug  string        `json:"variety_slug,omitempty"`
	VarietyLabel string        `json:"variety_label,omitempty"`
}

// ResolveEffective returns genetics > farm override > builtin for crop_key + optional variety.
func ResolveEffective(ctx context.Context, q db.Querier, farmID int64, cropKey string, variety *string) (EffectiveProfile, error) {
	cropKey = strings.ToLower(strings.TrimSpace(cropKey))
	if cropKey == "" {
		return EffectiveProfile{}, errors.New("crop_key required")
	}
	if variety != nil {
		if slug := SlugifyVariety(*variety); slug != "" {
			if ep, err := geneticsEffective(ctx, q, farmID, cropKey, slug, strings.TrimSpace(*variety)); err == nil {
				return ep, nil
			} else if !errors.Is(err, pgx.ErrNoRows) {
				return EffectiveProfile{}, err
			}
		}
	}
	farmPtr := farmID
	profile, err := q.GetCropProfileByKey(ctx, db.GetCropProfileByKeyParams{
		CropKey: cropKey,
		FarmID:  &farmPtr,
	})
	if err != nil {
		return EffectiveProfile{}, err
	}
	stages, err := q.ListCropProfileStages(ctx, profile.ID)
	if err != nil {
		return EffectiveProfile{}, err
	}
	if stages == nil {
		stages = []db.Gr33ncropsCropProfileStage{}
	}
	src := SourceBuiltin
	if profile.FarmID != nil && !profile.IsBuiltin {
		src = SourceFarm
	}
	return EffectiveProfile{
		ProfileWithStages: ProfileWithStages{Gr33ncropsCropProfile: profile, Stages: stages},
		Source:            src,
	}, nil
}

// ResolveProfileID returns the effective crop_profiles.id for Guardian/UI resolution.
func ResolveProfileID(ctx context.Context, q db.Querier, farmID int64, cropKey string, variety *string) (int64, error) {
	ep, err := ResolveEffective(ctx, q, farmID, cropKey, variety)
	if err != nil {
		return 0, err
	}
	return ep.ID, nil
}

func geneticsEffective(ctx context.Context, q db.Querier, farmID int64, cropKey, slug, label string) (EffectiveProfile, error) {
	link, err := q.GetGeneticsProfileLink(ctx, db.GetGeneticsProfileLinkParams{
		FarmID:      farmID,
		CropKey:     cropKey,
		VarietySlug: slug,
	})
	if err != nil {
		return EffectiveProfile{}, err
	}
	profile, err := q.GetCropProfile(ctx, link.CropProfileID)
	if err != nil {
		return EffectiveProfile{}, err
	}
	stages, err := q.ListCropProfileStages(ctx, profile.ID)
	if err != nil {
		return EffectiveProfile{}, err
	}
	if stages == nil {
		stages = []db.Gr33ncropsCropProfileStage{}
	}
	return EffectiveProfile{
		ProfileWithStages: ProfileWithStages{Gr33ncropsCropProfile: profile, Stages: stages},
		Source:            SourceGenetics,
		VarietySlug:       slug,
		VarietyLabel:        link.VarietyLabel,
	}, nil
}

// ProfileWithStages is a crop profile row plus stage targets.
type ProfileWithStages struct {
	db.Gr33ncropsCropProfile
	Stages []db.Gr33ncropsCropProfileStage `json:"stages"`
}
