// Package commonscatalog validates and applies commons catalog pack bodies on import.
package commonscatalog

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	CatalogVersion = "gr33n.commons_catalog.v1"

	KindFertigationRecipePack = "fertigation_recipe_pack"
	KindAgronomySeedPack      = "agronomy_seed_pack"
	KindDocumentationPack     = "documentation_pack"
)

// PackBody is the JSON shape stored in commons_catalog_entries.body.
type PackBody struct {
	CatalogVersion string `json:"catalog_version"`
	Kind           string `json:"kind"`
	ReadmeMD       string `json:"readme_md"`
	// fertigation_recipe_pack
	Programs []RecipeProgram `json:"programs"`
	// agronomy_seed_pack
	PlatformCatalogVersion int            `json:"platform_catalog_version"`
	ExpectedCounts         map[string]int `json:"expected_counts"`
}

// RecipeProgram is one fertigation program in a recipe pack.
type RecipeProgram struct {
	Name                string             `json:"name"`
	Description         *string            `json:"description"`
	TotalVolumeLiters   float64            `json:"total_volume_liters"`
	EcTriggerLow        float64            `json:"ec_trigger_low"`
	PhTriggerLow        float64            `json:"ph_trigger_low"`
	PhTriggerHigh       float64            `json:"ph_trigger_high"`
	IsActive            bool               `json:"is_active"`
	RecommendedCropKeys []string           `json:"recommended_crop_keys"`
	RecommendedStages   []string           `json:"recommended_stages"`
	ProfileECSource     *profileECSource   `json:"profile_ec_source"`
	ECBandMSCM          *ecBand            `json:"ec_band_mscm"`
}

type profileECSource struct {
	CropKey string `json:"crop_key"`
	Stage   string `json:"stage"`
}

type ecBand struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// ApplyResult is returned after import auto-apply.
type ApplyResult struct {
	Kind            string   `json:"kind"`
	Status          string   `json:"status"` // applied | verified | noop | skipped | failed
	Message         string   `json:"message"`
	ProgramsCreated int      `json:"programs_created,omitempty"`
	ProgramsUpdated int      `json:"programs_updated,omitempty"`
	ProgramsSkipped int      `json:"programs_skipped,omitempty"`
	NextSteps       []string `json:"next_steps,omitempty"`
	Details         []string `json:"details,omitempty"`
}

func ParsePackBody(raw json.RawMessage) (PackBody, error) {
	if len(raw) == 0 {
		return PackBody{}, fmt.Errorf("pack body is empty")
	}
	var b PackBody
	if err := json.Unmarshal(raw, &b); err != nil {
		return PackBody{}, fmt.Errorf("invalid pack body JSON: %w", err)
	}
	b.Kind = strings.TrimSpace(b.Kind)
	if b.Kind == "" {
		return b, fmt.Errorf("pack body missing kind")
	}
	if b.CatalogVersion != "" && b.CatalogVersion != CatalogVersion {
		return b, fmt.Errorf("unsupported catalog_version %q (want %s)", b.CatalogVersion, CatalogVersion)
	}
	return b, nil
}

func NormalizeSlug(slug string) (string, error) {
	s := strings.TrimSpace(strings.ToLower(slug))
	if s == "" {
		return "", fmt.Errorf("slug is required")
	}
	if len(s) > 120 {
		return "", fmt.Errorf("slug too long (max 120)")
	}
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			continue
		}
		return "", fmt.Errorf("slug must use lowercase letters, digits, and hyphens only")
	}
	return s, nil
}
