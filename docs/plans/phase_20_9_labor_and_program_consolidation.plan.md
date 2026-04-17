---
name: Phase 20.9 Labor Logging & Program→Actions Consolidation
overview: >
  Two small-but-high-leverage RAG prereqs that close the last structural gaps
  before Phase 21. WS A — labor logging: tasks get a `time_spent_minutes`
  column plus a `task_labor_log` join table so "cost per gram yield" becomes a
  real join including labor. WS B — consolidate gr33nfertigation.programs onto
  the same executable_actions table that rules and schedules already use, so
  "what did this automation do" is one uniform query across schedules / rules /
  programs. All additive. Target: 3–4 days.
todos:
  - id: ws1-labor-schema-and-cost
    content: "WS1: Additive migration — tasks.time_spent_minutes (nullable); new gr33ncore.task_labor_log table (task_id, user_id, started_at, ended_at, minutes, hourly_rate_snapshot, currency); auto-cost on labor log insert (category='labor_wages', links to task + crop_cycle_id via task scope)"
    status: pending
  - id: ws2-labor-ui
    content: "WS2: Task detail 'Time' section — start/stop timer + manual entry + rate picker; Tasks list shows aggregate time; 'Cost to date' card on Crop Cycle detail now includes labor breakdown"
    status: pending
  - id: ws3-program-actions-link
    content: "WS3: Additive — executable_actions.program_id BIGINT REFERENCES gr33nfertigation.programs(id) ON DELETE CASCADE; CHECK ensuring rule_id XOR schedule_id XOR program_id; migrate existing program action metadata into executable_actions rows via one backfill query"
    status: pending
  - id: ws4-program-actions-dispatch
    content: "WS4: Worker program-tick reads actions from executable_actions (not meta_data.steps); existing meta_data steps remain supported in parallel for one phase then deprecated in a banner; UI program editor uses the same action-list component as RuleForm.vue"
    status: pending
  - id: ws5-smoke-and-docs
    content: "WS5: Smoke — labor log roundtrip, auto-cost lands, program execution via executable_actions, cross-source run log uniformity; workflow-guide.md §5 (Tasks) + §4 (Fertigation programs) updated; OpenAPI audit"
    status: pending
isProject: false
---

# Phase 20.9 — Labor Logging & Program → Actions Consolidation

## Why this phase

Two independent cleanups that both pay off Phase 21 RAG. Each is small alone; bundling them is a deliberate trade-off — they share the same test harness (automation_runs + cost_transactions introspection) and the same documentation surface, so we land both in one phase rather than splitting.

**Related:** operator-facing **terminology** (JADAM vs natural farming in OpenAPI and UI) lives in [`phase_20_9b_terminology_and_copy_pass.plan.md`](phase_20_9b_terminology_and_copy_pass.plan.md).

1. **Labor logging.** `tasks` today has no `time_spent_minutes`. That means:
   - RAG cannot answer "labor hours per gram yield."
   - The auto-cost loop from Phase 20.7 covers inputs + electricity but not labor.
   - `cost_category_enum` already has `labor_wages` — it's been waiting.
   
   One nullable column on `tasks` plus one new `task_labor_log` table (for multi-person / start-stop timer use cases) closes it.

2. **Program → executable_actions consolidation.** Phase 20 WS1 put three classes of automation on three different storage shapes:
   - **Schedules**: cron + `meta_data` (freeform JSON) — the *actions* a schedule runs live inside `meta_data.steps[]`.
   - **Rules**: `automation_rules` + `executable_actions` table (structured rows) — Phase 20 WS1 did it right.
   - **Programs** (fertigation): cron-like + `meta_data` — same bad shape as schedules.
   
   RAG's "what did this automation just do" query is three different shapes depending on origin. Consolidating programs onto `executable_actions` (same pattern rules use) collapses that to one. Schedules keep `meta_data` for now — too risky to touch their action format in this phase, and they're the lowest-volume of the three. A future phase can migrate schedules if demand warrants.

Both pieces are strictly additive at the schema layer.

## Hand-offs from earlier phases (reuse, don't re-implement)

