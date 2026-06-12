---
name: Phase 72 — Money unification SPA
overview: >
  Merge the money domain — Money (farmer capture + monthly summary), Costs
  (accounting-grade ledger), and Supplies (on-hand stock + unit costs) — into one
  Money workspace with tabs. Same cost-transaction backend, one place: This month
  → Ledger → Supplies & costs. Removes the "Money / Supplies / Costs overlap" the
  operator flagged and surfaces the orphan /costs and /inventory routes that
  aren't in the sidebar today. UI-only.
todos:
  - id: ws1-summary-tab
    content: "WS1: This month tab = MoneyHub (spent/received/net, receipt capture, autologged rows with deep links)"
    status: completed
  - id: ws2-ledger-tab
    content: "WS2: Ledger tab = Costs (full transaction ledger, GL mapping, CSV/GL export, energy prices) behind progressive disclosure"
    status: completed
  - id: ws3-supplies-tab
    content: "WS3: Supplies & costs tab = SuppliesHub (on-hand batches, low-stock, restock, unit costs) + Inventory editor link; unit costs feed cost calc"
    status: completed
  - id: ws4-redirects-crosslinks
    content: "WS4: Redirect /operations/money,/costs,/operations/supplies,/inventory → /money?tab=; wire autolog deep-links to ledger/feed-water; update wiggle"
    status: completed
  - id: ws5-docs-tests
    content: "WS5: money-tabs Vitest, autolog deep-link test, phase-72-closure.test.js; operator-tour; OC-72"
    status: completed
isProject: false
---

# Phase 72 — Money unification SPA

## Status

**Shipped** on `main`. Closure: [`phase-72-closure.md`](phase-72-closure.md) (**OC-72**).

**Closure:** **OC-72** — tracked in this plan's DoD + [arc hub OC table](phase_68_73_spa_workspace_roadmap.plan.md#operational-closure-oc-rows).

---

## The one job

> **One place for the money — what I spent this month, the full ledger when I need it, and what supplies cost — instead of separate Money, Supplies, and (hidden) Costs pages.**

---

## Problem

The cost domain is split, and two of its pages aren't even in the sidebar:

| Today | Route | View | In sidebar? | Really is |
|-------|-------|------|-------------|-----------|
| Money | `/operations/money` | [`MoneyHub.vue`](../ui/src/views/MoneyHub.vue) | ✅ | Farmer: monthly spend, receipt capture, autolog |
| Costs | `/costs` | [`Costs.vue`](../ui/src/views/Costs.vue) | ❌ orphan | Accounting: ledger, GL mapping, CSV export, energy prices |
| Supplies | `/operations/supplies` | [`SuppliesHub.vue`](../ui/src/views/SuppliesHub.vue) | ✅ | On-hand batches, low-stock, restock, **unit costs** |
| Inventory | `/inventory` | [`Inventory.vue`](../ui/src/views/Inventory.vue) | ❌ orphan | Input definitions / batches / recipes editor |

All three (Money/Costs/Supplies) ride the **same cost-transaction backend** — Money is simplified capture + summary, Costs is the editor, and Supplies' unit costs feed the cost math. To the operator they "explain the same thing." And `/costs` + `/inventory` are reachable only by URL. Phase 72 makes Money one workspace and brings the orphans in as tabs.

---

## Design principles

1. **Progressive disclosure as tabs.** This month (farmer) → Ledger (accounting) → Supplies & costs. Depth is opt-in.
2. **Reuse components.** Host the existing views in tabs; no rewrite of cost logic.
3. **Keep the write-boundary.** Restock and receipts stay **inline hub actions** (not Guardian PRs) per the [Phase 55 spec](phase_55_guardian_pr_spec.md) / enterprise boundary — unchanged here.
4. **Surface the orphans.** `/costs` and `/inventory` become discoverable as tabs/links instead of URL-only.
5. **Contract-safe.** Redirects only; no schema/API change.

---

## WS1 — This month tab

- Tab body = [`MoneyHub.vue`](../ui/src/views/MoneyHub.vue): month spent / received / net, receipt capture form, autologged rows (mixes, labor, supplies, electricity) with "View →" deep links.
- Default tab on `/money`. Autolog "View →" links now jump within the workspace (to Ledger) or cross-workspace (e.g. a mix → Feed & Water Nutrients tab) with the wiggle.

---

## WS2 — Ledger tab

