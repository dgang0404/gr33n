---
name: Phase 19 Safety & Alert Rules
overview: >
  Foundation phase after Phase 18. Strengthens what already exists — no new big tables.
  Get CI green, teach existing alerts to respect duration/cooldown, let operators turn
  alerts into tasks in one click, and add a lightweight schedule precondition so the
  worker refuses to fire actuators when the real-world state is wrong ("interlock lite").
  Target: 4–6 days.
todos:
  - id: ws1-bootstrap-smoke
    content: "WS1: Fix TestFarmBootstrapOnCreate + TestOrgDefaultBootstrapOnFarmCreate (zone-count vs template) so go test ./cmd/api/... is green"
    status: completed
  - id: ws2-alert-duration
    content: "WS2: Alert duration + cooldown — add alert_duration_seconds/alert_cooldown_seconds to sensors; evaluator requires the breach to persist; UI form on sensor edit"
    status: completed
  - id: ws3-alert-to-task
    content: "WS3: Alert→Task — POST /alerts/{id}/create-task (or UI-side helper) pre-fills title/zone/priority; add source_alert_id on tasks; Alerts page button + task badge"
    status: completed
  - id: ws4-schedule-preconditions
    content: "WS4: Schedule preconditions (interlock lite) — preconditions JSON on schedules; worker evaluates before executing; run status=skipped reason=precondition_failed; UI editor on schedule form"
    status: completed
isProject: false
---

# Phase 19 — Safety & Alert Rules

## Why this phase

Phase 18 aligned the spec, shipped operator docs, and polished the UI. The platform now tells the operator *what happened* cleanly. Phase 19 is the first step toward **making it safer**:

1. **CI honesty** — two pre-existing bootstrap tests are red. Fix them before we add more tests on top.
2. **Alerts that don't spam** — today a single noisy reading fires an alert. We want "low for 10 minutes, then alert, then stay quiet for an hour."
3. **Alerts that drive action** — an alert should become a tracked task in one click, with a back-pointer for audit.
4. **Guardrails on schedules** — the worker should refuse to fire a pump if the tank is empty. The full automation-rule engine lands in Phase 20; this phase ships the minimum guard on the cron path that already exists.

No new big tables — every change is a column on an existing table plus small handler/UI updates.

## Scope

| WS | Focus | Location in repo |
|----|--------|------------------|
| **WS1** | Green CI — bootstrap smoke tests | `cmd/api/smoke_test.go`, `ui/src/constants/bootstrapTemplates.js` *or* seed/template |
| **WS2** | Alert duration + cooldown | `db/migrations/…`, `db/schema/gr33n-schema-v2-FINAL.sql`, `db/queries/sensors.sql`, `internal/handler/sensor/handler.go`, alert evaluator, `ui/src/views/SensorDetail.vue` (+ sensor edit form) |
| **WS3** | Alert → Task | `db/migrations/…` (add `source_alert_id` to tasks), `db/queries/tasks.sql`, `internal/handler/alert/handler.go` (new endpoint) or UI-only helper, `ui/src/views/Alerts.vue`, `ui/src/views/Tasks.vue` |
| **WS4** | Schedule preconditions | `db/migrations/…`, `db/queries/automation.sql`, `internal/handler/automation/handler.go`, `internal/automation/worker.go` (the evaluator), `ui/src/views/Schedules.vue`, `openapi.yaml` |

## Work-stream detail

### WS1 — Fix bootstrap smoke failures

- Investigate `TestFarmBootstrapOnCreate` and `TestOrgDefaultBootstrapOnFarmCreate` (zone count). Two acceptable fixes:
  - Update the org-default and farm-default templates in `ui/src/constants/bootstrapTemplates.js` / the server-side seed so they produce ≥ 4 zones (and match test expectations), **or**
  - Lower the assertions to reflect the intended minimum-viable template.
- Prefer the template change if 4 zones better matches the expected operator setup.
- Run `go test ./cmd/api/...` to confirm all tests green.

### WS2 — Alert duration + cooldown

- **Schema (migration):**
  - `gr33ncore.sensors.alert_duration_seconds INTEGER DEFAULT 0 NOT NULL`
  - `gr33ncore.sensors.alert_cooldown_seconds INTEGER DEFAULT 300 NOT NULL`
- **Evaluator:** the existing alert path in `internal/handler/sensor/handler.go` (PostReading / PostReadingsBatch) evaluates thresholds today. Extend it to:
  - Track per-sensor breach-start time (add a state column or derive from the last N readings).
  - Only emit an alert if `now - breach_start >= alert_duration_seconds`.
  - After emitting, suppress duplicates for `alert_cooldown_seconds` using `alerts_notifications.created_at`.
