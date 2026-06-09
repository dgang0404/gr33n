---
name: Phase 75 — Automation & comfort workspace
overview: >
  Collapse the climate/automation domain — Targets & schedules, Schedules (cron),
  Automations, and Setpoints (raw) — into one Comfort & automation workspace with
  progressive-disclosure tabs. Reuses the shipped Phase 42 ComfortTargetsHub farmer
  views and the Advanced power-user pages as tab bodies. Zone Climate tab stays the
  in-room editor; farm-wide admin lives in the workspace. Retire the entire Advanced
  sidebar group. UI-only — no schema, no API, no Pi.
todos:
  - id: ws1-workspace-shell
    content: "WS1: Add comfort workspace to workspaces.js — route /comfort-targets (keep path), tabs Comfort | Schedules | Automations | Raw; WorkspaceShell wrapper"
    status: completed
  - id: ws2-comfort-tab
    content: "WS2: Comfort tab — embed ComfortTargetsHub bands view (zone cards, ComfortBandEditor); default landing tab"
    status: completed
  - id: ws3-schedules-tab
    content: "WS3: Schedules tab — farmer 'what runs when' from hub + embed Schedules.vue cron editor behind 'Cron editor' expand or sub-tab"
    status: completed
  - id: ws4-automations-tab
    content: "WS4: Automations tab — farmer rules view from hub + embed Automation.vue for expression editing"
    status: completed
  - id: ws5-setpoints-tab
    content: "WS5: Raw tab — embed Setpoints.vue; farmer vocab ban on other tabs; 'setpoints' jargon allowed here only"
    status: completed
  - id: ws6-nav-redirects
    content: "WS6: Remove Advanced nav group; redirect /schedules,/automation,/setpoints → /comfort-targets?tab=; update ZoneAdvancedHint + navRelations + context_ref"
    status: completed
  - id: ws7-zone-climate
    content: "WS7: Zone Climate tab — remove link-outs to farm-wide Advanced; cross-workspace wiggle to comfort workspace; keep inline ZoneComfortTargets + ZoneAutomationPanel"
    status: completed
  - id: ws8-docs-tests
    content: "WS8: comfort-workspace.test.js, phase-75-closure.test.js; operator-tour §7h; OC-75"
    status: completed
isProject: false
---

# Phase 75 — Automation & comfort workspace

## Status

**Shipped.** Closes the largest remaining sidebar duplication after [Phase 74](phase_74_zone_ops_inbox.plan.md). UI-only — reuses [Phase 42](phase_42_comfort_targets_automation_plain_language.plan.md) farmer views and existing Advanced CRUD pages.

**Closure:** **OC-75** — tracked in this plan's DoD + [arc hub OC table](phase_68_73_spa_workspace_roadmap.plan.md#operational-closure-oc-rows). Do not add to the archived Phase 35 closure doc.

**Depends on:** [Phase 68](phase_68_workspace_shell_spa_nav.plan.md) workspace shell (shipped). Best after [Phase 69](phase_69_zone_workspace_hub.plan.md) zone inline automation toggles (shipped).

---

## The one job

> **One place for "how comfortable should it be" and "what runs when" — from plain comfort bands down to cron and raw setpoints — without four sidebar entries and an Advanced group.**

---

## Problem

The climate/automation domain is split across **four sidebar routes** plus a whole **Advanced** nav group, while the zone **Climate** tab already edits the same data inline:

| Route today | View | In sidebar | Zone already has |
|-------------|------|------------|------------------|
| `/comfort-targets` | [`ComfortTargetsHub.vue`](../ui/src/views/ComfortTargetsHub.vue) — bands + schedules + rules tabs | Grow & operate | `ZoneComfortTargets` + `ZoneAutomationPanel` on Climate |
| `/schedules` | [`Schedules.vue`](../ui/src/views/Schedules.vue) — raw cron CRUD | Advanced | Schedule pause/resume inline (Phase 69) |
| `/automation` | [`Automation.vue`](../ui/src/views/Automation.vue) — rule expression editor | Advanced | Rule active toggle inline (Phase 69) |
| `/setpoints` | [`Setpoints.vue`](../ui/src/views/Setpoints.vue) — farm-wide raw bands | Advanced | "Raw bands →" link from zone comfort |

