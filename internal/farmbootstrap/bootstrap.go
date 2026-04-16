package farmbootstrap

import "strings"

// Template keys (versioned in DB function gr33ncore.apply_farm_bootstrap_template).
const JadamIndoorPhotoperiodV1 = "jadam_indoor_photoperiod_v1"

// RequestedTemplate reports whether the client sent a non-empty bootstrap_template field.
func RequestedTemplate(p *string) (value string, ok bool) {
	if p == nil {
		return "", false
	}
	v := strings.TrimSpace(*p)
	if v == "" {
		return "", false
	}
	return v, true
}

// IsBlankChoice is true for explicit opt-out values.
func IsBlankChoice(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "none", "blank":
		return true
	default:
		return false
	}
}
