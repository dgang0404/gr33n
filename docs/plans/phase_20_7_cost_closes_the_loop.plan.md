---
name: Phase 20.7 Cost Closes the Loop
overview: >
  The cost ledger (gr33ncore.cost_transactions) is fully built but nothing in
  the platform writes to it automatically. Telemetry knows exactly what was
  consumed (mixing_event_components, fertigation_events, actuator_events) and
  the ledger has a polymorphic back-pointer (related_module_schema /
  related_table_name / related_record_id) built for this — it's just been
  waiting for the telemetry tables to land. They have. This phase wires the
  auto-deduct + auto-cost + low-stock-alert loops end-to-end, so every nutrient
  mix, every hour a grow light runs, and every feed consumption shows up in
  Cost and Inventory without an operator remembering to log it. Target:
  1.5–2 weeks (biggest of the pre-RAG phases, by design).
todos:
  - id: ws1-schema-additions
    content: "WS1: Additive migrations — input_definitions.unit_cost + unit_cost_currency + unit_cost_unit_id; input_batches.low_stock_threshold; actuators.watts; new gr33ncore.farm_energy_prices table; new gr33ncore.task_input_consumptions join table; cost_transactions.crop_cycle_id; broaden input_category_enum with animal_feed/bedding/veterinary_supply"
    status: pending
  - id: ws2-auto-deduct-and-cost-on-mixing
    content: "WS2: Post-commit auto-deduct on mixing_event_components insert — decrement input_batches.current_quantity_remaining and insert one cost_transaction per component (idempotent via cost_transaction_idempotency)"
    status: pending
  - id: ws3-task-consumption-api
    content: "WS3: task_input_consumptions CRUD — POST /tasks/{id}/consumptions, GET list, DELETE reverses the deduct+cost pair; UI on Task detail"
    status: pending
  - id: ws4-electricity-rollup
    content: "WS4: Nightly worker job — for each actuator with watts>0, compute on-duration from actuator_events across the day, multiply by current farm_energy_prices row, insert one cost_transaction per (actuator, date); idempotent; worker health surface mentions last run"
    status: pending
  - id: ws5-low-stock-rules
    content: "WS5: Low-stock alert — scheduled rule type (or inventory-specific worker pass) that fires when any input_batch.current_quantity_remaining < low_stock_threshold; alert → task via Phase 19 WS3 'create task from alert' flow"
    status: pending
  - id: ws6-ui-cost-surface
    content: "WS6: Costs page gets 'auto-logged' filter + source breadcrumb chip (chip reads cost_transactions.related_* to link back to the mixing event / actuator / task); Inventory page shows current_quantity_remaining + low_stock_threshold editor; Crop cycle detail shows 'cost to date' aggregate from cost_transactions.crop_cycle_id"
    status: pending
  - id: ws7-smoke-and-docs
    content: "WS7: Smoke per loop (mixing consumption, task consumption, electricity rollup, low-stock alert); docs/workflow-guide.md §7 (Costs) expanded; new cost-attribution playbook; OpenAPI audit"
    status: pending
isProject: false
---

# Phase 20.7 — Cost Closes the Loop

## Why this phase

Every question a farm operator actually wants answered is a cost question:

- *"What did this crop cycle cost me?"*
- *"Which zone is most expensive to run?"*
- *"How much did electricity cost during mid_flower last month?"*
- *"Am I about to run out of Nute X?"*
- *"What's my cost per gram of yield?"*

The schema already expects this: `cost_transactions` has a polymorphic back-pointer (`related_module_schema / related_table_name / related_record_id`) that has no other reason to exist, `input_batches.current_quantity_remaining` is a live stock counter that never gets decremented today, `cost_transaction_idempotency` exists precisely so offline/retry writes are safe, and `cost_category_enum` already includes `utilities_electricity_gas`, `fertilizers_soil_amendments`, `feed_livestock`, `water_irrigation`. The tables are waiting.

