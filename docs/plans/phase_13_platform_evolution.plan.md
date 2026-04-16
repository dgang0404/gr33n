---
name: Phase 13 Platform evolution
overview: Phase 13 picks up after Phase 12 productization — optional receiver-side Insert Commons, stronger audit and compliance surfaces, deeper finance and offline coverage, and early multi-tenant or commercial primitives without a full ERP or marketplace rewrite.
todos:
  - id: insert-commons-receiver
    content: "WS1-Commons receiver: Stand up or integrate an Insert Commons ingest service (validate schema_version, authenticate farms, idempotent store, retention); document contract vs farm sender; optional in-repo minimal receiver for self-hosted pilots."
    status: pending
  - id: audit-compliance
    content: "WS2-Audit: Productize audit trails for sensitive actions (RBAC, receipt access, sync toggles, finance mappings, destructive ops); export/query API or operator UI hooks; align with runbook retention."
    status: pending
  - id: offline-expansion
    content: "WS3-Offline: Extend PWA queued writes beyond tasks (e.g. cost quick-add with receipt queue, or zone notes); unify idempotency patterns with server; conflict UX parity."
    status: pending
  - id: finance-depth
    content: "WS4-Finance: Deeper bookkeeping — invoices/revenue docs, tax-oriented exports, or first external accounting connector behind a narrow adapter; keep schema reversible."
    status: pending
  - id: tenancy-billing
    content: "WS5-Tenancy: Early multi-farm org or tenant boundaries, usage metering hooks, and billing/pricing experiments beyond Phase 12 primitives (no full marketplace)."
    status: pending
  - id: mobile-distribution
    content: "WS6-Mobile: Capacitor or similar packaging roadmap, push notifications strategy, and store constraints — without replacing the Vue PWA core."
    status: pending
  - id: phase13-docs
    content: "WS7-Docs: README phase banner, OpenAPI for new surfaces, operator playbooks for audit and commons receiver deploy."
    status: pending
isProject: false
---

# Phase 13 — Platform evolution

## Prerequisites

Phase 12 on **`main`** includes production-oriented storage, offline task queue, finance COA mappings + GL export, farm-side Insert Commons sender (payload, ingest URL, history), and expanded operator docs/runbooks.

## Themes (pick priority order per release train)

| Theme | Outcome |
|-------|---------|
| **Federation completion** | Receiver trusts sender contract; benchmarks become real cross-farm signal |
| **Trust and compliance** | Auditable sensitive actions; defensible exports for finance and ops |
| **Field completeness** | More offline-first journeys without compromising data integrity |
| **Commercial readiness** | Org/tenant and billing hooks without boiling the ocean |

## Explicitly still out of scope for Phase 13 (defer to later)

- Full marketplace and hardware certification programs
- Full ERP replacement or single-vendor accounting lock-in
- Greenfield native app rewrite (Capacitor **wrapper** is in scope; rewrite is not)

## Suggested execution order

1. **Audit / compliance** — unlocks safe expansion of finance and federation.
2. **Insert Commons receiver** — completes the loop started in Phase 12.
3. **Offline expansion** — one additional vertical with the same queue discipline.
4. **Finance depth** — only after audit and export paths are trusted.
5. **Tenancy / billing** — when product needs org-level boundaries or metering.
6. **Mobile distribution** — when field rollout requires store or push channels.
7. **Docs** continuously.

## Using this plan in a new chat

Reference `@docs/plans/phase_13_platform_evolution.plan.md` and specify which workstreams to implement first; adjust todos to match your release goals.