- **Phase 20.7** wired the auto-cost pipeline with idempotency keys + polymorphic `related_*` back-pointers. Labor logs use the exact same pipeline; just a new `internal/costing/autologger.go::LogLaborEntry` entry point with category = `labor_wages` and idempotency key `"labor:" || labor_log.id`.
- **Phase 20.7 crop_cycle_id on cost_transactions** — labor rows set it from the task's zone's active crop_cycle (same resolution the mixing-component path does). Cost-to-date aggregations already group by crop_cycle_id.
- **Phase 20 WS1 executable_actions** — polymorphic `rule_id | schedule_id | program_id` is the target shape. The new CHECK constraint extends the existing one; migration is additive.
- **Phase 20 WS4 RuleForm.vue** shipped an ordered action-list editor. WS4 here extracts it into `ui/src/components/ActionListEditor.vue` (if not already) and reuses it on the Programs edit page.

## Scope

| WS | Focus | Location in repo |
|----|-------|------------------|
| **WS1** | Labor schema + auto-cost | `db/migrations/2026xxxx_phase209_labor.sql`, `internal/costing/autologger.go` |
| **WS2** | Labor UI | `ui/src/views/TaskDetail.vue`, `ui/src/views/Tasks.vue`, `ui/src/views/CropCycleDetail.vue` |
| **WS3** | Program→executable_actions schema + backfill | `db/migrations/2026xxxx_phase209_program_actions.sql` |
| **WS4** | Program dispatch via executable_actions | `internal/automation/worker.go`, `ui/src/views/Programs.vue`, `ui/src/components/ActionListEditor.vue` |
| **WS5** | Smoke + docs | `cmd/api/smoke_test.go`, `docs/workflow-guide.md` §5 + §4 |

## Work-stream detail

### WS1 — Labor schema + auto-cost

```sql
-- quick-path: operator just says "I spent 45 min on this"
ALTER TABLE gr33ncore.tasks
  ADD COLUMN IF NOT EXISTS time_spent_minutes INTEGER CHECK (time_spent_minutes IS NULL OR time_spent_minutes >= 0);

-- detailed path: timer starts / stops, multiple workers per task, rate captured at log time
CREATE TABLE IF NOT EXISTS gr33ncore.task_labor_log (
  id                    BIGSERIAL PRIMARY KEY,
  farm_id               BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
  task_id               BIGINT NOT NULL REFERENCES gr33ncore.tasks(id) ON DELETE CASCADE,
  user_id               UUID NOT NULL REFERENCES gr33ncore.profiles(user_id) ON DELETE RESTRICT,
  started_at            TIMESTAMPTZ NOT NULL,
  ended_at              TIMESTAMPTZ,
  minutes               INTEGER GENERATED ALWAYS AS (
    CASE WHEN ended_at IS NULL THEN NULL
    ELSE GREATEST(0, EXTRACT(EPOCH FROM (ended_at - started_at))::INTEGER / 60)
    END
  ) STORED,
  hourly_rate_snapshot  NUMERIC(10,2),
  currency              CHAR(3) CHECK (currency IS NULL OR currency ~ '^[A-Z]{3}$'),
  notes                 TEXT,
  created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_labor_task ON gr33ncore.task_labor_log (task_id);
CREATE INDEX idx_labor_user_started ON gr33ncore.task_labor_log (user_id, started_at DESC);
```

- `hourly_rate_snapshot` is captured at log-close time — rates change, but a historic cost must not. Default source is a per-user "hourly rate" on the profile (ADD COLUMN `gr33ncore.profiles.hourly_rate` + `hourly_rate_currency` in this migration) which the timer UI pre-fills; the operator can override. Profile rate is optional — if NULL, the labor log still saves, just with no auto-cost.
- `internal/costing/autologger.go::LogLaborEntry(ctx, labor_id)` called post-commit on INSERT/UPDATE when `ended_at` transitions from NULL to a value. Writes one `cost_transactions` row: category `labor_wages`, `related_table_name='task_labor_log'`, `related_record_id=<id>`, `crop_cycle_id` resolved from task→zone→active-cycle, idempotency_key `"labor:<id>"`. Updating `ended_at` later (to extend a timer) writes a compensating row delta-style, not a rewrite — ledger stays append-only.

