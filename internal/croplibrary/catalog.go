// Package croplibrary loads and validates data/crop_library.yaml (Phase 82 WS4a).
package croplibrary

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const DefaultCatalogPath = "data/crop_library.yaml"

// ValidGrowthStages matches gr33nfertigation.growth_stage_enum.
var ValidGrowthStages = map[string]struct{}{
	"clone": {}, "seedling": {}, "early_veg": {}, "late_veg": {},
	"transition": {}, "early_flower": {}, "mid_flower": {}, "late_flower": {},
	"flush": {}, "harvest": {}, "dry_cure": {},
}

var validWateringStyles = map[string]struct{}{
	"constant_feed": {}, "pulse_dryback": {}, "top_water_drydown": {},
	"mist_epiphyte": {}, "dry_down_succulent": {},
}

// Catalog is the canonical crop library (Phase 82 WS4a).
type Catalog struct {
	Version      int                `yaml:"version"`
	Aliases      map[string]string  `yaml:"aliases"`
	Crops        []Crop             `yaml:"crops"`
	Unsupported  []UnsupportedCrop  `yaml:"unsupported"`
	catalogPath  string
	cropKeys     map[string]struct{}
	unsupportedK map[string]struct{}
}

// Crop is one supported cultivator profile entry.
type Crop struct {
	Key              string      `yaml:"key"`
	DisplayName      string      `yaml:"display_name"`
	Category         string      `yaml:"category"`
	Source           string      `yaml:"source,omitempty"`
	Substrate        string      `yaml:"substrate,omitempty"`
	WateringStyle    string      `yaml:"watering_style,omitempty"`
	RunoffPctTarget  string      `yaml:"runoff_pct_target,omitempty"`
	MoistureGuidance string      `yaml:"moisture_guidance,omitempty"`
	CousinOf         *string     `yaml:"cousin_of"`
	Aliases          []string    `yaml:"aliases,omitempty"`
	ImageURL         string      `yaml:"image_url,omitempty"`
	Stages           []StageRow  `yaml:"stages,omitempty"`
}

// StageRow holds per-stage targets; EC fields are mS/cm.
type StageRow struct {
	Stage          string   `yaml:"stage"`
	ECMin          *float64 `yaml:"ec_min,omitempty"`
	ECTarget       *float64 `yaml:"ec_target,omitempty"`
	ECMax          *float64 `yaml:"ec_max,omitempty"`
	PHMin          *float64 `yaml:"ph_min,omitempty"`
	PHMax          *float64 `yaml:"ph_max,omitempty"`
	VPDMinKPa      *float64 `yaml:"vpd_min_kpa,omitempty"`
	VPDMaxKPa      *float64 `yaml:"vpd_max_kpa,omitempty"`
	TempMinC       *float64 `yaml:"temp_min_c,omitempty"`
	TempMaxC       *float64 `yaml:"temp_max_c,omitempty"`
	RHMinPct       *float64 `yaml:"rh_min_pct,omitempty"`
	RHMaxPct       *float64 `yaml:"rh_max_pct,omitempty"`
	DLITarget      *float64 `yaml:"dli_target,omitempty"`
	PhotoperiodHrs *float64 `yaml:"photoperiod_hrs,omitempty"`
	Notes          string   `yaml:"notes,omitempty"`
}

// UnsupportedCrop has no structured targets — honest Guardian handling (WS4e).
type UnsupportedCrop struct {
	Key         string   `yaml:"key"`
	DisplayName string   `yaml:"display_name,omitempty"`
	Aliases     []string `yaml:"aliases,omitempty"`
	Reason      string   `yaml:"reason"`
	CousinOf    *string  `yaml:"cousin_of"`
}

// LoadCatalog reads and validates the YAML catalog at repoRoot/data/crop_library.yaml
// (or catalogPath when relative to repoRoot).
func LoadCatalog(repoRoot, catalogPath string) (*Catalog, error) {
	if strings.TrimSpace(catalogPath) == "" {
		catalogPath = DefaultCatalogPath
	}
	abs := catalogPath
	if !filepath.IsAbs(catalogPath) {
		abs = filepath.Join(repoRoot, catalogPath)
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		return nil, fmt.Errorf("read crop library: %w", err)
	}
	var cat Catalog
	if err := yaml.Unmarshal(data, &cat); err != nil {
		return nil, fmt.Errorf("parse crop library: %w", err)
	}
	cat.catalogPath = abs
	if err := cat.Validate(); err != nil {
		return nil, err
	}
	return &cat, nil
}

