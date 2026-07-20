---
name: Phase 209 — Natural farming studio (operator UI)
overview: >
  Dedicated sidebar workspace for natural farming — not buried under Money.
  Switchover wizard for Mericle-style growers, make-a-batch flow, recipes &
  apply to zone/program/livestock. Rehosts Inventory.vue; links to Feed & water
  and Supplies for execution.
todos:
  - id: ws1-workspace-shell
    content: "WS1: /natural-farming workspace — tabs, route, router, workspaces.js + navGroups.js sidebar entry under Grow & operate"
    status: pending
  - id: ws2-switchover-wizard
    content: "WS2: Switchover tab — maps EC programs → canonical application recipes from recipe-canonical.yaml (208)"
    status: pending
  - id: ws3-make-batch
    content: "WS3: Make a batch tab — process type → ingredients → step cards from field guides (Ingredients/Steps/Timeline/Ready signs sections)"
    status: pending
  - id: ws3b-recipe-library
    content: "WS3b: In-workspace recipe library — browse all 14 application recipes + 15 inputs with read-only instructional panels from 208 guides"
    status: pending
  - id: ws4-recipes-apply
    content: "WS4: Recipes & apply tab — rehost Inventory recipe UI; link recipe to zone/program/crop stage; jump to Feed & water"
    status: pending
  - id: ws5-on-hand-bridge
    content: "WS5: On hand tab — embed or link SuppliesHub batch cards + Money unit costs; low-stock banner"
    status: pending
  - id: ws6-redirects-vocab
    content: "WS6: /inventory → /natural-farming?tab=...; farmer vocabulary on wizard; Fertigation deep-links updated"
    status: pending
  - id: ws7-tests-docs
    content: "WS7: phase-209-closure.test.js, nav-groups.test.js, operator-tour § Natural farming studio"
    status: pending
isProject: false
---

# Phase 209 — Natural farming studio (operator UI)

**Status:** planned · **Depends on:** [208 process knowledge](phase_208_natural_farming_process_knowledge.plan.md) — **hard gate:** WS0 recipe audit + field guides must land first so UI never shows wrong JMS dilutions

## The one job

> **One sidebar door** for "I ferment inputs and apply them" — with **step-by-step
> instructions** pulled from real field guides, not placeholder lorem.

## Problem

Today natural farming CRUD lives at **Money → Inventory & recipes** ([`Inventory.vue`](../../ui/src/views/Inventory.vue)) — correct for accountants, wrong for operators switching from bottle nutrients. The JADAM bootstrap seeds data operators never discover unless they already know to open Money.

## Workspace design

**Route:** `/natural-farming`  
**Sidebar:** `Grow & operate` group, after **Feed & water**, before **Comfort & automation**

| Tab | ID | Body | Reuse |
|-----|-----|------|-------|
| **Start here** | `start` | Switchover wizard + recipe library intro | Switchover + link to `library` tab |
| **Recipe library** | `library` | Browse all 15 inputs + 14 application recipes with instructional panels | Renders 208 field guide sections as step cards |
| **Make a batch** | `batch` | Process type → ingredients → ferment timeline | SuppliesHub batch form + 208 catalog + guide step cards |
| **Recipes & apply** | `recipes` | Application recipes + link to programs | Rehost [`Inventory.vue`](../../ui/src/views/Inventory.vue) recipe tabs |
| **On hand** | `stock` | Ready batches, low stock | [`SuppliesHub.vue`](../../ui/src/views/SuppliesHub.vue) embed or link |

### Progressive disclosure

- **Farmer path:** Start here → Make a batch → Recipes & apply
- **Accountant path:** On hand → Money for unit costs (existing)
- **Power user:** Feed & water → Advanced still has full Fertigation console

## WS2 — Switchover wizard

Target persona: understands EC, pH, mL/gal bottles; never fermented.

**Steps (v1):**

