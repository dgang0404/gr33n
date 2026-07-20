---
name: Phase 208 — Natural farming process & material knowledge
overview: >
  Structured knowledge layer grounded in real JADAM recipes already in master_seed.sql
  (Cho 2016). Field guides with step-by-step instructions, material→process catalog,
  recipe audit to fix dilution drift, RAG ingest. Unblocks Guardian read tools in 210.
todos:
  - id: ws0-recipe-audit
    content: "WS0: Audit master_seed.sql + bootstrap dilutions against Cho 2016 — fix JMS 1:500→1:10 soil / 1:20 foliar; add missing FAA input; document every change with page/ref"
    status: completed
  - id: ws1-process-vocabulary
    content: "WS1: Process type vocabulary + material role enum; distinguish JADAM (JLF/JMS) vs KNF (FPJ/LAB) vs Ingham (compost tea)"
    status: completed
  - id: ws2-field-guides
    content: "WS2: Field guides — one per core input (14) + application how-to + goldenrod extension + forest-garden + livestock primer"
    status: completed
  - id: ws3-yaml-catalog
    content: "WS3: process-material-catalog.yaml + recipe-canonical.yaml — material links, dilution bands, source_tier, links to seed row names"
    status: completed
  - id: ws4-rag-ingest
    content: "WS4: RAG ingest for new guides + catalog; manifest entries in field-guides/README.md"
    status: completed
  - id: ws5-read-api
    content: "WS5: GET /v1/field-guides/process-catalog + /recipes — static serve from YAML; no farm writes"
    status: completed
  - id: ws6-tests-docs
    content: "WS6: phase-208-closure.test.js — YAML shape, all canonical recipes present, JMS dilution audit test; no smoke changes"
    status: pending
isProject: false
---

# Phase 208 — Natural farming process & material knowledge

**Status:** WS0–WS3 complete · **Depends on:** [207 roadmap](phase_207_natural_farming_studio.plan.md) · **Blocks:** 209 studio UI, 210 Guardian

## The one job

> Guardian and the studio UI cite **real, sourced recipes** — not invented ratios —
> when answering "how do I make JMS?" or "what can I make from goldenrod?"

## Canonical recipe source (already in repo)

**Do not invent recipes in field guides.** Phase 208 extracts instructional copy from
[`db/seeds/master_seed.sql`](../../db/seeds/master_seed.sql) (`input_definitions` +
`application_recipes`), audited against primary references, then mirrors into field
guides + YAML.

Every recipe row carries `reference_source`. Default for JADAM inputs:

> `JADAM Organic Farming, Youngsang Cho, 2016`

Compost tea row cites Elaine Ingham (separate tradition). FPJ/LAB are **Korean Natural
Farming (KNF)** inputs also used alongside JADAM — label them honestly in guides.

### WS0 — Recipe audit (do first)

