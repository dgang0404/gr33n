---
name: Phase 56 — Grow schema + harvest analytics polish
overview: >
  Small schema and UX upgrades after Phase 53 harvest flow is proven — plant_id FK,
  stage history visibility, compare pre-select from harvest, income rollup on cycle
  summary. Keeps farmer paths simple; Advanced stays for power users.
todos:
  - id: ws1-plant-fk
    content: "WS1: Migration plant_id on crop_cycles; wizard dropdown; Plants page coherence"
    status: completed
  - id: ws2-stage-history
    content: "WS2: Stage transition history on cycle summary (read existing or light audit table)"
    status: completed
  - id: ws3-compare-flow
    content: "WS3: Harvest → compare with pre-selected prior cycle; cost-per-gram card"
    status: completed
  - id: ws4-income-rollup
    content: "WS4: Cycle summary income vs spend net line when is_income receipts tagged"
    status: completed
  - id: ws5-docs-tests
    content: "WS5: migration smoke; phase-56-closure; OC-56"
    status: completed
isProject: false
---

# Phase 56 — Grow schema + harvest analytics polish

## Status

**Shipped.** After [Phase 53](phase_53_grow_stock_money_closure.plan.md) harvest wizard.

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
| API | `PUT` cycle accepts `plant_id`; list filters `?plant_id=` |

Backfill: match `strain_name` text to plant name where possible (one-time script).

---

## WS2 — Stage history

- `crop_cycle_stage_events` table; insert on create + stage PATCH
- Farmer copy: "early flower" not raw enum slug on timeline
- `stage_history_supported: true` when events exist

---

## WS3 — Compare flow

- Post-harvest + summary → `/farms/{id}/crop-cycles/compare?ids=current,prior`
- **Cost per gram** on summary yield card (unchanged contract)
- Guardian starter: `compare_ids` on post-harvest + grow strip

---

## WS4 — Income rollup

- When receipts tagged `is_income` + `crop_cycle_id`, show **Harvest economics** banner on cycle summary
- Money hub filter: `/operations/money?cycle_id=` + `GET /costs?crop_cycle_id=`

---

## WS5 — Docs, tests, OC-56

- Migration: `20260608_phase56_grow_schema_harvest.sql`
- `phase-56-closure.test.js`, `smoke_crop_cycles_test.go` Phase 56 smoke
- operator-tour §6k, architecture §7.0t

---

## Definition of done

- [x] New grows link to Plants row
- [x] Compare opens with two cycles selected
- [x] Net line on summary when income tagged
- [x] OC-56 closed

---

## Appendix — rollback notes

1. Drop `gr33nfertigation.crop_cycle_stage_events`.
2. Drop `crop_cycles.plant_id` column + index.
3. Revert handler/UI to strain-only linking (no data loss on `strain_or_variety` text).
