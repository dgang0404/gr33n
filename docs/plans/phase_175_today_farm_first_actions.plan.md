---
name: Phase 175 — Today farm-first actions
overview: >
  Rebalance Today below the canvas: operational CTAs (feed, tasks, schedules)
  become the primary action bar; four Guardian chip rows collapse to one subtle
  Ask gr33n affordance so the page doesn't read as an AI launcher.
todos:
  - id: ws1-action-bar
    content: "WS1: FarmTodayActionBar.vue — Feed & water, New task, Schedules, My zones"
    status: completed
  - id: ws2-guardian-demotion
    content: "WS2: Collapse Guardian starters — one row max; morning/weather/ops into details or Ask drawer"
    status: completed
  - id: ws3-quick-actions-guardian
    content: "WS3: Keep zone-scoped Guardian in ZoneQuickActions only; remove duplicate attention starters on page"
    status: completed
  - id: ws4-empty-farm
    content: "WS4: Empty farm — setup chips stay; populated farm never shows 4 Guardian rows"
    status: completed
  - id: ws5-closure
    content: "WS5: phase-175-closure, update phase-170 tests for new chip placement"
    status: completed
isProject: false
---

# Phase 175 — Today farm-first actions

**Status:** shipped · **Follows:** [174](phase_174_today_visual_hierarchy.plan.md)

## Shipped

| WS | Deliverable |
|----|-------------|
| **WS1** | `FarmTodayActionBar.vue` — Feed & water, New task, What runs when, My zones |
| **WS2** | `FarmTodayAskGr33n.vue` + `farmTodayAskGr33n.js` — ≤2 curated chips; full set in details |
| **WS3** | Removed hero `attentionStarters` row; zone Guardian stays in `ZoneQuickActions` |
| **WS4** | Empty-farm setup chips unchanged (168) |
| **WS5** | `farm-today-ask-gr33n.test.js`, `phase-175-closure.test.js`; updated 169/60 closure |

## Why

After the canvas, `Dashboard.vue` currently renders **up to four**
`GuardianStarterChips` blocks in sequence:

1. `attentionStarters` (169)
2. `morningWalkthroughStarters`
3. `weatherStarters`
4. `dashboardOpsStarters`

That's 8–12 green pills before "All the details." New visitors assume gr33n is
a chatbot with a farm wallpaper. Guardian belongs on Today — but **under** the
farm, not instead of it.

## WS1 — FarmTodayActionBar (`ui/src/components/FarmTodayActionBar.vue`)

Horizontal action bar directly **below** canvas/stack (always visible on
populated farms):

| Action | Route / behavior |
|--------|------------------|
| **Feed & water** | `feedWaterRoute` |
| **New task** | `newTaskRoute` |
| **What runs when** | `/comfort-targets?tab=schedules` or schedules hub |
| **My zones** | `/zones` |

- Styled as solid farm actions (not green chat pills) — zinc/green primary
  buttons, `min-h-[44px]` on mobile
- Optional badge on Feed & water when low-stock alerts exist (reuse
  `lowStockCount` from Dashboard)

## WS2 — Guardian demotion

**Populated farm** (`zones.length > 0`):

- **Remove** standalone rows for `morningWalkthroughStarters`, `weatherStarters`,
  `dashboardOpsStarters` from default Today view
- **Remove** `attentionStarters` row — attention triage lives in
  `FarmTodayAttentionStrip` + zone quick actions (170 one-tap still works from
  sheet)
- Add **one** compact block:

```html
<FarmTodayAskGr33n :starters="curatedStarters" />
```

`curatedStarters` = at most **2** chips:
- "Morning check" (walkthrough) when before noon local OR first visit of day
- "Ask about your farm" generic opener

Everything else moves to:
- **All the details** disclosure → subsection "Ask gr33n" with full starter
  set (morning, weather, ops, attention) for power users
- Sidebar **Ask gr33n** button (unchanged)

**Empty farm** (168 contract): `buildSetupStarters` chips **stay** visible —
conversational onboarding is correct there.

## WS3 — Zone quick actions own zone Guardian

`ZoneQuickActions.vue` already has Guardian prompts scoped to the zone. Ensure
`buildTodayAttentionStarters` messages are reachable from:

- Attention strip chip → quick actions → "Plan for {zone}" / Guardian row

No farm-wide duplicate attention chips on the page body.

## WS4 — Capabilities off

When `capabilities.aiEnabled === false`, hide `FarmTodayAskGr33n` entirely;
action bar still shows. Offline field mode unchanged.

## WS5 — Closure

- `phase-175-closure.test.js` — Dashboard has `FarmTodayActionBar`; at most
  one `FarmTodayAskGr33n` on populated farm; no four consecutive
  `GuardianStarterChips` in default template
- Update `phase-170-closure.test.js` expectations: one-tap counsel triggered
  from quick actions / details subsection, not hero attention row
- Operator tour §7k: "Guardian on Today is optional depth, not the hero"

## Acceptance criteria

1. Demo farm Today: canvas → action bar → ≤2 Ask chips → details. No wall of
   green pills.
2. Empty farm: setup Guardian chips still show (168).
3. Morning walkthrough still reachable (details or Ask chip).
4. Phase 170 one-tap counsel works from zone quick actions.
5. Phase 175 bundle green.

## Verification

```bash
cd ui && npm test -- --run src/__tests__/phase-175-closure.test.js src/__tests__/phase-170-closure.test.js
```
