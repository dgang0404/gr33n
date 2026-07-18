---
name: Phase 173 — Today large-farm navigation
overview: >
  Scale the Today farm cockpit for farms with many zones: filter chips (attention,
  indoor/outdoor/greenhouse), paginated mobile stack, and a desktop overflow path
  when the spatial canvas gets crowded — without changing Zone Detail or layout APIs.
todos:
  - id: ws1-filter-lib
    content: "WS1: farmTodayZoneFilter.js — filter predicates, counts, session persistence helpers"
    status: completed
  - id: ws2-filter-bar
    content: "WS2: FarmTodayZoneFilterBar.vue — chip row (All, Needs attention, Indoor, Outdoor, Greenhouse)"
    status: completed
  - id: ws3-canvas-stack-wire
    content: "WS3: Wire filter into FarmCanvas + FarmZoneStack; hidden zones keep saved layouts"
    status: completed
  - id: ws4-mobile-paging
    content: "WS4: Mobile stack paging — show 8 zones per page + Prev/Next"
    status: completed
  - id: ws5-desktop-overflow
    content: "WS5: Desktop overflow — Map/List toggle when filtered count > 13"
    status: completed
  - id: ws6-seed-fixture
    content: "WS6: Large-farm test fixture (24 zones) for filter/paging tests"
    status: completed
  - id: ws7-closure
    content: "WS7: Dashboard wiring, phase-173-closure + unit tests, docs"
    status: completed
isProject: false
---

# Phase 173 — Today large-farm navigation

**Status:** shipped · **Follows:** [172](phase_172_field_guide_demo_docs.plan.md) · **Arc:** [173–177 roadmap](phase_173_177_today_excellence_roadmap.plan.md) · **Prerequisite bugfix:** CORS must allow `PUT` for `saveZoneLayout` (shipped in `cmd/api/cors.go` — restart API after deploy)

## Shipped

| WS | Deliverable |
|----|-------------|
| **WS1** | `farmTodayZoneFilter.js` — filter predicates, counts, paging math, session persistence |
| **WS2** | `FarmTodayZoneFilterBar.vue` — chip row, only visible at ≥9 zones |
| **WS3** | `Dashboard.vue` owns filter state; passes filtered zones + total count + label into `FarmCanvas`/`FarmZoneStack` |
| **WS4** | `FarmZoneStack.vue` — pages beyond 8 zones with Prev/Next footer |
| **WS5** | `FarmCanvas.vue` — Map/List toggle beyond 13 zones; list reuses `FarmCanvasZoneTile` in a scrollable column |
| **WS6** | `ui/src/__tests__/fixtures/largeFarmZones.js` — 24-zone synthetic fixture |
| **WS7** | `farm-today-zone-filter.test.js`, `phase-173-closure.test.js`, docs |

