---
name: Phase 20.95 RAG-Prep Column Adds & Housekeeping
overview: >
  A tight, one-sprint hardening pass before Phase 21. Adds the additive schema
  slivers that Phase 21 analytics and future RAG joins will want — labor
  logging, unit costs, energy pricing, program→actions linkage, animal + loop
  scope columns — without building the UI / worker plumbing that the full
  20.6–20.9 phases will eventually ship. Cost is cheap now (column adds with
  sane defaults); cost later is a backfill + re-surface on every cost and
  cycle page. Also cleans up three things that will only get worse with time:
  splits the 3,475-line cmd/api/smoke_test.go into feature files, adds a
  scripts/openapi_route_diff.sh CI guard, and sweeps the workflow guide for
  terminology-guideline links. All changes additive. Target: 2–3 days.
todos:
  - id: ws1-labor-schema
    content: "WS1: Labor logging schema (subset of 20.9 WS1) — tasks.time_spent_minutes + new gr33ncore.task_labor_log table + minimal CRUD (list/create/delete, no timer UI, no auto-cost); smoke test roundtrip"
    status: completed
  - id: ws2-cost-energy-columns
    content: "WS2: Cost/energy column adds (subset of 20.7 WS1) — input_definitions.{unit_cost, unit_cost_currency, unit_cost_unit_id}; input_batches.low_stock_threshold; actuators.watts; new gr33ncore.farm_energy_prices table; cost_transactions.crop_cycle_id; broaden input_category_enum with animal_feed/bedding/veterinary_supply; OpenAPI Sensor/Actuator/Input schemas reflect the new fields"
    status: completed
  - id: ws3-program-actions-link
    content: "WS3: executable_actions.program_id + exactly-one-source CHECK (20.9 WS3) — migration, sqlc regen, CRUD validator reject; one-shot backfill from programs.meta_data.steps if any exist (idempotent, safe if empty); smoke test that creating an action with program_id=42 round-trips and the CHECK rejects two-source rows"
    status: completed
  - id: ws4-animals-loops-columns
    content: "WS4: Animal + aquaponics scope columns (subset of 20.8 WS1) — animal_groups.(count, primary_zone_id, active, archived_at, archived_reason); aquaponics.loops.(fish_tank_zone_id, grow_bed_zone_id); all nullable/defaulted; OpenAPI schemas updated; smoke test column round-trip"
    status: completed
  - id: ws5-split-smoke-tests
    content: "WS5: Split cmd/api/smoke_test.go (3,475 lines, 60+ tests) into 12 feature files — smoke_auth_test.go, smoke_farms_test.go (includes Organization* tests), smoke_sensors_test.go, smoke_alerts_test.go, smoke_tasks_test.go, smoke_automation_test.go, smoke_fertigation_test.go, smoke_costs_test.go, smoke_inventory_test.go, smoke_plants_test.go, smoke_crop_cycles_test.go, smoke_commons_test.go; shared helpers in smoke_helpers_test.go; TestMain stays in smoke_test.go. Zero behavior change."
    status: completed
  - id: ws6-openapi-route-diff
    content: "WS6: scripts/openapi_route_diff.sh — extracts paths from cmd/api/routes.go (rg) and operations from openapi.yaml (yq), prints a unified diff, exits 1 if they disagree. Add a Makefile target `make audit-openapi`; wire into the existing CI step."
    status: completed
  - id: ws7-workflow-guide-sweep
    content: "WS7: Workflow-guide smoke — read §3 (schedules & automation) and §4 (fertigation); ensure docs/terminology-guideline.md is linked where JADAM / natural farming are first mentioned; add glossary entries for 'JADAM' (proper-noun method) and 'natural farming' (generic English umbrella) worded to stand alone with NO national/regional/ethnic qualifiers — not 'Korean', 'Japanese', 'Indian', etc.; final grep `rg -in 'korean|\\bknf\\b|japanese natural|indian natural|asian natural' docs/workflow-guide.md` must return zero hits"
    status: completed
isProject: false
---

# Phase 20.95 — RAG-Prep Column Adds & Housekeeping

## Why this phase

