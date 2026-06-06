---
name: Phase 43 — Operations hub (stock, feeding, money)
overview: >
  Unify inventory, NF recipes, fertigation admin, and costs/receipts into farmer-facing
  "Supplies & money" flows — zone- and job-oriented, not schema-oriented. APIs exist;
  this phase is navigation, copy, filters, and wizards.
todos:
  - id: ws1-operations-nav
    content: "WS1: Sidebar Operations group — Supplies, Feeding details, Money; collapse raw Inventory/Fertigation tabs"
    status: completed
  - id: ws2-supplies-hub
    content: "WS2: Supplies hub — batches low-stock, inputs by name, link to mix; reuse NF/inventory APIs"
    status: completed
  - id: ws3-feeding-admin
    content: "WS3: Feeding admin — programs/reservoirs/EC targets as cards; zone filter; plain irrigation vs fertigation badges"
    status: completed
  - id: ws4-money-hub
    content: "WS4: Money hub — costs, receipts, simple COA labels; link from tasks; no ledger jargon on first screen"
    status: completed
  - id: ws5-cross-links
    content: "WS5: Deep links from zone Water and 41 Fertigation context into filtered hubs"
    status: completed
  - id: ws6-guardian-ops
    content: "WS6: Guardian persona + impact for supplies/feeding/money vocabulary"
    status: completed
  - id: ws7-docs-tests
    content: "WS7: operator-tour §7 + §6f; architecture §7.0i; Vitest hub empty states; OC-43"
    status: completed
  - id: ws8-guardian-pr-slice
    content: "WS8: phase_43_guardian_pr_spec — summarize_farm_low_stock read + ops starters (no new Confirm tools)"
    status: pending
isProject: false
---

# Phase 43 — Operations hub (stock, feeding, money)

## Status

**WS1–WS7 shipped** on `main`. **WS8** (Guardian `summarize_farm_low_stock` + ops starter chips) pending — [phase_43_guardian_pr_spec.md](phase_43_guardian_pr_spec.md).

**Roadmap:** [farmer_ux_roadmap_40_plus.plan.md](farmer_ux_roadmap_40_plus.plan.md)

**Guardian slice (doc complete):** [phase_43_guardian_pr_spec.md](phase_43_guardian_pr_spec.md) — low-stock read + ops starters; no new Confirm tools.

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

**Daily feeding UX:** owned by [Phase 47](phase_47_feeding_water_plain_language.plan.md) (zone Water + Feeding hub). This WS is **farm-wide admin** — recipes, reservoirs, mixing log, bulk program edit.

Simplify [Fertigation.vue](../../ui/src/views/Fertigation.vue) as **Advanced / Operations** entry:

| Card type | Farmer copy |
|-----------|-------------|
| Program | Room name, next run, irrigation-only badge |
| Reservoir | Volume bar, "ready / needs top-up" |
| EC target | Stage name + range, not table ID |

Default: link from 47 Feeding hub → "Technical feeding admin →". Full tab bar for agronomists only.

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

## WS6 — Guardian persona

- `platform_context.go`: **Supplies**, **Feeding details**, **Money** — not Inventory / Fertigation / Costs in operator-facing replies.
- `impact.go`: verify `create_task_from_alert` copy when source is `inventory_low_stock`.

**Implementation spec (read + starters):** [phase_43_guardian_pr_spec.md](phase_43_guardian_pr_spec.md)

---

## WS7 — Docs, tests, closure (OC-43)

operator-tour §7 (operations hubs) + §6f (Guardian on supplies/money), architecture §7.0i, Vitest hub filters, OC-43 row in closure doc.

---

## WS8 — Guardian PR slice

| Item | Owner |
|------|--------|
| `summarize_farm_low_stock` read enrichment | Backend — spec §2 |
| Starters on supplies / feeding / money / dashboard | UI — spec §3 |
| Optional `pickAlertForIntent` for refill phrases | Backend — spec §4.2 |
| **No** new Confirm tools for batch/cost writes | — |

Starters ≠ Confirm. Matchers for stock PATCH deferred to Phase 46 if sit-in demands it.

---

## Out of scope

- New inventory schema
- Full accounting ERP
- Animals / aquaponics (→ 45 shells or later)

---

## Definition of done

- [x] Operations nav group live
- [x] Three hubs usable without reading schema names
- [x] Zone filter on feeding/supplies (`?zone_id=`)
- [x] Guardian persona + impact (WS6)
- [x] operator-tour §7 + §6f + OC-43 (WS7)
- [ ] Guardian: [phase_43_guardian_pr_spec.md](phase_43_guardian_pr_spec.md) WS8 complete

## Related

| Doc | Use |
|-----|-----|
| [phase_43_guardian_pr_spec.md](phase_43_guardian_pr_spec.md) | Low-stock read + ops starters |
| [guardian_pr_ux_through_farmer_phases.plan.md](guardian_pr_ux_through_farmer_phases.plan.md) | Cross-phase PR table |
