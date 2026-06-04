---
name: Phase 42 — Comfort targets & automation (plain language)
overview: >
  Replace the "Setpoints / Schedules / Rules are a database admin panel" feel with farmer
  jobs: comfort bands per room, "what runs when," and safe automation toggles. Reuses
  zone_setpoints, schedules, automation_rules, executable_actions APIs — no schema churn
  unless a read aggregate is required.
todos:
  - id: ws1-comfort-targets-hub
    content: "WS1: /comfort-targets (or Grow → Targets) — farm + zone list; plain labels; hide scope jargon behind chips"
    status: pending
  - id: ws2-band-editor
    content: "WS2: Band editor component — too low / ideal / too high per sensor need; stage chip; PATCH existing setpoint APIs"
    status: pending
  - id: ws3-schedules-plain
    content: "WS3: Schedules farmer view — humanized next run, linked program/lighting name; simple create (time picker not cron) for common cases"
    status: pending
  - id: ws4-rules-plain
    content: "WS4: Rules farmer view — one-line what-it-does; active toggle; template picker for GH/lighting; Advanced link for predicate JSON"
    status: pending
  - id: ws5-advanced-escape
    content: "WS5: Collapse raw CRUD under Advanced → Power settings; redirects from old URLs"
    status: pending
  - id: ws6-guardian-align
    content: "WS6: Guardian persona + tools prefer comfort-band language; patch_schedule/patch_rule summaries"
    status: pending
  - id: ws7-docs-tests
    content: "WS7: operator-tour §6 comfort; architecture §7.0h; Vitest band editor; OC-42 closure"
    status: pending
isProject: false
---

# Phase 42 — Comfort targets & automation (plain language)

## Status

**Planned.** After [Phase 40 WS2](phase_40_unified_farmer_ux_zone_cockpit.plan.md#ws2--inline-setpoint-editor) (inline wedge) and [Phase 41](phase_41_farm_hub_coherence.plan.md) (hub + why-empty).

**Roadmap:** [farmer_ux_roadmap_40_plus.plan.md](farmer_ux_roadmap_40_plus.plan.md)

---

## Problem

Operators report **setpoints are hard to understand**:

- Scope (`zone_id` vs `crop_cycle_id` vs `stage`) reads like a schema exercise.
- Rules and schedules expose **cron**, **predicate JSON**, and **executable_action** vocabulary.
- Phase 40 fixes **inline edit in the zone** but leaves `/setpoints`, `/automation`, `/schedules` as power-user tables.

---

## Design principles

1. **Same data, different words** — DB still has `zone_setpoints`; UI says **comfort band** or **target range**.
2. **Cron is an implementation detail** — farmers pick "Every day at 6:00 AM" (timezone-aware); store cron via existing schedule API.
3. **Rules = sentences** — "If humidity stays above the comfort band for 10 minutes → notify me."
4. **Advanced is explicit** — link "Edit technical rule →" opens today's RuleForm.

---

## WS1 — Comfort targets hub

**Route:** `/comfort-targets` or **Grow → Targets** (nav TBD with 40 WS7).

| UI | API |
|----|-----|
| List zones with band status (ok / missing / out of range) | `GET /farms/{id}/setpoints`, zone sensors, recent readings |
| Filter by zone | query param |
| Empty → why-empty `no_setpoint` (41 WS4) | — |

**Acceptance:** Operator finds Flower Room humidity band without opening Advanced Setpoints table.

---

## WS2 — Band editor

**Component:** `ComfortBandEditor.vue` — used in zone tabs (40) and hub (42).

| Field | Farmer label |
|-------|----------------|
| min | Too low (alert below) |
| ideal | Just right |
| max | Too high (alert above) |
| stage | "For **mid flower** stage" (from active cycle) |

Save → existing `POST/PATCH /setpoints` or farm setpoint routes.

**Acceptance:** No raw `zone_setpoints` column names in UI.

---

## WS3 — Schedules (plain)

**Route:** `/schedules` becomes farmer-first; cron hidden behind "Edit schedule times → Advanced".

| Feature | Notes |
|---------|--------|
| List | Name, **next run in plain English**, linked fertigation/lighting program |
| Create (simple) | Daily / weekly time + timezone; maps to cron server-side |
| Create (advanced) | Link to current schedule form |

Reuse `GET/POST/PUT /schedules`, lighting program schedule IDs.

---

## WS4 — Rules (plain)

| Feature | Notes |
|---------|--------|
| List | Human summary from predicate (reuse or extend `ruleSummary` helper) |
| Toggle active | existing PATCH |
| Templates | GH templates (36), lighting interlocks — "Add shade on hot day" wizard |
| Create | Wizard steps: trigger → action → zone; writes same rule rows |

**Acceptance:** Non-technical tester can disable a rule without reading JSON.

---

## WS5 — Advanced escape hatch

- Nav group **Advanced → Power settings**: Setpoints (raw table), Automation (full RuleForm), Schedules (cron editor).
- Old bookmarks keep working; banner: "You're in power-user mode."

---

## WS6 — Guardian alignment

- System prompt: prefer **comfort band** / **schedule plain English** when proposing patches.
- Impact lines for `patch_rule` / `patch_schedule` use farmer summaries.

---

## WS7 — Docs, tests, closure (OC-42)

| Artifact | |
|----------|--|
| operator-tour | §6 — comfort bands + "what runs when" |
| architecture | §7.0h — comfort vs Advanced |
| Vitest | Band editor save; schedule humanize |
| Smokes | Optional — setpoint PATCH unchanged |

---

## Out of scope

- New automation engine
- Replacing Confirm for Guardian writes
- Merging setpoints into a new table (use existing `zone_setpoints`)

---

## Definition of done

- [ ] Comfort targets hub + band editor labels
- [ ] Schedules list/create without exposing cron to farmers
- [ ] Rules list/toggle with plain summaries
- [ ] Advanced group contains legacy CRUD
- [ ] operator-tour §6 + OC-42
