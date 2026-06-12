---
name: Phase 102 — Fertigation program catalog metadata
overview: >
  Programs tagged with recommended crop_keys and growth stages — feeds Phase 96
  validation, Start grow suggestions, and Guardian feeding advice.
todos:
  - id: ws1-schema
    content: "WS1: fertigation_programs.meta recommended_crop_keys + recommended_stages JSONB"
    status: pending
  - id: ws2-seed
    content: "WS2: Seed demo programs (veg JLF, flower FFJ+WCA) with tags from master_seed"
    status: pending
  - id: ws3-api
    content: "WS3: GET programs includes tags; filter by crop_key + stage"
    status: pending
  - id: ws4-ui
    content: "WS4: Start grow program picker filtered by selected plant crop_key + stage"
    status: pending
  - id: ws5-phase96
    content: "WS5: Phase 96 validation uses metadata not name heuristics"
    status: pending
isProject: false
---

# Phase 102 — Fertigation program catalog metadata

## Status

**Planned.** Extends [Phase 96](phase_96_grow_feeding_program_validation.plan.md) with real metadata instead of name guessing.

**Depends on:** [Phase 86](phase_86_grow_ops_catalog_chain.plan.md).

**Closure:** **OC-102**

---

## The one job

> **Fertigation programs declare** which `crop_key`(s) and **growth stages** they target — Start grow suggests the right recipe; Phase 96 warns on mismatch.

---

## Example metadata

```json
{
  "recommended_crop_keys": ["cannabis", "tomato"],
  "recommended_stages": ["early_veg", "late_veg"],
  "ec_band_mscm": { "min": 1.0, "max": 1.6 }
}
```

---

## Acceptance

- [ ] Start grow with cannabis + early_flower hides veg-only programs (or flags them)
- [ ] Guardian cites program tags when recommending feed changes

**Prompt loop:** **`phase 102`**.
