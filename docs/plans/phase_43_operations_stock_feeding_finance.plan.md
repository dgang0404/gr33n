---
name: Phase 43 — Operations hub (stock, feeding, money)
overview: >
  Unify inventory, NF recipes, fertigation admin, and costs/receipts into farmer-facing
  "Supplies & money" flows — zone- and job-oriented, not schema-oriented. APIs exist;
  this phase is navigation, copy, filters, and wizards.
todos:
  - id: ws1-operations-nav
    content: "WS1: Sidebar Operations group — Supplies, Feeding details, Money; collapse raw Inventory/Fertigation tabs"
    status: pending
  - id: ws2-supplies-hub
    content: "WS2: Supplies hub — batches low-stock, inputs by name, link to mix; reuse NF/inventory APIs"
    status: pending
  - id: ws3-feeding-admin
    content: "WS3: Feeding admin — programs/reservoirs/EC targets as cards; zone filter; plain irrigation vs fertigation badges"
    status: pending
  - id: ws4-money-hub
    content: "WS4: Money hub — costs, receipts, simple COA labels; link from tasks; no ledger jargon on first screen"
    status: pending
  - id: ws5-cross-links
    content: "WS5: Deep links from zone Water and 41 Fertigation context into filtered hubs"
    status: pending
  - id: ws6-guardian-ops
    content: "WS6: Guardian read tools for stock summary (optional); no new write tools unless gap found"
    status: pending
  - id: ws7-docs-tests
    content: "WS7: operator-tour §7 operations; architecture §7.0i; Vitest hub empty states; OC-43"
    status: pending
isProject: false
---

# Phase 43 — Operations hub (stock, feeding, money)

## Status

**Planned.** After [Phase 41](phase_41_farm_hub_coherence.plan.md) (`?zone_id=` and why-empty).

**Roadmap:** [farmer_ux_roadmap_40_plus.plan.md](farmer_ux_roadmap_40_plus.plan.md)

---

## Problem

Farmers who need to **restock**, **log a mix**, or **attach a receipt** land in:

- `/inventory` (batches, definitions, recipes tabs)
- `/fertigation` (six tabs mirroring tables)
- `/costs` (finance vocabulary)

Each page is correct for the schema but feels like a **different product** from the zone cockpit.

---

## Design principles

1. **Jobs not tables** — "What's running low?" not `input_batches`.
2. **Zone lens when possible** — filter by `zone_id` from 41 query param pattern.
3. **Keep APIs** — `POST /farms/{id}/fertigation/...`, inventory routes, cost routes unchanged.
4. **Org / audit stay Advanced** — not farmer-first v1.

---

## WS1 — Operations navigation

| Farmer nav | Today routes (internal) |
|------------|-------------------------|
| **Supplies** | Inventory + low-stock alerts |
| **Feeding (details)** | Fertigation programs, reservoirs, EC targets, mixing log |
| **Money** | Costs, receipts, attachments |

Redirect or wrap existing Vue views — avoid duplicating business logic.

---

## WS2 — Supplies hub

| Screen | Content |
|--------|---------|
| Home | Low-stock banner (worker already fires alerts) |
| List | Input name, qty on hand, zone/farm scope |
| Action | Log mix → links to mixing event or 39 mix-jobs |

APIs: inventory batches, input definitions, `ListLowStockBatchesByFarm`.

---

## WS3 — Feeding admin

Simplify [Fertigation.vue](../../ui/src/views/Fertigation.vue) entry:

| Card type | Farmer copy |
|-----------|-------------|
| Program | Room name, next run, irrigation-only badge |
| Reservoir | Volume bar, "ready / needs top-up" |
| EC target | Stage name + range, not table ID |

Default tab: **Programs** filtered by `?zone_id=`. Advanced: full tab bar for agronomists.

---

## WS4 — Money hub

| Screen | Content |
|--------|---------|
| Home | This month spend summary (existing aggregates if any) |
| Receipts | Photo attach flow with plain "Save receipt" |
| Detail | Link to cost transaction — hide COA until Advanced |

Reuse [cost handler](../../internal/handler/cost/) routes.

---

## WS5 — Cross-links

- Zone Water → "Stock & recipes for this room →"
- Dashboard (41) → Supplies chip when low stock

---

## WS6 — Guardian (optional)

Read-only `summarize_farm_inventory` or extend existing alerts tool — only if sit-in shows gap.

---

## WS7 — Docs, tests, closure (OC-43)

operator-tour §7, architecture §7.0i, Vitest hub filters, OC-43 row in closure doc.

---

## Out of scope

- New inventory schema
- Full accounting ERP
- Animals / aquaponics (→ 45 shells or later)

---

## Definition of done

- [ ] Operations nav group live
- [ ] Three hubs usable without reading schema names
- [ ] Zone filter on feeding/supplies
- [ ] operator-tour §7 + OC-43
