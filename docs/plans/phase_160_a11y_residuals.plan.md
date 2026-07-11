---
name: Phase 160 — a11y residuals + advisory CI guard
overview: >
  Close Phase 158 deferred items: lighting modal dialog semantics, key form
  labels, mobile nav drawer focus trap, and a non-blocking Vitest a11y smoke.
todos:
  - id: ws1-lighting-modal
    content: "WS1: ZoneLightingEditor modal — role=dialog, aria-labelledby, useDialogFocusTrap, Escape"
    status: pending
  - id: ws2-lighting-form-labels
    content: "WS2: LightingProgramForm — for/id on preset name, photoperiod hours, actuator select"
    status: pending
  - id: ws3-mobile-drawer-trap
    content: "WS3: App.vue mobile hamburger drawer — focus trap + aria-label on close"
    status: pending
  - id: ws4-advisory-axe
    content: "WS4: Vitest advisory check — eslint-plugin-vuejs-accessibility or static closure expanded (no CI block)"
    status: pending
  - id: ws5-audit-update
    content: "WS5: Update a11y-audit doc — move D2/D3/D5 to closed; note D4 advisory"
    status: pending
  - id: ws6-closure
    content: "WS6: phase-160-closure.test.js"
    status: pending
isProject: false
---

# Phase 160 — a11y residuals + advisory CI guard

**Status:** planned · **Hub:** [159–160 backlog](phase_159_160_post_158_gaps_backlog.plan.md) · **Continues:** [158](phase_158_accessibility_pass.plan.md)

---

## Why this phase

Phase 158 fixed the **core farmer path** (Guardian chat, skip link, zone tabs). The audit doc deferred lighting modals, form label gaps, and mobile drawer trapping — surfaces operators hit when tuning **Light** schedules after Guardian proposes a change.

---

## Workstreams

### WS1 — Zone lighting modal

**Target:** `ui/src/components/ZoneLightingEditor.vue` (or equivalent modal shell)

| Fix | Detail |
|-----|--------|
| `role="dialog"` | On modal panel |
| `aria-modal="true"` | Trap semantics |
| `aria-labelledby` | Link to modal title |
| Focus trap | Reuse `useDialogFocusTrap` from Phase 158 |
| Escape | Close modal, restore focus to trigger |

### WS2 — Lighting program form labels

**Target:** `ui/src/components/LightingProgramForm.vue`

Associate labels for high-traffic controls only (not every optional field):

- Program / preset name
- On/off hour inputs
- Primary actuator `<select>`
- `aria-pressed` on preset picker buttons if toggle-style

Errors: tie validation text via `aria-describedby` where inline errors exist.

### WS3 — Mobile hamburger drawer

**Target:** `ui/src/App.vue` mobile drawer `<aside>`

- `ref` on drawer panel
- `useDialogFocusTrap(drawerOpen, …)` with `onEscape` → close
- Close button `aria-label="Close navigation menu"`
- Return focus to hamburger on close

### WS4 — Advisory CI guard

**Non-blocking** — matches Phase 158 WS5 intent.

**Option A (preferred):** expand `phase-160-closure.test.js` with static source assertions (same pattern as 158 — no new deps).

**Option B:** add `eslint-plugin-vuejs-accessibility` as `npm run lint:a11y` advisory script; document in `INSTALL.md`; do **not** add to blocking CI until P0 list empty.

### WS5 — Audit doc update

Refresh [`a11y-audit-2026-07-11.md`](../a11y-audit-2026-07-11.md) or add `a11y-audit-phase160.md` with closed D2/D3/D5 rows.

### WS6 — Closure

`ui/src/__tests__/phase-160-closure.test.js` — assert shipped attributes in modified files.

---

## Acceptance

- [ ] Lighting editor modal: keyboard trap + Escape + labelled dialog
- [ ] Lighting form: labelled name + photoperiod + actuator select
- [ ] Mobile drawer: trap + labelled close
- [ ] Advisory a11y check documented and runnable locally
- [ ] `phase-160-closure.test.js` passes

## Non-goals

- Full WCAG 2.2 AA / VPAT
- Sidebar roving `tabindex` (D1 — still deferred)
- Blocking axe in GitHub CI
- Rewriting every form in the SPA

## Test plan (after ship)

```bash
cd ui && npm test -- --run src/__tests__/phase-160-closure.test.js
# Keyboard: zone → Light → open program editor → Tab cycle → Escape
# Mobile width: hamburger → Tab trapped → Escape → focus on hamburger
```
