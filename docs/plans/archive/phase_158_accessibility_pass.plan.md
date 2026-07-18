---
name: Phase 158 — accessibility pass (core workspaces)
overview: >
  Keyboard navigation, focus management, and screen-reader labels for Guardian
  chat and the highest-traffic farmer workspaces.
todos:
  - id: ws1-audit
    content: "WS1: Baseline audit — axe-core or eslint-plugin-vuejs-accessibility on GuardianChatPanel + zone cockpit"
    status: completed
  - id: ws2-guardian-chat
    content: "WS2: Guardian chat — focus trap in panel, cite chips as links/buttons with aria-label, accuracy banner role=alert"
    status: completed
  - id: ws3-nav-keyboard
    content: "WS3: Workspace shell — skip link, sidebar roving tabindex, visible focus rings"
    status: completed
  - id: ws4-forms
    content: "WS4: High-traffic forms — zone Water/Light, proposal Confirm card, model selector"
    status: completed
  - id: ws5-ci-guard
    content: "WS5: Optional Vitest a11y smoke or axe in CI (advisory first)"
    status: completed
isProject: false
---

# Phase 158 — accessibility pass (core workspaces)

**Status:** shipped · **Hub:** [154–158 backlog](phase_154_158_infra_trust_gaps_backlog.plan.md)

## Shipped

| WS | Deliverable |
|----|-------------|
| **WS1** | [`a11y-audit-2026-07-11.md`](../a11y-audit-2026-07-11.md) — P0/P1/P2 issue list |
| **WS2** | Guardian chat: citation `aria-label`, accuracy `role="alert"`, textarea label, proposal focus, drawer focus trap |
| **WS3** | Skip link, `main#main-content`, `aria-current` on sidebar + mobile nav |
| **WS4** | Zone tablist semantics, proposal high-risk alert, settings model `aria-live` |
| **WS5** | `ui/src/__tests__/phase-158-closure.test.js` |

## Close when

- [x] Audit doc exists with P0 issues tracked to closure or wontfix
- [x] Guardian chat: citation links and accuracy banner ship a11y attributes
- [x] Skip link + sidebar keyboard nav on `/today` and `/zones/:id`
- [x] Proposal Confirm reachable and activatable without mouse
- [x] `phase-158-closure.test.js`

## Deferred

See audit doc § Deferred — roving tabindex, lighting form labels, axe in CI (blocking).