Phase 20 is done. Phase 21 is analytics. Between them sits a set of **additive schema slivers** that are drafted in the 20.6 / 20.7 / 20.8 / 20.9 plans but haven't landed yet. Each one is *cheap now* — a column add with a safe default — and *painful later* once Phase 21 starts rendering reports against the missing data:

- Report "cost per gram" without labor? Half-truth. Adding the column **after** the page exists means a backfill prompt on every task.
- Report "$ per cycle" when `cost_transactions.crop_cycle_id` doesn't exist yet? Every historical row is untaggable forever.
- Unify run logs across schedules + rules + programs? If we wait, we have **two** flavors of action rows to maintain.

This phase does **not** build the UI / worker / auto-cost plumbing that the full 20.6–20.9 phases will eventually ship. It lands the column adds, the tables, and the minimum-viable CRUD that keeps the shape stable so Phase 21 reports + future RAG joins can be written against a frozen surface.

Bolted on to the same sprint: three housekeeping tasks that only get more expensive with time — split the monolithic smoke test file, add an OpenAPI drift script, and a 15-minute workflow-guide sweep.

## Hand-offs (reuse, don't re-implement)

- **Migration shape** — mirror every new migration into `db/schema/gr33n-schema-v2-FINAL.sql` the same turn, same PR (Phase 19 + 20 precedent). `IF NOT EXISTS` on columns, `IF NOT EXISTS` on tables — reruns must be safe.
- **sqlc regeneration** — `internal/db/*.sql.go` is generated; always re-run `sqlc generate` after `db/queries/` changes.
- **OpenAPI ↔ routes discipline** — new routes touch both `cmd/api/routes.go` and `openapi.yaml`. WS6 lands the script that will bite next time we forget.
- **Smoke pattern** — follow the existing "create → list → update → delete" roundtrip pattern used by `TestNfInputDefinitionCRUD` etc. One smoke test per new column set is enough at this phase; the full flows come with 20.6–20.9.

## Scope

| WS | Focus | Locations in repo |
|----|--------|------------------|
| **WS1** | Labor schema (column + table + minimal CRUD) | `db/migrations/2026xxxx_phase2095_labor_schema.sql`, schema mirror, `db/queries/tasks.sql`, `internal/handler/task/handler.go`, `cmd/api/routes.go`, `openapi.yaml`, smoke |
| **WS2** | Cost / energy columns + enum | `db/migrations/2026xxxx_phase2095_cost_energy_columns.sql`, schema mirror, `db/queries/{inventory,costs,automation}.sql` (minimal), `openapi.yaml` schemas only (no new endpoints) |
| **WS3** | program_id on executable_actions | `db/migrations/2026xxxx_phase2095_exec_actions_program_id.sql`, schema mirror, `db/queries/automation.sql` (ExecAction CRUD validator), handler guard, smoke |
| **WS4** | animals + loops scope columns | `db/migrations/2026xxxx_phase2095_animals_loops_columns.sql`, schema mirror, `db/queries/{animals,aquaponics}.sql` (if the files exist — otherwise OpenAPI schema only), smoke |
| **WS5** | Split smoke test (12 feature files) | `cmd/api/smoke_*_test.go` (new), `cmd/api/smoke_helpers_test.go` (new), `cmd/api/smoke_test.go` (keeps TestMain + setup) |
| **WS6** | Route ↔ OpenAPI drift script | `scripts/openapi_route_diff.sh` (new), `Makefile` target |
| **WS7** | Workflow-guide sweep | `docs/workflow-guide.md` §3 + §4 + glossary |

## Work-stream detail

### WS1 — Labor schema (subset of 20.9 WS1)

**Migration** `db/migrations/2026xxxx_phase2095_labor_schema.sql` (+ schema mirror):

