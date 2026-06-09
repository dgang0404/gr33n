---
name: Phase 68 — Workspace shell / SPA nav consolidation
overview: >
  Collapse the deep 8-group / ~26-item left sidebar into a handful of full-page
  "workspaces" (SPAs), each a single route with internal tabs. Introduce a
  WorkspaceShell layout (header + tab strip + cross-workspace wiggle links),
  re-point navGroups to the workspaces, and redirect every retired route into its
  workspace tab so deep links, bookmarks, and Guardian route refs never break.
  UI-only — no schema, no API, no Pi. This is the shell that Phases 69–73 plug
  their content into.
todos:
  - id: ws1-workspace-model
    content: "WS1: Declare the workspace model — workspaces.js maps workspace id → route, tabs, and the legacy routes each tab absorbs; single source of truth consumed by nav + router"
    status: completed
  - id: ws2-workspace-shell
    content: "WS2: WorkspaceShell.vue layout — page header, tab strip (deep-linkable ?tab=), sticky sub-nav, related-workspace wiggle rail; mobile = tab dropdown"
    status: completed
  - id: ws3-nav-collapse
    content: "WS3: Rewrite navGroups.js to list workspaces instead of 26 leaf items; keep Today/Tasks/Alerts/Plants/Settings standalone; update SideNav + mobile drawer + bottom nav"
    status: completed
  - id: ws4-route-redirects
    content: "WS4: Add redirects from every retired route (/sensors, /actuators, /lighting, /fertigation, /operations/*, /costs, /inventory, ...) → workspace?tab=; keep named routes resolvable for context_ref.go"
    status: completed
  - id: ws5-cross-workspace-wiggle
    content: "WS5: Extend navRelations.js + v-nav-hint so in-workspace links wiggle the destination workspace (zone ↔ hardware ↔ feed-water ↔ money); honor prefers-reduced-motion"
    status: completed
  - id: ws6-docs-tests
    content: "WS6: nav-groups.test.js rewrite, workspace-routes redirect test, phase-68-closure.test.js; operator-tour 'workspaces' section; OC-68"
    status: completed
isProject: false
---

# Phase 68 — Workspace shell / SPA nav consolidation

## Status

**Shipped.** First phase of the [SPA workspace arc](phase_68_73_spa_workspace_roadmap.plan.md). UI-only. No DB, no API, no Pi. Everything else in the arc (69–73) plugs into the `WorkspaceShell` this phase introduces.

**Closure:** **OC-68** — tracked in this plan's DoD + [arc hub OC table](phase_68_73_spa_workspace_roadmap.plan.md#operational-closure-oc-rows). Do not add to the archived Phase 35 closure doc.

---

## Problem

The left sidebar is the navigation model, and it has grown to **8 groups and ~26 leaf items** in [`ui/src/lib/navGroups.js`](../ui/src/lib/navGroups.js), rendered by [`ui/src/components/SideNav.vue`](../ui/src/components/SideNav.vue) and the mobile drawer in [`ui/src/App.vue`](../ui/src/App.vue):

> Grow (My zones, Feed & water, Targets & schedules, Plants) · Today (Today, Tasks, Alerts) · Operations (Supplies, Feeding admin, Money) · Operate (Lighting) · Advanced (Schedules, Automations, Setpoints, Fertigation, Controls, Sensors) · Livestock · Monitor · System

Two problems the operator named directly:

1. **Too much to scroll; hard to tell what's where.** A flat 26-item list buries jobs.
2. **Pages explain the same thing.** The app deliberately tiers a domain across `hub → admin → raw editor` *and* duplicates it farm-wide vs per-zone — but to the user that just reads as "Supplies / Feeding admin / Fertigation are all the same," "Sensors / Controls / Plants / Lighting are all the same as the zone page."

This phase introduces the **workspace** concept: one full-page SPA per job, with internal tabs, so the sidebar shrinks to a short list of destinations and each destination holds a whole job. **It does not move any content yet** — it builds the shell and re-points navigation. Content consolidation happens in 69–72.