- Tab body = [`Costs.vue`](../ui/src/views/Costs.vue): full transaction ledger, chart-of-accounts / GL mapping, CSV/GL export, energy-price config, offline queue UI.
- Behind progressive disclosure — the accounting-grade surface power users need, no longer an orphan route.

---

## WS3 — Supplies & costs tab

- Tab body = [`SuppliesHub.vue`](../ui/src/views/SuppliesHub.vue): on-hand batches, low-stock banner, restock, unit-cost editing.
- Link to the full [`Inventory.vue`](../ui/src/views/Inventory.vue) editor (input definitions / batches / recipes) — as a sub-route or "Manage definitions" expander, bringing the second orphan in.
- Unit costs here feed the cost calculations shown on This month / Ledger — call that relationship out in the UI (the Money↔Supplies link the operator intuited).

---

## WS4 — Redirects & cross-links

- Redirects (declared Phase 68 WS4, confirm here): `/operations/money → /money?tab=summary`, `/costs → /money?tab=ledger`, `/operations/supplies → /money?tab=supplies`, `/inventory → /money?tab=supplies` (with a sub-route for the definitions editor).
- Wire autolog deep-links and update `v-nav-hint`/[`navRelations.js`](../ui/src/lib/navRelations.js): Money ↔ Supplies ↔ Feed & Water (mixes) ↔ Plants (grow-tagged spend).

> **Naming note:** "Supplies" is a real farmer job; consider keeping a top-level **Supplies** sidebar shortcut that deep-links to `/money?tab=supplies` so restock stays one click. Decide in WS4; lock in the closure test. (Default recommendation: keep the shortcut.)

---

## WS5 — Docs, tests, closure (OC-72)

| Artifact | Content |
|----------|---------|
| `ui/src/__tests__/money-tabs.test.js` (new) | Three tabs render; `?tab=` deep-link; default = summary |
| `ui/src/__tests__/money-autolog-links.test.js` (new) | Autolog "View →" resolves to ledger / cross-workspace target |
| `ui/src/__tests__/phase-72-closure.test.js` (new) | Costs + Inventory reachable as tabs; old routes redirect |
| [operator-tour.md](../operator-tour.md) | Money: one workspace; Costs/Inventory no longer hidden |

**OC-72** added and closed when WS1–WS5 ship.

---

## Out of scope

- New cost write-tools for Guardian (restock/receipt stay inline per Phase 55).
- Any schema/API/Pi change.
- Reworking GL/accounting logic — `Costs.vue` is hosted as-is.

---

## Definition of done

- [x] `/money` workspace has This month / Ledger / Supplies & costs tabs
- [x] Orphan `/costs` and `/inventory` are reachable as tabs/links, not URL-only
- [x] Supplies unit-costs' relationship to cost totals is visible
- [x] `/operations/money`, `/costs`, `/operations/supplies`, `/inventory` redirect into the workspace
- [x] Autolog deep-links resolve; wiggle connects Money↔Supplies↔Feed & Water↔Plants
- [x] Vitest green; OC-72 closed

---

## Suggested implementation order

1. WS1 This month (host MoneyHub) — default tab
2. WS3 Supplies & costs (host SuppliesHub + Inventory link)
3. WS2 Ledger (host Costs — de-orphan)
4. WS4 redirects + cross-links + Supplies shortcut decision
5. WS5 closure

---

## Related

| Doc | Use |
|-----|-----|
| [phase_68_workspace_shell_spa_nav.plan.md](phase_68_workspace_shell_spa_nav.plan.md) | Workspace shell + declared tabs |
| [phase_53_grow_stock_money_closure.plan.md](phase_53_grow_stock_money_closure.plan.md) | Money/stock farmer closure |
| [phase_55_guardian_pr_spec.md](phase_55_guardian_pr_spec.md) | Why restock/receipts stay inline, not PRs |
| [phase_59_enterprise_tier_boundary.plan.md](phase_59_enterprise_tier_boundary.plan.md) | Accounting boundary |
| [ui/src/views/Costs.vue](../ui/src/views/Costs.vue) | De-orphaned into Ledger tab |

---

## Using this in a new chat

> Read `docs/plans/phase_72_money_unification.plan.md`. UI-only. Merge MoneyHub + Costs + SuppliesHub (+ Inventory) into the `/money` workspace as tabs (This month → Ledger → Supplies & costs). De-orphan /costs and /inventory. Redirect old routes. Keep restock/receipts inline (no new Guardian write tools).
