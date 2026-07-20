// Package naturalfarmingcatalog loads Phase 208 YAML canon (process materials + recipes).
package naturalfarmingcatalog

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	MaterialCatalogPath = "data/process-material-catalog.yaml"
	RecipeCanonPath       = "data/recipe-canonical.yaml"
)

func loadYAMLMap(repoRoot, relPath string) (map[string]any, error) {
	root := strings.TrimSpace(repoRoot)
	if root == "" {
		root = "."
	}
	abs := relPath
	if !filepath.IsAbs(relPath) {
		abs = filepath.Join(root, relPath)
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", relPath, err)
	}
	var out map[string]any
	if err := yaml.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("parse %s: %w", relPath, err)
	}
	if out == nil {
		return nil, fmt.Errorf("parse %s: empty document", relPath)
	}
	return out, nil
}

// LoadMaterialCatalog reads data/process-material-catalog.yaml under repoRoot.
func LoadMaterialCatalog(repoRoot string) (map[string]any, error) {
	cat, err := loadYAMLMap(repoRoot, MaterialCatalogPath)
	if err != nil {
		return nil, err
	}
	if !yamlVersionAtLeast1(cat["version"]) {
		return nil, fmt.Errorf("%s: version must be >= 1", MaterialCatalogPath)
	}
	return cat, nil
}

// LoadRecipeCanon reads data/recipe-canonical.yaml under repoRoot.
func LoadRecipeCanon(repoRoot string) (map[string]any, error) {
	cat, err := loadYAMLMap(repoRoot, RecipeCanonPath)
	if err != nil {
		return nil, err
	}
	if !yamlVersionAtLeast1(cat["version"]) {
		return nil, fmt.Errorf("%s: version must be >= 1", RecipeCanonPath)
	}
	return cat, nil
}

// MaterialByID returns one material entry from a loaded material catalog.
func MaterialByID(catalog map[string]any, id string) (map[string]any, bool) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, false
	}
	raw, ok := catalog["materials"]
	if !ok {
		return nil, false
	}
	items, ok := raw.([]any)
	if !ok {
		return nil, false
	}
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		sid, _ := m["id"].(string)
		if strings.TrimSpace(sid) == id {
			return m, true
		}
	}
	return nil, false
}

func yamlVersionAtLeast1(v any) bool {
	switch n := v.(type) {
	case int:
		return n >= 1
	case int64:
		return n >= 1
	case float64:
		return n >= 1
	default:
		return false
	}
}
