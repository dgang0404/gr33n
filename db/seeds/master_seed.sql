-- =============================================================================
-- gr33n Master Seed File  v1.009
-- Phase 48: idempotent sensors/automation_rules; demo_showcase profile tag on farm 1.
-- + Demo input_batches (inventory), flower reservoir + fertigation program,
--   Phase 29 WS7 — three unread Guardian demo alerts (farm 1).
--   mixing_events + components, crop_cycles, protocol tasks (18/6 veg vs 12/12 flower).
-- v1.004: schedules table has no metadata column — notes moved to description
-- v1.008 (Phase 124): fixed stale crop_cycles.strain_or_variety → batch_label
-- v1.009: bundled demo farm Today layout background (data/files/farm-1/layout-background/)
--   (renamed in Phase 93, seed was never updated); added dev-hygiene purge of
--   smoke-test artifacts on farm 1; added 4 more zones (Propagation Room,
--   Herb & Greens Room, Outdoor Pepper Bed, Outdoor Berry Patch), a `plants`
--   catalog row per crop, and 11 crop_cycles across active + harvested states
--   so the demo farm reads like a real small farm, not phase-test debris.
-- =============================================================================
-- Run once after schema:
--   psql -d gr33n -f db/seeds/master_seed.sql
-- =============================================================================

BEGIN;

-- ===========================================================================
-- SECTION 0: BOOTSTRAP
-- ===========================================================================

INSERT INTO auth.users (id, email)
VALUES ('00000000-0000-0000-0000-000000000001', 'dev@gr33n.local')
ON CONFLICT (id) DO NOTHING;

INSERT INTO gr33ncore.profiles (user_id, full_name, email, role)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'Dev Farmer', 'dev@gr33n.local', 'farm_manager'
) ON CONFLICT (user_id) DO NOTHING;

-- bcrypt hash for password: devpassword (for local / smoke JWT login)
UPDATE auth.users
SET password_hash = convert_to('$2a$10$OTVuyp0CSrHuDkZ2F8ZIHe2rF56HUYyDp6haKYuwKaDBNQMPIfJe.', 'UTF8')
WHERE id = '00000000-0000-0000-0000-000000000001';

INSERT INTO gr33ncore.farms (
    id, name, description, owner_user_id,
    timezone, currency, scale_tier, operational_status
) VALUES (
    1, 'gr33n Demo Farm',
    'Demo farm pre-loaded with JADAM inputs, light schedules, and watering programs.',
    '00000000-0000-0000-0000-000000000001',
    'America/New_York', 'USD', 'small', 'active'
) ON CONFLICT (id) DO NOTHING;

INSERT INTO gr33ncore.farm_memberships (farm_id, user_id, role_in_farm, permissions, joined_at)
VALUES (
    1,
    '00000000-0000-0000-0000-000000000001',
    'owner',
    '{}'::jsonb,
    NOW()
) ON CONFLICT (farm_id, user_id) DO NOTHING;

-- Phase 48 — demo farm 1 uses full showcase profile (small_indoor via dev-reset-farm.sh).
UPDATE gr33ncore.farms
SET meta_data = COALESCE(meta_data, '{}'::jsonb) || '{"dev_seed_profile":"demo_showcase"}'::jsonb
WHERE id = 1;

-- Phase 124 WS0 — dev hygiene: the Go integration test suite (`cmd/api` smoke
-- tests) writes real plants/crop-cycles into farm 1 without cleanup when run
-- against a local DATABASE_URL. Purge that leftover test debris every time
-- the seed runs so the demo farm doesn't accumulate "Phase98", "smoke_cycle_…",
-- "typo A/B", etc. Root cause (smoke tests not cleaning up farm 1) tracked in
-- docs/plans/phase_124_realistic_seed_data.plan.md.
DELETE FROM gr33nfertigation.crop_cycles
WHERE farm_id = 1
  AND (
    name ~* '^(smoke_|phase[0-9]+|typo )'
    OR batch_label ~* '^(smoke_|phase[0-9]+|typo )'
  );
DELETE FROM gr33ncrops.plants
WHERE farm_id = 1
  AND (
    display_name ~* '^(typo |smoke_|phase[0-9]+)'
    OR variety_or_cultivar ~* '^(typo |smoke_|phase[0-9]+)'
    OR crop_key IS NULL
  );

INSERT INTO gr33ncore.zones (farm_id, name, description, zone_type)
SELECT 1, v.name, v.description, v.zone_type
FROM (VALUES
    ('Veg Room',        'Vegetative growth stage. 18/6 light, JLF+JMS feeding.',           'indoor'),
    ('Flower Room',     'Flowering and fruiting stage. 12/12 light, FFJ+WCA program.',     'indoor'),
    ('Outdoor Garden',  'Outdoor raised beds and garden rows. Natural light. JADAM soil program.', 'outdoor'),
    ('Propagation Room', 'Clones and seedlings under T5s until they graduate to Veg Room or outside.', 'indoor'),
    ('Herb & Greens Room', 'Perpetual indoor herbs and leafy greens under LED, cut-and-come-again.', 'indoor'),
    ('Outdoor Pepper Bed', 'Raised bed, full sun. Peppers direct-planted after last frost.', 'outdoor'),
    ('Outdoor Berry Patch', 'Perennial strawberry patch, drip-irrigated.', 'outdoor')
) AS v(name, description, zone_type)
WHERE NOT EXISTS (
    SELECT 1 FROM gr33ncore.zones z
    WHERE z.farm_id = 1 AND z.name = v.name AND z.deleted_at IS NULL
);

-- Phase 171 — demo farm canvas: persist spatial zone layouts (normalized 0–1).
UPDATE gr33ncore.zones z
SET meta_data = COALESCE(z.meta_data, '{}'::jsonb) || jsonb_build_object('layout', v.layout::jsonb)
FROM (VALUES
    ('Veg Room',            '{"x":0.04,"y":0.06,"w":0.20,"h":0.18}'),
    ('Flower Room',         '{"x":0.28,"y":0.06,"w":0.20,"h":0.18}'),
    ('Propagation Room',    '{"x":0.52,"y":0.06,"w":0.20,"h":0.18}'),
    ('Herb & Greens Room',  '{"x":0.76,"y":0.06,"w":0.20,"h":0.18}'),
    ('Outdoor Garden',      '{"x":0.10,"y":0.32,"w":0.24,"h":0.20}'),
    ('Outdoor Pepper Bed',  '{"x":0.38,"y":0.34,"w":0.20,"h":0.18}'),
    ('Outdoor Berry Patch', '{"x":0.62,"y":0.34,"w":0.20,"h":0.18}')
) AS v(zone_name, layout)
WHERE z.farm_id = 1 AND z.deleted_at IS NULL AND z.name = v.zone_name;

-- ===========================================================================
-- SECTION 1: JADAM INPUT DEFINITIONS
-- ===========================================================================

INSERT INTO gr33nnaturalfarming.input_definitions
    (farm_id, name, category, description, typical_ingredients,
     preparation_summary, storage_guidelines, safety_precautions, reference_source)
VALUES

(1, 'JMS (JADAM Microbial Solution)', 'microbial_inoculant',
 'Foundation of JADAM. Diverse microbial community from forest floor leaf mold. '
 'Applied to soil and foliage to build beneficial populations and suppress pathogens.',
 'Leaf mold humus (forest floor), boiled potato water (cooled), sea salt (pinch)',
 'Mix 1 cup leaf mold into 20L cooled potato water with a pinch of sea salt. '
 'Cover loosely, ferment 3-7 days at 20-30C until bubbling subsides.',
 'Use within 1 week once active. Cool, shaded, loosely covered.',
 'Chlorinated water kills microbes — use filtered or rain water only.',
 'JADAM Organic Farming, Youngsang Cho, 2016'),

(1, 'LAB (Lactic Acid Bacteria Serum)', 'microbial_inoculant',
 'Concentrated lactic acid bacteria from rice wash and milk culture. '
 'Suppresses harmful soil microorganisms and improves soil structure.',
 'Rice wash water (first rinse), fresh whole milk (non-UHT preferred)',
 'Ferment rice wash 3-5 days until soured. Mix 1 part into 10 parts milk. '
 'Wait 5-7 days. Extract golden serum from bottom layer.',
 'Mix with equal part raw sugar to preserve. Refrigerated 6-12 months.',
 'Use golden layer only. Discard curds and white top.',
 'JADAM Organic Farming, Youngsang Cho, 2016'),

(1, 'FPJ (Fermented Plant Juice)', 'fermented_plant_juice',
 'Made from rapidly growing plant tips (comfrey, nettle, mugwort, bamboo). '
 'Rich in plant growth hormones, enzymes, and amino acids. Promotes vigorous veg growth.',
 'Fresh growing tips of fast-growing plants, brown sugar (1:1 by weight)',
 'Layer equal weights of chopped plant material and sugar. Seal with breathable cloth. '
 'Ferment 3-7 days. Strain and bottle.',
 'Refrigerate after straining. Keeps 6-12 months.',
 'Keep sugar ratio accurate. Do not use moldy material.',
 'JADAM Organic Farming, Youngsang Cho, 2016'),

(1, 'FFJ (Fermented Fruit Juice)', 'fermented_plant_juice',
 'Made from ripe or overripe sweet fruits. High in sugars, enzymes, and potassium. '
 'Promotes flowering and fruiting. Apply at transition to reproductive stage.',
 'Ripe/overripe fruits (banana peels work well), brown sugar (1:1 by weight)',
 'Chop fruit, mix 1:1 with sugar. Ferment loosely covered 7 days. Strain.',
 'Refrigerate after straining. Use within 6 months.',
 'Use during flowering/fruiting only.',
 'JADAM Organic Farming, Youngsang Cho, 2016'),

(1, 'BRV (Brown Rice Vinegar)', 'fermented_plant_juice',
 'Fermented brown rice vinegar 4-8% acidity. Strengthens cell walls, '
 'improves calcium uptake, deters fungal issues.',
 'Organic brown rice vinegar (unpasteurized preferred)',
 'Purchase unpasteurized organic BRV. No preparation needed.',
 'Store sealed at room temperature indefinitely.',
 'Dilute properly — undiluted burns foliage.',
 'JADAM Organic Farming, Youngsang Cho, 2016'),

(1, 'OHN (Oriental Herbal Nutrient)', 'oriental_herbal_nutrient',
 'Extracted from aromatic herbs and roots (garlic, ginger, angelica, cinnamon). '
 'Powerful immune booster and pest deterrent. Used in very small quantities.',
 'Garlic, ginger, Angelica root, cinnamon bark, brown sugar, alcohol ~25% ABV',
 'Chop herbs, layer with sugar 1:1, ferment 7 days. Add equal alcohol. '
 'Ferment 7 more days. Strain. Combine extracts.',
 'Keeps 1-2 years sealed.',
 'Extremely potent — always dilute 1:1000 minimum. Avoid inhaling.',
 'JADAM Organic Farming, Youngsang Cho, 2016'),

(1, 'JHS (JADAM Herbal Solution)', 'oriental_herbal_nutrient',
 'Water-based extract of aromatic and pest-repellent herbs. Broader spectrum '
 'pest deterrent and foliar immune support. Mixed with JWA for natural pesticide sprays.',
 'Wormwood, artemisia, garlic chives, hot pepper, neem leaves, non-chlorinated water',
 'Simmer or cold-extract herbs in water 1-3 hours. Strain finely. Use fresh.',
 'Use within 2 weeks refrigerated. Strain very fine before loading sprayer.',
 'Do not apply on blooms — deters pollinators. Apply morning only.',
 'JADAM Organic Farming, Youngsang Cho, 2016'),

(1, 'WCA (Water-Soluble Calcium)', 'water_soluble_nutrient',
 'Calcium from eggshells dissolved in brown rice vinegar. Strengthens cell walls, '
 'improves fruit quality, prevents blossom end rot.',
 'Eggshells (or oyster shells), brown rice vinegar (4-8%)',
 'Roast eggshells until lightly brown. Cool. Cover with BRV 1:10. '
 'Fizzing will occur. Leave 7 days uncovered. Strain.',
 'Store in open-top container (gases form). Use within 30 days.',
 'Container must be breathable. Roast shells well.',
 'JADAM Organic Farming, Youngsang Cho, 2016'),

(1, 'WCS (Water-Soluble Calcium Phosphate)', 'water_soluble_nutrient',
 'Phosphorus and calcium from charred animal bones in brown rice vinegar. '
 'Promotes root development, flowering, and ripening.',
 'Beef or pork bones (charred to white ash), brown rice vinegar',
 'Char bones until white ash. Cool. Dissolve in BRV 1:10 for 7 days. Strain.',
 'Store in breathable container. Use within 30 days.',
 'Char bones fully to white — partial char gives inconsistent results.',
 'JADAM Organic Farming, Youngsang Cho, 2016'),

(1, 'JWA (JADAM Wetting Agent)', 'other_extract',
 'Homemade soap from plant oils and wood ash lye. Organic surfactant and '
 'contact insecticide for soft-bodied insects (aphids, mites, whitefly).',
 'Plant oil (soybean, canola, or coconut), wood ash lye water',
 'Boil wood ash in water, filter lye. Mix with oil 1:1. Boil until soap forms.',
 'Keeps indefinitely dry. Dilute 1:500-1:1000 for spraying.',
 'Lye is caustic — wear gloves when making. Do not apply in direct sun.',
 'JADAM Organic Farming, Youngsang Cho, 2016'),

(1, 'JS (JADAM Sulfur)', 'other_extract',
 'Sulfur solution for powdery mildew, rust, and spider mites. '
 'Core JADAM disease control input. Broad-spectrum fungicide and miticide.',
 'Wettable sulfur powder, water, JWA (as emulsifier)',
 'Dissolve sulfur powder at 0.5% in warm water with JWA. Mix fresh each use.',
 'Store dry sulfur powder sealed indefinitely. Mix fresh per application.',
 'Do not apply above 32C — sulfur burn risk. Wear mask and gloves.',
 'JADAM Organic Farming, Youngsang Cho, 2016'),

(1, 'JLF General (Weed and Grass)', 'other_ferment',
 'JADAM Liquid Fertilizer from locally available weeds and grasses. '
 'Returns native minerals to soil. Free from farm waste. '
 'Dilution 1:20 general, 1:30 seedlings. Much stronger than other JADAM inputs.',
 'Fresh untreated weeds and grass clippings, leaf mold (handful), non-chlorinated water',
 'Fill container 2/3 with chopped weeds. Add leaf mold as microbial starter. '
 'Fill to top with water. Seal. Ferment 7-14 days. Stir every few days. '
 'Ready when earthy smell. Strain through cloth before use.',
 'Use strained within 30 days. Sealed undiluted keeps 3 months.',
 'Non-chlorinated water only. No herbicide-treated material.',
 'JADAM Organic Farming, Youngsang Cho, 2016'),

(1, 'JLF Crop-Specific (Crop Residue)', 'other_ferment',
 'JLF from the same crop''s own residue — most targeted fertilizer possible. '
 'Tomato residue for tomatoes, corn stalks for corn.',
 'Crop residue (stems, leaves, roots, not fruit or seed), leaf mold, non-chlorinated water',
 'Chop crop residue small. Fill container 2/3. Add leaf mold. Fill with water. '
 'Seal. Ferment 10-14 days. Strain.',
 'Use within same season. Label with crop type and date.',
 'Healthy residue only — do not use diseased plant material.',
 'JADAM Organic Farming, Youngsang Cho, 2016'),

(1, 'JLF Spring (Nettle and Comfrey)', 'other_ferment',
 'High-nitrogen JLF from nitrogen-fixing plants. Best for spring vegetative '
 'growth push. Nettle and comfrey mine deep minerals.',
 'Fresh stinging nettle tops, comfrey leaves (or either alone), leaf mold, water',
 'Harvest tops wearing gloves. Fill container 2/3. Add leaf mold, fill with water. '
 'Ferment 7-10 days. Strain.',
 'Use within 2 weeks of straining.',
 'Wear gloves harvesting nettle. Very high N — do not over-apply to fruiting plants.',
 'JADAM Organic Farming, Youngsang Cho, 2016'),

(1, 'Compost Tea Actively Aerated', 'compost_tea_extract',
 'Brewed extract of finished compost, aerated 24-48h to multiply aerobic microbes. '
 'Builds soil food web, suppresses disease. Complements JMS.',
 'Finished compost, unsulfured molasses, kelp meal, de-chlorinated water',
 'Add compost in mesh bag to bucket with air stone. Add 1 tbsp molasses per 4L. '
 'Brew 24-48 hours. Use within 4 hours of finishing.',
 'Must use within 4 hours — microbes die without oxygen.',
 'Never store brewed tea.',
 'Elaine Ingham, Soil Biology Primer')

