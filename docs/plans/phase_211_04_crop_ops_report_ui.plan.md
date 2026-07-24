---
name: Phase 211.04 — Crop ops report UI
overview: >
  Thin operator UI for the crop-cycle ops timeline API shipped in 211.02 —
  feed, mix, light, and stage events with formula-at-time. Guardian read-only
  in 211.02; this phase is the page/chart surfacing only.
todos:
  - id: ws1-zone-or-money-surface
    content: "WS1: Pick home — Zone detail grow tab strip or Money → Grows row drill-down"
    status: pending
  - id: ws2-timeline-component
    content: "WS2: Reusable timeline component consuming GET crop-cycle ops API"
    status: pending
  - id: ws3-formula-at-time
    content: "WS3: Show revision snapshot / formula snapshot per mix and program run row"
    status: pending
  - id: ws4-closure
    content: "WS4: UI tests + operator-tour cross-link"
    status: pending
isProject: false
---

# Phase 211.04 — Crop ops report UI

**Status:** Planned · **Depends on:** [211.02 WS5 crop ops API](phase_211_02_recipe_formula_history.plan.md) · **After:** [211.03 farm permissions](phase_211_03_farm_permissions.plan.md) (timeline may show cost-sensitive rows — use `money.costs.read` if needed)

## The one job

> Answer “what was this room getting in February?” in the UI — not only via Guardian.

## Scope

- Read-only timeline UI wired to `GET …/crop-cycles/{id}/ops-timeline` (exact path from 211.02).
- No new schema; no write paths.

## Out of scope

- PDF export, cross-farm analytics (212+), Guardian write tools.
