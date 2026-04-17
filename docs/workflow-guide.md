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

### Devices & Actuators

A **device** is a piece of physical hardware (usually the Pi or a microcontroller bridged via MQTT). A device has `online`/`offline` status (`PATCH /devices/{id}/status`) and may own one or more **actuators** (relays, valves, pumps, lights).

Operator flow for controlling an actuator:

1. A **schedule** (cron) or a rule decides "turn light on now".
2. The API writes `config.pending_command` onto the device row.
3. The Pi's polling loop picks it up via `GET /farms/{id}/devices`.
4. The Pi executes it on GPIO, then:
   - `POST /actuators/{id}/events` (records what it actually did, with `schedule_id` if applicable),
   - `DELETE /devices/{id}/pending-command` (so it doesn't run twice).
5. The API fans this out: actuator state changes, Schedules page picks up the event via `GET /schedules/{id}/actuator-events`, automation run log updates.

The **Schedules** page shows each automation run side-by-side with the actuator events it caused — this is the audit trail for "did the light actually come on at 06:00?".

---

## 3. Schedules & automation runs

A **schedule** is a cron expression plus a small `meta_data` payload telling the worker what to do. Types in use today include irrigation pulses, alert evaluation, and fertigation program triggers. Schedules belong to a farm (`POST /farms/{id}/schedules`) and can be toggled active/inactive without deleting them (`PATCH /schedules/{id}/active`).

The **automation worker** polls due schedules, runs their action, and writes one row to `gr33ncore.automation_runs` per trigger — visible via `GET /farms/{id}/automation/runs`. Each run has `status` = `success | partial_success | failed | skipped` and a `details` JSON blob. Failed runs raise alerts.

When an automation run switches a relay, that shows up twice:

- In `automation_runs` (what the worker tried to do).
- In `actuator_events` for the affected actuator (what the Pi actually did — see §2).

If those two ever disagree, that's a diagnostic signal the hardware is drifting from what the scheduler asked.

**Tasks linked to schedules.** A task can reference a `schedule_id` (e.g. "before the 06:00 irrigation, check tank level"). The UI highlights these on the Schedules page and the Tasks page so operators can see what manual work is tied to which automation.

---

## 4. Fertigation: reservoirs, EC targets, programs, mixing, crops

Fertigation is the richest domain in gr33n. It combines real-world mixing (what went into the tank) with what the system told the tank to do.

### Reservoirs

A **reservoir** is a labelled tank (`GET/POST /farms/{id}/fertigation/reservoirs`, `PATCH/DELETE /fertigation/reservoirs/{rid}`). It has a volume, a current nutrient-solution state, and an optional EC target link. Mixing events, fertigation events, and programs all reference a reservoir.

### EC targets

An **EC target** is a named setpoint (e.g. "Veg EC 1.6 mS/cm"). Programs reference targets so that "increase EC for flower" is a single config change, not a fan-out across every schedule.

### Programs

A **fertigation program** is a recipe + EC target + schedule reference. It's the "standard operating procedure" for feeding a zone at a given growth stage. `GET/POST /farms/{id}/fertigation/programs`, `PATCH/DELETE /fertigation/programs/{rid}`.

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

**The operator story end-to-end:** set the EC target for "late veg" → assign that target to the program → mix a reservoir (record the mixing event with components) → the program triggers the schedule → the schedule fires an actuator → the Pi reports the actuator event → a fertigation event is recorded against the zone and the active crop cycle. Every step is auditable.

---

## 5. Tasks

**Tasks** are human checklists. Each task has a title, optional description, `status` (`todo | in_progress | on_hold | completed | cancelled | blocked_requires_input | pending_review`), a priority 0–3, an optional `due_date`, and optional links to a **zone** and/or a **schedule**.

Typical uses:

- One-off maintenance ("calibrate pH probe in Veg 1 by Friday").
- Recurring chores attached to a schedule ("check tank before every irrigation cron").
- Bug / operator action items from an alert.

Lifecycle: `POST /farms/{id}/tasks` → `PATCH /tasks/{id}/status` as work progresses → `PUT /tasks/{id}` to edit scope → `DELETE /tasks/{id}` (soft delete) when cancelled or duplicated.

The Tasks page groups by status and priority; the Dashboard shows high-priority and overdue tasks inline so the farm's daily work is visible without digging.

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

Costs are the one place where the platform intersects external finance; everything else stays inside the gr33n model.

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

## 10. Glossary (quick reference)

| Term | Meaning |
|------|---------|
| **Farm** | Top-level operational unit; owns everything else. |
| **Organization** | Group of farms for multi-site tenants (billing & usage). |
| **Zone** | Physical area inside a farm. All sensors / actuators / cycles belong to one zone. |
| **Sensor** | Logical measurement channel. Readings are time-series. |
| **Actuator** | Controllable output (relay, valve, pump, light). |
| **Device** | The hardware running actuators / bridging sensors (usually a Pi). Has online/offline status and a `pending_command` slot. |
| **Schedule** | Cron + meta_data; triggers an automation action. |
| **Automation run** | One execution of a schedule; has status and a details payload. |
| **Program** | A fertigation recipe/EC-target/schedule triplet. |
| **EC target** | A named EC setpoint (e.g. "flower EC 2.0"). |
| **Reservoir** | A tank you mix and dispense from. |
| **Mixing event** | "What physically went into the tank" — water + components + measured final EC/pH. |
| **Fertigation event** | "What the zone actually received" — zone-scoped, optionally tied to a crop cycle. |
| **Crop cycle** | A run of a crop in one zone, from start to harvest. Has stages. |
| **Plant** | Named farm-level plant (simpler than a crop cycle — good for perennials). |
| **Task** | Human checklist item; can link to a zone and/or schedule. |
| **Alert** | Auto-generated row from a threshold breach or failed run. Drives push notifications and the bell badge. |
| **Cost** | Farm-scoped expense or income with optional receipt attachment and COA mapping. |
| **Commons Catalog** | Public library of importable metadata packs. |
| **Insert Commons** | Opt-in farm → commons publishing pipeline. |

---

## 11. Where to go next

- For **every API contract**: [`openapi.yaml`](../openapi.yaml).
- For **Pi-side flows**: [`pi-integration-guide.md`](pi-integration-guide.md), [`mqtt-edge-operator-playbook.md`](mqtt-edge-operator-playbook.md).
- For **commons publishing**: [`insert-commons-pipeline-runbook.md`](insert-commons-pipeline-runbook.md), [`commons-catalog-operator-playbook.md`](commons-catalog-operator-playbook.md).
- For **alerts and push**: [`notifications-operator-playbook.md`](notifications-operator-playbook.md).
- For **phased feature history**: [`phase-13-operator-documentation.md`](phase-13-operator-documentation.md), [`phase-14-operator-documentation.md`](phase-14-operator-documentation.md).
