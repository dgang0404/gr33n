package croplibrary

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// OverridePack is a farm-specific EC/VPD/DLI delta file (Phase 83 WS2).
type OverridePack struct {
	Version   int             `yaml:"version"`
	FarmSlug  string          `yaml:"farm_slug,omitempty"`
	Source    string          `yaml:"source,omitempty"`
	Overrides []CropOverride  `yaml:"overrides"`
}

// CropOverride adjusts one builtin crop profile for a farm.
type CropOverride struct {
	CropKey     string          `yaml:"crop_key"`
	DisplayName string          `yaml:"display_name,omitempty"`
	Stages      []StageOverride `yaml:"stages"`
}

// StageOverride sets numeric targets for one growth stage (units: mS/cm, kPa, mol/m²/day).
type StageOverride struct {
	Stage            string   `yaml:"stage"`
	ECMin            *float64 `yaml:"ec_ms_cm_min,omitempty"`
	ECTarget         *float64 `yaml:"ec_ms_cm_target,omitempty"`
	ECMax            *float64 `yaml:"ec_ms_cm_max,omitempty"`
	PHMin            *float64 `yaml:"ph_min,omitempty"`
	PHMax            *float64 `yaml:"ph_max,omitempty"`
	VPDMinKPa        *float64 `yaml:"vpd_kpa_min,omitempty"`
	VPDMaxKPa        *float64 `yaml:"vpd_kpa_max,omitempty"`
	TempMinC         *float64 `yaml:"temp_min_c,omitempty"`
	TempMaxC         *float64 `yaml:"temp_max_c,omitempty"`
	RHMinPct         *float64 `yaml:"rh_min_pct,omitempty"`
	RHMaxPct         *float64 `yaml:"rh_max_pct,omitempty"`
	DLITarget        *float64 `yaml:"dli_target,omitempty"`
	PhotoperiodHrs   *float64 `yaml:"photoperiod_hrs,omitempty"`
	Notes            string   `yaml:"notes,omitempty"`
}

// LoadOverridePack reads a WS2 override YAML file.
func LoadOverridePack(path string) (*OverridePack, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var pack OverridePack
	if err := yaml.Unmarshal(data, &pack); err != nil {
		return nil, fmt.Errorf("parse override pack: %w", err)
	}
	if err := pack.Validate(); err != nil {
		return nil, err
	}
	return &pack, nil
}

// Validate checks override pack shape and stage names.
func (p *OverridePack) Validate() error {
	if p == nil {
		return fmt.Errorf("override pack is nil")
	}
	if p.Version < 1 {
		return fmt.Errorf("override pack version must be >= 1")
	}
	if len(p.Overrides) == 0 {
		return fmt.Errorf("override pack has no overrides")
	}
	seen := make(map[string]struct{})
	for _, o := range p.Overrides {
		key := strings.ToLower(strings.TrimSpace(o.CropKey))
		if key == "" {
			return fmt.Errorf("override missing crop_key")
		}
		if _, dup := seen[key]; dup {
			return fmt.Errorf("duplicate override crop_key %q", key)
		}
		seen[key] = struct{}{}
		if len(o.Stages) == 0 {
			return fmt.Errorf("override %q has no stages", key)
		}
		for _, st := range o.Stages {
			stage := strings.TrimSpace(st.Stage)
			if _, ok := ValidGrowthStages[stage]; !ok {
				return fmt.Errorf("override %q: invalid stage %q", key, stage)
			}
		}
	}
	return nil
}