```sql
ALTER TABLE gr33ncore.tasks
  ADD COLUMN IF NOT EXISTS time_spent_minutes INTEGER;

CREATE TABLE IF NOT EXISTS gr33ncore.task_labor_log (
  id                    BIGSERIAL PRIMARY KEY,
  farm_id               BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
  task_id               BIGINT NOT NULL REFERENCES gr33ncore.tasks(id) ON DELETE CASCADE,
  user_id               UUID   REFERENCES gr33ncore.profiles(user_id) ON DELETE SET NULL,
  started_at            TIMESTAMPTZ NOT NULL,
  ended_at              TIMESTAMPTZ,
  minutes               INTEGER NOT NULL CHECK (minutes >= 0),
  hourly_rate_snapshot  NUMERIC(10,2),
  currency              CHAR(3) CHECK (currency ~ '^[A-Z]{3}$'),
  notes                 TEXT,
  created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_task_labor_log_task ON gr33ncore.task_labor_log (task_id);
CREATE INDEX IF NOT EXISTS idx_task_labor_log_farm ON gr33ncore.task_labor_log (farm_id);
CREATE TRIGGER trg_task_labor_log_updated_at
  BEFORE UPDATE ON gr33ncore.task_labor_log
  FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();
```

**Routes** (JWT-protected, member authz):

- `GET  /tasks/{id}/labor` — list rows for one task.
- `POST /tasks/{id}/labor` — insert (shape: `started_at`, `ended_at?`, `minutes`, `hourly_rate_snapshot?`, `currency?`, `notes?`).
- `DELETE /labor/{id}` — hard delete.

Explicit **non-goals** for this phase:

- No timer UI (20.9 full WS2).
- No auto-cost transaction on insert (20.9 full WS1 second half; this phase just stores the minutes and the rate snapshot so nothing is lost).
- No `time_spent_minutes` aggregation trigger — the POST/DELETE handlers write `tasks.time_spent_minutes` as a **running SUM** over all surviving log rows for the task:

  ```sql
  UPDATE gr33ncore.tasks
  SET time_spent_minutes = COALESCE((
    SELECT SUM(minutes) FROM gr33ncore.task_labor_log WHERE task_id = $1
  ), 0)
  WHERE id = $1;
  ```

  Running SUM is the only semantic that doesn't silently lie after a DELETE of a non-latest log row, and it's what Phase 21's "total time on task" report will assume. Document the semantic in a `COMMENT ON COLUMN gr33ncore.tasks.time_spent_minutes IS 'denormalised SUM(task_labor_log.minutes) maintained by handler'`.

**OpenAPI:** `TaskLaborLog`, `TaskLaborLogCreate` schemas; three new paths.

**Smoke:** create task → POST two labor rows (30m, 45m) → GET list returns two → assert `tasks.time_spent_minutes == 75` → DELETE the 30m row → GET returns one → assert `tasks.time_spent_minutes == 45`.

### WS2 — Cost / energy columns (subset of 20.7 WS1)

**Migration** `db/migrations/2026xxxx_phase2095_cost_energy_columns.sql` (+ schema mirror):

```sql
-- input cost metadata
ALTER TABLE gr33nnaturalfarming.input_definitions
  ADD COLUMN IF NOT EXISTS unit_cost          NUMERIC(12,4),
  ADD COLUMN IF NOT EXISTS unit_cost_currency CHAR(3) CHECK (unit_cost_currency IS NULL OR unit_cost_currency ~ '^[A-Z]{3}$'),
  ADD COLUMN IF NOT EXISTS unit_cost_unit_id  BIGINT REFERENCES gr33ncore.units(id) ON DELETE SET NULL;

-- low-stock trigger
ALTER TABLE gr33nnaturalfarming.input_batches
  ADD COLUMN IF NOT EXISTS low_stock_threshold NUMERIC(12,4);

-- actuator wattage for the nightly electricity rollup (20.7 WS4, later)
ALTER TABLE gr33ncore.actuators
  ADD COLUMN IF NOT EXISTS watts NUMERIC(10,2) DEFAULT 0 NOT NULL;

-- farm-level energy pricing (additive new table)
CREATE TABLE IF NOT EXISTS gr33ncore.farm_energy_prices (
  id               BIGSERIAL PRIMARY KEY,
  farm_id          BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
  effective_from   DATE NOT NULL,
  effective_to     DATE,
  price_per_kwh    NUMERIC(10,4) NOT NULL CHECK (price_per_kwh >= 0),
  currency         CHAR(3) NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
  notes            TEXT,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_farm_energy_prices_active
  ON gr33ncore.farm_energy_prices (farm_id, effective_from DESC);
CREATE TRIGGER trg_farm_energy_prices_updated_at
  BEFORE UPDATE ON gr33ncore.farm_energy_prices
  FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();

-- cycle-scoped cost tagging
ALTER TABLE gr33ncore.cost_transactions
  ADD COLUMN IF NOT EXISTS crop_cycle_id BIGINT REFERENCES gr33nfertigation.crop_cycles(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_cost_tx_crop_cycle
  ON gr33ncore.cost_transactions (crop_cycle_id)
  WHERE crop_cycle_id IS NOT NULL;

-- broaden input category so animal feed etc. cost correctly (20.8)
ALTER TYPE gr33nnaturalfarming.input_category_enum ADD VALUE IF NOT EXISTS 'animal_feed';
ALTER TYPE gr33nnaturalfarming.input_category_enum ADD VALUE IF NOT EXISTS 'bedding';
ALTER TYPE gr33nnaturalfarming.input_category_enum ADD VALUE IF NOT EXISTS 'veterinary_supply';
```

