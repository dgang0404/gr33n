---
name: Phase 74 — Zone ops inbox (tasks, alerts, plants)
overview: >
  Finish the zone-as-hub story Phases 40 and 69 started. Absorb farm-wide Tasks,
  Alerts, and the sparse Plants page into the zone SPA: full zone-scoped ops
  (ack alerts, run the task board, manage grows) without sidebar link-outs.
  Today (/) remains the only farm-wide triage surface for cross-zone "what needs
  me first." UI-only — reuses existing task/alert/plant/cycle endpoints; no schema,
  no new API, no Pi.
todos:
  - id: ws1-zone-ops-tab
    content: "WS1: Zone detail Ops tab — full zone-filtered alerts inbox + tasks board (reuse Alerts/Tasks views embedded; Done/Snooze/Create task/Ack inline); deep-link ?tab=ops&ops=alerts|tasks"
    status: completed
  - id: ws2-zone-plants-tab
    content: "WS2: Zone detail Plants tab — strains and grows for this zone (active + history via crop_cycles); inline New strain + Start grow; absorb /plants farm-wide list into Zones workspace Strains sub-tab or redirect"
    status: completed
  - id: ws3-today-triage
    content: "WS3: Today dashboard — sole farm-wide ops triage; cross-zone alert/task widgets link into zone Ops tabs; drop duplicate sidebar entries"
    status: completed
  - id: ws4-nav-sidebar
    content: "WS4: Remove Tasks, Alerts, Plants from sidebar; Grow & operate shrinks to zones/feed-water/targets/hardware/money; mobile bottom nav unchanged except drop alert shortcut if redundant"
    status: completed
  - id: ws5-redirects-wiggle
    content: "WS5: Redirect /tasks, /alerts, /plants → zone or Today; preserve ?zone_id= query; update workspaces.js absorbs, navRelations, v-nav-hint, Guardian route refs"
    status: completed
  - id: ws6-docs-tests
    content: "WS6: zone-ops-tab.test.js, zone-plants-tab.test.js, phase-74-closure.test.js; operator-tour §7g; OC-74"
    status: completed
isProject: false
---

# Phase 74 — Zone ops inbox (tasks, alerts, plants)

## Status

**Shipped.** Closes the gap after [Phase 69](phase_69_zone_workspace_hub.plan.md) (zone hardware inline). UI-only — no DB, no API, no Pi.

**Closure:** **OC-74** — tracked in this plan's DoD + [arc hub OC table](phase_68_73_spa_workspace_roadmap.plan.md#operational-closure-oc-rows). Do not add to the archived Phase 35 closure doc.

---

## The one job

> **Everything that happens in a room — alerts, tasks, and grows — is handled inside that room. The sidebar stops listing three half-empty pages that duplicate what the zone already shows.**

Operator feedback (2026-06, post–Phase 69): the zone detail page "already has so much and is doing good" — but **Tasks**, **Alerts**, and **Plants** still sit in the sidebar as separate destinations, and the zone only shows **summaries** with link-outs (`All tasks →`, farm-wide Alerts).

---

## Problem

| Page today | Zone already has | Gap |
|------------|------------------|-----|
| `/alerts` — farm-wide inbox | `ZoneAlertsPanel` on Overview (zone-filtered, ack inline) | Full inbox + history still a separate nav item |
| `/tasks` — Kanban board | `ZoneTasksPanel` (due today, Done/Snooze) + strip counts | Full board + create/edit still separate |
| `/plants` — strain library grid | `ZoneCurrentGrowStrip` + Start grow wizard | Empty-feeling page; every grow starts **in a zone** anyway |

[Phase 40](phase_40_unified_farmer_ux_zone_cockpit.plan.md) intentionally shipped zone **panels** with farm-wide **link-outs** ("Ack in zone; history on Alerts"). [Phase 68](phase_68_workspace_shell_spa_nav.plan.md) kept **Today / Tasks / Alerts** as standalone sidebar items. That was the right incremental step — but after 69's inline hardware story, the remaining duplication is obvious in the UI.

**Plants nuance:** `plants` rows are farm-scoped strain definitions (`plant_id` on `crop_cycles`); there is no `plants.zone_id`. Grows always attach to a **zone** via `crop_cycles.zone_id`. The `/plants` page is a thin strain CRUD grid whose primary action is **Start a grow** → pick a zone. That belongs on the zone, not as a fourth top-level Grow item.

---

## Design principles

