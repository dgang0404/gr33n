---
name: Phase 76 — Today dashboard & mobile nav alignment
overview: >
  After Phases 71–75 absorb feeding, money, ops, and comfort/automation into
  workspaces, sweep the Today dashboard (/) and mobile bottom nav so every link
  targets a workspace or zone deep link — never a retired sidebar route. UI-only.
todos:
  - id: ws1-dashboard-quick-actions
    content: "WS1: Dashboard quick actions — /feeding,/fertigation → /feed-water?tab=; /tasks?create → zone or Today; /operator-guide kept"
    status: completed
  - id: ws2-dashboard-widgets
    content: "WS2: Tasks/Alerts/Schedules/Feeding widgets — link to /zones/:id?tab=ops, /feed-water, /comfort-targets; empty states → workspaces not /automation"
    status: completed
  - id: ws3-morning-strip-chips
    content: "WS3: FarmMorningStrip + Do next chips — canonical paths via workspaces.js / canonicalSidebarPath"
    status: completed
  - id: ws4-mobile-bottom-nav
    content: "WS4: Mobile bottom nav — drop Alerts after Phase 74; use Ops via Zones or Today; align with final sidebar (~8 items)"
    status: completed
  - id: ws5-empty-state-hints
    content: "WS5: EmptyStateHint + guardianStarters dashboard paths — audit action-to for absorbed routes"
    status: completed
  - id: ws6-docs-tests
    content: "WS6: dashboard-workspace-links.test.js, phase-76-closure.test.js; operator-tour Today section refresh; OC-76"
    status: completed
isProject: false
---

# Phase 76 — Today dashboard & mobile nav alignment

## Status

**Shipped.** **Capstone polish** for the workspace arc — run **after** [Phases 71–75](phase_68_73_spa_workspace_roadmap.plan.md). UI-only.

**Closure:** **OC-76** — arc hub OC table. Do not add to Phase 35 closure doc.

**Depends on:** Redirects from 71 (feed-water), 72 (money), 74 (tasks/alerts/plants), 75 (schedules/automation/setpoints) should exist before this phase is marked shipped — otherwise the sweep re-links to pages that still look like separate destinations.

---

## The one job

> **Today (/) is the only farm-wide triage screen — and every chip, widget, and quick action on it lands in a workspace or a zone tab, not a ghost route.**

[Phase 74 WS3](phase_74_zone_ops_inbox.plan.md) states this intent but does not enumerate every [`Dashboard.vue`](../ui/src/views/Dashboard.vue) link. After 71–75, `/` is still full of legacy paths.

---

## Problem

| Dashboard area today | Links to | Should link to (after arc) |
|---------------------|----------|----------------------------|
| Quick actions | `/tasks?create=1`, `/feeding`, `/fertigation` | `/feed-water?tab=daily`, zone Ops or create-task flow |
| Tasks widget | `/tasks`, empty → `/tasks` | Zone Ops or Today-only preview |
| Alerts widget | `/alerts` | Zone Ops or Today-only preview |
| Schedules widget | `/schedules` | `/comfort-targets?tab=schedules` |
| Alerts empty state | `/automation` | `/comfort-targets?tab=automations` |
| Feeding widget | `/feeding` | `/feed-water?tab=daily` |
| Fertigation events | `/feeding` | `/feed-water` |
| Zone chips | `/zones/:id?tab=water` | ✓ already correct |

**Mobile bottom nav** ([`navGroups.js`](../ui/src/lib/navGroups.js) `mobileBottomNav`): still includes **Alerts** after Phase 74 removes it from sidebar — duplicates zone Ops.

---

## WS1 — Quick actions row

Update [`Dashboard.vue`](../ui/src/views/Dashboard.vue) quick actions:

| Action | New target |
|--------|------------|
| + New Task | `/tasks?create=1` redirect → modal on Today or first zone; prefer inline create sheet on Today |
| Feed & water | `/feed-water?tab=daily` |
| Log mix (advanced) | `/feed-water?tab=nutrients` |
| Operator guide | `/operator-guide` (unchanged) |

Remove any remaining `/fertigation` / `/operations/*` literals.

---

## WS2 — Widget link sweep

For each dashboard section, `v-nav-hint` and `router-link` `to` must resolve through [`canonicalSidebarPath()`](../ui/src/lib/workspaces.js):

- **Tasks:** row zone link → `/zones/:id?tab=ops&ops=tasks`; "View all" → `/` expanded section or first zone with overdue tasks (document choice in operator-tour).
- **Alerts:** same with `ops=alerts`.
- **Schedules:** `/comfort-targets?tab=schedules`.
- **Feeding / fertigation events:** `/feed-water` tabs.

Add Vitest: read `Dashboard.vue` and assert no bare `/feeding`, `/fertigation`, `/schedules`, `/automation`, `/tasks`, `/alerts` in `to=` or `action-to` (redirect routes OK in router tests only).

---

## WS3 — Morning strip & Guardian starters

- [`FarmMorningStrip`](../ui/src/components/FarmMorningStrip.vue) chips — audit `to` paths.
- [`guardianStarters.js`](../ui/src/lib/guardianStarters.js) `buildDashboardOpsStarters`, morning walkthrough — route refs to workspace paths.
- Phase 60/61 nudges that deep-link `/alerts` → zone Ops or `/comfort-targets`.

---

## WS4 — Mobile bottom nav

**Recommended end state** (after 74):

| Slot | Route | Label |
|------|-------|-------|
| 1 | `/` | Today |
| 2 | `/zones` | Zones |
| 3 | `/feed-water` | Feed |
| 4 | `/comfort-targets` | Targets |
| 5 | `/settings` | More |

Remove **Alerts** slot — triage happens on Today + zone Ops.

Alternative: slot 4 = `/money` for operators who log receipts mobile-first — document in operator-tour if chosen.

---

## WS5 — EmptyStateHint & cross-app audit

Grep UI for absorbed paths still used as primary navigation:

```
/feeding /fertigation /operations/ /tasks /alerts /plants
/schedules /automation /setpoints (post-75)
```

Fix or accept as redirect-only in [`emptyStateHints.js`](../ui/src/lib/emptyStateHints.js), Guardian cards, [`OperatorGuide.vue`](../ui/src/views/OperatorGuide.vue) click paths (Phase 77 may consolidate Guide).

---

## Definition of done

- [x] Dashboard has zero primary links to absorbed legacy routes (Vitest guard)
- [x] Mobile bottom nav matches post-74/75 sidebar story
- [x] Morning strip + dashboard starters use workspace paths
- [x] operator-tour §7i Today refreshed; OC-76 closed

---

## Related

| Doc | Use |
|-----|-----|
| [phase_74_zone_ops_inbox.plan.md](phase_74_zone_ops_inbox.plan.md) | Today = triage intent |
| [phase_71_feed_water_unification.plan.md](phase_71_feed_water_unification.plan.md) | Feed-water redirects |
| [phase_75_automation_comfort_workspace.plan.md](phase_75_automation_comfort_workspace.plan.md) | Comfort workspace redirects |

---

## Using this in a new chat

> Read `docs/plans/phase_76_today_dashboard_nav_alignment.plan.md`. After 71–75 redirects ship, sweep Dashboard.vue + mobileBottomNav + morning chips so no widget links to retired routes. UI-only.
