-- =============================================================================
-- gr33n Master Seed File  v1.004
-- Fix: schedules table has no metadata column — notes moved to description
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

INSERT INTO gr33ncore.farms (
    id, name, description, owner_user_id,
    timezone, currency, scale_tier, operational_status
) VALUES (
    1, 'gr33n Demo Farm',
    'Demo farm pre-loaded with JADAM inputs, light schedules, and watering programs.',
    '00000000-0000-0000-0000-000000000001',
    'America/New_York', 'USD', 'small', 'active'
) ON CONFLICT (id) DO NOTHING;

INSERT INTO gr33ncore.zones (farm_id, name, description, zone_type)
VALUES
    (1, 'Seedling Room', 'Germination and early seedling stage. High humidity, gentle light.', 'indoor'),
    (1, 'Veg Room',      'Vegetative growth stage. 18/6 light, moderate feeding.',             'indoor'),
    (1, 'Flower Room',   'Flowering and fruiting stage. 12/12 light, full feeding program.',   'indoor'),
    (1, 'Outdoor Beds',  'Outdoor raised beds. Natural light. JADAM soil program.',            'outdoor')
ON CONFLICT DO NOTHING;

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
VALUES
(1, 'Light ON 24/0 Continuous',
 'Lights always on. Seedling propagation, cloning, autoflowering varieties.',
 'lighting', '0 0 * * *', 'America/New_York', false),

(1, 'Light ON 18/6 Veg',
 'Lights on at 06:00. 18 hours on for active vegetative growth.',
 'lighting', '0 6 * * *', 'America/New_York', false),

(1, 'Light OFF 18/6 Veg',
 'Lights off at midnight. 6 hours dark.',
 'lighting', '0 0 * * *', 'America/New_York', false),

(1, 'Light ON 16/8 Moderate Veg',
 'Lights on at 06:00. 16 hours on — good energy balance vs 18/6.',
 'lighting', '0 6 * * *', 'America/New_York', false),

(1, 'Light OFF 16/8 Moderate Veg',
 'Lights off at 22:00. 8 hours dark.',
 'lighting', '0 22 * * *', 'America/New_York', false),

(1, 'Light ON 12/12 Flower',
 'Lights on at 06:00. 12 hours on triggers flowering in photoperiod plants.',
 'lighting', '0 6 * * *', 'America/New_York', false),

(1, 'Light OFF 12/12 Flower',
 'Lights off at 18:00. 12 hours uninterrupted dark — critical for flowering.',
 'lighting', '0 18 * * *', 'America/New_York', false)

ON CONFLICT DO NOTHING;

-- ===========================================================================
-- SECTION 4: WATERING SCHEDULES
-- Note: schedules table has no metadata column — volume/stage info in description
-- ===========================================================================

INSERT INTO gr33ncore.schedules
    (farm_id, name, description, schedule_type, cron_expression, timezone, is_active)
VALUES

(1, 'Water Seedling Daily Light',
 'Seedling stage. ~75mL per plant. Water only when top inch of media is dry. '
 'Skip if still moist — overwatering kills more seedlings than underwatering. '
 'Zone: Seedling Room. Light: 24/0 or 18/6.',
 'irrigation', '0 8 * * *', 'America/New_York', false),

(1, 'Water Early Veg Every 2 Days',
 'Early veg. ~300mL per plant every 2 days. Allow slight dry-back between '
 'waterings to encourage roots to chase moisture downward. '
 'Zone: Veg Room. Light: 18/6.',
 'irrigation', '0 8 1-31/2 * *', 'America/New_York', false),

(1, 'Water Late Veg Daily',
 'Late veg with larger root zone. ~750mL per plant daily. '
 'Increase if wilting occurs before next scheduled watering. '
 'Zone: Veg Room. Light: 18/6 or 16/8.',
 'irrigation', '0 8 * * *', 'America/New_York', false),

(1, 'Water Early Flower Daily',
 'First 2 weeks of flowering. ~900mL per plant daily. Slight stress during '
 'stretch week is OK — builds stem density. '
 'Zone: Flower Room. Light: 12/12.',
 'irrigation', '0 8 * * *', 'America/New_York', false),

(1, 'Water Peak Flower 2x Daily',
 'Mid to late flowering — maximum demand. ~1.5L per plant twice daily. '
 'Never let medium go fully dry during peak flower. Watch for leaf curl. '
 'Zone: Flower Room. Light: 12/12.',
 'irrigation', '0 8,18 * * *', 'America/New_York', false),