ON CONFLICT DO NOTHING;

-- ===========================================================================
-- SECTION 2: APPLICATION RECIPES
-- ===========================================================================

INSERT INTO gr33nnaturalfarming.application_recipes
    (farm_id, name, description, target_application_type,
     dilution_ratio, instructions, frequency_guidelines,
     target_crop_categories, target_growth_stages)
VALUES

(1, 'JMS Soil Drench',
 'Base soil microbe inoculant. Foundation of all JADAM programs.',
 'soil_drench', '1:500 (JMS:water)',
 'Dilute 1:500. Apply 2-4L per sqm of root zone. Morning or evening.',
 'Every 2 weeks growing season. Monthly dormant.',
 ARRAY['All crops'], ARRAY['All stages']),

(1, 'JLF General Soil Drench',
 'Primary fertility input. Main fertilizer not a supplement.',
 'soil_drench', '1:20 (JLF:water)',
 'Strain JLF through cloth. Dilute 1:20. Apply 2-4L per sqm to root zone.',
 'Every 1-2 weeks active growth.',
 ARRAY['All crops'], ARRAY['All stages']),

(1, 'JLF Seedling Drench',
 'Gentler dilution safe for young seedlings and fresh transplants.',
 'soil_drench', '1:30 (JLF:water)',
 'Dilute 1:30. Apply 0.5L per tray or 1L per transplant hole.',
 'Weekly from germination through first 2 weeks after transplant.',
 ARRAY['All crops'], ARRAY['Seedling', 'Transplant']),

(1, 'JLF and JMS Combined Drench',
 'Nutrients and microbes in one pass. Core weekly application.',
 'soil_drench', 'JLF 1:20 + JMS 1:500 in same water',
 'Fill tank. Add JLF 1:20, then JMS 1:500. Apply same day.',
 'Weekly during peak growing season.',
 ARRAY['All crops'], ARRAY['All stages']),

(1, 'LAB Soil Conditioner',
 'Suppresses harmful soil pathogens, speeds organic matter breakdown.',
 'soil_drench', '1:1000 (LAB:water)',
 'Dilute LAB 1:1000. Apply evenly to soil surface. Water in lightly after.',
 'Every 2-4 weeks. Especially valuable before transplanting.',
 ARRAY['All crops'], ARRAY['Pre-plant', 'Transplant', 'All stages']),

(1, 'OHN Pest and Immunity Drench',
 'Stimulates plant immune response and deters insects.',
 'soil_drench', '1:1000 (OHN:water)',
 'Dilute OHN strictly 1:1000. Apply 1-2L per plant root zone.',
 'Every 2-4 weeks preventative. Weekly during pest or disease pressure.',
 ARRAY['All crops'], ARRAY['All stages']),

(1, 'JMS Foliar Spray',
 'Establishes beneficial microbes on leaf surfaces. Suppresses airborne pathogens.',
 'foliar_spray', '1:500 (JMS:water)',
 'Dilute 1:500. Spray upper and lower leaf surfaces to runoff. Early morning.',
 'Every 1-2 weeks. More often during high humidity.',
 ARRAY['All crops'], ARRAY['Vegetative', 'Early flowering']),

(1, 'FPJ Vegetative Foliar',
 'Promotes rapid vegetative growth. Stop at flowering transition.',
 'foliar_spray', '1:500 to 1:1000 (FPJ:water)',
 'Dilute 1:500 normal conditions, 1:1000 in hot weather. Add JWA 1:1000.',
 'Every 7-14 days during vegetative stage.',
 ARRAY['Leafy greens', 'Brassicas', 'Cucurbits', 'Tomatoes'],
 ARRAY['Seedling', 'Vegetative']),

(1, 'FFJ and WCA Flowering Boost',
 'Supports flowering transition and early fruit set.',
 'foliar_spray', 'FFJ 1:500 + WCA 1:1000 combined',
 'Mix FFJ 1:500 and WCA 1:1000 in same tank. Add JWA 1:1000. Morning.',
 'Weekly from first flower buds through early fruit set.',
 ARRAY['Tomatoes', 'Peppers', 'Cucumbers', 'Squash', 'Fruit trees'],
 ARRAY['Flowering', 'Early fruit']),

(1, 'BRV and WCA Cell Strengthener',
 'Hardens cell walls. Apply before rain, cold snaps, or disease pressure.',
 'foliar_spray', 'BRV 1:800 + WCA 1:1000',
 'Mix BRV 1:800 and WCA 1:1000. Do not exceed BRV concentration — burn risk.',
 'Every 2 weeks during fruiting or before stress events.',
 ARRAY['All crops'], ARRAY['Vegetative', 'Fruiting']),

(1, 'JHS and JWA Natural Pesticide',
 'Broad-spectrum organic pest deterrent. Effective against chewing and sucking insects.',
 'foliar_spray', 'JHS 1:50 + JWA 1:500',
 'Strain JHS very finely. Mix JHS 1:50 and JWA 1:500. '
 'Apply thorough coverage especially leaf undersides. Morning or evening.',
 'Weekly preventative. Every 3-5 days for active pest pressure.',
 ARRAY['All crops'], ARRAY['Any stage']),

(1, 'JS Fungicide Spray',
 'Controls powdery mildew, rust, and spider mites.',
 'foliar_spray', '0.5% JS + JWA 1:500',
 'Dissolve wettable sulfur at 0.5% in water with JWA. Mix fresh. '
 'Apply thorough coverage. Do NOT apply above 32C.',
 'At first sign of fungal disease. Repeat every 5-7 days.',
 ARRAY['All crops'], ARRAY['Any stage']),

(1, 'JLF Foliar Feed',
 'Fast nutrient uptake during plant stress or deficiency.',
 'foliar_spray', '1:30 to 1:50 (JLF:water)',
 'Strain JLF very finely. Dilute 1:30 min (1:50 hot weather). Add JWA 1:1000.',
 'Weekly during stress. Not a substitute for soil application.',
 ARRAY['All crops'], ARRAY['Any stage under stress']),

(1, 'JWA Insecticide Spray',
 'Contact insecticide for aphids, spider mites, whitefly, soft-bodied insects.',
 'foliar_spray', '1:500 (JWA:water)',
 'Dilute 1:500. Cover leaf surfaces including undersides. Morning or evening.',
 'Every 3-5 days for active infestations.',
 ARRAY['All crops'], ARRAY['Any stage'])

ON CONFLICT DO NOTHING;

-- Recipe components
DO $$
DECLARE
    v_farm   BIGINT := 1;
    u_frac   BIGINT;
    i_jms    BIGINT; i_lab  BIGINT; i_fpj  BIGINT; i_ffj  BIGINT;
    i_brv    BIGINT; i_ohn  BIGINT; i_jhs  BIGINT; i_js   BIGINT;
    i_jlf_g  BIGINT; i_wca  BIGINT; i_jwa  BIGINT;
    r_jms_s  BIGINT; r_jlf_s  BIGINT; r_jlf_sd BIGINT; r_combo  BIGINT;
    r_lab    BIGINT; r_ohn    BIGINT; r_jms_f  BIGINT; r_fpj_f  BIGINT;
    r_ffj_f  BIGINT; r_brv_f  BIGINT; r_jhs_f  BIGINT; r_js_f   BIGINT;
    r_jlf_f  BIGINT; r_jwa_f  BIGINT;
BEGIN
    SELECT id INTO u_frac  FROM gr33ncore.units                       WHERE name = 'decimal_fraction'               LIMIT 1;
    SELECT id INTO i_jms   FROM gr33nnaturalfarming.input_definitions WHERE farm_id=v_farm AND name LIKE 'JMS%'     LIMIT 1;
    SELECT id INTO i_lab   FROM gr33nnaturalfarming.input_definitions WHERE farm_id=v_farm AND name LIKE 'LAB%'     LIMIT 1;
    SELECT id INTO i_fpj   FROM gr33nnaturalfarming.input_definitions WHERE farm_id=v_farm AND name LIKE 'FPJ%'     LIMIT 1;
    SELECT id INTO i_ffj   FROM gr33nnaturalfarming.input_definitions WHERE farm_id=v_farm AND name LIKE 'FFJ%'     LIMIT 1;
    SELECT id INTO i_brv   FROM gr33nnaturalfarming.input_definitions WHERE farm_id=v_farm AND name LIKE 'BRV%'     LIMIT 1;
    SELECT id INTO i_ohn   FROM gr33nnaturalfarming.input_definitions WHERE farm_id=v_farm AND name LIKE 'OHN%'     LIMIT 1;
    SELECT id INTO i_jhs   FROM gr33nnaturalfarming.input_definitions WHERE farm_id=v_farm AND name LIKE 'JHS%'     LIMIT 1;
    SELECT id INTO i_js    FROM gr33nnaturalfarming.input_definitions WHERE farm_id=v_farm AND name LIKE 'JS (%'    LIMIT 1;
    SELECT id INTO i_jlf_g FROM gr33nnaturalfarming.input_definitions WHERE farm_id=v_farm AND name LIKE 'JLF Gen%' LIMIT 1;
    SELECT id INTO i_wca   FROM gr33nnaturalfarming.input_definitions WHERE farm_id=v_farm AND name LIKE 'WCA%'     LIMIT 1;
    SELECT id INTO i_jwa   FROM gr33nnaturalfarming.input_definitions WHERE farm_id=v_farm AND name LIKE 'JWA%'     LIMIT 1;

    SELECT id INTO r_jms_s  FROM gr33nnaturalfarming.application_recipes WHERE farm_id=v_farm AND name='JMS Soil Drench'          LIMIT 1;
    SELECT id INTO r_jlf_s  FROM gr33nnaturalfarming.application_recipes WHERE farm_id=v_farm AND name='JLF General Soil Drench'  LIMIT 1;
    SELECT id INTO r_jlf_sd FROM gr33nnaturalfarming.application_recipes WHERE farm_id=v_farm AND name='JLF Seedling Drench'      LIMIT 1;
    SELECT id INTO r_combo  FROM gr33nnaturalfarming.application_recipes WHERE farm_id=v_farm AND name LIKE 'JLF and JMS%'        LIMIT 1;
    SELECT id INTO r_lab    FROM gr33nnaturalfarming.application_recipes WHERE farm_id=v_farm AND name='LAB Soil Conditioner'      LIMIT 1;
    SELECT id INTO r_ohn    FROM gr33nnaturalfarming.application_recipes WHERE farm_id=v_farm AND name LIKE 'OHN%'                LIMIT 1;
    SELECT id INTO r_jms_f  FROM gr33nnaturalfarming.application_recipes WHERE farm_id=v_farm AND name='JMS Foliar Spray'         LIMIT 1;
    SELECT id INTO r_fpj_f  FROM gr33nnaturalfarming.application_recipes WHERE farm_id=v_farm AND name LIKE 'FPJ%'                LIMIT 1;
    SELECT id INTO r_ffj_f  FROM gr33nnaturalfarming.application_recipes WHERE farm_id=v_farm AND name LIKE 'FFJ%'                LIMIT 1;
    SELECT id INTO r_brv_f  FROM gr33nnaturalfarming.application_recipes WHERE farm_id=v_farm AND name LIKE 'BRV%'                LIMIT 1;
    SELECT id INTO r_jhs_f  FROM gr33nnaturalfarming.application_recipes WHERE farm_id=v_farm AND name LIKE 'JHS%'                LIMIT 1;
    SELECT id INTO r_js_f   FROM gr33nnaturalfarming.application_recipes WHERE farm_id=v_farm AND name LIKE 'JS Fungicide%'       LIMIT 1;
    SELECT id INTO r_jlf_f  FROM gr33nnaturalfarming.application_recipes WHERE farm_id=v_farm AND name LIKE 'JLF Foliar%'         LIMIT 1;
    SELECT id INTO r_jwa_f  FROM gr33nnaturalfarming.application_recipes WHERE farm_id=v_farm AND name LIKE 'JWA Insecticide%'    LIMIT 1;

    INSERT INTO gr33nnaturalfarming.recipe_input_components
        (application_recipe_id, input_definition_id, part_value, part_unit_id, notes)
    VALUES
        (r_jms_s,  i_jms,   1.0,   u_frac, '1 part JMS to 500 parts water'),
        (r_jlf_s,  i_jlf_g, 1.0,   u_frac, '1 part JLF to 20 parts water'),
        (r_jlf_sd, i_jlf_g, 1.0,   u_frac, '1 part JLF to 30 parts water'),
        (r_combo,  i_jlf_g, 1.0,   u_frac, 'JLF at 1:20'),
        (r_combo,  i_jms,   0.025, u_frac, 'JMS at 1:500 relative to 1:20 base'),
        (r_lab,    i_lab,   1.0,   u_frac, '1 part LAB to 1000 parts water'),
        (r_ohn,    i_ohn,   1.0,   u_frac, '1 part OHN to 1000 — never exceed'),
        (r_jms_f,  i_jms,   1.0,   u_frac, '1 part JMS to 500 parts water'),
        (r_fpj_f,  i_fpj,   1.0,   u_frac, '1 part FPJ to 500-1000 parts water'),
        (r_ffj_f,  i_ffj,   1.0,   u_frac, 'FFJ at 1:500'),
        (r_ffj_f,  i_wca,   0.5,   u_frac, 'WCA at 1:1000 relative'),
        (r_brv_f,  i_brv,   1.0,   u_frac, 'BRV at 1:800'),
        (r_brv_f,  i_wca,   0.8,   u_frac, 'WCA at 1:1000 relative'),
        (r_jhs_f,  i_jhs,   1.0,   u_frac, 'JHS at 1:50'),
        (r_jhs_f,  i_jwa,   0.1,   u_frac, 'JWA at 1:500 surfactant'),
        (r_js_f,   i_js,    1.0,   u_frac, '0.5% sulfur in water'),
        (r_js_f,   i_jwa,   0.1,   u_frac, 'JWA as emulsifier'),
        (r_jlf_f,  i_jlf_g, 1.0,   u_frac, 'JLF at 1:30 to 1:50'),
        (r_jlf_f,  i_jwa,   0.033, u_frac, 'JWA 1:1000 surfactant'),
        (r_jwa_f,  i_jwa,   1.0,   u_frac, '1 part JWA to 500 parts water')
    ON CONFLICT DO NOTHING;
END $$;

-- ===========================================================================
-- SECTION 3: LIGHT SCHEDULES
-- ===========================================================================

INSERT INTO gr33ncore.schedules
    (farm_id, name, description, schedule_type, cron_expression, timezone, is_active)
SELECT 1, v.name, v.description, v.schedule_type, v.cron_expression, v.timezone, v.is_active
FROM (VALUES
('Light ON 24/0 Continuous',
 'Lights always on. Seedling propagation, cloning, autoflowering varieties.',
 'lighting', '0 0 * * *', 'America/New_York', false),

('Light ON 18/6 Veg',
 'Lights on at 06:00. 18 hours on for active vegetative growth.',
 'lighting', '0 6 * * *', 'America/New_York', false),

('Light OFF 18/6 Veg',
 'Lights off at midnight. 6 hours dark.',
 'lighting', '0 0 * * *', 'America/New_York', false),

('Light ON 16/8 Moderate Veg',
 'Lights on at 06:00. 16 hours on — good energy balance vs 18/6.',
 'lighting', '0 6 * * *', 'America/New_York', false),

('Light OFF 16/8 Moderate Veg',
 'Lights off at 22:00. 8 hours dark.',
 'lighting', '0 22 * * *', 'America/New_York', false),

('Light ON 12/12 Flower',
 'Lights on at 06:00. 12 hours on triggers flowering in photoperiod plants.',
 'lighting', '0 6 * * *', 'America/New_York', false),

('Light OFF 12/12 Flower',
 'Lights off at 18:00. 12 hours uninterrupted dark — critical for flowering.',
 'lighting', '0 18 * * *', 'America/New_York', false)
) AS v(name, description, schedule_type, cron_expression, timezone, is_active)
WHERE NOT EXISTS (
    SELECT 1 FROM gr33ncore.schedules s WHERE s.farm_id = 1 AND s.name = v.name
);

-- ===========================================================================
-- SECTION 4: WATERING SCHEDULES
-- Note: schedules table has no metadata column — volume/stage info in description
-- ===========================================================================

