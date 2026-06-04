---
name: Phase 45 — Farmer validation & whole-app polish
overview: >
  Deliberate sit-in with non-technical farmers; fix friction backlog; mobile distribution
  execution; terminology pass v2; optional module shells (animals/aquaponics). Closes the
  "rough around the edges" gap that feature phases cannot predict.
todos:
  - id: ws1-sit-in-protocol
    content: "WS1: Sit-in protocol doc — tasks, script, success metrics; recruit 2–3 non-technical testers"
    status: pending
  - id: ws2-friction-backlog
    content: "WS2: Triage sit-in findings into P0/P1 fixes (UI-only preferred)"
    status: pending
  - id: ws3-copy-pass-v2
    content: "WS3: Copy pass v2 — grep technical terms site-wide; extend 20.9b pattern"
    status: pending
  - id: ws4-mobile-b4
    content: "WS4: Execute mobile-distribution.md checklist (icons, signing template, TestFlight path)"
    status: pending
  - id: ws5-module-shells
    content: "WS5: Animals/aquaponics — farmer-empty shells with why-empty + link to docs (no full CRUD redesign)"
    status: pending
  - id: ws6-accessibility-pass
    content: "WS6: Light a11y — focus order on wizards, contrast on chips, button labels"
    status: pending
  - id: ws7-docs-tests
    content: "WS7: operator-tour §9 validation; README farmer-ready statement; OC-45 closure"
    status: pending
isProject: false
---

# Phase 45 — Farmer validation & whole-app polish

## Status

**Planned.** After [Phases 40–44](farmer_ux_roadmap_40_plus.plan.md) feature work.

**Roadmap:** [farmer_ux_roadmap_40_plus.plan.md](farmer_ux_roadmap_40_plus.plan.md)

---

## Problem

Feature phases can ship correct **architecture** and still fail a **non-technical sit-in**:

- Inconsistent button labels
- Hidden prerequisites (Pi offline → empty graphs)
- Mobile WebView quirks
- Edge modules (animals) look "broken" when empty

Phase 45 is **evidence-driven polish**, not new domain features.

---

## WS1 — Sit-in protocol

**Artifact:** `docs/workstreams/farmer-sit-in-protocol.md` (create in WS1).

| Session | Script |
|---------|--------|
| Morning | Open Dashboard → handle alert → open zone → adjust comfort |
| Feed | Run program or pulse; read water story |
| Stock | Find low input; log mix (if applicable) |
| Setup (fresh profile) | New farm wizard only |

Record: confusion points, verbatim quotes, time-on-task.

---

## WS2 — Friction backlog

Triage into:

| Priority | Example |
|----------|---------|
| P0 | Cannot complete daily loop without help |
| P1 | Completes with wrong page |
| P2 | Copy / layout annoyance |

Prefer **UI + API composition** fixes; schema only if sit-in proves gap.

---

## WS3 — Copy pass v2

Extend [phase_20_9b](phase_20_9b_terminology_and_copy_pass.plan.md):

- Ban list: `cron`, `predicate`, `executable_action`, `zone_setpoints` in farmer routes
- HelpTips audit on 40–44 surfaces

---

## WS4 — Mobile (backlog B4)

Execute [mobile-distribution.md](../mobile-distribution.md) release checklist — at least one internal/TestFlight or sideload build documented end-to-end.

---

## WS5 — Module shells

For animals / aquaponics / low-use modules:

- Empty state: what this area is for + link to workflow doc
- Not full Phase 20.8 redesign unless sit-in demands it

---

## WS6 — Accessibility (light)

Focus visible, aria labels on Run now / Confirm, chip contrast — no full WCAG audit v1.

---

## WS7 — Docs, tests, closure (OC-45)

- README: "Farmer-ready v1" criteria met
- operator-tour §9
- OC-45 in closure doc

---

## Out of scope (remain Tier D)

- Closed-loop EC dosing
- Vendor hardware
- Enterprise multi-site dashboard
- Guardian without Confirm

---

## Definition of done

- [ ] ≥2 sit-ins completed; P0 backlog empty
- [ ] Copy pass v2 merged
- [ ] Mobile checklist executed or explicitly deferred with reason
- [ ] README + operator-tour updated
- [ ] OC-45 closed