[Phase 42](phase_42_comfort_targets_automation_plain_language.plan.md) shipped the **farmer** hub (`ComfortTargetsHub`) and explicitly parked power-user pages under **Advanced → Power settings** (WS5). That tiering was correct — but as **separate sidebar destinations** it reads as four jobs for one domain. Phase 75 keeps the tiering as **tab order inside one workspace** and **retires the Advanced group**.

[`ZoneAdvancedHint.vue`](../ui/src/components/ZoneAdvancedHint.vue) still links to `/automation`, `/comfort-targets`, and `/schedules` from every zone page — a smell that farm-wide automation is scattered.

---

## Design principles

1. **Progressive disclosure as tabs.** Comfort (farmer) → Schedules → Automations → Raw (power-user) — same pattern as Feed & Water (Phase 71) and Money (Phase 72).
2. **Reuse components.** Each tab hosts existing views/sections; no rewrite of setpoint/schedule/rule APIs.
3. **In-zone work stays in-zone.** Zone Climate tab keeps inline edit (Phase 69/40); workspace is for **farm-wide** admin and cross-zone audit — like Zones → Fleet for hardware.
4. **Vocabulary boundary.** Plain language on Comfort/Schedules/Automations tabs ([Phase 47](phase_47_feeding_water_plain_language.plan.md) grow-path rules); "setpoint", raw cron, JSON allowed on **Raw** tab only.
5. **Keep route `/comfort-targets`.** Operators and docs already know "Targets & schedules"; path becomes the workspace shell (no breaking rename).
6. **Contract-safe.** `/schedules`, `/automation`, `/setpoints` redirect; no deleted paths.

---

## Target workspace shape

| Tab id | Label | Body | Absorbs |
|--------|-------|------|---------|
| `comfort` | Comfort | ComfortTargetsHub bands section (zone cards + `ComfortBandEditor`) | default `/comfort-targets` |
| `schedules` | What runs when | Hub schedules farmer view + `Schedules.vue` (cron) as expand or nested sub-tab | `/schedules` |
| `automations` | Automations | Hub rules farmer view + `Automation.vue` full editor | `/automation` |
| `raw` | Raw setpoints | `Setpoints.vue` | `/setpoints` |

Deep links: `/comfort-targets?tab=schedules&zone_id=2` · `/comfort-targets?tab=raw`

---

## WS1 — Workspace shell

- Add `comfort` (or extend existing route) to [`workspaces.js`](../ui/src/lib/workspaces.js):

```js
comfort: {
  label: 'Comfort & automation',
  icon: '🎯',
  route: '/comfort-targets',
  subtitle: 'Comfort bands, what runs when, and automation toggles',
  tabs: [
    { id: 'comfort', label: 'Comfort' },
    { id: 'schedules', label: 'What runs when' },
    { id: 'automations', label: 'Automations' },
    { id: 'raw', label: 'Raw setpoints' },
  ],
  absorbs: {
    '/schedules': { tab: 'schedules' },
    '/automation': { tab: 'automations' },
    '/setpoints': { tab: 'raw' },
  },
},
```

- New [`ComfortWorkspace.vue`](../ui/src/views/workspaces/ComfortWorkspace.vue) — `WorkspaceShell` + tab bodies (pattern: `FeedWaterWorkspace.vue`).

---

## WS2 — Comfort tab

- Default tab when landing on `/comfort-targets`.
- Body = existing hub **bands** content from `ComfortTargetsHub` (zone list, status chips, `ComfortBandEditor`, `?zone_id=` banner from Phase 41).
- Refactor hub into tab-friendly sections or mount hub with `initialTab=comfort` only — avoid double headers.

---

## WS3 — Schedules tab

- Farmer schedules list (humanized next run, active toggle) — already in hub `schedules` tab.
- **Cron editor:** embed or link-expand `Schedules.vue` for power users ("Edit cron expressions →").
- Zone-scoped filter when `?zone_id=` present (Phase 41 pattern).

---

## WS4 — Automations tab

- Farmer rules summary + active toggle — hub `rules` tab content.
- **Expression editor:** embed `Automation.vue` below or behind "Advanced rule editor →".
- Greenhouse template entry preserved (Phase 36/42).