---

## Design principles

1. **UI-only, contract-safe.** No schema, no API, no Pi. **No route path is deleted** — retired paths `redirect` into a workspace tab so bookmarks, `RouterLink`s, and Guardian route refs ([`internal/farmguardian/context_ref.go`](../internal/farmguardian/context_ref.go)) keep resolving.
2. **Declarative.** Workspaces, their tabs, and the legacy routes each tab absorbs live in **one data file** — not hard-coded in components — so they're testable and the later phases just fill tabs in.
3. **Tabs are deep-linkable.** Every tab is `?tab=` addressable (the pattern `ZoneDetail` already uses), so Guardian and in-app links can land on an exact tab.
4. **Progressive disclosure preserved.** The farmer → advanced tiering survives as **tab order** inside a workspace (daily tab first, raw editor last), not as separate sidebar entries.
5. **Motion optional.** Cross-workspace wiggle honors `prefers-reduced-motion` (continues Phase 49/54).

---

## WS1 — Workspace model (data)

New [`ui/src/lib/workspaces.js`](../ui/src/lib/workspaces.js) — the single source of truth:

```js
// each workspace = one full-page SPA route with internal tabs.
// `absorbs` lists legacy routes that redirect into { route, tab }.
export const WORKSPACES = {
  zones: {
    label: 'Zones', icon: 'grid', route: '/zones',
    tabs: [
      { id: 'rooms',   label: 'Rooms' },          // zone list (was /zones)
      { id: 'fleet',   label: 'Fleet' },          // farm-wide sensors+controls+lighting (Phase 69)
    ],
    absorbs: {
      '/sensors':   { tab: 'fleet' },
      '/actuators': { tab: 'fleet' },
      '/lighting':  { tab: 'fleet' },
    },
  },
  hardware: {                                       // Phase 70 fills this
    label: 'Hardware', icon: 'chip', route: '/hardware',
    tabs: [
      { id: 'board',     label: 'GPIO board' },
      { id: 'devices',   label: 'Pi devices' },
      { id: 'reference', label: 'Wiring guide' },   // old PiSetupGuide constants
    ],
    absorbs: { '/pi-setup': { tab: 'reference' } },
  },
  feedwater: {                                      // Phase 71 fills this
    label: 'Feed & Water', icon: 'drop', route: '/feed-water',
    tabs: [
      { id: 'daily',     label: 'Daily' },          // FeedingHub
      { id: 'programs',  label: 'Programs & tanks' },// FeedingAdminHub + Fertigation reservoirs/EC
      { id: 'nutrients', label: 'Nutrients & mix' }, // supplies mixing
      { id: 'advanced',  label: 'Advanced' },        // Fertigation full console
    ],
    absorbs: {
      '/feeding':            { tab: 'daily' },
      '/operations/feeding': { tab: 'programs' },
      '/fertigation':        { tab: 'advanced' },
    },
  },
  money: {                                          // Phase 72 fills this
    label: 'Money', icon: 'coin', route: '/money',
    tabs: [
      { id: 'summary', label: 'This month' },        // MoneyHub
      { id: 'ledger',  label: 'Ledger' },            // Costs
      { id: 'supplies',label: 'Supplies & costs' },  // SuppliesHub + unit costs
    ],
    absorbs: {
      '/operations/money':    { tab: 'summary' },
      '/costs':               { tab: 'ledger' },
      '/operations/supplies': { tab: 'supplies' },
      '/inventory':           { tab: 'supplies' },
    },
  },
}
```

> **Note:** WS1 only *declares* the structure. The tab bodies for hardware/feedwater/money are stubbed in this phase (each renders the existing view component unchanged inside a tab) and properly merged in Phases 70/71/72. This keeps Phase 68 shippable on its own.

Helpers: `workspaceFor(legacyPath)` → `{ route, tab }`; `tabsFor(workspaceId)`. Unit-tested in isolation.

---

## WS2 — `WorkspaceShell.vue`

