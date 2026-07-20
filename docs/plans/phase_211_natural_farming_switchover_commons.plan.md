---
name: Phase 211 — Switchover packs & Commons recipe import
overview: >
  Close the Mericle-style switchover loop: importable natural-farming recipe
  packs via Commons, EC-program template mappings as bootstrap-adjacent packs,
  livestock feed use-case templates. Optional smoke promotion only after 210
  regression is stable.
todos:
  - id: ws1-recipe-packs
    content: "WS1: Commons recipe pack format — input_definitions + application_recipes JSON; import into gr33nnaturalfarming"
    status: completed
  - id: ws2-switchover-packs
    content: "WS2: Switchover pack keys — mericle_veg_to_jlf_v1, mericle_flower_to_ffj_v1 mapping tables in YAML + bootstrap helper"
    status: completed
  - id: ws3-livestock-templates
    content: "WS3: Livestock feed templates — comfrey slurry, sprouted grain → animal_feed inputs (demo seed + pack)"
    status: pending
  - id: ws4-studio-import
    content: "WS4: Natural farming studio Start tab — 'Import a recipe pack' → Commons catalog flow"
    status: pending
  - id: ws5-smoke-promotion
    content: "WS5 (optional): Promote cherry+JLF bar to smoke tier 2 OR add 5th smoke step — only after regression green 3+ runs"
    status: pending
  - id: ws6-tests-docs
    content: "WS6: Import smoke test, phase-211-closure, pattern-playbooks.md entry"
    status: pending
isProject: false
---

# Phase 211 — Switchover packs & Commons recipe import

**Status:** planned · **Depends on:** [208–210](phase_207_natural_farming_studio.plan.md) · **Last slice** of natural farming arc

## The one job

> Operators can **import** a vetted recipe pack and apply a **switchover mapping**
> instead of hand-entering JLF/JMS definitions — the Mericle→natural path becomes
> one Confirm click, not a weekend of typing.

## WS1 — Commons recipe pack format

**Pack content is exported from audited seed — not hand-written JSON.**

Source of truth after Phase 208 WS0:

1. Export [`db/seeds/master_seed.sql`](../../db/seeds/master_seed.sql) natural-farming INSERT blocks
2. Export bootstrap subset from [`20260703_phase124_fix_bootstrap_batch_label.sql`](../../db/migrations/20260703_phase124_fix_bootstrap_batch_label.sql)
3. Validate against `recipe-canonical.yaml`

```json
{
  "pack_key": "jadam_indoor_starter_recipes_v1",
  "reference_source": "JADAM Organic Farming, Youngsang Cho, 2016",
  "input_definitions": [ "...15 rows post-audit..." ],
  "application_recipes": [ "...14 rows..." ],
  "recipe_input_components": [ "..." ]
}
```

- Idempotent import per farm (`ON CONFLICT` on farm+name)
- Guardian tool: `import_natural_farming_pack` (admin, Confirm-gated) — optional if UI suffices

## WS2 — Switchover packs

Bootstrap-adjacent keys (like `jadam_indoor_photoperiod_v1`):

| Pack key | Creates |
|----------|---------|
| `mericle_veg_to_jlf_v1` | JMS + veg JLF definitions, 2 application recipes, links to existing veg zone programs (proposal) |
| `mericle_flower_to_ffj_v1` | FFJ + WCA foliar recipes for flower stage |

Implementation: new branch in `apply_farm_bootstrap_template` **or** standalone `POST /farms/{id}/naturalfarming/apply-pack` — prefer latter to avoid bloating bootstrap function (ponytail: separate endpoint, reuse bootstrap idempotency pattern).

Switchover wizard (209) gets "Apply pack" CTAs.

## WS3 — Livestock templates

- `livestock_comfrey_feed_v1` — comfrey → `animal_feed` input + simple consumption example
- Document in [`natural-farming-livestock-plant-feed.md`](../field-guides/natural-farming-livestock-plant-feed.md) (208)
- Wire to Animals module when enabled ([`farmModules.js`](../../ui/src/lib/farmModules.js))

**Not:** TMR balancing, NRC requirements, or automated feeding schedules.

## WS4 — Studio import UX

Natural farming **Start here** tab:

- Browse Commons packs tagged `natural_farming`
- Preview: inputs + recipes included
- Import → creates farm rows → "Make first batch" CTA

## WS5 — Smoke promotion (optional, gated)

**Preconditions:**

- `regression-cherry-goldenrod-jlf` passes 3+ consecutive runs on CPU path
- Process catalog ingested in default dev seed

**Options (pick one in implementation):**

| Option | Change |
|--------|--------|
| A | Add 5th smoke step `smoke-cherry-jlf` — keep original cherry as step 1 unchanged |
| B | Tier-2 smoke profile via env `GUARDIAN_SMOKE_TIER=2` |
| C | Never promote — regression only |

Default recommendation: **Option A** — additive fifth step, zero change to existing four.

## Acceptance criteria

- [ ] At least one Commons recipe pack imports cleanly on demo farm
- [ ] `mericle_veg_to_jlf_v1` creates definitions + recipes idempotently
- [ ] Switchover wizard can apply pack end-to-end
- [ ] Livestock template visible when Animals module on
- [ ] Smoke steps 1–4 still pass with original criteria
- [ ] `phase-211-closure` green

## Out of scope

- Marketplace / paid packs
- Multi-farm org-level pack defaults
- Replacing `jadam_indoor_photoperiod_v1` bootstrap