(1, 'Water Flush Week 2x Daily',
 'Final 7-14 days before harvest. Plain pH-adjusted water only — no nutrients. '
 '~2L per plant twice daily. 1.5-2x pot volume per session to clear salts. '
 'Zone: Flower Room. Light: 12/12.',
 'irrigation', '0 8,18 * * *', 'America/New_York', false),

(1, 'Water Outdoor Beds Morning',
 'Morning base irrigation for outdoor raised beds. ~3L per sqm. '
 'Disable during rain periods. Increase in heat waves. '
 'Apply JLF and JMS here. Zone: Outdoor Beds.',
 'irrigation', '0 7 * * *', 'America/New_York', false)

ON CONFLICT DO NOTHING;

-- ===========================================================================
-- SECTION 4B: FERTIGATION BASELINE DATA
-- ===========================================================================

INSERT INTO gr33nfertigation.reservoirs
    (farm_id, zone_id, name, description, capacity_liters, current_volume_liters, status)
SELECT
    1,
    (SELECT id FROM gr33ncore.zones WHERE farm_id = 1 AND name = 'Veg Room'),
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
  AND z.name IN ('Seedling Room', 'Veg Room', 'Flower Room')
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
VALUES

(1, 'AUTO Light ON 18/6 Veg',
 'Turn grow lights ON at 06:00 for 18/6 vegetative schedule.',
 false, 'specific_time_cron',
 '{"cron": "0 6 * * *", "timezone": "America/New_York",
   "action": "actuator_on", "target_zone": "Veg Room"}',
 'ALL'),

(1, 'AUTO Light OFF 18/6 Veg',
 'Turn grow lights OFF at midnight for 18/6 vegetative schedule.',
 false, 'specific_time_cron',
 '{"cron": "0 0 * * *", "timezone": "America/New_York",
   "action": "actuator_off", "target_zone": "Veg Room"}',
 'ALL'),

(1, 'AUTO Light ON 12/12 Flower',
 'Turn grow lights ON at 06:00 for 12/12 flowering schedule.',
 false, 'specific_time_cron',
 '{"cron": "0 6 * * *", "timezone": "America/New_York",
   "action": "actuator_on", "target_zone": "Flower Room"}',
 'ALL'),

(1, 'AUTO Light OFF 12/12 Flower',
 'Turn grow lights OFF at 18:00. 12 hours uninterrupted dark triggers flowering.',
 false, 'specific_time_cron',
 '{"cron": "0 18 * * *", "timezone": "America/New_York",
   "action": "actuator_off", "target_zone": "Flower Room"}',
 'ALL'),

(1, 'AUTO Light ON 16/8 Moderate Veg',
 'Turn grow lights ON at 06:00 for 16/8 schedule.',
 false, 'specific_time_cron',
 '{"cron": "0 6 * * *", "timezone": "America/New_York",
   "action": "actuator_on", "target_zone": "Veg Room"}',
 'ALL'),

(1, 'AUTO Light OFF 16/8 Moderate Veg',
 'Turn grow lights OFF at 22:00 for 16/8 schedule.',
 false, 'specific_time_cron',
 '{"cron": "0 22 * * *", "timezone": "America/New_York",
   "action": "actuator_off", "target_zone": "Veg Room"}',
 'ALL'),

(1, 'AUTO Light ON 24/0 Seedling',
 'Lights always on for seedling propagation.',
 false, 'specific_time_cron',
 '{"cron": "0 0 * * *", "timezone": "America/New_York",
   "action": "actuator_on", "target_zone": "Seedling Room"}',
 'ALL')

