---
name: Phase 169 — Today attention cockpit
overview: >
  Closes the "Today + Guardian for daily growing" loop: zones that need care
  surface above the farm map (not buried in collapsed tables), desktop canvas
  sorts attention-first like mobile, and Guardian starters reference the live
  attention rollup from the visual map.
todos:
  - id: ws1-attention-strip
    content: "WS1: FarmTodayAttentionStrip — compact chips for warn/alert zones; tap opens quick actions"
    status: completed
  - id: ws2-canvas-sort
    content: "WS2: Attention-first zone ordering on FarmCanvas (reuse sortZonesForStack)"
    status: completed
  - id: ws3-guardian-starters
    content: "WS3: buildTodayAttentionStarters — farm-wide prompts from flagged zones"
    status: completed
  - id: ws4-docs-tests
    content: "WS4: Dashboard wiring, operator-tour note, phase-169-closure + unit tests"
    status: completed
isProject: false
---

# Phase 169 — Today attention cockpit

**Status:** shipped · **Depends on:** [168](phase_168_today_cleanup_polish.plan.md)

## Why

Phases 166–168 shipped the visual farm map and quick actions, but attention
zones are still easy to miss on desktop: canvas tiles follow DB order, and
Guardian morning chips are generic until you open a zone. Mobile already sorts
attention-first (167); this phase brings parity and makes **flagged zones +
Guardian** the default morning path on Today.

## WS1 — Attention strip

- `FarmTodayAttentionStrip.vue` — renders when ≥1 zone has `health` warn/alert
  or `attention` items; each chip shows zone name + one-line reason; click
  emits `select-zone` (same as canvas/stack).
- Placed on Dashboard between site strip and canvas/stack.

## WS2 — Canvas sort parity

- `FarmCanvas.vue` uses `sortZonesForStack` + `zoneHasTasksDueToday` before
  layout math (same contract as `FarmZoneStack`).

## WS3 — Guardian attention starters

- `buildTodayAttentionStarters` in `guardianStarters.js` — 1–3 chips when
  attention zones exist; single-zone → "Why {zone}?"; multi → triage walk +
  top zone prompts with zone context refs.

## WS4 — Closure

- `zoneNeedsAttention` / `listAttentionZones` helpers in `zoneQuickActions.js`
- Tests: `farm-today-attention.test.js`, `phase-169-closure.test.js`
- Operator tour §7k bullet + `current-state.md` note

## Acceptance criteria

1. Demo farm: Flower Room appears in attention strip (humidity); chip opens
   quick-action sheet.
2. Desktop canvas lists Flower Room before healthy zones.
3. Guardian attention chips visible on Today when zones are flagged.
4. `cd ui && npm test -- --run` green for phase 169 bundle.

## Verification

```bash
cd ui && npm test -- --run src/__tests__/farm-today-attention.test.js src/__tests__/phase-169-closure.test.js
```
