---
name: Phase 158 — accessibility pass (core workspaces)
overview: >
  Keyboard navigation, focus management, and screen-reader labels for Guardian
  chat and the highest-traffic farmer workspaces. Lower priority than 154–156
  but important for field use (gloves, bright sun, one-handed operation).
todos:
  - id: ws1-audit
    content: "WS1: Baseline audit — axe-core or eslint-plugin-vuejs-accessibility on GuardianChatPanel + zone cockpit"
    status: pending
  - id: ws2-guardian-chat
    content: "WS2: Guardian chat — focus trap in panel, cite chips as links/buttons with aria-label, accuracy banner role=alert"
    status: pending
  - id: ws3-nav-keyboard
    content: "WS3: Workspace shell — skip link, sidebar roving tabindex, visible focus rings"
    status: pending
  - id: ws4-forms
    content: "WS4: High-traffic forms — zone Water/Light, proposal Confirm card, model selector"
    status: pending
  - id: ws5-ci-guard
    content: "WS5: Optional Vitest a11y smoke or axe in CI (advisory first)"
    status: pending
isProject: false
---

# Phase 158 — accessibility pass (core workspaces)

**Status:** planned · **Priority:** lower than [154](phase_154_test_suite_health.plan.md)–[156](phase_156_dependency_scanning.plan.md) · **Hub:** [154–158 backlog](phase_154_158_infra_trust_gaps_backlog.plan.md)

---

## Why this phase

Guardian answer quality (Phases 143–153) optimized **trust in text**. Accessibility optimizes **who can use the UI**:

- Farmers in greenhouses (bright backlight, gloves, one hand on a hose)
- Operators who navigate by keyboard after RSI or preference
- Screen-reader users reviewing alerts and proposal Confirm cards

No dedicated a11y pass has run on the SPA workspaces (68–81 arc) or post-152 Guardian chat (citation deep links, accuracy banner). This phase closes that gap without a full WCAG certification project.

---

## Workstreams

### WS1 — Baseline audit

Run automated scan on:

- `ui/src/components/GuardianChatPanel.vue`
- Zone cockpit / Today dashboard entry components
- Workspace shell / sidebar (`phase_68` nav)

**Tools (pick at implement time):**

- `@axe-core/cli` against `make dev-auth-test` static build, or
- `eslint-plugin-vuejs-accessibility` in `ui/` lint lane

Deliverable: `docs/a11y-audit-YYYY-MM-DD.md` — issue list prioritized P0/P1/P2.

### WS2 — Guardian chat

| Issue | Target fix |
|-------|------------|
| Citation chips with `router-link` | `aria-label` including source type + subject excerpt |
| Accuracy banner (`data-test="chat-accuracy-banner"`) | `role="alert"` or `role="status"` per severity |
| Chat input / send | Label associated with textarea; Enter vs Shift+Enter documented for SR |
| Proposal Confirm cards | Focus moves to card when proposal arrives; Confirm/Cancel are reachable by Tab |
| Model selector | Combobox pattern or native `<select>` with visible label |

### WS3 — Workspace shell

- **Skip to main content** link (first focusable element)
- Sidebar: roving `tabindex` on nav items; `aria-current="page"` on active route
- Focus ring visible on all interactive elements (don't remove outline without replacement)
- Mobile bottom nav: touch targets ≥ 44px (may already be close — verify in audit)

### WS4 — High-traffic forms

Priority surfaces:

- Zone → **Water / Light / Climate** tabs (Phase 38–40)
- Guardian **proposal Confirm** flow (Phase 30/46)
- **Settings → Guardian** model picker (Phase 111)

Fixes: `<label for=…>`, error messages tied via `aria-describedby`, disabled state not conveyed by color alone.

### WS5 — CI guard (advisory)

- Vitest + `jest-axe` or Playwright `axe` on 2–3 golden routes (login → Today → open Guardian)
- **Advisory** in CI first (annotation, don't block) — upgrade to blocking after P0 list is empty

---

## Acceptance

- [ ] Audit doc exists with P0 issues tracked to closure or wontfix
- [ ] Guardian chat: citation links and accuracy banner pass axe with 0 critical violations
- [ ] Skip link + sidebar keyboard nav work on `/today` and `/zones/:id`
- [ ] Proposal Confirm reachable and activatable without mouse
- [ ] `ui/src/__tests__/phase-158-closure.test.js` — closure test listing shipped a11y attributes (matches project convention)

## Non-goals

- Full WCAG 2.2 AA certification or VPAT
- Mobile native app (Capacitor) store accessibility review — see [`mobile-distribution.md`](../mobile-distribution.md) separately
- Rewriting all 80+ Vue components — scope is **core farmer path** only

## Test plan (after ship)

1. Keyboard-only: login → Today → zone → Guardian → ask → cite link → back
2. VoiceOver (macOS) or NVDA (Windows) on Guardian turn with accuracy banner
3. `npm run test` includes phase-158 closure test
