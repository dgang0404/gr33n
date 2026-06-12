---
name: Phase 96 — Grow feeding program validation
overview: >
  v1 warn/block when fertigation program mismatches crop_key or growth stage at
  start grow; Phase 102 adds program+recipe metadata — 96 ships warnings first.
todos:
  - id: ws0-deps
    content: "WS0: Phase 86 — cycle has plant_id + crop_key + current_stage"
    status: pending
  - id: ws1-rules-v1
    content: "WS1: v1 rules — stage/crop heuristics until Phase 102 meta seeded"
    status: pending
  - id: ws2-api
    content: "WS2: POST crop-cycles warns/blocks primary_program_id mismatch"
    status: pending
  - id: ws3-ui
    content: "WS3: Start grow + Water tab — mismatch banner (veg program + flower stage)"
    status: pending
  - id: ws4-guardian
    content: "WS4: Guardian prompt block — profile EC vs pump recipe may differ"
    status: pending
  - id: ws5-smokes
    content: "WS5: smoke — flower stage + veg JLF program → warning visible"
    status: pending
  - id: ws6-phase102-handoff
    content: "WS6: Hand off to Phase 102 — validation reads program.meta + recipe.meta"
    status: pending
isProject: false
---

# Phase 96 — Grow feeding program validation

## Status

**Planned.** Closes **blind spot #7** (EC strip says flower; pump runs veg recipe).

**Depends on:** [Phase 86](phase_86_grow_ops_catalog_chain.plan.md).

**Long-term metadata:** [Phase 102](phase_102_fertigation_program_catalog_metadata.plan.md) — **recipe ↔ crop profile linkage**.

**Closure:** **OC-96**

---

## The one job

> **Attach-time guardrail:** if grow stage or `crop_key` doesn’t fit the linked fertigation program (and its recipe), show a **clear warning** before the operator confirms — Guardian says the same thing in chat.

---

## Blind spot #7

| Layer | Can show |
|-------|----------|
| Zone strip | Flower EC from **crop profile** stage |
| Fertigation | Veg JLF **program + recipe** (untagged today) |

---

## Phase 96 vs Phase 102

| Phase | Role |
|-------|------|
| **96 (this)** | **Behavior** — warn/block on mismatch at Start grow + Water tab + Guardian |
| **102** | **Data** — `crop_key` + stage tags on programs and `application_recipes`; EC band from profile |

Ship **96 first** with heuristics; **102** replaces heuristics with metadata (WS6 handoff).

---

## WS1 — v1 validation rules (before Phase 102)

Until program meta exists:

| Signal | Rule |
|--------|------|
| Program name contains `veg` / `JLF` | Assume vegetative stages |
| Program name contains `flower` / `FFJ` | Assume flower stages |
| `cycle.current_stage` in flower enum | Mismatch if veg program |
| `plants.crop_key` | No crop filter in v1 (102 adds) |

Env `STRICT_PROGRAM_STAGE_MATCH=1` → **422** instead of warning response field.

---

## WS6 — Phase 102 handoff

When [Phase 102](phase_102_fertigation_program_catalog_metadata.plan.md) ships:

- Validation reads `fertigation_programs.meta.recommended_crop_keys` + `recommended_stages`
- Recipe check via `application_recipes.meta.crop_keys`
- Compare `ec_band_mscm` to effective `crop_profile_stages` for active stage
- Remove name heuristics from WS1

---

## WS4 — Guardian

When snapshot shows active cycle + program mismatch:

> "Your grow is in **early_flower** but the linked program **Veg JLF** targets vegetative stages. EC on the zone strip comes from the **crop profile**; the **pump recipe** may differ. See Water tab or switch program."

After Phase 102: cite program `recommended_crop_keys` and recipe name.

---

## Acceptance

- [ ] Start grow with mismatched program shows visible warning before confirm
- [ ] Water tab links to program edit
- [ ] Guardian mentions mismatch when asked about feeding
- [ ] Phase 102 WS7 reuses same validation functions with metadata

**Prompt loop:** **`phase 96`** (ship before or parallel to 102 WS1).