**OpenAPI:** extend `Sensor` no (not touched), `Actuator`/`ActuatorCreate` (+ `watts`), `InputDefinition`/`InputDefinitionCreate` (+ three unit_cost fields), `InputBatch` (+ `low_stock_threshold`), `CostTransaction`/`CostTransactionCreate` (+ `crop_cycle_id`), new schemas `FarmEnergyPrice`/`FarmEnergyPriceCreate`.

**Routes:** this phase intentionally adds **only** the energy-price routes (the rest are column-adds on existing endpoints):

- `GET  /farms/{id}/energy-prices` — list.
- `POST /farms/{id}/energy-prices` — create.
- `PUT  /energy-prices/{id}` — update (incl. closing `effective_to`).
- `DELETE /energy-prices/{id}` — hard delete.

**Smoke:** column round-trip on an existing input_definition / actuator / input_batch / cost_transaction (PUT the new field, GET it back). CRUD on energy_prices. Enum round-trip: create an input_definition with `category='animal_feed'` and confirm it persists.

### WS3 — `executable_actions.program_id` + exactly-one source (subset of 20.9 WS3)

**Migration** `db/migrations/2026xxxx_phase2095_exec_actions_program_id.sql` (+ schema mirror):

```sql
ALTER TABLE gr33ncore.executable_actions
  ADD COLUMN IF NOT EXISTS program_id BIGINT REFERENCES gr33nfertigation.programs(id) ON DELETE CASCADE;

-- Pre-check: the OLD CHECK was at-least-one (schedule_id OR rule_id), not
-- exactly-one. Any existing row with BOTH set would fail the new constraint
-- and brick the migration mid-deploy. Fail fast and loud instead, pointing
-- at the problem, so the operator fixes data before rerunning.
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM gr33ncore.executable_actions
    WHERE num_nonnulls(schedule_id, rule_id) <> 1
  ) THEN
    RAISE EXCEPTION
      'executable_actions has % rows violating exactly-one-source; fix before migrating',
      (SELECT COUNT(*) FROM gr33ncore.executable_actions
       WHERE num_nonnulls(schedule_id, rule_id) <> 1);
  END IF;
END $$;

-- Now safe: drop the at-least-one CHECK and replace with an exactly-one CHECK
-- that also accepts program_id. Existing rows all satisfy it (program_id is
-- NULL and exactly one of schedule_id/rule_id is set, verified above).
ALTER TABLE gr33ncore.executable_actions
  DROP CONSTRAINT IF EXISTS chk_executable_source;
ALTER TABLE gr33ncore.executable_actions
  ADD  CONSTRAINT chk_executable_source
  CHECK (
    num_nonnulls(schedule_id, rule_id, program_id) = 1
  );

CREATE INDEX IF NOT EXISTS idx_exec_actions_program
  ON gr33ncore.executable_actions (program_id)
  WHERE program_id IS NOT NULL;
```

**Optional backfill** (idempotent, no-op when no rows have steps yet — keep it in the migration so *if* any deployment has steps they land unified):