1. "What are you doing today?" — Indoor hydro / greenhouse / outdoor / livestock
2. "What commercial pattern?" — Single-part EC / A+B / dry salts / organic bottled
3. Show **mapped natural program** from `commercial_to_natural` in 208 YAML — cite real recipe names ("JLF and JMS Combined Drench", not generic "natural feed")
4. "Pick your first batch" — suggest **JMS** + **JLF General** (matches bootstrap)
5. CTA: **Make this batch** (→ batch tab with guide steps) or **Apply bootstrap** (`jadam_indoor_photoperiod_v1`)

Each wizard step shows a **"Learn how"** expander linking to the matching field guide section.

## WS3 — Make a batch (instructional)

Flow:

1. Select process type (JLF, JMS, FFJ, …) — filters catalog materials
2. Pick variant (e.g. JLF General vs Spring Nettle/Comfrey vs Crop-Specific)
3. **Step cards** from field guide: Ingredients → Step-by-step → Timeline → Ready signs → Safety
4. Create `input_definition` (if new) + `input_batch` with status lifecycle
5. Optional: linked task (`jadam_prep` category in seed)

**UI rule:** dilution and ratios displayed must come from `recipe-canonical.yaml` / field guide — never hardcoded in Vue.

## WS3b — Recipe library

Read-only browse of the full farmer inventory (208 canon):

- **Inputs tab:** 15 cards — each opens full preparation instructions
- **Application tab:** 14 cards — dilution, frequency, target stages, linked input batches
- **Programs tab:** veg / flower / outdoor bootstrap program explainer

Serves operators who won't ferment yet but need to understand what `jadam_indoor_photoperiod_v1` seeded.

**Vocabulary:** plain language on Start/Batch/Library tabs; JLF/JMS/FFJ acronyms with expanders ([`terminology-guideline.md`](../terminology-guideline.md)). Label KNF inputs (FPJ, LAB, OHN) separately from JADAM core.

Reuse [`farm.js`](../../ui/src/stores/farm.js) NF store methods — no new API.

## WS4 — Recipes & apply

- Rehost recipe + component CRUD from Inventory.vue
- **Apply** panel: pick zone → existing fertigation program or create link
- Show `target_application_type`, `dilution_ratio`, `target_growth_stages`
- Deep link: "Open in Feed & water → Programs" for schedule wiring

**Livestock:** recipes with `livestock_water_supplement` or `animal_feed` inputs link to Animals workspace when module enabled.

## WS5 — On hand bridge

Don't duplicate Money ledger — show:

- Ready batches (status `ready_for_use`, `partially_used`)
- Low-stock badges (existing [`suppliesHub.js`](../../ui/src/lib/suppliesHub.js))
- "Restock / edit costs → Money" link

## WS6 — Redirects & nav

| Legacy | Target |
|--------|--------|
| `/inventory` | `/natural-farming?tab=recipes` (or `stock` for batches) |
| Fertigation "Inventory batches →" | `/natural-farming?tab=stock` |
| Money inventory tab | Keep as power-user shortcut OR relabel "Natural farming (advanced)" linking to studio |

Update [`navGroups.js`](../../ui/src/lib/navGroups.js), [`workspaces.js`](../../ui/src/lib/workspaces.js), [`navRelations.js`](../../ui/src/lib/navRelations.js).

## Acceptance criteria

- [ ] Sidebar shows **Natural farming** under Grow & operate
- [ ] `/natural-farming` loads with five tabs; default `start`
- [ ] Recipe library shows all 15 inputs + 14 application recipes with step-by-step content from guides
- [ ] Make a batch shows field-guide step cards (not empty placeholders)
- [ ] Switchover wizard renders EC→natural mapping from YAML
- [ ] Make a batch creates input + batch via existing API
- [ ] Recipes tab can link recipe to fertigation program (deep link at minimum)
- [ ] `/inventory` redirects without 404
- [ ] `nav-groups.test.js` + `phase-209-closure.test.js` green
- [ ] No Guardian/smoke test file changes in this phase

## Out of scope

- Guardian propose/draft tools (210)
- New API endpoints
- IoT ferment monitoring
- Full Animals ration UI
