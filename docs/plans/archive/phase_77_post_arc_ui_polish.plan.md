---
name: Phase 77 — Post-arc UI polish (analytics, Guardian nav, help, farm config)
overview: >
  Final workspace-arc cleanup: give Analytics/grow-compare a real home, decide
  Guardian nav surfaces (drawer vs full page), optionally merge reference/help
  pages, and surface farm config (site coords, modules) without Settings sprawl.
  Mostly UI-only; may touch Settings sections and Guardian launcher only.
todos:
  - id: ws1-analytics-home
    content: "WS1: Analytics — CropCycleCompare + summary links under Zones Strains tab and/or Money Grows sub-tab; remove orphan More → Analytics or redirect"
    status: completed
  - id: ws2-guardian-nav-decision
    content: "WS2: Guardian nav — document drawer=primary, /chat=sessions+pending; remove or demote More → Guardian if global badge (Phase 73) suffices"
    status: completed
  - id: ws3-help-hub
    content: "WS3 (optional): Help & reference workspace — Guide + Knowledge + Catalog as tabs at /operator-guide or /help; redirect legacy paths"
    status: completed
  - id: ws4-farm-config-surface
    content: "WS4: Farm config — site lat/long (Phase 73), modules, timezone: slim 'Farm' card on Today or Settings tab; not a new sidebar item"
    status: completed
  - id: ws5-settings-trim
    content: "WS5: Settings trim — device wizard entry defers to Hardware (Phase 70); link out from Settings; reduce duplicate Pi copy"
    status: completed
  - id: ws6-docs-tests
    content: "WS6: phase-77-closure.test.js; operator-tour final sidebar map; arc hub 'target ~8 items' verified; OC-77"
    status: completed
isProject: false
---

# Phase 77 — Post-arc UI polish

## Status

**Shipped.** Last planned phase of the [SPA workspace arc](phase_68_73_spa_workspace_roadmap.plan.md). UI-only.

**Closure:** **OC-77** — arc hub OC table.

---

## The one job

> **When the arc is done, nothing important lives in "More" by accident — analytics, Guardian, help, and farm setup each have an obvious home, and the sidebar fits on one screen without scrolling.**

---

## Target sidebar (~8 items)

After Phases 68–76:

| Group | Items | Count |
|-------|-------|-------|
| **Today** | Today | 1 |
| **Grow & operate** | Zones, Feed & water, Comfort & automation, Hardware, Money | 5 |
| **More** | Animals*, Aquaponics*, Settings | 1–3 |

\*Hidden when farm module inactive (existing pattern).

**Phase 77** removes or relocates from More: ~~Analytics~~, ~~Guardian~~, ~~Guide~~, ~~Knowledge~~, ~~Catalog~~ (WS2–WS3).

---

## WS1 — Analytics / grow compare home

**Problem:** [`CropCycleCompare.vue`](../ui/src/views/CropCycleCompare.vue) and cycle summaries are grow economics — buried under **More → Analytics** with no workspace.

**Decision (document in plan, implement in WS1):**

| Surface | Home | Rationale |
|---------|------|-----------|
| Compare runs | **Zones → Strains** tab (Phase 74) — "Compare harvests" | Grow-centric |
| Cost-per-gram story | **Money** workspace — optional **Grows** sub-tab (Phase 72 extension) | Money-centric |
| Cycle summary `/crop-cycles/:id/summary` | Keep as detail route; link from zone Plants tab | Detail pages preserved |

- Remove `cycleCompareRoute` from sidebar **or** redirect Analytics nav item → `/zones?tab=strains&compare=1`.
- Update [`navGroups.js`](../ui/src/lib/navGroups.js) `buildNavGroups(cycleCompareRoute)` — compare route becomes in-workspace only.

---

## WS2 — Guardian nav surfaces

**Today:** three entry points — drawer (edge tab), `/chat` full page, Ask Guardian buttons.

