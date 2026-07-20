// Package naturalfarmingpack applies Phase 211 switchover packs (subset of starter canon).
package naturalfarmingpack

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	catalogpack "gr33n-api/internal/commonscatalog"
	"gr33n-api/internal/croplibrary"
	db "gr33n-api/internal/db"
)

const (
	SwitchoverPacksPath     = "data/natural-farming-packs/switchover-packs.yaml"
	StarterRecipePackPath   = "data/natural-farming-packs/jadam_indoor_starter_recipes_v1.json"
)

// SwitchoverPackSpec is one entry in switchover-packs.yaml.
type SwitchoverPackSpec struct {
	Title               string   `yaml:"title"`
	CommercialPattern   string   `yaml:"commercial_pattern"`
	InputNames          []string `yaml:"input_names"`
	ApplicationRecipes  []string `yaml:"application_recipes"`
	ProgramZoneHints    []string `yaml:"program_zone_hints"`
	NextSteps           []string `yaml:"next_steps"`
}

type switchoverCatalog struct {
	Version    int                              `yaml:"version"`
	SourcePack string                           `yaml:"source_pack"`
	Packs      map[string]SwitchoverPackSpec    `yaml:"packs"`
}

// ApplyResult wraps commons apply output with switchover metadata.
type ApplyResult struct {
	PackKey          string                    `json:"pack_key"`
	Title            string                    `json:"title"`
	CommercialPattern string                   `json:"commercial_pattern,omitempty"`
	Status           string                    `json:"status"` // applied | already_applied | noop
	Message          string                    `json:"message"`
	Apply            catalogpack.ApplyResult   `json:"apply"`
	ProgramHints     []ProgramHint             `json:"program_hints,omitempty"`
	NextSteps        []string                  `json:"next_steps,omitempty"`
}

// ProgramHint surfaces existing fertigation programs that may overlap the switchover.
type ProgramHint struct {
	ProgramID   int64  `json:"program_id"`
	ProgramName string `json:"program_name"`
	IsActive    bool   `json:"is_active"`
	ZoneHint    string `json:"zone_hint,omitempty"`
}

// LoadSwitchoverCatalog reads switchover-packs.yaml under repoRoot.
func LoadSwitchoverCatalog(repoRoot string) (map[string]SwitchoverPackSpec, error) {
	root := strings.TrimSpace(repoRoot)
	if root == "" {
		root = "."
	}
	path := filepath.Join(root, SwitchoverPacksPath)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", SwitchoverPacksPath, err)
	}
	var doc switchoverCatalog
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse %s: %w", SwitchoverPacksPath, err)
	}
	if doc.Version < 1 {
		return nil, fmt.Errorf("%s: version must be >= 1", SwitchoverPacksPath)
	}
	if len(doc.Packs) == 0 {
		return nil, fmt.Errorf("%s: no packs defined", SwitchoverPacksPath)
	}
	return doc.Packs, nil
}

// LoadStarterPackBody loads the full audited starter JSON pack.
func LoadStarterPackBody(repoRoot string) (catalogpack.PackBody, error) {
	root := strings.TrimSpace(repoRoot)
	if root == "" {
		root = "."
	}
	path := filepath.Join(root, StarterRecipePackPath)
	raw, err := os.ReadFile(path)
	if err != nil {
		return catalogpack.PackBody{}, fmt.Errorf("read %s: %w", StarterRecipePackPath, err)
	}
	body, err := catalogpack.ParsePackBody(raw)
	if err != nil {
		return catalogpack.PackBody{}, err
	}
	if body.Kind != catalogpack.KindNaturalFarmingRecipePack {
		return catalogpack.PackBody{}, fmt.Errorf("%s: want kind %s", StarterRecipePackPath, catalogpack.KindNaturalFarmingRecipePack)
	}
	return body, nil
}

// FilterStarterPack returns a subset pack body for one switchover key.
func FilterStarterPack(full catalogpack.PackBody, spec SwitchoverPackSpec) catalogpack.PackBody {
	recipeSet := nameSet(spec.ApplicationRecipes)
	inputSet := nameSet(spec.InputNames)

	var components []catalogpack.NFRecipeComponentSpec
	for _, c := range full.RecipeInputComponents {
		if _, ok := recipeSet[strings.TrimSpace(c.RecipeName)]; !ok {
			continue
		}
		components = append(components, c)
		inputSet[strings.TrimSpace(c.InputName)] = struct{}{}
	}

	var inputs []catalogpack.NFInputDefinitionSpec
	for _, inp := range full.InputDefinitions {
		if _, ok := inputSet[strings.TrimSpace(inp.Name)]; ok {
			inputs = append(inputs, inp)
		}
	}
	var recipes []catalogpack.NFApplicationRecipeSpec
	for _, rec := range full.ApplicationRecipes {
		if _, ok := recipeSet[strings.TrimSpace(rec.Name)]; ok {
			recipes = append(recipes, rec)
		}
	}

	return catalogpack.PackBody{
		CatalogVersion:        catalogpack.CatalogVersion,
		Kind:                  catalogpack.KindNaturalFarmingRecipePack,
		PackKey:               full.PackKey,
		ReferenceSource:       full.ReferenceSource,
		ReadmeMD:              full.ReadmeMD,
		InputDefinitions:      inputs,
		ApplicationRecipes:    recipes,
		RecipeInputComponents: components,
	}
}

