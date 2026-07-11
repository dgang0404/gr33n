---
name: Phase 157 — docs consolidation
overview: >
  Single current-state snapshot, plans archive for closed phases 88-92, trimmed
  phase-14 index, and make docs-current-state-hint for regeneration counts.
todos:
  - id: ws1-current-state
    content: "WS1: docs/current-state.md"
    status: completed
  - id: ws2-archive-folder
    content: "WS2: docs/plans/archive/ + stubs for phases 88-92"
    status: completed
  - id: ws3-index-trim
    content: "WS3: phase-14 Start here + condensed Quick links"
    status: completed
  - id: ws4-regen-hint
    content: "WS4: make docs-current-state-hint"
    status: completed
isProject: false
---

# Phase 157 — docs consolidation

**Status:** shipped · **Hub:** [154–158 backlog](phase_154_158_infra_trust_gaps_backlog.plan.md)

## Shipped

| WS | Deliverable |
|----|-------------|
| **WS1** | [`current-state.md`](../current-state.md) — what gr33n looks like today |
| **WS2** | [`plans/archive/`](../plans/archive/) — phases 88–92 moved; stubs at old paths |
| **WS3** | [`phase-14-operator-documentation.md`](../phase-14-operator-documentation.md) — Start here + arc hubs; Quick links ~80→12 rows |
| **WS4** | `make docs-current-state-hint` |

## Operator path

README → **current-state.md** → operator-tour → first-session-after-clone

## Close when

- [x] `current-state.md` linked from README + INSTALL + phase-14
- [x] Archive contains phases 88–92 with working stubs
- [x] Quick links table trimmed
- [x] `ui/src/__tests__/phase-157-closure.test.js`
