---
name: Phase 35 — Lighting domain (photoperiod, presets, timer UX)
overview: >
  Promote lighting from paired cron schedules + generic actuators into a first-class
  domain: lighting_programs with photoperiod windows, standard presets (22/2 peas,
  18/6 veg, 12/12 flower), schedule-action CRUD, timezone-aware automation, and a
  clean Start/End/Duration timer UI (concept from dgang0404_site, rewritten for Vue 3).
  Works for tent, indoor, greenhouse, and field zones — gr33n stays a generic grow program.
todos:
  - id: ws1-schema
    content: "WS1: Schema — gr33ncore.lighting_programs + photoperiod windows; link zone, actuator, optional crop_cycle; paired schedule generation"
    status: done
  - id: ws2-presets
    content: "WS2: Presets — library (22/2, 18/6, 12/12, custom); apply preset → lighting_program + ON/OFF schedules + executable_actions"
    status: done
  - id: ws3-schedule-actions-api
    content: "WS3: Schedule-action API — POST/GET/DELETE /schedules/{id}/actions (parity with rules/programs); worker unchanged dispatch path"
    status: done
  - id: ws4-timezone-worker
    content: "WS4: Timezone — honor schedules.timezone in shouldTriggerNow; farm/zone default TZ"
    status: done
  - id: ws5-timer-ux
    content: "WS5: Timer UX — PhotoperiodClockEditor (start/end/duration linked); LightingProgramForm + zone Lighting tab"
    status: done
  - id: ws6-migrate-seed
    content: "WS6: Migration path — master_seed lighting_program wrap done; jadam_indoor_photoperiod_v1 bootstrap upgraded (OC-35A, 20260603_phase35_oc35a_bootstrap_lighting_programs.sql)"
    status: done
  - id: ws7-guardian-read
    content: "WS7: Guardian read — summarize_zone_lighting read tool; optional create_lighting_program propose tool (medium tier) — defer full revise to Phase 34"
    status: done
  - id: ws8-docs-tests
    content: "WS8: Docs + tests — operator-tour lighting section; OpenAPI; smokes for preset apply + cron fire in TZ"
    status: done
isProject: false
---

# Phase 35 — Lighting domain (photoperiod, presets, timer UX)

## Status

**Shipped — WS6 + WS8 closed (OC-35A–C).** WS1–WS5, WS7, WS6, and WS8 are complete. Track cross-phase rollup in [`phase_35_37_operational_closure.plan.md`](phase_35_37_operational_closure.plan.md).

Depends on **Phase 14** schedules/automation baseline and existing `actuator_type='light'` + `control_actuator` worker path.

**Preconditions:**

- [`gr33ncore.schedules`](../../db/schema/gr33n-schema-v2-FINAL.sql) + [`executable_actions`](../../db/schema/gr33n-schema-v2-FINAL.sql) + worker ([`internal/automation/worker.go`](../../internal/automation/worker.go))
- Actuator + Pi `pending_command` path ([`pi_client/gr33n_client.py`](../../pi_client/gr33n_client.py))
- Zones with `zone_type` including `greenhouse`, `indoor`, etc. ([`internal/handler/zone/handler.go`](../../internal/handler/zone/handler.go))

**Today (gap):** "Lights 18/6" = **two unrelated cron schedules** (ON at 06:00, OFF at 00:00) with actions only in seed/SQL — **no schedule-action UI**, **timezone ignored**, **no photoperiod entity**.

---

## Why this phase

Lighting is required for peas (22h on / 2h off), veg (18/6), flower (12/12), tents, greenhouses, and supplemental field rigs. Operators can toggle actuators manually, but configuring photoperiod is painful and error-prone.

| Today | After Phase 35 |
|-------|----------------|
| Two loose schedules per photoperiod | One **lighting_program** owns ON/OFF pair + actuator |
| Cron evaluated in UTC | **Timezone-aware** triggers (farm/zone TZ) |
| No schedule-action API | **CRUD actions on schedules** like rules/programs |
| Implicit 18h from two crons | Explicit **on_hours / off_hours** + start anchor |
| No presets | **Preset library** (22/2, 18/6, 12/12, custom) |
| dgang0404 timer idea only external | **PhotoperiodClockEditor** in gr33n UI (clean rewrite) |

**Generic grow program:** lighting_program attaches to a **zone** (and optionally an active **crop_cycle**); same zone model as fertigation and greenhouse work.