Routes (JWT, member authz):
- `POST /tasks/{id}/labor-log/start` → inserts with `started_at=NOW(), ended_at=NULL`, returns the row.
- `POST /tasks/{id}/labor-log/stop` → updates the open log for this (task, user) pair with `ended_at=NOW()`, captures rate.
- `POST /tasks/{id}/labor-log` (manual entry) → `{started_at, ended_at, user_id?, notes?}`.
- `GET /tasks/{id}/labor-log`, `DELETE /labor-log/{id}` (admin; reverses via compensating cost row, same pattern as consumptions).
- `PATCH /tasks/{id}` already supports `time_spent_minutes` — that's the quick-entry path that skips the detailed log table. If `time_spent_minutes` is set and there are no labor_log rows, the auto-cost uses the task's zone's most-recently-active user's profile rate OR the farm default.

### WS2 — Labor UI

- **Task detail page**: new "Time" section.
  - Big "Start timer" button → starts a labor log for the current user + open task.
  - Running timer badge in header (cross-page) — reminds you a timer is running.
  - "Stop" button + confirmation → captures rate + closes.
  - Separate "Log manual entry" form for backfilled time.
  - List of past labor log entries per user with rates + minutes + cost.
  - Single-field "Quick log (X minutes)" for the non-timer path.
- **Tasks list**: new column "Time" showing sum of logs + quick-entry `time_spent_minutes`.
- **Crop Cycle detail "Cost to date" card** (shipped in Phase 20.7 WS6) grows a "Labor" line item that breaks down time × rate across cycle tasks.
- **Profile page** gets an `hourly_rate` + currency editor, with HelpTip "used as the default when you log time on tasks; historical logs keep whatever rate was captured at close time."

### WS3 — Program → executable_actions schema

```sql
ALTER TABLE gr33ncore.executable_actions
  ADD COLUMN IF NOT EXISTS program_id BIGINT REFERENCES gr33nfertigation.programs(id) ON DELETE CASCADE;

-- tighten the "owned by exactly one parent" invariant
ALTER TABLE gr33ncore.executable_actions
  DROP CONSTRAINT IF EXISTS chk_executable_action_parent;  -- if a previous phase added one
ALTER TABLE gr33ncore.executable_actions
  ADD CONSTRAINT chk_executable_action_parent CHECK (
    (rule_id IS NOT NULL AND schedule_id IS NULL AND program_id IS NULL) OR
    (rule_id IS NULL AND schedule_id IS NOT NULL AND program_id IS NULL) OR
    (rule_id IS NULL AND schedule_id IS NULL AND program_id IS NOT NULL)
  );

CREATE INDEX IF NOT EXISTS idx_executable_actions_program ON gr33ncore.executable_actions (program_id) WHERE program_id IS NOT NULL;
```

**Backfill** — iterate `gr33nfertigation.programs` rows whose `meta_data.steps[]` describes a list of actions; for each step insert an `executable_actions` row with `program_id=<program>`. Old `meta_data.steps[]` stays in place (additive — nothing deleted) but the worker stops reading it once the backfill runs and the schema version is bumped. Run the backfill idempotently: skip programs that already have an `executable_actions` row.

**Plain-JSONB programs with nothing in `meta_data.steps`** — skip. The operator just hadn't configured actions yet; the new empty state on the UI is fine.

### WS4 — Program dispatch via executable_actions

- `internal/automation/worker.go` program-tick (same module that currently reads `meta_data.steps`) is rewritten to `SELECT * FROM executable_actions WHERE program_id = $1 ORDER BY execution_order`. All action-dispatch code paths (`control_actuator`, `create_task`, `send_notification`) are already shared — this is just re-pointing the input.
- **One-phase parallel** — during this phase, the worker reads from BOTH `executable_actions.program_id` AND `meta_data.steps`, preferring executable_actions when present. A banner on the Programs page tells operators "this program still uses legacy action storage — click here to migrate." After Phase 20.9 ships, Phase 21+ can flip a switch and stop reading `meta_data.steps`.
- **UI**: `Programs.vue` edit form replaces its current steps editor with `ActionListEditor.vue` (the shared component from RuleForm.vue). Saving persists through the existing `/automation/rules/{id}/actions` pattern, but scoped to a program — new tiny endpoints `POST /programs/{id}/actions`, `PUT /actions/{id}` (the `PUT` is shared with rules; the polymorphic parent check in the handler accepts either a rule_id or program_id owner).

### WS5 — Smoke + docs

