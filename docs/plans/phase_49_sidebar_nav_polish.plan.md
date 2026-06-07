---
name: Phase 49 — Sidebar & navigation polish
overview: >
  Tighten the left sidebar: rename the technical fertigation console to "Fertigation"
  (agro term, allowed in Advanced), disambiguate the three "Feeding" entries, and add a
  hover affordance so related routes (zones ↔ feed & water ↔ targets) signal their
  connection. UI-only, no schema, no API contract changes. Respects prefers-reduced-motion.
todos:
  - id: ws1-fertigation-rename
    content: "WS1: Rename Advanced 'Feeding (technical)' → 'Fertigation'; disambiguate the three feeding nav labels; keep vocab grep ban (fertigation allowed in Advanced only)"
    status: completed
  - id: ws2-route-relationship-map
    content: "WS2: navRelations map — declare related nav items (zones, /feeding, /comfort-targets; controls↔sensors; lighting↔fertigation) consumed by SideNav"
    status: completed
  - id: ws3-hover-wiggle
    content: "WS3: Hover affordance — on hover/focus of a nav item, gently wiggle/highlight related items; prefers-reduced-motion fallback to static highlight"
    status: completed
  - id: ws4-docs-tests
    content: "WS4: nav-groups.test.js updates, new nav-relations Vitest, phase-49-closure.test.js; operator-tour nav note; OC-49"
    status: completed
isProject: false
---

# Phase 49 — Sidebar & navigation polish

## Status

**Shipped.** WS1–WS4 complete on `main`. UI-only polish. No DB, no API, no Pi.

**Roadmap:** [farmer_ux_roadmap_40_plus.plan.md](farmer_ux_roadmap_40_plus.plan.md) (polish track, follows Phase 45 whole-app polish).

**Closure:** **OC-49** in [phase_35_37_operational_closure.plan.md](phase_35_37_operational_closure.plan.md).

---

## Problem

Operator feedback after the Phase 48 clean-DB walkthrough:

1. **"Feeding" is overloaded.** Three sidebar entries all say some form of "Feeding," so the label no longer tells you which page you'll land on:

| Group | Today's label | Route | View | What it really is |
|-------|---------------|-------|------|-------------------|
| Grow | `Feed & water` | `/feeding` | `FeedingHub.vue` | Farmer per-room watering (Phase 47) |
| Operations | `Feeding (details)` | `/operations/feeding` | `FeedingAdminHub.vue` | Farm-wide feeding admin |
| Advanced | `Feeding (technical)` | `/fertigation` | `Fertigation.vue` | Six-tab fertigation console |