INSERT INTO gr33ncore.schedules
    (farm_id, name, description, schedule_type, cron_expression, timezone, is_active)
SELECT 1, v.name, v.description, v.schedule_type, v.cron_expression, v.timezone, v.is_active
FROM (VALUES
('Water Early Veg Every 2 Days',
 'Early veg. ~300mL per plant every 2 days. Allow slight dry-back between '
 'waterings to encourage roots to chase moisture downward. '
 'Zone: Veg Room. Light: 18/6.',
 'irrigation', '0 8 1-31/2 * *', 'America/New_York', false),

('Water Late Veg Daily',
 'Late veg with larger root zone. ~750mL per plant daily. '
 'Increase if wilting occurs before next scheduled watering. '
 'Zone: Veg Room. Light: 18/6 or 16/8.',
 'irrigation', '0 8 * * *', 'America/New_York', true),

('Water Early Flower Daily',
 'First 2 weeks of flowering. ~900mL per plant daily. Slight stress during '
 'stretch week is OK — builds stem density. '
 'Zone: Flower Room. Light: 12/12.',
 'irrigation', '0 8 * * *', 'America/New_York', true),

('Water Peak Flower 2x Daily',
 'Mid to late flowering — maximum demand. ~1.5L per plant twice daily. '
 'Never let medium go fully dry during peak flower. Watch for leaf curl. '
 'Zone: Flower Room. Light: 12/12.',
 'irrigation', '0 8,18 * * *', 'America/New_York', false),

('Water Flush Week 2x Daily',
 'Final 7-14 days before harvest. Plain pH-adjusted water only — no nutrients. '
 '~2L per plant twice daily. 1.5-2x pot volume per session to clear salts. '
 'Zone: Flower Room. Light: 12/12.',
 'irrigation', '0 8,18 * * *', 'America/New_York', false),

('Water Outdoor Garden Daily',
 'Morning irrigation for outdoor garden beds. ~3L per sqm. '
 'Disable during rain periods. Increase in heat waves. '
 'Apply JLF soil drench here. Zone: Outdoor Garden.',
 'irrigation', '0 7 * * *', 'America/New_York', true),

('Water Herbs Gravity Drip Daily',
 'Morning gravity drip for Herb & Greens tent. ~8L from elevated header tank; '
 'drip valve open 3 min — no pump required. Zone: Herb & Greens Room.',
 'irrigation', '0 7 * * *', 'America/New_York', true)
) AS v(name, description, schedule_type, cron_expression, timezone, is_active)
WHERE NOT EXISTS (
    SELECT 1 FROM gr33ncore.schedules s WHERE s.farm_id = 1 AND s.name = v.name
);

-- ===========================================================================
-- SECTION 4C: DEMO DEVICES + ACTUATORS + SCHEDULE ACTIONS
-- ===========================================================================

INSERT INTO gr33ncore.devices
    (farm_id, zone_id, name, device_uid, device_type, status, config)
SELECT
    1,
    (SELECT id FROM gr33ncore.zones WHERE farm_id = 1 AND name = 'Veg Room' AND deleted_at IS NULL ORDER BY id LIMIT 1),
    'Veg Relay Controller',
    'demo-veg-relay-01',
    'relay_controller',
    'online'::gr33ncore.device_status_enum,
    '{"simulation": true}'::jsonb
WHERE NOT EXISTS (
    SELECT 1 FROM gr33ncore.devices WHERE farm_id = 1 AND device_uid = 'demo-veg-relay-01'
);

INSERT INTO gr33ncore.devices
    (farm_id, zone_id, name, device_uid, device_type, status, config)
SELECT
    1,
    (SELECT id FROM gr33ncore.zones WHERE farm_id = 1 AND name = 'Flower Room' AND deleted_at IS NULL ORDER BY id LIMIT 1),
    'Flower Relay Controller',
    'demo-flower-relay-01',
    'relay_controller',
    'online'::gr33ncore.device_status_enum,
    '{"simulation": true}'::jsonb
WHERE NOT EXISTS (
    SELECT 1 FROM gr33ncore.devices WHERE farm_id = 1 AND device_uid = 'demo-flower-relay-01'
);

INSERT INTO gr33ncore.devices
    (farm_id, zone_id, name, device_uid, device_type, status, config)
SELECT
    1,
    (SELECT id FROM gr33ncore.zones WHERE farm_id = 1 AND name = 'Herb & Greens Room' AND deleted_at IS NULL ORDER BY id LIMIT 1),
    'Herb Relay Controller',
    'demo-herb-relay-01',
    'relay_controller',
    'online'::gr33ncore.device_status_enum,
    '{"simulation": true}'::jsonb
WHERE NOT EXISTS (
    SELECT 1 FROM gr33ncore.devices WHERE farm_id = 1 AND device_uid = 'demo-herb-relay-01'
);

INSERT INTO gr33ncore.actuators
    (device_id, farm_id, zone_id, name, actuator_type, hardware_identifier, current_state_text, config)
SELECT
    d.id,
    1,
    d.zone_id,
    'Veg Room Grow Light',
    'light',
    'relay_1',
    'offline',
    '{"channel": 1, "simulation": true}'::jsonb
FROM gr33ncore.devices d
WHERE d.farm_id = 1
  AND d.device_uid = 'demo-veg-relay-01'
  AND NOT EXISTS (
      SELECT 1 FROM gr33ncore.actuators a
      WHERE a.farm_id = 1 AND a.name = 'Veg Room Grow Light' AND a.deleted_at IS NULL
  );

INSERT INTO gr33ncore.actuators
    (device_id, farm_id, zone_id, name, actuator_type, hardware_identifier, current_state_text, config)
SELECT
    d.id,
    1,
    d.zone_id,
    'Flower Room Irrigation Pump',
    'pump',
    'relay_2',
    'offline',
    '{"channel": 2, "simulation": true}'::jsonb
FROM gr33ncore.devices d
WHERE d.farm_id = 1
  AND d.device_uid = 'demo-flower-relay-01'
  AND NOT EXISTS (
      SELECT 1 FROM gr33ncore.actuators a
      WHERE a.farm_id = 1 AND a.name = 'Flower Room Irrigation Pump' AND a.deleted_at IS NULL
  );

INSERT INTO gr33ncore.actuators
    (device_id, farm_id, zone_id, name, actuator_type, hardware_identifier, current_state_text, config)
SELECT
    d.id,
    1,
    d.zone_id,
    'Herb Room Gravity Drip Valve',
    'drip',
    'relay_1',
    'offline',
    '{"channel": 1, "simulation": true, "irrigation_mode": "gravity_drip"}'::jsonb
FROM gr33ncore.devices d
WHERE d.farm_id = 1
  AND d.device_uid = 'demo-herb-relay-01'
  AND NOT EXISTS (
      SELECT 1 FROM gr33ncore.actuators a
      WHERE a.farm_id = 1 AND a.name = 'Herb Room Gravity Drip Valve' AND a.deleted_at IS NULL
  );

INSERT INTO gr33ncore.executable_actions
    (schedule_id, execution_order, action_type, target_actuator_id, action_command, action_parameters)
SELECT
    s.id,
    0,
    'control_actuator'::gr33ncore.executable_action_type_enum,
    a.id,
    'on',
    '{"source":"seed_demo"}'::jsonb
FROM gr33ncore.schedules s
JOIN gr33ncore.actuators a ON a.farm_id = 1 AND a.name = 'Veg Room Grow Light' AND a.deleted_at IS NULL
WHERE s.farm_id = 1
  AND s.name = 'Light ON 18/6 Veg'
  AND NOT EXISTS (
      SELECT 1 FROM gr33ncore.executable_actions ea
      WHERE ea.schedule_id = s.id AND ea.target_actuator_id = a.id AND ea.action_command = 'on'
  );

INSERT INTO gr33ncore.executable_actions
    (schedule_id, execution_order, action_type, target_actuator_id, action_command, action_parameters)
SELECT
    s.id,
    0,
    'control_actuator'::gr33ncore.executable_action_type_enum,
    a.id,
    'off',
    '{"source":"seed_demo"}'::jsonb
FROM gr33ncore.schedules s
JOIN gr33ncore.actuators a ON a.farm_id = 1 AND a.name = 'Veg Room Grow Light' AND a.deleted_at IS NULL
WHERE s.farm_id = 1
  AND s.name = 'Light OFF 18/6 Veg'
  AND NOT EXISTS (
      SELECT 1 FROM gr33ncore.executable_actions ea
      WHERE ea.schedule_id = s.id AND ea.target_actuator_id = a.id AND ea.action_command = 'off'
  );

UPDATE gr33ncore.schedules
SET is_active = TRUE
WHERE farm_id = 1
  AND name IN ('Light ON 18/6 Veg', 'Light OFF 18/6 Veg');

-- ===========================================================================
-- SECTION 3B: LIGHTING PROGRAMS (Phase 35)
-- Wrap the existing 18/6 Veg schedule pair in the new lighting_programs entity.
-- ===========================================================================

DO $$
DECLARE
  v_zone_id      BIGINT;
  v_actuator_id  BIGINT;
  v_sch_on_id    BIGINT;
  v_sch_off_id   BIGINT;
  v_prog_id      BIGINT;
BEGIN
  SELECT id INTO v_zone_id     FROM gr33ncore.zones     WHERE farm_id = 1 AND name = 'Veg Room'           AND deleted_at IS NULL ORDER BY id LIMIT 1;
  SELECT id INTO v_actuator_id FROM gr33ncore.actuators WHERE farm_id = 1 AND name = 'Veg Room Grow Light' AND deleted_at IS NULL ORDER BY id LIMIT 1;
  SELECT id INTO v_sch_on_id   FROM gr33ncore.schedules WHERE farm_id = 1 AND name = 'Light ON 18/6 Veg'  ORDER BY id LIMIT 1;
  SELECT id INTO v_sch_off_id  FROM gr33ncore.schedules WHERE farm_id = 1 AND name = 'Light OFF 18/6 Veg' ORDER BY id LIMIT 1;

  -- Only insert if all references resolved and the program doesn't already exist.
  IF v_zone_id IS NOT NULL AND v_actuator_id IS NOT NULL
     AND v_sch_on_id IS NOT NULL AND v_sch_off_id IS NOT NULL
     AND NOT EXISTS (
         SELECT 1 FROM gr33ncore.lighting_programs
         WHERE farm_id = 1 AND name = 'Veg Room 18/6 Photoperiod'
     ) THEN

    INSERT INTO gr33ncore.lighting_programs
      (farm_id, zone_id, actuator_id, name, description,
       on_hours, off_hours, lights_on_at, timezone,
       schedule_on_id, schedule_off_id,
       is_active, metadata)
    VALUES
      (1, v_zone_id, v_actuator_id,
       'Veg Room 18/6 Photoperiod',
       'Standard vegetative photoperiod — 18h on / 6h off. Lights on at 06:00 America/New_York.',
       18, 6, '06:00', 'America/New_York',
       v_sch_on_id, v_sch_off_id,
       true, '{"preset_key":"veg_18_6","source":"seed_demo"}'::jsonb)
    RETURNING id INTO v_prog_id;

    -- Tag the schedules with the lighting_program_id so they can be found.
    UPDATE gr33ncore.schedules
       SET meta_data = jsonb_set(meta_data, '{lighting_program_id}', v_prog_id::text::jsonb)
     WHERE id IN (v_sch_on_id, v_sch_off_id);
  END IF;
END $$;

-- ===========================================================================
-- SECTION 4B: FERTIGATION BASELINE DATA
-- ===========================================================================

INSERT INTO gr33nfertigation.reservoirs
    (farm_id, zone_id, name, description, capacity_liters, current_volume_liters, status)
SELECT
    1,
    (SELECT id FROM gr33ncore.zones WHERE farm_id = 1 AND name = 'Veg Room' AND deleted_at IS NULL ORDER BY id LIMIT 1),
    'Main Nutrient Reservoir',
    'Primary fertigation reservoir for demo farm programs.',
    500.00,
    320.00,
    'ready'::gr33nfertigation.reservoir_status_enum
ON CONFLICT (farm_id, name) DO NOTHING;

INSERT INTO gr33nfertigation.ec_targets
    (farm_id, zone_id, growth_stage, ec_min_mscm, ec_max_mscm, ph_min, ph_max, rationale)
SELECT
    1,
    z.id,
    gs.stage::gr33nfertigation.growth_stage_enum,
    gs.ec_min,
    gs.ec_max,
    gs.ph_min,
    gs.ph_max,
    'Demo baseline target for fertigation MVP'
FROM gr33ncore.zones z
JOIN (
    VALUES
        ('seedling',     0.5::numeric, 1.2::numeric, 5.8::numeric, 6.6::numeric),
        ('early_veg',    1.0::numeric, 1.8::numeric, 5.8::numeric, 6.6::numeric),
        ('late_veg',     1.4::numeric, 2.2::numeric, 5.8::numeric, 6.6::numeric),
        ('transition',   1.6::numeric, 2.4::numeric, 5.8::numeric, 6.6::numeric),
        ('early_flower', 1.6::numeric, 2.4::numeric, 5.8::numeric, 6.6::numeric),
        ('mid_flower',   1.8::numeric, 2.6::numeric, 5.8::numeric, 6.6::numeric),
        ('late_flower',  1.6::numeric, 2.4::numeric, 5.8::numeric, 6.6::numeric),
        ('flush',        0.0::numeric, 0.5::numeric, 5.8::numeric, 6.8::numeric)
) AS gs(stage, ec_min, ec_max, ph_min, ph_max) ON TRUE
WHERE z.farm_id = 1
  AND z.name IN ('Veg Room', 'Flower Room', 'Outdoor Garden')
ON CONFLICT (farm_id, zone_id, growth_stage) DO NOTHING;

INSERT INTO gr33nfertigation.programs
    (farm_id, name, description, application_recipe_id, reservoir_id, target_zone_id, schedule_id,
     ec_target_id, total_volume_liters, run_duration_seconds, ec_trigger_low, ph_trigger_low, ph_trigger_high, is_active)
SELECT
    1,
    'Veg Daily JLF Program',
    'Daily veg-room fertigation run based on JLF + JMS soil drench recipe.',
    r.id,
    rv.id,
    z.id,
    s.id,
    et.id,
    120.000,
    900,
    1.200,
    5.8,
    6.8,
    TRUE
FROM gr33ncore.zones z
LEFT JOIN gr33nnaturalfarming.application_recipes r
    ON r.farm_id = 1 AND r.name = 'JLF and JMS Combined Drench'
LEFT JOIN gr33ncore.schedules s
    ON s.farm_id = 1 AND s.name = 'Water Late Veg Daily'
LEFT JOIN gr33nfertigation.reservoirs rv
    ON rv.farm_id = 1 AND rv.name = 'Main Nutrient Reservoir'
LEFT JOIN gr33nfertigation.ec_targets et
    ON et.farm_id = 1 AND et.zone_id = z.id
   AND et.growth_stage = 'late_veg'::gr33nfertigation.growth_stage_enum
WHERE z.farm_id = 1
  AND z.name = 'Veg Room'
  AND NOT EXISTS (
      SELECT 1
      FROM gr33nfertigation.programs p
      WHERE p.farm_id = 1
        AND p.name = 'Veg Daily JLF Program'
        AND p.deleted_at IS NULL
  );

-- Phase 39 WS8 — demo prerequisites for automated mix (base EC + delivery pump + parseable dilution)
INSERT INTO gr33ncore.actuators
    (device_id, farm_id, zone_id, name, actuator_type, hardware_identifier, current_state_text, config)
SELECT
    d.id,
    1,
    d.zone_id,
    'Veg Room Irrigation Pump',
    'pump',
    'relay_2',
    'offline',
    '{"channel": 2, "simulation": true}'::jsonb
FROM gr33ncore.devices d
WHERE d.farm_id = 1
  AND d.device_uid = 'demo-veg-relay-01'
  AND NOT EXISTS (
      SELECT 1 FROM gr33ncore.actuators a
      WHERE a.farm_id = 1 AND a.name = 'Veg Room Irrigation Pump' AND a.deleted_at IS NULL
  );

UPDATE gr33nfertigation.reservoirs rv
SET last_ec_mscm = 0.20,
    last_ph = 7.0,
    last_reading_time = NOW(),
    delivery_actuator_id = (
        SELECT a.id FROM gr33ncore.actuators a
        WHERE a.farm_id = 1 AND a.name = 'Veg Room Irrigation Pump' AND a.deleted_at IS NULL
        LIMIT 1
    )
