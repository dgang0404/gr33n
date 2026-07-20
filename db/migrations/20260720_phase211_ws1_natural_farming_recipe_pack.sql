-- Phase 211 WS1 — Commons natural farming recipe pack (JADAM indoor starter).
-- Body mirror: data/natural-farming-packs/jadam_indoor_starter_recipes_v1.json

INSERT INTO gr33ncore.commons_catalog_entries (
    slug, title, summary, body, contributor_display, contributor_uri,
    license_spdx, license_notes, tags, published, sort_order
) VALUES (
    'jadam-indoor-starter-recipes-v1',
    'JADAM indoor starter recipes v1',
    'Phase 208 WS0 audited input definitions, application recipes, and components — import via Commons catalog.',
    $body$
{
  "catalog_version": "gr33n.commons_catalog.v1",
  "kind": "natural_farming_recipe_pack",
  "pack_key": "jadam_indoor_starter_recipes_v1",
  "reference_source": "JADAM Organic Farming, Youngsang Cho, 2016",
  "readme_md": "# JADAM indoor starter recipes v1\n\nVetted **input definitions**, **application recipes**, and **components** exported from Phase 208 WS0 audited `master_seed.sql`. Import via Help → Catalog or `POST /farms/{id}/commons/catalog-imports`.\n\nIdempotent per farm on input/recipe **name** — safe to re-import; existing rows are skipped, components upserted.\n",
  "input_definitions": [
    {
      "name": "JMS (JADAM Microbial Solution)",
      "category": "microbial_inoculant",
      "description": "Foundation of JADAM. Diverse microbial community from forest floor leaf mold. Applied to soil and foliage to build beneficial populations and suppress pathogens.",
      "typical_ingredients": "Leaf mold humus (forest floor), boiled potato water (cooled), sea salt (pinch)",
      "preparation_summary": "Boil potato, suspend in mesh bag in 10-20L water with leaf mold and pinch of salt. Ferment 24-72 h at 20-30C until peak foam; use within 6-12 h of peak activity.",
      "storage_guidelines": "Use at peak foam — within 6-12 h of peak, not after a full week idle. Cool, shaded.",
      "safety_precautions": "Chlorinated water kills microbes — use filtered or rain water only.",
      "reference_source": "JADAM Organic Farming, Youngsang Cho, 2016"
    },
    {
      "name": "LAB (Lactic Acid Bacteria Serum)",
      "category": "microbial_inoculant",
      "description": "Concentrated lactic acid bacteria from rice wash and milk culture. Suppresses harmful soil microorganisms and improves soil structure.",
      "typical_ingredients": "Rice wash water (first rinse), fresh whole milk (non-UHT preferred)",
      "preparation_summary": "Ferment rice wash 3-5 days until soured. Mix 1 part into 10 parts milk. Wait 5-7 days. Extract golden serum from bottom layer.",
      "storage_guidelines": "Mix with equal part raw sugar to preserve. Refrigerated 6-12 months.",
      "safety_precautions": "Use golden layer only. Discard curds and white top.",
      "reference_source": "KNF (Cho Han-kyu); often used with JADAM"
    },
    {
      "name": "FPJ (Fermented Plant Juice)",
      "category": "fermented_plant_juice",
      "description": "Made from rapidly growing plant tips (comfrey, nettle, mugwort, bamboo). Rich in plant growth hormones, enzymes, and amino acids. Promotes vigorous veg growth.",
      "typical_ingredients": "Fresh growing tips of fast-growing plants, brown sugar (1:1 by weight)",
      "preparation_summary": "Layer equal weights of chopped plant material and sugar. Seal with breathable cloth. Ferment 3-7 days. Strain and bottle.",
      "storage_guidelines": "Refrigerate after straining. Keeps 6-12 months.",
      "safety_precautions": "Keep sugar ratio accurate. Do not use moldy material.",
      "reference_source": "KNF (Cho Han-kyu); often used with JADAM"
    },
    {
      "name": "FFJ (Fermented Fruit Juice)",
      "category": "fermented_plant_juice",
      "description": "Made from ripe or overripe sweet fruits. High in sugars, enzymes, and potassium. Promotes flowering and fruiting. Apply at transition to reproductive stage.",
      "typical_ingredients": "Ripe/overripe fruits (banana peels work well), brown sugar (1:1 by weight)",
      "preparation_summary": "Chop fruit, mix 1:1 with sugar. Ferment loosely covered 7 days. Strain.",
      "storage_guidelines": "Refrigerate after straining. Use within 6 months.",
      "safety_precautions": "Use during flowering/fruiting only.",
      "reference_source": "KNF (Cho Han-kyu); often used with JADAM"
    },
    {
      "name": "BRV (Brown Rice Vinegar)",
      "category": "fermented_plant_juice",
      "description": "Fermented brown rice vinegar 4-8% acidity. Strengthens cell walls, improves calcium uptake, deters fungal issues.",
      "typical_ingredients": "Organic brown rice vinegar (unpasteurized preferred)",
      "preparation_summary": "Purchase unpasteurized organic BRV. No preparation needed.",
      "storage_guidelines": "Store sealed at room temperature indefinitely.",
      "safety_precautions": "Dilute properly — undiluted burns foliage.",
      "reference_source": "JADAM Organic Farming, Youngsang Cho, 2016"
    },
    {
      "name": "OHN (Oriental Herbal Nutrient)",
      "category": "oriental_herbal_nutrient",
      "description": "Extracted from aromatic herbs and roots (garlic, ginger, angelica, cinnamon). Powerful immune booster and pest deterrent. Used in very small quantities.",
      "typical_ingredients": "Garlic, ginger, Angelica root, cinnamon bark, brown sugar, alcohol ~25% ABV",
      "preparation_summary": "Chop herbs, layer with sugar 1:1, ferment 7 days. Add equal alcohol. Ferment 7 more days. Strain. Combine extracts.",
      "storage_guidelines": "Keeps 1-2 years sealed.",
      "safety_precautions": "Extremely potent — always dilute 1:1000 minimum. Avoid inhaling.",
      "reference_source": "KNF (Cho Han-kyu); often used with JADAM"
    },
    {
      "name": "JHS (JADAM Herbal Solution)",
      "category": "oriental_herbal_nutrient",
      "description": "Water-based extract of aromatic and pest-repellent herbs. Broader spectrum pest deterrent and foliar immune support. Mixed with JWA for natural pesticide sprays.",
      "typical_ingredients": "Fresh herbs (wormwood, artemisia, garlic chives, hot pepper, neem) or Jerusalem artichoke, non-chlorinated water",
      "preparation_summary": "Boil 1 kg fresh plant in mesh bag in 4-5 L water 4-5 hours. Strain finely. Use fresh concentrate.",
      "storage_guidelines": "Use within 2 weeks refrigerated. Strain very fine before loading sprayer.",
      "safety_precautions": "Do not apply on blooms — deters pollinators. Apply morning only.",
      "reference_source": "JADAM Organic Farming, Youngsang Cho, 2016"
    },
    {
      "name": "WCA (Water-Soluble Calcium)",
      "category": "water_soluble_nutrient",
      "description": "Calcium from eggshells dissolved in brown rice vinegar. Strengthens cell walls, improves fruit quality, prevents blossom end rot.",
      "typical_ingredients": "Eggshells (or oyster shells), brown rice vinegar (4-8%)",
      "preparation_summary": "Roast eggshells until lightly brown. Cool. Cover with BRV 1:10. Fizzing will occur. Leave 7 days uncovered. Strain.",
      "storage_guidelines": "Store in open-top container (gases form). Use within 30 days.",
      "safety_precautions": "Container must be breathable. Roast shells well.",
      "reference_source": "KNF (Cho Han-kyu); often used with JADAM"
    },
    {
      "name": "WCS (Water-Soluble Calcium Phosphate)",
      "category": "water_soluble_nutrient",
      "description": "Phosphorus and calcium from charred animal bones in brown rice vinegar. Promotes root development, flowering, and ripening.",
      "typical_ingredients": "Beef or pork bones (charred to white ash), brown rice vinegar",
      "preparation_summary": "Char bones until white ash. Cool. Dissolve in BRV 1:10 for 7 days. Strain.",
      "storage_guidelines": "Store in breathable container. Use within 30 days.",
      "safety_precautions": "Char bones fully to white — partial char gives inconsistent results.",
      "reference_source": "KNF (Cho Han-kyu); often used with JADAM"
    },
    {
      "name": "JWA (JADAM Wetting Agent)",
      "category": "other_extract",
      "description": "Homemade soap from plant oils and wood ash lye. Organic surfactant and contact insecticide for soft-bodied insects (aphids, mites, whitefly).",
      "typical_ingredients": "Plant oil (soybean, canola, or coconut), wood ash lye water",
      "preparation_summary": "Boil wood ash in water, filter lye. Mix with oil 1:1. Boil until soap forms.",
      "storage_guidelines": "Keeps indefinitely dry. Dilute 1:500-1:1000 for spraying.",
      "safety_precautions": "Lye is caustic — wear gloves when making. Do not apply in direct sun.",
      "reference_source": "JADAM Organic Farming, Youngsang Cho, 2016"
    },
    {
      "name": "JS (JADAM Sulfur concentrate)",
      "category": "other_extract",
      "description": "Exothermic JADAM sulfur concentrate (~25% sulfur) for powdery mildew, rust, and mites. Not garden wettable sulfur — batch-made from sulfur, caustic soda, clay, and salt.",
      "typical_ingredients": "Elemental sulfur, caustic soda (NaOH), red clay, phyllite powder, sea salt, water",
      "preparation_summary": "Combine per Cho exothermic batch method; yields ~25% sulfur concentrate. Dilute 0.5-2 L concentrate per 500 L spray water. Add JWA for coverage.",
      "storage_guidelines": "Store concentrate sealed; label batch date. Mix spray same day.",
      "safety_precautions": "Caustic soda is hazardous — full PPE, ventilation. Do not apply above 32C.",
      "reference_source": "JADAM Organic Farming, Youngsang Cho, 2016"
    },
    {
      "name": "JLF General (Weed and Grass)",
      "category": "other_ferment",
      "description": "JADAM Liquid Fertilizer from locally available weeds and grasses. Returns native minerals to soil. Free from farm waste. Start 1:100 if unsure; experienced use 1:20. Seedlings 1:30. Much stronger than JMS.",
      "typical_ingredients": "Fresh untreated weeds and grass clippings, leaf mold (handful), non-chlorinated water",
      "preparation_summary": "Fill container 2/3 with chopped weeds. Add leaf mold as microbial starter. Fill to top with water. Seal. Ferment 7-14 days. Stir every few days. Ready when earthy smell. Strain through cloth before use.",
      "storage_guidelines": "Use strained within 30 days. Sealed undiluted keeps 3 months.",
      "safety_precautions": "Non-chlorinated water only. No herbicide-treated material.",
      "reference_source": "JADAM Organic Farming, Youngsang Cho, 2016"
    },
    {
      "name": "JLF Crop-Specific (Crop Residue)",
      "category": "other_ferment",
      "description": "JLF from the same crop's own residue — most targeted fertilizer possible. Tomato residue for tomatoes, corn stalks for corn.",
      "typical_ingredients": "Crop residue (stems, leaves, roots, not fruit or seed), leaf mold, non-chlorinated water",
      "preparation_summary": "Chop crop residue small. Fill container 2/3. Add leaf mold. Fill with water. Seal. Ferment 10-14 days. Strain.",
      "storage_guidelines": "Use within same season. Label with crop type and date.",
      "safety_precautions": "Healthy residue only — do not use diseased plant material.",
      "reference_source": "JADAM Organic Farming, Youngsang Cho, 2016"
    },
    {
      "name": "JLF Spring (Nettle and Comfrey)",
      "category": "other_ferment",
      "description": "High-nitrogen JLF from dynamic accumulator biomass (nettle, comfrey). Best for spring vegetative growth push. Mines deep minerals — not N-fixing crops.",
      "typical_ingredients": "Fresh stinging nettle tops, comfrey leaves (or either alone), leaf mold, water",
      "preparation_summary": "Harvest tops wearing gloves. Fill container 2/3. Add leaf mold, fill with water. Ferment 7-10 days. Strain.",
      "storage_guidelines": "Use within 2 weeks of straining.",
      "safety_precautions": "Wear gloves harvesting nettle. Very high N — do not over-apply to fruiting plants.",
      "reference_source": "JADAM Organic Farming, Youngsang Cho, 2016"
    },
    {
      "name": "FAA (Fish Amino Acid)",
      "category": "fermented_plant_juice",
      "description": "KNF fish amino acid from fish scraps and brown sugar. High nitrogen and trace minerals.",
      "typical_ingredients": "Fresh fish scraps (no salt), brown sugar (1:1 by weight)",
      "preparation_summary": "Layer fish and brown sugar 1:1. Ferment 3-6 months until bones dissolve. Strain.",
      "storage_guidelines": "Refrigerate after straining. Keeps 6-12 months.",
      "safety_precautions": "Strong odor during ferment — ventilate. Dilute 1:1000 minimum for application.",
      "reference_source": "KNF (Cho Han-kyu); often used with JADAM"
    },
    {
      "name": "Compost Tea Actively Aerated",
      "category": "compost_tea_extract",
      "description": "Brewed extract of finished compost, aerated 24-48h to multiply aerobic microbes. Builds soil food web, suppresses disease. Complements JMS.",
      "typical_ingredients": "Finished compost, unsulfured molasses, kelp meal, de-chlorinated water",
      "preparation_summary": "Add compost in mesh bag to bucket with air stone. Add 1 tbsp molasses per 4L. Brew 24-48 hours. Use within 4 hours of finishing.",
      "storage_guidelines": "Must use within 4 hours — microbes die without oxygen.",
      "safety_precautions": "Never store brewed tea.",
      "reference_source": "Elaine Ingham, Soil Biology Primer"
    }
  ],
  "application_recipes": [
    {
      "name": "JMS Soil Drench",
      "description": "Base soil microbe inoculant. Foundation of all JADAM programs.",
      "target_application_type": "soil_drench",
      "dilution_ratio": "1:10 (JMS:water)",
      "instructions": "Dilute 1:10 (1 part JMS to 10 parts water). Apply 2-4L per sqm of root zone. Morning or evening.",
      "frequency_guidelines": "Every 2 weeks growing season. Monthly dormant.",
      "target_crop_categories": [
        "All crops"
      ],
      "target_growth_stages": [
        "All stages"
      ]
    },
    {
      "name": "JLF General Soil Drench",
      "description": "Primary fertility input. Main fertilizer not a supplement.",
      "target_application_type": "soil_drench",
      "dilution_ratio": "1:20 (JLF:water)",
      "instructions": "Strain JLF through cloth. Start 1:100 if unsure; dilute 1:20 when tested. Apply 2-4L per sqm to root zone.",
      "frequency_guidelines": "Every 1-2 weeks active growth.",
      "target_crop_categories": [
        "All crops"
      ],
      "target_growth_stages": [
        "All stages"
      ]
    },
    {
      "name": "JLF Seedling Drench",
      "description": "Gentler dilution safe for young seedlings and fresh transplants.",
      "target_application_type": "soil_drench",
      "dilution_ratio": "1:30 (JLF:water)",
      "instructions": "Dilute 1:30. Apply 0.5L per tray or 1L per transplant hole.",
      "frequency_guidelines": "Weekly from germination through first 2 weeks after transplant.",
      "target_crop_categories": [
        "All crops"
      ],
      "target_growth_stages": [
        "Seedling",
        "Transplant"
      ]
    },
    {
      "name": "JLF and JMS Combined Drench",
      "description": "Nutrients and microbes in one pass. Core weekly application.",
      "target_application_type": "soil_drench",
      "dilution_ratio": "JLF 1:20 + JMS 1:10 in same water",
      "instructions": "Fill tank. Add JLF 1:20, then JMS 1:10. Apply same day.",
      "frequency_guidelines": "Weekly during peak growing season.",
      "target_crop_categories": [
        "All crops"
      ],
      "target_growth_stages": [
        "All stages"
      ]
    },
    {
      "name": "LAB Soil Conditioner",
      "description": "Suppresses harmful soil pathogens, speeds organic matter breakdown.",
      "target_application_type": "soil_drench",
      "dilution_ratio": "1:1000 (LAB:water)",
      "instructions": "Dilute LAB 1:1000. Apply evenly to soil surface. Water in lightly after.",
      "frequency_guidelines": "Every 2-4 weeks. Especially valuable before transplanting.",
      "target_crop_categories": [
        "All crops"
      ],
      "target_growth_stages": [
        "Pre-plant",
        "Transplant",
        "All stages"
      ]
    },
    {
      "name": "OHN Pest and Immunity Drench",
      "description": "Stimulates plant immune response and deters insects.",
      "target_application_type": "soil_drench",
      "dilution_ratio": "1:1000 (OHN:water)",
      "instructions": "Dilute OHN strictly 1:1000. Apply 1-2L per plant root zone.",
      "frequency_guidelines": "Every 2-4 weeks preventative. Weekly during pest or disease pressure.",
      "target_crop_categories": [
        "All crops"
      ],
      "target_growth_stages": [
        "All stages"
      ]
    },
    {
      "name": "JMS Foliar Spray",
      "description": "Establishes beneficial microbes on leaf surfaces. Suppresses airborne pathogens.",
      "target_application_type": "foliar_spray",
      "dilution_ratio": "1:20 (JMS:water) + JWA",
      "instructions": "Dilute 1:20. Add JWA for leaf coverage. Spray upper and lower leaf surfaces to runoff. Early morning.",
      "frequency_guidelines": "Every 1-2 weeks. More often during high humidity.",
      "target_crop_categories": [
        "All crops"
      ],
      "target_growth_stages": [
        "Vegetative",
        "Early flowering"
      ]
    },
    {
      "name": "FPJ Vegetative Foliar",
      "description": "Promotes rapid vegetative growth. Stop at flowering transition.",
      "target_application_type": "foliar_spray",
      "dilution_ratio": "1:500 to 1:1000 (FPJ:water)",
      "instructions": "Dilute 1:500 normal conditions, 1:1000 in hot weather. Add JWA 1:1000.",
      "frequency_guidelines": "Every 7-14 days during vegetative stage.",
      "target_crop_categories": [
        "Leafy greens",
        "Brassicas",
        "Cucurbits",
        "Tomatoes"
      ],
      "target_growth_stages": [
        "Seedling",
        "Vegetative"
      ]
    },
    {
      "name": "FFJ and WCA Flowering Boost",
      "description": "Supports flowering transition and early fruit set.",
      "target_application_type": "foliar_spray",
      "dilution_ratio": "FFJ 1:500 + WCA 1:1000 combined",
      "instructions": "Mix FFJ 1:500 and WCA 1:1000 in same tank. Add JWA 1:1000. Morning.",
      "frequency_guidelines": "Weekly from first flower buds through early fruit set.",
      "target_crop_categories": [
        "Tomatoes",
        "Peppers",
        "Cucumbers",
        "Squash",
        "Fruit trees"
      ],
      "target_growth_stages": [
        "Flowering",
        "Early fruit"
      ]
    },
    {
      "name": "BRV and WCA Cell Strengthener",
      "description": "Hardens cell walls. Apply before rain, cold snaps, or disease pressure.",
      "target_application_type": "foliar_spray",
      "dilution_ratio": "BRV 1:800 + WCA 1:1000",
      "instructions": "Mix BRV 1:800 and WCA 1:1000. Do not exceed BRV concentration — burn risk.",
      "frequency_guidelines": "Every 2 weeks during fruiting or before stress events.",
      "target_crop_categories": [
        "All crops"
      ],
      "target_growth_stages": [
        "Vegetative",
        "Fruiting"
      ]
    },
    {
      "name": "JHS and JWA Natural Pesticide",
      "description": "Broad-spectrum organic pest deterrent. Effective against chewing and sucking insects.",
      "target_application_type": "foliar_spray",
      "dilution_ratio": "JHS 1:50 + JWA 1:500",
      "instructions": "Strain JHS very finely. Mix JHS 1:50 and JWA 1:500. Apply thorough coverage especially leaf undersides. Morning or evening.",
      "frequency_guidelines": "Weekly preventative. Every 3-5 days for active pest pressure.",
      "target_crop_categories": [
        "All crops"
      ],
      "target_growth_stages": [
        "Any stage"
      ]
    },
    {
      "name": "JS Fungicide Spray",
      "description": "Controls powdery mildew, rust, and spider mites using JADAM sulfur concentrate.",
      "target_application_type": "foliar_spray",
      "dilution_ratio": "0.5-2 L JS concentrate per 500 L water + JWA 1:500",
      "instructions": "Dilute JS concentrate per Cho (0.5-2 L per 500 L). Add JWA. Mix fresh. Apply thorough coverage. Do NOT apply above 32C.",
      "frequency_guidelines": "At first sign of fungal disease. Repeat every 5-7 days.",
      "target_crop_categories": [
        "All crops"
      ],
      "target_growth_stages": [
        "Any stage"
      ]
    },
    {
      "name": "JLF Foliar Feed",
      "description": "Fast nutrient uptake during plant stress or deficiency.",
      "target_application_type": "foliar_spray",
      "dilution_ratio": "1:30 to 1:50 (JLF:water)",
      "instructions": "Strain JLF very finely. Dilute 1:30 min (1:50 hot weather). Add JWA 1:1000.",
      "frequency_guidelines": "Weekly during stress. Not a substitute for soil application.",
      "target_crop_categories": [
        "All crops"
      ],
      "target_growth_stages": [
        "Any stage under stress"
      ]
    },
    {
      "name": "JWA Insecticide Spray",
      "description": "Contact insecticide for aphids, spider mites, whitefly, soft-bodied insects.",
      "target_application_type": "foliar_spray",
      "dilution_ratio": "1:500 (JWA:water)",
      "instructions": "Dilute 1:500. Cover leaf surfaces including undersides. Morning or evening.",
      "frequency_guidelines": "Every 3-5 days for active infestations.",
      "target_crop_categories": [
        "All crops"
      ],
      "target_growth_stages": [
        "Any stage"
      ]
    }
  ],
  "recipe_input_components": [
    {
      "recipe_name": "JMS Soil Drench",
      "input_name": "JMS (JADAM Microbial Solution)",
      "part_value": 1.0,
      "part_unit_name": "decimal_fraction",
      "notes": "1 part JMS to 10 parts water"
    },
    {
      "recipe_name": "JLF General Soil Drench",
      "input_name": "JLF General (Weed and Grass)",
      "part_value": 1.0,
      "part_unit_name": "decimal_fraction",
      "notes": "1 part JLF to 20 parts water"
    },
    {
      "recipe_name": "JLF Seedling Drench",
      "input_name": "JLF General (Weed and Grass)",
      "part_value": 1.0,
      "part_unit_name": "decimal_fraction",
      "notes": "1 part JLF to 30 parts water"
    },
    {
      "recipe_name": "JLF and JMS Combined Drench",
      "input_name": "JLF General (Weed and Grass)",
      "part_value": 1.0,
      "part_unit_name": "decimal_fraction",
      "notes": "JLF at 1:20"
    },
    {
      "recipe_name": "JLF and JMS Combined Drench",
      "input_name": "JMS (JADAM Microbial Solution)",
      "part_value": 2.0,
      "part_unit_name": "decimal_fraction",
      "notes": "JMS at 1:10 relative to 1:20 JLF base"
    },
    {
      "recipe_name": "LAB Soil Conditioner",
      "input_name": "LAB (Lactic Acid Bacteria Serum)",
      "part_value": 1.0,
      "part_unit_name": "decimal_fraction",
      "notes": "1 part LAB to 1000 parts water"
    },
    {
      "recipe_name": "OHN Pest and Immunity Drench",
      "input_name": "OHN (Oriental Herbal Nutrient)",
      "part_value": 1.0,
      "part_unit_name": "decimal_fraction",
      "notes": "1 part OHN to 1000 — never exceed"
    },
    {
      "recipe_name": "JMS Foliar Spray",
      "input_name": "JMS (JADAM Microbial Solution)",
      "part_value": 1.0,
      "part_unit_name": "decimal_fraction",
      "notes": "1 part JMS to 20 parts water"
    },
    {
      "recipe_name": "FPJ Vegetative Foliar",
      "input_name": "FPJ (Fermented Plant Juice)",
      "part_value": 1.0,
      "part_unit_name": "decimal_fraction",
      "notes": "1 part FPJ to 500-1000 parts water"
    },
    {
      "recipe_name": "FFJ and WCA Flowering Boost",
      "input_name": "FFJ (Fermented Fruit Juice)",
      "part_value": 1.0,
      "part_unit_name": "decimal_fraction",
      "notes": "FFJ at 1:500"
    },
    {
      "recipe_name": "FFJ and WCA Flowering Boost",
      "input_name": "WCA (Water-Soluble Calcium)",
      "part_value": 0.5,
      "part_unit_name": "decimal_fraction",
      "notes": "WCA at 1:1000 relative"
    },
    {
      "recipe_name": "BRV and WCA Cell Strengthener",
      "input_name": "BRV (Brown Rice Vinegar)",
      "part_value": 1.0,
      "part_unit_name": "decimal_fraction",
      "notes": "BRV at 1:800"
    },
    {
      "recipe_name": "BRV and WCA Cell Strengthener",
      "input_name": "WCA (Water-Soluble Calcium)",
      "part_value": 0.8,
      "part_unit_name": "decimal_fraction",
      "notes": "WCA at 1:1000 relative"
    },
    {
      "recipe_name": "JHS and JWA Natural Pesticide",
      "input_name": "JHS (JADAM Herbal Solution)",
      "part_value": 1.0,
      "part_unit_name": "decimal_fraction",
      "notes": "JHS at 1:50"
    },
    {
      "recipe_name": "JHS and JWA Natural Pesticide",
      "input_name": "JWA (JADAM Wetting Agent)",
      "part_value": 0.1,
      "part_unit_name": "decimal_fraction",
      "notes": "JWA at 1:500 surfactant"
    },
    {
      "recipe_name": "JS Fungicide Spray",
      "input_name": "JS (JADAM Sulfur concentrate)",
      "part_value": 1.0,
      "part_unit_name": "decimal_fraction",
      "notes": "0.5-2 L JS concentrate per 500 L water"
    },
    {
      "recipe_name": "JS Fungicide Spray",
      "input_name": "JWA (JADAM Wetting Agent)",
      "part_value": 0.1,
      "part_unit_name": "decimal_fraction",
      "notes": "JWA as emulsifier"
    },
    {
      "recipe_name": "JLF Foliar Feed",
      "input_name": "JLF General (Weed and Grass)",
      "part_value": 1.0,
      "part_unit_name": "decimal_fraction",
      "notes": "JLF at 1:30 to 1:50"
    },
    {
      "recipe_name": "JLF Foliar Feed",
      "input_name": "JWA (JADAM Wetting Agent)",
      "part_value": 0.033,
      "part_unit_name": "decimal_fraction",
      "notes": "JWA 1:1000 surfactant"
    },
    {
      "recipe_name": "JWA Insecticide Spray",
      "input_name": "JWA (JADAM Wetting Agent)",
      "part_value": 1.0,
      "part_unit_name": "decimal_fraction",
      "notes": "1 part JWA to 500 parts water"
    }
  ]
}
$body$::jsonb,
    'gr33n platform (natural farming canon)',
    NULL,
    'CC-BY-4.0',
    'Derived from JADAM Organic Farming (Youngsang Cho, 2016) and KNF field guides in gr33n corpus.',
    ARRAY['natural_farming', 'jadam', 'commons', 'phase-211', 'recipe-pack'],
    TRUE,
    35
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
