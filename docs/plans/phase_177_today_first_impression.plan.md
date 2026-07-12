---
name: Phase 177 — Today first impression & arc closure
overview: >
  Close the Today excellence arc: demo farm tells a visual story on first open,
  lightweight coach marks (farm-first, not Guardian), performance/a11y pass,
  and screenshot-ready polish so `/` is the showcase screen.
todos:
  - id: ws1-demo-showcase
    content: "WS1: Demo seed + optional bundled layout background — screenshot-ready farm 1"
    status: pending
  - id: ws2-coach-marks
    content: "WS2: TodayCoachMarks — 3-step first visit (tap zone, attention, arrange); sessionStorage dismiss"
    status: pending
  - id: ws3-perf-a11y
    content: "WS3: Today load order, reduced motion, focus order, vocabulary final sweep"
    status: pending
  - id: ws4-docs-arc
    content: "WS4: operator-tour §7l, phase-14 table 173–177, current-state arc summary"
    status: pending
  - id: ws5-closure
    content: "WS5: phase-177-closure + today-excellence bundle test"
    status: pending
isProject: false
---

# Phase 177 — Today first impression & arc closure

**Status:** planned · **Follows:** [176](phase_176_today_farm_pulse.plan.md)

## Why

Phases 173–176 make Today **work** for real farms and **feel** farm-first.
Phase 177 makes it **sell** the product: the screen a grower screenshots, an
integrator demos on a projector, a new clone sees on `make dev-stack-fresh`.

Guardian stays available — but the first 10 seconds should be: *sun, pulse,
my zones, one thing needs attention.*

## WS1 — Demo showcase seed

**Goal:** Every demo zone tile tells a story — minimize "Not set up yet" on
the hero map.

| Zone | Target tile read |
|------|------------------|
| Veg Room | Blue Dream · healthy sensors |
| Flower Room | Gorilla Glue · bloom · humidity attention |
| Herb & Greens | Basil · gravity drip |
| Outdoor beds | Planted · calm unwired sensors OK |
| Propagation | Cuttings · scheduled light |

Tasks:

- Audit `master_seed.sql` readings/alerts/programs against `farmVisualStatus`
  — patch seed rows where tiles look empty
- **Optional:** ship a default `layout-background` for farm 1 (bundled WebP in
  `ui/public/demo/` uploaded via seed script OR documented operator step) —
  subtle greenhouse floor plan, not required for AC
- Re-verify Phase 171 layouts still look balanced with 174 canvas min-height

## WS2 — Today coach marks (`ui/src/components/TodayCoachMarks.vue`)

First visit to `/` with `zones.length > 0` (sessionStorage
`gr33n_today_coach_done`):

1. **"This is your farm"** — points at canvas/stack
2. **"Tap a zone"** — quick actions without leaving Today
3. **"Needs attention"** — attention strip when present, else pulse strip

- Non-modal tooltips (no fullscreen overlay); dismiss × or "Got it"
- **No Guardian step** — Ask gr33n discovered via sidebar
- Respect `prefers-reduced-motion`
- Skip on mobile if viewport too small (show step 2 only)

## WS3 — Performance & a11y closure

- `Dashboard.refreshAll()`: don't await `capabilities.fetch()` before painting
  canvas — skeleton OK for pulse, not for zones (zones already in store)
- Tab order: header → site → pulse → attention → filter → canvas tiles →
  action bar → details summary
- `aria-live="polite"` on attention strip when counts change after refresh
- Final `farmer-vocabulary` test coverage for all 173–177 components
- `phase-177-today-a11y.test.js` — focus order smoke (jsdom tab simulation)

## WS4 — Documentation arc

- `operator-tour.md` new **§7l Today excellence (173–177)**
- `phase-14-operator-documentation.md` — table rows for 174–177 when shipped
- `current-state.md` — replace single 173 bullet with full arc summary
- README one-liner: "Today is a visual farm cockpit" with link to §7l

## WS5 — Closure bundle

- `phase-177-closure.test.js`
- `today-excellence-arc.test.js` — imports chain: header, pulse, action bar,
  filter bar, coach marks; Dashboard does NOT import four bare
  `GuardianStarterChips` in hero flow
- Manual QA checklist in plan (screenshot viewports: 390px, 1280px, 1920px)

## Acceptance criteria

1. Fresh `make dev-stack-fresh`: Today demo is screenshot-ready within 2s of
   zone paint (no Guardian chip wall).
2. Coach marks show once, dismiss persists for session.
3. ≥5/7 demo zones show plants + water or light line on tiles.
4. Full `phase-173` through `phase-177` test bundles green.
5. Zone Detail still untouched.

## Verification

```bash
cd ui && npm test -- --run src/__tests__/phase-177-closure.test.js src/__tests__/today-excellence-arc.test.js
make dev-stack-fresh
# Manual: first visit coach marks, screenshot Today at 1280px
```

## Arc complete

After 177 ships, Today (`/`) is the canonical **grower cockpit** documented in
operator tour §7k–§7l. Further work (multi-site pages, canvas pan/zoom) opens
as Phase 178+ only if field integrators request it.