What's missing is the connective tissue — the worker passes and post-commit hooks that turn telemetry into ledger entries. This phase ships them. It is the single most valuable pre-RAG phase because every RAG money question becomes answerable once it lands, and remains unanswerable without it no matter how clever the retrieval is.

## Hand-offs from earlier phases (reuse, don't re-implement)

- **Polymorphic cost back-pointer** — `cost_transactions.related_module_schema / related_table_name / related_record_id` is already in place. Every auto-logged row MUST fill all three so the UI can render a "source" chip that navigates back to the mixing event / actuator / task / reservoir that created the cost.
- **Idempotency table** — `gr33ncore.cost_transaction_idempotency (farm_id, idempotency_key) UNIQUE` is built for this. All auto-writers must supply a deterministic key: mixing components use `"mixing_component:" || id`; electricity rollups use `"electricity:" || actuator_id || ":" || date`; task consumptions use `"task_consumption:" || id`. Retries + network splits are now safe.
- **Alert → Task flow** — Phase 19 WS3 landed `POST /alerts/{id}/create-task` with derivation from severity/subject/zone. Low-stock alerts use it unchanged — the alert fires with severity=`medium`, subject "Inventory: {input_name} below threshold", and operators can one-click it into a refill task.
- **Scheduled worker** — Phase 19 WS2/WS4 + Phase 20 WS2 already have a scheduled tick loop with `WithCooldown` and a test hook. The electricity rollup + low-stock check both register as schedule types (or a single new `inventory_and_cost_rollup` schedule type) so they benefit from existing worker health surface.
- **Farm-scoped currency** — `cost_transactions.currency` already exists as `CHAR(3) CHECK currency ~ '^[A-Z]{3}$'`. `input_definitions.unit_cost_currency` copies the same constraint. No new currency infrastructure needed.

## Scope

| WS | Focus | Location in repo |
|----|-------|------------------|
| **WS1** | Additive migrations | `db/migrations/2026xxxx_phase207_cost_loop.sql` + schema mirror; regenerated sqlc |
| **WS2** | Mixing → deduct + cost | `internal/handler/fertigation/mixing.go`, `internal/costing/autologger.go` (new) |
| **WS3** | Task consumption API + UI | `internal/handler/task/consumption.go` (new), `ui/src/views/TaskDetail.vue` (new or extended) |
| **WS4** | Electricity rollup worker | `internal/automation/worker.go` + `internal/automation/electricity_rollup.go` (new) |
| **WS5** | Low-stock alerts | `internal/automation/worker.go` + `internal/automation/lowstock.go` (new) |
| **WS6** | UI surface polish | `ui/src/views/Costs.vue`, `ui/src/views/Inventory.vue`, `ui/src/views/CropCycles.vue` |
| **WS7** | Smoke + docs | `cmd/api/smoke_test.go`, `docs/workflow-guide.md` §7, new playbook |

## Work-stream detail

### WS1 — Additive migrations

One migration file, all additive:

