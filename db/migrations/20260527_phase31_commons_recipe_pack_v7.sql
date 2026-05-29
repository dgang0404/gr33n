-- Phase 31 WS5 — demo Recipe Pack v7 in commons catalog (fertigation program payload + readme).
-- Source JSON mirror: scripts/enterprise/sample-recipe-pack-v7.body.json

INSERT INTO gr33ncore.commons_catalog_entries (
    slug, title, summary, body, contributor_display, contributor_uri,
    license_spdx, license_notes, tags, published, sort_order
) VALUES (
    'gr33n-recipe-pack-v7-lettuce-veg',
    'Recipe Pack v7 — Lettuce veg / flower (demo)',
    'Hypothetical HQ recipe promotion: fertigation program definitions as opaque JSON; apply per farm via import-recipe-pack.sh.',
    $body$
{
  "catalog_version": "gr33n.commons_catalog.v1",
  "kind": "fertigation_recipe_pack",
  "pack_version": "7",
  "pack_id": "gr33n-recipe-pack-v7-lettuce-veg",
  "readme_md": "# Recipe Pack v7 — Lettuce veg / flower (demo)\n\nHypothetical **HQ → site** promotion pack for multi-farm integrators. This JSON is **opaque payload** stored in `commons_catalog_entries.body` — the API does not auto-apply it.\n\n## What integrators do\n\n1. Curator publishes this row in the commons catalog.\n2. Per site, farm admin runs `scripts/enterprise/import-recipe-pack.sh` to record catalog import audit + create programs via public API.\n\nPrograms import with **`is_active: false`** — review before enabling automation.\n",
  "promotion_note": "Phase 31 WS5 demo — not a live agronomic prescription.",
  "programs": [
    {
      "pack_program_key": "lettuce-veg-standard",
      "name": "Recipe Pack v7 — Lettuce Veg Standard",
      "description": "Conservative veg-stage EC/pH triggers for indoor lettuce (demo pack).",
      "total_volume_liters": 2.0,
      "ec_trigger_low": 1.0,
      "ph_trigger_low": 5.8,
      "ph_trigger_high": 6.4,
      "is_active": false,
      "target_zone_name_hint": "Veg Room"
    },
    {
      "pack_program_key": "lettuce-flower-boost",
      "name": "Recipe Pack v7 — Lettuce Flower Boost",
      "description": "Flower-transition feed profile (demo pack). Enable only after operator review.",
      "total_volume_liters": 1.5,
      "ec_trigger_low": 1.4,
      "ph_trigger_low": 5.9,
      "ph_trigger_high": 6.3,
      "is_active": false,
      "target_zone_name_hint": "Flower Room"
    }
  ]
}
$body$::jsonb,
    'gr33n platform (demo pack)',
    NULL,
    'CC-BY-4.0',
    'Demo integrator pack for documentation; verify agronomy locally before production use.',
    ARRAY['fertigation', 'recipe-pack', 'enterprise', 'commons', 'phase-31'],
    TRUE,
    10
) ON CONFLICT (slug) DO NOTHING;
