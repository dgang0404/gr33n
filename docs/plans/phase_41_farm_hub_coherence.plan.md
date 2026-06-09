---
name: Phase 41 — Farm hub coherence (Dashboard + farm-wide pages)
overview: >
  After the zone cockpit (Phase 40), unify farm-wide surfaces so operators are not bounced
  between Dashboard, Fertigation tabs, Events, Schedules, and Tasks as separate "apps."
  Zone remains the grow lens; farm pages become contextual hubs with deep links and
  shared summaries — not a second schema pass.
todos:
  - id: ws1-dashboard-morning
    content: "WS1: Dashboard morning cockpit — farm-wide chips (tasks due, unread alerts, next schedule, queue depth); reuse grow-summary or parallel aggregate API"
    status: completed
  - id: ws2-fertigation-zone-context
    content: "WS2: Fertigation hub zone context — ?zone_id= preserves lens; default tab/events filter; 'back to zone Water' strip"
    status: completed
  - id: ws3-cross-page-deep-links
    content: "WS3: Schedules / Tasks / Events / Alerts — zone filter query param + breadcrumb when opened from zone"
    status: completed
  - id: ws4-why-empty-ux
    content: "WS4: Why-empty inline hints — reusable EmptyStateHint component (no telemetry / no setpoint / automation off / wrong farm)"
    status: completed
  - id: ws5-seed-zone-tasks
    content: "WS5: Seed + demo data — tasks with zone_id for zone Overview strip (unblocks 40 WS6 demos)"
    status: completed
  - id: ws6-lighting-farm-link
    content: "WS6: /lighting ↔ zone Light tab — consistent copy; zone page links program; lighting page shows zone filter"
    status: completed
  - id: ws7-docs-tests
    content: "WS7: operator-tour §3 farm hub; architecture §7.0g; Vitest dashboard + fertigation zone context; OC-41 closure"
    status: completed
isProject: false
---

# Phase 41 — Farm hub coherence

## Status

**Shipped.** OC-41 closed (WS7). **Historical:** ship-order notes in this plan are deprecated.

**Indexed in:** [`pre_development_gaps_index.plan.md`](pre_development_gaps_index.plan.md) (gap **A3**, archived).

---

## Problem

Phase 40 fixes **Zones → Overview / Water / Light / Climate**. Operators still report:

- **Dashboard** feels like a link farm, not “what matters this morning.”
- **Fertigation** (Programs, Events, Reservoirs, EC targets) feels disconnected from zone **Water**.
- **Schedules / Tasks / Events** do not remember they came from **Flower Room**.
- **Empty lists** with no hint (documented in [operator-tour §4](../operator-tour.md#4-why-is-this-empty-future-ux), [sit-in §1](../workstreams/sit-in-operator-experience.md) — **not implemented** in UI).

Phase 40 **out of scope** deliberately kept farm-wide Advanced CRUD. Phase 41 **coordinates** those pages without replacing them.

---

## Design principles

1. **Zone-first entry, farm-wide for bulk** — same as Phase 40.
2. **Query-param context** — `?zone_id=` on farm routes; no new session store required for v1.
3. **Reuse aggregates** — prefer `GET …/zones/{id}/grow-summary` from Phase 40 WS1; farm dashboard may use `GET …/farms/{id}/farm-summary` only if client N+1 is painful.
4. **Plain language** — match Phase 40 copy rules.
5. **Why-empty is educational, not apologetic** — one sentence + optional action link (“Add setpoint”, “Check Pi”, “Enable rule”).

---

## WS1 — Dashboard morning cockpit

**Goal:** `/` answers “what should I do first?” without opening six sidebar items.

**UI (top of Dashboard, below farm header):**

| Chip / row | Source |
|------------|--------|
| Tasks due today | `tasks` filtered by farm, status + due date |
| Unread alerts | `alerts_notifications` count |
| Next schedule fire | farm schedules — soonest next run (reuse cron humanize from 40 WS3) |
| Devices offline | device heartbeat summary |
| Queue pending (post-39) | sum `device_commands` pending per farm |

**Actions:** Each chip links to the right page **with filters pre-applied** (WS3).

**Acceptance:** Demo farm Dashboard shows ≥2 non-zero chips with seed data.

---

## WS2 — Fertigation hub zone context

**Goal:** Fertigation is “details for this room,” not a parallel product.

**Tasks:**

1. Accept `?zone_id=` on `/fertigation` and sub-routes (`?tab=events`).
2. Banner: “Viewing **Flower Room** — [Back to zone Water →](/zones/{id}?tab=water)”.
3. Default **Events** tab filter: events for actuators/programs linked to zone (client filter v1).
4. Programs list: highlight programs tied to zone actuators or crop cycles in zone.

**Acceptance:** From zone Water “Fertigation details →”, Events tab opens filtered; back link returns to zone.

---

## WS3 — Cross-page deep links

**Goal:** Schedules, Tasks, Events, Alerts honor zone context.

**Routes:** `/schedules`, `/tasks`, `/automation`, `/alerts` — read `zone_id` query param.

**UI:**

- Filter lists client-side (or server `?zone_id=` if API exists).
- Breadcrumb: `Zones › Flower Room › Tasks`.
- Zone Overview links use `?zone_id=` form.

**Acceptance:** Open Tasks from zone Overview; list shows only that zone’s tasks when `zone_id` set.

---

## WS4 — Why-empty inline hints

**Goal:** Close sit-in “why empty” backlog with one reusable pattern.

**Component:** `EmptyStateHint.vue` (or extend existing empty states) — props: `reason` enum:

| Reason | Copy pattern | Suggested action |
|--------|----------------|------------------|
| `no_data` | Nothing recorded yet for this farm. | Link to bootstrap / add zone |
| `no_telemetry` | No recent readings — check edge device. | Pi integration guide |
| `no_setpoint` | No target band for this sensor type. | Inline add (40 WS2) or Setpoints |
| `automation_off` | Rules/schedules exist but are inactive. | Toggle or Schedules |
| `wrong_farm` | (rare) Select another farm in header. | — |

**Rollout (v1 pages):** Dashboard widgets, ZoneNeedSection empty rows, Fertigation Events, Tasks board, Alerts list.

**Acceptance:** Flower Room with no humidity setpoint shows `no_setpoint` hint, not blank card.

---

## WS5 — Seed + demo `zone_id` on tasks

**Goal:** Phase 40 WS6 and Dashboard task chips have demo data.

**Tasks:**

1. Audit `master_seed.sql` tasks — set `zone_id` for representative open tasks per zone.
2. Document in operator-tour §4b/§3 that demo tasks appear on zone Overview.

**Acceptance:** ≥1 open task on Flower Room Overview after fresh seed.

---

## WS6 — Lighting farm page ↔ zone Light tab

**Goal:** `/lighting` and zone **Light** tab tell one story.

**Tasks:**

1. Zone Light tab: show linked `lighting_program` summary (name, ON/OFF window) — read-only if program managed on `/lighting`.
2. `/lighting`: optional `?zone_id=` filter; “Open zone →” link.
3. Align HelpTip copy with operator-tour §5.

**Acceptance:** Veg Room Light tab names the same program as Lighting page for that zone.

---

## WS7 — Docs, tests, closure (OC-41)

| Layer | Artifacts |
|-------|-----------|
| **operator-tour** | §3b farm hub + morning path (extends tasks-first guide; §3 stays data-flow) |
| **architecture** | §7.0g — farm hub vs zone cockpit |
| **Vitest** | Dashboard chips; Fertigation `?zone_id=` banner; EmptyStateHint |
| **Closure** | OC-41 row in [`phase_35_37_operational_closure.plan.md`](phase_35_37_operational_closure.plan.md) |

---

## Out of scope (v1)

- Merging Fertigation tabs into a single Vue route (still tabbed; only context + filters)
- Replacing Tasks Kanban with zone-only app
- New notification types
- Enterprise multi-site dashboard (see hypothetical topology doc)

---

## Recommended order

WS4 (component) → WS5 (seed) → WS1 → WS3 → WS2 → WS6 → WS7

WS4 + WS5 unblock demos; WS1 delivers morning cockpit value.

---

## Definition of done

- [x] Dashboard morning strip with tasks, alerts, schedule, devices, queue
- [x] Fertigation respects `?zone_id=` + back to zone Water
- [x] Tasks/Schedules/Alerts accept zone filter from zone links
- [x] Why-empty hints on ≥5 major surfaces
- [x] Demo tasks carry `zone_id`
- [x] Lighting page ↔ zone Light consistent
- [x] operator-tour §3 + architecture §7.0g + Vitest + OC-41

---

## Related

| Doc | Use |
|-----|-----|
| [phase_40_unified_farmer_ux_zone_cockpit.plan.md](phase_40_unified_farmer_ux_zone_cockpit.plan.md) | Prerequisite zone cockpit |
| [phase_39_edge_fertigation_execution.plan.md](phase_39_edge_fertigation_execution.plan.md) | Queue on Dashboard chip |
| [pre_development_gaps_index.plan.md](pre_development_gaps_index.plan.md) | Master gap list |
| [tasks-first-operator-guide.md](../tasks-first-operator-guide.md) | Morning path baseline |

---

## Using this plan in a new chat

> **Shipped.** For historical context only — do not treat dependency prose as a live gate. Active work: [68–73 SPA arc](phase_68_73_spa_workspace_roadmap.plan.md).
