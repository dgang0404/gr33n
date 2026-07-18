---
name: Phase 171 — Demo farm zone layouts seed
overview: >
  Persists the Phase 166 default canvas positions into zones.meta_data.layout
  for farm-1 demo zones so make dev-stack-fresh shows a spatial farm map
  immediately — layouts survive reload and match the Phase 165 API contract.
todos:
  - id: ws1-seed-layouts
    content: "WS1: master_seed.sql — meta_data.layout for all 7 demo zones"
    status: completed
  - id: ws2-closure
    content: "WS2: phase-171-closure test + current-state note"
    status: completed
isProject: false
---

# Phase 171 — Demo farm zone layouts seed

**Status:** shipped · **Depends on:** [165](phase_165_farm_layout_api.plan.md), [166](phase_166_today_visual_farm_canvas.plan.md)

## Shipped

- Idempotent `UPDATE` on farm-1 zones merging `meta_data.layout` `{x,y,w,h}`
- Positions match `DEFAULT_ZONE_LAYOUTS_BY_NAME` in `farmVisualStatus.js`

## Verification

```bash
cd ui && npm test -- --run src/__tests__/phase-171-closure.test.js
```
