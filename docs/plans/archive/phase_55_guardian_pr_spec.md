---
name: "Phase 55 — Guardian PR spec (ops read tools + starters)"
overview: >
  Implementation spec for Phase 55 Guardian slice: four new read enrichments for
  grow/stock/money jobs, expanded conversation starters, and ops persona copy.
  No new Confirm write tools in v1.
parent_plan: phase_55_guardian_ops_grow_money.plan.md
status: shipped
---

# Phase 55 — Guardian PR spec (ops read tools + starters)

**Parent:** [phase_55_guardian_ops_grow_money.plan.md](phase_55_guardian_ops_grow_money.plan.md)

**Prerequisites:** [phase_43_guardian_pr_spec.md](phase_43_guardian_pr_spec.md) (low-stock read + ops starters) · [phase_53_grow_stock_money_closure.plan.md](phase_53_grow_stock_money_closure.plan.md) (hub surfaces)

---

## 1. What Phase 55 adds

| Deliverable | Type | Outcome |
|-------------|------|---------|
| **`summarize_cycle_cost`** | Go read enrichment | Tagged spend per crop cycle with category breakdown |
| **`summarize_farm_spending`** | Go read enrichment | Current calendar month spend by category |
| **`restock_priority`** | Go read enrichment | Low-stock batches sorted by urgency (remaining/threshold) |
| **`summarize_active_grows`** | Go read enrichment | Active cycles listed per zone |
| **Starter expansion** | UI (`guardianStarters.js`) | Stage advice, tag receipt help, post-harvest compare/cost chips |
| **Ops persona** | Go (`context_ref.go`, `platform_context.go`) | Money/Supplies/Plants route hints reference read tools |
| **No new Confirm tools** | — | Restock, receipt capture, harvest stay hub UI |

---

## 2. Read tools

### 2.1 `summarize_cycle_cost`

| Field | Value |
|-------|--------|
| **Trigger** | Cost/spend intent + resolved cycle (context `crop_cycle_id`, zone active cycle, or cycle # in question) |
| **Query** | `GetCostTotalsByCropCycle` |
| **Block header** | `summarize_cycle_cost — cycle #{id} {name} ({zone})` |
| **Lines** | Total spent; top categories; cost/gram when yield logged |

**Banned phrases:** `cost_transactions table`, `GL mapping`

### 2.2 `summarize_farm_spending`

| Field | Value |
|-------|--------|
| **Trigger** | Month/category spend intent on Money hub or chat |
| **Query** | `GetCostCategoryTotalsByFarmForYear` (current month window) |
| **Block header** | `summarize_farm_spending — {farm} ({month})` |

### 2.3 `restock_priority`

| Field | Value |
|-------|--------|
| **Trigger** | “restock first”, “what should I restock”, “restock priority” |
| **Query** | `ListLowStockBatchesByFarm` sorted by remaining/threshold ratio |
| **Footer** | Point to Supplies **+ Add qty** — Guardian cannot PATCH batches |

Replaces `summarize_farm_low_stock` on the same turn when priority intent matches.

### 2.4 `summarize_active_grows`

| Field | Value |
|-------|--------|
| **Trigger** | “what’s growing”, “active grows”, “growing where” |
| **Query** | `ListCropCyclesByFarm` filtered `is_active` |

---

## 3. Conversation starters

| Surface | New / updated chips |
|---------|---------------------|
| Supplies hub | `restock-first` → **restock_priority** wording |
| Money hub | `tag-receipt-help` (+ spending by category uses **summarize_farm_spending**) |
| Zone grow strip | `stage-advice` (max 3) |
| Post-harvest | `how-did-we-do`, `cost-per-gram` |
| Dashboard low-stock | `open-supplies` |

Starters pass `crop_cycle_id` on zone grow context refs so **summarize_cycle_cost** resolves without the operator naming the cycle.

---

## 4. Matchers — UI wizards (no Confirm in v1)

| Operator phrase | Guardian response |
|-----------------|-------------------|
| restock / refill stock | Supplies hub **+ Add qty**; optional `create_task_from_alert` when alert in context |
| log receipt / tag spend | Money hub receipt form; no silent `createCost` |
| harvest weigh-in | Zone Overview **Harvest** button |

---

## 5. Impact previews (future Phase 46 proposals)

| Proposed action | Preview template |
|-----------------|------------------|
| Restock batch | “Adds {qty} to {input} on hand — confirm in Supplies UI today” |
| Log receipt | “Creates a {category} receipt for ${amount} — tag grow in Money hub” |

---

## 6. Acceptance

- [x] Four read tools in `ReadToolIDs()` and `EnrichPromptBlock`
- [x] Starters on Supplies, Money, grow strip, post-harvest, dashboard low-stock
- [x] `context_ref.crop_cycle_id` honored for cycle cost
- [x] Go tests in `readtools_ops_test.go`
- [x] `phase-55-closure.test.js`
