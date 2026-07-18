---
name: Phase 174 — Today visual hierarchy
overview: >
  Tighten Today above the canvas: rename Dashboard → Today, replace dry stats
  with a farm health header, unify section spacing, and polish tile/canvas
  visuals so the spatial map is the unmistakable hero.
todos:
  - id: ws1-today-naming
    content: "WS1: TopBar + document title + guardianStarters pageName → Today"
    status: completed
  - id: ws2-farm-header
    content: "WS2: FarmTodayHeader.vue — health rollup, tasks/alerts inline, drop duplicate attention row"
    status: completed
  - id: ws3-section-rhythm
    content: "WS3: Section spacing, single Your farm heading, canvas min-height + stage polish"
    status: completed
  - id: ws4-tile-polish
    content: "WS4: FarmCanvasZoneTile visual pass — borders, hover, attention glow, typography"
    status: completed
  - id: ws5-closure
    content: "WS5: phase-174-closure + farm-today-header.test.js"
    status: completed
isProject: false
---

# Phase 174 — Today visual hierarchy

**Status:** shipped · **Follows:** [173](phase_173_today_large_farm_navigation.plan.md)

## Shipped

| WS | Deliverable |
|----|-------------|
| **WS1** | TopBar `/` → **Today**; `document.title` → `Today · {farm}`; `guardianStarters` + `guardianRouteRef` use Today |
| **WS2** | `FarmTodayHeader.vue` + `farmTodayHeader.js` — health rollup pills; removed `dashboard-attention-row` |
| **WS3** | `space-y-5` rhythm; mobile stack heading **Your farm**; canvas `min-height` 420/480px; background `opacity-45` |
| **WS4** | Tile hover lift, warn/alert glow, empty-zone dashed inner border, `title` tooltip on plants line |
| **WS5** | `farm-today-header.test.js`, `phase-174-closure.test.js`, vocabulary scan |

## Why

Today still says **"Dashboard"** in the TopBar (`TopBar.vue` labels map) and
opens with `7 zones · 12 sensors · 3 devices` — accurate but cold. Below the
canvas, **tasks due** and **alerts** repeat in a separate row that duplicates
what the attention strip already communicates. The canvas is strong; everything
around it needs to feel as intentional.

## WS1 — "Today" everywhere

- `TopBar.vue`: `'/'` label → **Today**
- `document.title` or route meta if present → `Today · {farmName}`
- `guardianStarters.js`: `pageName` for dashboard surface → **Today** (not
  Dashboard)
- Grep repo for user-visible "Dashboard" on `/` path; fix Login subtitle only
  if it appears post-login (keep "Farm Automation Dashboard" on login page OK)

## WS2 — FarmTodayHeader (`ui/src/components/FarmTodayHeader.vue`)

Replace the current farm header block in `Dashboard.vue`:

```
gr33n Demo Farm
3 healthy · 2 need attention · 8 tasks today · 5 alerts
```

Pure rollup from existing store data + `computeZoneVisualStatus`:

| Pill | Source |
|------|--------|
| Healthy zones | `health === 'ok'` |
| Need attention | `zoneNeedsAttention` (169) |
| Tasks today | existing `todayTasks.length` |
| Alerts | `unreadAlerts` |

- Farm name stays prominent (h2)
- Pills are links: attention → filter attention (173) or scroll to strip;
  tasks → tasks route; alerts → alerts route
- **Remove** the standalone `dashboard-attention-row` section (tasks + alerts
  cards) — absorbed into header pills
- Subtitle line: optional local time greeting (`Good afternoon`) when coords +
  sun data available — subtle, not chatty

## WS3 — Section rhythm

- One section title for the map: **Your farm** (canvas owns it; mobile stack
  drops duplicate "Your zones" h3 or matches same label)
- Consistent `space-y-5` between major Today sections (header → site → pulse
  placeholder slot for 176 → attention → filter → canvas)
- `FarmCanvas` stage: `min-h-[420px] md:min-h-[480px]` so tiles breathe on
  wide screens
- Background image: slightly higher opacity when set (`opacity-40` → `45`) —
  farm photo should feel present, not a watermark

## WS4 — Tile polish

`FarmCanvasZoneTile.vue`:

- Health border: subtle outer glow on `warn`/`alert` (not just border color)
- Hover (non-arrange): `hover:shadow-xl hover:border-zinc-600` lift
- Truncate overflow with `title` tooltip on long crop names
- Empty zone: dashed inner border + "Ready to plant" accent (calm, not error)
- Arrange mode: clearer focus ring on keyboard-focused tile

No new data — presentation only.

## WS5 — Closure

- `farm-today-header.test.js` — rollup counts, pill links
- `phase-174-closure.test.js` — Dashboard imports header; no duplicate
  attention-row; TopBar says Today
- Vocabulary test: new header copy passes bans

## Acceptance criteria

1. TopBar on `/` reads **Today**.
2. Demo farm header: `4 healthy · 2 need attention` (or similar from seed) +
   task/alert counts — no separate tasks/alerts card row above Guardian.
3. Canvas min-height prevents crushed tiles on laptop viewports.
4. Zone Detail unchanged.
5. Phase 174 test bundle green.

## Verification

```bash
cd ui && npm test -- --run src/__tests__/farm-today-header.test.js src/__tests__/phase-174-closure.test.js
```
