package croplibrary

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	DefaultFieldGuideManifest = "docs/rag/field-guide-manifest.yaml"
	FieldGuidesDir            = "docs/field-guides"
)

// FieldGuideSeed is one agronomy field guide row for SQL seed generation.
type FieldGuideSeed struct {
	Slug           string
	Title          string
	CropKey        string
	GuideKind      string
	Domain         string
	SafetyTier     string
	BodyMD         string
	CatalogVersion int
	SortOrder      int
}

// LoadFieldGuideSeeds reads manifest + markdown bodies from repoRoot.
// When cat is non-nil, crop_key is resolved via catalog keys and global aliases.
func LoadFieldGuideSeeds(repoRoot, manifestPath string, cat *Catalog) ([]FieldGuideSeed, error) {
	if strings.TrimSpace(manifestPath) == "" {
		manifestPath = DefaultFieldGuideManifest
	}
	absManifest := manifestPath
	if !filepath.IsAbs(manifestPath) {
		absManifest = filepath.Join(repoRoot, manifestPath)
	}
	data, err := os.ReadFile(absManifest)
	if err != nil {
		return nil, fmt.Errorf("read field guide manifest: %w", err)
	}
	include, _, err := parseManifestIncludeList(string(data))
	if err != nil {
		return nil, err
	}
	guidesDir := filepath.Join(repoRoot, FieldGuidesDir)
	var out []FieldGuideSeed
	seen := make(map[string]struct{})
	for i, rel := range include {
		rel = strings.TrimSpace(strings.TrimPrefix(rel, "field-guides/"))
		if rel == "" {
			continue
		}
		if _, dup := seen[rel]; dup {
			continue
		}
		seen[rel] = struct{}{}
		raw, err := os.ReadFile(filepath.Join(guidesDir, rel))
		if err != nil {
			return nil, fmt.Errorf("read field guide %s: %w", rel, err)
		}
		body, meta := splitYAMLFrontmatter(string(raw))
		title := strings.TrimSpace(meta["title"])
		if title == "" {
			title = slugToTitle(strings.TrimSuffix(rel, ".md"))
		}
		slug := strings.TrimSuffix(rel, ".md")
		domain, safety := fieldGuideMetaDefaults(slug, meta)
		out = append(out, FieldGuideSeed{
			Slug:           slug,
			Title:          title,
			CropKey:        resolveGuideCropKey(inferGuideCropKey(slug), cat),
			GuideKind:      inferGuideKind(slug),
			Domain:         domain,
			SafetyTier:     safety,
			BodyMD:         strings.TrimSpace(body),
			CatalogVersion: 0, // filled by caller from catalog version
			SortOrder:      i,
		})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("field guide manifest resolved zero files")
	}
	return out, nil
}

func parseManifestIncludeList(raw string) (include, exclude []string, err error) {
	section := ""
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasSuffix(line, ":") {
			key := strings.TrimSuffix(line, ":")
			switch key {
			case "include":
				section = "include"
			case "exclude_globs":
				section = "exclude"
			default:
				section = ""
			}
			continue
		}
		switch section {
		case "include":
			line = strings.TrimPrefix(line, "- ")
			line = strings.TrimSpace(line)
			include = append(include, strings.Trim(line, `"'`))
		case "exclude":
			line = strings.TrimPrefix(line, "- ")
			line = strings.TrimSpace(line)
			exclude = append(exclude, strings.Trim(line, `"'`))
		}
	}
	if len(include) == 0 {
		return nil, nil, fmt.Errorf("manifest include list empty")
	}
	return include, exclude, nil
}

func splitYAMLFrontmatter(raw string) (body string, meta map[string]string) {
	raw = strings.TrimPrefix(raw, "\ufeff")
	if !strings.HasPrefix(raw, "---") {
		return raw, nil
	}
	rest := strings.TrimPrefix(raw, "---")
	rest = strings.TrimPrefix(rest, "\n")
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return raw, nil
	}
	header := rest[:end]
	body = strings.TrimSpace(rest[end+len("\n---"):])
	meta = make(map[string]string)
	for _, line := range strings.Split(header, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		meta[strings.TrimSpace(k)] = strings.Trim(strings.TrimSpace(v), `"'`)
	}
	return body, meta
}

func fieldGuideMetaDefaults(slug string, front map[string]string) (domain, safety string) {
	domain = strings.TrimSpace(front["domain"])
	safety = strings.TrimSpace(strings.ToLower(front["safety_tier"]))
	if domain == "" {
		switch {
		case strings.Contains(slug, "electrical"):
			domain = "electrical"
		case strings.Contains(slug, "irrigation"), strings.Contains(slug, "plumb"):
			domain = "plumbing"
		case strings.Contains(slug, "sensor"):
			domain = "sensor"
		case strings.Contains(slug, "relay"), strings.Contains(slug, "actuator"):
			domain = "actuator"
		case strings.Contains(slug, "pi"):
			domain = "pi"
		default:
			domain = "general"
		}
	}
	if safety == "" {
		if domain == "electrical" || strings.Contains(slug, "electrical-safety") {
			safety = "caution"
		} else {
			safety = "safe"
		}
	}
	return domain, safety
}

func inferGuideKind(slug string) string {
	switch {
	case strings.HasPrefix(slug, "crop-unsupported-"):
		return "unsupported"
	case strings.HasPrefix(slug, "crop-"):
		if strings.Contains(slug, "-care") {
			return "crop_care"
		}
		return "crop_nutrition"
	default:
		return "trades"
	}
}

func inferGuideCropKey(slug string) string {
	if !strings.HasPrefix(slug, "crop-") || strings.HasPrefix(slug, "crop-unsupported-") {
		return ""
	}
	base := strings.TrimPrefix(slug, "crop-")
	for _, suffix := range []string{"-nutrition", "-care", "-nursery", "-container", "-vine"} {
		if strings.HasSuffix(base, suffix) {
			key := strings.TrimSuffix(base, suffix)
			key = strings.ReplaceAll(key, "-", "_")
			return key
		}
	}
	return strings.ReplaceAll(base, "-", "_")
}

func slugToTitle(slug string) string {
	s := strings.ReplaceAll(slug, "-", " ")
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func resolveGuideCropKey(inferred string, cat *Catalog) string {
	inferred = strings.TrimSpace(inferred)
	if inferred == "" || cat == nil {
		return inferred
	}
	if cat.hasCropKey(inferred) {
		return inferred
	}
	if target, ok := cat.Aliases[strings.ToLower(inferred)]; ok && cat.hasCropKey(target) {
		return target
	}
	return ""
}

func (c *Catalog) hasCropKey(key string) bool {
	if c == nil {
		return false
	}
	if c.cropKeys == nil {
		c.buildIndexes()
	}
	_, ok := c.cropKeys[key]
	return ok
}
