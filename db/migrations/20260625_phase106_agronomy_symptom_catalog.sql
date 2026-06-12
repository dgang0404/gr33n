-- Phase 106 — structured deficiency & pest symptom catalog for Guardian lookup + RAG.

CREATE TABLE IF NOT EXISTS gr33ncrops.agronomy_symptom_entries (
    id              BIGSERIAL PRIMARY KEY,
    symptom_key     TEXT NOT NULL UNIQUE,
    display_name    TEXT NOT NULL,
    crop_keys       TEXT[] NOT NULL DEFAULT '{}',
    categories      TEXT[] NOT NULL DEFAULT '{}',
    body_md         TEXT NOT NULL,
    severity_hint   TEXT,
    catalog_version INTEGER NOT NULL DEFAULT 1,
    published       BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order      INTEGER NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_agronomy_symptom_crop_keys
    ON gr33ncrops.agronomy_symptom_entries USING GIN (crop_keys);

CREATE INDEX IF NOT EXISTS idx_agronomy_symptom_categories
    ON gr33ncrops.agronomy_symptom_entries USING GIN (categories);

DROP TRIGGER IF EXISTS trg_agronomy_symptom_entries_updated_at ON gr33ncrops.agronomy_symptom_entries;
CREATE TRIGGER trg_agronomy_symptom_entries_updated_at
    BEFORE UPDATE ON gr33ncrops.agronomy_symptom_entries
    FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();

INSERT INTO gr33ncrops.agronomy_symptom_entries (
    symptom_key, display_name, crop_keys, categories, body_md, severity_hint, sort_order
) VALUES
(
    'interveinal_yellowing',
    'Interveinal yellowing (lower/mid leaves)',
    ARRAY['tomato', 'pepper', 'cannabis', 'eggplant', 'cucumber'],
    ARRAY['fruiting', 'vegetables'],
    E'**Observation:** Leaf tissue between veins turns yellow while veins stay green.\n\n**Hypotheses (confirm with tests):**\n- Magnesium deficiency (common in fruiting tomatoes/peppers on heavy fruit load)\n- Iron deficiency (often upper leaves; check pH lockout)\n- Overwatering / root stress mimicking deficiency\n\n**Checks:** Compare feed EC and pH to stage targets (lookup_crop_targets). Inspect runoff EC drift. Check lower-leaf age — senescence vs true deficiency.',
    'moderate',
    10
),
(
    'tip_burn',
    'Tip burn / leaf edge necrosis',
    ARRAY['lettuce', 'basil', 'cannabis', 'tomato', 'strawberry'],
    ARRAY['leafy', 'herbs', 'fruiting'],
    E'**Observation:** Brown or crispy tips and margins, often starting at leaf apex.\n\n**Hypotheses:**\n- Nutrient burn (EC too high vs stage target)\n- Calcium transport issues ( inconsistent irrigation, low transpiration)\n- Salt buildup in media / inadequate flush\n\n**Checks:** Measure in-solution or runoff EC vs profile target. Review irrigation frequency and dryback. Never diagnose from appearance alone — adjust feed only after EC/pH evidence.',
    'moderate',
    20
),
(
    'yellow_lower_leaves',
    'Yellowing lower leaves (mobile deficiency pattern)',
    ARRAY['tomato', 'cannabis', 'pepper', 'cucumber', 'lettuce'],
    ARRAY['vegetables', 'fruiting', 'leafy'],
    E'**Observation:** Oldest leaves yellow from the bottom up.\n\n**Hypotheses:**\n- Nitrogen deficiency or normal senescence under heavy fruit load\n- Overwatering reducing root uptake\n- pH drift locking out nitrogen\n\n**Checks:** Feed EC in range? pH stable? If EC is on target and only lowest leaves affected, may be age — compare to symptom guide and stage.',
    'low',
    30
),
(
    'purple_stems',
    'Purple stems / petioles',
    ARRAY['cannabis', 'tomato', 'pepper', 'kale'],
    ARRAY['fruiting', 'vegetables', 'leafy'],
    E'**Observation:** Stems or leaf undersides develop purple/red anthocyanin.\n\n**Hypotheses:**\n- Phosphorus deficiency (especially cool temps or early veg)\n- Genetics / light stress (some cultivars normal)\n- Root zone too cold\n\n**Checks:** Root zone temp, feed EC vs early/late veg targets. Do not treat with bloom boost without confirming P need and EC headroom.',
    'low',
    40
),
(
    'wilting',
    'Wilting / drooping leaves',
    ARRAY[]::text[],
    ARRAY['vegetables', 'herbs', 'fruiting', 'leafy'],
    E'**Observation:** Leaves lose turgor — wilted despite wet or dry media.\n\n**Hypotheses:**\n- Underwatering / dryback too deep\n- Overwatering / root rot (wet media + wilt)\n- Heat stress / VPD too high\n- Stem damage or vascular blockage\n\n**Checks:** Substrate moisture, runoff presence, zone RH/VPD vs targets, root health. Compare sensor readings to crop profile comfort band.',
    'high',
    50
),
(
    'leaf_spotting',
    'Leaf spots (brown/black lesions)',
    ARRAY['tomato', 'cucumber', 'basil', 'strawberry'],
    ARRAY['fruiting', 'vegetables', 'herbs'],
    E'**Observation:** Discrete brown, tan, or black spots on leaves — may have yellow halos.\n\n**Hypotheses:**\n- Fungal leaf spot (humidity, poor airflow)\n- Bacterial spot (splash, warm wet foliage)\n- Nutrient burn mimicking spots (check EC first)\n\n**Checks:** EC/pH on target? Leaf wetness duration? Isolate affected leaves; improve airflow. Guardian offers checklist — not a lab diagnosis.',
    'moderate',
    60
),
(
    'chewing_damage',
    'Chewing damage on leaves',
    ARRAY[]::text[],
    ARRAY['vegetables', 'herbs', 'leafy'],
    E'**Observation:** Irregular holes, notched margins, or skeletonized tissue.\n\n**Hypotheses:**\n- Caterpillars, slugs, beetles, or mammal browse\n- Mechanical damage from handling\n\n**Checks:** Inspect undersides at night for larvae. Sticky traps for monitoring only. Document pattern — new vs old growth. No pesticide product recommendations from Guardian.',
    'moderate',
    70
),
(
    'powdery_white_coating',
    'White powdery coating on leaves',
    ARRAY['cucumber', 'squash', 'cannabis', 'basil'],
    ARRAY['fruiting', 'vegetables', 'herbs'],
    E'**Observation:** White dusty patches on upper leaf surfaces.\n\n**Hypotheses:**\n- Powdery mildew (often dry foliage + high humidity swings)\n- Mineral residue from hard water or foliar feeds (wipes off)\n\n**Checks:** Wipe test — mildew does not wipe clean. Improve airflow, reduce leaf wetness. Confirm EC/pH still on target — stressed plants invite disease.',
    'moderate',
    80
),
(
    'leaf_curl',
    'Leaf curl / cupping',
    ARRAY['tomato', 'pepper', 'cannabis', 'cucumber'],
    ARRAY['fruiting', 'vegetables'],
    E'**Observation:** Leaves roll upward or downward, cupped or twisted.\n\n**Hypotheses:**\n- Heat / light stress\n- Herbicide drift (outdoor) or residue\n- Calcium / boron issues in fast growth\n- Aphid infestation (check undersides)\n\n**Checks:** Zone temp/RH/VPD vs profile. Inspect new growth. Feed EC and pH vs stage row.',
    'moderate',
    90
),
(
    'blossom_end_rot',
    'Blossom-end rot on fruit',
    ARRAY['tomato', 'pepper', 'eggplant', 'squash'],
    ARRAY['fruiting'],
    E'**Observation:** Dark sunken patch on blossom end of fruit.\n\n**Hypotheses:**\n- Calcium transport failure (often irrigation inconsistency, not only low Ca in feed)\n- Extreme VPD swings limiting uptake\n\n**Checks:** Irrigation consistency, runoff pattern, EC stable through fruit set. Compare to tomato/pepper fruiting EC targets — do not chase Ca without fixing water rhythm.',
    'moderate',
    100
)
ON CONFLICT (symptom_key) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    crop_keys = EXCLUDED.crop_keys,
    categories = EXCLUDED.categories,
    body_md = EXCLUDED.body_md,
    severity_hint = EXCLUDED.severity_hint,
    sort_order = EXCLUDED.sort_order,
    updated_at = NOW();

INSERT INTO gr33ncrops.agronomy_field_guides (
    slug, title, crop_key, guide_kind, domain, safety_tier, body_md, sort_order
) VALUES (
    'crop-deficiency-patterns',
    'Crop deficiency & pest symptom patterns',
    NULL,
    'symptom_catalog',
    'agronomy',
    'safe',
    E'# Crop deficiency & pest symptom patterns\n\nGuardian uses **structured symptom rows** (`lookup_crop_symptoms`) plus **live EC/pH/VPD** from `lookup_crop_targets` — never diagnose from narrative alone.\n\n## How to use\n\n1. Name the crop (`crop_key`) and what you see (yellow, tip burn, spots).\n2. Compare feed and runoff EC/pH to the stage profile.\n3. Treat read-tool output as **hypotheses + checks**, not a lab diagnosis.\n\n## Category notes\n\n- **Fruiting** (tomato, pepper, cannabis flower): watch EC drift, tip burn, interveinal yellow on mid-canopy.\n- **Leafy** (lettuce, basil, kale): tip burn often tracks EC vs gentle targets.\n- **Pest symptoms:** chewing, spotting, powdery patches — confirm with inspection; Guardian does not identify pests from photos as fact.\n',
    900
)
ON CONFLICT (slug) DO UPDATE SET
    title = EXCLUDED.title,
    guide_kind = EXCLUDED.guide_kind,
    body_md = EXCLUDED.body_md,
    updated_at = NOW();
