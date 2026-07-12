---
name: Phase 167 — Mobile zone stack + quick actions
overview: >
  Makes the visual Today work in the hand and adds the "do things" layer:
  on phones the canvas renders as stacked zone cards (same tiles, same
  status), and tapping any zone — canvas or stack — opens a quick-action
  sheet: water now, toggle light, complete/acknowledge, ask Guardian about
  this zone, open Zone Detail. Today + Guardian becomes enough to run a
  normal growing day without touching the workspaces.
todos:
  - id: ws1-responsive-stack
    content: "WS1: FarmZoneStack.vue — stacked FarmCanvasZoneTile list below md breakpoint; canvas hidden on phone"
    status: pending
  - id: ws2-quick-action-sheet
    content: "WS2: ZoneQuickActions.vue — bottom sheet: water now, light toggle, task/alert actions, links"
    status: pending
  - id: ws3-water-now
    content: "WS3: Water now — run-now on zone's program, or pulse pump/drip actuator when no program"
    status: pending
  - id: ws4-guardian-zone-link
    content: "WS4: 'Ask Guardian about this zone' — open drawer with zone-scoped starter prompt"
    status: pending
  - id: ws5-inline-triage
    content: "WS5: Complete task / acknowledge alert inline from the sheet (existing endpoints)"
    status: pending
  - id: ws6-closure
    content: "WS6: Touch-target a11y, tests, phase-167-closure"
    status: pending
isProject: false
---

# Phase 167 — Mobile zone stack + quick actions

**Status:** planned · **Depends on:** [166](phase_166_today_visual_farm_canvas.plan.md)

## WS1 — Responsive stack

- `FarmZoneStack.vue`: vertical list of the same `FarmCanvasZoneTile`
  components (full-width variant), ordered by attention (alert zones first,
  then tasks due, then healthy, empty last). Site strip stays on top.
- Dashboard renders canvas ≥ `md`, stack below; one source of tile data
  (WS1 rollup from 166) so status can never diverge between layouts.
- Arrange mode is desktop-only; the stack has no drag.

## WS2 — Quick-action sheet (`ZoneQuickActions.vue`)

Tapping a tile (canvas or stack) opens a bottom sheet (mobile) / popover
(desktop) replacing 166's interim direct link:

| Action | Backing |
|--------|---------|
| 💧 Water now | WS3 |
| 💡 Light on/off | existing actuator command path (`ActuatorCard` logic extracted to a composable, not duplicated) |
| ✅ Today's tasks for this zone | complete inline (WS5) or open zone Ops |
| 🔔 Alerts for this zone | acknowledge inline (WS5) |
| 🧙 Ask Guardian about this zone | WS4 |
| ⚙️ Open zone | `/zones/:id` (Zone Detail unchanged — the sheet must name the zone clearly so the user knows what they're editing) |

Rows only render when applicable (no light → no light row; greenhouse zones
add vent/shade toggles reusing the greenhouse command path).

## WS3 — Water now

- Zone has an active fertigation/irrigation program → `POST
  /fertigation/programs/{id}/run-now` (exists), confirm dialog shows program
  name + duration ("Run Herb Room Gravity Drip — 3 min?").
- No program but a pump/drip/valve actuator → pulse command with
  `duration_seconds` (existing command endpoint), default 60 s with a picker.
- Neither → row reads "Set up watering" → zone Water tab.
- Result feedback via the command queue depth the dashboard already polls.

## WS4 — Guardian zone link

- Extend `guardianStarters.js` with `buildZoneQuickStarters({zone, status})`
  producing 1–2 prompts grounded in the tile state ("Why is humidity high in
  Flower Room?", "What should I do in Herb & Greens today?").
- Sheet button opens the Guardian drawer prefilled — same mechanism the
  dashboard starter chips use today. No Guardian backend changes.

## WS5 — Inline triage

- Task row: checkbox completes via existing task update endpoint; optimistic
  update, store refresh on failure.
- Alert row: acknowledge via existing alert ack endpoint (same one Guardian's
  `ack_alert` confirm path hits).
- Both capped to ~3 items in the sheet with "View all in zone Ops" link.

## WS6 — Closure

- All touch targets ≥ 44 px (repo already follows this pattern); sheet is
  focus-trapped, ESC/backdrop dismissable, aria-labeled.
- Tests: stack ordering, sheet action gating per zone shape, water-now
  branching (program vs pulse vs none), inline complete/ack optimistic paths,
  Guardian starter payload.
- `phase-167-closure.test.js`.

## Acceptance criteria

1. Phone viewport: Today = site strip + ordered zone cards; no horizontal
   scroll; every card opens the sheet.
2. Herb & Greens sheet: "Water now" runs the gravity-drip program and the
   event appears in recent feeds after refresh.
3. Flower Room sheet: humidity alert acknowledgeable inline; "Ask Guardian"
   opens the drawer with a zone-scoped prompt.
4. Zone Detail untouched; sheet's "Open zone" lands on the right zone.
5. `cd ui && npm test -- --run` green.

## Verification

```bash
cd ui && npm test -- --run src/__tests__/zone-quick-actions.test.js src/__tests__/farm-zone-stack.test.js src/__tests__/phase-167-closure.test.js
npm run dev  # manual: narrow viewport walkthrough, water-now on demo farm
```
