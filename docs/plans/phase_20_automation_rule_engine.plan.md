---
name: Phase 20 Automation Rule Engine
overview: >
  Turns gr33n from "records what happened" into "reacts to what sensors do." Builds
  out the dormant gr33ncore.automation_rules + executable_actions tables (schema
  exists, zero code today) into a working rule engine with CRUD, a worker evaluator,
  three core action types, and a UI rule builder. Target: 2–3 weeks.
todos:
  - id: ws1-rules-crud
    content: "WS1: sqlc queries + CRUD handlers for automation_rules and executable_actions; OpenAPI paths; smoke tests for both tables"
    status: completed
  - id: ws2-rule-evaluator
    content: "WS2: Worker rule evaluator — poll active rules, eval conditions_jsonb against latest readings, honor cooldown + last_evaluated/triggered_time; write automation_runs"
    status: completed
  - id: ws3-action-dispatchers
    content: "WS3: Action dispatchers for control_actuator, create_task, send_notification (defer http_webhook_call and update_record_in_gr33n)"
    status: completed
  - id: ws4-rule-builder-ui
    content: "WS4: Rule builder UI — rule list page + three-pane form (Trigger / Conditions / Actions); HelpTips; tie to existing sensors + actuators pickers"
    status: completed
  - id: ws5-docs-and-smoke
    content: "WS5: Smoke coverage for every action type; update docs/workflow-guide.md §3 to describe rules vs schedules; OpenAPI audit"
    status: completed
isProject: false
---

# Phase 20 — Automation Rule Engine

## Why this phase

Phase 19 made the existing cron scheduler safer (duration, cooldown, interlock). Phase 20 adds the second leg: **rules that fire on sensor state changes, not a clock**. The schema has been sitting dormant since the initial data model — `gr33ncore.automation_rules` + `gr33ncore.executable_actions` exist with rich columns (trigger_source, conditions_jsonb, cooldown_period_seconds, polymorphic action types) but zero handler code, zero worker code, and zero UI. This phase wires it up end-to-end, opinionated about the three most valuable action types, and leaves the rest for a later phase.

## Hand-offs from Phase 19 (reuse, don't re-implement)

Phase 19 landed four reusable pieces. Lean on them instead of forking:

- **Predicate shape** — `{ sensor_id: <int>, op: lt|lte|eq|gte|gt|ne, value: <number> }`. The WS4 precondition evaluator lives in `internal/automation/worker.go` (`evalPrecondition`, `numericToFloat64`, type `schedulePrecondition`). Promote these to a shared `internal/automation/predicates.go` in WS2 so both the schedule interlock and the rule evaluator read the same code path. `conditions_jsonb` should canonicalise to `{ "logic": "ALL"|"ANY", "predicates": [<same shape>] }`.
- **Worker test hook** — `Worker.Tick(ctx)` is already exported. The rule evaluator should get an analogous `TickRules(ctx)` (or be absorbed into the same `Tick`) so `cmd/api/smoke_test.go` can drive it deterministically with cooldown=0 the same way WS4 does.
- **Alert → Task pattern** — `POST /alerts/{id}/create-task` in `internal/handler/alert/handler.go` is the reference implementation for deriving task fields (title / description / priority / zone) from context. The `create_task` action dispatcher should mirror it, including the `source_alert_id`-style provenance column (see WS1 below).
- **Task provenance precedent** — Phase 19 added `gr33ncore.tasks.source_alert_id BIGINT REFERENCES alerts_notifications(id) ON DELETE SET NULL` with a partial index. WS1 should add a sibling column `source_rule_id BIGINT REFERENCES automation_rules(id) ON DELETE SET NULL` with the same partial-index pattern so "this task came from rule #42" is queryable.
- **Details JSON shape** — Phase 19 writes `automation_runs.details = { "phase": "<stage>", "failed": [...], ... }`. Rule runs should follow suit: `{ "phase": "conditions"|"actions", "conditions_met": true, "actions_total": N, "actions_success": M, "errors": [{action_id, message}] }`. Keep the key names stable so future RAG can join Phase 19 and Phase 20 runs with one query.

## Scope

