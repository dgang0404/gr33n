---
name: Phase 104 — Harvest analytics by crop_key
overview: >
  Compare grows and farm summaries group by catalog crop_key — not display_name
  or batch_label — Money and zone strip aligned with knowledge base.
todos:
  - id: ws1-api
    content: "WS1: Compare/summary responses include crop_key from plant_id join"
    status: completed
  - id: ws2-aggregate
    content: "WS2: GET /farms/{id}/crop-analytics — yield/cost/EC by crop_key"
    status: completed
  - id: ws3-ui
    content: "WS3: CropCycleCompare + Money grows section — group by catalog crop"
    status: completed
  - id: ws4-guardian
    content: "WS4: Read tool summarize_farm_crops_by_key for compare questions"
    status: completed
  - id: ws5-smokes
    content: "WS5: Compare two cannabis cycles — same crop_key bucket"
    status: completed
isProject: false
---

# Phase 104 — Harvest analytics by crop_key

## Status

**Shipped** on `main`. Closure: [`phase-104-closure.md`](phase-104-closure.md) (**OC-104**).

Analytics keyed on catalog **`crop_key`** via `plant_id` — not cycle name or legacy strain fields.

**Depends on:** [Phase 86](phase_86_grow_ops_catalog_chain.plan.md), [Phase 93](phase_93_plant_identity_vocabulary_cleanup.plan.md).

**Closure:** **OC-104**

---

## The one job

> **“Compare my last two cannabis runs”** uses **`crop_key=cannabis`**, not string matching on “Flower run (12/12)”.

---

## Gap today

`GET /farms/{id}/crop-cycles/compare` returns per-cycle summaries without stable **`crop_key`** grouping. `MoneyGrowsSection` and Guardian compare starters cannot bucket by knowledge-base crop.

---

## Target

- Summary includes `crop_key`, `catalog_display_name`, `batch_label`
- Farm-level rollup: yield grams / cost / avg EC by `crop_key` and stage
- UI compare picker: filter cycles by same `crop_key`

---

## Acceptance

- [x] Post-harvest compare route pre-filters same crop_key cycles
- [x] Guardian “compare last two tomato runs” resolves via plant → crop_key

**Prompt loop:** **`phase 104`**.