New [`ui/src/components/WorkspaceShell.vue`](../ui/src/components/WorkspaceShell.vue) — the layout every workspace route uses:

- **Header:** workspace title, short subtitle, primary action slot.
- **Tab strip:** from `tabsFor(id)`; selecting a tab sets `?tab=`; active tab restores on reload/deep-link. Desktop = horizontal tabs; mobile = compact dropdown (no horizontal scroll trap).
- **Body:** `<router-view>`-style slot or `<component :is>` per active tab.
- **Related rail:** small "Jump to" chips for sibling workspaces (drives the WS5 wiggle, e.g. Zones → Hardware → Feed & Water).
- **Sticky sub-nav** on scroll so the operator never loses the tabs in a long page (addresses "you've got to scroll").

Workspace route components (`ZonesWorkspace.vue`, `HardwareWorkspace.vue`, `FeedWaterWorkspace.vue`, `MoneyWorkspace.vue`) are thin: they pass an id to `WorkspaceShell` and map tabs to content. In Phase 68 the content is the **existing view**, mounted inside the shell.

---

## WS3 — Sidebar collapse

Rewrite [`ui/src/lib/navGroups.js`](../ui/src/lib/navGroups.js) so the sidebar lists **workspaces + the already-single-job items**, not 26 leaves:

| Sidebar (after) | Route | Notes |
|-----------------|-------|-------|
| Today | `/` | unchanged |
| Tasks | `/tasks` | unchanged |
| Alerts | `/alerts` | unchanged |
| **Zones** | `/zones` | workspace (Rooms / Fleet) |
| **Feed & Water** | `/feed-water` | workspace |
| Targets & schedules | `/comfort-targets` | unchanged (its own job) |
| Plants | `/plants` | strain library (unchanged) |
| **Hardware** | `/hardware` | workspace |
| **Money** | `/money` | workspace |
| Guardian | `/chat` | unchanged (Phase 73) |
| Guide | `/operator-guide` | unchanged |
| Settings | `/settings` | unchanged |
| Livestock / Aquaponics / Catalog / Knowledge / Analytics | — | keep, optionally under a "More" group |

- Collapses **8 groups → ~3** (Today, Grow & Operate, System) or a flat short list — pick during WS3, lock in the closure test.
- Update `SideNav.vue`, the `App.vue` mobile drawer, and `mobileBottomNav` (swap `/feeding`-style leaves for workspace routes; bottom nav stays 5 items).
- `navTitle` tooltips describe the workspace ("Zones — every room, its sensors, controls and lighting").

---

## WS4 — Route redirects (no broken links)

In [`ui/src/router/index.js`](../ui/src/router/index.js):

- Add the new workspace routes (`/zones` already exists; add `/hardware`, `/feed-water`, `/money`) pointing at the workspace components.
- For every **absorbed** legacy path, add a `redirect` that maps to `workspace?tab=` using `workspaceFor()`:

```js
{ path: '/sensors',            redirect: { path: '/zones',      query: { tab: 'fleet' } } },
{ path: '/actuators',         redirect: { path: '/zones',      query: { tab: 'fleet' } } },
{ path: '/fertigation',       redirect: { path: '/feed-water', query: { tab: 'advanced' } } },
{ path: '/operations/money',  redirect: { path: '/money',      query: { tab: 'summary' } } },
{ path: '/costs',             redirect: { path: '/money',      query: { tab: 'ledger' } } },
{ path: '/pi-setup',          redirect: { path: '/hardware',   query: { tab: 'reference' } } },
// ...one per absorbed route
```

