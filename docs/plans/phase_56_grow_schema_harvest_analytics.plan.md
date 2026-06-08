---
name: Phase 56 — Grow schema + harvest analytics polish
overview: >
  Small schema and UX upgrades after Phase 53 harvest flow is proven — plant_id FK,
  stage history visibility, compare pre-select from harvest, income rollup on cycle
  summary. Keeps farmer paths simple; Advanced stays for power users.
todos:
  - id: ws1-plant-fk
    content: "WS1: Migration plant_id on crop_cycles; wizard dropdown; Plants page coherence"
    status: pending
  - id: ws2-stage-history
    content: "WS2: Stage transition history on cycle summary (read existing or light audit table)"
    status: pending
  - id: ws3-compare-flow
    content: "WS3: Harvest → compare with pre-selected prior cycle; cost-per-gram card"
    status: pending
  - id: ws4-income-rollup
    content: "WS4: Cycle summary income vs spend net line when is_income receipts tagged"
    status: pending
  - id: ws5-docs-tests
    content: "WS5: migration smoke; phase-56-closure; OC-56"
    status: pending
isProject: false
---

# Phase 56 — Grow schema + harvest analytics polish

## Status

**Planned.** After [Phase 53](phase_53_grow_stock_money_closure.plan.md) harvest wizard ships.

**Boundary:** Not METRC/compliance — see [Phase 59](phase_59_enterprise_tier_boundary.plan.md).

---

## The one job

> **Plants, cycles, and harvest numbers tell one coherent story with real links and comparisons.**

---

## WS1 — plant_id FK

| Step | Work |
|------|------|
| Migration | `crop_cycles.plant_id` nullable FK → `plants` |
| Start grow wizard | Dropdown strains from Plants catalog (create-on-the-fly optional) |
| Plants page | Show active cycles per plant |
| API | `PATCH` cycle accepts `plant_id`; list filters |

Backfill: match `strain_name` text to plant name where possible (one-time script).

---

## WS2 — Stage history

- Surface `stage` changes on cycle summary timeline (if audit exists) or add lightweight `crop_cycle_stage_events` table
- Farmer copy: "Moved to Flower on …" not "stage enum changed"

---

## WS3 — Compare flow

- Post-harvest CTA → `/crop-cycles/compare?a={current}&b={prior}` with both IDs set
- **Cost per gram** card when yield + cost summary present
- Guardian starter: "Compare this harvest to last run" (Phase 55 tool hint)

---

## WS4 — Income rollup

- When receipts tagged `is_income` + `crop_cycle_id`, show **Net** on cycle summary
- Money hub filter: "Income for this grow"

---

## WS5 — Docs, tests, OC-56

- Migration rollback notes in plan appendix
- `phase-56-closure.test.js`
- operator-tour compare + net harvest economics paragraph

---

## Definition of done

- [ ] New grows link to Plants row
- [ ] Compare opens with two cycles selected
- [ ] Net line on summary when income tagged
- [ ] OC-56 closed