WHERE rv.farm_id = 1
  AND rv.name = 'Main Nutrient Reservoir'
  AND rv.deleted_at IS NULL;

-- Override combo recipe text with a parseable ratio for cloud MixPlan demo (recipe row unchanged)
UPDATE gr33nfertigation.programs
SET dilution_ratio = '1:500'
WHERE farm_id = 1
  AND name = 'Veg Daily JLF Program'
  AND deleted_at IS NULL
  AND dilution_ratio IS NULL;

-- Phase 39b — plain irrigation (RO/well): pulse only, no recipe or mix_batch
INSERT INTO gr33nfertigation.programs
    (farm_id, name, description, application_recipe_id, reservoir_id, target_zone_id, schedule_id,
     total_volume_liters, run_duration_seconds, ec_trigger_low, ph_trigger_low, ph_trigger_high,
     is_active, irrigation_only)
SELECT
    1,
    'Outdoor Well Pulse',
    'Municipal/RO water — timed pump pulse only. No nutrient mix (Phase 39b demo).',
    NULL,
    rv.id,
    z.id,
    s.id,
    40.000,
    120,
    0.0,
    6.0,
    8.0,
    TRUE,
    TRUE
FROM gr33ncore.zones z
LEFT JOIN gr33ncore.schedules s
    ON s.farm_id = 1 AND s.name = 'Water Outdoor Garden Daily'
LEFT JOIN gr33nfertigation.reservoirs rv
    ON rv.farm_id = 1 AND rv.name = 'Outdoor Drench Tank'
WHERE z.farm_id = 1
  AND z.name = 'Outdoor Garden'
  AND NOT EXISTS (
      SELECT 1 FROM gr33nfertigation.programs p
      WHERE p.farm_id = 1 AND p.name = 'Outdoor Well Pulse' AND p.deleted_at IS NULL
  );

INSERT INTO gr33nfertigation.fertigation_events
    (farm_id, program_id, reservoir_id, zone_id, applied_at, growth_stage,
     volume_applied_liters, run_duration_seconds, ec_before_mscm, ec_after_mscm,
     ph_before, ph_after, trigger_source, notes)
SELECT
    1,
    p.id,
    rv.id,
    z.id,
    TIMESTAMPTZ '2026-03-01 08:00:00+00',
    'late_veg'::gr33nfertigation.growth_stage_enum,
    112.500,
    860,
    1.150,
    1.720,
    6.05,
    6.22,
    'schedule_cron'::gr33nfertigation.program_trigger_enum,
    'Seeded historical fertigation event for API/UI demo baseline.'
FROM gr33ncore.zones z
LEFT JOIN gr33nfertigation.programs p
    ON p.farm_id = 1 AND p.name = 'Veg Daily JLF Program' AND p.deleted_at IS NULL
LEFT JOIN gr33nfertigation.reservoirs rv
    ON rv.farm_id = 1 AND rv.name = 'Main Nutrient Reservoir'
WHERE z.farm_id = 1
  AND z.name = 'Veg Room'
  AND NOT EXISTS (
      SELECT 1
      FROM gr33nfertigation.fertigation_events fe
      WHERE fe.farm_id = 1
        AND fe.applied_at = TIMESTAMPTZ '2026-03-01 08:00:00+00'
  );

-- ===========================================================================
-- SECTION 5: AUTOMATION RULES
-- ===========================================================================

INSERT INTO gr33ncore.automation_rules
    (farm_id, name, description, is_active, trigger_source,
     trigger_configuration, condition_logic)
SELECT
    1,
    v.name,
    v.description,
    FALSE,
    'specific_time_cron'::gr33ncore.automation_trigger_source_enum,
    v.trigger_configuration::jsonb,
    'ALL'
FROM (VALUES
    ('AUTO Light ON 18/6 Veg',
     'Turn grow lights ON at 06:00 for 18/6 vegetative schedule.',
     '{"cron": "0 6 * * *", "timezone": "America/New_York", "action": "actuator_on", "target_zone": "Veg Room"}'),
    ('AUTO Light OFF 18/6 Veg',
     'Turn grow lights OFF at midnight for 18/6 vegetative schedule.',
     '{"cron": "0 0 * * *", "timezone": "America/New_York", "action": "actuator_off", "target_zone": "Veg Room"}'),
    ('AUTO Light ON 12/12 Flower',
     'Turn grow lights ON at 06:00 for 12/12 flowering schedule.',
     '{"cron": "0 6 * * *", "timezone": "America/New_York", "action": "actuator_on", "target_zone": "Flower Room"}'),
    ('AUTO Light OFF 12/12 Flower',
     'Turn grow lights OFF at 18:00. 12 hours uninterrupted dark triggers flowering.',
     '{"cron": "0 18 * * *", "timezone": "America/New_York", "action": "actuator_off", "target_zone": "Flower Room"}'),
    ('AUTO Light ON 16/8 Moderate Veg',
     'Turn grow lights ON at 06:00 for 16/8 schedule.',
     '{"cron": "0 6 * * *", "timezone": "America/New_York", "action": "actuator_on", "target_zone": "Veg Room"}'),
    ('AUTO Light OFF 16/8 Moderate Veg',
     'Turn grow lights OFF at 22:00 for 16/8 schedule.',
     '{"cron": "0 22 * * *", "timezone": "America/New_York", "action": "actuator_off", "target_zone": "Veg Room"}')
) AS v(name, description, trigger_configuration)
WHERE NOT EXISTS (
    SELECT 1 FROM gr33ncore.automation_rules r
    WHERE r.farm_id = 1 AND r.name = v.name
);

-- ===========================================================================
-- SECTION 6: SENSOR TEMPLATES
-- ===========================================================================

INSERT INTO gr33ncore.sensors
    (farm_id, name, sensor_type, unit_id,
     value_min_expected, value_max_expected,
     alert_threshold_low, alert_threshold_high,
     reading_interval_seconds, config)
SELECT
    1,
    s.name,
    s.sensor_type,
    u.id,
    s.vmin, s.vmax,
    s.alert_low, s.alert_high,
    s.interval_sec,
    s.config::jsonb
FROM (VALUES
    ('PAR Sensor Indoor',     'par',          'par_umol',          0,   2000,   100,  1800, 60,
     '{"notes":"Seedling 100-300, Veg 400-600, Flower 600-900 umol/m2/s"}'),
    ('Lux Sensor Indoor',     'light_lux',    'lux',               0,   100000, 1000, 80000,60,
     '{"notes":"Seedling 5000-15000 lux, Veg 15000-40000 lux"}'),
    ('Air Temp Indoor',       'temperature',  'celsius',           10,  40,     16,   32,   60,
     '{"notes":"Seedling 22-26C, Veg 20-28C, Flower 18-26C"}'),
    ('Root Zone Temp',        'temperature',  'celsius',           15,  30,     18,   26,   120,
     '{"notes":"Optimal 18-22C. Below 15C stresses roots."}'),
    ('Air Humidity Indoor',   'humidity',     'percent',           20,  90,     35,   75,   60,
     '{"notes":"Seedling 65-70%, Veg 50-70%, Flower 40-50%, Late Flower 35-45%"}'),
    ('Soil Moisture Outdoor', 'soil_moisture','percent',           0,   100,    20,   85,   300,
     '{"notes":"Water at 30-40%. Avoid above 85% anaerobic."}'),
    ('Media Moisture Indoor', 'soil_moisture','percent',           0,   100,    25,   80,   120,
     '{"notes":"Allow dry-back to 30-40% in veg. Less dry-back in flower."}'),
    ('EC Sensor',             'conductivity', 'ms_per_cm',         0,   5,      0.5,  3.5,  60,
     '{"notes":"Seedling 0.5-1.2, Veg 1.2-2.0, Flower 1.6-2.4, Flush <0.5 mS/cm"}'),
    ('pH Sensor',             'ph',           'ph_unit',           4,   9,      5.5,  7.0,  60,
     '{"notes":"Soil 6.0-7.0, Hydro/Coco 5.5-6.5. Check daily in hydro."}'),
    ('CO2 Sensor Indoor',     'co2',          'parts_per_million', 300, 2000,   400,  1500, 60,
     '{"notes":"Ambient 400ppm. Enrichment 800-1200ppm veg, 1000-1500ppm flower."}')
) AS s(name, sensor_type, unit_name, vmin, vmax, alert_low, alert_high, interval_sec, config)
JOIN gr33ncore.units u ON u.name = s.unit_name
WHERE NOT EXISTS (
    SELECT 1 FROM gr33ncore.sensors existing
    WHERE existing.farm_id = 1
      AND existing.name = s.name
      AND existing.deleted_at IS NULL
);

COMMIT;

-- ===========================================================================
-- VERIFY
-- ===========================================================================
SELECT 'auth_users'          AS table_name, count(*) AS rows FROM auth.users                                   UNION ALL
SELECT 'profiles',                          count(*)         FROM gr33ncore.profiles                           UNION ALL
SELECT 'farms',                             count(*)         FROM gr33ncore.farms                              UNION ALL
SELECT 'zones',                             count(*)         FROM gr33ncore.zones                              UNION ALL
SELECT 'input_definitions',                 count(*)         FROM gr33nnaturalfarming.input_definitions        UNION ALL
SELECT 'application_recipes',               count(*)         FROM gr33nnaturalfarming.application_recipes      UNION ALL
SELECT 'recipe_components',                 count(*)         FROM gr33nnaturalfarming.recipe_input_components  UNION ALL
SELECT 'schedules',                         count(*)         FROM gr33ncore.schedules                          UNION ALL
SELECT 'automation_rules',                  count(*)         FROM gr33ncore.automation_rules                   UNION ALL
SELECT 'sensor_templates',                  count(*)         FROM gr33ncore.sensors
ORDER BY 1;

-- ── Sensor → Zone assignments ─────────────────────────────────────────────
-- Assigned 2026-03-05. Zone IDs match gr33n Demo Farm (farm_id = 1).
UPDATE gr33ncore.sensors SET zone_id = (SELECT id FROM gr33ncore.zones WHERE farm_id = 1 AND name = 'Veg Room' AND deleted_at IS NULL ORDER BY id LIMIT 1)      WHERE farm_id = 1 AND name = 'Root Zone Temp';
UPDATE gr33ncore.sensors SET zone_id = (SELECT id FROM gr33ncore.zones WHERE farm_id = 1 AND name = 'Veg Room' AND deleted_at IS NULL ORDER BY id LIMIT 1)      WHERE farm_id = 1 AND name = 'Air Temp Indoor';
UPDATE gr33ncore.sensors SET zone_id = (SELECT id FROM gr33ncore.zones WHERE farm_id = 1 AND name = 'Veg Room' AND deleted_at IS NULL ORDER BY id LIMIT 1)      WHERE farm_id = 1 AND name = 'Media Moisture Indoor';
UPDATE gr33ncore.sensors SET zone_id = (SELECT id FROM gr33ncore.zones WHERE farm_id = 1 AND name = 'Outdoor Garden' AND deleted_at IS NULL ORDER BY id LIMIT 1)  WHERE farm_id = 1 AND name = 'Soil Moisture Outdoor';
UPDATE gr33ncore.sensors SET zone_id = (SELECT id FROM gr33ncore.zones WHERE farm_id = 1 AND name = 'Flower Room' AND deleted_at IS NULL ORDER BY id LIMIT 1)    WHERE farm_id = 1 AND name = 'Air Humidity Indoor';
UPDATE gr33ncore.sensors SET zone_id = (SELECT id FROM gr33ncore.zones WHERE farm_id = 1 AND name = 'Veg Room' AND deleted_at IS NULL ORDER BY id LIMIT 1)       WHERE farm_id = 1 AND name = 'CO2 Sensor Indoor';
UPDATE gr33ncore.sensors SET zone_id = (SELECT id FROM gr33ncore.zones WHERE farm_id = 1 AND name = 'Veg Room' AND deleted_at IS NULL ORDER BY id LIMIT 1)       WHERE farm_id = 1 AND name = 'Lux Sensor Indoor';
UPDATE gr33ncore.sensors SET zone_id = (SELECT id FROM gr33ncore.zones WHERE farm_id = 1 AND name = 'Flower Room' AND deleted_at IS NULL ORDER BY id LIMIT 1)    WHERE farm_id = 1 AND name = 'PAR Sensor Indoor';
UPDATE gr33ncore.sensors SET zone_id = (SELECT id FROM gr33ncore.zones WHERE farm_id = 1 AND name = 'Veg Room' AND deleted_at IS NULL ORDER BY id LIMIT 1)      WHERE farm_id = 1 AND name = 'EC Sensor';
UPDATE gr33ncore.sensors SET zone_id = (SELECT id FROM gr33ncore.zones WHERE farm_id = 1 AND name = 'Veg Room' AND deleted_at IS NULL ORDER BY id LIMIT 1)      WHERE farm_id = 1 AND name = 'pH Sensor';

-- Phase 41 WS5: zone_id on open tasks powers zone Overview + Dashboard morning chips.
INSERT INTO gr33ncore.tasks
  (farm_id, zone_id, schedule_id, title, description, task_type, status, priority, due_date)
SELECT
  1,
  (SELECT id FROM gr33ncore.zones WHERE farm_id=1 AND name = z AND deleted_at IS NULL ORDER BY id LIMIT 1),
  (SELECT id FROM gr33ncore.schedules WHERE farm_id=1 AND name = sched ORDER BY id LIMIT 1),
  title, description, task_type,
  status::gr33ncore.task_status_enum,
  priority,
  due_date::date
FROM (VALUES
  ('Veg Room',        'Water Late Veg Daily',        'Mix JMS batch for veg reservoir',           'Brew 20L JMS from forest leaf mold. Needs 3–7 days ferment. Use in next veg reservoir mix.',  'jadam_prep',    'todo',        2, CURRENT_DATE + 1),
  ('Veg Room',        'Water Late Veg Daily',        'Check veg room EC levels',                  'Target 1.2–2.0 mS/cm for late veg. Adjust JLF drench ratio if drifting.', 'monitoring',    'todo',        2, CURRENT_DATE),
  ('Flower Room',     'Water Early Flower Daily',    'Apply FFJ + WCA foliar spray',              'FFJ 1:500 + WCA 1:1000. Morning spray before lights peak. Follow schedule.', 'jadam_apply',   'in_progress', 3, CURRENT_DATE),
  ('Flower Room',     'Water Early Flower Daily',    'Inspect flower room for powdery mildew',    'Check leaf undersides. Prep JS spray if found. Critical during bloom.', 'scouting',     'in_progress', 2, CURRENT_DATE),
  ('Outdoor Garden',  'Water Outdoor Garden Daily',  'Apply JLF soil drench — outdoor beds',      '1:20 JLF dilution. 3L per sqm. Combine with JMS 1:500 in drench tank.', 'jadam_apply',   'todo',        1, CURRENT_DATE + 2),
  ('Veg Room',        NULL,                          'Calibrate pH sensor',                       'pH drifting — recalibrate with 6.86 and 4.01 buffer solution.', 'maintenance',  'on_hold',     2, CURRENT_DATE + 1),
  ('Flower Room',     NULL,                          'Harvest Flower Room A',                     'Week 9 short-day crop. Flush complete. Check bloom openness and stem length.', 'harvest',       'completed',   3, CURRENT_DATE - 2),
  ('Outdoor Garden',  NULL,                          'Turn compost pile',                         'Aerate pile. Check temp 55–65C. Moisture should clump not drip.', 'soil_prep',    'completed',   1, CURRENT_DATE - 5)
) AS t(z, sched, title, description, task_type, status, priority, due_date)
WHERE NOT EXISTS (
    SELECT 1 FROM gr33ncore.tasks existing
    WHERE existing.farm_id = 1 AND existing.title = t.title
);

-- ===========================================================================
-- SECTION 8: INVENTORY BATCHES + FLOWER FERTIGATION + MIXING HISTORY
-- Typical protocol: 18/6 veg uses JLF+JMS style feeding; 12/12 flower shifts
-- toward flowering inputs (FFJ+WCA here). Farmers mostly tune volumes & cron.
-- ===========================================================================

