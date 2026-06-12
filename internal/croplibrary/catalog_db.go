package croplibrary

import (
	"context"
	"fmt"
	"os"
	"strings"

	db "gr33n-api/internal/db"
)

// CatalogQuerier loads platform catalog rows (sqlc Queries subset).
type CatalogQuerier interface {
	ListCropCatalogEntries(ctx context.Context) ([]db.Gr33ncropsCropCatalogEntry, error)
	ListCropCatalogAliases(ctx context.Context) ([]db.Gr33ncropsCropCatalogAlias, error)
}

// CatalogSource returns yaml or db from CROP_CATALOG_SOURCE (default yaml).
func CatalogSource() string {
	s := strings.ToLower(strings.TrimSpace(os.Getenv("CROP_CATALOG_SOURCE")))
	if s == "db" {
		return "db"
	}
	return "yaml"
}

// LoadCatalogForRuntime picks YAML file or DB based on CROP_CATALOG_SOURCE.
func LoadCatalogForRuntime(ctx context.Context, repoRoot, catalogPath string, q CatalogQuerier) (*Catalog, error) {
	if CatalogSource() == "db" {
		if q == nil {
			return nil, fmt.Errorf("CROP_CATALOG_SOURCE=db requires database querier")
		}
		return LoadCatalogFromDB(ctx, q)
	}
	return LoadCatalog(repoRoot, catalogPath)
}

// LoadCatalogFromDB builds a Catalog from platform catalog tables.
func LoadCatalogFromDB(ctx context.Context, q CatalogQuerier) (*Catalog, error) {
	entries, err := q.ListCropCatalogEntries(ctx)
	if err != nil {
		return nil, fmt.Errorf("list crop catalog entries: %w", err)
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("crop catalog empty — run migrate and catalog seed")
	}
	aliases, err := q.ListCropCatalogAliases(ctx)
	if err != nil {
		return nil, fmt.Errorf("list crop catalog aliases: %w", err)
	}

	version := 1
	aliasByCrop := make(map[string][]string)
	globalAliases := make(map[string]string)
	for _, a := range aliases {
		globalAliases[strings.ToLower(a.Alias)] = a.CropKey
		aliasByCrop[a.CropKey] = append(aliasByCrop[a.CropKey], a.Alias)
	}

	cat := &Catalog{
		Aliases: globalAliases,
	}
	for _, e := range entries {
		if int(e.CatalogVersion) > version {
			version = int(e.CatalogVersion)
		}
		cropAliases := filterCropAliases(aliasByCrop[e.CropKey], e.CropKey)
		if e.Supported {
			crop := Crop{
				Key:              e.CropKey,
				DisplayName:      e.DisplayName,
				Category:         derefStr(e.Category),
				Source:           derefStr(e.Source),
				Substrate:        derefStr(e.Substrate),
				WateringStyle:    derefStr(e.WateringStyle),
				RunoffPctTarget:  derefStr(e.RunoffPctTarget),
				MoistureGuidance: derefStr(e.MoistureGuidance),
				CousinOf:         optionalStringPtr(e.CousinOf),
				Aliases:          cropAliases,
			}
			cat.Crops = append(cat.Crops, crop)
			continue
		}
		display := e.DisplayName
		cat.Unsupported = append(cat.Unsupported, UnsupportedCrop{
			Key:         e.CropKey,
			DisplayName: display,
			Aliases:     cropAliases,
			Reason:      derefStr(e.UnsupportedReason),
			CousinOf:    optionalStringPtr(e.CousinOf),
		})
	}
	cat.Version = version
	if err := cat.Validate(); err != nil {
		return nil, err
	}
	return cat, nil
}

func filterCropAliases(all []string, cropKey string) []string {
	seen := make(map[string]struct{})
	var out []string
	cropKey = strings.ToLower(strings.TrimSpace(cropKey))
	for _, a := range all {
		a = strings.ToLower(strings.TrimSpace(a))
		if a == "" || a == cropKey {
			continue
		}
		if _, dup := seen[a]; dup {
			continue
		}
		seen[a] = struct{}{}
		out = append(out, a)
	}
	return out
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return strings.TrimSpace(*s)
}

func optionalStringPtr(s *string) *string {
	if s == nil {
		return nil
	}
	v := strings.TrimSpace(*s)
	if v == "" {
		return nil
	}
	return &v
}
