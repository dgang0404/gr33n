// Package programmeta parses fertigation program metadata (Phase 102).
package programmeta

import (
	"encoding/json"
	"strings"
)

// Meta is stored in gr33nfertigation.programs.metadata.
type Meta struct {
	RecommendedCropKeys []string         `json:"recommended_crop_keys"`
	RecommendedStages   []string         `json:"recommended_stages"`
	ProfileECSource     *ProfileECSource `json:"profile_ec_source"`
	ECBandMSCM          *ECBand          `json:"ec_band_mscm"`
}

// ProfileECSource points at a crop profile stage used to derive EC band.
type ProfileECSource struct {
	CropKey string `json:"crop_key"`
	Stage   string `json:"stage"`
}

// ECBand is a denormalized EC range in mS/cm.
type ECBand struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// FitResult describes how well a program matches a grow context.
type FitResult struct {
	OK       bool     `json:"ok"`
	Warnings []string `json:"warnings,omitempty"`
}

// Parse reads crop catalog metadata from program metadata JSONB.
func Parse(raw json.RawMessage) Meta {
	if len(raw) == 0 {
		return Meta{}
	}
	var wrapper struct {
		RecommendedCropKeys []string         `json:"recommended_crop_keys"`
		RecommendedStages   []string         `json:"recommended_stages"`
		ProfileECSource     *ProfileECSource `json:"profile_ec_source"`
		ECBandMSCM          *ECBand          `json:"ec_band_mscm"`
	}
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		return Meta{}
	}
	return Meta{
		RecommendedCropKeys: wrapper.RecommendedCropKeys,
		RecommendedStages:   wrapper.RecommendedStages,
		ProfileECSource:     wrapper.ProfileECSource,
		ECBandMSCM:          wrapper.ECBandMSCM,
	}
}

// HasCatalogTags reports whether metadata includes Phase 102 crop/stage tags.
func (m Meta) HasCatalogTags() bool {
	return len(m.RecommendedCropKeys) > 0 || len(m.RecommendedStages) > 0
}

// CheckFit compares program tags to an active grow context.
func (m Meta) CheckFit(cropKey, stage string) FitResult {
	if !m.HasCatalogTags() {
		return FitResult{OK: true}
	}
	var warnings []string
	ck := strings.ToLower(strings.TrimSpace(cropKey))
	st := strings.ToLower(strings.TrimSpace(stage))

	if len(m.RecommendedCropKeys) > 0 && ck != "" {
		if !containsFold(m.RecommendedCropKeys, ck) {
			warnings = append(warnings,
				"program is tagged for crops "+strings.Join(m.RecommendedCropKeys, ", ")+
					" but this grow uses "+cropKey)
		}
	}
	if len(m.RecommendedStages) > 0 && st != "" {
		if !containsFold(m.RecommendedStages, st) {
			warnings = append(warnings,
				"program is tagged for stages "+strings.Join(m.RecommendedStages, ", ")+
					" but this grow is in "+stage)
		}
	}
	return FitResult{OK: len(warnings) == 0, Warnings: warnings}
}

func containsFold(list []string, want string) bool {
	for _, v := range list {
		if strings.EqualFold(strings.TrimSpace(v), want) {
			return true
		}
	}
	return false
}

// MergeMetadata shallow-merges catalog keys into existing metadata JSON.
func MergeMetadata(existing json.RawMessage, patch Meta) (json.RawMessage, error) {
	base := map[string]any{}
	if len(existing) > 0 {
		_ = json.Unmarshal(existing, &base)
	}
	if len(patch.RecommendedCropKeys) > 0 {
		base["recommended_crop_keys"] = patch.RecommendedCropKeys
	}
	if len(patch.RecommendedStages) > 0 {
		base["recommended_stages"] = patch.RecommendedStages
	}
	if patch.ProfileECSource != nil {
		base["profile_ec_source"] = patch.ProfileECSource
	}
	if patch.ECBandMSCM != nil {
		base["ec_band_mscm"] = patch.ECBandMSCM
	}
	return json.Marshal(base)
}