-- Ready-to-use demo lots (adjust quantities in the UI as you draw them down)
INSERT INTO gr33nnaturalfarming.input_batches (
    farm_id, input_definition_id, batch_identifier, creation_start_date, actual_ready_date,
    quantity_produced, quantity_unit_id, current_quantity_remaining, status,
    storage_location, observations_notes, made_by_user_id
)
SELECT
 1,
    d.id,
    v.batch_identifier,
    v.started::date,
    v.ready::date,
    v.qty_l,
    u.id,
    v.remaining_l,
    'ready_for_use'::gr33nnaturalfarming.input_batch_status_enum,
    v.location,
    v.notes,
    '00000000-0000-0000-0000-000000000001'::uuid
FROM (VALUES
    ('SEED-JLF-GEN-001',  DATE '2026-01-10', DATE '2026-01-24', 45.0::numeric, 38.0::numeric,
     'Veg Room — concentrate shelf',
     'JLF General (weed/grass ferment). Demo lot for 18/6 veg reservoir mixes; dilute 1:20–1:30 to tank.'),
    ('SEED-JMS-001',      DATE '2026-02-01', DATE '2026-02-05', 25.0::numeric, 22.0::numeric,
     'Veg Room — fridge',
     'JMS concentrate. Pair with JLF for combined drench per master recipe.'),
    ('SEED-FFJ-001',      DATE '2026-02-15', DATE '2026-03-01', 8.0::numeric, 6.5::numeric,
     'Flower Room — fridge',
     'FFJ for 12/12 flowering phase — use in lighter feed or foliar per recipe.'),
    ('SEED-WCA-001',      DATE '2026-01-20', DATE '2026-02-10', 12.0::numeric, 10.0::numeric,
     'Flower Room — bench',
     'WCA (eggshell vinegar calcium). Pairs with FFJ during flower.'),
    ('SEED-OHN-001',      DATE '2026-03-01', DATE '2026-03-20', 5.0::numeric,  0.35::numeric,
     'Veg Room — concentrate shelf',
     'OHN (Oriental Herbal Nutrient). Demo lot — remaining below 0.5 L reorder threshold.')
) AS v(batch_identifier, started, ready, qty_l, remaining_l, location, notes)
JOIN gr33ncore.units u ON u.name = 'liter'
JOIN gr33nnaturalfarming.input_definitions d
  ON d.farm_id = 1
 AND d.deleted_at IS NULL
 AND (
 (v.batch_identifier = 'SEED-JLF-GEN-001' AND d.name LIKE 'JLF General%')
   OR (v.batch_identifier = 'SEED-JMS-001' AND d.name LIKE 'JMS%')
   OR (v.batch_identifier = 'SEED-FFJ-001' AND d.name LIKE 'FFJ%')
   OR (v.batch_identifier = 'SEED-WCA-001' AND d.name LIKE 'WCA%')
   OR (v.batch_identifier = 'SEED-OHN-001' AND d.name LIKE 'OHN%')
 )
WHERE NOT EXISTS (
 SELECT 1 FROM gr33nnaturalfarming.input_batches b
    WHERE b.farm_id = 1 AND b.batch_identifier = v.batch_identifier AND b.deleted_at IS NULL
);

INSERT INTO gr33nfertigation.reservoirs
    (farm_id, zone_id, name, description, capacity_liters, current_volume_liters, status)
SELECT
    1,
    (SELECT id FROM gr33ncore.zones WHERE farm_id = 1 AND name = 'Flower Room' AND deleted_at IS NULL ORDER BY id LIMIT 1),
    'Flower Nutrient Reservoir',
    'Dedicated tank for 12/12 flower feeding (FFJ+WCA-style program). Keep separate from veg JLF+JMS tank.',
    400.00,
    220.00,
    'ready'::gr33nfertigation.reservoir_status_enum
ON CONFLICT (farm_id, name) DO NOTHING;

INSERT INTO gr33nfertigation.reservoirs
    (farm_id, zone_id, name, description, capacity_liters, current_volume_liters, status)
SELECT
    1,
    (SELECT id FROM gr33ncore.zones WHERE farm_id = 1 AND name = 'Outdoor Garden' AND deleted_at IS NULL ORDER BY id LIMIT 1),
    'Outdoor Drench Tank',
    'JLF soil drench tank for outdoor raised beds. Fill-and-apply, no recirculation.',
    200.00,
    150.00,
    'ready'::gr33nfertigation.reservoir_status_enum
ON CONFLICT (farm_id, name) DO NOTHING;

INSERT INTO gr33nfertigation.reservoirs
    (farm_id, zone_id, name, description, capacity_liters, current_volume_liters, status)
SELECT
    1,
    (SELECT id FROM gr33ncore.zones WHERE farm_id = 1 AND name = 'Herb & Greens Room' AND deleted_at IS NULL ORDER BY id LIMIT 1),
    'Herb Room Gravity Header',
    'Elevated bucket/tank feeding the gravity drip line. Plain water only — no pump.',
    40.00,
    32.00,
    'ready'::gr33nfertigation.reservoir_status_enum
ON CONFLICT (farm_id, name) DO NOTHING;

UPDATE gr33nfertigation.reservoirs rv
SET delivery_actuator_id = (
    SELECT a.id FROM gr33ncore.actuators a
    WHERE a.farm_id = 1 AND a.name = 'Herb Room Gravity Drip Valve' AND a.deleted_at IS NULL
    LIMIT 1
),
    last_ec_mscm = 0.15,
    last_ph = 7.0,
    last_reading_time = NOW()
WHERE rv.farm_id = 1
  AND rv.name = 'Herb Room Gravity Header'
  AND rv.deleted_at IS NULL;

INSERT INTO gr33nfertigation.programs
    (farm_id, name, description, application_recipe_id, reservoir_id, target_zone_id, schedule_id,
     ec_target_id, total_volume_liters, run_duration_seconds, ec_trigger_low, ph_trigger_low, ph_trigger_high, is_active)
SELECT
    1,
    'Flower Daily FFJ+WCA Program',
    '12/12 flower room: scheduled irrigations aligned with "Water Early Flower Daily" using FFJ+WCA flowering recipe. Tune EC/pH and volumes to your cultivar.',
    r.id,
    rv.id,
    z.id,
    s.id,
    et.id,
    95.000,
    840,
    1.400,
    5.8,
    6.8,
    TRUE
FROM gr33ncore.zones z
JOIN gr33nnaturalfarming.application_recipes r
    ON r.farm_id = 1 AND r.name = 'FFJ and WCA Flowering Boost' AND r.deleted_at IS NULL
JOIN gr33ncore.schedules s
    ON s.farm_id = 1 AND s.name = 'Water Early Flower Daily'
JOIN gr33nfertigation.reservoirs rv
    ON rv.farm_id = 1 AND rv.name = 'Flower Nutrient Reservoir'
JOIN gr33nfertigation.ec_targets et
    ON et.farm_id = 1 AND et.zone_id = z.id   AND et.growth_stage = 'early_flower'::gr33nfertigation.growth_stage_enum
WHERE z.farm_id = 1
  AND z.name = 'Flower Room'
  AND NOT EXISTS (
      SELECT 1
      FROM gr33nfertigation.programs p
      WHERE p.farm_id = 1
        AND p.name = 'Flower Daily FFJ+WCA Program'
        AND p.deleted_at IS NULL
  );

INSERT INTO gr33nfertigation.programs
    (farm_id, name, description, application_recipe_id, reservoir_id, target_zone_id, schedule_id,
     total_volume_liters, run_duration_seconds, ec_trigger_low, ph_trigger_low, ph_trigger_high, is_active)
SELECT
    1,
    'Outdoor JLF Soil Drench',
    'Daily outdoor drench: JLF General 1:20 via drench tank. Covers raised beds ~3L/sqm.',
    r.id,
    rv.id,
    z.id,
    s.id,
    60.000,
    600,
    0.800,
    5.8,
    7.0,
    TRUE
FROM gr33ncore.zones z
LEFT JOIN gr33nnaturalfarming.application_recipes r
    ON r.farm_id = 1 AND r.name = 'JLF General Soil Drench'
LEFT JOIN gr33ncore.schedules s
    ON s.farm_id = 1 AND s.name = 'Water Outdoor Garden Daily'
LEFT JOIN gr33nfertigation.reservoirs rv
    ON rv.farm_id = 1 AND rv.name = 'Outdoor Drench Tank'
WHERE z.farm_id = 1
  AND z.name = 'Outdoor Garden'
  AND NOT EXISTS (
      SELECT 1
      FROM gr33nfertigation.programs p
      WHERE p.farm_id = 1
        AND p.name = 'Outdoor JLF Soil Drench'
        AND p.deleted_at IS NULL
  );

-- Phase 164 WS4 — gravity-fed drip demo (plain irrigation, timed drip valve).
INSERT INTO gr33nfertigation.programs
    (farm_id, name, description, application_recipe_id, reservoir_id, target_zone_id, schedule_id,
     total_volume_liters, run_duration_seconds, ec_trigger_low, ph_trigger_low, ph_trigger_high,
     is_active, irrigation_only)
SELECT
    1,
    'Herb Room Gravity Drip',
    'Gravity-fed drip: elevated header tank, drip line, timed valve — no pump required (Phase 164 demo).',
    NULL,
    rv.id,
    z.id,
    s.id,
    8.000,
    180,
    0.0,
    6.0,
    8.0,
    TRUE,
    TRUE
FROM gr33ncore.zones z
LEFT JOIN gr33ncore.schedules s
    ON s.farm_id = 1 AND s.name = 'Water Herbs Gravity Drip Daily'
LEFT JOIN gr33nfertigation.reservoirs rv
    ON rv.farm_id = 1 AND rv.name = 'Herb Room Gravity Header'
WHERE z.farm_id = 1
  AND z.name = 'Herb & Greens Room'
  AND NOT EXISTS (
      SELECT 1 FROM gr33nfertigation.programs p
      WHERE p.farm_id = 1 AND p.name = 'Herb Room Gravity Drip' AND p.deleted_at IS NULL
  );

INSERT INTO gr33ncore.executable_actions
    (program_id, execution_order, action_type, target_actuator_id, action_command, action_parameters)
SELECT
    p.id,
    0,
    'control_actuator'::gr33ncore.executable_action_type_enum,
    a.id,
    'on',
    '{"source":"seed_phase164_gravity_drip"}'::jsonb
FROM gr33nfertigation.programs p
JOIN gr33ncore.actuators a
    ON a.farm_id = 1 AND a.name = 'Herb Room Gravity Drip Valve' AND a.deleted_at IS NULL
WHERE p.farm_id = 1
  AND p.name = 'Herb Room Gravity Drip'
  AND p.deleted_at IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM gr33ncore.executable_actions ea
      WHERE ea.program_id = p.id AND ea.target_actuator_id = a.id
  );

INSERT INTO gr33nfertigation.fertigation_events
    (farm_id, program_id, reservoir_id, zone_id, applied_at, growth_stage,
     volume_applied_liters, run_duration_seconds, ec_before_mscm, ec_after_mscm,
     ph_before, ph_after, trigger_source, notes)
SELECT
    1,
    p.id,
    rv.id,
    z.id,
    NOW() - INTERVAL '18 hours',
    'late_veg'::gr33nfertigation.growth_stage_enum,
    8.000,
    180,
    0.12,
    0.15,
    7.00,
    7.02,
    'schedule_cron'::gr33nfertigation.program_trigger_enum,
    '[seed:herb-gravity-drip-demo] Plain water gravity drip — header tank to herb bed.'
FROM gr33ncore.zones z
JOIN gr33nfertigation.programs p
    ON p.farm_id = 1 AND p.name = 'Herb Room Gravity Drip' AND p.deleted_at IS NULL
JOIN gr33nfertigation.reservoirs rv
    ON rv.farm_id = 1 AND rv.name = 'Herb Room Gravity Header'
WHERE z.farm_id = 1
  AND z.name = 'Herb & Greens Room'
  AND NOT EXISTS (
      SELECT 1 FROM gr33nfertigation.fertigation_events fe
      WHERE fe.farm_id = 1 AND fe.notes LIKE '%[seed:herb-gravity-drip-demo]%'
  );

INSERT INTO gr33nfertigation.fertigation_events
    (farm_id, program_id, reservoir_id, zone_id, applied_at, growth_stage,
     volume_applied_liters, run_duration_seconds, ec_before_mscm, ec_after_mscm,
     ph_before, ph_after, trigger_source, notes)
SELECT
    1,
    p.id,
    rv.id,
    z.id,
    TIMESTAMPTZ '2026-03-05 08:00:00+00',
    'early_flower'::gr33nfertigation.growth_stage_enum,
    88.000,
    800,
    1.550,
    1.920,
    6.00,
    6.18,
    'schedule_cron'::gr33nfertigation.program_trigger_enum,
    'Seeded flower-room fertigation run (12/12 protocol baseline).'
FROM gr33ncore.zones z
JOIN gr33nfertigation.programs p
    ON p.farm_id = 1 AND p.name = 'Flower Daily FFJ+WCA Program' AND p.deleted_at IS NULL
JOIN gr33nfertigation.reservoirs rv
    ON rv.farm_id = 1 AND rv.name = 'Flower Nutrient Reservoir'
WHERE z.farm_id = 1
  AND z.name = 'Flower Room'
  AND NOT EXISTS (
      SELECT 1
      FROM gr33nfertigation.fertigation_events fe
      WHERE fe.farm_id = 1
        AND fe.applied_at = TIMESTAMPTZ '2026-03-05 08:00:00+00'
        AND fe.zone_id = z.id
  );

INSERT INTO gr33nfertigation.fertigation_events
    (farm_id, program_id, reservoir_id, zone_id, applied_at, growth_stage,
     volume_applied_liters, run_duration_seconds, ec_before_mscm, ec_after_mscm,
     ph_before, ph_after, trigger_source, notes)
SELECT
    1,
    p.id,
    rv.id,
    z.id,
    TIMESTAMPTZ '2026-03-08 07:15:00+00',
    'early_veg'::gr33nfertigation.growth_stage_enum,
    55.000,
    580,
    0.350,
    0.950,
    6.40,
    6.55,
    'schedule_cron'::gr33nfertigation.program_trigger_enum,
    'Seeded outdoor JLF soil drench — raised beds morning run.'
FROM gr33ncore.zones z
JOIN gr33nfertigation.programs p
    ON p.farm_id = 1 AND p.name = 'Outdoor JLF Soil Drench' AND p.deleted_at IS NULL
JOIN gr33nfertigation.reservoirs rv
    ON rv.farm_id = 1 AND rv.name = 'Outdoor Drench Tank'
WHERE z.farm_id = 1
  AND z.name = 'Outdoor Garden'
  AND NOT EXISTS (
      SELECT 1
      FROM gr33nfertigation.fertigation_events fe
      WHERE fe.farm_id = 1
        AND fe.applied_at = TIMESTAMPTZ '2026-03-08 07:15:00+00'
        AND fe.zone_id = z.id
  );

-- Mixing logs: what went into each reservoir (inventory draw + water volume)
DO $$
DECLARE
    r_veg      BIGINT;
    r_flower   BIGINT;
    r_outdoor  BIGINT;
    p_veg      BIGINT;
    p_flower   BIGINT;
    p_outdoor  BIGINT;
    mix_veg    BIGINT;
    mix_fl     BIGINT;
    mix_out    BIGINT;
    b_jlf      BIGINT;
    b_jms      BIGINT;
    b_ffj      BIGINT;
    b_wca      BIGINT;
    i_jlf      BIGINT;
    i_jms      BIGINT;
    i_ffj      BIGINT;
    i_wca      BIGINT;
    cc_veg     BIGINT;
    cc_flower  BIGINT;
    cc_outdoor BIGINT;