Both `FarmCanvas` and `FarmZoneStack` distinguish a **farm-empty** state
("Add your first zone…") from a **filter-empty** state ("No zones match
Outdoor right now") so a strict filter never looks like an empty farm.

## Why

Phases 166–171 shipped a spatial Today map tuned for the **7-zone demo farm**. Real
commercial sites can have **20–80+ zones** across indoor rooms, greenhouses, and
outdoor blocks. Without navigation aids:

- Desktop canvas tiles overlap and become unreadable
- Mobile stack scrolls forever with no way to focus a site or problem set
- Attention zones are still surfaced (169), but finding a specific bed in a crowd
  is slow

This phase adds **lightweight filter + paging** on Today only. Zone Detail, My
zones, and the Phase 165 layout API stay unchanged.

## Product intent

A grower with 40 zones opens Today and can:

1. Tap **Needs attention** and see only flagged zones (strip + canvas/stack agree)
2. Tap **Outdoor** and work the beds without indoor rooms in the way
3. On phone, page through zones **8 at a time** instead of one endless scroll
4. On desktop, switch to a **compact list** when the spatial map is too dense

Default remains **All zones** — demo farm behavior unchanged.

## WS1 — Filter library (`ui/src/lib/farmTodayZoneFilter.js`)

Pure functions (unit-tested):

```js
export const TODAY_ZONE_FILTERS = [
  { id: 'all', label: 'All zones' },
  { id: 'attention', label: 'Needs attention' },
  { id: 'indoor', label: 'Indoor' },
  { id: 'outdoor', label: 'Outdoor' },
  { id: 'greenhouse', label: 'Greenhouse' },
]

filterZonesForToday(zones, filterId, getStatus)
// → zones matching filter; 'attention' uses zoneNeedsAttention (169)

countZonesPerFilter(zones, getStatus)
// → { all: N, attention: N, indoor: N, ... } for chip badges

shouldShowTodayPaging(zoneCount, { breakpoint: 'mobile'|'desktop' })
// → boolean thresholds: mobile page at ≥9, desktop list hint at ≥13
```

Rules:

- **Indoor / outdoor / greenhouse** — match `zone.zone_type` substring (same
  contract as `farmVisualStatus.zoneTypeLabel` and `FarmSiteStrip` outdoor rollup)
- **Attention** — reuse `zoneNeedsAttention` from `zoneQuickActions.js`
- **Sort** — after filter, still apply `sortZonesForStack` (attention-first)
- **Persistence** — `readTodayZoneFilter()` / `writeTodayZoneFilter()` via
  `sessionStorage` key `gr33n_today_zone_filter` (per browser tab session; not
  cross-device)

## WS2 — Filter bar (`ui/src/components/FarmTodayZoneFilterBar.vue`)

- Horizontal scrollable chip row; placed on Dashboard **below** attention strip,
  **above** canvas/stack
- Only render when `zones.length >= 9` **or** any non-`all` filter has count > 0
  (keeps 7-zone demo uncluttered)
- Each chip shows optional count badge (`3` on Needs attention)
- `aria-pressed` on active chip; keyboard left/right moves focus
- Emits `update:filter` — Dashboard owns filter state and passes filtered zones
  into canvas/stack

## WS3 — Canvas + stack wiring

**FarmCanvas.vue**

- Accept optional `filterId` prop or pre-filtered `zones` from Dashboard (prefer
  pre-filtered list from parent — single source of truth)
- Tiles not in filtered set are **not rendered**; their `meta_data.layout` is
  untouched (arrange mode only shows filtered zones; document in HelpTip)
- Subtitle when filter active: "Showing 4 of 28 zones · Outdoor"

**FarmZoneStack.vue**

- Same filtered zone list from Dashboard
- WS4 adds paging inside the stack component

**FarmTodayAttentionStrip**

- Unchanged — always lists **all** attention zones farm-wide (not filter-scoped)
  so urgent items never hide behind a filter

## WS4 — Mobile paging

When filtered zone count **> 8** on `md:hidden` stack:

- Show zones `[page * 8, (page+1) * 8)`
- Footer: `← Previous` · `Page 2 of 4` · `Next →`
- Reset `page` to 0 when filter changes
- Preserve sort order within page

## WS5 — Desktop overflow (compact list)

When filtered zone count **> 12** on `md:block` canvas:

- Show a **View** toggle next to "Arrange layout": `Map` | `List` (default Map)
- **List view** — vertical scroll of `FarmCanvasZoneTile` rows (reuse tile component,
  no absolute positioning); click → same quick-action sheet
- **Map view** — existing spatial canvas with filtered tiles only
- Toggle state in `sessionStorage` `gr33n_today_desktop_view`
- Arrange mode disabled in list view (switch to Map to arrange)

No pan/zoom in v1 — list fallback is simpler and accessible.

## WS6 — Large-farm fixture

For tests and manual QA (no mandatory seed change):

- `ui/src/__tests__/fixtures/largeFarmZones.js` — 24 synthetic zones (mix of
  types, 3 with attention status mocks)
- Optional follow-up: `db/seeds/large_farm_zones.sql` pack for integrators (defer
  unless QA asks)

## WS7 — Closure

- `farm-today-zone-filter.test.js` — predicates, counts, paging math
- `phase-173-closure.test.js` — filter bar imported by Dashboard; canvas/stack
  accept filtered props; mobile paging markup
- Operator tour §7k bullet + `current-state.md` section
- `phase-14-operator-documentation.md` table row

## Acceptance criteria

1. Demo farm (7 zones): no filter bar unless user deep-links with `?filter=attention`
   and ≥1 flagged zone — bar can show with low threshold only when counts warrant
   it (≥9 zones **or** integrator enables via env — default ≥9 hides bar on demo).
2. Fixture farm (24 zones): filter bar visible; Outdoor shows only outdoor beds;
   Needs attention shows Flower Room + any warn zones.
3. Mobile: page 1 shows 8 zones; page 2 shows remainder; filter resets page.
4. Desktop: 24 zones → list toggle appears; list view scrolls; map view still
   arranges and persists layout for visible zones.
5. Zone Detail and `/zones` unchanged.
6. `cd ui && npm test -- --run` green for phase 173 bundle.

## Out of scope (defer)

- Multi-floor / site **grouping** pages (requires `zones.meta_data.site` schema)
- Canvas pan/zoom or minimap
- Server-side zone pagination API (all zones already load in farm store)
- Guardian starters filtered by Today chip (starters stay farm-wide)

## Verification

```bash
cd ui && npm test -- --run src/__tests__/farm-today-zone-filter.test.js src/__tests__/phase-173-closure.test.js
npm run dev  # manual: import large fixture or add zones; exercise chips + paging + list toggle
```

## Implementation order

WS1 → WS2 → WS3 → WS4 → WS5 → WS6 → WS7 (lib first, then UI wiring, then overflow paths, then docs).