The Advanced one **is the fertigation console**. "Fertigation" is a legitimate agronomy term and is fine to use in the **Advanced / power-user** section (it stays banned on farmer grow paths — see [Vocabulary note](#vocabulary-note)).

2. **No sense of connection between related routes.** The sidebar is a flat list of `RouterLink`s ([`ui/src/components/SideNav.vue`](../ui/src/components/SideNav.vue)). Zones, Feed & water, and Targets & schedules are deeply related ("this room → how it's fed → what runs when"), but nothing tells the user that. A light hover affordance — when you hover one, its relatives gently wiggle/highlight — improves discoverability without adding chrome.

---

## Design principles

1. **UI-only.** No schema, no API contracts, no route-path changes (labels only). Renames must not break deep links.
2. **Power-user vocabulary stays in Advanced.** "Fertigation" appears only in the Advanced group; farmer grow paths keep plain language.
3. **Motion is optional, never required.** Any wiggle honors `prefers-reduced-motion` and degrades to a static highlight. No layout shift, no focus theft.
4. **Declarative relationships.** Related-route links live in one small map, not hard-coded in the component, so they're testable and easy to extend.

---

## WS1 — Fertigation rename + feeding disambiguation

Edit [`ui/src/lib/navGroups.js`](../ui/src/lib/navGroups.js):

| Group | New label | Route (unchanged) |
|-------|-----------|-------------------|
| Grow | `Feed & water` *(unchanged — farmer entry)* | `/feeding` |
| Operations | `Feeding admin` *(was "Feeding (details)")* | `/operations/feeding` |
| Advanced | **`Fertigation`** *(was "Feeding (technical)")* | `/fertigation` |

- Update `navTitle` tooltips to match (Advanced: "Fertigation console — programs, reservoirs, EC targets, mixing log").
- No `to:` paths change — existing bookmarks and Guardian route refs keep working.
- Confirm no farmer-facing grow view picks up "Fertigation" from a shared constant.

### Vocabulary note

Phase 47 WS5 added a grep ban on technical terms (`fertigation`, `setpoint`, `cron`, …) on **farmer grow paths** ([`ui/src/__tests__/farmer-vocabulary-grow-path.test.js`](../ui/src/__tests__/farmer-vocabulary-grow-path.test.js)). The rename targets the **Advanced** group, which is explicitly out of the farmer grow-path scope. WS4 extends that test to assert "Fertigation" is allowed in Advanced nav and still absent from grow paths.

---

## WS2 — Route relationship map

New `ui/src/lib/navRelations.js` — declares which nav items are "siblings on the same job":

```js
// keyed by route `to`; values are related `to` targets
export const NAV_RELATIONS = {
  '/zones':            ['/feeding', '/comfort-targets'],
  '/feeding':          ['/zones', '/comfort-targets'],
  '/comfort-targets':  ['/zones', '/feeding'],
  '/actuators':        ['/sensors', '/fertigation'],
  '/sensors':          ['/actuators'],
  '/lighting':         ['/fertigation'],
  // extend as jobs grow
}
```

- Pure data + a `relatedTo(to)` helper. No Vue here.
- Unit-tested in isolation (symmetry where intended, no dangling routes).

---

## WS3 — Hover affordance (the "wiggle")

In [`ui/src/components/SideNav.vue`](../ui/src/components/SideNav.vue):

- Track a `hoveredRoute` ref set on `@mouseenter` / `@focus` of each `RouterLink`, cleared on leave/blur.
- A nav item gets a `is-related` class when its `to` is in `relatedTo(hoveredRoute)`.
- `is-related` applies a subtle **wiggle** keyframe (small rotate/translate, ~400ms, 1–2 cycles) **plus** a faint highlight ring so the cue survives reduced motion.
- `@media (prefers-reduced-motion: reduce)` → drop the keyframe, keep only the highlight ring.
- Works collapsed (icon-only) and expanded. No width/layout shift; transform-only animation.
- Keyboard parity: focusing an item triggers the same related-highlight (accessibility).

---

## WS4 — Docs, tests, closure (OC-49)

| Artifact | Content |
|----------|---------|
| [`ui/src/__tests__/nav-groups.test.js`](../ui/src/__tests__/nav-groups.test.js) | Assert Advanced label is "Fertigation"; Operations is "Feeding admin"; routes unchanged |
| `ui/src/__tests__/nav-relations.test.js` (new) | `relatedTo` returns expected siblings; no relation points at a non-existent route |
| [`farmer-vocabulary-grow-path.test.js`](../ui/src/__tests__/farmer-vocabulary-grow-path.test.js) | "Fertigation" allowed in Advanced nav; still banned on grow paths |
| `ui/src/__tests__/phase-49-closure.test.js` (new) | Closure bundle: rename + relations + reduced-motion class present |
| [operator-tour.md](../operator-tour.md) | One line: Advanced → Fertigation; hover shows related pages |

**OC-49** row added to closure plan and closed when WS1–WS4 ship.

---

## Out of scope

- Restructuring nav groups or moving routes between groups (label-only here).
- Mobile bottom-nav changes ([`mobileBottomNav`](../ui/src/lib/navGroups.js) untouched).
- Breadcrumbs or in-page related-links (sidebar affordance only).
- Any schema, API, or Pi/GPIO work — that's [Phase 50](phase_50_hardware_wiring_visibility.plan.md).

---

## Definition of done

- [x] Advanced nav reads **Fertigation**; the three feeding entries are unambiguous
- [x] No route `to:` paths changed; deep links + Guardian route refs still resolve
- [x] Hovering/focusing a related nav item wiggles + highlights its siblings
- [x] `prefers-reduced-motion` falls back to static highlight (no animation)
- [x] Vitest green; OC-49 closed

---

## Suggested implementation order

1. WS1 rename (smallest, immediate clarity) + update `nav-groups.test.js`
2. WS2 relations map + unit test
3. WS3 hover affordance in `SideNav.vue`
4. WS4 closure test + operator-tour line

---

## Related

| Doc | Use |
|-----|-----|
| [phase_47_feeding_water_plain_language.plan.md](phase_47_feeding_water_plain_language.plan.md) | Why "fertigation" is banned on grow paths but fine in Advanced |
| [phase_45_farmer_validation_whole_app_polish.plan.md](phase_45_farmer_validation_whole_app_polish.plan.md) | Prior whole-app polish; a11y pass precedent |
| [phase_50_hardware_wiring_visibility.plan.md](phase_50_hardware_wiring_visibility.plan.md) | Companion phase — Pi/GPIO wiring visibility |
| [ui/src/lib/navGroups.js](../ui/src/lib/navGroups.js) | Primary file edited |

---

## Using this in a new chat

> Read `docs/plans/phase_49_sidebar_nav_polish.plan.md`. Implement one workstream (WS1–WS4). UI-only; do not change route paths or schema. Honor prefers-reduced-motion. Keep "fertigation" out of farmer grow paths.