- **Parameterized routes are not redirected away** in this phase: `/sensors/:id`, `/zones/:id`, `/crop-cycles/:id/...` stay (they're detail pages, handled in 69/70).
- **Guardian route refs:** audit [`internal/farmguardian/context_ref.go`](../internal/farmguardian/context_ref.go) and [`ui/src/lib/navRelations.js`](../ui/src/lib/navRelations.js) for hard-coded paths; redirects cover them, but update the canonical targets so Guardian deep-links land on the tab, not a redirect bounce.

---

## WS5 — Cross-workspace wiggle

Extend the Phase 49/54 affordance from "sibling nav items" to "sibling workspaces":

- Add workspace-level relations to [`ui/src/lib/navRelations.js`](../ui/src/lib/navRelations.js): `/zones ↔ /hardware ↔ /feed-water ↔ /money`.
- The `WorkspaceShell` "Jump to" chips carry `v-nav-hint` ([`ui/src/directives/navHint.js`](../ui/src/directives/navHint.js)) so hovering a chip wiggles the destination workspace in the sidebar (the "pimp" cross-linking the operator asked for).
- Keep `prefers-reduced-motion` → static highlight fallback.

---

## WS6 — Docs, tests, closure (OC-68)

| Artifact | Content |
|----------|---------|
| [`ui/src/__tests__/nav-groups.test.js`](../ui/src/__tests__/nav-groups.test.js) | Rewrite: sidebar lists workspaces; group count reduced; no orphan leaf routes |
| `ui/src/__tests__/workspaces.test.js` (new) | `workspaceFor()` maps each legacy path to the right `{route, tab}`; `tabsFor()` complete |
| `ui/src/__tests__/workspace-redirects.test.js` (new) | Every absorbed legacy route resolves (redirects, no 404); param routes untouched |
| `ui/src/__tests__/phase-68-closure.test.js` (new) | Shell renders tabs; `?tab=` deep-link restores; reduced-motion class present |
| [operator-tour.md](../operator-tour.md) | New "Workspaces" section — sidebar is now jobs, tabs inside each |

**OC-68** added and closed when WS1–WS6 ship.

---

## Out of scope (handed to later phases)

- **Merging the content** inside Feed & Water / Money / Hardware tabs — they wrap existing views unchanged here (Phases 70/71/72).
- **Per-zone inline editing** of sensors/controls/lighting and the Fleet tab content (Phase 69).
- **Live GPIO board** and Pi backend work (Phase 70).
- **Guardian PR discoverability** (Phase 73).
- Any schema / API / Pi change.

---

## Definition of done

- [ ] Sidebar shows workspaces + single-job items; group count materially reduced
- [ ] `WorkspaceShell` renders deep-linkable `?tab=` tabs (desktop tabs, mobile dropdown)
- [ ] Every absorbed legacy route redirects into the correct workspace tab — no 404, no broken bookmark
- [ ] Detail/param routes (`/zones/:id`, `/sensors/:id`, …) still resolve
- [ ] Cross-workspace hover wiggle works; reduced-motion falls back to highlight
- [ ] Vitest green; OC-68 closed

---

## Suggested implementation order

1. WS1 workspace model + unit test (data foundation)
2. WS4 redirects + redirect test (prove nothing breaks before moving UI)
3. WS2 `WorkspaceShell` (tabs wrap existing views)
4. WS3 sidebar collapse
5. WS5 cross-workspace wiggle
6. WS6 closure

---

## Related

| Doc | Use |
|-----|-----|
| [phase_68_73_spa_workspace_roadmap.plan.md](phase_68_73_spa_workspace_roadmap.plan.md) | Arc hub |
| [phase_49_sidebar_nav_polish.plan.md](phase_49_sidebar_nav_polish.plan.md) | Nav data model + wiggle this extends |
| [phase_54_zone_connection_nav.plan.md](phase_54_zone_connection_nav.plan.md) | Connection-nav precedent |
| [ui/src/lib/navGroups.js](../ui/src/lib/navGroups.js) | Primary file edited |
| [ui/src/router/index.js](../ui/src/router/index.js) | Redirects added here |

---

## Using this in a new chat

> Read `docs/plans/phase_68_workspace_shell_spa_nav.plan.md`. Implement one workstream (WS1–WS6). UI-only: never delete a route — redirect it into `workspace?tab=`. Tabs must be `?tab=` deep-linkable. Honor prefers-reduced-motion. Do not merge tab content yet (that's Phases 69–72).