1. **In-zone ops stay in-zone.** Ack an alert, complete a task, start a grow — all on `/zones/:id` without visiting `/alerts`, `/tasks`, or `/plants`.
2. **Today = farm-wide triage only.** One morning surface for cross-zone "what needs me first" (existing `FarmMorningStrip`, dashboard widgets, Guardian walkthrough). Not a third copy of the task board.
3. **Reuse, don't rebuild.** Embed [`Alerts.vue`](../ui/src/views/Alerts.vue), [`Tasks.vue`](../ui/src/views/Tasks.vue), and plant/grow components with a `zoneId` lock — same endpoints, hosted in zone context (pattern: Phase 41 `?zone_id=` on farm-wide pages).
4. **Strain library has one farm-wide home.** Either a **Strains** sub-tab on the Zones workspace (alongside Rooms / Fleet) for CRUD across all strains, or redirect `/plants` → `/zones?tab=strains`. Zone **Plants** tab shows strains **used in this zone** (via active/historical cycles).
5. **Contract-safe.** No deleted routes — `/tasks`, `/alerts`, `/plants` redirect with query preserved; Guardian `context_ref.go` route hints updated.

---

## WS1 — Zone Ops tab (alerts + tasks)

Add an **Ops** tab on [`ZoneDetail.vue`](../ui/src/views/ZoneDetail.vue) (alongside Overview, Water, Light, Climate):

| Sub-view | Source | Zone-scoped behavior |
|----------|--------|----------------------|
| **Alerts** | Extract or embed alert list from [`Alerts.vue`](../ui/src/views/Alerts.vue) | Filter to zone sensors/name; full ack/mark-read/create-task actions |
| **Tasks** | Embed [`Tasks.vue`](../ui/src/views/Tasks.vue) with `zoneId` prop | Kanban or list locked to `tasks.zone_id`; `+ New task` pre-fills zone |

Deep links: `/zones/:id?tab=ops&ops=alerts` · `…&ops=tasks`

**Overview cleanup:** Keep the Today strip counts (open alerts, overdue tasks) as **summary chips** that deep-link into Ops sub-views — same pattern as Phase 69's need-tab jumps. Optionally collapse `ZoneAlertsPanel` / `ZoneTasksPanel` into short previews with "See all in Ops →" instead of duplicating full lists on Overview.

**Promote from Phase 40 panels:**
- [`ZoneAlertsPanel.vue`](../ui/src/components/ZoneAlertsPanel.vue) — ack/mark-read/create-task (already inline)
- [`ZoneTasksPanel.vue`](../ui/src/components/ZoneTasksPanel.vue) — due-today Done/Snooze (already inline)

---

## WS2 — Zone Plants tab (strains + grows)

Add a **Plants** tab on `ZoneDetail`:

- **Active grow** — extend [`ZoneCurrentGrowStrip.vue`](../ui/src/components/ZoneCurrentGrowStrip.vue) (summary, harvest, Guardian starters).
- **History** — prior `crop_cycles` in this zone (links to cycle summary / compare).
- **Strains in this zone** — plants that have at least one cycle in this zone; quick **Start grow** with [`StartGrowWizard.vue`](../ui/src/components/StartGrowWizard.vue).
- **Add strain** — inline create plant (modal from [`Plants.vue`](../ui/src/views/Plants.vue)) when starting a new variety.

