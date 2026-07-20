-- Phase 211 WS3 — livestock comfrey/sprout feed pack + demo farm seed rows.
-- Body mirror: data/natural-farming-packs/livestock_comfrey_feed_v1.json

INSERT INTO gr33ncore.commons_catalog_entries (
    slug, title, summary, body, contributor_display, contributor_uri,
    license_spdx, license_notes, tags, published, sort_order
) VALUES (
    'livestock-comfrey-feed-v1',
    'Livestock comfrey & sprout feed v1',
    'Simple animal_feed inputs — comfrey slurry and sprouted grain — with flock supplement examples. Requires Animals module.',
    $body$
{
  "catalog_version": "gr33n.commons_catalog.v1",
  "kind": "natural_farming_recipe_pack",
  "pack_key": "livestock_comfrey_feed_v1",
  "reference_source": "Operator practice; gr33n animal_feed category — not full ration math",
  "readme_md": "# Livestock comfrey & sprout feed v1\n\nSimple **animal_feed** inputs — comfrey slurry and sprouted grain — with flock supplement examples. **Not** TMR balancing or veterinary formulation.\n\nApply via `POST /farms/{id}/naturalfarming/apply-pack` when Animals module is enabled.\n",
  "input_definitions": [
    {
      "name": "Comfrey Slurry (Livestock Supplement)",
      "category": "animal_feed",
      "description": "Fresh comfrey leaves blended or soaked to slurry for flock supplement. Not sole ration.",
      "typical_ingredients": "Fresh comfrey leaves, water",
      "preparation_summary": "Harvest comfrey; chop; soak or blend with water; feed fresh within 24 h.",
      "storage_guidelines": "Use same day — do not store slurry long; anaerobic spoilage risk.",
      "safety_precautions": "Comfrey contains pyrrolizidine alkaloids — moderation for poultry; not majority of ration.",
      "reference_source": "Operator practice; see natural-farming-livestock-plant-feed.md"
    },
    {
      "name": "Sprouted Grain (Livestock Supplement)",
      "category": "animal_feed",
      "description": "Soaked and sprouted grain (barley, wheat, or similar) as a treat or supplement.",
      "typical_ingredients": "Whole grain, water",
      "preparation_summary": "Soak grain 8–12 h, drain, sprout 2–5 days with daily rinse until short tails appear.",
      "storage_guidelines": "Refrigerate sprouts max 1–2 days; discard moldy trays.",
      "safety_precautions": "Moldy sprouts — discard. Treat/supplement only with balanced forage or ration.",
      "reference_source": "Operator practice; see natural-farming-livestock-plant-feed.md"
    }
  ],
  "application_recipes": [
    {
      "name": "Comfrey Slurry Flock Supplement",
      "description": "Daily comfrey slurry treat for chickens or goats — small supplement only.",
      "target_application_type": "livestock_water_supplement",
      "dilution_ratio": "Fresh slurry — operator volume by flock size",
      "instructions": "Offer in shallow trough; remove uneaten slurry within a few hours.",
      "frequency_guidelines": "2–3 times per week as treat — not primary ration.",
      "primary_input_name": "Comfrey Slurry (Livestock Supplement)"
    },
    {
      "name": "Sprouted Grain Treat Batch",
      "description": "Short-tail sprouted grain offered as flock treat or supplement.",
      "target_application_type": "livestock_water_supplement",
      "dilution_ratio": "Whole sprouts — handful per bird or small bowl for goats",
      "instructions": "Feed when white root tails appear; discard mold immediately.",
      "frequency_guidelines": "Daily small treat during active sprouting cycle.",
      "primary_input_name": "Sprouted Grain (Livestock Supplement)"
    }
  ],
  "recipe_input_components": [
    {
      "recipe_name": "Comfrey Slurry Flock Supplement",
      "input_name": "Comfrey Slurry (Livestock Supplement)",
      "part_value": 1.0,
      "part_unit_name": "decimal_fraction",
      "notes": "Fresh slurry volume by flock size"
    },
    {
      "recipe_name": "Sprouted Grain Treat Batch",
      "input_name": "Sprouted Grain (Livestock Supplement)",
      "part_value": 1.0,
      "part_unit_name": "decimal_fraction",
      "notes": "Whole sprouts — treat portion only"
    }
  ]
}
$body$::jsonb,
    'gr33n platform (natural farming extension)',
    NULL,
    'CC-BY-4.0',
    'Extension method — not veterinary formulation or TMR balancing.',
    ARRAY['natural_farming', 'livestock', 'animal_feed', 'commons', 'phase-211'],
    TRUE,
    36
) ON CONFLICT (slug) DO UPDATE SET
    title = EXCLUDED.title,
    summary = EXCLUDED.summary,
    body = EXCLUDED.body,
    license_spdx = EXCLUDED.license_spdx,
    license_notes = EXCLUDED.license_notes,
    tags = EXCLUDED.tags,
    published = EXCLUDED.published,
    sort_order = EXCLUDED.sort_order,
    updated_at = NOW();

-- Demo farm (id=1): seed animal_feed inputs if missing (idempotent).
INSERT INTO gr33nnaturalfarming.input_definitions
    (farm_id, name, category, description, typical_ingredients,
     preparation_summary, storage_guidelines, safety_precautions, reference_source)
SELECT 1, x.name, x.cat::gr33nnaturalfarming.input_category_enum, x.descr, x.ting, x.prep, x.store, x.safe, x.ref
FROM (
    VALUES
        ('Comfrey Slurry (Livestock Supplement)', 'animal_feed',
         'Fresh comfrey leaves blended or soaked to slurry for flock supplement. Not sole ration.',
         'Fresh comfrey leaves, water',
         'Harvest comfrey; chop; soak or blend with water; feed fresh within 24 h.',
         'Use same day — do not store slurry long.',
         'Comfrey contains pyrrolizidine alkaloids — moderation for poultry.',
         'Operator practice; see natural-farming-livestock-plant-feed.md'),
        ('Sprouted Grain (Livestock Supplement)', 'animal_feed',
         'Soaked and sprouted grain as a treat or supplement.',
         'Whole grain, water',
         'Soak 8–12 h, drain, sprout 2–5 days with daily rinse.',
         'Refrigerate sprouts max 1–2 days; discard moldy trays.',
         'Moldy sprouts — discard. Treat/supplement only.',
         'Operator practice; see natural-farming-livestock-plant-feed.md')
) AS x(name, cat, descr, ting, prep, store, safe, ref)
WHERE NOT EXISTS (
    SELECT 1 FROM gr33nnaturalfarming.input_definitions d
    WHERE d.farm_id = 1 AND d.name = x.name AND d.deleted_at IS NULL
);