```sql
INSERT INTO gr33ncore.executable_actions (program_id, action_type, execution_order, action_command, action_parameters)
SELECT p.id,
       'control_actuator'::gr33ncore.executable_action_type_enum,
       COALESCE((step->>'order')::int, 0),
       step->>'command',
       step - 'order' - 'command'
FROM gr33nfertigation.programs p
CROSS JOIN LATERAL jsonb_array_elements(COALESCE(p.metadata->'steps', '[]'::jsonb)) AS step
WHERE NOT EXISTS (
  SELECT 1 FROM gr33ncore.executable_actions ea WHERE ea.program_id = p.id
)
AND step ? 'command';
```

**CRUD validator** — the only handler that inserts into `executable_actions` today is `internal/handler/automation/rules_handler.go` (confirmed: `h.q.CreateExecutableActionForRule` at ~line 617; schedule-side inserts go through the worker, not a handler). Add a guard in that handler: reject a POST body where `num_nonnulls(schedule_id, rule_id, program_id) != 1` with a 400 before DB round-trip. The DB CHECK is belt-and-suspenders. Also extend the sqlc query + generated Params struct to take `program_id` (sqlc regen after updating `db/queries/automation.sql`).

**Worker:** no change this phase. 20.9 full WS4 moves program execution onto the new rows; 20.95 just lets them be written and read.

**Smoke:** insert an executable_action with `program_id` set → round-trip through `GET`. Negative: insert with both `program_id` and `schedule_id` → 400.

### WS4 — Animal + loop scope columns (subset of 20.8 WS1)

**Migration** `db/migrations/2026xxxx_phase2095_animals_loops_columns.sql` (+ schema mirror):

```sql
ALTER TABLE gr33nanimals.animal_groups
  ADD COLUMN IF NOT EXISTS count             INTEGER DEFAULT 0 NOT NULL CHECK (count >= 0),
  ADD COLUMN IF NOT EXISTS primary_zone_id   BIGINT REFERENCES gr33ncore.zones(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS active            BOOLEAN DEFAULT TRUE NOT NULL,
  ADD COLUMN IF NOT EXISTS archived_at       TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS archived_reason   TEXT;
CREATE INDEX IF NOT EXISTS idx_animal_groups_primary_zone
  ON gr33nanimals.animal_groups (primary_zone_id)
  WHERE primary_zone_id IS NOT NULL;

ALTER TABLE gr33naquaponics.loops
  ADD COLUMN IF NOT EXISTS fish_tank_zone_id BIGINT REFERENCES gr33ncore.zones(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS grow_bed_zone_id  BIGINT REFERENCES gr33ncore.zones(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_loops_fish_tank_zone
  ON gr33naquaponics.loops (fish_tank_zone_id)
  WHERE fish_tank_zone_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_loops_grow_bed_zone
  ON gr33naquaponics.loops (grow_bed_zone_id)
  WHERE grow_bed_zone_id IS NOT NULL;
```

**OpenAPI:** extend `AnimalGroup`/`AnimalGroupCreate` and `AquaponicsLoop`/`AquaponicsLoopCreate` schemas with the new fields. No new routes — existing CRUD accepts them.

**Smoke:** existing animal_groups CRUD test (`cmd/api/smoke_test.go` — find via `grep "animal_groups"`; add one if missing) round-trips the new columns. Same for loops.

### WS5 — Split cmd/api/smoke_test.go

`cmd/api/smoke_test.go` is **3,475 lines** with **60+** `func TestXxx` tests. Keep:

- `smoke_test.go` — `TestMain`, package-level init, any `TestHealth` / `TestAuthModeEndpoint` flavour. ~400 lines.
- `smoke_helpers_test.go` — shared helpers (`login`, `createFarm`, `seedSensor`, the body-builder functions used in 10+ places). **Move, don't duplicate.**

Split the rest along feature lines (roughly in order of appearance today):