BEGIN
    SELECT id INTO r_veg FROM gr33nfertigation.reservoirs WHERE farm_id = 1 AND name = 'Main Nutrient Reservoir' LIMIT 1;
    SELECT id INTO r_flower FROM gr33nfertigation.reservoirs WHERE farm_id = 1 AND name = 'Flower Nutrient Reservoir' LIMIT 1;
    SELECT id INTO r_outdoor FROM gr33nfertigation.reservoirs WHERE farm_id = 1 AND name = 'Outdoor Drench Tank' LIMIT 1;
    SELECT id INTO p_veg FROM gr33nfertigation.programs WHERE farm_id = 1 AND name = 'Veg Daily JLF Program' AND deleted_at IS NULL LIMIT 1;
    SELECT id INTO p_flower FROM gr33nfertigation.programs WHERE farm_id = 1 AND name = 'Flower Daily FFJ+WCA Program' AND deleted_at IS NULL LIMIT 1;
    SELECT id INTO p_outdoor FROM gr33nfertigation.programs WHERE farm_id = 1 AND name = 'Outdoor JLF Soil Drench' AND deleted_at IS NULL LIMIT 1;

    SELECT id INTO i_jlf FROM gr33nnaturalfarming.input_definitions WHERE farm_id = 1 AND name LIKE 'JLF General%' AND deleted_at IS NULL LIMIT 1;
    SELECT id INTO i_jms FROM gr33nnaturalfarming.input_definitions WHERE farm_id = 1 AND name LIKE 'JMS%' AND deleted_at IS NULL LIMIT 1;
    SELECT id INTO i_ffj FROM gr33nnaturalfarming.input_definitions WHERE farm_id = 1 AND name LIKE 'FFJ%' AND deleted_at IS NULL LIMIT 1;
    SELECT id INTO i_wca FROM gr33nnaturalfarming.input_definitions WHERE farm_id = 1 AND name LIKE 'WCA%' AND deleted_at IS NULL LIMIT 1;

    SELECT id INTO b_jlf FROM gr33nnaturalfarming.input_batches WHERE farm_id = 1 AND batch_identifier = 'SEED-JLF-GEN-001' AND deleted_at IS NULL LIMIT 1;
    SELECT id INTO b_jms FROM gr33nnaturalfarming.input_batches WHERE farm_id = 1 AND batch_identifier = 'SEED-JMS-001' AND deleted_at IS NULL LIMIT 1;
    SELECT id INTO b_ffj FROM gr33nnaturalfarming.input_batches WHERE farm_id = 1 AND batch_identifier = 'SEED-FFJ-001' AND deleted_at IS NULL LIMIT 1;
    SELECT id INTO b_wca FROM gr33nnaturalfarming.input_batches WHERE farm_id = 1 AND batch_identifier = 'SEED-WCA-001' AND deleted_at IS NULL LIMIT 1;

    IF r_veg IS NOT NULL AND NOT EXISTS (
        SELECT 1 FROM gr33nfertigation.mixing_events me
        WHERE me.reservoir_id = r_veg AND me.notes LIKE '%[seed:veg-mix-demo]%'
    ) THEN
        INSERT INTO gr33nfertigation.mixing_events (
            farm_id, reservoir_id, program_id, mixed_by_user_id, mixed_at,
            water_volume_liters, water_source, water_ec_mscm, water_ph,
            final_ec_mscm, final_ph, ec_target_met,
            notes
        ) VALUES (
            1, r_veg, p_veg, '00000000-0000-0000-0000-000000000001'::uuid, TIMESTAMPTZ '2026-03-01 07:15:00+00',
            300.0, 'RO + rain blend', 0.05, 6.50,
            1.65, 6.12, TRUE,
            '18/6 veg protocol demo mix: JLF+JMS style batch before irrigations. [seed:veg-mix-demo]'
        ) RETURNING id INTO mix_veg;

        IF i_jlf IS NOT NULL THEN
            INSERT INTO gr33nfertigation.mixing_event_components
                (mixing_event_id, input_definition_id, input_batch_id, volume_added_ml, dilution_ratio, notes)
            VALUES (mix_veg, i_jlf, b_jlf, 15000.000, 'concentrate to ~1:20 tank',
                    'Demo: ~15 L equivalent concentrate contribution — tune to your real JLF strength.');
        END IF;
        IF i_jms IS NOT NULL THEN
            INSERT INTO gr33nfertigation.mixing_event_components
                (mixing_event_id, input_definition_id, input_batch_id, volume_added_ml, dilution_ratio, notes)
            VALUES (mix_veg, i_jms, b_jms, 600.000, '1:500 in tank',
                    'Demo JMS contribution; adjust to recipe.');
        END IF;
    END IF;

    IF r_flower IS NOT NULL AND NOT EXISTS (
        SELECT 1 FROM gr33nfertigation.mixing_events me
        WHERE me.reservoir_id = r_flower AND me.notes LIKE '%[seed:flower-mix-demo]%'
    ) THEN
        INSERT INTO gr33nfertigation.mixing_events (
            farm_id, reservoir_id, program_id, mixed_by_user_id, mixed_at,
            water_volume_liters, water_source, water_ec_mscm, water_ph,
            final_ec_mscm, final_ph, ec_target_met,
            notes
        ) VALUES (
            1, r_flower, p_flower, '00000000-0000-0000-0000-000000000001'::uuid, TIMESTAMPTZ '2026-03-05 07:20:00+00',
            220.0, 'RO', 0.04, 6.45,
            1.78, 6.05, TRUE,
            '12/12 flower protocol demo mix: FFJ+WCA oriented batch. [seed:flower-mix-demo]'
        ) RETURNING id INTO mix_fl;

        IF i_ffj IS NOT NULL THEN
            INSERT INTO gr33nfertigation.mixing_event_components
                (mixing_event_id, input_definition_id, input_batch_id, volume_added_ml, dilution_ratio, notes)
            VALUES (mix_fl, i_ffj, b_ffj, 2200.000, 'light feed contribution',
                    'Demo FFJ draw — flowering phase; follow FFJ+WCA recipe for final ratios.');
        END IF;
        IF i_wca IS NOT NULL THEN
            INSERT INTO gr33nfertigation.mixing_event_components
                (mixing_event_id, input_definition_id, input_batch_id, volume_added_ml, dilution_ratio, notes)
            VALUES (mix_fl, i_wca, b_wca, 800.000, '1:1000 relative',
                    'Demo WCA contribution paired with FFJ.');
        END IF;
    END IF;

    -- Outdoor mixing event: JLF soil drench from shared JLF batch
    IF r_outdoor IS NOT NULL AND NOT EXISTS (
        SELECT 1 FROM gr33nfertigation.mixing_events me
        WHERE me.reservoir_id = r_outdoor AND me.notes LIKE '%[seed:outdoor-mix-demo]%'
    ) THEN
        INSERT INTO gr33nfertigation.mixing_events (
            farm_id, reservoir_id, program_id, mixed_by_user_id, mixed_at,
            water_volume_liters, water_source, water_ec_mscm, water_ph,
            final_ec_mscm, final_ph, ec_target_met,
            notes
        ) VALUES (
            1, r_outdoor, p_outdoor, '00000000-0000-0000-0000-000000000001'::uuid, TIMESTAMPTZ '2026-03-08 06:45:00+00',
            150.0, 'Rain barrel', 0.03, 6.80,
            0.92, 6.55, TRUE,
            'Outdoor JLF drench mix: 1:20 dilution from shared JLF batch. [seed:outdoor-mix-demo]'
        ) RETURNING id INTO mix_out;

        IF i_jlf IS NOT NULL THEN
            INSERT INTO gr33nfertigation.mixing_event_components
                (mixing_event_id, input_definition_id, input_batch_id, volume_added_ml, dilution_ratio, notes)
            VALUES (mix_out, i_jlf, b_jlf, 7500.000, '1:20 in drench tank',
                    'Draw from shared JLF batch SEED-JLF-GEN-001 for outdoor beds.');
        END IF;
    END IF;
END $$;

UPDATE gr33nfertigation.fertigation_events fe
SET mixing_event_id = me.id
FROM gr33nfertigation.mixing_events me
JOIN gr33nfertigation.reservoirs rv ON me.reservoir_id = rv.id AND rv.farm_id = 1
WHERE fe.farm_id = 1
  AND fe.mixing_event_id IS NULL
  AND fe.applied_at = TIMESTAMPTZ '2026-03-01 08:00:00+00'
  AND rv.name = 'Main Nutrient Reservoir'
  AND me.notes LIKE '%[seed:veg-mix-demo]%';

UPDATE gr33nfertigation.fertigation_events fe
SET mixing_event_id = me.id
FROM gr33nfertigation.mixing_events me
JOIN gr33nfertigation.reservoirs rv ON me.reservoir_id = rv.id AND rv.farm_id = 1
WHERE fe.farm_id = 1
  AND fe.mixing_event_id IS NULL
  AND fe.applied_at = TIMESTAMPTZ '2026-03-05 08:00:00+00'
  AND rv.name = 'Flower Nutrient Reservoir'
  AND me.notes LIKE '%[seed:flower-mix-demo]%';

UPDATE gr33nfertigation.fertigation_events fe
SET mixing_event_id = me.id
FROM gr33nfertigation.mixing_events me
JOIN gr33nfertigation.reservoirs rv ON me.reservoir_id = rv.id AND rv.farm_id = 1
WHERE fe.farm_id = 1
  AND fe.mixing_event_id IS NULL
  AND fe.applied_at = TIMESTAMPTZ '2026-03-08 07:15:00+00'
  AND rv.name = 'Outdoor Drench Tank'
  AND me.notes LIKE '%[seed:outdoor-mix-demo]%';

-- Phase 164 WS1 — retheme demo farm: cannabis → chrysanthemum (catalog + field
-- guides keep cannabis). Idempotent for already-seeded DBs.
DELETE FROM gr33ncrops.plants
WHERE farm_id = 1 AND crop_key = 'cannabis' AND deleted_at IS NULL;

UPDATE gr33nfertigation.crop_cycles
SET batch_label = 'Anastasia Green',
    cycle_notes = 'Match light schedule "Light ON/OFF 18/6 Veg" and Veg Daily JLF fertigation program.'
WHERE farm_id = 1
  AND name = 'Veg canopy (18/6)'
  AND batch_label = 'Blue Dream';

UPDATE gr33nfertigation.crop_cycles
SET name = 'Bloom run (12/12)',
    batch_label = 'Zembla White',
    cycle_notes = 'Match "Light ON/OFF 12/12 Flower" and Flower Daily FFJ+WCA program.'
WHERE farm_id = 1
  AND name = 'Flower run (12/12)'
  AND batch_label = 'Gorilla Glue #4';

UPDATE gr33nfertigation.crop_cycles
SET name = 'Anastasia Green — Run 3 (harvested)',
    batch_label = 'Anastasia Green',
    cycle_notes = 'Previous bloom run. Held in cooler 5 days; graded for stem length and ship date.'
WHERE farm_id = 1
  AND name = 'Blue Dream — Run 3 (harvested)';

UPDATE gr33nfertigation.crop_cycles
SET name = 'Chrysanthemum — Cutting Batch 12',
    batch_label = 'Zembla White',
    cycle_notes = 'Rooting under dome, 24h light, misted 2x daily. Chrysanthemum tip cuttings.'
WHERE farm_id = 1
  AND name = 'OG Kush — Clone Batch 12';

UPDATE gr33ncore.tasks
SET description = 'Week 9 short-day crop. Flush complete. Check bloom openness and stem length.'
WHERE farm_id = 1 AND deleted_at IS NULL AND title = 'Harvest Flower Room A'
  AND description LIKE '%trichomes%';

UPDATE gr33ncore.tasks
SET description = 'Check leaf undersides. Prep JS spray if found. Critical during bloom.'
WHERE farm_id = 1 AND deleted_at IS NULL AND title = 'Inspect flower room for powdery mildew'
  AND description LIKE '%Critical during flower%';

UPDATE gr33ncore.alerts_notifications
SET message_text_rendered = 'Air Humidity Indoor read 72.4% RH (alert threshold 65% for bloom stage). '
    || 'Zone: Flower Room. Consider dehumidification or increased airflow before powdery mildew risk.'
WHERE farm_id = 1
  AND subject_rendered = 'Humidity high — Flower Room'
  AND message_text_rendered LIKE '%late flower%';

-- Phase 124 WS1 — one `plants` catalog row per crop grown on this farm. The
-- specific variety per grow lives on crop_cycles.batch_label instead
-- (Phase 93); this row is just "we grow chrysanthemum / tomato / etc. here".
INSERT INTO gr33ncrops.plants (farm_id, display_name, variety_or_cultivar, crop_key)
VALUES
    (1, 'Chrysanthemum', 'Mixed spray varieties', 'chrysanthemum'),
    (1, 'Tomato',     'Roma',                 'tomato'),
    (1, 'Pepper',     'California Wonder',    'pepper'),
    (1, 'Basil',      'Genovese',             'basil'),
    (1, 'Strawberry', 'Albion (everbearing)', 'strawberry'),
    (1, 'Cilantro',   'Santo',                'cilantro'),
    (1, 'Lettuce',    'Buttercrunch',         'lettuce')
ON CONFLICT DO NOTHING;

-- Re-link chrysanthemum cycles after cannabis plant row removal (Phase 164).
UPDATE gr33nfertigation.crop_cycles cc
SET plant_id = p.id
FROM gr33ncrops.plants p
WHERE cc.farm_id = 1
  AND p.farm_id = 1 AND p.crop_key = 'chrysanthemum' AND p.deleted_at IS NULL
  AND cc.name IN (
    'Veg canopy (18/6)',
    'Bloom run (12/12)', 'Flower run (12/12)',
    'Anastasia Green — Run 3 (harvested)', 'Blue Dream — Run 3 (harvested)',
    'Chrysanthemum — Cutting Batch 12', 'OG Kush — Clone Batch 12'
  );

INSERT INTO gr33nfertigation.crop_cycles
    (farm_id, zone_id, name, batch_label, current_stage, is_active, started_at, cycle_notes)
SELECT
    1,
    z.id,
    'Veg canopy (18/6)',
    'Anastasia Green',
    'late_veg'::gr33nfertigation.growth_stage_enum,
    TRUE,
    CURRENT_DATE - 35,
    'Match light schedule "Light ON/OFF 18/6 Veg" and Veg Daily JLF fertigation program.'
FROM gr33ncore.zones z
WHERE z.farm_id = 1 AND z.name = 'Veg Room'
  AND NOT EXISTS (
    SELECT 1 FROM gr33nfertigation.crop_cycles cc
    WHERE cc.zone_id = z.id AND cc.is_active = TRUE
  );

INSERT INTO gr33nfertigation.crop_cycles
    (farm_id, zone_id, name, batch_label, current_stage, is_active, started_at, cycle_notes)
SELECT
    1,
    z.id,
    'Bloom run (12/12)',
    'Zembla White',
    'early_flower'::gr33nfertigation.growth_stage_enum,
    TRUE,
    CURRENT_DATE - 14,
    'Match "Light ON/OFF 12/12 Flower" and Flower Daily FFJ+WCA program.'
FROM gr33ncore.zones z
WHERE z.farm_id = 1 AND z.name = 'Flower Room'
  AND NOT EXISTS (
    SELECT 1 FROM gr33nfertigation.crop_cycles cc
    WHERE cc.zone_id = z.id AND cc.is_active = TRUE
  );

INSERT INTO gr33nfertigation.crop_cycles
    (farm_id, zone_id, name, batch_label, current_stage, is_active, started_at, cycle_notes)
SELECT
    1,
    z.id,
    'Outdoor raised beds — spring',
    'Roma',
    'early_veg'::gr33nfertigation.growth_stage_enum,
    TRUE,
    CURRENT_DATE - 21,
    'Outdoor garden: JADAM soil drench program + natural light. No photoperiod control.'
FROM gr33ncore.zones z
WHERE z.farm_id = 1 AND z.name = 'Outdoor Garden'
  AND NOT EXISTS (
    SELECT 1 FROM gr33nfertigation.crop_cycles cc
    WHERE cc.zone_id = z.id AND cc.is_active = TRUE
  );

-- Phase 124 WS2 — attach the anchor cycles above to their plant catalog row
-- (they predate the plants table backfill) and give the 3 new beds/rooms
-- + harvested history so the demo farm shows a real mix of grow stages.
UPDATE gr33nfertigation.crop_cycles cc
SET plant_id = p.id
FROM gr33ncrops.plants p, gr33ncore.zones z
WHERE cc.farm_id = 1 AND cc.plant_id IS NULL AND cc.zone_id = z.id
  AND ((z.name = 'Veg Room' AND p.crop_key = 'chrysanthemum' AND cc.name = 'Veg canopy (18/6)')
    OR (z.name = 'Flower Room' AND p.crop_key = 'chrysanthemum' AND cc.name = 'Bloom run (12/12)')
    OR (z.name = 'Outdoor Garden' AND p.crop_key = 'tomato' AND cc.name = 'Outdoor raised beds — spring'))
  AND p.farm_id = 1;

INSERT INTO gr33nfertigation.crop_cycles
    (farm_id, zone_id, plant_id, name, batch_label, current_stage, is_active, started_at, harvested_at, yield_grams, cycle_notes)
