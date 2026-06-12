package croplibrary

import (
	"encoding/json"
	"strings"
)

// ProfileCatalogMeta holds substrate/watering fields copied from crop catalog (WS-I).
type ProfileCatalogMeta struct {
	Substrate        string `json:"substrate,omitempty"`
	WateringStyle    string `json:"watering_style,omitempty"`
	RunoffPctTarget  string `json:"runoff_pct_target,omitempty"`
	MoistureGuidance string `json:"moisture_guidance,omitempty"`
	CatalogVersion   int    `json:"catalog_version,omitempty"`
}

// ParseProfileCatalogMeta extracts catalog metadata from crop_profiles.meta JSONB.
func ParseProfileCatalogMeta(raw json.RawMessage) ProfileCatalogMeta {
	if len(raw) == 0 {
		return ProfileCatalogMeta{}
	}
	var m ProfileCatalogMeta
	_ = json.Unmarshal(raw, &m)
	m.Substrate = strings.TrimSpace(m.Substrate)
	m.WateringStyle = strings.TrimSpace(m.WateringStyle)
	m.RunoffPctTarget = strings.TrimSpace(m.RunoffPctTarget)
	m.MoistureGuidance = strings.TrimSpace(m.MoistureGuidance)
	return m
}

// HasMoistureInfo reports whether any substrate/watering field is populated.
func (m ProfileCatalogMeta) HasMoistureInfo() bool {
	return m.Substrate != "" || m.WateringStyle != "" || m.RunoffPctTarget != "" || m.MoistureGuidance != ""
}

// FormatMoistureBlock returns Guardian-safe substrate/watering lines (no EC numbers).
func (m ProfileCatalogMeta) FormatMoistureBlock() string {
	if !m.HasMoistureInfo() {
		return ""
	}
	var b strings.Builder
	b.WriteString("Substrate / watering (catalog metadata — not EC targets):")
	if m.Substrate != "" {
		b.WriteString("\nSubstrate: " + m.Substrate)
	}
	if m.WateringStyle != "" {
		b.WriteString("\nWatering style: " + humanizeToken(m.WateringStyle))
	}
	if m.RunoffPctTarget != "" {
		b.WriteString("\nRunoff target: " + m.RunoffPctTarget)
	}
	if m.MoistureGuidance != "" {
		b.WriteString("\nMoisture guidance: " + m.MoistureGuidance)
	}
	return b.String()
}

func humanizeToken(s string) string {
	s = strings.ReplaceAll(s, "_", " ")
	return strings.TrimSpace(s)
}
