---
name: Phase 54 — Zone connection nav & wiggle completion
overview: >
  Complete the "see how it all connects" affordance started in Phase 49/52.
  Interactive zone pipeline, remaining orphan links, and navRelations so ADHD-friendly
  hover chains work across grow, feed, comfort, stock, and money — without new APIs.
todos:
  - id: ws1-connection-pipeline
    content: "WS1: Interactive How it connects on zone tabs — each segment wiggles sidebar"
    status: pending
  - id: ws2-orphan-links
    content: "WS2: v-nav-hint on zone names, task zones, connection card Details, See history, Edit in Automations"
    status: pending
  - id: ws3-nav-relations
    content: "WS3: Expand navRelations — tasks↔alerts, fertigation↔feeding, grow↔money"
    status: pending
  - id: ws4-docs-tests
    content: "WS4: operator-tour § connection nav; phase-54-closure.test.js; OC-54"
    status: pending
isProject: false
---

# Phase 54 — Zone connection nav & wiggle completion

## Status

**Planned.** Depends on [Phase 53](phase_53_grow_stock_money_closure.plan.md) WS4 starting the wiggle pattern on new surfaces.

**Predecessor:** [Phase 52](phase_52_guardian_ui_context.plan.md) ✅ · [Phase 49](phase_49_sidebar_nav_polish.plan.md) ✅

---

## The one job

> **Hover anywhere in the grow story and the sidebar shows where that piece lives.**

---

## WS1 — Interactive connection pipeline

Replace static text in [ZoneNeedSection.vue](../../ui/src/components/ZoneNeedSection.vue):

```
sensor reading → target band → automation → pump/light/fan → device
```

Each segment is a hover target with `v-nav-hint`:

| Segment | Wiggles |
|---------|---------|
| sensor reading | `/sensors` |
| target band | `/comfort-targets` |
| automation | `/automation` |
| pump/light/fan | `/actuators` |
| device | `/pi-setup` (if offline) or `/actuators` |

Optional: subtle underline on hover; respect `prefers-reduced-motion` (wiggle only, no animation on text).

---

## WS2 — Orphan link pass

| Location | Hint target |
|----------|-------------|
| Actuator card zone name | `/zones/:id` |
| Task row zone link | `/zones/:id` |
| ZoneNeedConnectionCard Details → | `manageTo` path |
| ZoneWaterGrowStory See history → | `/operations/feeding` or feed history route |
| ZoneAutomationPanel Edit in Automations → | `/automation` |
| TargetsRulesPanel greenhouse templates → | `/zones` |
| Pi setup guide After wiring links | already in 53 — verify |

---

## WS3 — navRelations expansion

```javascript
'/tasks': ['/alerts', '/schedules', '/zones']
'/alerts': ['/tasks', '/zones']
'/fertigation': ['/feeding', '/operations/feeding']
'/operations/money': ['/operations/supplies']  // spend ↔ stock
'/plants': ['/zones', '/comfort-targets']
```

Sidebar self-hover continues to ripple related routes.

---

## WS4 — Docs, tests, OC-54

- operator-tour: "connection pipeline" paragraph
- Vitest: pipeline segments have v-nav-hint; navRelations keys
- Guardian: when `context_ref` is zone + water tab, mention pipeline in prompt block (Go — small `context_ref.go` addition)

---

## Definition of done

- [ ] Zone Water/Climate/Overview shows interactive pipeline
- [ ] No major in-app router-link without v-nav-hint (audit list in closure test)
- [ ] OC-54 closed