SELECT 1, z.id, p.id, v.name, v.batch_label, v.stage::gr33nfertigation.growth_stage_enum, v.is_active, v.started_at::date, v.harvested_at::date, v.yield_grams, v.notes
FROM (VALUES
    ('Flower Room',         'chrysanthemum', 'Anastasia Green — Run 3 (harvested)',      'Anastasia Green',    'dry_cure',     FALSE, (CURRENT_DATE-100)::text, (CURRENT_DATE-15)::text, 412.5, 'Previous bloom run. Held in cooler 5 days; graded for stem length and ship date.'),
    ('Propagation Room',    'chrysanthemum', 'Chrysanthemum — Cutting Batch 12',         'Zembla White',       'clone',        TRUE,  (CURRENT_DATE-9)::text,   NULL,                    NULL,  'Rooting under dome, 24h light, misted 2x daily. Chrysanthemum tip cuttings.'),
    ('Propagation Room',    'tomato',     'Roma — Seedling Tray 4 (transplanted)',        'Roma',               'seedling',     FALSE, (CURRENT_DATE-35)::text, NULL,                    NULL,  'Transplanted to Outdoor Garden after hardening off.'),
    ('Herb & Greens Room',  'basil',      'Genovese Basil — Perpetual Bed',               'Genovese',           'late_veg',     TRUE,  (CURRENT_DATE-25)::text, NULL,                    NULL,  'Cut-and-come-again. Harvest outer leaves weekly.'),
    ('Herb & Greens Room',  'cilantro',   'Santo Cilantro — Cut Batch 2 (harvested)',     'Santo',              'harvest',      FALSE, (CURRENT_DATE-60)::text, (CURRENT_DATE-10)::text, 180,   'Bolted in warm spell, harvested whole plants.'),
    ('Outdoor Pepper Bed',  'pepper',     'California Wonder — Bed 2',                    'California Wonder',  'late_veg',     TRUE,  (CURRENT_DATE-45)::text, NULL,                    NULL,  'Direct-planted after last frost. First flowers forming.'),
    ('Outdoor Berry Patch', 'strawberry', 'Albion Strawberries — Patch A',                'Albion',             'early_flower', TRUE,  (CURRENT_DATE-60)::text, NULL,                    NULL,  'Perennial everbearing patch, second season.'),
    ('Outdoor Garden',      'lettuce',    'Buttercrunch Lettuce — Spring Bed (harvested)','Buttercrunch',       'harvest',      FALSE, (CURRENT_DATE-70)::text, (CURRENT_DATE-25)::text, 2200,  'Spring succession crop, cleared to make way for tomatoes.')
) AS v(zone_name, crop_key, name, batch_label, stage, is_active, started_at, harvested_at, yield_grams, notes)
JOIN gr33ncore.zones z ON z.farm_id = 1 AND z.name = v.zone_name AND z.deleted_at IS NULL
JOIN gr33ncrops.plants p ON p.farm_id = 1 AND p.crop_key = v.crop_key
WHERE NOT EXISTS (
    SELECT 1 FROM gr33nfertigation.crop_cycles cc WHERE cc.zone_id = z.id AND cc.name = v.name
);

-- Phase 124 WS3 — one representative sensor per new bed/room (unwired until
-- an operator assigns real hardware via Virtual Pi, same as other sensors).
INSERT INTO gr33ncore.sensors
    (farm_id, zone_id, name, sensor_type, unit_id, value_min_expected, value_max_expected,
     alert_threshold_low, alert_threshold_high, reading_interval_seconds, config)
SELECT 1, z.id, s.name, s.sensor_type, u.id, s.vmin, s.vmax, s.alert_low, s.alert_high, s.interval_sec, s.config::jsonb
FROM (VALUES
    ('Propagation Room',    'Propagation Dome Temp',     'temperature',   'celsius', 15, 35, 20, 29, 60,  '{"notes":"Chrysanthemum cuttings like it warm: 22-26C dome temp."}'),
    ('Herb & Greens Room',  'Herb Room Air Temp',        'temperature',   'celsius', 10, 35, 16, 28, 60,  '{"notes":"Leafy greens/herbs prefer 18-24C."}'),
    ('Outdoor Pepper Bed',  'Pepper Bed Soil Moisture',  'soil_moisture', 'percent', 0,  100,20, 85, 300, '{"notes":"Peppers: water at 30-40%, drought-tolerant once established."}'),
    ('Outdoor Berry Patch', 'Berry Patch Soil Moisture', 'soil_moisture', 'percent', 0,  100,25, 80, 300, '{"notes":"Strawberries: shallow roots, keep evenly moist 40-60%."}')
) AS s(zone_name, name, sensor_type, unit_name, vmin, vmax, alert_low, alert_high, interval_sec, config)
JOIN gr33ncore.zones z ON z.farm_id = 1 AND z.name = s.zone_name AND z.deleted_at IS NULL
JOIN gr33ncore.units u ON u.name = s.unit_name
WHERE NOT EXISTS (
    SELECT 1 FROM gr33ncore.sensors existing WHERE existing.farm_id = 1 AND existing.name = s.name AND existing.deleted_at IS NULL
);

-- Set primary_program_id on crop cycles
UPDATE gr33nfertigation.crop_cycles cc
SET primary_program_id = p.id
FROM gr33nfertigation.programs p
WHERE cc.farm_id = 1
  AND cc.primary_program_id IS NULL
  AND cc.is_active = TRUE
  AND p.farm_id = 1
  AND p.deleted_at IS NULL
  AND (
    (cc.name = 'Veg canopy (18/6)'             AND p.name = 'Veg Daily JLF Program')
    OR (cc.name = 'Bloom run (12/12)'            AND p.name = 'Flower Daily FFJ+WCA Program')
    OR (cc.name = 'Outdoor raised beds — spring' AND p.name = 'Outdoor JLF Soil Drench')
    OR (cc.name = 'Genovese Basil — Perpetual Bed' AND p.name = 'Herb Room Gravity Drip')
  );

-- Link fertigation events to their crop cycles
UPDATE gr33nfertigation.fertigation_events fe
SET crop_cycle_id = cc.id
FROM gr33nfertigation.crop_cycles cc
WHERE fe.farm_id = 1
  AND fe.crop_cycle_id IS NULL
  AND cc.farm_id = 1
  AND cc.is_active = TRUE
  AND fe.zone_id = cc.zone_id;

-- Protocol tasks with schedule links
INSERT INTO gr33ncore.tasks
 (farm_id, zone_id, schedule_id, title, description, task_type, status, priority, due_date)
SELECT
  1,
  (SELECT id FROM gr33ncore.zones WHERE farm_id = 1 AND name = z AND deleted_at IS NULL ORDER BY id LIMIT 1),
  (SELECT id FROM gr33ncore.schedules WHERE farm_id = 1 AND name = sched ORDER BY id LIMIT 1),
  title, description, task_type,
  status::gr33ncore.task_status_enum,
  priority,
  due_date::date
FROM (VALUES
  ('Veg Room',       'Water Late Veg Daily',       'Refresh veg reservoir mix (18/6)',
   'Main Nutrient Reservoir: JLF+JMS style batch per mixing log. Hit EC 1.4–2.2 late veg.',
   'jadam_mix', 'todo', 2, CURRENT_DATE),
  ('Flower Room',    'Water Early Flower Daily',    'Refresh flower reservoir mix (12/12)',
   'Flower Nutrient Reservoir: FFJ+WCA batch for early flower. Check EC vs early_flower targets.',
   'jadam_mix', 'todo', 2, CURRENT_DATE + 1),
  ('Outdoor Garden', 'Water Outdoor Garden Daily',  'Refresh outdoor drench tank',
   'Outdoor Drench Tank: top up JLF 1:20 mix. Check volume before morning schedule.',
   'jadam_mix', 'todo', 1, CURRENT_DATE + 1)
) AS t(z, sched, title, description, task_type, status, priority, due_date)
WHERE NOT EXISTS (
  SELECT 1 FROM gr33ncore.tasks o
  WHERE o.farm_id = 1 AND o.deleted_at IS NULL AND o.title = t.title
);

-- ===========================================================================
-- SECTION 9: GUARDIAN DEMO ALERTS (Phase 29 WS7)
-- Three unread alerts for farm 1 so Guardian live snapshot + drawer demos work
-- after `make dev-stack-fresh` without manual SQL. Idempotent on subject line.
-- ===========================================================================

INSERT INTO gr33ncore.alerts_notifications (
    farm_id, triggering_event_source_type, triggering_event_source_id,
    severity, subject_rendered, message_text_rendered,
    status, is_read, is_acknowledged, created_at
)
SELECT
    1,
    'input_batch',
    b.id,
    'medium'::gr33ncore.notification_priority_enum,
    'OHN batch below minimum — reorder or brew soon',
    'Batch SEED-OHN-001 has 0.35 L remaining (threshold 0.5 L). '
    || 'OHN (Oriental Herbal Nutrient) is used for immunity drenches at 1:1000. '
    || 'Brew a fresh batch or adjust the reorder point in Inventory.',
    'pending',
    FALSE,
    FALSE,
    NOW() - INTERVAL '3 hours'
FROM gr33nnaturalfarming.input_batches b
WHERE b.farm_id = 1
  AND b.batch_identifier = 'SEED-OHN-001'
  AND b.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM gr33ncore.alerts_notifications a
    WHERE a.farm_id = 1
      AND a.subject_rendered = 'OHN batch below minimum — reorder or brew soon'
  );

INSERT INTO gr33ncore.alerts_notifications (
    farm_id, triggering_event_source_type, triggering_event_source_id,
    severity, subject_rendered, message_text_rendered,
    status, is_read, is_acknowledged, created_at
)
SELECT
    1,
    'sensor',
    s.id,
    'high'::gr33ncore.notification_priority_enum,
    'Humidity high — Flower Room',
    'Air Humidity Indoor read 72.4% RH (alert threshold 65% for bloom stage). '
    || 'Zone: Flower Room. Consider dehumidification or increased airflow before powdery mildew risk.',
    'pending',
    FALSE,
    FALSE,
    NOW() - INTERVAL '45 minutes'
FROM gr33ncore.sensors s
WHERE s.farm_id = 1
  AND s.name = 'Air Humidity Indoor'
  AND s.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM gr33ncore.alerts_notifications a
    WHERE a.farm_id = 1
      AND a.subject_rendered = 'Humidity high — Flower Room'
  );

INSERT INTO gr33ncore.alerts_notifications (
    farm_id, triggering_event_source_type, triggering_event_source_id,
    severity, subject_rendered, message_text_rendered,
    status, is_read, is_acknowledged, created_at
)
SELECT
    1,
    'schedule',
    sch.id,
    'low'::gr33ncore.notification_priority_enum,
    'Light schedule change in 48 hours — Flower Room',
    'Photoperiod transition reminder: Light OFF 12/12 Flower fires in ~48 hours (18:00 America/New_York). '
    || 'Confirm timers and blackout curtains in Flower Room before the flip.',
    'pending',
    FALSE,
    FALSE,
    NOW() - INTERVAL '90 minutes'
FROM gr33ncore.schedules sch
WHERE sch.farm_id = 1
  AND sch.name = 'Light OFF 12/12 Flower'
  AND NOT EXISTS (
    SELECT 1 FROM gr33ncore.alerts_notifications a
    WHERE a.farm_id = 1
      AND a.subject_rendered = 'Light schedule change in 48 hours — Flower Room'
  );

-- ===========================================================================
-- Phase 164 WS2+WS3 — demo sensor readings (living farm, three health states)
-- ===========================================================================
-- Wired sensors (SECTION 6): sparse 6 h history + latest row tagged seed:phase164.
--   healthy  — Veg Room cluster + Outdoor soil + Flower PAR (in-range baselines)
--   attention — Flower Room Air Humidity Indoor latest 72.4% (matches humidity alert)
--   not set up — Phase 124 bed sensors intentionally have NO readings:
--     Propagation Dome Temp, Herb Room Air Temp, Pepper/Berry soil moisture
-- Re-run safe: delete prior phase164_demo rows then reinsert.
DELETE FROM gr33ncore.sensor_readings sr
USING gr33ncore.sensors s
WHERE sr.sensor_id = s.id
  AND s.farm_id = 1
  AND s.deleted_at IS NULL
  AND sr.meta_data @> '{"seed":"phase164_demo"}'::jsonb;

-- Align humidity alert threshold with seeded alert copy (65% bloom-stage band).
UPDATE gr33ncore.sensors
SET alert_threshold_high = 65
WHERE farm_id = 1 AND deleted_at IS NULL AND name = 'Air Humidity Indoor';

INSERT INTO gr33ncore.sensor_readings (reading_time, sensor_id, value_raw, is_valid, meta_data)
SELECT
  NOW() - (gs.n * INTERVAL '30 minutes'),
  s.id,
  ROUND(
    (v.base_val + CASE WHEN gs.n = 0 THEN 0 ELSE 0.08 * ((gs.n % 3) - 1) END)::numeric,
    2
  ),
  TRUE,
  '{"seed":"phase164_demo"}'::jsonb
FROM gr33ncore.sensors s
JOIN (VALUES
  ('PAR Sensor Indoor',      620.0),
  ('Lux Sensor Indoor',    28000.0),
  ('Air Temp Indoor',         24.2),
  ('Root Zone Temp',          21.5),
  ('Air Humidity Indoor',     72.4),
  ('Soil Moisture Outdoor',   41.0),
  ('Media Moisture Indoor',   46.0),
  ('EC Sensor',                1.6),
  ('pH Sensor',                6.1),
  ('CO2 Sensor Indoor',      950.0)
) AS v(sensor_name, base_val) ON v.sensor_name = s.name
CROSS JOIN generate_series(0, 12) AS gs(n)
WHERE s.farm_id = 1 AND s.deleted_at IS NULL;

-- ===========================================================================
-- Phase 164 WS6 — VERIFY (living demo farm)
-- ===========================================================================
SELECT 'phase164_cannabis_plants_farm1' AS check_name, count(*)::int AS n
FROM gr33ncrops.plants
WHERE farm_id = 1 AND crop_key = 'cannabis' AND deleted_at IS NULL
UNION ALL
SELECT 'phase164_chrysanthemum_plants_farm1', count(*)::int
FROM gr33ncrops.plants
WHERE farm_id = 1 AND crop_key = 'chrysanthemum' AND deleted_at IS NULL
UNION ALL
SELECT 'phase164_legacy_cannabis_batch_labels', count(*)::int
FROM gr33nfertigation.crop_cycles
WHERE farm_id = 1
  AND batch_label ~* 'blue dream|gorilla|og kush'
UNION ALL
SELECT 'phase164_demo_sensor_readings', count(*)::int
FROM gr33ncore.sensor_readings sr
JOIN gr33ncore.sensors s ON s.id = sr.sensor_id
WHERE s.farm_id = 1 AND s.deleted_at IS NULL
  AND sr.meta_data @> '{"seed":"phase164_demo"}'::jsonb
UNION ALL
SELECT 'phase164_gravity_drip_program', count(*)::int
FROM gr33nfertigation.programs
WHERE farm_id = 1 AND name = 'Herb Room Gravity Drip' AND deleted_at IS NULL
UNION ALL
SELECT 'phase164_gravity_drip_event', count(*)::int
FROM gr33nfertigation.fertigation_events
WHERE farm_id = 1 AND notes LIKE '%[seed:herb-gravity-drip-demo]%';

-- ===========================================================================
-- SECTION 10: PHASE 177 — propagation light for demo tile story
-- Gives Propagation Room a 24h T5 schedule so Today tiles show plants + light.
-- ===========================================================================

INSERT INTO gr33ncore.devices
    (farm_id, zone_id, name, device_uid, device_type, status, config)
SELECT
    1,
    (SELECT id FROM gr33ncore.zones WHERE farm_id = 1 AND name = 'Propagation Room' AND deleted_at IS NULL ORDER BY id LIMIT 1),
    'Propagation Relay Controller',
    'demo-propagation-relay-01',
    'relay_controller',
    'online'::gr33ncore.device_status_enum,
    '{"simulation": true}'::jsonb
WHERE NOT EXISTS (
    SELECT 1 FROM gr33ncore.devices WHERE farm_id = 1 AND device_uid = 'demo-propagation-relay-01'
);

INSERT INTO gr33ncore.actuators
    (device_id, farm_id, zone_id, name, actuator_type, hardware_identifier, current_state_text, config)
SELECT
    d.id,
    1,
    d.zone_id,
    'Propagation T5 Rack',
    'light',
    'relay_1',
    'on',
    '{"channel": 1, "simulation": true}'::jsonb