---

## WS5 — Raw setpoints tab

- Full `Setpoints.vue` farm-wide editor.
- Tab label "Raw setpoints" in UI; nav title explains power-user scope.
- Zone detail "Raw bands →" link targets `/comfort-targets?tab=raw&zone_id=:id`.

---

## WS6 — Nav, redirects, wiggle

**Sidebar after Phase 75:**

| Group | Items |
|-------|-------|
| Today | Today |
| Grow & operate | Zones, Feed & water, **Comfort & automation**, Hardware, Money |
| ~~Advanced~~ | *(removed)* |
| More | (unchanged until Phase 76–77) |

- Remove **Advanced** group from [`navGroups.js`](../ui/src/lib/navGroups.js).
- Rename sidebar label from "Targets & schedules" → **Comfort & automation** (or keep farmer label "Targets & schedules" with updated `navTitle`).
- Spread `buildLegacyRedirectRoutes()` absorbs for `/schedules`, `/automation`, `/setpoints`.
- Update [`navRelations.js`](../ui/src/lib/navRelations.js): legacy paths wiggle `/comfort-targets`.
- Update [`ZoneAdvancedHint.vue`](../ui/src/components/ZoneAdvancedHint.vue): single link "Farm-wide comfort & automation →" to workspace; remove three-way split.
- Audit Guardian `context_ref.go` route names for schedules/automation/setpoints.

---

## WS7 — Zone Climate tab alignment

- [`ZoneNeedSection.vue`](../ui/src/components/ZoneNeedSection.vue) Climate / [`ZoneAutomationPanel.vue`](../ui/src/components/ZoneAutomationPanel.vue): replace `/automation` and `/comfort-targets` link-outs with workspace deep links or remove when inline edit suffices.
- [`ZoneComfortTargets.vue`](../ui/src/components/ZoneComfortTargets.vue): "Raw bands →" → `/comfort-targets?tab=raw&zone_id=…`.
- Cross-workspace "Jump to" on comfort workspace: Zones, Feed & water (existing wiggle pattern).

---

## Out of scope

- Merging **lighting photoperiod schedules** into this workspace — lighting stays Zones → Fleet / zone Light tab (Phase 69).
- **Fertigation feed schedules** — Feed & Water workspace (Phase 71).
- Schema/API changes to setpoints, schedules, or rules.

---

## Definition of done

- [ ] `/comfort-targets` is a WorkspaceShell with four tabs; Comfort is default
- [ ] `/schedules`, `/automation`, `/setpoints` redirect into the correct tab
- [ ] Advanced nav group removed; no duplicate sidebar entries for this domain
- [ ] Zone Climate tab edits inline; farm-wide admin in workspace only
- [ ] ZoneAdvancedHint and navRelations updated; Vitest green; operator-tour §7h; OC-75 closed

---

## Suggested implementation order

1. WS1 shell + redirects (low risk, proves pattern)
2. WS2 Comfort tab (move hub bands)
3. WS3–WS5 embed Advanced views as tabs
4. WS6 nav + ZoneAdvancedHint
5. WS7 zone link cleanup
6. WS8 closure

---

## Related

| Doc | Use |
|-----|-----|
| [phase_42_comfort_targets_automation_plain_language.plan.md](phase_42_comfort_targets_automation_plain_language.plan.md) | Original farmer hub + Advanced escape |
| [phase_69_zone_workspace_hub.plan.md](phase_69_zone_workspace_hub.plan.md) | Inline zone schedule/rule toggles |
| [phase_76_today_dashboard_nav_alignment.plan.md](phase_76_today_dashboard_nav_alignment.plan.md) | Dashboard links to comfort workspace |
| [phase_68_73_spa_workspace_roadmap.plan.md](phase_68_73_spa_workspace_roadmap.plan.md) | Arc hub |

---

## Using this in a new chat

> Read `docs/plans/phase_75_automation_comfort_workspace.plan.md`. UI-only. Wrap ComfortTargetsHub + Schedules + Automation + Setpoints in a WorkspaceShell at `/comfort-targets` with four tabs. Redirect legacy Advanced routes. Remove Advanced nav group. Zone Climate tab stays inline.
