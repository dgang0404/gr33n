package croplibrary

import (
	"fmt"
	"sort"
	"strings"
)

// GenerateCatalogSeedSQL emits idempotent platform catalog + field guide seed SQL.
func GenerateCatalogSeedSQL(cat *Catalog, guides []FieldGuideSeed) string {
	var b strings.Builder
	b.WriteString("-- Generated from data/crop_library.yaml + docs/field-guides — do not edit by hand.\n")
	b.WriteString("-- Regenerate: ./scripts/generate-crop-catalog-seed.sql.sh\n")
	b.WriteString(fmt.Sprintf("-- crop_library version: %d\n\n", cat.Version))

	writeCatalogEntries(&b, cat)
	writeCatalogCousinUpdates(&b, cat)
	writeCatalogAliases(&b, cat)
	writeFieldGuides(&b, cat.Version, guides)
	return b.String()
}

func writeCatalogEntries(b *strings.Builder, cat *Catalog) {
	b.WriteString(`INSERT INTO gr33ncrops.crop_catalog_entries (
    crop_key, display_name, supported, category, source, substrate, watering_style,
    runoff_pct_target, moisture_guidance, cousin_of, unsupported_reason, catalog_version
)
SELECT v.crop_key, v.display_name, v.supported, v.category, v.source, v.substrate, v.watering_style,
       v.runoff_pct_target, v.moisture_guidance, v.cousin_of, v.unsupported_reason, v.catalog_version
FROM (VALUES
`)
	type row struct {
		key, display, category, source, substrate, watering, runoff, moisture, cousin, reason string
		supported                                                                              bool
	}
	var rows []row
	for _, c := range cat.Crops {
		rows = append(rows, row{
			key: c.Key, display: c.DisplayName, supported: true,
			category: c.Category, source: c.Source, substrate: c.Substrate,
			watering: c.WateringStyle, runoff: c.RunoffPctTarget, moisture: c.MoistureGuidance,
			cousin: cousinKey(c.CousinOf), reason: "",
		})
	}
	for _, u := range cat.Unsupported {
		display := strings.TrimSpace(u.DisplayName)
		if display == "" {
			display = u.Key
		}
		rows = append(rows, row{
			key: u.Key, display: display, supported: false,
			source: "", substrate: "", watering: "", runoff: "", moisture: "",
			cousin: cousinKey(u.CousinOf), reason: u.Reason,
		})
	}
	for i, r := range rows {
		comma := ","
		if i == len(rows)-1 {
			comma = ""
		}
		fmt.Fprintf(b, "    (%s, %s, %t, %s, %s, %s, %s, %s, %s, NULL, %s, %d)%s\n",
			sqlQuote(r.key), sqlQuote(r.display), r.supported,
			sqlQuoteNullStr(r.category), sqlQuoteNullStr(r.source),
			sqlQuoteNullStr(r.substrate), sqlQuoteNullStr(r.watering),
			sqlQuoteNullStr(r.runoff), sqlQuoteNullStr(r.moisture),
			sqlQuoteNullStr(r.reason), cat.Version, comma,
		)
	}
	b.WriteString(`) AS v(crop_key, display_name, supported, category, source, substrate, watering_style,
         runoff_pct_target, moisture_guidance, cousin_of, unsupported_reason, catalog_version)
ON CONFLICT (crop_key) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    supported = EXCLUDED.supported,
    category = EXCLUDED.category,
    source = EXCLUDED.source,
    substrate = EXCLUDED.substrate,
    watering_style = EXCLUDED.watering_style,
    runoff_pct_target = EXCLUDED.runoff_pct_target,
    moisture_guidance = EXCLUDED.moisture_guidance,
    unsupported_reason = EXCLUDED.unsupported_reason,
    catalog_version = EXCLUDED.catalog_version,
    updated_at = NOW();

`)
}

func writeCatalogCousinUpdates(b *strings.Builder, cat *Catalog) {
	type pair struct{ key, cousin string }
	var pairs []pair
	for _, c := range cat.Crops {
		if ck := cousinKey(c.CousinOf); ck != "" {
			pairs = append(pairs, pair{c.Key, ck})
		}
	}
	for _, u := range cat.Unsupported {
		if ck := cousinKey(u.CousinOf); ck != "" {
			pairs = append(pairs, pair{u.Key, ck})
		}
	}
	if len(pairs) == 0 {
		return
	}
	b.WriteString("-- cousin_of FK (second pass)\n")
	for _, p := range pairs {
		fmt.Fprintf(b, "UPDATE gr33ncrops.crop_catalog_entries SET cousin_of = %s WHERE crop_key = %s;\n",
			sqlQuote(p.cousin), sqlQuote(p.key))
	}
	b.WriteString("\n")
}