```sql
-- input cost
ALTER TABLE gr33nnaturalfarming.input_definitions
  ADD COLUMN IF NOT EXISTS unit_cost          NUMERIC(12,4),
  ADD COLUMN IF NOT EXISTS unit_cost_currency CHAR(3) CHECK (unit_cost_currency IS NULL OR unit_cost_currency ~ '^[A-Z]{3}$'),
  ADD COLUMN IF NOT EXISTS unit_cost_unit_id  BIGINT REFERENCES gr33ncore.units(id) ON DELETE RESTRICT;

-- low stock threshold (nullable — opt-in per batch)
ALTER TABLE gr33nnaturalfarming.input_batches
  ADD COLUMN IF NOT EXISTS low_stock_threshold NUMERIC(10,2);

-- watts per actuator (nullable — opt-in)
ALTER TABLE gr33ncore.actuators
  ADD COLUMN IF NOT EXISTS watts NUMERIC(8,2);

-- farm-level energy price, time-effective (no ALTERs to historical rows; insert-only)
CREATE TABLE IF NOT EXISTS gr33ncore.farm_energy_prices (
  id             BIGSERIAL PRIMARY KEY,
  farm_id        BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
  effective_from TIMESTAMPTZ NOT NULL,
  price_per_kwh  NUMERIC(10,6) NOT NULL CHECK (price_per_kwh >= 0),
  currency       CHAR(3) NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_farm_energy_prices_farm_effective ON gr33ncore.farm_energy_prices (farm_id, effective_from DESC);

-- task-driven consumption join table (manual feed / manual fertigation / anything)
CREATE TABLE IF NOT EXISTS gr33ncore.task_input_consumptions (
  id              BIGSERIAL PRIMARY KEY,
  task_id         BIGINT NOT NULL REFERENCES gr33ncore.tasks(id) ON DELETE CASCADE,
  input_batch_id  BIGINT NOT NULL REFERENCES gr33nnaturalfarming.input_batches(id) ON DELETE RESTRICT,
  quantity        NUMERIC(10,3) NOT NULL CHECK (quantity > 0),
  unit_id         BIGINT NOT NULL REFERENCES gr33ncore.units(id) ON DELETE RESTRICT,
  notes           TEXT,
  recorded_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  recorded_by     UUID REFERENCES gr33ncore.profiles(user_id) ON DELETE SET NULL
);
CREATE INDEX idx_task_consumptions_task ON gr33ncore.task_input_consumptions (task_id);
CREATE INDEX idx_task_consumptions_batch ON gr33ncore.task_input_consumptions (input_batch_id);

-- crop-cycle attribution on the ledger
ALTER TABLE gr33ncore.cost_transactions
  ADD COLUMN IF NOT EXISTS crop_cycle_id BIGINT REFERENCES gr33nfertigation.crop_cycles(id) ON DELETE SET NULL;
CREATE INDEX idx_cost_transactions_crop_cycle ON gr33ncore.cost_transactions (crop_cycle_id) WHERE crop_cycle_id IS NOT NULL;

-- widen input categories for animal husbandry (additive ALTER TYPE)
ALTER TYPE gr33nnaturalfarming.input_category_enum ADD VALUE IF NOT EXISTS 'animal_feed';
ALTER TYPE gr33nnaturalfarming.input_category_enum ADD VALUE IF NOT EXISTS 'bedding';
ALTER TYPE gr33nnaturalfarming.input_category_enum ADD VALUE IF NOT EXISTS 'veterinary_supply';
```

All additive. Nothing existing changes shape. Enum ADDs are safe.

### WS2 — Auto-deduct + auto-cost on mixing

- New package `internal/costing/autologger.go` with one entry point per source:
  - `LogMixingComponent(ctx, mixingComponentID)` — called post-commit in `internal/handler/fertigation/mixing.go` after inserting each `mixing_event_components` row.
  - It resolves `input_batch_id → input_batches → input_definitions.unit_cost`, computes `cost = volume_added * unit_cost`, decrements `current_quantity_remaining`, and writes one `cost_transactions` row with `category='fertilizers_soil_amendments'`, `related_table_name='mixing_event_components'`, `related_record_id=<id>`, `idempotency_key='mixing_component:<id>'`, and `crop_cycle_id` resolved from the fertigation program / zone's active cycle.
- **If `unit_cost` is NULL** → the deduct still happens (stock tracking still works) but no cost row is written. Operators can fill in unit_cost later; a separate backfill endpoint (`POST /cost/backfill?from=...&to=...`) retroactively logs costs for rows that now have a price. Deferred to Phase 21 — out of scope here.
- **Negative-stock guard** — decrementing below zero is a red flag, not a hard error. Log a WARNING and insert a `system_logs` row, but don't block the operation. Operators physically can run a batch to empty without gr33n knowing the exact endpoint.

### WS3 — Task consumption API + UI

