---
name: Phase 47 — Feeding & water in plain language
overview: >
  One farmer job: "How does this room get water?" Unify zone Water tab, feeding plans,
  irrigation-only vs nutrient runs, EC/volume/reservoir/next run — without the six-tab
  fertigation console or schema words (executable_action, schedule_id). Complements Phase 43
  (stock/money); does not replace inventory or mixing logistics.
todos:
  - id: ws1-feeding-plan-model
    content: "WS1: Feeding plan view-model per zone — program + schedule + reservoir + last/next event (client or GET grow-summary)"
    status: completed
  - id: ws2-zone-water-primary
    content: "WS2: Zone Water tab owns daily feeding — story, badges, Run now/pulse; demote Manage→Fertigation"
    status: completed
  - id: ws3-feeding-plan-editor
    content: "WS3: Inline feeding plan editor — volume, EC range, schedule plain time, irrigation-only toggle; PATCH existing APIs"
    status: completed
  - id: ws4-farm-feeding-hub
    content: "WS4: Farm Feeding hub (route) — all rooms as cards; ?zone_id= from 41; Advanced link to full Fertigation.vue"
    status: completed
  - id: ws5-vocabulary-pass
    content: "WS5: farmer-vocabulary.md + grep ban list — no setpoint/cron/executable_action on grow paths; plantNeeds.js copy"
    status: completed
  - id: ws6-guardian-feeding
    content: "WS6: Guardian starters on Water + feeding hub; summarize_zone_fertigation prominence; patch program matchers"
    status: completed
  - id: ws7-docs-tests
    content: "WS7: operator-tour §7b feeding; architecture §7.0m; Vitest feeding plan card; OC-47"
    status: pending
isProject: false
---

# Phase 47 — Feeding & water in plain language

## Status

**In progress — WS1–WS6 shipped.** Ties the farmer UX arc together for **non-technical growers** who understand EC, irrigation, and fertigation in the field but not in the database.

**Roadmap:** [farmer_ux_roadmap_40_plus.plan.md](farmer_ux_roadmap_40_plus.plan.md) (chapter after farm hub; completes the Water story started in Phase 40).

**Next chunk:** WS7 docs/OC-47.

**Does not replace:** [Phase 43](phase_43_operations_stock_feeding_finance.plan.md) (Supplies, farm-wide feeding **admin**, Money). Phase 47 owns **per-room watering**; Phase 43 owns **restock, recipes catalog, receipts**.

---

## The one job

> **"How does this room get water?"**

| Farmer question | They should find it here | Not here (default path) |
|-----------------|--------------------------|-------------------------|
| When is the next feed? | Zone **Water** · Feeding hub card | Schedules table with cron |
| Last feed / did it run? | Water **grow story** | Fertigation → Events tab |
| EC / how much per run? | **Feeding plan** on Water | EC targets table by ID |
| Plain irrigation only? | **Water only** badge (39b) | Mix preview / recipes required |
| Run it now | **Run now** / **Pulse** buttons | Program PATCH via Guardian only |
| Reservoir low? | **Reservoir: needs top-up** | Reservoirs tab first |
| Change the plan | Inline editor on Water | Six-tab Fertigation console |

---

## Problem (why a dedicated phase)

Phases **38–39** shipped honest **pulse**, **queue**, **mix**, and **irrigation_only**. Phase **40 WS5** starts a **grow story** on the Water tab. Farmers still hit:

| Symptom | Root cause |
|---------|------------|
| "What's all this fertigation stuff?" | [Fertigation.vue](../../ui/src/views/Fertigation.vue) — six tabs, `executable_actions`, cron in dropdowns |
| Water tab says "Manage → Fertigation" | [plantNeeds.js](../../ui/src/lib/plantNeeds.js) routes to admin app |
| EC in three places | Program, EC target row, mixing log — no single **feeding plan** |
| Irrigation vs nutrients unclear | 39b badge exists but not the **primary** framing |
| Operator Guide teaches "Setpoint" | Glossary fights grow path ([OperatorGuide.vue](../../ui/src/views/OperatorGuide.vue)) |

Phase **43 WS3** simplifies farm-wide feeding admin — **47** goes deeper on **zone-first** language and makes Water the home for daily feeding.

---

## Position in the arc

```mermaid
flowchart LR
  P40[40 Zone cockpit]
  P41[41 Farm hub]
  P47[47 Feeding and water]
  P42[42 Comfort targets]
  P43[43 Stock and money]
  P44[44 Setup]
  P45[45 Sit-in]
  P46[46 LLM PRs]
  P40 --> P41 --> P47 --> P42 --> P43 --> P44 --> P45 --> P46
```

