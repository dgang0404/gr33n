package zonephotos

import (
	"encoding/json"
	"fmt"
	"sort"
)

const metaKeyPhotoIDs = "photo_attachment_ids"

// MaxPhotosPerZone caps reference photos stored on a single zone.
const MaxPhotosPerZone = 24

// Meta is the JSON shape stored in gr33ncore.zones.meta_data for Guardian walkthrough photos.
type Meta struct {
	PhotoAttachmentIDs []int64 `json:"photo_attachment_ids,omitempty"`
}

// ParseMeta decodes zone meta_data; unknown keys are preserved on Marshal back.
func ParseMeta(raw []byte) (Meta, map[string]json.RawMessage, error) {
	var extra map[string]json.RawMessage
	m := Meta{}
	if len(raw) == 0 {
		return m, extra, nil
	}
	if err := json.Unmarshal(raw, &m); err != nil {
		return Meta{}, nil, fmt.Errorf("decode zone meta_data: %w", err)
	}
	if err := json.Unmarshal(raw, &extra); err != nil {
		return Meta{}, nil, fmt.Errorf("decode zone meta_data envelope: %w", err)
	}
	delete(extra, metaKeyPhotoIDs)
	return m, extra, nil
}

// MarshalMeta merges photo IDs with any other meta_data keys.
func MarshalMeta(m Meta, extra map[string]json.RawMessage) ([]byte, error) {
	out := make(map[string]any, len(extra)+1)
	for k, v := range extra {
		out[k] = v
	}
	if len(m.PhotoAttachmentIDs) > 0 {
		out[metaKeyPhotoIDs] = m.PhotoAttachmentIDs
	}
	return json.Marshal(out)
}

// AppendPhotoID adds an attachment id if not already present.
func AppendPhotoID(m *Meta, id int64) error {
	if id < 1 {
		return fmt.Errorf("invalid attachment id")
	}
	for _, existing := range m.PhotoAttachmentIDs {
		if existing == id {
			return nil
		}
	}
	if len(m.PhotoAttachmentIDs) >= MaxPhotosPerZone {
		return fmt.Errorf("zone already has the maximum of %d photos", MaxPhotosPerZone)
	}
	m.PhotoAttachmentIDs = append(m.PhotoAttachmentIDs, id)
	return nil
}

// RemovePhotoID drops an attachment id; returns whether it was present.
func RemovePhotoID(m *Meta, id int64) bool {
	ids := m.PhotoAttachmentIDs
	for i, existing := range ids {
		if existing == id {
			m.PhotoAttachmentIDs = append(ids[:i], ids[i+1:]...)
			return true
		}
	}
	return false
}

// LatestID returns the most recently appended photo attachment id, or 0.
func LatestID(m Meta) int64 {
	if len(m.PhotoAttachmentIDs) == 0 {
		return 0
	}
	return m.PhotoAttachmentIDs[len(m.PhotoAttachmentIDs)-1]
}

// SortedCopy returns attachment ids in stable ascending order (for API listing).
func SortedCopy(m Meta) []int64 {
	if len(m.PhotoAttachmentIDs) == 0 {
		return nil
	}
	out := append([]int64(nil), m.PhotoAttachmentIDs...)
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}