- Smoke (labor):
  - Start timer, wait (mock `NOW()`), stop → labor_log row has computed `minutes`, one `cost_transactions` row at category `labor_wages`, idempotency holds on re-save.
  - Manual entry with custom `hourly_rate_snapshot` overriding profile default → cost matches supplied rate.
  - Quick-path `time_spent_minutes=45` on a task → labor cost logs using profile default rate.
- Smoke (programs):
  - Backfill migration against a seeded program with `meta_data.steps=[{...}, {...}]` → two `executable_actions` rows with correct `program_id` and preserved execution_order.
  - Running the program via worker tick writes `automation_runs` row with `related_program_id` (new? no — `automation_runs.schedule_id` already exists, add `program_id` only if it doesn't — check before ALTER) and the same shape of details JSON rules use.
  - `chk_executable_action_parent` CHECK fires on attempts to insert an action with two parents set.
- Docs:
  - `workflow-guide.md` §5 (Tasks) gains a "Time & labor" subsection.
  - `workflow-guide.md` §4 (Fertigation programs) rewrites the programs paragraph to describe executable_actions + reference §3 (schedules/rules) for the shared action vocabulary.
  - Glossary: `labor_log`, `hourly_rate_snapshot`.

## After Phase 20.9

- **Labor rounds out the cost story** — input + electricity + labor all auto-log. Phase 21 RAG can answer "cost per gram yield" with a single query.
- **One action vocabulary across rules, schedules (future), and programs** — RAG retrieval queries become uniform. "What did this automation do" is now `SELECT ... FROM automation_runs JOIN executable_actions ON ...` regardless of trigger source.
- **Last pre-RAG additive schema change.** After this, schema freezes until Phase 21 ships and we have real production data to evaluate against. Phase 21 is read-only on the relational schema (vector indices + RAG plumbing are additive on their own new tables).

## Risks / things to watch

- **Running-timer UX** — a forgotten timer is a common failure mode in time-tracking tools. Ship the cross-page badge + a daily summary email-or-notification in-phase, don't defer.
- **Rate snapshots on open logs** — do NOT capture rate at timer-start; only at close. An operator might have their rate raised mid-shift. Closing the log with the current rate is the defensible behavior.
- **Backfill dry-run** — the program-actions backfill touches every farm's programs. The migration MUST be read-only first (log counts, confirm shape) before writing. Consider a two-migration split: read-only audit first, write-second.
- **schedule actions left alone** — resist the temptation to also migrate schedules onto executable_actions in this phase. Too much blast radius. That's a later phase (Phase 22+) when we have a day to dedicate to it.

---

## Using this plan in a new chat (copy-paste prompt)

```text
Implement Phase 20.9 per @docs/plans/phase_20_9_labor_and_program_consolidation.plan.md.

Scope:
1) WS1 — Additive migration: tasks.time_spent_minutes, gr33ncore.task_labor_log (new), profiles.hourly_rate + hourly_rate_currency. Extend internal/costing/autologger.go with LogLaborEntry (category='labor_wages', idempotent). Routes: start/stop/manual labor log + list + delete.
2) WS2 — Task detail 'Time' section with timer + manual entry; Tasks list time column; Profile page rate editor; Crop Cycle detail Cost-to-date gets labor breakdown.
3) WS3 — Additive: executable_actions.program_id + tightened chk_executable_action_parent; idempotent backfill of gr33nfertigation.programs.meta_data.steps into executable_actions rows.
4) WS4 — Worker program-tick reads from executable_actions; one-phase parallel read (meta_data.steps + executable_actions, preferring actions); Programs.vue uses the shared ActionListEditor.vue.
5) WS5 — Smoke (labor roundtrip + idempotency, backfill, CHECK violation, uniform action vocabulary); workflow-guide §5 + §4 updates; OpenAPI audit.

Constraints: additive schema only (nullable columns, new tables, new CHECK tightening that old rows already satisfy). NO touching of schedules.meta_data.steps — deliberately out of scope. Reuse Phase 20.7 autologger + idempotency keys + related_* back-pointers. Run go test ./cmd/api/..., go test ./..., python3 -m pytest pi_client/test_gr33n_client.py -q, and npm run build in ui/ after each WS. Update this plan's YAML todo statuses when each WS lands.
```