| Phase | Relationship to 47 |
|-------|-------------------|
| **38–39, 39b** | Runtime: queue, mix, pulse, irrigation_only ✅ |
| **40** | Water tab wedge (WS5) — 47 **completes** narrative + actions |
| **41** | `?zone_id=`, Dashboard links into feeding hub |
| **42** | Climate comfort — parallel concern, not blocked by 47 |
| **43** | Supplies + farm feeding admin — link from 47 "Stock & recipes" |
| **44–46** | Setup, sit-in, Guardian NL — 47 adds feeding starters |

**Recommended order:** **40 → 41 → 47 → 42 → 43 → 44 → 45 → 46**

Rationale: daily **feed/water** is as central as humidity for many farms; shipping 47 before 42 prevents growers living in Fertigation while waiting for comfort-band UI.

---

## Design principles

1. **Room-first** — every feeding screen defaults to **one zone**; farm-wide is a list of rooms.
2. **Feeding plan, not tables** — one card: program name, next run plain English, volume, EC band, reservoir status, irrigation-only badge.
3. **Same APIs** — `fertigation_programs`, `schedules`, `fertigation_events`, run-now, pulse; optional read aggregate only.
4. **Advanced escape hatch** — full Fertigation.vue, mixing log, executable_actions under **Advanced → Feeding (technical)** or Operations (43).
5. **Vocabulary contract** — [farmer-vocabulary.md](../farmer-vocabulary.md) (WS5); sit-in (45) checks grow path.
6. **Guardian secondary** — Run now / pulse on the tab; PR for patch program when chat-first ([42](phase_42_guardian_pr_spec.md) matchers, [46](phase_46_guardian_llm_tool_proposals.plan.md) fallback).

---

## WS1 — Feeding plan view-model

**Goal:** One object per zone for UI and optional API.

| Field | Source |
|-------|--------|
| `program_name` | Active program for zone |
| `irrigation_only` | 39b flag |
| `next_run_at` | Schedule linked to program — humanized in farm TZ |
| `next_run_label` | "Tomorrow at 6:00 AM" |
| `last_event_at` | Last fertigation/irrigation event for zone |
| `last_event_summary` | "Fed 0.3 L · EC 1.2 · OK" |
| `volume_liters` | Program `total_volume_liters` |
| `ec_range` | EC target low/high or program fields |
| `reservoir_status` | ready / low / unknown |
| `queue_depth` | Phase 39 device_commands for zone pump |

**Implementation options:**

| Option | When |
|--------|------|
| **Client compose** | ZoneDetail + farm store already load programs/schedules/events |
| **`GET /zones/{id}/feeding-plan`** | If >5 round trips or mobile perf needs one payload |

**Acceptance:** Flower Room feeding plan object renders without opening Fertigation root.

---

## WS2 — Zone Water tab (primary surface)

**Goal:** Daily feeding happens on **Zones → {room} → Water** — not Fertigation sidebar.

**UI blocks (top to bottom):**

| Block | Copy / behavior |
|-------|-----------------|
| **Status line** | "Next feed: Tomorrow 6:00 AM · 0.3 L · EC 1.1–1.3" |
| **Last feed** | Time + outcome; link "See history" → filtered events (not six tabs) |
| **Feeding plan card** | Edit volume, EC, schedule time (WS3); irrigation-only toggle |
| **Reservoir** | Bar or chip: Ready / Needs top-up |
| **Edge** | Queue: "1 command waiting" (39); honest offline copy |
| **Actions** | **Run feed now** (backlog B1) · **Pulse pump** (38) · **Preview mix** (hidden if irrigation_only) |
| **Stock link** | "Supplies & recipes for this room →" (43) |

**Remove / replace:**

- `plantNeeds.js` manage link: **Feeding plan** (in-tab) not "Fertigation programs"
- ZoneNeedSection: no "Manage → Fertigation" as primary CTA

**Depends on:** Phase 40 WS5 (extend, do not duplicate in 40 — 47 owns completion).

**Acceptance:** Operator answers "when did we last feed Flower Room?" without `/fertigation`.

---

## WS3 — Inline feeding plan editor

**Goal:** Change feed time, volume, EC, pause schedule — on the Water tab.

| Edit | API (existing) |
|------|----------------|
| Volume | PATCH program `total_volume_liters` |
| EC band | PATCH program EC fields or linked EC target |
| Pause / resume | PATCH schedule or program `is_active` |
| Irrigation only | PATCH program `irrigation_only` (39b) |
| Next run time | Plain time picker → cron stored server-side (42 pattern for schedules; reuse helper) |

**Empty state:** "No feeding plan for this room" → **Start feeding plan** wizard (3 steps: name, volume, daily time, irrigation-only?) → POST program + schedule.

**Acceptance:** Change daily feed time without seeing the word `cron_expression` in the UI.