// Validate enforces schema rules: growth stages, mS/cm EC ranges, alias integrity.
func (c *Catalog) Validate() error {
	if c.Version < 1 {
		return fmt.Errorf("crop library: version must be >= 1")
	}
	c.buildIndexes()
	var errs []string

	for alias, target := range c.Aliases {
		if strings.TrimSpace(alias) == "" {
			errs = append(errs, "aliases: empty alias key")
			continue
		}
		if !c.knownKey(target) {
			errs = append(errs, fmt.Sprintf("aliases[%q]: unknown target %q", alias, target))
		}
	}

	seenCrop := make(map[string]struct{})
	for _, crop := range c.Crops {
		if crop.Key == "" {
			errs = append(errs, "crops: missing key")
			continue
		}
		if _, dup := seenCrop[crop.Key]; dup {
			errs = append(errs, fmt.Sprintf("crops: duplicate key %q", crop.Key))
		}
		seenCrop[crop.Key] = struct{}{}
		if crop.DisplayName == "" {
			errs = append(errs, fmt.Sprintf("crop %q: display_name required", crop.Key))
		}
		if crop.WateringStyle != "" {
			if _, ok := validWateringStyles[crop.WateringStyle]; !ok {
				errs = append(errs, fmt.Sprintf("crop %q: unknown watering_style %q", crop.Key, crop.WateringStyle))
			}
		}
		if crop.CousinOf != nil && strings.TrimSpace(*crop.CousinOf) != "" {
			if !c.isCropKey(*crop.CousinOf) {
				errs = append(errs, fmt.Sprintf("crop %q: cousin_of %q is not a supported crop key", crop.Key, *crop.CousinOf))
			}
		}
		for _, a := range crop.Aliases {
			if strings.TrimSpace(a) == "" {
				errs = append(errs, fmt.Sprintf("crop %q: empty alias", crop.Key))
			}
		}
		seenStage := make(map[string]struct{})
		for _, st := range crop.Stages {
			errs = append(errs, validateStageRow(crop.Key, st, seenStage)...)
		}
	}

	seenUnsup := make(map[string]struct{})
	for _, u := range c.Unsupported {
		if u.Key == "" {
			errs = append(errs, "unsupported: missing key")
			continue
		}
		if _, dup := seenUnsup[u.Key]; dup {
			errs = append(errs, fmt.Sprintf("unsupported: duplicate key %q", u.Key))
		}
		seenUnsup[u.Key] = struct{}{}
		if strings.TrimSpace(u.Reason) == "" {
			errs = append(errs, fmt.Sprintf("unsupported %q: reason required", u.Key))
		}
		if u.CousinOf != nil && strings.TrimSpace(*u.CousinOf) != "" {
			if !c.isCropKey(*u.CousinOf) {
				errs = append(errs, fmt.Sprintf("unsupported %q: cousin_of %q is not a supported crop key", u.Key, *u.CousinOf))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("crop library validation failed (%s):\n  - %s",
			c.catalogPath, strings.Join(errs, "\n  - "))
	}
	return nil
}

func (c *Catalog) buildIndexes() {
	c.cropKeys = make(map[string]struct{}, len(c.Crops))
	for _, crop := range c.Crops {
		c.cropKeys[crop.Key] = struct{}{}
	}
	c.unsupportedK = make(map[string]struct{}, len(c.Unsupported))
	for _, u := range c.Unsupported {
		c.unsupportedK[u.Key] = struct{}{}
	}
}

func (c *Catalog) isCropKey(key string) bool {
	_, ok := c.cropKeys[key]
	return ok
}

func (c *Catalog) knownKey(key string) bool {
	if c.isCropKey(key) {
		return true
	}
	_, ok := c.unsupportedK[key]
	return ok
}

func validateStageRow(cropKey string, st StageRow, seen map[string]struct{}) []string {
	var errs []string
	prefix := fmt.Sprintf("crop %q stage %q", cropKey, st.Stage)
	if st.Stage == "" {
		return []string{fmt.Sprintf("crop %q: stage row missing stage key", cropKey)}
	}
	if _, ok := ValidGrowthStages[st.Stage]; !ok {
		errs = append(errs, fmt.Sprintf("%s: invalid stage (must match growth_stage_enum)", prefix))
	}
	if _, dup := seen[st.Stage]; dup {
		errs = append(errs, fmt.Sprintf("%s: duplicate stage", prefix))
	}
	seen[st.Stage] = struct{}{}

	for label, v := range map[string]*float64{
		"ec_min": st.ECMin, "ec_target": st.ECTarget, "ec_max": st.ECMax,
	} {
		if v == nil {
			continue
		}
		if err := validateECmScm(prefix, label, *v); err != "" {
			errs = append(errs, err)
		}
	}
	if st.ECMin != nil && st.ECTarget != nil && *st.ECMin > *st.ECTarget {
		errs = append(errs, fmt.Sprintf("%s: ec_min > ec_target", prefix))
	}
	if st.ECTarget != nil && st.ECMax != nil && *st.ECTarget > *st.ECMax {
		errs = append(errs, fmt.Sprintf("%s: ec_target > ec_max", prefix))
	}
	if st.ECMin != nil && st.ECMax != nil && *st.ECMin > *st.ECMax {
		errs = append(errs, fmt.Sprintf("%s: ec_min > ec_max", prefix))
	}
	return errs
}

// validateECmScm rejects values outside hydroponic mS/cm range (guards against % EC mistakes).
func validateECmScm(prefix, field string, v float64) string {
	if v < 0 || v > 5.0 {
		return fmt.Sprintf("%s: %s=%g out of mS/cm range (0–5)", prefix, field, v)
	}
	return ""
}

// CropsWithStages returns crops that have at least one stage row (seed SQL candidates).
func (c *Catalog) CropsWithStages() []Crop {
	var out []Crop
	for _, crop := range c.Crops {
		if len(crop.Stages) > 0 {
			out = append(out, crop)
		}
	}
	return out
}