Farm-wide strain CRUD (today's `/plants` grid):

| Option | Route | When to use |
|--------|-------|-------------|
| **A (preferred)** | `/zones?tab=strains` | Zones workspace third tab — all farm strains, same cards as `Plants.vue` |
| **B (minimal)** | redirect `/plants` → `/zones?tab=strains` | Same UI, no new workspace tab name |

Zone tab answers *"what's growing here?"*; Strains tab answers *"what varieties does the farm know about?"*

---

## WS3 — Today as sole farm-wide ops entry

[`Dashboard.vue`](../ui/src/views/Dashboard.vue) keeps:

- Morning strip, Guardian walkthrough, first-run checklist
- Cross-zone **Alerts** + **Tasks** preview widgets (existing grid row)
- Chips link to **`/zones/:id?tab=ops&…`** when the item is zone-scoped, not to `/alerts` or `/tasks`

Remove mental model of "three Today-group sidebar items." After WS4, sidebar **Today** group is only **Today** (`/`).

Unassigned / farm-wide tasks (no `zone_id`): triage from Today widget → optional **Unassigned** row on a future Fleet-style ops view, or filter on Today only for Phase 74 v1.

---

## WS4 — Sidebar collapse

Update [`navGroups.js`](../ui/src/lib/navGroups.js):

**Remove from Grow & operate:**
- ~~Plants~~

**Remove entire items from Today group:**
- ~~Tasks~~
- ~~Alerts~~

**Today group:** `{ label: 'Today', items: [{ to: '/', label: 'Today' }] }` only.

**Mobile bottom nav:** keep Today / Zones / Feed / Alerts / More — or swap Alerts chip for Zones-only triage (decide in WS4; document in operator-tour).

---

## WS5 — Redirects & cross-links

Add to [`workspaces.js`](../ui/src/lib/workspaces.js) `zones` workspace absorbs (or dedicated redirect table):

| Legacy path | Redirect target |
|-------------|-----------------|
| `/tasks` | `/` if no zone_id, else `/zones/:id?tab=ops&ops=tasks` |
| `/alerts` | `/` if no zone_id, else `/zones/:id?tab=ops&ops=alerts` |
| `/plants` | `/zones?tab=strains` (Zones workspace) |

Preserve:
- `/tasks?create=1&zone_id=2` → zone Ops tasks + open create
- `/alerts` with alert tied to zone → deep-link that zone's Ops alerts
- `/crop-profiles/:id`, `/crop-cycles/:id/summary` — unchanged

Update [`navRelations.js`](../ui/src/lib/navRelations.js): legacy `/tasks`, `/alerts`, `/plants` → wiggle `/zones` sidebar item.

Audit [`internal/farmguardian/context_ref.go`](../internal/farmguardian/context_ref.go) for `/tasks`, `/alerts`, `/plants` route names.

---

## WS6 — Docs, tests, closure (OC-74)

| Artifact | Content |
|----------|---------|
| `ui/src/__tests__/zone-ops-tab.test.js` | Ops tab renders embedded alerts/tasks; zone filter applied |
| `ui/src/__tests__/zone-plants-tab.test.js` | Plants tab lists cycles for zone; start-grow opens wizard |
| `ui/src/__tests__/phase-74-closure.test.js` | Sidebar omits tasks/alerts/plants; redirects resolve |
| [operator-tour.md](../operator-tour.md) | §7g — "Ops and grows live in the zone" |

**OC-74** closed when WS1–WS6 ship.

---

## Out of scope

- Replacing **Today** dashboard — it stays as farm-wide triage (Phase 41/60 morning walkthrough).
- **Catalog / crop profiles** — global reference data; not merged into zone (links from grow strip remain).
- **Animals / Aquaponics** — separate domains in More.
- Schema changes (e.g. `plants.zone_id`) — not required; zone association via `crop_cycles` is sufficient.

---

## Definition of done

- [ ] Zone Ops tab: full alerts + tasks for this zone inline; no link-out to `/alerts` or `/tasks` for zone work
- [ ] Zone Plants tab: active grow, history, start grow, add strain — without visiting `/plants`
- [ ] Farm-wide strain library lives under Zones workspace (Strains tab) or equivalent redirect target
- [ ] Sidebar: no Tasks, Alerts, or Plants entries; Today group is Today only
- [ ] `/tasks`, `/alerts`, `/plants` redirect; `?zone_id=` deep links land on correct zone tab
- [ ] Vitest green; operator-tour §7g; OC-74 closed

---

## Suggested implementation order

1. WS1 Ops tab (highest operator pain — alerts/tasks duplication)
2. WS2 Plants tab + Strains workspace sub-tab
3. WS5 redirects (safe to land with feature flags)
4. WS3 Today widget link targets
5. WS4 sidebar removal
6. WS6 closure

---

## Related

| Doc | Use |
|-----|-----|
| [phase_69_zone_workspace_hub.plan.md](phase_69_zone_workspace_hub.plan.md) | Zone inline hardware — prerequisite |
| [phase_40_unified_farmer_ux_zone_cockpit.plan.md](phase_40_unified_farmer_ux_zone_cockpit.plan.md) | Original zone alerts/tasks panels |
| [phase_41_farm_hub_coherence.plan.md](phase_41_farm_hub_coherence.plan.md) | `?zone_id=` on farm-wide pages |
| [phase_68_73_spa_workspace_roadmap.plan.md](phase_68_73_spa_workspace_roadmap.plan.md) | Arc hub — add Phase 74 row |

---

## Using this in a new chat

> Read `docs/plans/archive/phase_74_zone_ops_inbox.plan.md`. UI-only. Absorb Tasks, Alerts, and Plants into the zone SPA (Ops + Plants tabs on ZoneDetail; Strains under Zones workspace). Keep Today as farm-wide triage. Remove Tasks/Alerts/Plants from sidebar. Redirect legacy routes; never delete paths.