---

## WS4 — Farm Feeding hub

**Goal:** Farm-wide view = **list of rooms**, not six tabs.

**Route:** `/feeding` or `/grow/feeding` (farmer nav: **Feed & water**).

| Card | Content |
|------|---------|
| Per zone | Room name, next run, last feed, irrigation-only badge, alert chip if feed failed |
| Filter | `?zone_id=` from 41 |
| Footer | Advanced → full [Fertigation.vue](../../ui/src/views/Fertigation.vue) (programs, reservoirs, EC targets, mixing log, recipes) |

**Relationship to Phase 43:** 43 **Operations → Feeding (details)** = admin/supplies context; 47 **Feeding hub** = grower daily lens. Can be same route with two modes later — v1: 47 route is card list; 43 links here as default entry.

**Acceptance:** Dashboard "Feeding" chip opens room cards, not tab bar.

---

## WS5 — Farmer vocabulary

**Artifact:** [farmer-vocabulary.md](../farmer-vocabulary.md)

| Banned on grow routes (`/zones`, `/feeding`, zone Water) | Use instead |
|--------------------------------------------------------|-------------|
| Setpoint | Comfort target / target range (climate; 42) |
| Fertigation (nav label) | Feed & water |
| Schedule (alone) | What runs when / next feed |
| Rule (alone) | Automation |
| executable_action | Step in feeding plan |
| predicate, cron_expression | *(not shown)* |
| application_recipe_id | Recipe (linked from Supplies) |

**Tasks:**

1. Publish vocabulary doc + link from roadmap and 45 sit-in.
2. Update `plantNeeds.js`, `ZoneNeedSection.vue`, `navGroups.js` labels.
3. Optional CI: `rg` ban-list on `ui/src/views/Zone*.vue` and feeding routes (warn-only v1).

**Acceptance:** Grow path grep finds zero user-visible "Setpoints →" links.

---

## WS6 — Guardian (feeding)

| Item | Detail |
|------|--------|
| Starters on Water + Feeding hub | "When is the next feed for {zone}?", "Run feed now safe?", "Switch to water-only irrigation" |
| Read | Promote existing `summarize_zone_fertigation` in persona when on Water |
| PR | Reuse `patch_fertigation_program`, `patch_schedule` matchers ([42 spec](phase_42_guardian_pr_spec.md)); gaps → 46 |
| Wizards win | New program via Water wizard (WS3), not setup pack alone |

No new Confirm tools in 47 v1 unless sit-in demands `create_fertigation_program` matcher phrase.

---

## WS7 — Docs, tests, closure (OC-47)

| Doc | Section |
|-----|---------|
| [operator-tour.md](../operator-tour.md) | §7b Feeding & water |
| [farm-guardian-architecture.md](../farm-guardian-architecture.md) | §7.0m |
| [workflow-guide.md](../workflow-guide.md) | Plain irrigation + feeding plan cross-link |
| Vitest | Feeding plan card, irrigation-only hides mix, Run now visible |

**OC-47** row in [phase_35_37_operational_closure.plan.md](phase_35_37_operational_closure.plan.md).

---

## Out of scope (47 v1)

| Item | Owner |
|------|--------|
| Inventory / low-stock / receipts | Phase 43 |
| Closed-loop EC dosing | Tier D |
| New mix calculator schema | Phase 39 ✅ |
| Replacing Guardian | 44, 46 |
| Comfort bands / GH rules | Phase 42 |
| Peristaltic / vendor pump buses | Tier D |

---

## Definition of done

- [ ] Zone Water answers last/next feed and reservoir without `/fertigation`
- [ ] Farm Feeding hub lists rooms as cards; Advanced opens technical Fertigation
- [ ] Irrigation-only farms never see required mix/recipe on Water
- [ ] [farmer-vocabulary.md](../farmer-vocabulary.md) published; grow path ban list applied
- [ ] operator-tour §7b + architecture §7.0m + OC-47
- [ ] Guardian feeding starters shipped (WS6)

---

## Related

| Doc | Use |
|-----|-----|
| [phase_40_unified_farmer_ux_zone_cockpit.plan.md](phase_40_unified_farmer_ux_zone_cockpit.plan.md) | WS5 wedge |
| [phase_39b_plain_irrigation.plan.md](phase_39b_plain_irrigation.plan.md) | irrigation_only |
| [phase_43_operations_stock_feeding_finance.plan.md](phase_43_operations_stock_feeding_finance.plan.md) | Stock & admin |
| [product_backlog_operator_runtime.plan.md](product_backlog_operator_runtime.plan.md) | Run now B1 ✅ |
| [farmer_ux_roadmap_40_plus.plan.md](farmer_ux_roadmap_40_plus.plan.md) | Arc order |
