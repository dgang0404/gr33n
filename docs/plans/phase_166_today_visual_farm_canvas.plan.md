---
name: Phase 166 — Today visual farm canvas
overview: >
  Rebuild the Today tab (Dashboard.vue) around a spatial farm canvas: zones as
  draggable tiles over an optional photo/sketch background, each tile showing
  plants, light, water, and sensor health in farmer language with three
  explicit hardware states (healthy / needs attention / not set up yet); a
  site layer for sun, weather, and irrigation source; greenhouse zones
  visually differentiated. Zone Detail and every other page stay untouched —
  the canvas deep-links into them. Desktop-first; Phase 167 delivers the
  mobile stacked-card rendering and quick actions.
todos:
  - id: ws1-zone-status-lib
    content: "WS1: farmVisualStatus.js — per-zone rollup (plants, light, water, sensor health, attention) + 3 hardware states"
    status: completed
  - id: ws2-canvas-component
    content: "WS2: FarmCanvas.vue — positioned zone tiles, drag-to-arrange (persist via Phase 165), background image slot"
    status: completed
  - id: ws3-zone-tile
    content: "WS3: FarmCanvasZoneTile.vue — crop/stage, light state, next/last water, health color, greenhouse variant"
    status: completed
  - id: ws4-site-layer
    content: "WS4: Site strip — sunrise/sunset/daylength arc, outdoor summary, irrigation source → zone hints"
    status: completed
  - id: ws5-dashboard-rewire
    content: "WS5: Dashboard.vue — canvas as hero; demote tables to collapsed 'More' section; keep actuator control + Guardian starters"
    status: completed
  - id: ws6-closure
    content: "WS6: Component/unit tests, a11y pass on drag + tiles, phase-166-closure"
    status: completed
isProject: false
---

# Phase 166 — Today visual farm canvas

**Status:** shipped · **Depends on:** [164](phase_164_demo_seed_living_farm.plan.md) (living seed data) · [165](phase_165_farm_layout_api.plan.md) (layout persistence)

## Shipped

| WS | Deliverable |
|----|-------------|
| **WS1** | `farmVisualStatus.js` — 3 hardware states, water/light/plants rollup |
| **WS2** | `FarmCanvas.vue` — drag/resize arrange mode, background photo, layout persist |
| **WS3** | `FarmCanvasZoneTile.vue` — zone tile with farmer copy + health border |
| **WS4** | `FarmSiteStrip.vue` — sun dial, outdoor rollup, water source line |
| **WS5** | `Dashboard.vue` rewire — canvas hero, compact attention row, details disclosure |
| **WS6** | `farm-visual-status.test.js`, `farm-canvas.test.js`, `phase-166-closure.test.js` |

## Product intent

Today stops being a widget stack and becomes a picture of the farm. A grower
opens `/` and sees their zones laid out the way *they* arranged them — tent
here, outdoor bed there, greenhouse over there — each with plants, light,
water, and health readable at a glance. "I open Today, I see my farm, I do my
day." Deep edits still happen in Zone Detail (`/zones/:id`), which this phase
does not modify.

## WS1 — Zone status rollup (`ui/src/lib/farmVisualStatus.js`)

Pure functions (testable without components) that turn store data into one
tile model per zone:

```js
computeZoneVisualStatus({ zone, sensors, readings, devices, tasks, alerts,
                          schedules, programs, cropCycles, fertigationEvents })
// → {
//   plants:   { state: 'growing'|'empty', cropName, stage, batchLabel },
//   light:    { state: 'on'|'off'|'scheduled'|'none', scheduleLabel },
//   water:    { kind: 'pump'|'gravity_drip'|'manual'|'none', nextRun, lastEvent },
//   sensors:  { state: 'healthy'|'attention'|'not_set_up'|'mixed', summary, worst },
//   attention:[ { kind: 'alert'|'task', label, severity, link } ],
//   health:   'ok'|'warn'|'alert'|'unconfigured'   // tile border/badge color
// }
```

Rules:

- **Three hardware states** (Phase 164 contract): sensor with recent reading
  in thresholds → healthy; reading outside thresholds or linked open alert →
  attention; sensor with no readings ever / stale beyond N× interval →
  **"Not set up yet"** (calm gray, never an error tone, never "NO DATA").
- **Water kind:** program on zone → pump vs gravity drip via actuator type
  (`drip`) / program naming; `irrigation_only` → "plain water." Next run from
  the linked schedule (`scheduleRunsLabel`), last event from fertigation_events.
- **Farmer language everywhere** — run copy through
  `farmerVocabulary.js` bans; e.g. "Needs water," "Humidity high," "Bloom
  stage, day 14," "Empty — ready to plant."
- Greenhouse zones add `greenhouse: { policy, ventState, shadeState }` from
  `meta_data.greenhouse_climate` + actuator states (read-only reuse of what
  `ZoneGreenhouseTab` reads; no new automation).