- **API:** surface the two new fields in sensor create / update; update `openapi.yaml` (`Sensor` + `SensorCreate` / `SensorUpdate`).
- **UI:** add duration + cooldown inputs to the sensor edit form (minutes-based with a helper that converts to seconds). HelpTip with a plain-language example.
- **Test:** smoke test that a reading below threshold for < duration does **not** create an alert, and for > duration does — add to `cmd/api/smoke_test.go`.

### WS3 — Alert → Task linkage

- **Schema:** `gr33ncore.tasks.source_alert_id BIGINT REFERENCES gr33ncore.alerts_notifications(id) ON DELETE SET NULL`.
- **API:** two options — pick one:
  - **A. Server-side helper** `POST /alerts/{id}/create-task` that reads the alert, synthesizes title/description/zone/priority, inserts a task, returns it. Cleaner, testable, RAG-friendly.
  - **B. Client-side** — the UI composes a TaskCreate from the alert row and POSTs to `/farms/{id}/tasks`.
  - **Recommend A**; simpler operator UX, one round trip, leaves the linkage on the server.
- **UI:** button on each alert row — "Create task". Prefilled modal; on submit, the Alerts page shows a small "→ Task #N" badge next to the alert.
- **Task page:** small "from alert" link back to the originating alert when `source_alert_id` is set.
- **Test:** smoke test that alert→task preserves farm, zone, and records `source_alert_id`.

### WS4 — Schedule preconditions (interlock lite)

- **Schema:** `gr33ncore.schedules.preconditions JSONB DEFAULT '[]'::jsonb NOT NULL`.
  - Shape: `[{ "sensor_id": 12, "op": "gte", "value": 10.0 }, ...]` (ops: `lt | lte | eq | gte | gt | ne`).
- **Worker:** before executing any schedule's `executable_actions` (control_actuator, etc.), fetch the latest reading for each sensor in `preconditions` and evaluate. If any fails:
  - `automation_runs.status = 'skipped'`
  - `automation_runs.message = 'precondition_failed'`
  - `automation_runs.details = { failed: [{sensor_id, op, expected, actual}] }`
  - No actuator commands written.
- **API:** accept/return `preconditions` on `POST /farms/{id}/schedules` and `PUT /schedules/{id}`; update `openapi.yaml` (`Schedule`, `ScheduleCreate`, `ScheduleUpdate`).
- **UI:** on the schedule edit form, "Preconditions" section — pick sensor, pick op, enter value; repeat. Empty list = no interlock, same behavior as today.
- **Test:** smoke test a schedule with a precondition that fails (latest reading below threshold) produces a skipped run and leaves the actuator untouched.

## After Phase 19

- **Unlocks Phase 20** — the automation-rule engine can re-use the precondition evaluator plumbing.
- **Workflow guide update** — add a subsection to [`docs/workflow-guide.md`](../workflow-guide.md) under "Alerts" and "Schedules & automation runs" describing duration/cooldown and interlocks.
- **OpenAPI** stays 1:1 with `cmd/api/routes.go`; re-run the diff audit at the end.

## Risks / things to watch

- **Evaluator state model** (WS2) — if we need breach-start per sensor, we either add a column on `sensors` (`alert_breach_started_at TIMESTAMPTZ NULL`) or derive from the latest N readings. Column is simpler and survives restarts.
- **Back-compat** — every new column must have a non-null default (or be nullable) so existing deployments keep working without a data migration.
- **Precondition freshness** — "latest reading" can be stale; document that preconditions are best-effort and not a substitute for a real interlock sensor on the hardware side.

---

## Using this plan in a new chat (copy-paste prompt)

```text
Implement Phase 19 per @docs/plans/phase_19_safety_and_alert_rules.plan.md.

Scope:
1) WS1 — Fix TestFarmBootstrapOnCreate + TestOrgDefaultBootstrapOnFarmCreate so `go test ./cmd/api/...` is green.
2) WS2 — Add alert_duration_seconds + alert_cooldown_seconds to sensors; extend the alert evaluator; add sensor-edit UI inputs; smoke test duration gating.
3) WS3 — Add tasks.source_alert_id + POST /alerts/{id}/create-task; add an Alerts-page button and a Tasks-page back-link; smoke test the linkage.
4) WS4 — Add schedules.preconditions JSONB; worker evaluates before executing and logs status=skipped reason=precondition_failed when any fails; UI editor on the schedule form; smoke test a failing precondition.

Constraints: keep changes backwards-compatible (nullable / default values on new columns), keep openapi.yaml 1:1 with routes.go, run `go test ./cmd/api/...` and `pytest pi_client/test_gr33n_client.py` at the end, and update this plan's YAML todo statuses when each WS lands.
```