**UI inspiration (honest):** [`dgang0404_site` option.vue](file:///home/davidg/projects/dgang0404_site/src/components/inv/option.vue) — **Start Time / End Time / Duration** + I/O is the right UX model; **do not port** the Vuex/store code (rewrite for gr33n).

---

## Design principles

1. **First-class domain, thin execution** — `lighting_program` is the operator-facing object; it **generates** paired `schedules` + `executable_actions` the existing worker already runs.
2. **One photoperiod, one pair** — applying a preset creates/updates ON and OFF schedules atomically; disabling a program disables both.
3. **Timezone is not optional for lights** — fix worker to respect `schedules.timezone` (stored column exists today but is ignored).
4. **Presets are data, not magic** — named templates in DB or manifest (`photoperiod_preset` seed rows) operators can clone.
5. **Still actuators underneath** — no new Pi protocol; `control_actuator` `on`/`off` unchanged.
6. **Watering vs fertigation stays separate** — this phase does **not** merge plain RO/well watering (noted for a future inputs/watering slice); lighting only.

---

## Architecture

```
Operator picks preset "18/6 Veg" + zone + grow light actuator
   └─► POST /zones/{id}/lighting-programs
        ├─ gr33ncore.lighting_programs (on_hours=18, off_hours=6, lights_on_at=06:00, tz=America/New_York)
        ├─ schedules: "LP-{id} ON" cron, "LP-{id} OFF" cron (linked via lighting_program_id)
        └─ executable_actions: control_actuator on / off

Worker (30s tick, TZ-aware)
   └─► schedule fires → pending_command → Pi → actuator_events
   (Phase 39 WS1: enqueue via device_commands queue — same ON/OFF payload, no last-write-wins)

UI PhotoperiodClockEditor
   └─► edit start · end · duration (any two compute the third) → PATCH lighting_program → regenerate crons
```

Optional link: `lighting_program.crop_cycle_id` for stage-specific photoperiod (veg vs flower) without new zone.

---

## Scope

| WS | Focus | Primary artifacts |
|----|-------|-------------------|
| **WS1** | Schema | migration `lighting_programs`; sqlc; handler CRUD |
| **WS2** | Presets | preset table or YAML; `ApplyLightingPreset` service |
| **WS3** | Schedule-action API | [`automation/handler.go`](../../internal/handler/automation/handler.go) schedule actions routes |
| **WS4** | Timezone worker | [`worker.go`](../../internal/automation/worker.go) `shouldTriggerNow` + tests |
| **WS5** | Timer UX | `PhotoperiodClockEditor.vue`, `Lighting.vue` or ZoneDetail tab |
| **WS6** | Migrate seed | [`master_seed.sql`](../../db/seeds/master_seed.sql), bootstrap template update |
| **WS7** | Guardian | read tool + optional create tool |
| **WS8** | Docs + tests | operator-tour, OpenAPI, smokes |

---

## Work-stream detail

### WS1 — Schema & lighting_program CRUD

**Goal:** Persistent photoperiod entity linked to zone + actuator.

**Tasks:**

1. Table `gr33ncore.lighting_programs`:
   - `id`, `farm_id`, `zone_id`, `actuator_id` (FK)
   - `name`, `description`
   - `on_hours`, `off_hours` (numeric; validate sum ≤ 24 or allow >24 for specialty — default enforce 24h cycle)
   - `lights_on_at` TIME (anchor; OFF derived or explicit `lights_off_at`)
   - `timezone` TEXT (default from farm)
   - `schedule_on_id`, `schedule_off_id` UUID FK → schedules (nullable until materialized)
   - `crop_cycle_id` optional FK
   - `is_active`, `metadata` JSONB (preset_id, notes)
   - audit columns
2. API: `GET/POST /farms/{id}/lighting-programs`, `GET/PATCH/DELETE /lighting-programs/{id}`, `POST /lighting-programs/{id}/activate|deactivate`
3. Service: create/update regenerates paired schedules + actions transactionally.

**Acceptance:** Create program → two schedules + two `control_actuator` actions exist; deactivate sets `is_active=false` on program and both schedules.

### WS2 — Presets library

**Goal:** One-click standard photoperiods.

**Tasks:**

1. Seed or manifest presets:

   | Preset key | On | Off | Typical use |
   |------------|-----|-----|-------------|
   | `peas_22_2` | 22h | 2h | Peas / long-day veg |
   | `veg_18_6` | 18h | 6h | Vegetative |
   | `flower_12_12` | 12h | 12h | Flowering |
   | `seedling_16_8` | 16h | 8h | Seedlings (optional) |

2. `POST /lighting-programs/from-preset` body: `{preset_key, zone_id, actuator_id, lights_on_at?, timezone?}`
3. Allow **custom** without preset (operator sets hours in UI).

**Acceptance:** `from-preset` with `veg_18_6` yields 18h ON window starting at chosen anchor; cron expressions match documented times in farm TZ.

### WS3 — Schedule-action API

**Goal:** Operators (and Guardian) can attach actions to schedules without raw SQL.

**Tasks:**

1. Mirror program/rule patterns:
   - `GET /schedules/{id}/actions`
   - `POST /schedules/{id}/actions`
   - `PUT /automation/actions/{id}` (exists)
   - `DELETE /automation/actions/{id}` (exists)
2. Schedules UI: expand row → list actions, add `control_actuator` (for advanced users overriding generated pair).

**Acceptance:** POST creates action bound to schedule; worker fires it on cron tick smoke.

### WS4 — Timezone-aware worker

**Goal:** Lighting fires at local wall clock.

**Tasks:**

1. `shouldTriggerNow` loads schedule `timezone`, converts `now` with `time.LoadLocation`, evaluates cron in that location.
2. Default TZ chain: schedule → farm setting → `UTC`.
3. Unit tests: America/New_York 06:00 ON fires at correct UTC instant across DST boundary (at least one case).

**Acceptance:** Schedule with `timezone=America/Los_Angeles` and `0 6 * * *` fires at 06:00 Pacific, not 06:00 UTC.

### WS5 — Photoperiod timer UX

**Goal:** Friendly clock setter; no dgang0404 port.

**Tasks:**

1. `PhotoperiodClockEditor.vue`:
   - Three linked fields: **Start** (lights on), **End** (lights off), **Duration** (on period)
   - Editing any two recomputes the third; 24hr picker; show computed off-hours
   - Preset chips (22/2, 18/6, 12/12) prefill duration
2. `LightingPrograms.vue` or **ZoneDetail → Lighting** tab: list programs, actuator picker, active toggle, link to schedules/events
3. Show next ON/OFF from cron + TZ in plain language ("Next on today at 6:00 AM ET")

**Acceptance:** Vitest: change duration 18h → end time updates; preset chip sets 12/12; form submits PATCH that updates DB crons.

### WS6 — Migration from demo seed / bootstrap

**Tasks:**

1. Replace seed "Light ON/OFF 18/6 Veg" pair with one `lighting_program` + generated schedules (or migration script for existing DBs).
2. Update `jadam_indoor_photoperiod_v1` bootstrap to emit `lighting_program` instead of orphan schedule names.
3. Doc note: existing farms with legacy paired schedules can coexist until operator migrates.

**Acceptance:** Fresh seed has ≥1 lighting_program; demo farm Schedules page shows linked pair under program name.

### WS7 — Guardian (optional slice)

**Tasks:**

1. Read tool `summarize_zone_lighting` — active programs, photoperiod, next trigger, actuator state.
2. Optional medium-tier `create_lighting_program` from preset + zone (propose→Confirm); defer revise-specific lighting edits to Phase 34.

**Acceptance:** Grounded chat "what's the light schedule in Tent A?" includes read-tool block when zone known.

### WS8 — Docs + tests

**Tasks:**

- `operator-tour.md` — "Set up 18/6 lights" with preset + clock editor screenshot path
- `farm-guardian-architecture.md` — lighting in grow environment stack
- OpenAPI schemas: `LightingProgram`, `LightingPreset`, schedule-action paths
- Smokes: create from preset → list → deactivate; TZ cron unit + optional integration

**Acceptance:** Docs + OpenAPI + `go test` green for new packages.

---

## Out of scope (this phase)

- DLI / PAR targets / dimming curves (sensor-driven dim not in v1)
- Greenhouse shade cloth (Phase 36)
- Plain-water irrigation programs (future watering slice; worm casting = new `input_definition` row)
- `schedule_type=interval` / `one_time` worker support

---

## Recommended order

WS1 → WS4 (TZ before trusting cron generation) → WS2 → WS3 → WS5 → WS6 → WS7 → WS8. WS4 can parallel WS1 if cron generation uses TZ at create time only first.

---

## Definition of done (phase ship)

- [x] `lighting_programs` table + CRUD API; preset apply creates paired schedules + actions
- [x] Worker honors `schedules.timezone`
- [x] Schedule-action POST/GET for schedules
- [x] PhotoperiodClockEditor + zone/farm lighting UI
- [x] Demo seed / bootstrap migrated to lighting_program model
- [x] Operator docs + OpenAPI + tests

---

## Using this plan in a new chat

> Implement Phase 35 from `docs/plans/phase_35_lighting_domain.plan.md`. Start WS1 schema + transactional schedule pair generation, then WS4 timezone fix. Build PhotoperiodClockEditor fresh (Start/End/Duration linked) — do not port dgang0404 Vuex code. Presets: peas 22/2, veg 18/6, flower 12/12.

---

## Related

| Doc | Use |
|-----|-----|
| [phase_36_greenhouse_climate.plan.md](phase_36_greenhouse_climate.plan.md) | Shade/fans (complements supplemental light) |
| [phase_34_guardian_pr_iteration.plan.md](phase_34_guardian_pr_iteration.plan.md) | Propose lighting changes with refine loop |
| [pi-integration-guide.md](../pi-integration-guide.md) | Actuator / pending_command execution |
| [phase_37_guardian_offline_field_assistant.plan.md](phase_37_guardian_offline_field_assistant.plan.md) | Wiring the light relay to the Pi = a guided Phase 37 procedure |
| [phase_32_guardian_grow_setup_prs.plan.md](phase_32_guardian_grow_setup_prs.plan.md) | Grow setup; photoperiod in plant meta optional |
| [phase_35_37_operational_closure.plan.md](phase_35_37_operational_closure.plan.md) | Seed / bootstrap / OpenAPI / operator-tour / smokes rollup (35→37) |