| WS | Focus | Location in repo |
|----|--------|------------------|
| **WS1** | CRUD API for rules + actions | `db/queries/automation.sql`, `internal/db/automation.sql.go` (generated), `internal/handler/automation/rules_handler.go` (new), `cmd/api/routes.go`, `openapi.yaml` |
| **WS2** | Worker evaluator | `internal/automation/worker.go` + new `internal/automation/rules.go` |
| **WS3** | Action dispatchers | `internal/automation/actions/*.go` (new), wiring into the worker |
| **WS4** | Rule builder UI | `ui/src/views/Automation.vue` (new), `ui/src/components/RuleForm.vue` (new), router entry, SideNav entry under "Operate" |
| **WS5** | Tests + docs | `cmd/api/smoke_test.go`, `docs/workflow-guide.md`, OpenAPI audit |

## Work-stream detail

### WS1 — Rules + actions CRUD

- **Migration first** (`db/migrations/20260503_phase20_task_source_rule.sql` + mirror in `db/schema/gr33n-schema-v2-FINAL.sql`):
  - `ALTER TABLE gr33ncore.tasks ADD COLUMN source_rule_id BIGINT REFERENCES gr33ncore.automation_rules(id) ON DELETE SET NULL;`
  - Partial index `CREATE INDEX idx_tasks_source_rule_id ON gr33ncore.tasks (source_rule_id) WHERE source_rule_id IS NOT NULL;`
  - Update `Task` / `TaskCreate` schemas in `openapi.yaml` to include `source_rule_id` (mirrors the Phase 19 WS3 treatment of `source_alert_id`).
- **Routes** (add to `cmd/api/routes.go`, JWT-protected, member + operate authz as appropriate):
  - `GET /farms/{id}/automation/rules` — list
  - `POST /farms/{id}/automation/rules` — create
  - `GET /automation/rules/{id}` — get
  - `PUT /automation/rules/{id}` — update
  - `DELETE /automation/rules/{id}` — soft delete or hard delete (decide; recommend hard delete with cascade to `executable_actions`)
  - `PATCH /automation/rules/{id}/active` — enable/disable toggle (mirrors schedule active toggle)
  - Nested actions:
    - `GET /automation/rules/{id}/actions`
    - `POST /automation/rules/{id}/actions`
    - `PUT /automation/actions/{id}`
    - `DELETE /automation/actions/{id}`
- **Validation** in handlers:
  - `trigger_source` must be a valid enum value.
  - `conditions_jsonb` must parse to the shape `{ "logic": "ALL"|"ANY", "predicates": [{sensor_id, op, value}] }`; the predicate shape and op whitelist MUST match the Phase 19 WS4 precondition shape — lean on the shared validator (see Hand-offs above).
  - Every `predicates[].sensor_id` must belong to the rule's farm (same farm-scoping check used in `parsePreconditions`).
  - Actions: enforce the CHECK constraint client-side so operators get a readable error, not a 500. Explicitly return `400` with message "action_type X is defined in the schema but not yet supported" for `trigger_another_automation_rule | http_webhook_call | update_record_in_gr33n | log_custom_event`.
- **OpenAPI:** add `AutomationRule`, `AutomationRuleCreate`, `AutomationRuleUpdate`, `ExecutableAction`, `ExecutableActionCreate`, `ExecutableActionUpdate`, and `RulePredicate` (or `$ref` the Phase 19 `SchedulePrecondition`) schemas. Re-run the routes.go ↔ openapi.yaml diff.
- **Smoke:**
  - Create a rule with two conditions and one control_actuator action; toggle active; update; delete; confirm cascade removes actions.
  - Separate negative test: POSTing an action with `action_type = http_webhook_call` returns `400` with the deferred-type message (proves the guardrail bites at write-time, not tick-time).

### WS2 — Worker rule evaluator

