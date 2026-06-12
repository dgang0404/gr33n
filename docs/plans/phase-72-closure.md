# Phase 72 — closure (OC-72)

**Status:** **Shipped** on `main`.

**Canonical plan:** [`phase_72_money_unification.plan.md`](phase_72_money_unification.plan.md)

**Depends on:** [Phase 68](phase_68_workspace_shell_spa_nav.plan.md) workspace shell; [Phase 53](phase_53_grow_stock_money_closure.plan.md) money/stock farmer closure.

**Closes:** One `/money` workspace for monthly spend, accounting ledger, and supplies — de-orphaning `/costs` and `/inventory`.

---

## The one job (done)

> **One place for farm money** — what you spent this month, the full ledger when you need it, and supply unit costs — instead of hidden Costs pages and scattered operations routes.

---

## Workstream checklist

| WS | Deliverable | Verify |
|----|-------------|--------|
| **WS1** | This month = `MoneyHub` | `MoneyWorkspace.vue` |
| **WS2** | Ledger = `Costs` | same |
| **WS3** | Supplies & costs = `SuppliesHub` + `Inventory` | same; unit-cost copy in `SuppliesHub.vue` |
| **WS4** | Legacy redirects + autolog deep-links | `workspaces.js`, `moneyHub.js`, `MoneyHub.vue` |
| **WS5** | Tests | `money-tabs.test.js`, `money-autolog-links.test.js`, `phase-72-closure.test.js` |

---

## Routes

| Legacy | Target |
|--------|--------|
| `/operations/money` | `/money?tab=summary` |
| `/costs` | `/money?tab=ledger` |
| `/operations/supplies` | `/money?tab=supplies` |
| `/inventory` | `/money?tab=inventory` |

Autolog **View →** links: mixes → Feed & Water nutrients; energy → Money ledger.

---

## Automated tests

| Test | Path |
|------|------|
| Tab model + workspace hosting | `ui/src/__tests__/money-tabs.test.js` |
| Autolog deep-link targets | `ui/src/__tests__/money-autolog-links.test.js` |
| Redirects + UI copy | `ui/src/__tests__/phase-72-closure.test.js` |

---

## OC-72

Phase 72 is **closed** when `/money` hosts all money-domain views, orphan routes redirect correctly, autolog links stay in-workspace, and supplies unit-cost relationship is visible in the UI.
