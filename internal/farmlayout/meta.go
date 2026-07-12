package farmlayout

import (
	"encoding/json"
	"fmt"
)

const metaKeyLayoutBackground = "layout_background_attachment_id"

// Meta is the farm meta_data slice used for the Today farm canvas background.
type Meta struct {
	LayoutBackgroundAttachmentID *int64 `json:"layout_background_attachment_id,omitempty"`
}

// ParseMeta decodes farm meta_data; unknown keys are preserved on Marshal back.
func ParseMeta(raw []byte) (Meta, map[string]json.RawMessage, error) {
	var extra map[string]json.RawMessage
	m := Meta{}
	if len(raw) == 0 {
		return m, extra, nil
	}
	if err := json.Unmarshal(raw, &m); err != nil {
		return Meta{}, nil, fmt.Errorf("decode farm meta_data: %w", err)
	}
	if err := json.Unmarshal(raw, &extra); err != nil {
		return Meta{}, nil, fmt.Errorf("decode farm meta_data envelope: %w", err)
	}
	delete(extra, metaKeyLayoutBackground)
	return m, extra, nil
}

// MarshalMeta merges layout background id with any other meta_data keys.
func MarshalMeta(m Meta, extra map[string]json.RawMessage) ([]byte, error) {
	out := make(map[string]any, len(extra)+1)
	for k, v := range extra {
		out[k] = v
	}
	if m.LayoutBackgroundAttachmentID != nil && *m.LayoutBackgroundAttachmentID > 0 {
		out[metaKeyLayoutBackground] = *m.LayoutBackgroundAttachmentID
	}
	return json.Marshal(out)
}

// SetLayoutBackgroundID records the attachment id on farm meta.
func SetLayoutBackgroundID(m *Meta, id int64) error {
	if id < 1 {
		return fmt.Errorf("invalid attachment id")
	}
	m.LayoutBackgroundAttachmentID = &id
	return nil
}

// ClearLayoutBackgroundID removes the background attachment id.
func ClearLayoutBackgroundID(m *Meta) {
	m.LayoutBackgroundAttachmentID = nil
}
