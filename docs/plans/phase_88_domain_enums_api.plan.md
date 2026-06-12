---
name: Phase 88 — Domain enums API
overview: >
  GET /platform/domain-enums serves growth stages, reservoir status, cost categories,
  and NF inventory enums from one source; UI removes duplicate hardcoded arrays.
todos:
  - id: ws1-api
    content: "WS1: GET /platform/domain-enums — growth_stages, reservoir_status, cost_categories, …"
    status: completed
  - id: ws2-openapi
    content: "WS2: OpenAPI schema + smoke test"
    status: completed
  - id: ws3-ui-loader
    content: "WS3: ui/lib/domainEnums.js — fetch once, cache in pinia or farm store"
    status: completed
  - id: ws4-migrate-ui
    content: "WS4: growHub, SetpointRow, Fertigation, Fertigation reservoir, Inventory, moneyHub"
    status: completed
  - id: ws5-guardian
    content: "WS5: Document alignment with croplibrary.ValidGrowthStages + persona"
    status: completed
  - id: ws6-parity-link
    content: "WS6: Phase 99 check-ui-domain-parity — SetpointRow count guard"
    status: completed
isProject: false
---

# Phase 88 — Domain enums API

## Status

**Shipped.** Foundation for keeping **UI forms** aligned with **Postgres enums** and **Guardian stage vocabulary**.

**Closure:** [`phase-88-closure.md`](phase-88-closure.md) · **OC-88**

---

## The one job

> **One HTTP call** returns every platform enum the UI needs for dropdowns — growth stages first (11 values, including `transition` and `flush`).

---

## Bug today (operator impact)

`SetpointRow.vue` default `stageOptions` has **9** stages — missing **`transition`** and **`flush`**:

```99:99:ui/src/components/SetpointRow.vue
    default: () => ['clone', 'seedling', 'early_veg', 'late_veg', 'early_flower', 'mid_flower', 'late_flower', 'harvest', 'dry_cure'],
```

`Fertigation.vue` duplicates the full list inline instead of importing `growHub.js`.

---

## API shape (proposed)

```
GET /platform/domain-enums
```

```json
{
  "growth_stages": [
    { "value": "early_flower", "label": "early flower" }
  ],
  "reservoir_statuses": [ … ],
  "cost_categories": [ … ],
  "application_targets": [ … ],
  "input_definition_categories": [ … ],
  "batch_statuses": [ … ]
}
```

**Source of truth:** Go maps generated from same lists as OpenAPI / sqlc enums (single package e.g. `internal/platform/domainenums`).

**Labels:** Humanized from snake_case; optional `operator_label` later.

---

## UI migration (WS4)

| File | Change |
|------|--------|
| `lib/growHub.js` | `GROWTH_STAGES` from API cache; keep `formatStageLabel` |
| `components/SetpointRow.vue` | `stageOptions` from domain enums |
| `views/Fertigation.vue` | Remove inline array |
| `views/Fertigation.vue` reservoir select | `reservoir_statuses` from API |
| `lib/moneyHub.js` | Full `cost_categories`; separate income GL mapping |
| `views/Inventory.vue` | NF enums from API |

**Fallback:** If API unavailable, use bundled snapshot in `domainEnums.fallback.js` (generated at build from OpenAPI — optional).

---

## Guardian (WS5)

- `lookup_crop_targets` already uses DB stage enum — no change required
- Persona: stage names in chat match API labels
- Smoke: setpoint created with `transition` stage persists and displays

---

## Acceptance

- [x] All 11 growth stages in every stage dropdown
- [x] Single `formatStageLabel` / no Fertigation duplicate
- [x] Money hub shows full cost category list from API
- [x] `smoke_phase88` or extend existing enum parity test

**Prompt loop:** `phase 88 ws1` … or **`phase 88`**.