FROM gr33ncore.devices d
WHERE d.farm_id = 1
  AND d.device_uid = 'demo-propagation-relay-01'
  AND NOT EXISTS (
      SELECT 1 FROM gr33ncore.actuators a
      WHERE a.farm_id = 1 AND a.name = 'Propagation T5 Rack' AND a.deleted_at IS NULL
  );

INSERT INTO gr33ncore.executable_actions
    (schedule_id, execution_order, action_type, target_actuator_id, action_command, action_parameters)
SELECT
    s.id,
    0,
    'control_actuator'::gr33ncore.executable_action_type_enum,
    a.id,
    'on',
    '{"source":"seed_phase177"}'::jsonb
FROM gr33ncore.schedules s
JOIN gr33ncore.actuators a ON a.farm_id = 1 AND a.name = 'Propagation T5 Rack' AND a.deleted_at IS NULL
WHERE s.farm_id = 1
  AND s.name = 'Light ON 24/0 Continuous'
  AND NOT EXISTS (
      SELECT 1 FROM gr33ncore.executable_actions ea
      WHERE ea.schedule_id = s.id AND ea.target_actuator_id = a.id AND ea.action_command = 'on'
  );

UPDATE gr33ncore.schedules
SET is_active = TRUE
WHERE farm_id = 1
  AND name = 'Light ON 24/0 Continuous';

DO $$
DECLARE
  v_zone_id      BIGINT;
  v_actuator_id  BIGINT;
  v_sch_on_id    BIGINT;
  v_prog_id      BIGINT;
BEGIN
  SELECT id INTO v_zone_id     FROM gr33ncore.zones     WHERE farm_id = 1 AND name = 'Propagation Room' AND deleted_at IS NULL ORDER BY id LIMIT 1;
  SELECT id INTO v_actuator_id FROM gr33ncore.actuators WHERE farm_id = 1 AND name = 'Propagation T5 Rack' AND deleted_at IS NULL ORDER BY id LIMIT 1;
  SELECT id INTO v_sch_on_id   FROM gr33ncore.schedules WHERE farm_id = 1 AND name = 'Light ON 24/0 Continuous' ORDER BY id LIMIT 1;

  IF v_zone_id IS NOT NULL AND v_actuator_id IS NOT NULL AND v_sch_on_id IS NOT NULL
     AND NOT EXISTS (
         SELECT 1 FROM gr33ncore.lighting_programs
         WHERE farm_id = 1 AND name = 'Propagation 24h Photoperiod'
     ) THEN

    INSERT INTO gr33ncore.lighting_programs
      (farm_id, zone_id, actuator_id, name, description,
       on_hours, off_hours, lights_on_at, timezone,
       schedule_on_id, schedule_off_id,
       is_active, metadata)
    VALUES
      (1, v_zone_id, v_actuator_id,
       'Propagation 24h Photoperiod',
       'Clone dome — T5 rack on 24/0 for rooting cuttings. Zone: Propagation Room.',
       24, 0, '00:00', 'America/New_York',
       v_sch_on_id, NULL,
       true, '{"preset_key":"propagation_24_0","source":"seed_phase177"}'::jsonb)
    RETURNING id INTO v_prog_id;

    UPDATE gr33ncore.schedules
       SET meta_data = jsonb_set(COALESCE(meta_data, '{}'::jsonb), '{lighting_program_id}', to_jsonb(v_prog_id))
     WHERE id = v_sch_on_id;

    UPDATE gr33ncore.zones
       SET meta_data = jsonb_set(COALESCE(meta_data, '{}'::jsonb), '{lighting_program_id}', to_jsonb(v_prog_id))
     WHERE id = v_zone_id;
  END IF;
END $$;

-- Demo farm Today canvas background (blob tracked in data/files/)
DELETE FROM gr33ncore.file_attachments
WHERE farm_id = 1 AND file_type = 'farm_layout_background';

WITH ins AS (
  INSERT INTO gr33ncore.file_attachments (
    farm_id, related_module_schema, related_table_name, related_record_id,
    file_name, file_type, file_size_bytes, storage_path, mime_type
  ) VALUES (
    1, 'gr33ncore', 'farms', '1',
    'demo-farm-layout.jpg', 'farm_layout_background', 5676341,
    'farm-1/layout-background/6f6c26e8-f753-4aef-839f-8b68729f0524.jpg',
    'image/jpeg'
  )
  RETURNING id
)
UPDATE gr33ncore.farms f
SET meta_data = jsonb_set(
  COALESCE(f.meta_data, '{}'::jsonb),
  '{layout_background_attachment_id}',
  to_jsonb(ins.id),
  true
)
FROM ins
WHERE f.id = 1;

-- Phase 164 guard — ensure demo farm 1 has no legacy cannabis plant row (seed
-- sections after the first COMMIT can re-run on partial applies).
DELETE FROM gr33ncrops.plants
WHERE farm_id = 1 AND crop_key = 'cannabis' AND deleted_at IS NULL;

-- Phase 182 guard — `go test` against a persistent (non-Docker-fresh) local DB
-- runs the Go smoke suite against farm 1 too. Most smokes clean up after
-- themselves, but a few raw-INSERT test rows without a soft-delete endpoint;
-- their names all end in uniqueName()'s trailing `_<bigrandomint>` (or a
-- known literal test prefix), which no real farm zone/flock/loop name would.
-- crop_cycles.zone_id is FK RESTRICT, so drop those first or the zone DELETE
-- below aborts the whole seed script.
DELETE FROM gr33nfertigation.crop_cycles
WHERE zone_id IN (
    SELECT id FROM gr33ncore.zones
    WHERE farm_id = 1 AND deleted_at IS NULL
      AND (name ~ '_[0-9]{9,}$' OR name ~* '^ws[0-9]+_smoke')
);
DELETE FROM gr33ncore.zones
WHERE farm_id = 1 AND deleted_at IS NULL
  AND (name ~ '_[0-9]{9,}$' OR name ~* '^ws[0-9]+_smoke');
DELETE FROM gr33nanimals.animal_groups
WHERE farm_id = 1 AND deleted_at IS NULL
  AND (label ~ '_[0-9]{9,}$' OR label ~* '^ws[0-9]+_smoke');
DELETE FROM gr33naquaponics.loops
WHERE farm_id = 1 AND deleted_at IS NULL
  AND (label ~ '_[0-9]{9,}$' OR label ~* '^ws[0-9]+_smoke');

-- Phase 181 — demo farm showcases every module, not just crops. Animals and
-- aquaponics default OFF for new farms (farmmodules.SeedDefaults never runs
-- for farm 1 — it's inserted directly above, not via POST /farms), so without
-- this the Animals/Aquaponics nav items 404 on a fresh clone.
INSERT INTO gr33ncore.farm_active_modules (farm_id, module_schema_name, is_enabled)
VALUES
    (1, 'gr33ncrops', true),
    (1, 'gr33nnaturalfarming', true),
    (1, 'gr33nanimals', true),
    (1, 'gr33naquaponics', true)
ON CONFLICT (farm_id, module_schema_name) DO UPDATE SET is_enabled = true;

INSERT INTO gr33ncore.zones (farm_id, name, description, zone_type)
SELECT 1, v.name, v.description, v.zone_type
FROM (VALUES
    ('Chicken Coop',            'Layer flock — coop + run, deep litter bedding.', 'outdoor'),
    ('Sheep Pasture',           'Rotational grazing paddock for the small flock.', 'outdoor'),
    ('Fish Tank',               'Tilapia grow-out tank feeding the aquaponics grow bed.', 'greenhouse'),
    ('Grow Bed (Aquaponics)',   'Media bed grow bed fed by the fish tank loop.', 'greenhouse')
) AS v(name, description, zone_type)
WHERE NOT EXISTS (
    SELECT 1 FROM gr33ncore.zones z
    WHERE z.farm_id = 1 AND z.name = v.name AND z.deleted_at IS NULL
);

INSERT INTO gr33nanimals.animal_groups (farm_id, label, species, count, primary_zone_id)
SELECT 1, v.label, v.species, v.count,
       (SELECT id FROM gr33ncore.zones WHERE farm_id = 1 AND name = v.zone_name AND deleted_at IS NULL ORDER BY id LIMIT 1)
FROM (VALUES
    ('Laying flock',  'chicken', 12, 'Chicken Coop'),
    ('Grazing flock', 'sheep',    6, 'Sheep Pasture')
) AS v(label, species, count, zone_name)
WHERE NOT EXISTS (
    SELECT 1 FROM gr33nanimals.animal_groups g
    WHERE g.farm_id = 1 AND g.label = v.label AND g.deleted_at IS NULL
);

INSERT INTO gr33nanimals.animal_lifecycle_events (farm_id, animal_group_id, event_type, delta_count, notes)
SELECT 1, g.id, 'added', g.count, 'Initial flock — seeded demo data'
FROM gr33nanimals.animal_groups g
WHERE g.farm_id = 1 AND g.label IN ('Laying flock', 'Grazing flock')
  AND NOT EXISTS (
      SELECT 1 FROM gr33nanimals.animal_lifecycle_events e
      WHERE e.animal_group_id = g.id AND e.event_type = 'added'
  );

INSERT INTO gr33naquaponics.loops (farm_id, label, fish_tank_zone_id, grow_bed_zone_id)
SELECT 1, 'Tilapia loop',
       (SELECT id FROM gr33ncore.zones WHERE farm_id = 1 AND name = 'Fish Tank' AND deleted_at IS NULL ORDER BY id LIMIT 1),
       (SELECT id FROM gr33ncore.zones WHERE farm_id = 1 AND name = 'Grow Bed (Aquaponics)' AND deleted_at IS NULL ORDER BY id LIMIT 1)
WHERE NOT EXISTS (
    SELECT 1 FROM gr33naquaponics.loops l WHERE l.farm_id = 1 AND l.label = 'Tilapia loop' AND l.deleted_at IS NULL
);

-- Phase 183 — feeder/waterer/gate actuators for the animal zones, and
-- pump/temp/water-quality sensors for the aquaponics loop, so the demo shows
-- more than tracking-only animal groups. No device_id yet (not wired to a
-- real Pi) — same "add hardware later" state as a fresh crop zone before
-- Pi setup; the UI's wiring badge flags it as unwired either way.
INSERT INTO gr33ncore.actuators (farm_id, zone_id, name, actuator_type, current_state_text, config)
SELECT 1, z.id, v.name, v.actuator_type, 'offline', '{"simulation": true}'::jsonb
FROM (VALUES
    ('Chicken Coop',  'Coop Feeder Hopper',  'feeder_hopper'),
    ('Chicken Coop',  'Coop Waterer Valve',  'water_valve'),
    ('Chicken Coop',  'Coop Run Gate',       'gate'),
    ('Sheep Pasture', 'Pasture Trough Valve','water_valve'),
    ('Sheep Pasture', 'Pasture Gate',        'gate'),
    ('Fish Tank',     'Tilapia Circulation Pump', 'pump'),
    ('Fish Tank',     'Tilapia Air Pump',    'air_pump')
) AS v(zone_name, name, actuator_type)
JOIN gr33ncore.zones z ON z.farm_id = 1 AND z.name = v.zone_name AND z.deleted_at IS NULL
WHERE NOT EXISTS (
    SELECT 1 FROM gr33ncore.actuators a
    WHERE a.farm_id = 1 AND a.zone_id = z.id AND a.name = v.name AND a.deleted_at IS NULL
);

INSERT INTO gr33ncore.sensors (farm_id, zone_id, name, sensor_type, unit_id, value_min_expected, value_max_expected, alert_threshold_low, alert_threshold_high)
SELECT 1, z.id, v.name, v.sensor_type, u.id, v.vmin, v.vmax, v.alert_low, v.alert_high
FROM (VALUES
    ('Fish Tank', 'Fish Tank Water Temp',       'water_temp',        'celsius',           10, 35,   22,   28),
    ('Fish Tank', 'Fish Tank Dissolved Oxygen', 'dissolved_oxygen',  'parts_per_million',  0, 15,    5,   12),
    ('Fish Tank', 'Fish Tank Water Level',      'water_level',       'percent',            0, 100,  60,  100)
) AS v(zone_name, name, sensor_type, unit_name, vmin, vmax, alert_low, alert_high)
JOIN gr33ncore.zones z ON z.farm_id = 1 AND z.name = v.zone_name AND z.deleted_at IS NULL
JOIN gr33ncore.units u ON u.name = v.unit_name
WHERE NOT EXISTS (
    SELECT 1 FROM gr33ncore.sensors s
    WHERE s.farm_id = 1 AND s.zone_id = z.id AND s.name = v.name AND s.deleted_at IS NULL
);

-- Live-ish readings for the fish tank sensors above (phase164_demo block runs earlier
-- in this file, before these sensors exist — seed them here so aquaponics zones aren't
-- all NO DATA on a fresh clone).
DELETE FROM gr33ncore.sensor_readings sr
USING gr33ncore.sensors s
JOIN gr33ncore.zones z ON z.id = s.zone_id
WHERE sr.sensor_id = s.id
  AND s.farm_id = 1
  AND z.name = 'Fish Tank'
  AND s.deleted_at IS NULL
  AND sr.meta_data @> '{"seed":"phase183_aquaponics_demo"}'::jsonb;

INSERT INTO gr33ncore.sensor_readings (reading_time, sensor_id, value_raw, is_valid, meta_data)
SELECT
  NOW() - (gs.n * INTERVAL '20 minutes'),
  s.id,
  ROUND((v.base_val + CASE WHEN gs.n = 0 THEN 0 ELSE 0.05 * ((gs.n % 3) - 1) END)::numeric, 2),
  TRUE,
  '{"seed":"phase183_aquaponics_demo"}'::jsonb
FROM gr33ncore.sensors s
JOIN gr33ncore.zones z ON z.id = s.zone_id AND z.farm_id = 1 AND z.name = 'Fish Tank' AND z.deleted_at IS NULL
JOIN (VALUES
  ('Fish Tank Water Temp',       26.5),
  ('Fish Tank Dissolved Oxygen',  7.2),
  ('Fish Tank Water Level',      88.0)
) AS v(sensor_name, base_val) ON v.sensor_name = s.name
CROSS JOIN generate_series(0, 8) AS gs(n)
WHERE s.farm_id = 1 AND s.deleted_at IS NULL;

-- Grow bed needs a light on the Light tab (fish tank already has pumps on Water tab).
INSERT INTO gr33ncore.actuators (farm_id, zone_id, name, actuator_type, current_state_text, config)
SELECT 1, z.id, 'Grow Bed LED', 'grow_light', 'on', '{"simulation": true}'::jsonb
FROM gr33ncore.zones z
WHERE z.farm_id = 1 AND z.name = 'Grow Bed (Aquaponics)' AND z.deleted_at IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM gr33ncore.actuators a
      WHERE a.farm_id = 1 AND a.zone_id = z.id AND a.name = 'Grow Bed LED' AND a.deleted_at IS NULL
  );

-- Phase 179 — resync every serial/identity sequence to max(id) across the seeded
-- schemas. This file inserts many rows with explicit ids (farm 1, its zones,
-- sensors, etc.) so their sequences never advance via nextval(). Without this,
-- the *first* real INSERT through the API after seeding (e.g. POST /farms)
-- calls nextval() -> 1, collides with the seeded row, and fails with a 500 —
-- this bites both smoke tests and any operator who seeds demo data then
-- creates their first real farm/zone/etc.
DO $$
DECLARE
  r RECORD;
  max_id BIGINT;
BEGIN
  FOR r IN
    SELECT n.nspname AS schema_name, t.relname AS table_name,
           a.attname AS col_name, s.relname AS seq_name
    FROM pg_class s
    JOIN pg_depend d ON d.objid = s.oid AND d.deptype = 'a'
    JOIN pg_class t ON d.refobjid = t.oid
    JOIN pg_namespace n ON n.oid = t.relnamespace
    JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = d.refobjsubid
    WHERE s.relkind = 'S'
      AND n.nspname IN (
        'gr33ncore', 'gr33ncrops', 'gr33nfertigation', 'gr33nnaturalfarming',
        'gr33naquaponics', 'gr33nanimals', 'auth'
      )
  LOOP
    EXECUTE format('SELECT COALESCE(MAX(%I), 0) FROM %I.%I', r.col_name, r.schema_name, r.table_name)
      INTO max_id;
    PERFORM setval(format('%I.%I', r.schema_name, r.seq_name), GREATEST(max_id, 1), max_id > 0);
  END LOOP;
END $$;
