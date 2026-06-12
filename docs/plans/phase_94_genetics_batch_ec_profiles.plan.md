---
name: Phase 94 — Genetics & batch EC profiles
overview: >
  Per-genetics EC overrides beyond farm-wide crop_key Settings — closes operator
  expectation gap for Blue Dream vs OG Kush.
todos:
  - id: ws0-scope
    content: "WS0: Phase 87 runbook — v1 = farm crop_key only; link to Phase 94 for genetics"
    status: completed
  - id: ws1-schema
    content: "WS1: plant_genetics_profiles link table + synthetic genetics crop_key"
    status: completed
  - id: ws2-api
    content: "WS2: PUT genetics profile; effective profile resolution order documented"
    status: completed
  - id: ws3-ui
    content: "WS3: Plants workspace — 'Tune EC for this variety' link"
    status: completed
  - id: ws4-guardian
    content: "WS4: lookup_crop_targets prefers genetics override when variety linked"
    status: completed
  - id: ws5-smokes
    content: "WS5: Farm cannabis override vs genetics override precedence smoke"
    status: completed
isProject: false
---

# Phase 94 — Genetics & batch EC profiles

## Status

**Shipped (OC-94).** Closure: [`phase-94-closure.md`](phase-94-closure.md).

**Depends on:** [Phase 87](phase_87_crop_knowledge_operator_closure.plan.md) (v1 docs must ship first).

**Closure:** **OC-94**

---

## Blind spot #3

**Today (v1 — Phase 83):** Settings → Crops & targets overrides EC for **`crop_key`** on the whole farm.

**Operator expectation:** “This Blue Dream batch runs hotter EC than Wedding Cake.”

**v1 answer (Phase 87 runbook):** All grows of **cannabis** on this farm share one EC table unless you use **separate farms** or wait for Phase 94.

---

## Resolution order (target)

```
1. genetics_profile override (plant.variety_or_cultivar linked profile) — if set
2. farm crop_key override (Settings)
3. platform builtin profile
```

Same chain for UI strip, picker preview, and **`lookup_crop_targets`**.

---

## Schema options (pick one in WS1)

| Option | Pros |
|--------|------|
| **A** `plant_genetics_profiles` table keyed by `(farm_id, crop_key, variety_slug)` | Reuse across plants |
| **B** `plants.profile_override_id` FK → farm-scoped profile clone | Simple per-slot |

---

## Acceptance

- [x] Phase 87 runbook states v1 vs v2 clearly
- [x] When genetics override set, Guardian EC ≠ farm-wide cannabis override
- [x] Precedence documented in architecture §7.0af

**Prompt loop:** **`phase 94`**.