// ApplyPack applies a named switchover pack to a farm (idempotent on input/recipe names).
func ApplyPack(ctx context.Context, q db.Querier, repoRoot string, farmID int64, packKey string) (ApplyResult, error) {
	packKey = strings.TrimSpace(packKey)
	if packKey == "" {
		return ApplyResult{}, fmt.Errorf("pack_key is required")
	}
	catalog, err := LoadSwitchoverCatalog(repoRoot)
	if err != nil {
		return ApplyResult{}, err
	}
	spec, ok := catalog[packKey]
	if !ok {
		return ApplyResult{}, fmt.Errorf("unknown switchover pack %q", packKey)
	}
	starter, err := LoadStarterPackBody(repoRoot)
	if err != nil {
		return ApplyResult{}, err
	}
	filtered := FilterStarterPack(starter, spec)
	if err := catalogpack.ValidatePublishBody(filtered); err != nil {
		return ApplyResult{}, fmt.Errorf("pack %q: %w", packKey, err)
	}
	raw, err := json.Marshal(filtered)
	if err != nil {
		return ApplyResult{}, err
	}
	applyRes, err := catalogpack.ApplyPack(ctx, q, farmID, raw)
	if err != nil {
		return ApplyResult{
			PackKey:           packKey,
			Title:             spec.Title,
			CommercialPattern: spec.CommercialPattern,
			Status:            "failed",
			Message:           err.Error(),
			Apply:             applyRes,
		}, err
	}

	res := ApplyResult{
		PackKey:           packKey,
		Title:             spec.Title,
		CommercialPattern: spec.CommercialPattern,
		Apply:             applyRes,
		NextSteps:         append([]string(nil), spec.NextSteps...),
	}
	if len(res.NextSteps) == 0 {
		res.NextSteps = applyRes.NextSteps
	}

	hints, hErr := programHints(ctx, q, farmID, spec.ProgramZoneHints)
	if hErr == nil && len(hints) > 0 {
		res.ProgramHints = hints
	}

	switch {
	case applyRes.InputsCreated == 0 && applyRes.RecipesCreated == 0 &&
		applyRes.InputsSkipped > 0 && applyRes.RecipesSkipped > 0:
		res.Status = "already_applied"
		res.Message = fmt.Sprintf("%s — definitions already on this farm; components refreshed.", spec.Title)
	case applyRes.Status == "noop":
		res.Status = "noop"
		res.Message = applyRes.Message
	default:
		res.Status = "applied"
		res.Message = fmt.Sprintf("%s applied — review Natural Farming → Recipes.", spec.Title)
	}
	return res, nil
}

// ApplyPackFromRepo finds repo root and applies packKey.
func ApplyPackFromRepo(ctx context.Context, q db.Querier, farmID int64, packKey string) (ApplyResult, error) {
	root, err := croplibrary.FindRepoRoot()
	if err != nil {
		return ApplyResult{}, err
	}
	return ApplyPack(ctx, q, root, farmID, packKey)
}

func programHints(ctx context.Context, q db.Querier, farmID int64, zoneHints []string) ([]ProgramHint, error) {
	if len(zoneHints) == 0 {
		return nil, nil
	}
	programs, err := q.ListProgramsByFarm(ctx, farmID)
	if err != nil {
		return nil, err
	}
	hints := make([]string, 0, len(zoneHints))
	for _, h := range zoneHints {
		h = strings.TrimSpace(h)
		if h != "" {
			hints = append(hints, strings.ToLower(h))
		}
	}
	if len(hints) == 0 {
		return nil, nil
	}
	var out []ProgramHint
	for _, p := range programs {
		nameLower := strings.ToLower(p.Name)
		for _, hint := range hints {
			if strings.Contains(nameLower, hint) {
				out = append(out, ProgramHint{
					ProgramID:   p.ID,
					ProgramName: p.Name,
					IsActive:    p.IsActive,
					ZoneHint:    hint,
				})
				break
			}
		}
	}
	return out, nil
}

func nameSet(names []string) map[string]struct{} {
	out := make(map[string]struct{}, len(names))
	for _, n := range names {
		n = strings.TrimSpace(n)
		if n != "" {
			out[n] = struct{}{}
		}
	}
	return out
}
