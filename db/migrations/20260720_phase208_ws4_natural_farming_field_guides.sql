-- Phase 208 WS4 — natural farming field guides in agronomy_field_guides (RAG DB source)
-- Bodies synced from docs/field-guides/natural-farming-*.md

INSERT INTO gr33ncrops.agronomy_field_guides (
    slug, title, crop_key, guide_kind, domain, safety_tier, body_md, catalog_version, published, sort_order
)
SELECT v.slug, v.title, v.crop_key, v.guide_kind, v.domain, v.safety_tier, v.body_md, v.catalog_version, v.published, v.sort_order
FROM (VALUES
    ('natural-farming-application-recipes', 'Natural farming application recipes (canon)', NULL, 'natural_farming', 'natural_farming', 'safe', $nf_natural-farming-application-recipes$---
domain: natural_farming
title: Natural farming application recipes (canon)
safety_tier: safe
tradition: jadam
reference_source: "db/seeds/master_seed.sql application_recipes (post Phase 208 WS0 audit)"
source_tier: cho_named
---

# Natural farming application recipes (canon)

## What it is (1 paragraph)

Reference table for all **14 application recipes** in farm seed data — dilutions audited against Cho 2016 / KNF standards in Phase 208 WS0. Use with input batch inventory and fertigation programs.

## When to use

Look up how to apply a ready batch before mixing a tank or spraying. Cross-link input guides for making each concentrate.

## Ingredients (list with amounts)

See input guides per component — this table is **application** only.

## Step-by-step preparation

1. Identify recipe name below matching your program or task.
2. Strain concentrates as needed (JLF, JHS).
3. Mix at listed dilution in non-chlorinated water same day.
4. Add JWA when recipe notes coverage.

## Ferment / wait timeline

N/A — application step; batches must already be at ready status.

## Ready signs (smell, foam, color)

Input batches at **ready_for_use** per batch notes before applying these dilutions.

## Storage

Mixed tank: use same day. Do not store diluted spray overnight.

## Safety & water (non-chlorinated, PPE)

Follow each input''s safety tier — OHN and JS never above labeled dilution.

## How to apply (link to application recipe name)

| # | Recipe | Type | Dilution (canon) | Frequency |
|---|--------|------|------------------|-----------|
| 1 | JMS Soil Drench | soil_drench | **1:10** JMS:water | Every 2 weeks |
| 2 | JLF General Soil Drench | soil_drench | 1:20 (start **1:100**) | Weekly–biweekly |
| 3 | JLF Seedling Drench | soil_drench | **1:30** | Weekly seedlings |
| 4 | JLF and JMS Combined Drench | soil_drench | JLF 1:20 + JMS **1:10** | Weekly peak season |
| 5 | LAB Soil Conditioner | soil_drench | **1:1000** | Every 2–4 weeks |
| 6 | OHN Pest and Immunity Drench | soil_drench | **1:1000** max | Preventative / pressure |
| 7 | JMS Foliar Spray | foliar_spray | **1:20** + JWA | Every 1–2 weeks |
| 8 | FPJ Vegetative Foliar | foliar_spray | 1:500–1:1000 + JWA | Veg only |
| 9 | FFJ and WCA Flowering Boost | foliar_spray | FFJ 1:500 + WCA 1:1000 + JWA | Flower → early fruit |
| 10 | BRV and WCA Cell Strengthener | foliar_spray | BRV 1:800 + WCA 1:1000 | Before stress |
| 11 | JHS and JWA Natural Pesticide | foliar_spray | JHS 1:50 + JWA 1:500 | Weekly preventative |
| 12 | JS Fungicide Spray | foliar_spray | **0.5–2 L JS conc. / 500 L** + JWA | ≤32 °C; repeat 5–7 d |
| 13 | JLF Foliar Feed | foliar_spray | 1:30–1:50 + JWA | Stress only |
| 14 | JWA Insecticide Spray | foliar_spray | **1:500** JWA | Active soft-bodied pests |

## Dilution table (start conservative → stronger)

JLF drenches: always **start 1:100** if unsure (see [JLF general](natural-farming-jlf-general.md)). JMS never weaker than **1:10** soil / **1:20** foliar per audit.

## Common mistakes

- Using pre-audit **1:500 JMS** — wrong (WS0 fixed)
- OHN or FAA stronger than 1:1000
- JS as 0.5% wettable sulfur — wrong input; use JADAM JS concentrate recipe
$nf_natural-farming-application-recipes$, 5, TRUE, 100),
    ('natural-farming-brv', 'Brown rice vinegar (BRV) — purchased input', NULL, 'natural_farming', 'natural_farming', 'safe', $nf_natural-farming-brv$---
domain: natural_farming
title: Brown rice vinegar (BRV) — purchased input
safety_tier: safe
tradition: other
reference_source: "JADAM Organic Farming, Youngsang Cho, 2016"
source_tier: third_party
---

# Brown rice vinegar (BRV) — purchased input

## What it is (1 paragraph)

Unpasteurized organic brown rice vinegar (4–8% acidity) — purchased input for WCA/WCS extraction and foliar cell-strengthening with WCA. Not a ferment you make on farm in v1 seed data.

## When to use

- WCA and WCS extraction solvent (1:10 with shells/bones)
- **BRV and WCA Cell Strengthener** foliar before stress

## Ingredients (list with amounts)

- Organic unpasteurized BRV — purchase ready-made

## Step-by-step preparation

1. Purchase unpasteurized organic BRV.
2. Use directly for extracts or dilute per application recipe.

## Ferment / wait timeline

N/A — purchased product.

## Ready signs (smell, foam, color)

Clear amber vinegar; live culture sediment normal in unpasteurized bottles.

## Storage

Sealed at room temperature indefinitely.

## Safety & water (non-chlorinated, PPE)

Undiluted on foliage burns — always follow recipe dilution.

## How to apply (link to application recipe name)

**BRV and WCA Cell Strengthener** — BRV **1:800** + WCA **1:1000**; do not exceed BRV rate.

## Dilution table (start conservative → stronger)

| Recipe | BRV dilution |
|--------|--------------|
| Cell strengthener foliar | **1:800** with WCA 1:1000 |

## Common mistakes

- Pasteurized clear vinegar for WCA — weak extraction
- Full-strength foliar — leaf burn
$nf_natural-farming-brv$, 5, TRUE, 101),
    ('natural-farming-compost-tea-aact', 'Actively aerated compost tea (AACT)', NULL, 'natural_farming', 'natural_farming', 'safe', $nf_natural-farming-compost-tea-aact$---
domain: natural_farming
title: Actively aerated compost tea (AACT)
safety_tier: safe
tradition: other
reference_source: "Elaine Ingham, Soil Biology Primer"
source_tier: third_party
---

# Actively aerated compost tea (AACT)

## What it is (1 paragraph)

AACT is **not JADAM** — an Elaine Ingham–style aerobic compost extract brewed with air stone and molasses to multiply beneficial microbes. Complements JMS but follows a **4-hour use window** after brew finishes.

## When to use

- Soil drench or foliar when soil food web boost is needed
- Disease suppression support alongside good compost source

## Ingredients (list with amounts)

- Finished quality compost (mesh bag)
- Unsulfured molasses — ~1 tbsp per 4 L water
- Optional kelp meal
- De-chlorinated water

## Step-by-step preparation

1. Suspend compost bag in bucket with air stone running.
2. Add molasses (and kelp if used).
3. Brew **24–48 h** with continuous aeration.
4. Use entire batch within **4 hours** of stopping aeration.

## Ferment / wait timeline

Brew **24–48 h** aerated; apply within **4 h** of finish.

## Ready signs (smell, foam, color)

Earthy smell; slight foam; no anaerobic rotten odor.

## Storage

**Do not store** brewed tea — microbes crash without O₂.

## Safety & water (non-chlorinated, PPE)

Use finished mature compost; E. coli risk if compost is immature or aeration fails.

## How to apply (link to application recipe name)

Apply as soil drench or foliar using standard compost-tea dilution for your volume (undiluted to 1:10 depending on compost strength — start weak).

## Dilution table (start conservative → stronger)

| Pass | Guidance |
|------|----------|
| First use | Weak tea / longer water ratio — watch plant response |
| Follow-up | Stronger only if no burn and good compost source |

## Common mistakes

- Storing tea overnight — anaerobic crash
- Turning off air early — pathogen bloom risk
- Calling it JADAM JMS substitute — different tradition and timing
$nf_natural-farming-compost-tea-aact$, 5, TRUE, 102),
    ('natural-farming-faa', 'Fish amino acid (FAA) — KNF', NULL, 'natural_farming', 'natural_farming', 'caution', $nf_natural-farming-faa$---
domain: natural_farming
title: Fish amino acid (FAA) — KNF
safety_tier: caution
tradition: knf
reference_source: "KNF (Cho Han-kyu); often used with JADAM"
source_tier: knf_standard
---

# Fish amino acid (FAA) — KNF

## What it is (1 paragraph)

FAA is **KNF** fish scrap ferment with brown sugar — high nitrogen and trace minerals. Long ferment until bones dissolve; apply only at high dilution.

## When to use

- Supplemental nitrogen foliar or soil at **1:1000** minimum
- Paired with JADAM programs when fish waste is available

## Ingredients (list with amounts)

- Fresh fish scraps (no salt)
- Brown sugar **1:1 by weight**

## Step-by-step preparation

1. Layer fish and brown sugar 1:1 in ferment vessel.
2. Cover breathable; ferment **3–6 months** until bones soften/dissolve.
3. Strain; bottle.

## Ferment / wait timeline

**3–6 months**; longer in cool climates.

## Ready signs (smell, foam, color)

Fish breaks down; bones crumble; sauce-like liquid.

## Storage

Refrigerate after strain; **6–12 months**.

## Safety & water (non-chlorinated, PPE)

Strong odor — ventilate outdoor ferment. Salted fish scraps ruin batch.

## How to apply (link to application recipe name)

Soil or foliar at **≥1:1000** FAA:water (KNF standard — never stronger).

## Dilution table (start conservative → stronger)

| Use | Dilution |
|-----|----------|
| All passes | **1:1000 minimum** |

## Common mistakes

- Short ferment — incomplete breakdown
- Strong smell indoors without ventilation
- Undiluted application — salt/ammonia burn
$nf_natural-farming-faa$, 5, TRUE, 103),
    ('natural-farming-ffj', 'Fermented Fruit Juice (FFJ) — KNF', NULL, 'natural_farming', 'natural_farming', 'safe', $nf_natural-farming-ffj$---
domain: natural_farming
title: Fermented Fruit Juice (FFJ) — KNF
safety_tier: safe
tradition: knf
reference_source: "KNF (Cho Han-kyu); often used with JADAM"
source_tier: knf_standard
---

# Fermented Fruit Juice (FFJ) — KNF

## What it is (1 paragraph)

FFJ is a **KNF** sugar ferment of ripe fruit — enzymes, sugars, and potassium for flowering and early fruit. Often paired with JADAM programs but it is sugar-based KNF, not JADAM core.

## When to use

- Transition to flowering through early fruit set
- With WCA in **FFJ and WCA Flowering Boost** foliar program

## Ingredients (list with amounts)

- Ripe or overripe fruit (banana peels work well)
- Brown sugar — **1:1 by weight** with fruit
- **No added water** (KNF standard)

## Step-by-step preparation

1. Chop fruit; mix 1:1 with brown sugar by weight.
2. Pack in jar; cover with breathable cloth (not airtight).
3. Ferment ~7 days at room temperature.
4. Strain liquid; bottle.

## Ferment / wait timeline

~**7 days** ferment; strain when juice separates.

## Ready signs (smell, foam, color)

Sweet-sour ferment smell; liquid syrup; fruit collapsed.

## Storage

Refrigerate after straining; use within **6 months**.

## Safety & water (non-chlorinated, PPE)

Do not add water. Avoid moldy fruit.

## How to apply (link to application recipe name)

**FFJ and WCA Flowering Boost** — FFJ 1:500 + WCA 1:1000 + JWA in same tank, weekly from first buds.

## Dilution table (start conservative → stronger)

| Use | Dilution | Notes |
|-----|----------|-------|
| Flowering foliar | 1:500 FFJ | With WCA 1:1000 |
| Hot weather | 1:800–1:1000 | Lighter pass |

## Common mistakes

- Adding water to ferment — not KNF FFJ
- Using in heavy veg — promotes wrong growth stage
- Labeling as pure JADAM — honesty: KNF input
$nf_natural-farming-ffj$, 5, TRUE, 104),
    ('natural-farming-forest-garden-understory', 'Forest garden understory (cherry, blackberry, goldenrod)', NULL, 'natural_farming', 'natural_farming', 'safe', $nf_natural-farming-forest-garden-understory$---
domain: natural_farming
title: Forest garden understory (cherry, blackberry, goldenrod)
safety_tier: safe
tradition: extension
reference_source: "Operator horticulture + gr33n crop guides; no invented EC targets"
source_tier: extension_method
---

# Forest garden understory (cherry, blackberry, goldenrod)

## What it is (1 paragraph)

Counsel for **forest-garden / orchard understory** questions — cherry with goldenrod, blackberries, and mixed volunteers. gr33n does **not** ship EC/VPD programs for wild understory polycultures; this guide gives honest ecology and links to natural-farming extensions without inventing bottle-nutrient schedules.

## When to use

- Operator asks about cherry + goldenrod + blackberry coexistence
- Guardian smoke/regression forest-garden prompts with **farm context on** after RAG ingest

## Ingredients (list with amounts)

N/A — ecology guide, not a ferment recipe.

## Step-by-step preparation

1. Identify operator goals: fruit quality, dye harvest, blackberry keep/remove, goldenrod management.
2. For goldenrod biomass → [goldenrod JLF extension](natural-farming-goldenrod-jlf.md) at **1:100** start.
3. For sweet cherry production targets see [cherry nursery guide](crop-cherry-nursery.md) — not identical to backyard forest garden.
4. For unsupported woodland forage crops see [crop-unsupported-woodland.md](crop-unsupported-woodland.md).

## Ferment / wait timeline

If using goldenrod JLF — see extension guide ferment timeline.

## Ready signs (smell, foam, color)

N/A for ecology; for JLF extension see goldenrod guide.

## Storage

N/A.

## Safety & water (non-chlorinated, PPE)

Blackberry thorns — PPE for clearing. Do not recommend herbicide blanket on polyculture without operator consent.

## How to apply (link to application recipe name)

Optional understory fertility: **JLF General Soil Drench** via goldenrod extension — **1:100** conservative around cherry root zone; never claim EC match to indoor veg programs.

## Dilution table (start conservative → stronger)

| Material | Suggestion |
|----------|------------|
| Goldenrod → JLF | Start **1:100** drench |
| Blackberry | Management choice — not a fertigation recipe |

## Common mistakes

- Inventing EC 1.8 veg feed for forest garden cherry
- Telling operator they must eradicate goldenrod — operator may keep for dye + JLF biomass
- Confusing nursery cherry production guide with backyard understory
$nf_natural-farming-forest-garden-understory$, 5, TRUE, 105),
    ('natural-farming-fpj', 'Fermented Plant Juice (FPJ) — KNF', NULL, 'natural_farming', 'natural_farming', 'safe', $nf_natural-farming-fpj$---
domain: natural_farming
title: Fermented Plant Juice (FPJ) — KNF
safety_tier: safe
tradition: knf
reference_source: "KNF (Cho Han-kyu); often used with JADAM"
source_tier: knf_standard
---

# Fermented Plant Juice (FPJ) — KNF

## What it is (1 paragraph)

FPJ is **KNF** fermented growing tips (comfrey, nettle, mugwort, bamboo) with brown sugar — plant hormones and amino acids for vegetative growth. Sugar-based; label as KNF when used beside JADAM.

## When to use

- Vegetative stage only — **stop at flower transition**
- **FPJ Vegetative Foliar** every 7–14 days

## Ingredients (list with amounts)

- Fresh fast-growing plant tips
- Brown sugar **1:1 by weight**

## Step-by-step preparation

1. Chop tips; layer equal weight sugar.
2. Seal jar with breathable cover.
3. Ferment 3–7 days; strain and bottle.

## Ferment / wait timeline

**3–7 days** at room temp.

## Ready signs (smell, foam, color)

Sweet ferment; liquid separates; tips collapsed.

## Storage

Refrigerate; **6–12 months** sealed.

## Safety & water (non-chlorinated, PPE)

Accurate 1:1 sugar ratio; no moldy material.

## How to apply (link to application recipe name)

**FPJ Vegetative Foliar** — 1:500 (1:1000 hot weather) + JWA 1:1000.

## Dilution table (start conservative → stronger)

| Conditions | FPJ:water |
|------------|-----------|
| Normal | 1:500 |
| Hot / stress | 1:1000 |

## Common mistakes

- Continuing FPJ after flowers — wrong stage
- Calling it JADAM core — it is KNF (sugar ferment)
$nf_natural-farming-fpj$, 5, TRUE, 106),
    ('natural-farming-goldenrod-jlf', 'Goldenrod JLF (extension method)', NULL, 'natural_farming', 'natural_farming', 'safe', $nf_natural-farming-goldenrod-jlf$---
domain: natural_farming
title: Goldenrod JLF (extension method)
safety_tier: safe
tradition: extension
reference_source: "JLF General method — JADAM Organic Farming, Youngsang Cho, 2016"
source_tier: extension_method
---

# Goldenrod JLF (extension method)

## What it is (1 paragraph)

**Not a named Cho goldenrod recipe.** Valid **extension**: apply the standard [JLF general](natural-farming-jlf-general.md) method to **Solidago** (Canadian goldenrod) biomass — dynamic accumulator weed usable for orchard understory fertility. Compatible with dye harvest if biomass is collected responsibly.

## When to use

- Orchard understory (cherry, apple, plum) where goldenrod is present
- Operator keeps goldenrod for dyes **and** wants fertigation use — not prescriptive removal

## Ingredients (list with amounts)

- Fresh goldenrod biomass (stems/leaves — not necessarily flowers if reserved for dye)
- Leaf mold handful
- Non-chlorinated water — same ratios as JLF general (2/3 vessel biomass)

## Step-by-step preparation

Follow [JLF general](natural-farming-jlf-general.md) steps 1–5 using goldenrod as the weed feedstock.

## Ferment / wait timeline

**7–14 days** ferment; strain before use.

## Ready signs (smell, foam, color)

Earthy ferment — same ready signs as general JLF.

## Storage

Strained: **30 days** cool/shaded.

## Safety & water (non-chlorinated, PPE)

Do not use sprayed roadside plants. Label batch **goldenrod JLF extension**.

## How to apply (link to application recipe name)

**JLF General Soil Drench** dilution bands — start **1:100**, stronger only after plant response (up to **1:30** experienced).

Guardian must say: *"JLF from goldenrod using the standard JLF method"* — never *"Cho''s goldenrod recipe."*

## Dilution table (start conservative → stronger)

| Pass | Dilution | Notes |
|------|----------|-------|
| First | **1:100** | Cherry understory conservative start |
| Tested OK | 1:30–1:20 | Only with observed plant response |

## Common mistakes

- Claiming cho_named source tier — must stay **extension_method**
- Starting at 1:20 on unknown understory — too strong
- Conflicting with dye harvest — plan biomass cuts so both uses fit operator intent

See also [forest garden understory](natural-farming-forest-garden-understory.md).
$nf_natural-farming-goldenrod-jlf$, 5, TRUE, 107),
    ('natural-farming-indoor-photoperiod-program', 'Indoor photoperiod JADAM programs (bootstrap)', NULL, 'natural_farming', 'natural_farming', 'safe', $nf_natural-farming-indoor-photoperiod-program$---
domain: natural_farming
title: Indoor photoperiod JADAM programs (bootstrap)
safety_tier: safe
tradition: jadam
reference_source: "gr33ncore._bootstrap_jadam_indoor_photoperiod_v1"
source_tier: cho_named
---

# Indoor photoperiod JADAM programs (bootstrap)

## What it is (1 paragraph)

Maps the **`jadam_indoor_photoperiod_v1`** bootstrap template — veg (18/6), flower (12/12), and outdoor JLF drench programs wired to audited application recipes on demo and new farms.

## When to use

After applying bootstrap template or when mirroring demo farm fertigation layout. Commercial EC programs stay on **Feed & water** — these are parallel natural paths.

## Ingredients (list with amounts)

Program-linked batches (template IDs **TPL-JLF-GEN-001**, **TPL-JMS-001**, **TPL-FFJ-001**, **TPL-WCA-001**) — see input guides.

## Step-by-step preparation

1. Apply bootstrap `jadam_indoor_photoperiod_v1` or use demo farm seed.
2. Confirm batches exist for JLF, JMS, FFJ, WCA.
3. Refresh reservoir mix tasks before scheduled irrigations.

## Ferment / wait timeline

Maintain rolling JMS at peak foam; JLF/FFJ batches per input guide storage windows.

## Ready signs (smell, foam, color)

Batch status **ready_for_use** in inventory before linking to a mix event.

## Storage

Veg **Main Nutrient Reservoir** and flower **Flower Nutrient Reservoir** — mix same day as irrigation schedule.

## Safety & water (non-chlorinated, PPE)

Non-chlorinated make-up water in reservoirs. JMS at **1:10** with JLF **1:20** in combined veg tank (not legacy 1:500 JMS).

## How to apply (link to application recipe name)

| Program | Zone | Recipe | Schedule |
|---------|------|--------|----------|
| **Veg Daily JLF Program** | Veg Room 18/6 | **JLF and JMS Combined Drench** | Water Late Veg Daily |
| **Flower Daily FFJ+WCA Program** | Flower Room 12/12 | **FFJ and WCA Flowering Boost** | Water Early Flower Daily |
| **Outdoor JLF Soil Drench** | Outdoor Garden | **JLF General Soil Drench** | Water Outdoor Garden Daily |

See [application recipes](natural-farming-application-recipes.md) for dilutions.

## Dilution table (start conservative → stronger)

Combined veg tank: JLF **1:20** + JMS **1:10**. Flower tank: FFJ 1:500 + WCA 1:1000. Outdoor: JLF start **1:100** if unsure.

## Common mistakes

- Expecting EC 1.6–1.8 mS/cm Mericle targets without converting mindset — tune volume/cron, not bottle A/B
- Legacy JMS 1:500 in mix notes — pre-audit; refresh to 1:10
$nf_natural-farming-indoor-photoperiod-program$, 5, TRUE, 108),
    ('natural-farming-jlf-crop-specific', 'JLF from same-crop residue', NULL, 'natural_farming', 'natural_farming', 'safe', $nf_natural-farming-jlf-crop-specific$---
domain: natural_farming
title: JLF from same-crop residue
safety_tier: safe
tradition: jadam
reference_source: "JADAM Organic Farming, Youngsang Cho, 2016"
source_tier: cho_named
---

# JLF from same-crop residue

## What it is (1 paragraph)

Crop-specific JLF uses residue from the **same crop** you will feed — the most targeted JADAM fertilizer (tomato trimmings for tomatoes, corn stalks for corn).

## When to use

- Recurring fertility for a single crop through a season
- When you have clean, healthy residue volume

## Ingredients (list with amounts)

- Same-crop residue (stems, leaves — not fruit or seed) — 2/3 vessel
- Leaf mold handful
- Non-chlorinated water to top

## Step-by-step preparation

1. Chop residue small; fill container 2/3.
2. Add leaf mold; fill with water; seal.
3. Ferment **10–14 days**; stir occasionally.
4. Strain before use; label crop + date.

## Ferment / wait timeline

**10–14 days** minimum; use within same growing season.

## Ready signs (smell, foam, color)

Earthy ferment smell; residue broken down; brown liquid.

## Storage

Use within season; label crop type. Do not mix crop-specific batches blindly.

## Safety & water (non-chlorinated, PPE)

**Never** use diseased plant material. Non-chlorinated water.

## How to apply (link to application recipe name)

Apply via **JLF General Soil Drench** dilution bands (start **1:100**, up to **1:20** when tested).

## Dilution table (start conservative → stronger)

| Stage | Dilution | Notes |
|-------|----------|-------|
| Start | 1:100 | Conservative first pass |
| Established | 1:20–1:30 | Match crop vigor |

## Common mistakes

- Cross-crop residue — loses targeted benefit
- Diseased trimmings — spreads pathogens
- Over-applying on fruiting plants — too much N
$nf_natural-farming-jlf-crop-specific$, 5, TRUE, 109),
    ('natural-farming-jlf-general', 'JLF from weeds and grass (general)', NULL, 'natural_farming', 'natural_farming', 'safe', $nf_natural-farming-jlf-general$---
domain: natural_farming
title: JLF from weeds and grass (general)
safety_tier: safe
tradition: jadam
reference_source: "JADAM Organic Farming, Youngsang Cho, 2016"
source_tier: cho_named
---

# JLF from weeds and grass (general)

## What it is (1 paragraph)

JLF (JADAM Liquid Fertilizer) from local weeds and grasses returns native minerals to soil. It is the primary fertility input in JADAM programs — much stronger than JMS; dilute carefully.

## When to use

- Primary soil fertility weekly to biweekly in active growth
- Outdoor beds and indoor reservoirs on JADAM-style programs
- Base method for extension materials (e.g. goldenrod) — see [goldenrod JLF](natural-farming-goldenrod-jlf.md)

## Ingredients (list with amounts)

- Fresh untreated weeds/grass clippings — fill container **2/3**
- Handful leaf mold (microbial starter)
- Non-chlorinated water to top

## Step-by-step preparation

1. Chop weeds; fill ferment vessel 2/3 full.
2. Add leaf mold starter.
3. Top with non-chlorinated water; seal (burp if needed).
4. Ferment 7–14 days; stir every few days.
5. Strain through cloth before use.

## Ferment / wait timeline

- Minimum usable **7–14 days**; can mature weeks to months for richer brew
- Strain before applying

## Ready signs (smell, foam, color)

- Earthy fermented smell — not rotten-egg anaerobic
- Plant material softened; liquid amber to brown

## Storage

Strained: use within **30 days**. Sealed concentrate up to **3 months** cool/shaded.

## Safety & water (non-chlorinated, PPE)

No herbicide-treated clippings. Non-chlorinated water only.

## How to apply (link to application recipe name)

- **JLF General Soil Drench** — main fertility
- **JLF Seedling Drench** — gentler 1:30
- **JLF and JMS Combined Drench** — with JMS 1:10
- **JLF Foliar Feed** — stress only, finely strained

## Dilution table (start conservative → stronger)

| Situation | JLF:water | Notes |
|-----------|-----------|-------|
| First time / unsure | **1:100** | Start here per Cho/FigJam |
| Tested on your soil | **1:20** | Primary experienced default |
| Seedlings | **1:30** | See JLF Seedling Drench |

## Common mistakes

- Starting at 1:20 on unknown soil — leaf burn or salt shock
- Using diseased or sprayed weeds — pathogen carryover
- Foliar without fine strain — clogged sprayer
$nf_natural-farming-jlf-general$, 5, TRUE, 110),
    ('natural-farming-jlf-spring-nettle-comfrey', 'Spring JLF — nettle and comfrey', NULL, 'natural_farming', 'natural_farming', 'caution', $nf_natural-farming-jlf-spring-nettle-comfrey$---
domain: natural_farming
title: Spring JLF — nettle and comfrey
safety_tier: caution
tradition: jadam
reference_source: "JADAM Organic Farming, Youngsang Cho, 2016"
source_tier: cho_named
---

# Spring JLF — nettle and comfrey

## What it is (1 paragraph)

High-nitrogen spring JLF from **dynamic accumulator** biomass (nettle, comfrey) — deep-mining herbs, not nitrogen-fixing legumes. Strong vegetative push.

## When to use

- Early spring vegetative growth
- Before heavy fruiting — avoid over-N on fruiting plants later

## Ingredients (list with amounts)

- Fresh stinging nettle tops and/or comfrey leaves — 2/3 vessel
- Leaf mold handful
- Non-chlorinated water

## Step-by-step preparation

1. Harvest nettle wearing gloves; chop with comfrey.
2. Fill 2/3; add leaf mold; top with water; seal.
3. Ferment **7–10 days**; strain.

## Ferment / wait timeline

**7–10 days**; use strained liquid within **2 weeks**.

## Ready signs (smell, foam, color)

Rich earthy smell; dark green-brown liquid after strain.

## Storage

Use within 2 weeks of straining — high N degrades in storage.

## Safety & water (non-chlorinated, PPE)

Gloves for nettle harvest. High N — do not over-apply to fruiting crops.

## How to apply (link to application recipe name)

**JLF General Soil Drench** — start **1:100**, test before **1:20**.

## Dilution table (start conservative → stronger)

| Pass | Dilution | Notes |
|------|----------|-------|
| First | 1:100 | Spring push, watch leaf color |
| Follow-up | 1:30–1:20 | Only if plants respond well |

## Common mistakes

- Calling nettle "nitrogen-fixing" — it is a dynamic accumulator
- Strong dilution on fruit trees in bloom — vegetative push at wrong time
$nf_natural-farming-jlf-spring-nettle-comfrey$, 5, TRUE, 111),
    ('natural-farming-jms', 'JADAM Microbial Solution (JMS)', NULL, 'natural_farming', 'natural_farming', 'safe', $nf_natural-farming-jms$---
domain: natural_farming
title: JADAM Microbial Solution (JMS)
safety_tier: safe
tradition: jadam
reference_source: "JADAM Organic Farming, Youngsang Cho, 2016"
source_tier: cho_named
---

# JADAM Microbial Solution (JMS)

## What it is (1 paragraph)

JMS is the foundation JADAM microbial inoculant — diverse bacteria and fungi from forest leaf mold, activated with potato starch water. It builds soil and leaf-surface biology and suppresses pathogens when used at Cho dilutions.

## When to use

- Soil drenches every 2 weeks in active season; before transplant
- Foliar sprays in vegetative and early flower for leaf-surface microbes
- Combined with JLF in peak-season weekly drenches

## Ingredients (list with amounts)

- Leaf mold humus (local forest floor), ~1 cup per 10–20 L batch
- 1 potato boiled in non-chlorinated water; use cooled potato water as base
- Pinch of sea salt
- Non-chlorinated water to 10–20 L total volume

## Step-by-step preparation

1. Boil potato in non-chlorinated water; cool completely.
2. Place potato in a mesh bag; suspend in a bucket with leaf mold and pinch of salt.
3. Fill to 10–20 L with non-chlorinated water; cover loosely (not airtight).
4. Ferment 24–72 h at 20–30 °C until peak foam activity.
5. Use at peak — strain if needed for sprayers.

## Ferment / wait timeline

- **24–72 h** active fermentation to peak foam
- **Use within 6–12 h of peak** — not after sitting a full week idle

## Ready signs (smell, foam, color)

- Vigorous foam at surface at peak activity
- Earthy, not putrid, smell
- Cloudy water; potato breaks down in bag

## Storage

Use at peak foam; do not store active JMS long-term. Make fresh batches weekly during season.

## Safety & water (non-chlorinated, PPE)

Chlorinated tap water kills microbes — rain, RO, or de-chlorinated water only.

## How to apply (link to application recipe name)

- **JMS Soil Drench** — primary soil inoculant
- **JMS Foliar Spray** — leaf biology (+ JWA for coverage)
- **JLF and JMS Combined Drench** — weekly peak-season pass

## Dilution table (start conservative → stronger)

| Use | Dilution (JMS:water) | Notes |
|-----|----------------------|-------|
| Soil drench | **1:10** | 2–4 L per sqm root zone |
| Foliar | **1:20** + JWA | Early morning; both leaf sides |
| With JLF tank | **1:10** in same water as JLF 1:20 | Apply same day |

## Common mistakes

- Storing finished JMS a week+ after peak — weak or anaerobic
- Using 1:500 dilution (old drift) — far too weak per Cho
- Skipping JWA on foliar — poor leaf coverage
$nf_natural-farming-jms$, 5, TRUE, 112),
    ('natural-farming-jwa-js-jhs', 'JWA, JHS, and JS — wetting and pest inputs (JADAM)', NULL, 'natural_farming', 'natural_farming', 'expert', $nf_natural-farming-jwa-js-jhs$---
domain: natural_farming
title: JWA, JHS, and JS — wetting and pest inputs (JADAM)
safety_tier: expert
tradition: jadam
reference_source: "JADAM Organic Farming, Youngsang Cho, 2016"
source_tier: cho_named
---

# JWA, JHS, and JS — wetting and pest inputs (JADAM)

## What it is (1 paragraph)

**JWA** is JADAM wetting-agent soap (wood-ash lye + oil). **JHS** is boiled herbal concentrate for pest deterrent sprays. **JS** is exothermic **JADAM sulfur concentrate** (~25% sulfur) — not garden wettable sulfur. JWA is added to foliar mixes for coverage; JHS and JS are pest/disease programs at seed dilutions.

## When to use

- JWA: surfactant in JMS foliar, JHS/JS sprays, FFJ/FPJ tanks
- JHS + JWA: weekly preventative pest spray
- JS + JWA: powdery mildew, rust, mites at first sign (≤32 °C)

## Ingredients (list with amounts)

**JWA:** wood ash lye water + plant oil (soy/canola/coconut) → soap

**JHS:** 1 kg fresh herb (wormwood, artemisia, garlic chives, hot pepper, neem, or Jerusalem artichoke) + 4–5 L water

**JS concentrate:** elemental sulfur, caustic soda (NaOH), red clay, phyllite, sea salt — Cho exothermic batch (~25% sulfur concentrate)

## Step-by-step preparation

**JWA:** boil ash for lye water; filter; mix 1:1 with oil; boil to soap.

**JHS:** boil 1 kg plant in mesh bag in 4–5 L water **4–5 hours**; strain very fine.

**JS:** follow Cho exothermic batch method; label concentrate; dilute only at spray time.

## Ferment / wait timeline

JWA soap keeps dry indefinitely. JHS use within **2 weeks** refrigerated. JS concentrate stored sealed; mix spray **same day**.

## Ready signs (smell, foam, color)

JWA: firm soap paste. JHS: dark aromatic broth, fine strain required. JS: labeled concentrate strength.

## Storage

JWA dry soap; JHS refrigerated short term; JS concentrate sealed labeled.

## Safety & water (non-chlorinated, PPE)

**JWA:** lye caustic — gloves, no sun spray burns.

**JS:** caustic soda batch — full PPE, ventilation; **do not apply above 32 °C**.

**JHS:** avoid open blooms — deters pollinators.

## How to apply (link to application recipe name)

- **JMS Foliar Spray** — add JWA for coverage
- **JHS and JWA Natural Pesticide** — JHS 1:50 + JWA 1:500
- **JS Fungicide Spray** — **0.5–2 L JS concentrate per 500 L water** + JWA 1:500
- **JWA Insecticide Spray** — JWA 1:500 alone for soft-bodied insects

## Dilution table (start conservative → stronger)

| Product | Application dilution |
|---------|---------------------|
| JWA alone | 1:500 |
| JHS + JWA | 1:50 + 1:500 |
| JS concentrate | 0.5–2 L per 500 L water + JWA |

## Common mistakes

- Wettable sulfur labeled JS — wrong input (pre-audit drift)
- JHS cold steep 1–3 h — not Cho boil method
- JS spray in hot midday sun — sulfur burn
$nf_natural-farming-jwa-js-jhs$, 5, TRUE, 113),
    ('natural-farming-lab', 'Lactic Acid Bacteria serum (LAB) — KNF', NULL, 'natural_farming', 'natural_farming', 'safe', $nf_natural-farming-lab$---
domain: natural_farming
title: Lactic Acid Bacteria serum (LAB) — KNF
safety_tier: safe
tradition: knf
reference_source: "KNF (Cho Han-kyu); often used with JADAM"
source_tier: knf_standard
---

# Lactic Acid Bacteria serum (LAB) — KNF

## What it is (1 paragraph)

LAB serum from soured rice wash cultured in milk — golden layer of lactic acid bacteria that suppresses harmful soil microbes and improves structure. KNF input, often paired with JADAM.

## When to use

- Soil conditioning every 2–4 weeks
- Before transplanting
- **LAB Soil Conditioner** drench

## Ingredients (list with amounts)

- Rice wash (first rinse water)
- Fresh whole milk (non-UHT) — 10 parts milk to 1 part soured rice wash

## Step-by-step preparation

1. Ferment rice wash 3–5 days until soured.
2. Mix 1 part soured rice wash into 10 parts milk.
3. Wait 5–7 days; collect **golden serum** from bottom.
4. Mix equal part raw sugar to preserve (optional).

## Ferment / wait timeline

Rice wash **3–5 d**; milk culture **5–7 d**.

## Ready signs (smell, foam, color)

Golden translucent serum layer below curds; sour-clean smell.

## Storage

Refrigerated with sugar preservative **6–12 months**. Use serum only — discard curds and white top.

## Safety & water (non-chlorinated, PPE)

Use golden layer only.

## How to apply (link to application recipe name)

**LAB Soil Conditioner** — **1:1000** LAB:water; water in lightly.

## Dilution table (start conservative → stronger)

| Use | Dilution |
|-----|----------|
| Soil drench | **1:1000** |

## Common mistakes

- Using curds or top milk layer — wrong fraction
- Stronger than 1:1000 — unnecessary; KNF standard is dilute
$nf_natural-farming-lab$, 5, TRUE, 114),
    ('natural-farming-livestock-plant-feed', 'Livestock plant feed (simple inputs)', NULL, 'natural_farming', 'natural_farming', 'safe', $nf_natural-farming-livestock-plant-feed$---
domain: natural_farming
title: Livestock plant feed (simple inputs)
safety_tier: safe
tradition: extension
reference_source: "Operator practice; gr33n animal_feed category — not full ration math"
source_tier: extension_method
---

# Livestock plant feed (simple inputs)

## What it is (1 paragraph)

Simple on-farm **animal_feed** inputs — comfrey slurry, sprouted grain, chop-and-drop — tracked in `gr33nnaturalfarming` **animal_feed** category. **Not** total mixed ration (TMR) balancing or veterinary formulation.

## When to use

- Chickens, goats, or other livestock with on-farm plant feed supplements
- Linking comfrey or grain sprouts to inventory batches (see demo chicken bootstrap for flock context)

## Ingredients (list with amounts)

**Comfrey slurry:** fresh comfrey leaves + water — wilt/blend to slurry (operator volume by flock size)

**Sprouted grain:** grain soak 8–12 h, drain, sprout 2–5 days until short tails

## Step-by-step preparation

**Comfrey:** harvest comfrey; chop; soak or blend with water; feed fresh within 24 h.

**Sprouts:** rinse daily; feed when sprout tail appears; discard moldy trays.

## Ferment / wait timeline

Comfrey slurry: use **same day**. Sprouts: **2–5 days** from soak to feed-ready.

## Ready signs (smell, foam, color)

Sprouts: white root tails, no mold. Comfrey: fresh green smell.

## Storage

Do not store comfrey slurry long — anaerobic spoilage. Sprouts refrigerated max 1–2 days.

## Safety & water (non-chlorinated, PPE)

Comfrey contains pyrrolizidine alkaloids — **moderation** for poultry; not sole diet. Moldy sprouts — discard.

## How to apply (link to application recipe name)

Record as **animal_feed** input batch in Natural farming inventory — not fertigation application recipes.

## Dilution table (start conservative → stronger)

| Feed | Guidance |
|------|----------|
| Comfrey | Small supplement — not majority of ration |
| Sprouts | Treat as treat/supplement with balanced grain/forage |

## Common mistakes

- Using this guide as complete ration math — out of scope v1
- Feeding comfrey as unlimited primary forage
- Confusing with JLF comfrey ferment for plants — different use path

Cross-link [JLF spring nettle/comfrey](natural-farming-jlf-spring-nettle-comfrey.md) for **plant** fertility, not livestock ration.
$nf_natural-farming-livestock-plant-feed$, 5, TRUE, 115),
    ('natural-farming-ohn', 'Oriental Herbal Nutrient (OHN) — KNF', NULL, 'natural_farming', 'natural_farming', 'caution', $nf_natural-farming-ohn$---
domain: natural_farming
title: Oriental Herbal Nutrient (OHN) — KNF
safety_tier: caution
tradition: knf
reference_source: "KNF (Cho Han-kyu); often used with JADAM"
source_tier: knf_standard
---

# Oriental Herbal Nutrient (OHN) — KNF

## What it is (1 paragraph)

OHN is a potent **KNF** extract of garlic, ginger, angelica, cinnamon and other aromatics — immune support and pest deterrent in **very small** doses.

## When to use

- Preventative soil drench every 2–4 weeks
- Pest pressure — weekly at labeled dilution only

## Ingredients (list with amounts)

- Garlic, ginger, angelica root, cinnamon bark
- Brown sugar 1:1 with chopped herbs
- Alcohol ~25% ABV — equal volume to fermented herb mix after first ferment

## Step-by-step preparation

1. Chop herbs; layer 1:1 sugar; ferment 7 days.
2. Add equal volume alcohol; ferment 7 more days.
3. Strain; combine individual herb extracts into OHN stock.

## Ferment / wait timeline

**7 d** sugar ferment + **7 d** with alcohol per herb component.

## Ready signs (smell, foam, color)

Strong aromatic extract; clear to amber liquid after strain.

## Storage

Sealed **1–2 years**. Extremely concentrated.

## Safety & water (non-chlorinated, PPE)

**Never exceed 1:1000** application. Avoid inhaling concentrate.

## How to apply (link to application recipe name)

**OHN Pest and Immunity Drench** — strictly **1:1000** OHN:water.

## Dilution table (start conservative → stronger)

| Use | Dilution | Max |
|-----|----------|-----|
| All applications | **1:1000** | Never stronger |

## Common mistakes

- Full-strength or 1:500 — burn and phytotoxicity
- Treating as JADAM core — KNF sugar/alcohol extract
$nf_natural-farming-ohn$, 5, TRUE, 116),
    ('natural-farming-wca-wcs', 'Water-soluble calcium inputs (WCA and WCS) — KNF', NULL, 'natural_farming', 'natural_farming', 'safe', $nf_natural-farming-wca-wcs$---
domain: natural_farming
title: Water-soluble calcium inputs (WCA and WCS) — KNF
safety_tier: safe
tradition: knf
reference_source: "KNF (Cho Han-kyu); often used with JADAM"
source_tier: knf_standard
---

# Water-soluble calcium inputs (WCA and WCS) — KNF

## What it is (1 paragraph)

**WCA** dissolves calcium from roasted eggshells in brown rice vinegar (1:10). **WCS** (WCAP) dissolves phosphorus and calcium from white-ashed bones in vinegar (1:10). Both are **KNF** mineral extracts used in flowering and cell-strength programs.

## When to use

- WCA with FFJ during flower (**FFJ and WCA Flowering Boost**)
- WCA with BRV before stress (**BRV and WCA Cell Strengthener**)
- WCS when root/flower phosphorus support is needed (foliar at 1:1000 band)

## Ingredients (list with amounts)

**WCA:** roasted eggshells + unpasteurized BRV 1:10 (shells covered by vinegar)

**WCS:** beef/pork bones charred to white ash + BRV 1:10

## Step-by-step preparation

**WCA**

1. Roast eggshells light brown; cool.
2. Cover with BRV 1:10 in breathable container.
3. Fizz 7 days; strain.

**WCS**

1. Char bones to white ash completely.
2. Dissolve in BRV 1:10 for 7 days; strain.

## Ferment / wait timeline

**7 days** extraction each; gases evolve — breathable lid.

## Ready signs (smell, foam, color)

Fizzing slows; clear to amber extract; ash/shells mostly dissolved.

## Storage

Breathable container; use within **30 days** after strain.

## Safety & water (non-chlorinated, PPE)

Vinegar acid — eye protection. Roast shells/bones fully.

## How to apply (link to application recipe name)

- **FFJ and WCA Flowering Boost** — WCA **1:1000** with FFJ 1:500
- **BRV and WCA Cell Strengthener** — WCA **1:1000** with BRV 1:800

## Dilution table (start conservative → stronger)

| Input | Typical foliar dilution |
|-------|------------------------|
| WCA | **1:1000** |
| WCS | **1:1000** (same band as WCA in programs) |

## Common mistakes

- Sealed tight jar on WCA — gas buildup
- Partially charred bones — inconsistent P and off flavors
- Undiluted on leaves — burn
$nf_natural-farming-wca-wcs$, 5, TRUE, 117)
) AS v(slug, title, crop_key, guide_kind, domain, safety_tier, body_md, catalog_version, published, sort_order)
ON CONFLICT (slug) DO UPDATE SET
    title = EXCLUDED.title,
    guide_kind = EXCLUDED.guide_kind,
    domain = EXCLUDED.domain,
    safety_tier = EXCLUDED.safety_tier,
    body_md = EXCLUDED.body_md,
    catalog_version = EXCLUDED.catalog_version,
    published = EXCLUDED.published,
    sort_order = EXCLUDED.sort_order,
    updated_at = NOW();
