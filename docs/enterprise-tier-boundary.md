# Enterprise tier boundary

**Purpose:** Say clearly what gr33n **will not** build in the **farmer tier** by accident — so Phases 53–58 (grow, stock, money, tasks, Guardian ops) stay operator-focused.

**Status:** Published (Phase 59). **No enterprise feature code ships on `main` until this document is explicitly superseded.**

---

## Product tiers

| Tier | Audience | Status on `main` |
|------|----------|------------------|
| **Farmer** | Single-farm operator — homestead, market garden, small commercial indoor/outdoor | **Current default** — all shipped UX through Phase 58 |
| **Enterprise** | Multi-site ops, compliance, accountant-grade GL, vendor workflows | **Future** — hooks documented below; not implemented |

Read this doc before adding PO flows, traceability integrations, or multi-entity accounting to the farmer UI.

---

## In scope (farmer tier)

What we **are** building today:

- **Single-farm operator UX** — one farm context, zones, tasks, alerts, Pi edge
- **Supplies** — input definitions, batches, on-hand qty, low-stock banners, task consumptions on complete
- **Simple costs** — tagged receipts, autolog lines from mixing/electricity/consumptions, monthly Money hub summary
- **Grow** — crop cycles, harvest weigh-in, post-harvest compare, comfort targets, feeding programs
- **Guardian read tools** — walkthrough, low stock, grow advisor, Pi diagnostics (propose→Confirm for writes)
- **Edge** — per-device API keys, config sync, actuator queue when Pi is offline
- **Accountant handoff (light)** — CSV export of costs (`GET /farms/{id}/costs/export?format=csv`); optional GL-formatted export on the Advanced **Costs** page only

Farmer language: **receipts**, **batches**, **rooms/zones**, **tasks**, **spend** — not POs, SKU masters, or warehouses.

---

## Out of scope (enterprise tier — deferred)

| Capability | Why deferred | Farmer-tier alternative |
|------------|--------------|------------------------|
| **Purchase orders / vendors** | ERP complexity; approval chains; receiving vs invoice matching | Log a **receipt** on Money; restock qty on Supplies |
| **METRC / state traceability** | Separate compliance product line; state APIs; tag lifecycle | Crop cycles + harvest notes for **your** records only |
| **Multi-entity GL / chart of accounts** | Accountant tier; consolidation; period close | Simple categories + optional **accountant CSV** export |
| **Inventory valuation (FIFO/LIFO)** | Finance-grade costing; audit trails | Batch **qty remaining** + cost autolog lines |
| **SKU master / WMS** | Warehouse management; pick/pack; bin locations | **Input definitions** + **batches** per farm |
| **HR / payroll** | N/A for grow OS | Tasks for labor reminders only |
| **Multi-farm holding company rollup** | Org model TBD; cross-farm P&L | One farm per login context today |

If a phase plan proposes any row above **without** an explicit deferral to this doc, treat it as out of scope until Enterprise tier is chartered.

---

## METRC note

gr33n is **not** a cannabis compliance system. We do not integrate with METRC, BioTrack, or state traceability APIs in the farmer tier. Operators who need tag-level compliance should export their own records or use a dedicated compliance product. Harvest and cycle data in gr33n is for **operational and costing** use on your LAN — not regulatory submission.

---

## Hooks (document only — not enterprise UI)

Reserved extension points so a future Enterprise module can attach without rewriting farmer tables:

| Hook | Today | Enterprise use (future) |
|------|-------|-------------------------|
| **`batch_identifier`** on `input_batches` | Operator-visible batch label | External lot / vendor lot cross-ref |
| **`meta_data` JSONB** on `farms` (and similar pattern on other core rows) | Farm-specific flags | `external_id`, ERP sync cursor, org rollup key |
| **`organization_id` + `plan_tier`** on orgs | Schema exists; farmer UX is single-farm | Billing / multi-farm admin |
| **`scale_tier` enum** includes `enterprise` on farms | Enum value only | Feature gating when implemented |
| **Cost CSV export** | `format=csv`, `format=summary_csv`, Advanced `format=gl_csv` | Full GL bridge module |
| **API versioning** | OpenAPI `0.4.x` on farmer routes | `/v2/enterprise/...` modules behind separate auth |

No `external_id` columns are required for farmer-tier v1. When Enterprise ships, prefer additive migrations and opt-in modules — not farmer-surface jargon.

---

## Copy rules (farmer Vue)

Banned on **farmer surfaces** (Money, Supplies, Tasks, zones, dashboard, Pi setup, alerts):

- purchase order
- METRC
- general ledger
- SKU master
- warehouse

Replace with farmer language or keep accountant/GL tooling on **Advanced** routes only (`/costs`, raw CRUD). Vitest: `phase-59-closure.test.js`.

---

## Related

| Doc | Use |
|-----|-----|
| [phase_59 plan](plans/archive/phase_59_enterprise_tier_boundary.plan.md) | Implementation checklist |
| [phase_53_59 roadmap](plans/phase_53_59_roadmap.plan.md) | Farmer closure arc |
| [pre_development_gaps_index](plans/pre_development_gaps_index.plan.md) | Tier D deferrals |
| [operator-tour §7](operator-tour.md#7-supplies-feeding--money-phase-43) | Farmer money & supplies jobs |
| [cost-attribution-playbook](cost-attribution-playbook.md) | Autolog + receipts (farmer tier) |