- `smoke_auth_test.go` — Login, JWT, profile, notification prefs, push tokens.
- `smoke_farms_test.go` — GetFarm, CrossFarmWriteForbidden, OrganizationCreateListUsageAndFarmLink, CoaMappings, bootstrap templates (`TestFarmBootstrapOnCreate`, `TestOrgDefaultBootstrapOnFarmCreate`, `TestPhase205BootstrapTemplates`).
- `smoke_sensors_test.go` — ListSensors, SensorReadingsAndStats, SensorAlertDurationAndCooldown.
- `smoke_alerts_test.go` — AlertLifecycle, AlertToTaskLinkage.
- `smoke_tasks_test.go` — TaskCreate, TaskUpdateAndDelete.
- `smoke_automation_test.go` — ListAutomationRuns, WorkerHealth, ScheduleActiveToggle, SchedulePreconditionFailsRun, ScheduleCreateUpdateDelete, and **all** 9 `TestAutomationRule*` tests.
- `smoke_fertigation_test.go` — FertigationReservoirRoundtrip + UpdateDelete, EcTargetRoundtrip, ProgramRoundtrip + UpdateDelete, EventRoundtripWithCropCycle, MixingEventCreateWithComponents.
- `smoke_costs_test.go` — CostsSummaryListExport, CostReceiptUpload/Replacement/Delete.
- `smoke_inventory_test.go` — NfInputDefinitionCRUD, NfBatchCRUD, RecipeList, RecipeFullCRUD.
- `smoke_plants_test.go` — PlantCRUD.
- `smoke_crop_cycles_test.go` — CropCycleCreateAndStage, CropCycleFullCRUD.
- `smoke_commons_test.go` — CommonsCatalogBrowseAndImport, InsertCommonsPreview.

**Rules for the split:**

- All files in `package main` (same package — `TestMain` lives in only one).
- No test body changes. Zero behaviour diff.
- Run `go test ./cmd/api/... -count=1` before + after and diff the test count — must match exactly.
- **One-time** post-split audit: `go test ./cmd/api/... -count=1 -shuffle=on`. Expect some flakes — the existing smoke tests share helper state (login tokens, seeded farms) that will misbehave under shuffle for reasons unrelated to the split. Use the audit to catch *obvious* hidden order deps (e.g. a test that relies on a DB row created two tests earlier), not to gate the PR. **Do NOT wire `-shuffle=on` into CI** — it'll flake forever.
- Update the README test-count line if one exists.

### WS6 — `scripts/openapi_route_diff.sh`

Single bash file using `rg` + `python3`, exits 1 on drift. **Router note:** `cmd/api/routes.go` uses Go 1.22+ `http.ServeMux` with `mux.HandleFunc("METHOD /path", ...)` and `mux.Handle("METHOD /path", <wrapper>)` — **not** chi. The regex must match both forms.

```bash
#!/usr/bin/env bash
set -euo pipefail

ROUTES_FILE=/tmp/gr33n_routes.txt
OPENAPI_FILE=/tmp/gr33n_openapi.txt

# Extract (METHOD PATH) pairs from cmd/api/routes.go. Matches:
#   mux.HandleFunc("GET /health", ...)
#   mux.Handle("POST /sensors/{id}/readings", requireAPIKey(...))
rg -oN '\bmux\.(Handle|HandleFunc)\("(GET|POST|PUT|PATCH|DELETE) ([^"]+)"' cmd/api/routes.go \
  | sed -E 's/.*"(GET|POST|PUT|PATCH|DELETE) ([^"]+)".*/\1 \2/' \
  | sort -u > "$ROUTES_FILE"

# Paranoia: if the extractor returned zero routes, the regex is broken, not
# the openapi file. Fail loudly instead of silently passing against an empty
# set.
if ! [ -s "$ROUTES_FILE" ]; then
  echo "❌ route extractor returned 0 routes from cmd/api/routes.go — regex is broken" >&2
  exit 2
fi

# Extract (METHOD PATH) pairs from openapi.yaml.
python3 - <<'PY' > "$OPENAPI_FILE"
import yaml
spec = yaml.safe_load(open("openapi.yaml"))
rows = []
for path, item in (spec.get("paths") or {}).items():
    for verb in ("get","post","put","patch","delete"):
        if verb in (item or {}):
            rows.append(f"{verb.upper()} {path}")
print("\n".join(sorted(set(rows))))
PY

if diff -u "$ROUTES_FILE" "$OPENAPI_FILE"; then
  echo "openapi.yaml is 1:1 with routes.go ✓"
else
  echo "❌ drift detected — update openapi.yaml or routes.go" >&2
  exit 1
fi
```