- Runs on its own cadence (e.g. every 15s configurable) in `internal/automation/worker.go`. Prefer extending the existing `Worker` struct over spawning a second goroutine manager — one ticker, two passes (`scheduleTick` + `ruleTick`) keeps simulation-mode, cooldown options, and the test hook uniform.
- Steps per tick:
  1. `SELECT id, farm_id, trigger_source, trigger_configuration, conditions_jsonb, condition_logic, cooldown_period_seconds, last_triggered_time FROM automation_rules WHERE is_active = TRUE`.
  2. For each rule, resolve the triggering sensor(s) from `trigger_configuration`, fetch latest readings via `GetLatestReadingBySensor` (same query WS4 uses).
  3. Evaluate predicates with `condition_logic` (ALL / ANY) using the shared predicate evaluator promoted out of `worker.go`.
  4. If the rule is currently in cooldown (`now - last_triggered_time < cooldown_period_seconds`), skip and record a `skipped` run with `message = "cooldown"` (mirror the schedule cooldown row shape).
  5. If satisfied: write `automation_runs` row (`rule_id`, status, message, details), dispatch each `executable_action` in `execution_order`, update `last_triggered_time`.
  6. Always update `last_evaluated_time`.
- **Testability** — expose `Worker.TickRules(ctx)` (or fold into `Tick`) the same way Phase 19 WS4 exposed `Tick`. `cmd/api/smoke_test.go` already initialises the worker with `automationworker.WithCooldown(0)` — reuse.
- Bookkeeping: if any action dispatcher fails, run status = `partial_success` or `failed`; record which action failed in `details.errors[] = { action_id, message }`. Successful runs: `details = { "phase": "actions", "actions_total": N, "actions_success": M }`.
- Observability: the existing `GET /farms/{id}/automation/runs` endpoint already lists runs for both schedules and rules (it has `rule_id` and `schedule_id` columns); update `ui/src/views/Schedules.vue` run rendering (currently hard-codes `schedule #{{ r.schedule_id }}`) so it also shows `rule #{{ r.rule_id }}` when set. When the new `Automation.vue` page lands, link rule runs there.

### WS3 — Action dispatchers

Implement three action types; the others stay as 400 "not yet supported" from the CRUD layer so operators can't create unrunnable rules.

- **control_actuator** — mirror the existing schedule path in `internal/automation/worker.go` (`executeAction` / `case "control_actuator"`) that writes `pending_command` on the device and records the resulting actuator event when the Pi confirms. Set `triggered_by_rule_id` on the actuator event (schema column already exists). Respects `delay_before_execution_seconds`.
- **create_task** — Phase 19 WS3 has already landed; reuse the derivation logic from `internal/handler/alert/handler.go::CreateTaskFromAlert` (severity→priority, subject→title, etc.). Map `action_parameters` to title / description / zone_id / priority / due_date (relative days). Set the new `source_rule_id` column (added in WS1) for provenance.
- **send_notification** — `internal/pushnotify` is wired and exercised in existing tests; dispatch through it. `action_parameters.title` + `action_parameters.body` + target user set (default: all farm members with push tokens). Also write the rendered notification to `gr33ncore.alerts_notifications` so the Alerts page surfaces rule-driven notifications alongside threshold-driven ones.
- Everything else (`trigger_another_automation_rule`, `http_webhook_call`, `update_record_in_gr33n`, `log_custom_event`) remains valid at the DB layer but is explicitly rejected by the CRUD validator with a clear message pointing to a future phase. The worker also has a defensive `default:` branch that records `status='failed', message='unsupported action_type'` so an out-of-band DB insert can't brick a tick.

### WS4 — Rule builder UI

- **New page** `ui/src/views/Automation.vue` under "Operate" in the sidebar, next to Schedules.
- **List** shows name, trigger summary ("When sensor X EC < 1.2"), cooldown, last triggered, active toggle, action count.
- **Form** (`RuleForm.vue`) has three panes:
  - **Trigger** — source dropdown, sensor picker (if source = sensor), schedule picker (if source = schedule).
  - **Conditions** — ALL / ANY selector plus a table of `{sensor, op, value}` rows (same component used for Phase 19 WS4 preconditions — reuse).
  - **Actions** — ordered list, each with a type picker and a type-specific mini-form (actuator picker for `control_actuator`, task-field prefill for `create_task`, title/body for `send_notification`).
- **HelpTips** on every pane explaining in plain language.
- **Dry-run** button (stretch goal): call a `POST /automation/rules/{id}/dry-run` that evaluates conditions against current sensor state and returns whether it *would* fire, without dispatching actions.