ON CONFLICT DO NOTHING;

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
ON CONFLICT DO NOTHING;

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
UPDATE gr33ncore.sensors SET zone_id = (SELECT id FROM gr33ncore.zones WHERE farm_id = 1 AND name = 'Veg Room')      WHERE farm_id = 1 AND name = 'Root Zone Temp';
UPDATE gr33ncore.sensors SET zone_id = (SELECT id FROM gr33ncore.zones WHERE farm_id = 1 AND name = 'Veg Room')      WHERE farm_id = 1 AND name = 'Air Temp Indoor';
UPDATE gr33ncore.sensors SET zone_id = (SELECT id FROM gr33ncore.zones WHERE farm_id = 1 AND name = 'Veg Room')      WHERE farm_id = 1 AND name = 'Media Moisture Indoor';
UPDATE gr33ncore.sensors SET zone_id = (SELECT id FROM gr33ncore.zones WHERE farm_id = 1 AND name = 'Outdoor Beds')  WHERE farm_id = 1 AND name = 'Soil Moisture Outdoor';
UPDATE gr33ncore.sensors SET zone_id = (SELECT id FROM gr33ncore.zones WHERE farm_id = 1 AND name = 'Flower Room')   WHERE farm_id = 1 AND name = 'Air Humidity Indoor';
UPDATE gr33ncore.sensors SET zone_id = (SELECT id FROM gr33ncore.zones WHERE farm_id = 1 AND name = 'Veg Room')      WHERE farm_id = 1 AND name = 'CO2 Sensor Indoor';
UPDATE gr33ncore.sensors SET zone_id = (SELECT id FROM gr33ncore.zones WHERE farm_id = 1 AND name = 'Seedling Room') WHERE farm_id = 1 AND name = 'Lux Sensor Indoor';
UPDATE gr33ncore.sensors SET zone_id = (SELECT id FROM gr33ncore.zones WHERE farm_id = 1 AND name = 'Flower Room')   WHERE farm_id = 1 AND name = 'PAR Sensor Indoor';
UPDATE gr33ncore.sensors SET zone_id = (SELECT id FROM gr33ncore.zones WHERE farm_id = 1 AND name = 'Veg Room')      WHERE farm_id = 1 AND name = 'EC Sensor';
UPDATE gr33ncore.sensors SET zone_id = (SELECT id FROM gr33ncore.zones WHERE farm_id = 1 AND name = 'Veg Room')      WHERE farm_id = 1 AND name = 'pH Sensor';

-- pre seed scedual
INSERT INTO gr33ncore.tasks
  (farm_id, zone_id, title, description, task_type, status, priority, due_date)
SELECT
  1,
  (SELECT id FROM gr33ncore.zones WHERE farm_id=1 AND name = z),
  title, description, task_type,
  status::gr33ncore.task_status_enum,
  priority,
  due_date::date
FROM (VALUES
  ('Veg Room',     'Mix JMS batch',              'Brew 20L JMS from forest leaf mold. Needs 3–7 days ferment.',  'jadam_prep',    'todo',        2, CURRENT_DATE + 1),
  ('Veg Room',     'Check EC levels',             'Target 1.2–2.0 mS/cm for veg stage. Adjust JLF drench ratio.','monitoring',    'todo',        2, CURRENT_DATE),
  ('Flower Room',  'Apply FFJ + WCA foliar',      'FFJ 1:500 + WCA 1:1000. Morning spray before lights peak.',   'jadam_apply',   'in_progress', 3, CURRENT_DATE),
  ('Flower Room',  'Inspect for powdery mildew',  'Check leaf undersides in flower room. Prep JS spray if found.','scouting',     'in_progress', 2, CURRENT_DATE),
  ('Seedling Room','Transplant seedlings to veg', 'Move ready seedlings from tray to Veg Room. JLF seedling drench first.','transplant','todo', 1, CURRENT_DATE + 3),
  ('Outdoor Beds', 'Apply JLF general drench',    '1:20 dilution. 3L per sqm. Combine with JMS 1:500.',          'jadam_apply',   'todo',        1, CURRENT_DATE + 2),
  ('Veg Room',     'Calibrate pH sensor',         'pH drifting — recalibrate with 6.86 and 4.01 buffer solution.','maintenance',  'on_hold',     2, CURRENT_DATE + 1),
  ('Flower Room',  'Harvest Flower Room A',        'Week 9 photoperiod crop. Flush complete. Check trichomes.',   'harvest',       'completed',   3, CURRENT_DATE - 2),
  ('Outdoor Beds', 'Turn compost pile',            'Aerate pile. Check temp 55–65C. Moisture should clump not drip.','soil_prep',  'completed',   1, CURRENT_DATE - 5),
  (NULL,           'Order OHN ingredients',        'Garlic, ginger, angelica root, cinnamon bark, soju.',         'procurement',   'pending_review',1, CURRENT_DATE + 7)
) AS t(z, title, description, task_type, status, priority, due_date)
ON CONFLICT DO NOTHING;

