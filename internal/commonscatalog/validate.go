package commonscatalog

import (
	"fmt"
	"strings"

	db "gr33n-api/internal/db"
)

// ValidatePublishBody checks a pack before POST /commons/catalog.
func ValidatePublishBody(b PackBody) error {
	switch b.Kind {
	case KindFertigationRecipePack:
		if len(b.Programs) == 0 {
			return fmt.Errorf("fertigation_recipe_pack requires at least one program")
		}
		for i, p := range b.Programs {
			if strings.TrimSpace(p.Name) == "" {
				return fmt.Errorf("programs[%d].name is required", i)
			}
			if p.TotalVolumeLiters <= 0 {
				return fmt.Errorf("programs[%d].total_volume_liters must be positive", i)
			}
		}
	case KindAgronomySeedPack:
		if b.PlatformCatalogVersion < 1 {
			return fmt.Errorf("agronomy_seed_pack requires platform_catalog_version >= 1")
		}
	case KindDocumentationPack:
		if strings.TrimSpace(b.ReadmeMD) == "" {
			return fmt.Errorf("documentation_pack requires readme_md")
		}
	default:
		return fmt.Errorf("unsupported pack kind %q", b.Kind)
	}
	return nil
}

// ValidateRecipeCropKeys ensures recommended_crop_keys exist in platform catalog.
func ValidateRecipeCropKeys(programs []RecipeProgram, entries []db.Gr33ncropsCropCatalogEntry) error {
	valid := map[string]struct{}{}
	for _, e := range entries {
		k := strings.ToLower(strings.TrimSpace(e.CropKey))
		if k != "" {
			valid[k] = struct{}{}
		}
	}
	if len(valid) == 0 {
		return fmt.Errorf("platform crop catalog is empty")
	}
	for _, p := range programs {
		for _, ck := range p.RecommendedCropKeys {
			norm := strings.ToLower(strings.TrimSpace(ck))
			if norm == "" {
				continue
			}
			if _, ok := valid[norm]; !ok {
				return fmt.Errorf("unknown crop_key %q in program %q", ck, p.Name)
			}
		}
		if p.ProfileECSource != nil {
			norm := strings.ToLower(strings.TrimSpace(p.ProfileECSource.CropKey))
			if norm != "" {
				if _, ok := valid[norm]; !ok {
					return fmt.Errorf("unknown profile_ec_source.crop_key %q in program %q", p.ProfileECSource.CropKey, p.Name)
				}
			}
		}
	}
	return nil
}
