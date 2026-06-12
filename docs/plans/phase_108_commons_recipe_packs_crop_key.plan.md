---
name: Phase 108 — Commons recipe packs crop_key tags
overview: >
  Recipe pack manifest requires crop_keys + stages per recipe; import script
  writes application_recipes.meta — enterprise promotion aligned with Phase 102.
todos:
  - id: ws0-deps
    content: "WS0: Phase 102 WS2 recipe.meta schema shipped"
    status: completed
  - id: ws1-manifest
    content: "WS1: commons pack JSON schema — recipes[].crop_keys, stages, ec_band"
    status: completed
  - id: ws2-import
    content: "WS2: import-recipe-pack.sh writes meta; validate against crop_catalog"
    status: completed
  - id: ws3-seed
    content: "WS3: Update phase31/enterprise demo packs with crop_key tags"
    status: completed
  - id: ws4-ui
    content: "WS4: Commons catalog UI shows crop fit badges on recipe packs"
    status: completed
  - id: ws5-smokes
    content: "WS5: Import pack → Start grow suggests tagged recipe for cannabis flower"
    status: completed
isProject: false
---

# Phase 108 — Commons recipe packs crop_key tags

## Status

**Planned.** Extends [Phase 102 WS8](phase_102_fertigation_program_catalog_metadata.plan.md) for **commons promotion** of recipes across sites.

**Depends on:** [Phase 102](phase_102_fertigation_program_catalog_metadata.plan.md), [Phase 98](phase_98_enterprise_catalog_promotion.plan.md).

**Closure:** **OC-108**

---

## The one job

> **Import recipe pack from commons** → every `application_recipe` row carries **`crop_keys`** and **stages** so Phase 96/102 validation works on enterprise sites without hand-tagging.

---

## Pack manifest extension (WS1)

```yaml
recipes:
  - slug: ffj-wca-flower-boost
    crop_keys: [cannabis, tomato]
    stages: [early_flower, mid_flower, late_flower]
    profile_ec_source: { crop_key: cannabis, stage: mid_flower }
```

Import rejects unknown `crop_key` (parity with catalog).

---

## Relationship to Phase 102

| Phase 102 | Phase 108 |
|-----------|-----------|
| Farm-local program + recipe meta | **Commons pack** carries meta at import |
| Seed demo farm | **Promote** tagged packs org-wide |

---

## Acceptance

- [ ] `import-recipe-pack.sh` fails on invalid crop_key
- [ ] Imported recipes visible in Phase 102 program suggest API
- [ ] Document in commons-catalog-operator-playbook

**Prompt loop:** **`phase 108`**.
