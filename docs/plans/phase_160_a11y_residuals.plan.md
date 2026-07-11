---
name: Phase 160 — a11y residuals + advisory CI guard
overview: >
  Close Phase 158 deferred items: lighting modal dialog semantics, key form
  labels, mobile nav drawer focus trap, and a non-blocking Vitest a11y smoke.
todos:
  - id: ws1-lighting-modal
    content: "WS1: ZoneLightingEditor modal — role=dialog, aria-labelledby, useDialogFocusTrap, Escape"
    status: completed
  - id: ws2-lighting-form-labels
    content: "WS2: LightingProgramForm — for/id on preset name, photoperiod hours, actuator select"
    status: completed
  - id: ws3-mobile-drawer-trap
    content: "WS3: App.vue mobile hamburger drawer — focus trap + aria-label on close"
    status: completed
  - id: ws4-advisory-axe
    content: "WS4: Vitest advisory check — eslint-plugin-vuejs-accessibility or static closure expanded (no CI block)"
    status: completed
  - id: ws5-audit-update
    content: "WS5: Update a11y-audit doc — move D2/D3/D5 to closed; note D4 advisory"
    status: completed
  - id: ws6-closure
    content: "WS6: phase-160-closure.test.js"
    status: completed
isProject: false
---

# Phase 160 — a11y residuals + advisory CI guard

**Status:** shipped · **Hub:** [159–160 backlog](phase_159_160_post_158_gaps_backlog.plan.md) · **Continues:** [158](phase_158_accessibility_pass.plan.md)

## Shipped

| WS | Deliverable |
|----|-------------|
| **WS1** | `ZoneLightingEditor` modal — dialog + focus trap + Escape |
| **WS2** | `LightingProgramForm` + `PhotoperiodClockEditor` — `for`/`id`, `aria-pressed`, error `role="alert"` |
| **WS3** | Mobile drawer focus trap + labelled close |
| **WS4** | Advisory `phase-160-closure.test.js` (no new CI blocker) |
| **WS5** | [`a11y-audit-2026-07-11.md`](../a11y-audit-2026-07-11.md) updated |
| **WS6** | Closure test |

## Close when

- [x] Lighting editor modal: keyboard trap + Escape + labelled dialog
- [x] Lighting form: labelled name + photoperiod + actuator select
- [x] Mobile drawer: trap + labelled close
- [x] Advisory a11y check runnable locally
- [x] `phase-160-closure.test.js` passes

## Still deferred

- Sidebar roving `tabindex` (D1)
- Blocking axe in GitHub CI (D4)
