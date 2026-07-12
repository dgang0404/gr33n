---
name: Phase 168 — Today cleanup: checklist removal + polish
overview: >
  Finishes the Today redesign: remove the IT-flavored GettingStartedChecklist
  from the dashboard (growers should see a farm, not sysadmin setup steps),
  replace its one useful job (guiding a brand-new empty farm) with the
  canvas's grower-native empty state, sweep remaining technical copy on
  Today through the farmer-vocabulary bans, and close out cross-phase tests
  and docs for the 164–167 arc.
todos:
  - id: ws1-remove-checklist
    content: "WS1: Remove GettingStartedChecklist from Dashboard.vue; retire or quarantine component + firstRunChecklist lib usage"
    status: pending
  - id: ws2-empty-farm-path
    content: "WS2: Grower-native first-run — canvas empty state + Guardian setup starters cover the new-farm journey"
    status: pending
  - id: ws3-copy-sweep
    content: "WS3: Farmer-vocabulary sweep of all new Today surfaces (canvas, tiles, sheet, site strip)"
    status: pending
  - id: ws4-docs-tests
    content: "WS4: Update operator-tour/current-state docs, prune dead tests, phase-168-closure"
    status: pending
isProject: false
---

# Phase 168 — Today cleanup: checklist removal + polish

**Status:** planned · **Depends on:** [166](phase_166_today_visual_farm_canvas.plan.md), [167](phase_167_mobile_stack_quick_actions.plan.md)

## WS1 — Remove the getting-started checklist

The checklist ("Connect edge device," "Set comfort targets," …) trains users
to think like IT staff on day one. Remove:

- `Dashboard.vue`: drop `GettingStartedChecklist` render + import,
  `firstRunItems` / `showFirstRunChecklist` / `firstRunStarters` /
  `firstRunDismissed` computed/state.
- `ui/src/components/GettingStartedChecklist.vue` +
  `ui/src/lib/firstRunChecklist.js`: delete if Dashboard was the only
  consumer (grep first — tests reference both). Update/remove
  `first-run-checklist.test.js`, `phase-44-closure.test.js`,
  `phase-53-ws4-crosslinks.test.js` expectations; keep git history as the
  record rather than quarantining dead code.

## WS2 — Grower-native first run

What the checklist actually covered must not silently vanish for a truly
empty farm:

- Canvas empty state (166 WS2) is the entry: "Add your first zone" → `/zones`.
- `buildSetupStarters` Guardian chips stay, resurfaced under the empty-state
  canvas only when the farm has 0 zones or 0 devices — Guardian walks the
  user through setup conversationally instead of a static todo list.
- Once ≥1 zone exists, Today never shows setup framing again — unwired
  sensors already read as calm "Not set up yet" on tiles (164/166 contract).

## WS3 — Copy sweep

- Run every new Today surface against `GROW_PATH_VOCABULARY_BANS` /
  `GROW_PATH_GENERIC_ROOM_BANS` (extend the existing vocabulary test to
  include the new component files).
- Audit for leftover dev jargon on Today: raw `zone_type` values, cron
  strings, "actuator"/"fertigation program" strings in default view (allowed
  inside the collapsed power-user section).
- Confirm the zone HelpTip copy from 166 WS3 shipped; if planning moved it,
  it lands here.

## WS4 — Docs + closure

- `docs/operator-tour.md` + `docs/current-state.md`: Today described as the
  visual farm cockpit; screenshots/click-path updates.
- Prune tests asserting the old dashboard layout (task/alert table
  always-visible expectations).
- `phase-168-closure.test.js`: Dashboard bundle contains FarmCanvas and not
  GettingStartedChecklist; vocabulary test covers new components.

## Acceptance criteria

1. No getting-started checklist anywhere; brand-new farm still has an obvious
   guided path (empty canvas CTA + Guardian setup chips).
2. `rg -i "getting.?started" ui/src` returns only historical test names or
   nothing.
3. Full UI suite green: `cd ui && npm test -- --run`.
4. Operator tour reflects the new Today.

## Verification

```bash
cd ui && npm test -- --run
rg -i "GettingStartedChecklist" ui/src
```
