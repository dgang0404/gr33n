# gr33n Operator Workflow Guide

> **Audience:** farm operators, managers, and anyone learning the platform end-to-end.
>
> **Purpose:** explain in plain language how the moving parts — farms, zones, sensors/actuators, schedules, automation runs, fertigation, crop cycles, plants, tasks, alerts, and costs — fit together and when to use each.
>
> **Format:** this file is structured so each section can be used as a standalone chunk for retrieval-augmented (RAG) help. Section titles stay stable; cross-references use relative paths so they work locally and in chunked form.
>
> **Companion docs:**
> - [`openapi.yaml`](../openapi.yaml) — canonical machine-readable spec of every route.
> - [`pi-integration-guide.md`](pi-integration-guide.md) — how on-farm hardware reaches the platform.
> - [`raspberry-pi-and-deployment-topology.md`](raspberry-pi-and-deployment-topology.md) — Pi OS packages, hosting DB/API/UI together or separately.
> - [`pattern-playbooks.md`](pattern-playbooks.md) — hardware + tuning notes for optional farm **bootstrap templates** (chicken coop, greenhouse, drying room, aquaponics, JADAM indoor starter).
> - [`phase-13-operator-documentation.md`](phase-13-operator-documentation.md), [`phase-14-operator-documentation.md`](phase-14-operator-documentation.md) — phased feature rollouts.

---

## 1. The mental model

A gr33n deployment is one **API** that many **clients** talk to:

```
┌──────────────────────┐                    ┌──────────────────────┐
│  On-farm hardware    │   X-Api-Key        │                      │
│  Raspberry Pi(s)     │ ─────────────────▶ │                      │
│  (sensors, relays)   │                    │                      │
└──────────────────────┘                    │     gr33n API        │   Postgres
                                            │   (Go, net/http)     │ ◀──────▶ DB
┌──────────────────────┐    JWT (Bearer)    │                      │
│  Vue dashboard (UI)  │ ─────────────────▶ │                      │
│  + mobile (PWA/FCM)  │                    │                      │
└──────────────────────┘                    └──────────────────────┘
```

A **farm** is the top-level unit of operation: its own zones, sensors, schedules, costs, alerts, and members. **Organizations** group farms for multi-site tenants and billing but do not change day-to-day operation. Everything below assumes "inside one farm" unless stated otherwise.

### Roles (high level)

- **Owner / Manager** — can edit everything, invite members, change roles, manage billing, opt-in to commons.
- **Operator** — can create/update schedules, zones, sensors, tasks, fertigation entities, costs.
- **Viewer** — read-only across the farm.
- **Pi / MQTT edge** — not a human role; a pre-shared API key that can only post sensor data, heartbeat devices, and record actuator events.

Authorization is enforced per route in `cmd/api/routes.go` via `farmauthz.RequireFarmMember` / `RequireFarmOperate`.

---

## 2. Farm → Zones → Sensors/Actuators

### Zones

Zones are named physical areas inside a farm (e.g. "Veg Room 1", "Greenhouse East"). Every sensor, actuator, and crop cycle lives inside exactly one zone. This is the unit most dashboards pivot on — the **Zones** view lists them; **Zone Detail** shows live readings, active schedules, and crop cycles for one zone.

Key routes: `GET /farms/{id}/zones`, `POST /farms/{id}/zones`, `GET/PUT/DELETE /zones/{id}`.

### Sensors

A **sensor** is a logical channel (e.g. "root-zone EC probe #2"). The Pi reads it on an interval and `POST /sensors/{id}/readings` each time. Readings are stored in `gr33ncore.sensor_readings` and used by:

- Dashboard live SSE stream (`GET /farms/{id}/sensors/stream`).
- Per-sensor history charts (`GET /sensors/{id}/readings`, `.../stats`, `.../latest`).
- Alert rules (thresholds → `sensor_alerts` → push notifications).
- Automation runs that read "the latest EC in zone X".

The Pi can also batch readings (`POST /sensors/readings/batch`) which is how the offline queue drains after reconnect. See [`pi-integration-guide.md`](pi-integration-guide.md) for details.

### Derived channels (optional)

Some channels are **computed on the Pi** from other sensors (for example **dew point**, **VPD**, **heat index** from air temperature + relative humidity). You still register them as normal rows in `gr33ncore.sensors` with a distinct `sensor_type`; the Pi posts readings the same way as for a physical probe. That keeps automation **rules** and **schedule preconditions** unchanged — they keep using the same `{sensor_id, op, value}` predicate shape. Configure the client in `pi_client/config.yaml` under `source: derived` (see [`pattern-playbooks.md`](pattern-playbooks.md) — Greenhouse section and the Pi client source).

### Devices & Actuators

A **device** is a piece of physical hardware (usually the Pi or a microcontroller bridged via MQTT). A device has `online`/`offline` status (`PATCH /devices/{id}/status`) and may own one or more **actuators** (relays, valves, pumps, lights).

Operator flow for controlling an actuator:

1. A **schedule** (cron), a **rule**, or a **fertigation program** tick decides "turn this actuator now" (worker writes the pending payload).
2. The API merges **`pending_command`** into the device’s **`config`** JSONB in Postgres (`devices.config` — see `internal/db` device queries).
3. The **Pi daemon** or an **MQTT → HTTP bridge** polls **`GET /farms/{id}/devices`** (JWT or **`X-API-Key`**, same as sensor ingest — [`pi-integration-guide.md`](pi-integration-guide.md) §7).
4. **Wire format:** the JSON you get from the API encodes `config` as a **base64 string** (Go `json.Marshal` on `[]byte`), not a nested JSON object. Decode it to UTF-8 JSON before reading `pending_command`; the in-repo Pi client does this in `_schedule_loop` ([`pi-integration-guide.md`](pi-integration-guide.md), [`mqtt-edge-operator-playbook.md`](mqtt-edge-operator-playbook.md)).
5. The edge executes the command (GPIO, relay, downstream MQTT), then:
   - **`POST /actuators/{id}/events`** — copy provenance from the pending JSON into **`triggered_by_schedule_id`**, **`triggered_by_rule_id`**, and **`program_id`** as applicable. The API rejects **`triggered_by_rule_id` together with `program_id`** (400); all referenced ids must belong to the **same farm** as the actuator ([`cmd/api/smoke_pi_contract_test.go`](../cmd/api/smoke_pi_contract_test.go) `TestPiContract*`).
   - **`DELETE /devices/{id}/pending-command`** — clears the slot so the command does not repeat.