### WS5 — Tests + docs

- **Smoke** per action type: rule fires → actuator gets a pending_command / task gets created (and `source_rule_id` is populated) / push gets queued. Use the existing fake FCM test doubles where available. Drive the worker with `testWorker.TickRules(ctx)` (added in WS2) exactly like the Phase 19 WS4 precondition test.
- **Cooldown** test: back-to-back `TickRules` calls respect the cooldown window — first tick fires, second tick writes a `status='skipped', message='cooldown'` run, third tick after `cooldown_period_seconds` elapses fires again.
- **ALL vs ANY** test: evaluator math matches `condition_logic`; one rule with two predicates, toggle the logic and verify the fire/skip decision flips.
- **Deferred action type** test (mirrors WS1 negative): confirms the handler returns `400` for `action_type = http_webhook_call` so operators can't create rules that silently do nothing.
- **Docs:** update `docs/workflow-guide.md` §3 (Schedules & automation runs) to describe rules as a peer of schedules, not a replacement. Add a glossary entry. Cross-link to Phase 19 WS4 for the shared predicate shape.

## After Phase 20

- **Actuator control is complete** — between Phase 19's schedule interlocks and Phase 20's rule engine, the platform can safely automate a farm end-to-end.
- **Phase 21 analytics** can now segment by "caused by schedule X" vs "caused by rule Y" when summarizing a crop cycle.
- **RAG prep** — automation_runs with both schedule_id and rule_id gives the future assistant a clean, structured "why did this happen?" substrate.

## Risks / things to watch

- **Evaluator hot path** — polling every active rule every 15s across many farms won't scale forever. Phase 20 ships the naive implementation; cache "latest reading per sensor" if it shows up as a bottleneck before Phase 22.
- **Action fan-out ordering** — `execution_order` matters; respect it. Deterministic order = deterministic debugging.
- **Silent cooldown** — operators might wonder "why didn't my rule fire?" when it's cooling down. Surface last_triggered_time + remaining cooldown on the rule card.
- **Schema drift** — the existing check constraint on `executable_actions` is strict; the handler must produce valid tuples or the insert fails opaquely. Validate before insert.

---

## Using this plan in a new chat (copy-paste prompt)

```text
Implement Phase 20 per @docs/plans/phase_20_automation_rule_engine.plan.md.

Scope:
1) WS1 — sqlc queries + CRUD handlers + OpenAPI paths for gr33ncore.automation_rules and gr33ncore.executable_actions (list/create/get/update/delete + active toggle, plus nested actions routes). Smoke tests.
2) WS2 — Worker rule evaluator in internal/automation/worker.go (new rules.go) — polls active rules, evaluates conditions_jsonb with ALL/ANY logic against latest sensor readings, respects cooldown_period_seconds, writes automation_runs rows, updates last_evaluated/triggered_time.
3) WS3 — Action dispatchers for control_actuator, create_task, send_notification only; other action types valid at DB layer but rejected by the CRUD validator with a clear message.
4) WS4 — Automation.vue + RuleForm.vue UI in ui/src/views and ui/src/components; router + SideNav entries under "Operate"; HelpTips; reuse the predicate-list component from Phase 19 WS4.
5) WS5 — Smoke per action type, cooldown test, ALL/ANY test; update docs/workflow-guide.md §3; OpenAPI audit.

Constraints: keep openapi.yaml 1:1 with routes.go; run `go test ./cmd/api/...`, `go test ./...`, `python3 -m pytest pi_client/test_gr33n_client.py -q`, and `npm run build` in `ui/` after each WS; update this plan's YAML todo statuses when each WS lands. Reuse the Phase 19 WS4 predicate evaluator (`evalPrecondition` in `internal/automation/worker.go` — promote to a shared file in WS2). Reuse the Phase 19 WS3 alert→task derivation logic for the `create_task` action dispatcher. Add `source_rule_id` on `gr33ncore.tasks` in WS1 mirroring `source_alert_id`. Explicitly defer `trigger_another_automation_rule`, `http_webhook_call`, `update_record_in_gr33n`, and `log_custom_event` to a later phase — reject at the CRUD layer with a clear 400.
```