func writeCatalogAliases(b *strings.Builder, cat *Catalog) {
	type aliasRow struct{ alias, cropKey string }
	seen := make(map[string]struct{})
	var rows []aliasRow
	add := func(alias, cropKey string) {
		alias = strings.ToLower(strings.TrimSpace(alias))
		cropKey = strings.TrimSpace(cropKey)
		if alias == "" || cropKey == "" || alias == cropKey {
			return
		}
		if _, dup := seen[alias]; dup {
			return
		}
		seen[alias] = struct{}{}
		rows = append(rows, aliasRow{alias, cropKey})
	}
	for alias, target := range cat.Aliases {
		add(alias, target)
	}
	for _, c := range cat.Crops {
		for _, a := range c.Aliases {
			add(a, c.Key)
		}
	}
	for _, u := range cat.Unsupported {
		for _, a := range u.Aliases {
			add(a, u.Key)
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].alias != rows[j].alias {
			return rows[i].alias < rows[j].alias
		}
		return rows[i].cropKey < rows[j].cropKey
	})

	b.WriteString(`INSERT INTO gr33ncrops.crop_catalog_aliases (alias, crop_key)
SELECT v.alias, v.crop_key
FROM (VALUES
`)
	for i, r := range rows {
		comma := ","
		if i == len(rows)-1 {
			comma = ""
		}
		fmt.Fprintf(b, "    (%s, %s)%s\n", sqlQuote(r.alias), sqlQuote(r.cropKey), comma)
	}
	b.WriteString(`) AS v(alias, crop_key)
ON CONFLICT (alias) DO UPDATE SET crop_key = EXCLUDED.crop_key;

`)
}

func writeFieldGuides(b *strings.Builder, version int, guides []FieldGuideSeed) {
	if len(guides) == 0 {
		return
	}
	b.WriteString(`INSERT INTO gr33ncrops.agronomy_field_guides (
    slug, title, crop_key, guide_kind, domain, safety_tier, body_md, catalog_version, published, sort_order
)
SELECT v.slug, v.title, v.crop_key, v.guide_kind, v.domain, v.safety_tier, v.body_md, v.catalog_version, v.published, v.sort_order
FROM (VALUES
`)
	for i, g := range guides {
		comma := ","
		if i == len(guides)-1 {
			comma = ""
		}
		ck := g.CropKey
		fmt.Fprintf(b, "    (%s, %s, %s, %s, %s, %s, %s, %d, TRUE, %d)%s\n",
			sqlQuote(g.Slug), sqlQuote(g.Title), sqlQuoteNullStr(ck),
			sqlQuote(g.GuideKind), sqlQuoteNullStr(g.Domain), sqlQuote(g.SafetyTier),
			sqlDollarQuote(g.Slug, g.BodyMD), version, g.SortOrder, comma,
		)
	}
	b.WriteString(`) AS v(slug, title, crop_key, guide_kind, domain, safety_tier, body_md, catalog_version, published, sort_order)
ON CONFLICT (slug) DO UPDATE SET
    title = EXCLUDED.title,
    crop_key = EXCLUDED.crop_key,
    guide_kind = EXCLUDED.guide_kind,
    domain = EXCLUDED.domain,
    safety_tier = EXCLUDED.safety_tier,
    body_md = EXCLUDED.body_md,
    catalog_version = EXCLUDED.catalog_version,
    published = EXCLUDED.published,
    sort_order = EXCLUDED.sort_order,
    updated_at = NOW();

`)
}

func cousinKey(p *string) string {
	if p == nil {
		return ""
	}
	return strings.TrimSpace(*p)
}

func sqlQuoteNullStr(s string) string {
	if strings.TrimSpace(s) == "" {
		return "NULL"
	}
	return sqlQuote(s)
}

func sqlDollarQuote(tag, content string) string {
	tag = "fg_" + strings.ReplaceAll(tag, "-", "_")
	for strings.Contains(content, "$"+tag+"$") {
		tag = tag + "_"
	}
	return "$" + tag + "$" + content + "$" + tag + "$"
}
