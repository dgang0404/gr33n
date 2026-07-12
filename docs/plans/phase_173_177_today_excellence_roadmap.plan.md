---
name: Today excellence roadmap (Phases 173–177)
overview: >
  Locked execution order after Phase 172. Scale Today for large farms, tighten
  visual hierarchy, demote Guardian below the farm hero, add an operational
  "farm pulse," and close with first-impression / demo showcase polish — so
  `/` reads as a grower cockpit, not an AI chat launcher.
isProject: false
---

# Today excellence roadmap — Phases 173–177

**Status:** shipped (arc complete) · **Follows:** [172](phase_172_field_guide_demo_docs.plan.md) · **Prerequisite:** CORS `PUT` for zone layout saves (`cmd/api/cors.go`)

## North star

> Someone opens **Today** and thinks: *"This is my farm — I can see what's
> growing, what needs water, what's off — and it's beautiful."* Guardian is a
> helpful sidebar, not the product.

Phases 164–171 built the spatial map. Phases 169–170 added attention + one-tap
Guardian — valuable, but Today still **reads Guardian-heavy** below the canvas
(four chip rows) and **reads utilitarian** above it (TopBar says "Dashboard,"
header is zone/sensor counts). This arc fixes that without touching Zone Detail.

## Locked order

| Phase | Name | One-line goal |
|-------|------|----------------|
| **173** | [Large-farm navigation](phase_173_today_large_farm_navigation.plan.md) | Filters, mobile paging, desktop list overflow |
| **174** | [Visual hierarchy](phase_174_today_visual_hierarchy.plan.md) | "Today" naming, farm health header, spacing + tile polish |
| **175** | [Farm-first actions](phase_175_today_farm_first_actions.plan.md) | Operational CTAs hero; Guardian demoted |
| **176** | [Farm pulse](phase_176_today_farm_pulse.plan.md) | "What's happening now" strip — no AI |
| **177** | [First impression](phase_177_today_first_impression.plan.md) | Demo showcase, coach marks, arc closure |

**Execute strictly in order.** 173 is structural; 174–175 reshape what users
see first; 176 adds depth; 177 is the screenshot-ready pass.

## Today page — target layout (after 177)

**Clutter check (revised after review):** the original draft stacked a *new*
Farm Pulse strip (176) on top of the *existing* Site Strip (166) — two rows
answering the same "what's happening now" question. Fixed: **176 enriches
Site Strip in place** (extra cells, same card, zero added rows). Attention
strip and filter bar stay conditional — they earn their space only when
there's something to flag or a farm big enough to need filtering. Default
small-farm view is 4 rows before the hero, not 6.

```
┌─────────────────────────────────────────────────────────┐
│ TopBar: "Today" · farm name · time · API status         │
├─────────────────────────────────────────────────────────┤
│ FarmTodayHeader — health rollup + tasks/alerts counts   │
├─────────────────────────────────────────────────────────┤
│ FarmSiteStrip — sun · outdoor · water · next run ·      │
│                 active crops · devices (176 enriches)   │
├─────────────────────────────────────────────────────────┤
│ Needs attention strip (only when zones are flagged)     │
├─────────────────────────────────────────────────────────┤
│ Zone filter chips (only when ≥9 zones, Phase 173)       │
├─────────────────────────────────────────────────────────┤
│ ★ FARM CANVAS / STACK — the hero ★                      │
├─────────────────────────────────────────────────────────┤
│ Farm action bar — Feed & water · New task · Schedules   │
├─────────────────────────────────────────────────────────┤
│ Ask gr33n (single row, collapsed when idle)             │
├─────────────────────────────────────────────────────────┤
│ ▸ All the details (power users)                         │
└─────────────────────────────────────────────────────────┘
```

## What stays out of this arc

- Zone Detail / My zones redesign
- New backend endpoints (except bugfixes)
- Guardian model / RAG changes
- Pan/zoom canvas (defer past 177 unless integrators ask)
- Multi-site / `meta_data.site` grouping (defer)

## Verification (full arc)

```bash
cd ui && npm test -- --run src/__tests__/phase-173-closure.test.js \
  src/__tests__/phase-174-closure.test.js \
  src/__tests__/phase-175-closure.test.js \
  src/__tests__/phase-176-closure.test.js \
  src/__tests__/phase-177-closure.test.js
npm run dev  # manual: demo farm walkthrough — farm-first, not Guardian-first
```

## Docs touchpoints (each phase)

- Own `phase_NNN_*.plan.md`
- Row in `phase-14-operator-documentation.md`
- Bullet in `current-state.md`
- §7k append in `operator-tour.md` when shipped