**Primary references:** Youngsang Cho, *JADAM Organic Farming*, 2016 · [en.jadam.kr](https://en.jadam.kr) · UH CTAHR KNF leaflets (FPJ/FFJ/WCA) · Cho Han-kyu KNF for inputs JADAM book does not fully cover.

#### Tradition labeling (fix `reference_source`)

| Input | Seed cites | Actually is |
|-------|------------|-------------|
| JMS, JLF×3, JWA, JS, JHS, WCA, WCS, BRV | Cho 2016 | **JADAM** (Youngsang Cho) — verify each prep against book |
| FPJ, FFJ, LAB, OHN | Cho 2016 | **KNF** (Cho Han-kyu) — commonly paired with JADAM; relabel `reference_source` to `"KNF (Cho Han-kyu); often used with JADAM"` |
| Compost Tea AACT | Ingham | **Correct** — not JADAM |

JADAM explicitly **eliminated sugar** from its core inputs; FPJ/FFJ/LAB/OHN are sugar-based KNF preparations.

---

#### Input definitions — accuracy matrix

| Input | Preparation in seed | Verdict | Fix |
|-------|---------------------|---------|-----|
| **JMS** | Leaf mold + potato water + salt; 3–7 days | **Mostly OK** | Cho: boil potato, suspend in mesh, **24–72 h active window**, use at peak foam. Change copy to "use within 6–12 h of peak activity" not generic 1 week. Scale: 1 potato / ~10–20 L batch is book-aligned. |
| **LAB** | Rice wash → milk; golden serum | **OK (KNF)** | Relabel source to KNF. Prep matches standard KNF LAB. |
| **FPJ** | Tips + sugar 1:1; 3–7 days | **OK (KNF)** | Relabel source. Matches CTAHR SA-7. |
| **FFJ** | Fruit + sugar 1:1; 7 days | **OK (KNF)** | Relabel source. Bootstrap wrongly adds **water** — KNF FFJ is fruit+sugar only, no water. |
| **BRV** | Purchased BRV | **OK** | — |
| **OHN** | Herbs + sugar + alcohol | **OK (KNF)** | Relabel source. Standard KNF OHN (5 herbs). |
| **JHS** | Wormwood, sim 1–3 h | **Wrong process** | Cho JHS: **boil** fresh plant (Jerusalem artichoke default) **4–5 hours** in mesh bag; 1 kg plant : 4–5 L water. Wormwood/neem/garlic-chives are valid *materials* but cold/simmer 1–3 h is not the book method. |
| **WCA** | Eggshell + BRV 1:10; roast; 7 days | **OK (KNF)** | Matches CTAHR / KNF 1:10 shells:vinegar. Relabel to KNF. |
| **WCS** | Charred bones + BRV 1:10 | **OK (KNF WCAP)** | Relabel to KNF. |
| **JWA** | Ash lye + oil → soap | **Directionally OK** | Cho book has full caustic potash + canola process; seed summary acceptable for v1 if guide has full steps. |
| **JS** | Wettable sulfur 0.5% + JWA | **WRONG INPUT** | Real **JADAM JS** = sulfur + **caustic soda (NaOH)** + red clay + phyllite + sea salt — exothermic batch, ~25% sulfur concentrate. Application: **0.5–2 L JS per 500 L water** (~1:250–1:1000 of concentrate). **Not** garden wettable sulfur. Rename seed row or split: `JS (JADAM Sulfur concentrate)` vs remove wettable-sulfur shortcut. |
| **JLF General** | Weeds 2/3 + leaf mold + water; 7–14 d | **OK (JADAM)** | Matches Cho/FigJam. No sugar — correct. Ferment can be **weeks to months**; 7–14 d is minimum usable, not ideal max. |
| **JLF Crop-Specific** | Same-crop residue | **OK (JADAM)** | — |
| **JLF Spring** | Nettle/comfrey; "nitrogen-fixing plants" | **Prep OK; description wrong** | Nettle/comfrey are **dynamic accumulators / high-N biomass**, not nitrogen-fixing. Fix description text. |
| **Compost Tea** | AACT 24–48 h; 4 h use window | **OK (Ingham)** | — |
| **FAA** | — | **Missing** | Add KNF FAA: fish + brown sugar 1:1; dilute ~1:1000. |

---

#### Application recipes — accuracy matrix

| Recipe | Seed dilution | Authoritative target | Verdict |
|--------|---------------|----------------------|---------|
| **JMS Soil Drench** | 1:500 | **1:10** (1 part JMS : 10 water) | **WRONG — fix** |
| **JMS Foliar Spray** | 1:500 | **1:20** minimum + **JWA** for coverage | **WRONG — fix** |
| **JLF and JMS Combined** | JLF 1:20 + JMS 1:500 | JLF 1:20 (or start 1:100) + JMS **1:10** | **JMS part wrong** |
| **JLF General Soil Drench** | 1:20 | Cho/FigJam: **start 1:100**, range 1:20–1:500 by age/soil | **Strong default** — keep 1:20 as "experienced" tier; add UI note "start 1:100" |
| **JLF Seedling Drench** | 1:30 | 1:30–1:50 for seedlings | **OK** |
| **JLF Foliar Feed** | 1:30–1:50 + JWA | Reasonable stress foliar | **OK** |
| **LAB Soil Conditioner** | 1:1000 | KNF standard 1:1000 | **OK** |
| **OHN Pest Drench** | 1:1000 | KNF 1:1000 (never stronger) | **OK** |
| **FPJ Vegetative Foliar** | 1:500–1:1000 + JWA | CTAHR 1:500; 1:800–1:1000 when stacking | **OK** |
| **FFJ + WCA Flowering** | FFJ 1:500 + WCA 1:1000 | KNF growth-stage recipes use ~1:1000 for each | **OK** |
| **BRV + WCA Cell Strengthener** | BRV 1:800 + WCA 1:1000 | KNF-style; conservative | **OK** |
| **JHS + JWA Pesticide** | JHS 1:50 + JWA 1:500 | Cho: 3–20 L JHS per 500 L (~1:25–1:167) + JWA | **Ballpark OK** if JHS is real boiled concentrate |
| **JS Fungicide** | 0.5% wettable sulfur | Real JS: 0.5–2 L per 500 L of **JADAM JS concentrate** | **Wrong if labeled JS** — fix after JS input corrected |
| **JWA Insecticide** | 1:500 | Cho JWA foliar/pesticide dilutions | **OK** |

**Component math note:** Combined drench stores JMS as `0.025` relative part assuming 1:500 — after fix, relative part for 1:10 vs 1:20 JLF base needs recalculation.

---

#### Goldenrod / extension (not in seed yet)

| Claim | Verdict |
|-------|---------|
| Goldenrod → JLF | **Valid extension** — JADAM principle is "local weeds"; goldenrod is usable biomass |
| Named Cho recipe | **No** — must label `source_tier: extension_method` |
| 1:20 drench default | **Too strong** — start **1:100** per Cho/FigJam conservative guidance |
| Dye harvest compatible | **OK** — biomass use does not conflict with dye harvest |

---

#### WS0 deliverables

1. Fix `master_seed.sql` + bootstrap migration dilutions (JMS trio)
2. Fix JLF Spring description (nitrogen-fixing → high-N/dynamic accumulator)
3. Replace or rename JS input to real JADAM JS process (or mark wettable-sulfur row as deprecated shortcut with warning)
4. Expand JHS preparation to boiled-herb method
5. Fix bootstrap FFJ ingredients (remove water)
6. Relabel KNF `reference_source` on FPJ, FFJ, LAB, OHN, WCA, WCS
7. Add FAA input + optional WCS application recipe
8. Write `docs/field-guides/procedures/recipe-audit-log.md` with one line per change + citation

Known drift summary (quick reference):

| Item | Seed today | Correct |
|------|------------|---------|
| JMS soil drench | 1:500 | **1:10** |
| JMS foliar | 1:500 | **1:20** + JWA |
| JMS in combined drench | 1:500 | **1:10** |
| JS input | Wettable sulfur 0.5% | **JADAM JS** (caustic soda batch) |
| JHS prep | Simmer 1–3 h | **Boil 4–5 h** (Cho method) |
| JLF Spring desc | "N-fixing plants" | **Dynamic accumulators** |
| FFJ bootstrap | fruit, sugar, water | **fruit + sugar only** |
| JLF default drench | 1:20 only | **Start 1:100**, up to 1:20 when tested |

---

## Complete farmer recipe inventory (v1 canon)

These **15 input definitions** and **14 application recipes** must appear in field guides,
YAML canon, and (after audit) match seed/bootstrap.

### Input definitions (make / buy / ferment)

| # | Seed name | Category | Tradition | Preparation summary (from seed) |
|---|-----------|----------|-----------|--------------------------------|
| 1 | JMS (JADAM Microbial Solution) | microbial_inoculant | JADAM | Leaf mold + potato water + pinch salt; ferment 3–7 days |
| 2 | LAB (Lactic Acid Bacteria Serum) | microbial_inoculant | KNF | Rice wash → milk culture; extract golden serum |
| 3 | FPJ (Fermented Plant Juice) | fermented_plant_juice | KNF | Growing tips + brown sugar 1:1; 3–7 days |
| 4 | FFJ (Fermented Fruit Juice) | fermented_plant_juice | JADAM/KNF | Ripe fruit + brown sugar 1:1; ~7 days |
| 5 | BRV (Brown Rice Vinegar) | fermented_plant_juice | JADAM | Purchased unpasteurized BRV |
| 6 | OHN (Oriental Herbal Nutrient) | oriental_herbal_nutrient | KNF | Garlic, ginger, angelica, cinnamon + sugar + alcohol |
| 7 | JHS (JADAM Herbal Solution) | oriental_herbal_nutrient | JADAM | Wormwood, artemisia, garlic chives, hot pepper, neem — water extract |
| 8 | WCA (Water-Soluble Calcium) | water_soluble_nutrient | JADAM | Eggshells + BRV 1:10; 7 days |
| 9 | WCS (Water-Soluble Calcium Phosphate) | water_soluble_nutrient | JADAM | Charred bones + BRV 1:10; 7 days |
| 10 | JWA (JADAM Wetting Agent) | other_extract | JADAM | Wood ash lye + plant oil → soap |
| 11 | JS (JADAM Sulfur) | other_extract | JADAM | Wettable sulfur 0.5% + JWA |
| 12 | JLF General (Weed and Grass) | other_ferment | JADAM | Weeds 2/3 + leaf mold + water; 7–14 days |
| 13 | JLF Crop-Specific (Crop Residue) | other_ferment | JADAM | Same-crop residue — targeted fertilizer |
| 14 | JLF Spring (Nettle and Comfrey) | other_ferment | JADAM | Nettle/comfrey tops + leaf mold + water |
| 15 | Compost Tea Actively Aerated | compost_tea_extract | Ingham | Aerated compost + molasses; 24–48 h brew |

**WS0 add:** FAA (Fish Amino Acid) — fish scraps + brown sugar 1:1; KNF standard.

### Application recipes (how to apply)

| # | Recipe | Type | Dilution (post-audit target) | When |
|---|--------|------|------------------------------|------|
| 1 | JMS Soil Drench | soil_drench | **1:10** JMS:water | Every 2 weeks; before transplant |
| 2 | JLF General Soil Drench | soil_drench | 1:20 JLF:water (start 1:100 if unsure) | Primary fertility; weekly–biweekly |
| 3 | JLF Seedling Drench | soil_drench | 1:30 | Germination through 2 weeks post-transplant |
| 4 | JLF and JMS Combined Drench | soil_drench | JLF 1:20 + JMS 1:10 in same tank | Peak season weekly |
| 5 | LAB Soil Conditioner | soil_drench | 1:1000 | Every 2–4 weeks; pre-transplant |
| 6 | OHN Pest and Immunity Drench | soil_drench | 1:1000 (never stronger) | Preventative / pest pressure |
| 7 | JMS Foliar Spray | foliar_spray | **1:20** + JWA | Every 1–2 weeks; filter well |
| 8 | FPJ Vegetative Foliar | foliar_spray | 1:500–1:1000 + JWA | Veg only — stop at flower |
| 9 | FFJ and WCA Flowering Boost | foliar_spray | FFJ 1:500 + WCA 1:1000 + JWA | Flower through early fruit |
| 10 | BRV and WCA Cell Strengthener | foliar_spray | BRV 1:800 + WCA 1:1000 | Before stress / fruiting |
| 11 | JHS and JWA Natural Pesticide | foliar_spray | JHS 1:50 + JWA 1:500 | Weekly preventative |
| 12 | JS Fungicide Spray | foliar_spray | 0.5% sulfur + JWA | ≤32°C; repeat 5–7 days |
| 13 | JLF Foliar Feed | foliar_spray | 1:30–1:50 + JWA | Stress only — not primary feed |
| 14 | JWA Insecticide Spray | foliar_spray | 1:500 | Soft-bodied insects |

Bootstrap [`jadam_indoor_photoperiod_v1`](../../db/migrations/20260703_phase124_fix_bootstrap_batch_label.sql) wires three **programs** to these recipes:

- Veg Daily JLF Program → JLF and JMS Combined Drench
- Flower Daily FFJ+WCA Program → FFJ and WCA Flowering Boost
- Outdoor JLF Soil Drench → JLF General Soil Drench

---

## WS1 — Process vocabulary & honesty labels

| Label | Meaning in UI/guides |
|-------|---------------------|
| **JADAM** | Methods from Cho 2016 — JMS, JLF, JWA, JS, JHS, WCA, WCS |
| **KNF** | Korean Natural Farming — FPJ, LAB, OHN, FAA (often paired with JADAM) |
| **Extension** | Valid JADAM *method* applied to local material not named in Cho (e.g. goldenrod JLF) |
| **Other** | Compost tea (Ingham), purchased BRV |

Schema enum mapping unchanged — see [`data/natural_farming_process_vocabulary.yaml`](../../data/natural_farming_process_vocabulary.yaml) `schema_category_map` and `seed_inputs`.

---

## WS2 — Field guides (instructional — required structure)

Each guide **must** include these sections (studio UI renders them as step cards):

```markdown
---
title: ...
safety_tier: safe | caution | expert
tradition: jadam | knf | extension | other
reference_source: "..."
source_tier: cho_named | knf_standard | extension_method | third_party
---

# Title

## What it is (1 paragraph)
## When to use
## Ingredients (list with amounts)
## Step-by-step preparation
  1. ...
  2. ...
## Ferment / wait timeline
## Ready signs (smell, foam, color)
## Storage
## Safety & water (non-chlorinated, PPE)
## How to apply (link to application recipe name)
## Dilution table (start conservative → stronger)
## Common mistakes
```

### Required guide files (18 minimum)

**Core inputs (match seed names):**

| File | Covers |
|------|--------|
| `natural-farming-jms.md` | JMS make + apply ( soil 1:10, foliar 1:20 ) |
| `natural-farming-jlf-general.md` | JLF from weeds — canonical Cho method |
| `natural-farming-jlf-crop-specific.md` | Same-crop residue JLF |
| `natural-farming-jlf-spring-nettle-comfrey.md` | High-N spring push |
| `natural-farming-ffj.md` | FFJ make + flower use |
| `natural-farming-fpj.md` | FPJ (KNF) — label as KNF, not JADAM core |
| `natural-farming-lab.md` | LAB serum |
| `natural-farming-ohn.md` | OHN — 1:1000 max |
| `natural-farming-wca-wcs.md` | Calcium inputs |
| `natural-farming-jwa-js-jhs.md` | Wetting agent + pest/disease sprays |
| `natural-farming-brv.md` | Purchased BRV + foliar use |
| `natural-farming-faa.md` | Fish amino acid (WS0 add) |
| `natural-farming-compost-tea-aact.md` | Ingham AACT — 4-hour use window |

**Application & program guides:**

| File | Covers |
|------|--------|
| `natural-farming-application-recipes.md` | All 14 application recipes in one reference table |
| `natural-farming-indoor-photoperiod-program.md` | Maps bootstrap veg/flower/outdoor programs |

**Extensions (honest sourcing):**

| File | Covers |
|------|--------|
| `natural-farming-goldenrod-jlf.md` | **Extension:** JLF-general method on goldenrod biomass; dye harvest compatible; start dilution 1:100 |
| `natural-farming-forest-garden-understory.md` | Cherry/blackberry/goldenrod ecology — no invented EC |
| `natural-farming-livestock-plant-feed.md` | Comfrey, sprouted grain → `animal_feed` (simple, not ration math) |

Cross-link [`crop-unsupported-woodland.md`](../field-guides/crop-unsupported-woodland.md).

### Goldenrod — sourcing honesty

Goldenrod JLF is **not** a named Cho recipe. It is:

1. **JLF General method** (Cho 2016 — "use local weeds")
2. Applied to **Solidago** biomass (dynamic accumulator — extension tier)
3. For orchard understory — operator choice, not prescriptive removal of goldenrod

Guardian must say "JLF from goldenrod using the standard JLF method" — never "Cho's goldenrod recipe."

---

## WS3 — YAML catalogs

### `process-material-catalog.yaml`

Material → process links for Guardian + studio picker. Every entry needs:

```yaml
materials:
  - id: goldenrod
    common_names: ["Canadian goldenrod", "Solidago canadensis"]
    roles: [dye, fertigation]
    source_tier: extension_method
    base_method_guide: natural-farming-jlf-general
    processes:
      - type: jlf
        season: spring
        guide: natural-farming-goldenrod-jlf
        target_crops: [cherry, apple, plum]
        dilution_start: "1:100"
        dilution_strong: "1:30"
        notes: "Start conservative per Cho; stronger only after plant response"
  - id: nettle
    source_tier: cho_named
    processes:
      - type: jlf
        guide: natural-farming-jlf-spring-nettle-comfrey
  - id: comfrey
    source_tier: cho_named
    processes:
      - type: jlf
        guide: natural-farming-jlf-spring-nettle-comfrey
      - type: animal_feed
        guide: natural-farming-livestock-plant-feed
```

### `recipe-canonical.yaml`

Machine-readable mirror of seed inventory for API + closure tests:

```yaml
inputs:
  - seed_name: "JMS (JADAM Microbial Solution)"
    guide: natural-farming-jms.md
    reference_source: "JADAM Organic Farming, Youngsang Cho, 2016"
application_recipes:
  - seed_name: "JLF and JMS Combined Drench"
    components: ["JLF General (Weed and Grass)", "JMS (JADAM Microbial Solution)"]
    dilution: "JLF 1:20 + JMS 1:10"
```

**Switchover mappings** (for Phase 211):

```yaml
commercial_to_natural:
  - commercial: "Daily EC veg feed 1.6–1.8 mS/cm"
    natural_equivalent:
      - recipe: "JLF and JMS Combined Drench"
        frequency: "Weekly peak season"
      - recipe: "JMS Foliar Spray"
        frequency: "Every 1–2 weeks"
  - commercial: "Flower boost A+B"
    natural_equivalent:
      - recipe: "FFJ and WCA Flowering Boost"
        frequency: "Weekly from first buds"
```

---

## WS4 — RAG ingest

- Manifest all guides in [`docs/field-guides/README.md`](../field-guides/README.md)
- Tag: `domain: natural_farming`, `tradition: jadam|knf|extension`
- Reingest via existing Guardian path

## WS5 — Read API

```
GET /v1/field-guides/process-catalog
GET /v1/field-guides/process-catalog/materials/{id}
GET /v1/field-guides/recipe-canon          # from recipe-canonical.yaml
```

## WS6 — Tests & docs

- `phase-208-closure.test.js`:
  - All 15+ input names in `recipe-canonical.yaml`
  - All 14 application recipes present
  - `goldenrod` in material catalog with `source_tier: extension_method`
  - JMS soil recipe dilution is `1:10` not `1:500` (post-audit)
  - Every guide file exists on disk
- **Do not touch** smoke fixtures

## Acceptance criteria

- [ ] WS0 audit complete; `recipe-audit-log.md` documents JMS fix
- [ ] 18+ field guides with full step-by-step instructional sections
- [ ] Zero invented ratios — every number traceable to seed (post-audit) or cited extension
- [ ] Goldenrod labeled `extension_method`, not `cho_named`
- [ ] FAA input added to seed + guide
- [ ] YAML catalogs + read API return canonical data
- [ ] RAG retrieves "how to make JMS" with correct 1:10 / 1:20 dilutions
- [ ] Smoke suite unchanged

## Out of scope

- Guardian read tools (210)
- Studio UI (209)
- Per-farm custom material DB rows
