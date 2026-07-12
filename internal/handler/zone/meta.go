package zone

import (
	"encoding/json"
	"fmt"
)

// MergeZoneMetaData overlays incoming meta_data keys onto the existing zone
// meta_data. Keys absent from incoming are preserved. An empty or omitted
// incoming payload leaves existing meta_data unchanged.
func MergeZoneMetaData(existing, incoming json.RawMessage) (json.RawMessage, error) {
	base := map[string]json.RawMessage{}
	if len(existing) > 0 {
		if err := json.Unmarshal(existing, &base); err != nil {
			return nil, fmt.Errorf("decode existing meta_data: %w", err)
		}
	}

	if len(incoming) > 0 && json.Valid(incoming) {
		var inc map[string]json.RawMessage
		if err := json.Unmarshal(incoming, &inc); err != nil {
			return nil, fmt.Errorf("decode incoming meta_data: %w", err)
		}
		for k, v := range inc {
			base[k] = v
		}
	}

	if len(base) == 0 {
		return json.RawMessage("{}"), nil
	}

	merged, err := json.Marshal(base)
	if err != nil {
		return nil, err
	}
	return NormalizeZoneLayoutMeta(merged)
}