- Routes:
  - `GET /tasks/{id}/consumptions`
  - `POST /tasks/{id}/consumptions` → body `{input_batch_id, quantity, unit_id, notes?}`; post-commit calls `autologger.LogTaskConsumption(ctx, id)`.
  - `DELETE /consumptions/{id}` → reverses: re-credit the batch, mark the cost_transaction as voided (keep the audit row — new column NOT needed; use `description` prefix `[VOIDED]` and subtract via a compensating transaction so net = 0 and the ledger stays append-only).
- Task detail page grows a "Consumed" section with an add-row form (input picker → batch picker → quantity + unit). Same pattern as the existing mixing-event components editor.

### WS4 — Electricity rollup worker

- New file `internal/automation/electricity_rollup.go` + registration as a built-in schedule type (or a direct new worker pass — decide in WS4, prefer schedule type for uniformity with the existing worker health surface).
- Logic (runs daily, default 01:00 farm time but operator-tunable):
  1. For each farm:
     1. Resolve the active `farm_energy_prices` row (greatest `effective_from <= now()`).
     2. For each actuator with `watts IS NOT NULL AND watts > 0`:
        - Pull yesterday's `actuator_events` for that actuator.
        - Reconstruct on/off intervals (ignore degenerate ON→ON or OFF→OFF pairs; they indicate noise, log but don't fail).
        - Sum on-duration in seconds; kWh = `watts * seconds / 3_600_000`; cost = `kWh * price_per_kwh`.
        - Insert one `cost_transactions` row (category `utilities_electricity_gas`, related_table `actuators`, related_record `actuator_id`, idempotency_key `electricity:<actuator_id>:<YYYY-MM-DD>`, `description` = `"Electricity: {actuator_name} ran {hh:mm} × {watts}W @ {price}/kWh"`).
  2. If no `farm_energy_prices` row exists for the farm, skip silently. An inline tooltip on the Costs page says "configure a $/kWh to enable automatic electricity cost logging."
- **Testability** — expose `Worker.TickElectricityRollup(ctx, date)` the same way Phase 20 WS2 exposed `TickRules`, so the smoke test can force a rollup for a specific date with seeded actuator_events.

### WS5 — Low-stock alerts

- New worker pass (or periodic rule, see WS4 decision): every 15 minutes by default, `SELECT ... FROM input_batches WHERE low_stock_threshold IS NOT NULL AND current_quantity_remaining < low_stock_threshold AND deleted_at IS NULL`.
- For each match: call the existing alert pipeline (`alerts_notifications` insert + push fan-out — the same one Phase 20 WS3 wired for `send_notification` actions). Subject: "Inventory low: {input_name} at {remaining} / {threshold} {unit}". Severity: `medium`. Idempotent per `(batch_id, current_day)` so operators don't get spammed every 15 minutes; use a helper table or check-before-insert against latest alert for this batch.
- **Alert → Task one-click** already works (Phase 19 WS3). No code needed there.
- **Auto-clear** — when a batch is refilled (operator edits `current_quantity_remaining` upward, or a new batch for the same input_definition is received — future inventory inbound feature), the corresponding open alert transitions to `status='system_cleared'`. This is an additive bookkeeping pass in the same worker.

### WS6 — UI surface polish

- **Costs page** grows two things:
  - A filter chip "Auto-logged" (shows only rows where `related_module_schema IS NOT NULL`).
  - A source chip on each row: clicking it navigates to the mixing event / actuator detail / task detail that produced the cost. This is the "audit trail" RAG will leverage; operators love it too.
- **Inventory page** (or wherever `input_batches` are surfaced): show `current_quantity_remaining / quantity_produced`, an inline `low_stock_threshold` editor, and a visible "warn below" badge.
- **Crop Cycle detail page**: a small "Cost to date" card that aggregates `cost_transactions` where `crop_cycle_id = this`. Breakdown by category. This is the first RAG-precursor visualization.
- **Farm settings page** grows an "Energy price" section — list of `farm_energy_prices` rows with add/edit (effective_from is immutable once set; operators add a new row to change the rate).

### WS7 — Smoke + docs

- Smoke:
  - Mixing event insert → exactly one cost_transactions row + `current_quantity_remaining` decremented; second insert with same key is a no-op (idempotency).
  - Task consumption POST then DELETE → batch re-credited, net cost = 0 (via compensating row), ledger append-only.
  - Electricity rollup with seeded actuator_events → one cost row per actuator per day; running rollup twice for the same day is a no-op.
  - Low-stock alert fires exactly once per day per batch even when the worker runs 96 times that day.
  - `NULL unit_cost` → stock decrements, no cost row written, no error.
- Docs:
  - Update `docs/workflow-guide.md` §7 (Costs & finance) with a "Automatic cost attribution" subsection.
  - New playbook `docs/cost-attribution-playbook.md` — how to set `watts` on actuators, how to set `unit_cost` on inputs, how to configure `farm_energy_prices`, how to read the auto-logged chip.
  - Glossary: `low_stock_threshold`, `farm_energy_prices`, `task_input_consumption`.

## After Phase 20.7

- Every question listed in "Why this phase" becomes answerable by a single SQL query — and therefore by Phase 21 RAG without any clever indexing.
- Inventory turns from a passive ledger into an active safety net (low-stock → alert → task → refill).
- Operators can compare crop cycles on true cost-per-gram, the first real business-ops lens the platform offers.

## Risks / things to watch

- **Idempotency keys are sacred** — any auto-writer that produces non-deterministic keys will double-bill. Every WS2/WS4/WS5 smoke test must include a second-invocation check.
- **Currency consistency** — a farm might stock ingredients priced in EUR while logging costs in USD. WS1 adds `unit_cost_currency` per input; the auto-logger must fail loudly (`system_logs` + no write) if it would mix currencies in one transaction. Defer per-row FX conversion to a later phase — out of scope.
- **Electricity rollup timezones** — "yesterday" is ambiguous for multi-region farms. Use the farm's timezone (`gr33ncore.farms.timezone`) if present, else UTC. Document clearly.
- **Schema additivity** — the temptation to add a proper `inventory_movements` event-sourced table will be strong. Resist for this phase. The direct-decrement-on-batch model is enough for pre-RAG; a ledger-style movements table is a Phase 22+ refactor.

---

## Using this plan in a new chat (copy-paste prompt)

```text
Implement Phase 20.7 per @docs/plans/phase_20_7_cost_closes_the_loop.plan.md.

Scope:
1) WS1 — Additive migrations + schema mirror: input_definitions.unit_cost / unit_cost_currency / unit_cost_unit_id, input_batches.low_stock_threshold, actuators.watts, gr33ncore.farm_energy_prices (new), gr33ncore.task_input_consumptions (new), cost_transactions.crop_cycle_id, broadened input_category_enum.
2) WS2 — New internal/costing/autologger.go, wired post-commit into mixing_event_components insert; decrements stock + writes cost_transaction (idempotent).
3) WS3 — task_input_consumptions CRUD + autologger hook + Task detail UI.
4) WS4 — Electricity rollup worker pass in internal/automation/electricity_rollup.go, testable via Worker.TickElectricityRollup(ctx, date).
5) WS5 — Low-stock alert worker pass in internal/automation/lowstock.go; per-batch-per-day idempotency; integrates with the existing alert → task pipeline.
6) WS6 — Costs/Inventory/CropCycle UI polish: auto-logged filter chip, source breadcrumbs, low-stock editor, Cost to Date card, Energy Price editor on farm settings.
7) WS7 — Smoke per loop + second-invocation idempotency checks; workflow-guide.md §7; new cost-attribution playbook; OpenAPI audit.

Constraints: additive schema only — no changes to existing tables' shape beyond new nullable columns + new tables + additive enum ADD VALUE. All auto-writes MUST use cost_transaction_idempotency deterministic keys. Run go test ./cmd/api/..., go test ./..., python3 -m pytest pi_client/test_gr33n_client.py -q, and npm run build in ui/ after each WS. Update this plan's YAML todo statuses when each WS lands.
```
