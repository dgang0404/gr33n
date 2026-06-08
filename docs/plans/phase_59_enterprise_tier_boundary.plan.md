---
name: Phase 59 — Enterprise tier boundary (doc-only)
overview: >
  Explicit product gate — what gr33n is NOT building in the farmer tier.
  Prevents scope creep into ERP, METRC, POs, multi-entity GL. Documents
  future "Enterprise" tier hooks without implementing them.
todos:
  - id: ws1-boundary-doc
    content: "WS1: docs/enterprise-tier-boundary.md — in/out table, API stubs, METRC note"
    status: pending
  - id: ws2-readme-gaps
    content: "WS2: README + pre_development_gaps_index + phase_53_59_roadmap pointers"
    status: pending
  - id: ws3-ui-copy-audit
    content: "WS3: Ban ERP jargon in farmer surfaces — grep pass checklist in closure test"
    status: pending
  - id: ws4-oc-59
    content: "WS4: OC-59 in operational closure; phase-59-closure.test.js"
    status: pending
isProject: false
---

# Phase 59 — Enterprise tier boundary

## Status

**Planned.** Can ship anytime — **no feature code required** for v1.

---

## The one job

> **Say clearly what we won't build by accident — so Phases 53–58 stay farmer-focused.**

---

## WS1 — Boundary document

Create `docs/enterprise-tier-boundary.md`:

### In scope (farmer tier)

- Single-farm operator UX
- Supplies batches, simple costs, crop cycles
- Pi edge, Guardian read tools
- Tagged receipts, autolog lines

### Out of scope (enterprise tier — future)

| Capability | Why deferred |
|------------|--------------|
| Purchase orders / vendors | ERP complexity; farmer uses receipts |
| METRC / state traceability | Compliance product line |
| Multi-entity GL / chart of accounts | Accountant tier |
| Inventory valuation (FIFO/LIFO) | Batch qty sufficient for v1 |
| HR / payroll | N/A |
| Multi-farm holding company rollup | Org model TBD |

### Hooks (document only)

- `external_id` fields reserved on cycles/batches
- Export CSV for accountant
- API versioning note for future enterprise modules

---

## WS2 — Index updates

- README "Product tiers" subsection
- [pre_development_gaps_index.plan.md](pre_development_gaps_index.plan.md) — enterprise row
- [phase_53_59_roadmap.plan.md](phase_53_59_roadmap.plan.md) — boundary callout

---

## WS3 — Copy audit checklist

Closure test greps farmer Vue for banned terms:

- "purchase order", "METRC", "general ledger", "SKU master", "warehouse"

Replace with farmer language or hide behind Advanced flag.

---

## WS4 — OC-59

- Doc published + indexes linked
- No open "accidental ERP" gaps in phase-53–58 plans without explicit deferral row

---

## Definition of done

- [ ] enterprise-tier-boundary.md on main
- [ ] README points to it
- [ ] OC-59 closed