**Makefile target:**

```makefile
.PHONY: audit-openapi
audit-openapi:
	@scripts/openapi_route_diff.sh
```

**CI wiring:** add a step to whatever GitHub Action runs `go test` today (grep `.github/workflows/` — if CI isn't wired yet, skip; the local `make audit-openapi` is still valuable). Document in README under the test-running section.

### WS7 — Workflow-guide sweep

A `rg -i 'korean natural farming|\bknf\b' docs/` at plan-authoring time returns hits only in `docs/plans/` and `docs/terminology-guideline.md` itself — `docs/workflow-guide.md` is already clean — so this WS collapses to ~10 minutes of glossary + link work on `docs/workflow-guide.md`.

**Hard rule for this WS (and the project going forward):** no national, regional, or ethnic qualifier on "natural farming" anywhere in user-facing copy — not "Korean", not "Japanese", not "Indian", not "Asian". The two accepted terms are:

- **JADAM** — a proper-noun method name. Use it when referring to the specific method, its named starter cultures (JMS, JLF, FFJ, WCA), or seed data that cites it.
- **natural farming** — lowercase generic English umbrella for fermented inputs, soil drenches, and recipes. Preferred in module titles, navigation, API tag descriptions, and any copy that should read well for international operators.

Tasks:

- §3 (Schedules & automation) — first mention of "rules" gets a parenthetical link to `docs/terminology-guideline.md` if any JADAM/natural-farming phrasing is nearby. Unlikely in §3 but check.
- §4 (Fertigation) — confirm the one or two mentions of natural-farming inputs link to `docs/terminology-guideline.md`.
- Glossary / appendix: add two entries, worded to stand alone without national framing:
  - **JADAM** — "A documented method using named starter cultures (JMS, JLF, FFJ, WCA). Use as a proper noun. See `docs/terminology-guideline.md`."
  - **Natural farming** — "Generic English label this app uses for fermented inputs, soil drenches, and related recipes. Preferred term in UI and docs. See `docs/terminology-guideline.md`."
- Re-run `rg -in 'korean|\bknf\b|japanese natural|indian natural|asian natural' docs/workflow-guide.md` at the end as a belt-and-suspenders check — zero hits required.

## Smoke + test pass (all WS)

After each WS:

```bash
go test ./... -count=1
python3 -m pytest pi_client/test_gr33n_client.py -q
npm --prefix ui run build
make audit-openapi    # WS6 onward
```

WS5 adds one extra check: `go test -count=1 ./cmd/api/...` before and after the split — test count must match exactly.

## After Phase 20.95

- **Phase 21** can now report `cost per gram including labor`, `$ per cycle`, and `kWh per cycle` against columns that have existed since day one of the report — no backfill prompts.
- **Future RAG** has one unified `executable_actions` table covering schedules + rules + programs — a single join explains "what did this automation do."
- **OpenAPI drift** is caught at CI time, not during a customer support ticket.
- **Smoke test file** is navigable — adding new tests in Phase 21 doesn't need a `rg` to find the right spot.

## Risks / things to watch

- **Exactly-one-source CHECK** (WS3) — the OLD `chk_executable_source` is at-least-one (`schedule_id IS NOT NULL OR rule_id IS NOT NULL`), not exactly-one. Any pre-existing row with BOTH set would fail the new `num_nonnulls(...) = 1` CHECK. **Mitigation baked into the migration**: a `DO $$ ... RAISE EXCEPTION` pre-check runs *before* `DROP CONSTRAINT`, fails loudly with a row count, and leaves production untouched if data needs cleanup. The operator then fixes rows and reruns.
- **Test split ordering** (WS5) — some smoke tests share state via package-level vars (login tokens, seeded farm ids). The split itself preserves file-order execution (Go runs tests in the order they're declared per file, files alphabetically), so behaviour is identical. A *one-time* `go test -shuffle=on` audit after the split is a useful debugging aid, but not a PR gate — existing order dependencies will flake it and that's fine.
- **Enum additions + seed in same transaction** — `ALTER TYPE ... ADD VALUE` is a transaction boundary in some Postgres versions. *Not triggered in 20.95* because nothing in this phase seeds `animal_feed` / `bedding` / `veterinary_supply`; the WS2 smoke test runs against a fresh DB well after the migration commits. Flagging for Phase 20.8 / 21: if either seeds these values, the seed must live in a **separate** migration file than the `ADD VALUE`.

---

## Using this plan in a new chat (copy-paste prompt)

```text
Implement Phase 20.95 per @docs/plans/phase_20_95_rag_prep_and_housekeeping.plan.md.

Scope (additive schema only + housekeeping; NO new UI / worker plumbing — that belongs to the full 20.6–20.9 phases):

1) WS1 — Labor schema subset: tasks.time_spent_minutes + gr33ncore.task_labor_log + GET/POST /tasks/{id}/labor + DELETE /labor/{id}. Smoke roundtrip. Defer auto-cost and timer UI.

2) WS2 — Cost/energy column adds: input_definitions.(unit_cost, unit_cost_currency, unit_cost_unit_id); input_batches.low_stock_threshold; actuators.watts DEFAULT 0; new gr33ncore.farm_energy_prices table + four CRUD routes; cost_transactions.crop_cycle_id; broaden input_category_enum with animal_feed / bedding / veterinary_supply. OpenAPI schemas updated for all touched models.

3) WS3 — executable_actions.program_id BIGINT REFERENCES gr33nfertigation.programs(id) ON DELETE CASCADE; migration includes a DO-block pre-check that RAISE EXCEPTIONs if any existing row violates num_nonnulls(schedule_id, rule_id) = 1 (the OLD CHECK was at-least-one, not exactly-one); then replace chk_executable_source with num_nonnulls(schedule_id, rule_id, program_id) = 1; optional idempotent backfill from programs.metadata->'steps'; CRUD validator in internal/handler/automation/rules_handler.go rejects two-source writes with 400. Smoke.

4) WS4 — animal_groups.(count, primary_zone_id, active, archived_at, archived_reason); aquaponics.loops.(fish_tank_zone_id, grow_bed_zone_id); all nullable/defaulted. OpenAPI schemas updated. Smoke column round-trip.

5) WS5 — Split cmd/api/smoke_test.go (3,475 lines, 60+ tests) into 12 files: smoke_{auth,farms,sensors,alerts,tasks,automation,fertigation,costs,inventory,plants,crop_cycles,commons}_test.go + smoke_helpers_test.go. Organization* tests fold into smoke_farms_test.go. TestMain stays in smoke_test.go. Zero behaviour change. Run `go test -count=1 ./cmd/api/...` before + after to confirm test count matches. One-time audit with `-shuffle=on` is OK; do NOT wire shuffle into CI.

6) WS6 — scripts/openapi_route_diff.sh + `make audit-openapi` target; wire into CI if .github/workflows/ has a test step. The router is Go 1.22+ `http.ServeMux` (mux.HandleFunc / mux.Handle with "METHOD /path" patterns), NOT chi — the regex must match `mux\.(Handle|HandleFunc)\("(GET|POST|PUT|PATCH|DELETE) [^"]+"`. Script must exit 2 if the extractor returns zero routes (paranoia guard against a future regex breakage silently reporting green).

7) WS7 — docs/workflow-guide.md §3 + §4 sweep; add JADAM / natural-farming glossary entries worded to stand alone with NO national/regional/ethnic qualifiers (not "Korean", "Japanese", "Indian", "Asian" — "natural farming" is the generic English umbrella, "JADAM" is the proper-noun method name); link to docs/terminology-guideline.md where relevant. Final grep `rg -in 'korean|\bknf\b|japanese natural|indian natural|asian natural' docs/workflow-guide.md` must return zero hits.

Constraints: NO changes to existing columns (only ADD COLUMN IF NOT EXISTS); every new column has a sane NULL / 0 / TRUE default so existing deployments need no backfill. Every migration mirrors into db/schema/gr33n-schema-v2-FINAL.sql the same commit. Keep openapi.yaml 1:1 with cmd/api/routes.go. Run `go test ./...`, `python3 -m pytest pi_client/test_gr33n_client.py -q`, `npm --prefix ui run build`, and `make audit-openapi` after each WS. Update this plan's YAML todo statuses when each WS lands.
```
