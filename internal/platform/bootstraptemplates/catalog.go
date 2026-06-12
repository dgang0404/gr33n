// Package bootstraptemplates loads gr33ncore.bootstrap_templates (Phase 91).
package bootstraptemplates

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Template is one bootstrap starter pack row.
type Template struct {
	TemplateKey        string   `json:"template_key"`
	Label              string   `json:"label"`
	ShortLabel         string   `json:"short_label,omitempty"`
	Tagline            string   `json:"tagline,omitempty"`
	SummaryTitle       string   `json:"summary_title"`
	SummaryBullets     []string `json:"summary_bullets"`
	ModuleHints        []string `json:"module_hints,omitempty"`
	Icon               string   `json:"icon,omitempty"`
	Recommended        bool     `json:"recommended"`
	WizardPrimary      bool     `json:"wizard_primary"`
	PlaybookSection    string   `json:"playbook_section,omitempty"`
	RelatedCommonsSlug string   `json:"related_commons_slug,omitempty"`
	SortOrder          int      `json:"sort_order"`
}

// Catalog is an indexed bootstrap template list.
type Catalog struct {
	byKey map[string]Template
	list  []Template
}

var (
	cacheMu sync.RWMutex
	cached  *Catalog
)

// Load reads active templates from Postgres.
func Load(ctx context.Context, pool *pgxpool.Pool) (*Catalog, error) {
	if pool == nil {
		return nil, fmt.Errorf("bootstrap templates: nil pool")
	}
	cacheMu.RLock()
	if cached != nil {
		c := cached
		cacheMu.RUnlock()
		return c, nil
	}
	cacheMu.RUnlock()

	rows, err := pool.Query(ctx, `
SELECT template_key, label, COALESCE(short_label, ''), COALESCE(tagline, ''), summary_title,
       summary_bullets, module_hints, COALESCE(icon, ''), recommended, wizard_primary,
       COALESCE(playbook_section, ''), COALESCE(related_commons_slug, ''), sort_order
FROM gr33ncore.bootstrap_templates
WHERE is_active = TRUE
ORDER BY sort_order, template_key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cat := &Catalog{byKey: make(map[string]Template)}
	for rows.Next() {
		var t Template
		var bulletsJSON, hintsJSON []byte
		if err := rows.Scan(
			&t.TemplateKey, &t.Label, &t.ShortLabel, &t.Tagline, &t.SummaryTitle,
			&bulletsJSON, &hintsJSON, &t.Icon, &t.Recommended, &t.WizardPrimary,
			&t.PlaybookSection, &t.RelatedCommonsSlug, &t.SortOrder,
		); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(bulletsJSON, &t.SummaryBullets)
		_ = json.Unmarshal(hintsJSON, &t.ModuleHints)
		cat.byKey[t.TemplateKey] = t
		cat.list = append(cat.list, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	cacheMu.Lock()
	cached = cat
	cacheMu.Unlock()
	return cat, nil
}

// ResetCache clears the in-process cache (tests).
func ResetCache() {
	cacheMu.Lock()
	cached = nil
	cacheMu.Unlock()
}

// Current returns DB-backed catalog when loaded, else embedded seed.
func Current() *Catalog {
	cacheMu.RLock()
	defer cacheMu.RUnlock()
	if cached != nil {
		return cached
	}
	return embeddedCatalog()
}

func (c *Catalog) List() []Template {
	if c == nil {
		return nil
	}
	out := make([]Template, len(c.list))
	copy(out, c.list)
	return out
}

// IsValid reports whether template_key can be applied via apply_farm_bootstrap_template.
func (c *Catalog) IsValid(templateKey string) bool {
	if c == nil {
		return false
	}
	_, ok := c.byKey[templateKey]
	return ok
}

// LoadList is a convenience for handlers.
func LoadList(ctx context.Context, pool *pgxpool.Pool) ([]Template, error) {
	cat, err := Load(ctx, pool)
	if err != nil {
		return embeddedCatalog().List(), nil
	}
	if cat == nil || len(cat.list) == 0 {
		return embeddedCatalog().List(), nil
	}
	return cat.List(), nil
}
