package zone

import (
	"encoding/json"
	"fmt"
	"math"
)

const (
	defaultLayoutW = 0.20
	defaultLayoutH = 0.18
)

// ZoneLayout is the typed JSON schema stored at meta_data.layout on zones.
// Coordinates are normalized 0–1 relative to the farm canvas.
type ZoneLayout struct {
	X float64  `json:"x"`
	Y float64  `json:"y"`
	W *float64 `json:"w,omitempty"`
	H *float64 `json:"h,omitempty"`
}

func clamp01(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0
	}
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// ValidateZoneLayout parses and validates a layout value extracted from zone
// meta_data. Optional w/h receive server defaults; out-of-range values are
// clamped. Returns an error when the tile would extend past the canvas.
func ValidateZoneLayout(raw json.RawMessage) (*ZoneLayout, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var layout ZoneLayout
	if err := json.Unmarshal(raw, &layout); err != nil {
		return nil, fmt.Errorf("layout: %w", err)
	}

	layout.X = clamp01(layout.X)
	layout.Y = clamp01(layout.Y)

	w := defaultLayoutW
	if layout.W != nil {
		w = clamp01(*layout.W)
	}
	h := defaultLayoutH
	if layout.H != nil {
		h = clamp01(*layout.H)
	}

	if layout.X+w > 1+1e-9 {
		return nil, fmt.Errorf("layout: x + width must not exceed 1 (got x=%.3f w=%.3f)", layout.X, w)
	}
	if layout.Y+h > 1+1e-9 {
		return nil, fmt.Errorf("layout: y + height must not exceed 1 (got y=%.3f h=%.3f)", layout.Y, h)
	}

	out := ZoneLayout{X: layout.X, Y: layout.Y, W: &w, H: &h}
	return &out, nil
}

// ExtractZoneLayout reads the layout key from zone meta_data.
// Returns nil, nil if the key is absent.
func ExtractZoneLayout(meta json.RawMessage) (json.RawMessage, error) {
	if len(meta) == 0 {
		return nil, nil
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(meta, &m); err != nil {
		return nil, err
	}
	return m["layout"], nil
}

// NormalizeZoneLayoutMeta re-validates and re-serializes a layout key in meta_data.
func NormalizeZoneLayoutMeta(meta json.RawMessage) (json.RawMessage, error) {
	raw, err := ExtractZoneLayout(meta)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 || string(raw) == "null" {
		return meta, nil
	}
	validated, err := ValidateZoneLayout(raw)
	if err != nil {
		return nil, err
	}
	normalized, err := json.Marshal(validated)
	if err != nil {
		return nil, err
	}

	var m map[string]json.RawMessage
	if len(meta) == 0 {
		m = map[string]json.RawMessage{}
	} else if err := json.Unmarshal(meta, &m); err != nil {
		return nil, err
	}
	m["layout"] = normalized
	return json.Marshal(m)
}