## WS2 — Canvas (`ui/src/components/FarmCanvas.vue`)

- Aspect-fixed canvas region; zones absolutely positioned from
  `meta_data.layout` (Phase 165). Zones without layout auto-flow into a
  default grid, then persist wherever the user drops them.
- **Arrange mode** toggle: off by default (taps = open/quick actions); on →
  drag (pointer events, no heavy dnd dependency) + resize handles; save
  debounced via `store.saveZoneLayout`. Keyboard: arrow keys nudge the
  focused tile in arrange mode (a11y).
- **Background slot:** renders the farm background image (Phase 165) dimmed
  under tiles; "Add a photo or sketch of your space" affordance in arrange
  mode when none set.
- Empty farm → friendly "Add your first zone" state linking to `/zones`.

## WS3 — Zone tile (`ui/src/components/FarmCanvasZoneTile.vue`)

Compact card rendered on the canvas (and reused as the stacked card in 167):

- Header: zone name + type icon (tent/room 🏠, outdoor bed 🌱, greenhouse 🪴 —
  final icon set in implementation) + health color edge.
- Rows (only what applies): plants (crop + stage, or "Empty"), light
  (on/off/schedule), water (kind + next run — "Gravity drip · 7 AM daily"),
  sensor summary ("3 sensors healthy" / "Humidity high" / "Not set up yet").
- Attention badges (alert/task count) in severity color.
- Greenhouse variant: extra climate line (inside temp, vent/shade state) and
  distinct silhouette/accent so it reads as a different structure.
- Click (non-arrange mode): Phase 167 wires the quick-action sheet; this
  phase links tile → `/zones/:id` (correct deep route so the user knows what
  they're editing) as the interim behavior.
- Tooltip on the zone-type icon (HelpTip): "A zone is your grow area — a bed,
  tent, room, greenhouse section, or outdoor plot. You assign plants, lights,
  sensors, and watering to it." (New copy; audit first — no such tooltip
  exists today outside the farm-header HelpTip.)

## WS4 — Site layer

Slim strip above the canvas (`FarmSiteStrip.vue`):

- **Sun:** sunrise ↑, sunset ↓, daylength from `fetchSiteWeather` — the API
  already returns all three; today's UI only shows daylength. Simple arc/dial
  showing where "now" sits between sunrise and sunset.
- **Outdoor:** rollup of outdoor-zone sensors (or "No outdoor sensors yet").
- **Water source:** when reservoirs/programs exist, a compact
  `tank → pump/gravity → N zones` line; links to `/feed-water`. No full
  irrigation schematic in v1.
- Lat/long unset → single gentle prompt chip reusing `FarmConfigCard`'s save
  action (card itself moves out of the hero flow).

## WS5 — Dashboard rewire (`ui/src/views/Dashboard.vue`)

New order:

1. Farm header (kept, plus existing HelpTip)
2. **Site strip** (WS4)
3. **Farm canvas** (WS2/WS3) — the hero
4. Attention row — compact: overdue/today tasks + unread alerts count, each
   linking to existing views (replaces the two full-width tables)
5. Guardian starter chips (kept — morning walkthrough + weather + ops)
6. Collapsed **"All the details"** disclosure containing the current
   schedules/feeds/all-sensors/all-actuators sections for power users —
   markup mostly moves, doesn't die; actuator ON/OFF control stays available.

Removals/moves: the giant always-open Live Sensors grid, duplicate zone card
grid, and full task/alert tables leave the default view. The
GettingStartedChecklist stays for now — **Phase 168 removes it** (keeping this
phase's diff reviewable). `refreshAll()` data loading is already sufficient;
no new endpoints.

## WS6 — Closure

- Unit tests: `farmVisualStatus` (all three hardware states, water kinds,
  greenhouse rollup, empty zone), tile rendering per state, canvas
  positioning math + drag persistence call, site strip sun math.
- A11y: tiles focusable with descriptive aria-labels; arrange mode
  keyboard-operable; drag has no keyboard trap.
- `phase-166-closure.test.js` bundle guard (components imported by Dashboard).

## Acceptance criteria

1. Demo farm (post-164): canvas shows 7 zones — Veg Room healthy/green,
   Flower Room amber with "Humidity high," beds calm "Not set up yet,"
   Herb & Greens showing "Gravity drip" water line.
2. Drag a zone, reload → position persists; set a background photo → renders
   under tiles.
3. Zone tile click lands on the right `/zones/:id`.
4. No regressions on Zone Detail, My zones, Comfort, Feed & water (untouched).
5. `cd ui && npm test -- --run` green.

## Verification

```bash
cd ui && npm test -- --run src/__tests__/farm-visual-status.test.js src/__tests__/farm-canvas.test.js src/__tests__/phase-166-closure.test.js
npm run dev  # manual: arrange mode, background upload, tile links
```
