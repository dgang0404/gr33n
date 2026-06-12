---
name: Phase 102 — Fertigation & recipe ↔ crop profile linkage
overview: >
  Programs and application_recipes tagged with crop_key + stages; EC bands aligned
  to crop_profile_stages; Phase 96 validates at attach; Start grow suggests matched
  recipes. Full chain: crop_key → profile targets → program → recipe.
todos:
  - id: ws0-deps
    content: "WS0: Phase 86 crop_key on cycles; effective crop_profiles per farm"
    status: pending
  - id: ws1-program-meta
    content: "WS1: fertigation_programs.meta — recommended_crop_keys, recommended_stages, profile_ec_source"
    status: pending
  - id: ws2-recipe-meta
    content: "WS2: application_recipes.meta — crop_keys, stages, links to program ids"
    status: pending
  - id: ws3-ec-align
    content: "WS3: ec_band_mscm derived from crop_profile_stages for tagged crop_key+stage"
    status: pending
  - id: ws4-seed
    content: "WS4: Seed demo programs + recipes (veg JLF, flower FFJ+WCA) with crop_key tags"
    status: pending
  - id: ws5-api
    content: "WS5: GET programs/recipes filter by crop_key + stage; suggest for Start grow"
    status: pending
  - id: ws6-ui
    content: "WS6: Start grow + Water tab — program picker filtered; recipe name shows crop fit"
    status: pending
  - id: ws7-phase96
    content: "WS7: Phase 96 validation reads program+recipe metadata (not name heuristics)"
    status: pending
  - id: ws8-commons
    content: "WS8: Commons recipe pack import preserves crop_key tags in recipe meta"
    status: pending
  - id: ws9-guardian
    content: "WS9: Guardian feeding advice cites program crop_key + profile EC chain"
    status: pending
isProject: false
---

# Phase 102 — Fertigation & recipe ↔ crop profile linkage

## Status

**Planned.** Long-term home for **recipe ↔ crop profile** linkage (not Phase 101 — that is Guardian write tools).

**Depends on:** [Phase 86](phase_86_grow_ops_catalog_chain.plan.md), [Phase 96](phase_96_grow_feeding_program_validation.plan.md) (v1 warnings first).

**Closure:** **OC-102**

---

## The one job

> **One feeding chain:** `crop_key` → **crop profile stage EC** → **fertigation program** → **application recipe** (mix). Operators and Guardian never attach a veg recipe to a flower-stage cannabis grow without a visible warning.

---

## Problem today

| Layer | Knows crop? | Knows stage? |
|-------|-------------|--------------|
| `crop_profiles` + stages | ✅ `crop_key` | ✅ per stage EC/DLI |
| `fertigation_programs` | ❌ | ❌ (implicit in name only) |
| `application_recipes` | ❌ | ❌ |
| Zone strip | ✅ via plant/cycle | ✅ `current_stage` |
| Active pump program | ❌ | ❌ |

Programs have `application_recipe_id` but **no `crop_key` tags**. Phase **96** adds **attach-time validation**; Phase **102** adds **catalog metadata** so validation is data-driven as recipes multiply.

---

## Target pipeline

```
crop_catalog_entries / crop_key
        │
        ▼
crop_profiles + crop_profile_stages   ← EC mS/cm, pH (Settings override per farm)
        │
        │  profile_ec_source: { crop_key, stage }  (on program meta)
        ▼
fertigation_programs                  ← recommended_crop_keys, recommended_stages
        │
        │  application_recipe_id
        ▼
application_recipes                   ← crop_keys[], stages[], mix components
        │
        ▼
crop_cycles.primary_program_id        ← Phase 96 validates fit
```

---

## WS1 — Program metadata

On `gr33nfertigation.fertigation_programs.meta` (JSONB):

```json
{
  "recommended_crop_keys": ["cannabis", "tomato"],
  "recommended_stages": ["early_veg", "late_veg"],
  "profile_ec_source": { "crop_key": "cannabis", "stage": "late_veg" },
  "ec_band_mscm": { "min": 1.4, "max": 2.0 }
}
```

- **`profile_ec_source`** — optional pointer; server can **derive** `ec_band_mscm` from effective farm profile at seed time
- **`ec_band_mscm`** — denormalized cache for quick UI/Guardian compare vs live reservoir EC

---

## WS2 — Recipe metadata

On `gr33nnaturalfarming.application_recipes.meta`:

```json
{
  "crop_keys": ["cannabis"],
  "stages": ["early_flower", "mid_flower", "late_flower"],
  "feeding_style": "ffj_wca_boost",
  "linked_program_names": ["Flower FFJ+WCA"]
}
```

Recipes inherit program tags when created via program; standalone recipes (Inventory) get tags at author time.

---

## WS3 — EC alignment with crop profile

When tagging a program for `crop_key=cannabis`, stage `early_flower`:

1. Load **effective** farm profile (builtin or Settings override)
2. Read `crop_profile_stages` row for `early_flower`
3. Set `ec_band_mscm` from `ec_min` / `ec_max` / `ec_target`

UI Start grow preview:

> **Profile target:** 1.6–2.2 mS/cm · **Program band:** 1.6–2.2 mS/cm ✓

Mismatch → amber before attach (Phase 96).

---

## Phase 96 vs Phase 102 (split)

| | Phase 96 | Phase 102 |
|---|----------|-----------|
| **When** | After Phase 86 | After 96 (or parallel WS7) |
| **Scope** | Warn/block on attach | Metadata model + seed + filters |
| **v1** | Name/stage heuristics OK | — |
| **v2** | Reads program+recipe meta | Defines meta schema + API |

**Do not skip 96** waiting for 102 — operators need mismatch warnings early.

---

## WS8 — Commons recipe packs

Enterprise recipe import ([`import-recipe-pack.sh`](../scripts/enterprise/import-recipe-pack.sh)) must preserve **`crop_keys`** / **`stages`** in recipe `meta` when pack manifest includes them.

Cross-link [Phase 98](phase_98_enterprise_catalog_promotion.plan.md) — platform recipes promote; farm program links stay local.

---

## WS9 — Guardian

When `primary_program_id` set on active cycle:

- Read program `recommended_crop_keys` vs `plants.crop_key`
- Compare program `ec_band_mscm` vs `lookup_crop_targets` output
- Narrative: “Strip EC from **crop profile**; pump runs **Flower FFJ+WCA** (tagged cannabis, flower stages)”

---

## Acceptance

- [ ] Demo seed: veg JLF program tagged `early_veg`/`late_veg`; flower FFJ tagged `early_flower`+
- [ ] Start grow cannabis + early_flower → flower program suggested; veg program flagged
- [ ] Recipe row shows crop_keys in API GET
- [ ] Phase 96 smokes pass using metadata (not name substring)
- [ ] Guardian mentions program/recipe crop fit when asked about feeding

**Prompt loop:** `phase 102 ws1` … or **`phase 102`**.
