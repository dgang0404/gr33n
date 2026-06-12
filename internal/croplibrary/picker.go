package croplibrary

import (
	"sort"
	"strings"
)

// PickerResponse powers the grouped crop library UI (Phase 82 WS4f).
type PickerResponse struct {
	Version int           `json:"version"`
	Counts  PickerCounts  `json:"counts"`
	Groups  []PickerGroup `json:"groups"`
}

// PickerCounts summarizes selectable vs catalog-only entries.
type PickerCounts struct {
	WithTargets int `json:"with_targets"`
	CatalogOnly int `json:"catalog_only"`
	Total       int `json:"total"`
}

// PickerGroup is one category section in the picker.
type PickerGroup struct {
	Key   string       `json:"key"`
	Label string       `json:"label"`
	Items []PickerItem `json:"items"`
}

// PickerItem is one crop row in the picker.
type PickerItem struct {
	CropKey       string   `json:"crop_key"`
	DisplayName   string   `json:"display_name"`
	Category      string   `json:"category"`
	CropProfileID *int64   `json:"crop_profile_id,omitempty"`
	HasTargets    bool     `json:"has_targets"`
	IsCustom      bool     `json:"is_custom"`
	Substrate     string   `json:"substrate,omitempty"`
	WateringStyle string   `json:"watering_style,omitempty"`
	Aliases       []string `json:"aliases,omitempty"`
	CousinOf      *string  `json:"cousin_of,omitempty"`
	CousinLabel   string   `json:"cousin_label,omitempty"`
	ImageURL      string   `json:"image_url,omitempty"`
	SearchTerms   []string `json:"search_terms,omitempty"`
}

// ProfileRow is the minimal DB profile fields needed to build a picker.
type ProfileRow struct {
	ID          int64
	CropKey     string
	DisplayName string
	Category    *string
	IsBuiltin   bool
	FarmID      *int64
	StageCount  int
}

var categoryOrder = []struct {
	key   string
	label string
}{
	{"fruit_tree", "Fruit trees"},
	{"fruiting", "Fruiting"},
	{"leafy", "Leafy greens"},
	{"herb", "Herbs"},
	{"flower", "Flowers"},
	{"epiphyte", "Flower / epiphyte"},
	{"industrial", "Hemp / industrial"},
	{"ornamental", "Ornamental"},
	{"custom", "My farm profiles"},
}

// BuildPicker merges YAML catalog with farm/builtin crop profiles.
func BuildPicker(cat *Catalog, profiles []ProfileRow) PickerResponse {
	byKey := map[string]ProfileRow{}
	var customs []ProfileRow
	for _, p := range profiles {
		if p.FarmID != nil && !p.IsBuiltin {
			customs = append(customs, p)
			continue
		}
		key := strings.TrimSpace(p.CropKey)
		if key == "" {
			continue
		}
		if existing, ok := byKey[key]; !ok || profilePrefers(p, existing) {
			byKey[key] = p
		}
	}

	cousinLabels := map[string]string{}
	for _, crop := range cat.Crops {
		cousinLabels[crop.Key] = crop.DisplayName
	}

	groupItems := make(map[string][]PickerItem)
	var withTargets, catalogOnly int

	for _, crop := range cat.Crops {
		item := pickerItemFromCrop(crop, cat.Aliases, cousinLabels)
		if prof, ok := byKey[crop.Key]; ok && prof.StageCount > 0 {
			id := prof.ID
			item.CropProfileID = &id
			item.HasTargets = true
			withTargets++
		} else {
			catalogOnly++
		}
		catKey := normalizeCategory(crop.Category)
		groupItems[catKey] = append(groupItems[catKey], item)
	}

	for _, p := range customs {
		catKey := "custom"
		if p.Category != nil && strings.TrimSpace(*p.Category) != "" {
			catKey = normalizeCategory(*p.Category)
		}
		id := p.ID
		has := p.StageCount > 0
		item := PickerItem{
			CropKey:       p.CropKey,
			DisplayName:   p.DisplayName,
			Category:      catKey,
			CropProfileID: &id,
			HasTargets:    has,
			IsCustom:      true,
			SearchTerms:   searchTerms(p.CropKey, p.DisplayName, nil, nil),
		}
		groupItems[catKey] = append(groupItems[catKey], item)
		if has {
			withTargets++
		}
	}

	var groups []PickerGroup
	for _, co := range categoryOrder {
		items := groupItems[co.key]
		if len(items) == 0 {
			continue
		}
		sort.Slice(items, func(i, j int) bool {
			return strings.ToLower(items[i].DisplayName) < strings.ToLower(items[j].DisplayName)
		})
		groups = append(groups, PickerGroup{Key: co.key, Label: co.label, Items: items})
	}

	return PickerResponse{
		Version: cat.Version,
		Counts: PickerCounts{
			WithTargets: withTargets,
			CatalogOnly: catalogOnly,
			Total:       withTargets + catalogOnly,
		},
		Groups: groups,
	}
}

func pickerItemFromCrop(crop Crop, globalAliases map[string]string, cousinLabels map[string]string) PickerItem {
	item := PickerItem{
		CropKey:       crop.Key,
		DisplayName:   crop.DisplayName,
		Category:      crop.Category,
		Substrate:     crop.Substrate,
		WateringStyle: crop.WateringStyle,
		Aliases:       append([]string(nil), crop.Aliases...),
		CousinOf:      crop.CousinOf,
		ImageURL:      strings.TrimSpace(crop.ImageURL),
		SearchTerms:   searchTerms(crop.Key, crop.DisplayName, crop.Aliases, globalAliases),
	}
	if crop.CousinOf != nil && strings.TrimSpace(*crop.CousinOf) != "" {
		if lbl, ok := cousinLabels[*crop.CousinOf]; ok {
			item.CousinLabel = lbl
		}
	}
	return item
}

func profilePrefers(a, b ProfileRow) bool {
	if a.FarmID != nil && b.FarmID == nil {
		return true
	}
	return a.StageCount > b.StageCount
}

func normalizeCategory(category string) string {
	c := strings.ToLower(strings.TrimSpace(category))
	switch c {
	case "epiphyte":
		return "epiphyte"
	case "flower":
		return "flower"
	case "industrial":
		return "industrial"
	case "ornamental":
		return "ornamental"
	case "herb":
		return "herb"
	case "leafy":
		return "leafy"
	case "fruit_tree":
		return "fruit_tree"
	default:
		return "fruiting"
	}
}

func searchTerms(cropKey, displayName string, aliases []string, globalAliases map[string]string) []string {
	seen := make(map[string]struct{})
	add := func(s string) {
		s = strings.ToLower(strings.TrimSpace(s))
		if s == "" {
			return
		}
		seen[s] = struct{}{}
	}
	add(cropKey)
	add(displayName)
	for _, a := range aliases {
		add(a)
	}
	for alias, target := range globalAliases {
		if target == cropKey {
			add(alias)
		}
	}
	out := make([]string, 0, len(seen))
	for s := range seen {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}
