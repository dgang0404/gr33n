# Cost-attribution playbook

This playbook is the operator + developer reference for **how a row ends
up in `gr33ncore.cost_transactions`** and how to debug, extend, or
back-fill that ledger. It complements §7 of the [Workflow
Guide](workflow-guide.md), which describes the user-facing flows; this
file goes one level deeper into invariants, idempotency keys, and
common failure modes.

---

## 1. Why a "playbook"?

Costs are append-only. A botched insert ships out the door as a real
financial discrepancy. Phase 20.7 introduced three **autologged**
sources, each with its own deterministic key, plus the existing manual
path. When something looks wrong on the Costs page, the answer is
almost always one of:

1. The autologger replayed and (correctly) wrote nothing.
2. An auto row landed against the wrong `crop_cycle_id` because the
   triggering record didn't carry one yet.
3. The electricity rollup found no active `farm_energy_prices` row and
   silently skipped the day.
4. A manual cost was entered with a stale or missing `crop_cycle_id`
   tag, so it never shows up in `GET /crop-cycles/{id}/cost-summary`.

The rest of this doc is a structured tour of each.

---

## 2. The four cost-row sources

Every row in `cost_transactions` came from exactly one of these:

| Source                            | `related_module_schema` | `related_table_name`        | `related_record_id`       | Idempotency key                          |
| --------------------------------- | ----------------------- | --------------------------- | ------------------------- | ---------------------------------------- |
| **Manual** (operator)             | `NULL`                  | `NULL`                      | `NULL`                    | none — operator owns the dedupe          |
| **Mixing component autologger**   | `gr33nfertigation`      | `mixing_event_components`   | the component id          | `mixing_component:<id>`                  |
| **Task-consumption autologger**   | `gr33ncore`             | `task_input_consumptions`   | the consumption id        | `task_consumption:<id>` (and `task_consumption_void:<id>` for the compensating row) |
| **Electricity rollup worker**     | `gr33ncore`             | `actuators`                 | the actuator id           | `electricity:<actuator_id>:<YYYY-MM-DD>` |

The first column is what powers the **`auto · <table>`** chip on the
Costs page. The "Auto-logged only" filter is just
`WHERE related_module_schema IS NOT NULL`.

---

## 3. Idempotency contract

All three autologger paths consult `gr33ncore.cost_transaction_idempotency`
**before** writing. That table is `(farm_id, idempotency_key) UNIQUE`,
so a parallel writer that races to the same key gets a duplicate-key
error and the autologger treats it as "already logged".

This means it is **always safe** to:

- Replay a `POST /tasks/{id}/consumptions` body from a queued client.
- Re-run `TickElectricityRollup(ctx, sameDate)` after a worker crash.
- Resubmit a `POST /farms/{id}/fertigation/mixing-events` with the
  same component shape (new component IDs are created though, so a
  resubmit via the API path produces a *new* cost row — autologger
  idempotency only protects internal replays of the same row).

The smoke tests (`cmd/api/smoke_phase207_test.go`) assert "second
invocation writes zero rows" for each path. If a future change
breaks this, that file fails loud.

---

## 4. Stamping `crop_cycle_id` on auto rows (the RAG precursor)

The first user-visible payoff of all this plumbing is
`GET /crop-cycles/{id}/cost-summary`, which buckets every cost row by
`(category, currency)` for one cycle. For that view to be useful, auto
rows need to carry the right `crop_cycle_id`:

- **Mixing components** — the mixing event itself doesn't always know
  which cycle it's for (one mix can fertigate multiple cycles via
  later fertigation events). Today the autologger writes the cost row
  with `crop_cycle_id = NULL`. The cycle attribution happens later
  via the fertigation event link.
- **Task consumptions** — if the parent task carries a `crop_cycle_id`
  (Phase 20.95 added the column on `tasks`), the autologger should
  copy it onto the cost row. *Open follow-up:* this copy is not yet
  wired; see issue queue.
- **Electricity rollup** — actuators are zone-scoped, not cycle-scoped.
  We deliberately leave `crop_cycle_id = NULL`; cycle-level energy
  attribution would require a "what cycle was active in this zone on
  this day" join, which is a Phase 20.8+ topic.

**Manual rows** still take whatever `crop_cycle_id` the operator
supplies in the `POST /farms/{id}/costs` body.

---

## 5. Debug recipes

### 5a. "Why didn't the electricity rollup write anything?"

In order, check:

1. **Is there an active price?**
   `SELECT * FROM gr33ncore.farm_energy_prices WHERE farm_id = $1
    AND effective_from <= '2026-04-15' AND (effective_to IS NULL OR effective_to > '2026-04-15');`
   If empty → fix in the Costs page energy editor, then re-run the
   tick. The rollup is idempotent so it's safe.
2. **Does the actuator have `watts > 0`?**
   `ListBillableActuatorsByFarm` filters on `watts > 0`. If the
   actuator was seeded with the default `0`, set it via SQL or the
   actuator handler (Phase 20.95 added the column).
3. **Did any ON/OFF events land that day?**
   `SELECT event_time, command_sent FROM gr33ncore.actuator_events
    WHERE actuator_id = $1 AND event_time >= '2026-04-15'::date AND event_time < '2026-04-16'::date;`
   The rollup also looks at `GetLastActuatorEventBefore` to see if the
   actuator was ON at the start of the window.
4. **Was an idempotency row already written?**
   `SELECT * FROM gr33ncore.cost_transaction_idempotency
    WHERE idempotency_key = 'electricity:<actuator>:2026-04-15';`
   If yes the rollup already ran and (correctly) skipped.

### 5b. "The mixing event ran but no stock was deducted."

The autologger only deducts when the component carries
`input_batch_id`. If the operator didn't attach one, the cost row
still gets written (priced via the input definition) but no stock
moves. Check `mixing_event_components.input_batch_id` for the row in
question.

### 5c. "I deleted a task consumption and the ledger looks weird."

This is by design. `DELETE /consumptions/{id}` does **not** delete the
original `cost_transactions` row — it appends a new
`[VOIDED]`-prefixed row with a negative amount, idempotent on
`task_consumption_void:<id>`. Net cost = 0, ledger stays append-only.
Both rows show up in the Costs view; the void's `description` carries
the link back to the original.

---

## 6. Backfill / batch corrections

Two safe knobs exist for retroactive fixes:

1. **`UpdateCostTransactionCropCycle`** — re-tag a specific row's
   `crop_cycle_id` (manual SQL; no API surface today).
2. **Worker re-run** — `TickElectricityRollup(ctx, anyPastDate)` is
   idempotent and safe to call repeatedly during a backfill window.

Avoid editing `cost_transactions.amount` in place. If the price was
wrong, write a compensating row (positive or negative) so the ledger
audit trail tells the truth.

---

## 7. Where to extend next

- **Per-cycle electricity attribution** — needs zone-active-cycle
  resolution per day; lands when Phase 20.8 wires animal husbandry
  cost flows because the same join applies to barns/pens.
- **Crop-cycle copy on task consumption** — small autologger change,
  blocked on `tasks.crop_cycle_id` becoming populated more
  consistently from the UI.
- **RAG over the ledger** — once enough rows carry `crop_cycle_id`,
  feeding the per-cycle summary into a recommendation prompt is the
  natural next step (Phase 21+).