**After [Phase 73](phase_73_guardian_pr_discoverability.plan.md):** global pending badge on launcher.

**Recommended policy:**

| Surface | Role | In sidebar? |
|---------|------|-------------|
| **Drawer** | Primary — ask, confirm PRs, quick context | No (edge tab + TopBar) |
| **`/chat`** | Sessions history, long transcripts, Pending inbox tab | **Demote** — link from drawer "Open full chat →" only |
| **Ask Guardian buttons** | Contextual entry | Keep everywhere |

- Remove **Guardian (full page)** from More if drawer + badge cover 95% of jobs.
- Keep `/chat` route and redirect `/guardian/requests` → `/chat?tab=pending` (Phase 73).
- operator-tour §6: when to open full page vs drawer.

---

## WS3 — Help & reference hub (optional)

Lower priority than WS1–WS2; bundle if scope allows.

| Tab | Source |
|-----|--------|
| **Guide** | [`OperatorGuide.vue`](../ui/src/views/OperatorGuide.vue) |
| **Knowledge** | [`FarmKnowledge.vue`](../ui/src/views/FarmKnowledge.vue) |
| **Catalog** | [`CommonsCatalog.vue`](../ui/src/views/CommonsCatalog.vue) |

Route: `/operator-guide` (keep) with internal tabs **or** new `/help` with redirects.

Enterprise/catalog users keep deep links; sidebar shows one **Help** item.

---

## WS4 — Farm config surface

**Problem:** [`Settings.vue`](../ui/src/views/Settings.vue) is very large; Phase 73 needs **farm lat/long** for weather; modules/timezone scattered.

**Approach:** add **Farm** card on **Today** dashboard (below morning strip) or top of Settings:

- Farm name, timezone, site coordinates (edit inline)
- Active modules toggles (link to Settings detail)
- Link "All settings →" for account, audit, labor rate

Not a new sidebar item — reduces "where is my farm?" hunting.

---

## WS5 — Settings vs Hardware

[Phase 70](phase_70_hardware_pi_control_spa.plan.md) owns device wizard on **Hardware → Pi devices**.

- Settings: remove duplicate "add device" flows; link to `/hardware?tab=devices`.
- Pi integration copy points to Hardware workspace, not Settings.

---

## WS6 — Arc completion verification

Vitest **`phase-77-closure.test.js`**:

- Sidebar item count ≤ 10 (excluding module-gated)
- No Analytics/Guardian/Guide/Knowledge/Catalog in sidebar if WS2–WS3 shipped
- `buildNavGroups` snapshot matches operator-tour "final map"

Update [arc hub](phase_68_73_spa_workspace_roadmap.plan.md) status to **Arc planned through Phase 77**.

---

## Definition of done

- [x] Analytics/compare has workspace home; sidebar Analytics removed or redirected
- [x] Guardian sidebar policy documented and implemented
- [x] (Optional) Help hub merged
- [x] Farm config surfaced on Today or Settings without new nav item
- [x] Settings defers devices to Hardware
- [x] OC-77 closed; operator-tour final sidebar map

---

## Out of scope

- Animals / Aquaponics workspace SPAs — domain modules; stay in More until those products deepen
- Splitting Settings.vue into microservices — future refactor, not required for arc closure

---

## Related

| Doc | Use |
|-----|-----|
| [phase_74_zone_ops_inbox.plan.md](phase_74_zone_ops_inbox.plan.md) | Strains tab for analytics |
| [phase_72_money_unification.plan.md](phase_72_money_unification.plan.md) | Optional Grows sub-tab |
| [phase_73_guardian_pr_discoverability.plan.md](phase_73_guardian_pr_discoverability.plan.md) | Pending badge, site coords nudge |

---

## Using this in a new chat

> Read `docs/plans/archive/phase_77_post_arc_ui_polish.plan.md`. Final arc polish: analytics home, Guardian nav policy, optional help hub, farm config card, Settings trim. Verify ~8-item sidebar.
