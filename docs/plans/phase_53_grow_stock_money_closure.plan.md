---
name: Phase 53 — Grow + stock + money closure
overview: >
  Close the farmer gap on Plants/crop cycles, Supplies actions, and Money tagging —
  using existing APIs only (no new tables). Includes cross-links, checklist hooks,
  post-harvest flow, nav-hint wiggles, and Guardian starters. See phase_53_59_roadmap
  for Phases 54–59.
todos:
  - id: ws1-grow-closure
    content: "WS1: Grow — zone strip, start wizard, harvest weigh-in, post-harvest screen, Plants hints"
    status: completed
  - id: ws2-stock-closure
    content: "WS2: Stock — restock, quick new batch, unit cost, low-stock→task, consumptions thin UI"
    status: completed
  - id: ws3-money-closure
    content: "WS3: Money — cycle tag, spend chip, plain autolog, zone cost peek, income receipts, energy nudge"
    status: completed
  - id: ws4-cross-links
    content: "WS4: Cross-links — v-nav-hint on all new CTAs; getting-started checklist rows; zone grow→feed→cost line"
    status: completed
  - id: ws5-guardian
    content: "WS5: Guardian starters on grow strip, Supplies, Money; cycle cost question routing"
    status: completed
  - id: ws6-docs-tests
    content: "WS6: operator-tour, architecture §7.0q, phase-53-closure.test.js, OC-53, roadmap row"
    status: completed
isProject: false
---

# Phase 53 — Grow + stock + money closure

## Status

**Shipped (WS1–WS6).** No new DB migrations for v1. Full arc: [phase_53_59_roadmap.plan.md](phase_53_59_roadmap.plan.md). **OC-53** closed.

**Predecessors:** Phase 43 (hubs) ✅ · Phase 28 (cycle analytics) ✅ · Phase 47 (Water) ✅ · Phase 52 (wiggles) ✅

---

## The three jobs

| # | Farmer job | Target |
|---|------------|--------|
| 1 | **"What's growing in this room?"** | Zone strip + start/harvest wizards |
| 2 | **"Restock what ran low"** | Supplies hub inline actions |
| 3 | **"What did this grow cost?"** | Tagged receipts + plain autolog |

---

## WS1 — Grow closure

| # | Deliverable |
|---|-------------|
| 1.1 | **ZoneCurrentGrowStrip** — name, stage, days, link to summary; empty → Start a grow |
| 1.2 | **Start grow wizard** — plant/strain → zone → optional program; `POST` crop cycle |
| 1.3 | **Harvest weigh-in** — `yield_grams`, notes, deactivate; `PATCH` cycle |
| 1.4 | **Post-harvest one-screen** — summary cards + **Compare to last cycle** (pre-fill compare URL with last harvested cycle in same zone — client-side) |
| 1.5 | **Plants page** — `EmptyStateHint`, link to wizard, strain pre-fill (text field, no FK v1) |
| 1.6 | **Zone connection line** — one row: grow → feeding plan → cost peek (read-only links) |

**APIs:** existing crop cycle + plant + summary endpoints (see prior plan table).

---

## WS2 — Stock closure

| # | Deliverable |
|---|-------------|
| 2.1 | **Restock form** — `+ Add qty` on batch card → `updateNfBatch` |
| 2.2 | **Quick new batch** — dialog when qty zero: input pick + qty + unit → `createNfBatch` |
| 2.3 | **Unit cost** — `$ / L` on input definition → `PATCH` NF input |
| 2.4 | **Low-stock → refill task** — banner CTA → `POST /tasks` with alert source |
| 2.5 | **Task consumptions (thin)** — on task complete: optional "Used X from batch Y" → existing consumptions API if wired in store; else document gap for Phase 58 |
| 2.6 | **Nav hints** — restock wiggles Supplies; low-stock wiggles Supplies + Tasks |

---

## WS3 — Money closure

| # | Deliverable |
|---|-------------|
| 3.1 | **Tag receipt to grow** — zone → active cycle dropdown → `crop_cycle_id` on `createCost` |
| 3.2 | **Income receipts** — expose `is_income` toggle on farmer form ("Sold harvest", grants) |
| 3.3 | **Today spend chip** — month spent on dashboard strip → Money hub |
| 3.4 | **Plain autolog lines** — farmer copy + tap-through to mix/feed context |
| 3.5 | **Zone cost peek** — "This grow: ~$X" from cost-summary API |
| 3.6 | **Energy price nudge** — link to Costs energy form when autolog electricity blocked |

---

## WS4 — Cross-links & checklist

| # | Deliverable |
|---|-------------|
| 4.1 | **v-nav-hint** on every new CTA (restock, harvest, tag receipt, compare, spend chip) |
| 4.2 | **Getting started checklist** — optional rows: "Start a grow", "Restock one input", "Log first receipt" (extend [firstRunChecklist.js](../../ui/src/lib/firstRunChecklist.js)) |
| 4.3 | **navRelations** — grow strip → `/plants`, `/comfort-targets`; money tag → `/operations/money`; restock → `/tasks` |
| 4.4 | **Operator guide** — link post-harvest + restock paths |

---

## WS5 — Guardian (required slice)

| Surface | Starter | Read tool |
|---------|---------|-----------|
| Zone grow strip | "What did this room cost so far?" | cycle cost summary |
| Zone grow strip | "How does this cycle compare to last time?" | compare route hint |
| Supplies hub | "What should I restock first?" | `summarize_farm_low_stock` |
| Money hub | "Summarize spending this month by category" | cost summary by category |
| Harvest flow | "What yield did we hit last run in this zone?" | prior cycle summary |

No new Confirm tools — starters send chat; matchers optional in Phase 55.

---

## WS6 — Docs, tests, OC-53

- `operator-tour.md` § grow / restock / receipt / post-harvest
- `farm-guardian-architecture.md` §7.0q
- `ui/src/__tests__/phase-53-closure.test.js`
- [farmer_ux_roadmap_40_plus.plan.md](farmer_ux_roadmap_40_plus.plan.md) + [phase_53_59_roadmap.plan.md](phase_53_59_roadmap.plan.md)
- OC-53 in operational closure

---

## Definition of done

- [x] Start grow → restock → tag receipt → see cycle cost without Advanced default path
- [x] Post-harvest lands on summary + compare link
- [x] All new links wiggle sidebar destinations
- [x] Guardian starters on three hub surfaces
- [x] No migrations in v1

---

## Deferred to later phases

| Item | Phase |
|------|-------|
| `plant_id` FK | [56](phase_56_grow_schema_harvest_analytics.plan.md) |
| Full task consumptions UI | [58](phase_58_task_consumptions_runtime.plan.md) |
| Interactive sensor→target→pump pipeline | [54](phase_54_zone_connection_nav.plan.md) |
| Guardian read-tool depth + persona | [55](phase_55_guardian_ops_grow_money.plan.md) |
| Per-device API keys | [57](phase_57_pi_device_api_keys.plan.md) |