6. The API fans this out: actuator state, **`GET /schedules/{id}/actuator-events`**, and **`gr33ncore.automation_runs`** / program run rows stay joinable for audit.

The **Schedules** page shows each automation run side-by-side with the actuator events it caused — this is the audit trail for "did the light actually come on at 06:00?".

### Field edge troubleshooting for Pi and MQTT

Use this subsection with [`mqtt-edge-operator-playbook.md`](mqtt-edge-operator-playbook.md) when wiring bridges; it documents the **base64 `config`** wire format and **`GET /farms/{id}/devices`** auth in one place.

| Symptom | What to check |
|---------|----------------|
| Pending command never clears | Edge must **`DELETE`** after success; wrong device id or stuck HTTP errors leave JSON in `config`. |
| `GET /farms/{id}/devices` returns 401/403 | Missing/wrong **`X-API-Key`** or JWT; API needs **`PI_API_KEY`** in non-dev auth modes ([`pi-integration-guide.md`](pi-integration-guide.md#7-pi-api-key-security-middleware-and-least-privilege)). |
| `pending_command` looks missing after decode | Confirm you **base64-decode** `config` from the list response; raw JSON text in DB tools is not the same as the HTTP body. |
| Actuator events not linking to schedules/rules | Echo **`triggered_by_schedule_id`**, **`triggered_by_rule_id`**, **`program_id`** from pending JSON into the POST body (same names); IDs must belong to the actuator’s farm; do not send **`triggered_by_rule_id` together with `program_id`**. |
| MQTT path only | Bridge topic / map issues: [`mqtt-edge-operator-playbook.md`](mqtt-edge-operator-playbook.md) **Troubleshooting** table. |

### Pattern playbooks (bootstrap templates)

When you **create a farm** or **apply a starter pack** (Settings), you can pick a **template key** instead of starting completely blank. Each template seeds zones, sensors, actuators, schedules, rules, and starter tasks appropriate to a pattern (indoor photoperiod + JADAM-style inputs, chicken coop, greenhouse climate, drying room, small aquaponics). Rules are usually seeded **inactive** and schedules **off** until you wire hardware and tune thresholds. See **[`pattern-playbooks.md`](pattern-playbooks.md)** for hardware notes, what each template creates, and how to tune safely.

---

## 3. Schedules & automation runs

The automation worker has **two peer triggers**: the **clock** (schedules) and **sensor state** (rules). Both ultimately write to the same `gr33ncore.automation_runs` table and the same Pi-side actuator pipeline, so operators get one unified audit log regardless of what made the worker act.

### 3a. Schedules (time-driven)

A **schedule** is a cron expression plus a small `meta_data` payload telling the worker what to do. Types in use today include irrigation pulses, alert evaluation, and fertigation program triggers. Schedules belong to a farm (`POST /farms/{id}/schedules`) and can be toggled active/inactive without deleting them (`PATCH /schedules/{id}/active`).

Schedules can also carry **preconditions** — a JSON `{ logic: ALL|ANY, predicates: [{sensor_id, op, value}] }` list that the worker checks before executing. If preconditions fail, the run lands with `status='skipped', details.reason='precondition_failed'` instead of firing the actuator. This is the "interlock lite" guardrail from Phase 19 WS4.

### 3b. Automation rules (sensor-driven)

An **automation rule** fires when sensor state changes, not when a clock ticks. Rules live in `gr33ncore.automation_rules` and are managed via `GET|POST /farms/{id}/automation/rules`, `GET|PUT|DELETE /automation/rules/{id}`, and `PATCH /automation/rules/{id}/active`. Each rule has three pieces:

1. **Trigger** — `trigger_source` says what kind of event the rule listens to (today: `manual_api_trigger` and `sensor_reading_threshold`) plus a `trigger_configuration` JSON for trigger-specific context (e.g. `{ sensor_id: 42 }`).
2. **Conditions** — `conditions_jsonb = { logic: ALL|ANY, predicates: [...] }` using the **same predicate shape as schedule preconditions**. One canonical evaluator (`internal/automation/predicates.go`) runs in both places, so a predicate that works on a schedule guard works identically on a rule.
3. **Actions** — an ordered list of rows in `gr33ncore.executable_actions` attached via `GET|POST /automation/rules/{id}/actions` and mutated via `PUT|DELETE /automation/actions/{id}`. Phase 20 ships three action types:
   - `control_actuator` — writes `pending_command` into the device `config` and logs an `actuator_events` row with provenance (`triggered_by_schedule_id` and/or `triggered_by_rule_id` depending on the trigger).
   - `create_task` — inserts a task whose `source_rule_id` points back at the rule (same provenance pattern as Phase 19's `source_alert_id`).
   - `send_notification` — renders a `notification_templates` row into `alerts_notifications` and fans it through the push pipeline.

   The remaining action-type enum values (`http_webhook_call`, `update_record_in_gr33n`, `trigger_another_automation_rule`, `log_custom_event`) are deferred and the handler rejects them with HTTP 400 so operators can't accidentally ship rules that silently do nothing.

**Dew point / VPD as rule inputs.** Climate rules often key off **dew point** or **vapor pressure deficit (VPD)**. Those can be **computed on the Pi** from air temperature + humidity and ingested as normal sensor readings (see §2 *Derived channels* and [`pattern-playbooks.md`](pattern-playbooks.md)). Predicates stay the usual `{ sensor_id, op, value }` — no special case in the worker.

**Stage-scoped setpoints (Phase 20.6).** Instead of baking a numeric threshold into every rule, operators can store the *ideal environment* for a zone or crop cycle at a given stage as first-class data in `gr33ncore.zone_setpoints` (`GET|POST /farms/{id}/setpoints`, `GET|PUT|DELETE /setpoints/{id}`). A setpoint row carries `sensor_type` (e.g. `dew_point`), optional `min_value` / `max_value` / `ideal_value`, and a scope (`zone_id` and/or `crop_cycle_id`, plus an optional `stage`). Rules can reference them via a second predicate shape — `{ type: "setpoint", sensor_type, scope: "current_stage"|"zone_default", op: "out_of_range"|"below_ideal"|"above_ideal"|"inside_range" }` — and the worker resolves the most-specific row at eval time (precedence: `cycle+stage` > `cycle-any-stage` > `zone+stage` > `zone-any-stage`). If no row matches, the run is recorded as `skipped` with `message='no_setpoint_for_scope'` so the operator knows to configure one rather than thinking the rule failed. The net effect: one rule says *"dew point is out of ideal"* once, and it auto-adjusts as cycles advance through stages.

**Photoperiod / fermented-input automations.** Rules don't know whether the sensor they're reading came from a hydroponic rack, a natural-farming ferment bucket, or a livestock barn — they only see thresholds and predicates. The **JADAM indoor photoperiod starter** bootstrap (see [`pattern-playbooks.md`](pattern-playbooks.md)) ships a canonical schedule + rule set for lighting and fermented-input inventory alerts; the worker runs those rules with the exact same code paths as any other rule. The term *JADAM* here is a proper noun for the method; when we mean the broader product area we say **natural farming** (see [`terminology-guideline.md`](terminology-guideline.md) and the glossary in §10).

Rules honor `cooldown_period_seconds`. Two ticks inside the cooldown window produce one `success` run and one `skipped` run with `details.reason='cooldown'`; once the window elapses the rule can fire again. On a successful tick the worker advances `last_triggered_time`; every tick (fire or not) advances `last_evaluated_time`.

Deleting a rule **cascades** its `executable_actions` (they're meaningless without the parent rule) but **nulls** `tasks.source_rule_id` so operator-facing work survives an administrator tidying up automations. The same `ON DELETE SET NULL` pattern Phase 19 used for `source_alert_id`.

### 3c. The unified run log

The **automation worker** polls due schedules *and* active rules on each tick, and writes one row to `gr33ncore.automation_runs` per trigger — visible via `GET /farms/{id}/automation/runs`. Each run has `status` = `success | partial_success | failed | skipped`, a nullable **`schedule_id`** (cron-driven), **`rule_id`** (sensor-driven), and/or **`program_id`** (fertigation program tick), plus a `details` JSON blob shaped as `{ phase, conditions_met?, actions_total, actions_success, errors: [{action_id, message}], reason?, action_source? }`. Failed runs raise alerts. The UI's Schedules page and the new Automation page both read this table, filtered by `schedule_id` / `rule_id` respectively, so each surface shows only its own history; program fires also populate `program_id` on the same table.

When an automation run switches a relay, that shows up twice:

- In `automation_runs` (what the worker tried to do — regardless of whether the trigger was clock or sensor).
- In `actuator_events` for the affected actuator (what the Pi actually did — see §2). Events can carry **`triggered_by_schedule_id`**, **`triggered_by_rule_id`**, and/or **`program_id`** (from the pending payload / Pi POST) so you can join back to the originating schedule, rule, or fertigation program.

If those two ever disagree, that's a diagnostic signal the hardware is drifting from what the worker asked.

**Tasks linked to schedules and rules.** A task can reference a `schedule_id` ("before the 06:00 irrigation, check tank level") **or** a `source_rule_id` ("this task was auto-created by rule #42 when EC dropped below 1.2"). The UI highlights both on the Tasks page and the relevant source page (Schedules / Automation).

---

## 4. Fertigation: reservoirs, EC targets, programs, mixing, crops

Fertigation is the richest domain in gr33n. It combines real-world mixing (what went into the tank) with what the system told the tank to do.

### Reservoirs

A **reservoir** is a labelled tank (`GET/POST /farms/{id}/fertigation/reservoirs`, `PATCH/DELETE /fertigation/reservoirs/{rid}`). It has a volume, a current nutrient-solution state, and an optional EC target link. Mixing events, fertigation events, and programs all reference a reservoir.

### EC targets

An **EC target** is a named setpoint (e.g. "Veg EC 1.6 mS/cm"). Programs reference targets so that "increase EC for flower" is a single config change, not a fan-out across every schedule.

### Programs

A **fertigation program** is a recipe + EC target + schedule reference. It's the "standard operating procedure" for feeding a zone at a given growth stage. `GET/POST /farms/{id}/fertigation/programs`, `PATCH/DELETE /fertigation/programs/{rid}`.

**Program actions (Phase 20.9 WS4 → Phase 22 WS1/WS2).** A program attaches structured **executable actions** — `control_actuator`, `create_task`, or `send_notification` — through `GET/POST /fertigation/programs/{id}/actions` (list / attach) and `PUT/DELETE /automation/actions/{id}` (edit / detach). Rows live in the same `gr33ncore.executable_actions` table that automation rules write to, with the single-parent CHECK (`rule_id`, `schedule_id`, `program_id` — exactly one) preventing accidental dual-binding. The runtime resolver (`internal/automation.ResolveProgramActions`) prefers DB rows and falls back to synthesising actions from any legacy `metadata.steps` array — but on a freshly-migrated production database, every program should resolve via the DB.

**Worker execution (Phase 22 WS1).** The automation worker now runs a dedicated program tick (`runProgramTick`) every 30 seconds alongside the schedule and rule ticks. For each active program with a bound schedule, the tick:

1. Skips unless the schedule's cron expression fires this minute _and_ `programs.last_triggered_time` isn't already stamped for the current minute.
2. Resolves the action list via `ResolveProgramActions` (logs a structured warning if it takes the `metadata.steps` fallback path so operators can chase stragglers down).
3. Dispatches each action through `dispatchProgramAction`. Actuator commands land on `actuator_events` with `triggered_by_schedule_id = program.schedule_id` plus `meta_data.program_id` for attribution; created tasks carry `schedule_id = program.schedule_id` and a `[program_id=N]` prefix on the description; notifications use `triggering_event_source_type='automation_program'` so the Alerts page can filter.
4. Writes a single `automation_runs` row with the new `program_id` column populated, `details.action_source` set to `executable_actions` or `metadata_steps_fallback`, and a deterministic idempotency key — so a second tick in the same minute is a no-op even if `last_triggered_time` hasn't flushed yet.

**Backfill lifecycle (Phase 22 WS2).** The 20260515 migration installed the `_backfill_program_actions(program_id)` function and ran it over every program in that window. The 20260517 "sweep" migration re-runs the function once more across the whole corpus on deploy, with a per-program `RAISE NOTICE` for any row that actually needed migrating. The function itself is idempotent — re-running it against a program that already has `executable_actions` rows returns `0` inserts — so the sweep is safe on any mature database. Malformed steps continue to be skipped (both the SQL backfill and the Go resolver log a NOTICE / warning instead of aborting) so one bad row never blocks a deploy. The fallback resolver path remains available as a safety net but should be considered deprecated once the sweep has run — the worker log warning is the breadcrumb for ops.

**Worker tick order and monitoring (Phase 23).** The API process runs one automation worker (see `internal/automation/worker.go`). Every **30 seconds** it executes `runTick` in this order: **(1)** cron **schedules** (including precondition checks and schedule idempotency), **(2)** sensor-driven **rules**, **(3)** schedule-bound **fertigation programs** (`runProgramTick`), **(4)** low-stock inventory sweep, **(5)** optional once-per-day electricity rollup after 01:00 UTC. Program ticks intentionally run **after** schedule and rule work so actuator ordering stays predictable when a program shares a schedule with other automation.

**`metadata.steps` fallback — what to grep and what to fix.** When a program still has no rows in `gr33ncore.executable_actions` but usable legacy JSON under `programs.metadata.steps`, the resolver returns `action_source = metadata_steps_fallback`. On each fire the worker prints a line you can grep in API logs:

- `using metadata.steps fallback` — program ID and name are in the message; remediate by adding actions via `POST /fertigation/programs/{id}/actions` or running the SQL backfill helper `_backfill_program_actions(program_id)` (idempotent).

If the top-level `metadata` JSON cannot be parsed, you may see `program N: metadata.steps unusable` — that program yields **no** synthesized actions until an operator repairs `metadata` or attaches DB actions.

**Database checks (same signal as logs).** Program fires always write `gr33ncore.automation_runs` with `program_id` set when the worker evaluated that program. Successful or partial runs include `details.action_source` (`executable_actions` vs `metadata_steps_fallback` vs `empty` on zero-action skips). Example — recent fires that still used the legacy path:

```sql
SELECT id, farm_id, program_id, status, executed_at, message
FROM gr33ncore.automation_runs
WHERE program_id IS NOT NULL
  AND details->>'action_source' = 'metadata_steps_fallback'
ORDER BY executed_at DESC
LIMIT 50;
```

**Other useful log prefixes** (schedules, rules, programs share the same stderr stream): `automation tick failed` (list schedules), `automation rule tick failed`, `automation program tick failed` (list programs), `failed to record automation run` / `failed to record program run` (persistence issues), `transient error (attempt` from schedule/program action retries, `cron parse error` for bad cron text, `precondition_failed` / `skipped: cooldown` (also recorded as `automation_runs` rows where applicable). `GET /automation/worker/health` exposes last tick time and last error string for quick dashboard checks.

### Mixing events (what physically went into the tank)

When an operator mixes a fresh batch of nutrient solution, they record a **mixing event** (`POST /farms/{id}/fertigation/mixing-events`) with:

- the reservoir it went into,
- how much water (and optionally source / starting EC / pH),
- the final EC / pH / temp they measured,
- whether it met the EC target,
- optional **components** — per-input draws like "added 250 ml of FPJ batch #17". Components subtract from natural-farming input inventory so you can see real consumption over a crop cycle.

Line items per mixing event are available at `GET /farms/{id}/fertigation/mixing-events/{mid}/components`.

### Fertigation events (what the zone received)

A **fertigation event** (`POST /farms/{id}/fertigation/events`) is a zone-scoped record of "zone Z received N liters of reservoir R at time T". This is the unit that pairs with a crop cycle so you can ask "how much did Strain A drink this week?". Events can be linked to a `crop_cycle_id` to filter the list (`GET /farms/{id}/fertigation/events?crop_cycle_id=…`).

### Crop cycles & plants

- **Crop cycle** — a run of a crop in one zone from seed/clone to harvest (`gr33n-fertigation`, because EC targets and programs are pinned to cycles). Stages advance via `PATCH /crop-cycles/{id}/stage`.
- **Plant** — a simpler, farm-scoped named entity ("Blueberry Bush, North Row"), useful for perennials, mothers, or catalog bookkeeping that isn't a single cycle. `GET/POST /farms/{id}/plants`, `GET/PUT/DELETE /plants/{id}`.

Both accept arbitrary `meta` JSON for tags, notes, or integrations with the **Commons Catalog** (see §8).

**Stage matters outside EC, too.** EC targets cover the fertigation side of "what should this zone look like right now"; **zone setpoints** (Phase 20.6 — see §3b) cover the environment side per stage (dew point / VPD / temperature ranges, etc.). The two are intentionally separate tables: EC targets drive mixing and programs; setpoints drive rule predicates. When a crop cycle's `current_stage` advances, any setpoint predicate on a rule keyed to that zone immediately starts resolving against the new stage's row — no rule edit required.

**The operator story end-to-end:** set the EC target for "late veg" → assign that target to the program → mix a reservoir (record the mixing event with components) → the program triggers the schedule → the schedule fires an actuator → the Pi reports the actuator event → a fertigation event is recorded against the zone and the active crop cycle. Every step is auditable.

**Fertigation with natural-farming inputs.** Components on a mixing event can draw from either commercial nutrient batches or **natural-farming input batches** (fermented extracts, microbial inoculants, etc.) — the schema doesn't distinguish, it just debits whatever `input_batches.id` you cite. The **JADAM indoor photoperiod starter** bootstrap seeds a handful of JADAM-style inputs (JMS, JLF, FFJ, WCA) so operators following that method have realistic demo data out of the box; operators using other approaches add their own input definitions and the rest of the fertigation pipeline is unchanged. See [`terminology-guideline.md`](terminology-guideline.md) for why we call the API module **natural farming** (the generic umbrella) rather than tying it to any nationality or tradition.

---

## 5. Tasks

**Tasks** are human checklists. Each task has a title, optional description, `status` (`todo | in_progress | on_hold | completed | cancelled | blocked_requires_input | pending_review`), a priority 0–3, an optional `due_date`, and optional links to a **zone** and/or a **schedule**.

Typical uses:

- One-off maintenance ("calibrate pH probe in Veg 1 by Friday").
- Recurring chores attached to a schedule ("check tank before every irrigation cron").
- Bug / operator action items from an alert.

Lifecycle: `POST /farms/{id}/tasks` → `PATCH /tasks/{id}/status` as work progresses → `PUT /tasks/{id}` to edit scope → `DELETE /tasks/{id}` (soft delete) when cancelled or duplicated.

The Tasks page groups by status and priority; the Dashboard shows high-priority and overdue tasks inline so the farm's daily work is visible without digging.

### Labor time tracking (Phase 20.9 WS1–WS2)

Every task carries a **labor log** — a list of `task_labor_log` rows capturing "who worked, when, how long, at what rate." Entries are created either:

- **Manually** — `POST /tasks/{id}/labor` with `started_at`, `ended_at`, `minutes`, and optionally `hourly_rate_snapshot` + `currency`. Good for backfilling work done offline.
- **Via a timer** — `POST /tasks/{id}/labor/start` opens an entry for the logged-in user and leaves `ended_at` NULL; `POST /tasks/{id}/labor/stop` closes it, captures elapsed minutes, and snapshots the user's profile rate (or an explicit override body). A second `start` while one is open 409s; a `stop` with no open entry 404s.

`tasks.time_spent_minutes` is recomputed from the summed `minutes` column after every insert / update / delete (`RecalcTaskTimeSpentMinutes` helper), so the Tasks list can surface cumulative time without joining.

**Auto-cost on close.** When a labor log lands with `ended_at` set and `minutes > 0`, the autologger (`internal/costing.LogLaborEntry`) resolves the rate (log snapshot > profile default > skip), multiplies by minutes/60, and writes one idempotent `cost_transactions` row with `category='labor_wages'`, `related_table_name='task_labor_log'`, `related_record_id=<labor_log.id>`. The idempotency key is `labor:<id>`, so retries and timer-edit round-trips never double-book. Deleting a labor row fires `ReverseLaborEntry`, which stamps a compensating negative row under `labor_void:<id>`; nothing is physically deleted, the ledger nets to zero.

**User rate defaults.** Each user's profile carries `hourly_rate` + `hourly_rate_currency` (nullable, managed via `PATCH /profile/hourly-rate`). The handler enforces that both fields are set or both cleared — a lone rate with no currency is useless to the autologger. Operators without a default rate still show up on labor logs; their time just doesn't produce a cost row until a per-log snapshot is supplied.

---

## 6. Alerts

**Alerts** are automatically generated when a sensor reading crosses a configured threshold or when an automation run fails. They are **not** push notifications in themselves — they're rows in `gr33ncore.sensor_alerts` that:

- drive the unread count in the TopBar (`GET /farms/{id}/alerts/unread-count`),
- list in the Alerts view (`GET /farms/{id}/alerts`),
- fan out to **push notifications** via FCM using tokens the operator registered in Settings → Push tokens.

Operator actions:

- `PATCH /alerts/{id}/read` — clears it from the unread count but keeps the history.
- `PATCH /alerts/{id}/acknowledge` — marks it as acted on (signals "I saw this and handled it").

Because alerts are rows, they can become **tasks** (copy the summary and link the zone/schedule) or **costs** (if resolving required a purchase). The platform does not auto-create those — the operator decides.

Push delivery respects per-user **notification preferences** (`GET/PATCH /profile/notification-preferences`) so operators can mute categories they don't want at 3am.

---

## 7. Costs & finance

A **cost** is any farm-scoped expense or income (`POST /farms/{id}/costs`) with amount, currency, category, date, optional description and optional **receipt attachment**. Costs can be tagged to a zone or crop cycle so margin reporting is possible later.

Key flows:

- **Upload a receipt** — `POST /farms/{id}/cost-receipts` (multipart). The API stores the file in the configured file store and returns an attachment ID, which the cost references.
- **Download** — `GET /file-attachments/{id}/download` (pre-signed) or `.../content` for direct streaming.
- **Export** — `GET /farms/{id}/costs/export` returns a CSV for the accountant.
- **COA mappings** — `GET/PUT /farms/{id}/finance/coa-mappings` maps gr33n categories to a chart-of-accounts, and `DELETE` variants reset either one category or all.
- **Per-cycle P&L** — `GET /crop-cycles/{id}/cost-summary` returns category totals for every cost transaction tagged with that crop cycle (set by the autologger at write time, or manually on legacy rows). This is the first **RAG-precursor lens**: "what did this cycle actually cost me?"

Costs are the one place where the platform intersects external finance; everything else stays inside the gr33n model.

### 7a. Autologged costs (Phase 20.7)

Three flows now write `cost_transactions` rows automatically — operators don't (and shouldn't) hand-enter them:

| Source                          | When it fires                                                                                                                                  | Idempotency key                          | What it does                                                                                                                                   |
| ------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| **Mixing event component**      | Inside the same transaction as `POST /farms/{id}/fertigation/mixing-events` (one row per component).                                           | `mixing_component:<id>`                  | Decrements `gr33nnaturalfarming.input_batches.current_quantity_remaining`; if the input has a `unit_cost`, writes a cost row priced `volume_added_ml × unit_cost`. |
| **Task input consumption**      | `POST /tasks/{id}/consumptions` records a manual draw (e.g. "top-dressed row 3 with 0.5 L of FAA").                                            | `task_consumption:<id>`                  | Same shape as above. The consumption row stores the resulting `cost_transaction_id`. `DELETE /consumptions/{id}` re-credits the batch and writes a paired `[VOIDED]` cost row so net = 0 with an append-only ledger. |
| **Electricity rollup**          | Once per UTC day (~01:00 local), per actuator with `watts > 0` and an active `farm_energy_prices` row.                                         | `electricity:<actuator_id>:<YYYY-MM-DD>` | Reconstructs ON/OFF intervals from `actuator_events.command_sent`, computes `kWh = watts × hours / 1000`, writes a cost row priced `kWh × price_per_kwh`. |

Every auto-write is **idempotent** via `gr33ncore.cost_transaction_idempotency` (PK: `farm_id, idempotency_key`). Replays produce a silent no-op instead of duplicate rows. The `cost_transactions` row carries `related_module_schema` + `related_table_name` + `related_record_id` so the Costs page can render an **`auto · <table>`** chip and operators can filter "auto-logged only".

### 7b. Energy prices

`gr33ncore.farm_energy_prices` is a per-farm rate table with `effective_from` / `effective_to` windows so historical days are priced at the rate in effect that day. Manage via:

- `GET / POST /farms/{id}/energy-prices`
- `PUT / DELETE /energy-prices/{id}`

The Costs page has an inline editor. **No active price = no electricity rollup row** (the worker silently skips farms without one).

### 7c. Low-stock alerts

`gr33nnaturalfarming.input_batches.low_stock_threshold` is opt-in. When `current_quantity_remaining < low_stock_threshold`, the automation worker fires a `medium`-severity alert tagged `triggering_event_source_type = 'inventory_low_stock'`. The Phase 19 "create task from alert" flow turns it into a refill task with one click.

Dedupe: at most one alert per batch per UTC day. The Inventory page shows a **`low`** chip when the batch is below threshold so operators see the state without waiting for the alert.

---

## 8. Commons Catalog & Insert Commons

Two related but distinct systems:

- **Commons Catalog** (`gr33n_inserts`) — a public, browsable library of metadata packs (e.g. starter recipes, input definitions, schedule templates). The UI's **Catalog** view reads `GET /commons/catalog` and `GET /commons/catalog/{slug}`. Operators can **import** a catalog entry into their farm (`POST /farms/{id}/commons/catalog-imports`), and audit history is kept per farm.
- **Insert Commons** (`/farms/{id}/insert-commons/*`) — the opposite direction: the farm can opt-in to publish anonymized bundles of its own schema rows to the commons so other farms benefit. There is a full approve/reject/deliver/export workflow with a bundle audit trail, documented in [`insert-commons-pipeline-runbook.md`](insert-commons-pipeline-runbook.md).

Day-to-day operators mostly use the Catalog to bootstrap a new farm. Insert Commons is an explicit, opt-in act by an owner/manager.

---

## 9. Putting it all together — a day in the life

1. **Morning Dashboard check** — operator opens the dashboard, sees each zone's live temp/humidity/EC, the unread alert count, today's tasks, and the last automation run status.
2. **Alert fires** — a zone's EC drifted low overnight. Push arrived; operator taps it, goes to Alerts, marks **read**, opens the zone detail chart to confirm, and creates a **task** "top up Reservoir A, recheck EC" linked to the zone.
3. **Fertigation** — operator mixes a new batch, records a **mixing event** with components (drawing from natural-farming batches). The program's next cron run (a **schedule**) triggers irrigation; the Pi runs the relay and records an **actuator event**; the API logs a **fertigation event** for the zone and the active **crop cycle**.
4. **Cost** — operator bought new nutrients; uploads the receipt, files the cost under the right COA mapping.
5. **End of day** — Schedules page shows every automation run and the actuator event for each. Tasks page shows progress. Dashboard shows today's readings, today's cost, and any still-unread alerts.

Every action above is recorded as a row in Postgres; nothing depends on an external SaaS for operation. The Pi can even keep posting readings into its offline queue during a network outage and drain the backlog to `/sensors/readings/batch` when connectivity returns.

---

## 10. Animal husbandry & aquaponics

Phase 20.8 lights up the livestock half of the platform on the same three
rails every other subject area uses — **record → consume → cost**.

### 10.1 Animal groups

`gr33nanimals.animal_groups` is the unit of management. One row = one
flock / herd / hive / cohort. A group has:

- `label`, optional `species` (free text — "chicken", "tilapia",
  "honeybee", "cattle"…), and a running `count`.
- An optional `primary_zone_id` — the coop / paddock / tank the group
  lives in. Cross-farm zone ids are rejected at create/update.
- `active` (defaults true) and an `archived_reason` once the group is
  retired — groups are never hard-deleted until the row is soft-deleted
  by the owner.

Routes: `GET /farms/{farm_id}/animal-groups`,
`POST /farms/{farm_id}/animal-groups`,
`GET /animal-groups/{id}`, `PUT /animal-groups/{id}`,
`PATCH /animal-groups/{id}/archive`, `DELETE /animal-groups/{id}`.
UI lives under **Livestock → Animals**.

### 10.2 Lifecycle events — the audit trail

Every non-trivial change to a group is recorded in
`gr33nanimals.animal_lifecycle_events` (added, born, died, sold,
harvested, moved, note). Each row carries a signed `delta_count`, an
optional `notes` / `related_task_id`, and is stamped with
`recorded_by = authctx.UserID`.

The `GET /animal-groups/{id}` detail response returns `delta_total`
(sum of signed deltas) next to the manually edited `count`. The UI
surfaces a reconciliation hint when the two diverge — we intentionally
do **not** overwrite `count` from events, so operators can keep the
displayed headcount authoritative while still logging what actually
happened.

Routes: `GET /animal-groups/{group_id}/lifecycle-events`,
`POST /animal-groups/{group_id}/lifecycle-events`,
`DELETE /lifecycle-events/{id}`.

### 10.3 Feed, bedding, vet supplies → costs

Feed is just a `gr33nnaturalfarming.input_definitions` row with
`category = 'animal_feed'` (bedding → `bedding`, vet drugs →
`veterinary_supply`). When a task consumption drains a batch of one of
these inputs, the Phase 20.7 autologger fires as usual *and* the Phase
20.8 `mapInputCategoryToCostCategory` helper maps the input category
to the right `cost_category_enum`:

| Input category      | Cost category         |
|---------------------|-----------------------|
| `animal_feed`       | `feed_livestock`      |
| `bedding`           | `bedding_supplies`    |
| `veterinary_supply` | `veterinary_services` |
| (all other inputs)  | `natural_farming_inputs` (legacy default) |

That means "feed the layer flock" costs land in the feed rollup
automatically — no separate pathway for livestock costs.

### 10.4 Aquaponics loops

`gr33naquaponics.loops` now has two typed FKs —
`fish_tank_zone_id` and `grow_bed_zone_id` — plus an `active` flag.
One row = one fish-tank/grow-bed pair. The Phase 20.8 bootstrap
upgrade (`_bootstrap_small_aquaponics_v1`) seeds exactly one loop
and back-patches old rows that predate the typed FKs via
`UPDATE … COALESCE`, so re-running a bootstrap on a legacy farm is
safe and lossless.

Routes: `GET /farms/{farm_id}/aquaponics-loops`,
`POST /farms/{farm_id}/aquaponics-loops`,
`GET /aquaponics-loops/{id}`, `PUT /aquaponics-loops/{id}`,
`DELETE /aquaponics-loops/{id}`. UI lives under **Livestock →
Aquaponics**. The loop card links to both zones so zone-level
readings (DO, pH, temperature) are one click away.

### 10.5 Bootstrap idempotency

Both `chicken_coop_v1` and `small_aquaponics_v1` bootstraps are
idempotent on two levels: the dispatcher short-circuits at
`gr33ncore.farm_bootstrap_applications`, **and** the inner
`_bootstrap_*_v1` functions use `NOT EXISTS` / `UPDATE … COALESCE`
guards so a direct re-run writes zero new rows.

### 10.6 Farm knowledge (RAG retrieval)

Phase **24** adds **farm-scoped semantic retrieval** over text that has been **embedded** into Postgres (**pgvector**)
and optional **LLM answer synthesis** that cites those chunks. Nothing trains a model on your data by default;
third-party chat endpoints are an **explicit operator choice** via `LLM_BASE_URL` / `LLM_MODEL` on the API host.

| UI | API (JWT + farm member) |
|----|-------------------------|
| **Monitor → Knowledge** | `GET`/`POST /farms/{id}/rag/search` — vector similarity + optional `module` / date filters on chunk rows |
| Same page: **Ask (LLM)** | `POST /farms/{id}/rag/answer` — retrieve top‑k chunks, then chat completion with bracket citations `[n]` |

Ingestion from operational tables is via the **`rag-ingest`** CLI (see repo `cmd/rag-ingest`). Operator-facing
constraints (PII, secrets, Insert Commons boundaries) are documented in
[`rag-scope-and-threat-model.md`](rag-scope-and-threat-model.md).

---

## 11. Glossary (quick reference)

| Term | Meaning |
|------|---------|
| **Farm** | Top-level operational unit; owns everything else. |
| **Organization** | Group of farms for multi-site tenants (billing & usage). |
| **Zone** | Physical area inside a farm. All sensors / actuators / cycles belong to one zone. |
| **Sensor** | Logical measurement channel. Readings are time-series. |
| **Actuator** | Controllable output (relay, valve, pump, light). |
| **Device** | The hardware running actuators / bridging sensors (usually a Pi). Has online/offline status; automation stores **`pending_command`** inside **`config`** (JSONB in Postgres). In **`GET /farms/{id}/devices`** JSON, `config` is a **base64** encoding of those bytes — decode before reading `pending_command`. |
| **Schedule** | Cron + meta_data; triggers an automation action on a clock. Can carry preconditions (sensor-state interlocks) that must pass before firing. |
| **Automation rule** | Sensor-driven peer of a schedule. Fires when `conditions_jsonb` (ALL/ANY of `{sensor_id, op, value}` predicates) evaluates true, then runs an ordered list of `executable_actions`. Honors `cooldown_period_seconds`. |
| **Executable action** | One step attached to exactly one parent — an automation rule, a schedule, or (Phase 20.9 WS3/WS4) a fertigation program. Action types supported in Phase 20 are `control_actuator`, `create_task`, `send_notification`; the other enum values are reserved for later phases and rejected at creation today. The single-parent CHECK (`chk_executable_action_parent`) prevents a row from binding to more than one source. |
| **Labor log** | Row in `gr33ncore.task_labor_log`. Captures (user, started_at, ended_at, minutes) with an optional (hourly_rate_snapshot, currency). Open entries (NULL `ended_at`) are timers; closed entries (populated `ended_at` + `minutes`) fire the labor autologger. Deletion writes a compensating negative cost row instead of erasing history. |
| **Automation run** | One execution of a schedule, a rule, or a fertigation program tick; has status (`success|partial_success|failed|skipped`), a `details` JSON, and nullable `schedule_id`, `rule_id`, and/or `program_id` pointing back at what triggered it. |
| **Program** | A fertigation recipe/EC-target/schedule triplet. |
| **EC target** | A named EC setpoint (e.g. "flower EC 2.0"). |
| **Zone setpoint** | A stage-scoped row in `gr33ncore.zone_setpoints` that says "for this zone / crop cycle / growth stage, the ideal `sensor_type` value is X (min/ideal/max)". Rules can reference setpoints via a `type: "setpoint"` predicate so one rule auto-adjusts as stages advance. Resolver precedence is `cycle+stage` > `cycle-any-stage` > `zone+stage` > `zone-any-stage`. |
| **Reservoir** | A tank you mix and dispense from. |
| **Mixing event** | "What physically went into the tank" — water + components + measured final EC/pH. |
| **Fertigation event** | "What the zone actually received" — zone-scoped, optionally tied to a crop cycle. |
| **Crop cycle** | A run of a crop in one zone, from start to harvest. Has stages. |
| **Plant** | Named farm-level plant (simpler than a crop cycle — good for perennials). |
| **Task** | Human checklist item; can link to a zone and/or schedule. |
| **Alert** | Auto-generated row from a threshold breach or failed run. Drives push notifications and the bell badge. |
| **Cost** | Farm-scoped expense or income with optional receipt attachment and COA mapping. |
| **Autologger** | The `internal/costing` hook set that turns mixing components, task consumptions, and electricity rollups into idempotent `cost_transactions` rows + inventory deductions. Replays are silent no-ops via `cost_transaction_idempotency`. See §7a. |
| **Energy price** | A row in `gr33ncore.farm_energy_prices` with `effective_from` / `effective_to` and `price_per_kwh`. Required for the nightly electricity rollup. |
| **Low-stock threshold** | Opt-in `gr33nnaturalfarming.input_batches.low_stock_threshold`. When `current_quantity_remaining` drops below it, the worker fires one `medium`-severity alert per batch per UTC day. |
| **Animal group** | Row in `gr33nanimals.animal_groups` — one flock / herd / cohort. Carries `species`, `count`, optional `primary_zone_id`, and an `archived_reason` once retired. See §10.1. |
| **Lifecycle event** | Row in `gr33nanimals.animal_lifecycle_events`. Signed `delta_count` audit entry for added / born / died / sold / moved / note events on an animal group. `recorded_by` is stamped from JWT. See §10.2. |
| **Aquaponics loop** | Row in `gr33naquaponics.loops` pairing a `fish_tank_zone_id` with a `grow_bed_zone_id`. Managed under **Livestock → Aquaponics**; seeded by `small_aquaponics_v1`. See §10.4. |
| **Commons Catalog** | Public library of importable metadata packs. |
| **Insert Commons** | Opt-in farm → commons publishing pipeline. |
| **Natural farming** | Generic English umbrella term used in module titles, API tags, and UI copy for farming that relies on on-site fermented extracts, microbial cultures, and soil amendments (FPJ, FAA, JMS, etc.). Intentionally unqualified — no national / regional / ethnic modifier — because the system doesn't privilege any single tradition. See [`terminology-guideline.md`](terminology-guideline.md). |
| **JADAM** | Proper noun for a specific documented method and its starter cultures (JMS, JLF, FFJ, WCA, …). Used when referring to that method precisely — e.g. the `jadam_indoor_photoperiod_v1` bootstrap template or `reference_source = "JADAM Organic Farming"` seed metadata. Not interchangeable with *natural farming*, which is the broader category. See [`terminology-guideline.md`](terminology-guideline.md). |
| **RAG (retrieval)** | **R**etrieval‑**a**ugmented context: embed farm text into vectors, search by meaning, optionally send ranked chunks to an LLM. UI: **Knowledge**; storage: `gr33ncore.rag_embedding_chunks`. |
| **Embedding chunk** | One row of indexed text + vector for a farm (`source_type`, `source_id`, `chunk_index`, `content_text`, `embedding`). |
| **pgvector** | PostgreSQL extension storing `vector` columns; cosine distance `<=>` ranks neighbors **within the same farm** in API queries. |
| **Knowledge (UI)** | Dashboard screen under **Monitor → Knowledge** calling `/farms/{id}/rag/search` and optionally `/rag/answer`. |

---

## 12. Where to go next

- For **every API contract**: [`openapi.yaml`](../openapi.yaml).
- For **bootstrap templates & wiring patterns**: [`pattern-playbooks.md`](pattern-playbooks.md).
- For **Pi-side flows**: [`pi-integration-guide.md`](pi-integration-guide.md), [`mqtt-edge-operator-playbook.md`](mqtt-edge-operator-playbook.md).
- For **commons publishing**: [`insert-commons-pipeline-runbook.md`](insert-commons-pipeline-runbook.md), [`commons-catalog-operator-playbook.md`](commons-catalog-operator-playbook.md).
- For **alerts and push**: [`notifications-operator-playbook.md`](notifications-operator-playbook.md).
- For **phased feature history**: [`phase-13-operator-documentation.md`](phase-13-operator-documentation.md), [`phase-14-operator-documentation.md`](phase-14-operator-documentation.md).
- For **RAG scope, data classes, and LLM egress**: [`rag-scope-and-threat-model.md`](rag-scope-and-threat-model.md).
