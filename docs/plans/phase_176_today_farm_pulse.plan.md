---
name: Phase 176 — Today farm pulse
overview: >
  Enrich the existing FarmSiteStrip with operational "what's happening now"
  cells — next water/light runs, active crop stages, devices online — so the
  farm feels alive without adding a new row to Today (revised after clutter
  review: no standalone Farm Pulse component/row).
todos:
  - id: ws1-pulse-lib
    content: "WS1: farmTodayPulse.js — next runs, active cycles, device rollup"
    status: pending
  - id: ws2-pulse-component
    content: "WS2: Add pulse cells INTO FarmSiteStrip.vue (no new component/row)"
    status: pending
  - id: ws3-dashboard-wire
    content: "WS3: No Dashboard wiring change needed — FarmSiteStrip already gets the data it needs via existing props"
    status: pending
  - id: ws4-seed-align
    content: "WS4: Demo seed sanity — pulse shows meaningful lines on farm 1"
    status: pending
  - id: ws5-closure
    content: "WS5: farm-today-pulse.test.js + phase-176-closure"
    status: pending
isProject: false
---

# Phase 176 — Today farm pulse

**Status:** planned · **Follows:** [175](phase_175_today_farm_first_actions.plan.md)

## Why

The site strip (166) covers sun, outdoor rollup, and water **source** — good
context, but not **schedule**. Growers want: *"When is the next feed? What's in
bloom? Is my Pi online?"* That data already loads in `Dashboard.refreshAll()`;
it's just buried in "All the details."

**Revised scope:** an earlier draft of this phase added a second strip
(`FarmTodayPulse`) below Site Strip. Review caught that as redundant —
two cards asking "what's happening on my farm right now" back to back. This
version adds the same information as **more cells in the same card**, so the
farm feels operational and alive without growing Today's vertical footprint.

## WS1 — Pulse library (`ui/src/lib/farmTodayPulse.js`)

Pure functions from existing Dashboard refs (no new API):

```js
buildFarmTodayPulse({
  zones, programs, schedules, cropCycles, devices, actuators,
  siteWeather, queueDepth, fertigationEvents,
})
// → {
//   cells: [
//     { id: 'next_water', label: 'Next water', value: 'Veg Room · 3:00 AM', link },
//     { id: 'next_light', label: 'Lights', value: 'Flower Room on until 8 PM', link },
//     { id: 'crops', label: 'Growing', value: '5 runs · 2 in bloom', link },
//     { id: 'edge', label: 'Devices', value: '3 online · queue 0', link },
//   ],
//   outdoorTemp: '72°F' | null,  // from siteWeather when coords set
// }
```

Rules:

- **Next water** — earliest upcoming active program run across zones;
  `scheduleRunsLabel` + zone name; link → Feed & water
- **Lights** — next light schedule flip or "N zones on" summary; link →
  Comfort / schedules
- **Growing** — count active `crop_cycles` by stage bucket (veg / bloom /
  propagate); link → My zones or first active cycle
- **Devices** — `online/total` from `store.devices`; queue depth badge when
  >0; link → Hardware
- Omit cells with no data (don't show "Next water: —")
- Farmer language; no cron strings in values

## WS2 — Pulse cells inside FarmSiteStrip

- Extend `FarmSiteStrip.vue`'s existing flex row with 2–3 additional compact
  cells (`next_water`, `crops`, `devices`) matching the sun/outdoor/water cell
  style already there (`text-[10px]` label + `text-[11px]` value)
- On narrow viewports, cells wrap onto a second line within the **same card**
  (`flex-wrap` already present) rather than becoming a new section
- Each cell links out (Feed & water, My zones, Hardware) — same pattern as
  the existing water-source link
- Accept new optional props on `FarmSiteStrip` (`programs`, `cropCycles`,
  `devices`, `queueDepth`) — Dashboard already loads all of these, just pass
  them down alongside the props it already sends

## WS3 — Dashboard wiring

No new component import, no new row. `Dashboard.vue` passes a few extra
existing refs into the already-rendered `<FarmSiteStrip>`:

```
FarmTodayHeader
FarmSiteStrip        ← same row, now with pulse cells
FarmTodayAttentionStrip
FarmTodayZoneFilterBar (173, when shown)
FarmCanvas / FarmZoneStack
FarmTodayActionBar (175)
FarmTodayAskGr33n (175)
details
```

Pulse cells hide individually when their data is empty (existing per-cell
omit rule below) — no farm-empty special case needed since Site Strip
already renders for any farm state.

## WS4 — Demo seed alignment

Verify `master_seed.sql` farm 1 produces:

- At least one "Next water" line (Veg 3 AM, Herb gravity drip, etc.)
- "Growing" shows chrysanthemum runs
- Devices online count matches seeded Pi

Adjust seed copy only if pulse would be empty on fresh `make seed` — no schema
changes.

## WS5 — Closure

- `farm-today-pulse.test.js` — next run selection, stage counts, empty omit
- `phase-176-closure.test.js` — asserts `FarmSiteStrip` renders pulse cells;
  asserts **no new standalone pulse component** is imported by Dashboard
- `current-state.md` bullet

## Acceptance criteria

1. Demo farm Site Strip shows ≥2 populated pulse cells on first load,
   alongside the existing sun/outdoor/water cells — same card, no new row.
2. Links land on correct workspaces.
3. No Guardian or LLM dependency.
4. Phase 176 bundle green.

## Verification

```bash
cd ui && npm test -- --run src/__tests__/farm-today-pulse.test.js src/__tests__/phase-176-closure.test.js
make seed  # optional: confirm pulse lines on farm 1
```
