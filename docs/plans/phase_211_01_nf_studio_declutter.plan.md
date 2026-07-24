---
name: Phase 211.01 — Natural farming studio declutter
overview: >
  Demote the switchover wizard to onboarding-only: rename tab, remove duplicate
  Make a batch CTAs, move Commons import to operational tabs, default seeded
  farms to Make a batch. No schema or API changes.
todos:
  - id: ws1-rename-tab
    content: "WS1: Rename Start here → Switchover guide; step 5 Apply → Seed farm (optional)"
    status: completed
  - id: ws2-collapse-ctas
    content: "WS2: Replace dual Make batch buttons with one link to Make a batch tab"
    status: completed
  - id: ws3-move-commons-import
    content: "WS3: Move CommonsRecipePackImport from wizard footer to Recipes & apply"
    status: completed
  - id: ws4-default-tab
    content: "WS4: Default tab = batch when farm has recipes; start when blank"
    status: completed
  - id: ws5-tests
    content: "WS5: Update phase-209/211 closure tests; add naturalFarmingStudio helper test"
    status: completed
isProject: false
---

# Phase 211.01 — Natural farming studio declutter

**Status:** Shipped · **Depends on:** [211](phase_211_natural_farming_switchover_commons.plan.md) · **Before:** [211.02 recipe formula history](phase_211_02_recipe_formula_history.plan.md) · **Before:** [212 dual-install federation](phase_212_dual_farm_federation_test.plan.md)

## The one job

> Returning operators land on **Make a batch** or **Recipes & apply**, not a five-step
> wizard that duplicates those tabs. The switchover guide stays for bottle→natural
> onboarding and one-time farm seed actions only.

## Problem

The **Start here** tab mixed education, navigation shortcuts, and DB bootstrap actions
under a label (**Apply**) that collided with **Recipes & apply → Apply**. Users with
seeded demo farms saw redundant **Make JMS / Make JLF** buttons while **Make a batch**
already existed in the tab bar.

## Solution (Option A)

| Area | Before | After |
|------|--------|--------|
| Tab label | Start here | **Switchover guide** |
| Step 5 rail label | Apply | **Seed farm (optional)** |
| Step 5 CTAs | Make JMS + Make JLF + pack + bootstrap | One **Ready to ferment? → Make a batch** link + pack + bootstrap |
| Commons import | Wizard footer | **Recipes & apply** panel |
| Default tab (no `?tab=`) | `start` | **`batch`** if farm has ≥1 application recipe; else `start` |

Wizard steps 1–4 unchanged (context → bottle program → natural match → first batch pick).

## Out of scope (211.02+)

- Recipe version history / run snapshots for reporting
- Removing the switchover guide tab entirely
- Changing pack import idempotency semantics

## Acceptance

- `/natural-farming` on demo farm opens **Make a batch** when `?tab` omitted
- Switchover guide step 5 has no duplicate ferment buttons
- Commons pack import visible on **Recipes & apply**
- `npm test` — phase-209*, phase-211*, natural-farming* green

## Files

- `ui/src/lib/workspaces.js` — tab label
- `ui/src/lib/naturalFarmingStudio.js` — default tab helper
- `ui/src/views/workspaces/NaturalFarmingWorkspace.vue` — default tab redirect
- `ui/src/components/naturalfarming/SwitchoverWizard.vue` — collapsed CTAs
- `ui/src/components/naturalfarming/RecipesApplyPanel.vue` — Commons import
