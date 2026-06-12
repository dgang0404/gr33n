package cropcycle

import (
	"strings"

	"gr33n-api/internal/croplibrary"
	db "gr33n-api/internal/db"
)

// CycleCropIdentity is catalog identity for a grow run (Phase 104).
type CycleCropIdentity struct {
	CropKey            *string
	CatalogDisplayName *string
	BatchLabel         *string
}

// ResolveCycleCropIdentity derives crop_key and catalog display name from the
// linked plant row. batch_label comes from the cycle when set.
func ResolveCycleCropIdentity(cycle db.Gr33nfertigationCropCycle, plant *db.Gr33ncropsPlant) CycleCropIdentity {
	out := CycleCropIdentity{BatchLabel: ResolveBatchLabel(cycle.BatchLabel, nil)}
	if plant == nil || plant.CropKey == nil {
		return out
	}
	ck := strings.TrimSpace(*plant.CropKey)
	if ck == "" {
		return out
	}
	out.CropKey = &ck
	dn := CatalogDisplayName(ck)
	if dn == "" {
		dn = ck
	}
	out.CatalogDisplayName = &dn
	return out
}

// CatalogDisplayName returns the YAML catalog display_name for a crop_key.
func CatalogDisplayName(cropKey string) string {
	cropKey = strings.TrimSpace(cropKey)
	if cropKey == "" {
		return ""
	}
	cat, err := croplibrary.DefaultCatalog()
	if err != nil {
		return humanizeCropKey(cropKey)
	}
	for _, c := range cat.Crops {
		if c.Key == cropKey {
			if dn := strings.TrimSpace(c.DisplayName); dn != "" {
				return dn
			}
			break
		}
	}
	return humanizeCropKey(cropKey)
}

func humanizeCropKey(key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return ""
	}
	return strings.ReplaceAll(key, "_", " ")
}
