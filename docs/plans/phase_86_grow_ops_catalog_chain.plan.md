---
name: Phase 86 — Grow ops & Guardian crop chain
overview: >
  Start grow, zone strip, Water/Light tabs, and Guardian lookup_crop_targets all
  resolve the same crop_key from catalog-bound plants; feeding follows profile stages.
todos:
  - id: ws0-deps
    content: "WS0: Phase 85 shipped — plants.crop_key on every plant row"
    status: pending
  - id: ws1-start-grow
    content: "WS1: StartGrowWizard — catalog plant required; target preview; optional variety"
    status: pending
  - id: ws2-zone-plants
    content: "WS2: Zone Plants — link plants to zone; Start grow from plant row"
    status: pending
  - id: ws3-cycle-api
    content: "WS3: Active cycle requires plant_id with crop_key; batch label optional"
    status: pending
  - id: ws4-targets-ui
    content: "WS4: Grow strip + Water/Light show EC/DLI/photoperiod from profile stage"
    status: pending
  - id: ws5-guardian
    content: "WS5: lookup_crop_targets uses cycle→plant→crop_key; DB catalog registry; no LLM EC"
    status: pending
  - id: ws6-smokes
    content: "WS6: smoke_phase86 — Flower Room cannabis path; Guardian EC matches strip"
    status: pending
isProject: false
---

# Phase 86 — Grow ops & Guardian crop chain

## Status

**Planned.** Connects **Plants** (Phase 85) to **grows**, **feeding**, **light**, and **Guardian**.

**Depends on:** [Phase 85](phase_85_catalog_bound_plants.plan.md).

**Closure:** **OC-86**

---

## The one job

> **One chain from Plants tab to Guardian:** pick **Cannabis** from the knowledge base → start **Flower run** → zone strip shows EC for **early_flower** → Water tab hints feed strength → Guardian **`lookup_crop_targets`** returns the **same numbers** (including farm EC override from Settings).

---

## End-to-end flow (Flower Room example)

1. **Zone → Plants → + Add plant** → select **Cannabis** (dropdown, targets preview visible)
2. Optional variety: "Blue Dream"
3. **Start grow** from strip or plant row → stage **early_flower**, feeding program optional
4. **Current grow strip** shows EC mS/cm chip for active stage
5. **Water** tab links reservoir target to profile stage EC
6. **Ask Guardian** “Is my EC on target for early flower?” → `lookup_crop_targets` uses `plants.crop_key=cannabis` + farm override

---

## Current gaps

| Gap | Symptom (your screenshots) |
|-----|----------------------------|
| Grow without catalog plant | Strip has no EC; Guardian says “pick a profile” |
| Picker 404 | Empty/error dropdown — API not migrated/restarted |
| `strain_or_variety` as crop identity | Conflicts with catalog semantics |
| Guardian alias registry off YAML | Mismatch if checkout stale vs DB catalog |
| Cycle without `plant_id` | Context ref cannot resolve crop |

---

## WS5 — Guardian crop API access (required)

Guardian **must not invent** EC, pH, VPD, DLI, photoperiod, or watering advice. It uses the **same crop data path as the UI**:

### Resolution order for `lookup_crop_targets`

1. **Active grow context** — `context_ref.crop_cycle_id` or zone → active cycle → `plant_id` → **`plants.crop_key`**
2. **Effective profile** — farm override row if present, else builtin — same sqlc path as picker
3. **Stage** — cycle `current_stage` or inferred from question
4. **Multi-crop / compare** — mentions resolved via **`crop_catalog_entries` + aliases** (`CROP_CATALOG_SOURCE=db`, querier wired at boot)
5. **Unsupported** — `supported=false` → honest block + `cousin_of`; **no fake EC**

### APIs Guardian aligns with (internal DB, same as HTTP)

| HTTP (UI) | Guardian internal |
|-----------|-------------------|
| `GET /farms/{id}/crop-library/picker` | Effective profiles + catalog metadata for farm |
| `GET /commons/crop-catalog/{crop_key}` | Alias / unsupported / substrate / watering |
| `PUT /farms/{id}/crop-profiles/{crop_key}` | Farm override visible on next chat turn (no re-ingest) |
| `agronomy_field_guides` (RAG) | Narrative “how to grow” — **supplement**, not replace structured targets |

### Persona rule (existing, enforce in smokes)

> NEVER state EC, pH, VPD, DLI, or photoperiod unless `lookup_crop_targets` output provides it. EC is **mS/cm**. If no plant/crop assigned, direct operator to **Zone → Plants** or **Start grow**.

### WS5 deliverables

- [ ] `resolveCropProfileContext` prefers `plants.crop_key` over free-text cycle fields
- [ ] `defaultCropRegistry()` always uses DB catalog when `CROP_CATALOG_SOURCE=db`
- [ ] Smoke: Guardian EC string == zone strip EC for same farm + stage + override
- [ ] Smoke: “compare cannabis and tomato” uses DB stage rows
- [ ] Smoke: “EC for ramps” → unsupported block, no percentages

---

## WS1 — Start grow wizard

| Rule | Detail |
|------|--------|
| Required | Farm plant with `crop_key` OR inline catalog pick (creates plant slot) |
| Picker | Same `CropLibraryPicker` — feeding & light preview before confirm |
| Optional | `variety_or_cultivar`; cycle name = batch label |
| Removed | Starting a grow with no catalog crop |

---

## WS3 — Cycle API

```json
POST /farms/{id}/crop-cycles
{
  "zone_id": 2,
  "plant_id": 5,
  "name": "Flower run (12/12)",
  "current_stage": "early_flower",
  "strain_or_variety": "Blue Dream batch A"
}
```

- `plant_id` required for `is_active=true`
- Validate `plants.crop_key` is supported
- `strain_or_variety` = batch/genetics label only

---

## WS4 — Targets in UI

| Surface | Shows |
|---------|--------|
| Zone grow strip | EC range for `current_stage` from effective profile |
| CropLibraryPicker | Stage target lines (EC, DLI, photoperiod) |
| Water tab | Feed strength hint from profile EC |
| Light tab | Photoperiod / DLI from profile stage |
| Link | “Adjust targets” → Settings → Crops & targets |

---

## Acceptance

- [ ] Flower Room: catalog cannabis → start grow → EC chip on strip
- [ ] Farm EC override in Settings → strip + Guardian match on next question
- [ ] Cannot start active grow without catalog plant
- [ ] Guardian never returns EC as “%” or generic veg/flower weeks without tool output
- [ ] `smoke_phase86` green

**Prompt loop:** `phase 86 ws5` for Guardian-only pass, or **`phase 86`** for full phase.
