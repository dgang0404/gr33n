---
name: Phase 20.6 Stage-Scoped Setpoints
overview: >
  Introduces one additive table — `gr33ncore.zone_setpoints` — so the "ideal
  environment" for a zone or crop cycle at a given growth stage is first-class
  data, not freeform meta_data. Teaches the rule engine to optionally read
  setpoints (so one rule "dew_point out of ideal" auto-adjusts as stages
  advance), and gives Phase 21 RAG a clean "intent vs reality" axis to reason
  over. No existing tables change. Target: 4–5 days.
todos:
  - id: ws1-migration-and-queries
    content: "WS1: Migration for gr33ncore.zone_setpoints (additive only); mirror in schema file; sqlc queries; Gr33ncoreZoneSetpoint model"
    status: completed
  - id: ws2-crud-handlers
    content: "WS2: CRUD handlers — GET|POST /farms/{id}/setpoints, GET|PUT|DELETE /setpoints/{id}; validate scope (zone_id XOR crop_cycle_id), sensor_type not empty, numeric coherence (min <= ideal <= max); OpenAPI paths + schemas"
    status: completed
  - id: ws3-rule-engine-hook
    content: "WS3: Optional setpoint-driven predicate shape in rules — `{ setpoint_key: 'dew_point', scope: 'current_stage' }` resolves at eval time to the zone's active setpoint row, falls back to a hard-coded predicate if no setpoint exists; promote to shared predicates.go"
    status: completed
  - id: ws4-ui-setpoints-page
    content: "WS4: Setpoints page under Operate — list per zone/cycle, inline edit, HelpTips explaining stage matching; link from Zone detail + Crop Cycle detail; extend RuleForm predicate picker with a 'use setpoint' toggle"
    status: completed
  - id: ws5-smoke-and-docs
    content: "WS5: Smoke — setpoint CRUD, rule fires against current_stage setpoint, falls back when no row exists; update workflow-guide.md §4 (Fertigation / stages) to describe setpoints; OpenAPI audit"
    status: completed
isProject: false
---

# Phase 20.6 — Stage-Scoped Setpoints

## Why this phase

`gr33nfertigation.crop_cycles.current_stage` is an enum (`seedling`, `early_veg`, …, `dry_cure`) but there's nowhere to record *"for strain X in `mid_flower`, the ideal dew point is 50–55°F."* Today an operator who wants stage-aware automation has to either (a) hand-edit rule thresholds every time a cycle advances, or (b) maintain parallel rules and toggle them — both are fragile and invisible to RAG.

One small additive table fixes it, and the ergonomics ripple outward immediately:

1. Operators get a single **"what should this zone look like right now"** page.
2. Rules can say "dew_point out of ideal" *once*, and the evaluator resolves what "ideal" means for the current stage on every tick.
3. Phase 21 RAG can finally answer **"intent vs reality"** questions — "you wanted dew point 50–55°F in mid_flower; actual averaged 62°F for 4 hours; yield came in at 82% of target."

The new table is strictly additive. No existing tables change. The rule engine's new "setpoint-driven predicate" is backwards-compatible — existing hard-coded predicates keep working exactly as they do today.

## Hand-offs from earlier phases (reuse, don't re-implement)

- **Predicate plumbing** — Phase 20 WS2 promoted `evalPrecondition` to the shared `internal/automation/predicates.go`. This phase extends that file with a second predicate variant — **don't fork**, add a type tag (`"hard"` vs `"setpoint"`) to the existing struct and keep one evaluator entry point.
- **Growth stage vocabulary** — `gr33nfertigation.growth_stage_enum` already exists (`clone`, `seedling`, `early_veg`, …, `dry_cure`). The new `zone_setpoints.stage` column is TEXT matching those labels (TEXT not the enum, so the same table can carry setpoints for non-crop zones too — e.g. drying rooms that don't have a fertigation crop cycle).
- **Derived sensors** — Phase 20.5 WS1 ships dew_point / VPD / heat_index as derived channels. Setpoints key off `sensor_type` (TEXT) so they work identically for physical and derived sensors.
- **Rule builder UI** — Phase 20 WS4 landed `RuleForm.vue` with a predicate editor. WS4 of this phase adds a checkbox "use setpoint from zone/cycle" that collapses the `{min,max,ideal}` inputs and instead picks a `setpoint_key` from a dropdown.

## Scope

| WS | Focus | Location in repo |
|----|-------|------------------|
| **WS1** | Additive migration + sqlc | `db/migrations/2026xxxx_phase206_zone_setpoints.sql`, `db/schema/gr33n-schema-v2-FINAL.sql`, `db/queries/setpoints.sql`, regenerated `internal/db/...` |
| **WS2** | CRUD handlers + OpenAPI | `internal/handler/setpoint/handler.go` (new), `cmd/api/routes.go`, `openapi.yaml` |
| **WS3** | Rule-engine hook | `internal/automation/predicates.go`, `internal/automation/rules.go` |
| **WS4** | UI | `ui/src/views/Setpoints.vue` (new), `ui/src/components/SetpointRow.vue` (new), router + SideNav, extension to `RuleForm.vue` |
| **WS5** | Smoke + docs | `cmd/api/smoke_test.go`, `docs/workflow-guide.md` §4 |

## Work-stream detail

### WS1 — Migration + sqlc

Migration file `db/migrations/2026xxxx_phase206_zone_setpoints.sql` (+ identical block inserted into `db/schema/gr33n-schema-v2-FINAL.sql` in the same PR):

```sql
CREATE TABLE IF NOT EXISTS gr33ncore.zone_setpoints (
  id              BIGSERIAL PRIMARY KEY,
  farm_id         BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
  zone_id         BIGINT REFERENCES gr33ncore.zones(id) ON DELETE CASCADE,
  crop_cycle_id   BIGINT REFERENCES gr33nfertigation.crop_cycles(id) ON DELETE CASCADE,
  stage           TEXT,
  sensor_type     TEXT NOT NULL,
  min_value       NUMERIC,
  max_value       NUMERIC,
  ideal_value     NUMERIC,
  meta            JSONB NOT NULL DEFAULT '{}',
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT chk_setpoint_scope CHECK (zone_id IS NOT NULL OR crop_cycle_id IS NOT NULL),
  CONSTRAINT chk_setpoint_numeric_coherent CHECK (
    (min_value IS NULL OR max_value IS NULL OR min_value <= max_value) AND
    (ideal_value IS NULL OR min_value IS NULL OR ideal_value >= min_value) AND
    (ideal_value IS NULL OR max_value IS NULL OR ideal_value <= max_value)
  )
);
CREATE INDEX idx_zone_setpoints_zone_stage ON gr33ncore.zone_setpoints (zone_id, stage, sensor_type);
CREATE INDEX idx_zone_setpoints_cycle_stage ON gr33ncore.zone_setpoints (crop_cycle_id, stage, sensor_type);
CREATE TRIGGER trg_zone_setpoints_updated_at
  BEFORE UPDATE ON gr33ncore.zone_setpoints
  FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();
```

- `zone_id` OR `crop_cycle_id` is required (via `chk_setpoint_scope`). A zone-scoped row applies to any cycle running in that zone; a cycle-scoped row overrides. Resolution order at eval time: cycle > zone > default (no setpoint).
- `stage` nullable: `NULL` stage means "all stages for this scope" (the fallback).
- sqlc queries: `ListSetpointsByFarm`, `ListSetpointsByZone`, `ListSetpointsByCropCycle`, `GetSetpointByID`, `CreateSetpoint`, `UpdateSetpoint`, `DeleteSetpoint`, **plus** the evaluator helper `GetActiveSetpointForScope(zone_id, crop_cycle_id, sensor_type, stage)` that runs the precedence resolution server-side.

### WS2 — CRUD handlers + OpenAPI

- Routes (JWT-protected, member authz):
  - `GET /farms/{id}/setpoints` (optional `?zone_id=` / `?crop_cycle_id=` / `?sensor_type=` filters)
  - `POST /farms/{id}/setpoints`
  - `GET /setpoints/{id}`, `PUT /setpoints/{id}`, `DELETE /setpoints/{id}`
- Validation mirrors the CHECK constraints client-side so operators get readable errors, not a 500. Additionally enforce that any `zone_id` / `crop_cycle_id` supplied belongs to the farm in the path (same pattern as schedule preconditions in Phase 19 WS4).
- OpenAPI: `Setpoint`, `SetpointCreate`, `SetpointUpdate` schemas; document precedence in the schema description.

### WS3 — Rule-engine hook

- Extend the predicate shape (in `internal/automation/predicates.go`) with a type discriminator:
  ```json
  { "type": "hard", "sensor_id": 42, "op": "lt", "value": 1.2 }
  { "type": "setpoint", "sensor_type": "dew_point", "scope": "current_stage", "op": "out_of_range" }
  ```
  `type` defaults to `"hard"` when absent (backwards compatible — existing rules keep working without migration).
- `"setpoint"` predicates resolve at eval time:
  1. Find the active `crop_cycle_id` for the rule's zone (via `crop_cycles.is_active = true AND zone_id = ...`).
  2. Call `GetActiveSetpointForScope` with the cycle's `current_stage`.
  3. If no setpoint row exists → the predicate resolves to "inconclusive" and the rule is skipped with `message="no_setpoint_for_scope"`. This is *not* a failure — it's a normal state when setpoints haven't been configured yet.
  4. `op="out_of_range"` → `reading < min OR reading > max`. `op="below_ideal"` / `above_ideal` / `inside_range` supported analogously.
- The evaluator reuses the existing cooldown + details-JSON bookkeeping from Phase 20 WS2.

### WS4 — UI

- **New page** `ui/src/views/Setpoints.vue` under Operate → Setpoints. Lists setpoints grouped by zone → cycle → stage. Inline row editing (sensor_type autocomplete from the zone's sensors, min/max/ideal NumberInput, stage dropdown sourced from `growth_stage_enum`).
- **New component** `SetpointRow.vue` — the actual editor, reused on the Zone detail page and the Crop Cycle detail page (two extra entry points operators will actually use in practice).
- **RuleForm.vue extension** — predicate editor gets a "Use setpoint from zone/cycle" checkbox. When checked, `sensor_id` input collapses and is replaced with `sensor_type` (autocomplete) + `scope` (dropdown `current_stage | zone_default`) + `op` (dropdown `out_of_range | below_ideal | above_ideal | inside_range`).
- **HelpTips** explain the precedence order plainly: "Cycle setpoints override zone setpoints. If no setpoint is configured, the rule skips with message `no_setpoint_for_scope` — configure one and it'll start firing."

### WS5 — Smoke + docs

- Smoke tests:
  - CRUD roundtrip (list/create/get/update/delete) with scope validation.
  - `chk_setpoint_scope` CHECK violation — creating a setpoint with both zone_id and crop_cycle_id NULL → 400.
  - Rule-engine precedence — cycle setpoint shadows zone setpoint when both exist.
  - Graceful skip — setpoint-predicate rule against a zone with no setpoint → run status=`skipped`, `message=no_setpoint_for_scope`, no actions fired.
- Docs:
  - Update `docs/workflow-guide.md` §4 (Fertigation) with a new subsection "Stage-scoped setpoints" and add a glossary entry.
  - Cross-link from `docs/pattern-playbooks.md` (created in 20.5) — "for the drying_room_v1 pattern, here's how to configure stage setpoints for cannabis flower vs dry/cure."

## After Phase 20.6

- Operators can express "ideal environment per stage" as structured data; rules written once auto-adjust when cycles advance.
- Phase 21 RAG has its first-class **intent vs reality** signal — "what the operator *wanted* this zone to look like" joins cleanly against sensor_readings + automation_runs to explain yield variance.

## Risks / things to watch

- **Precedence bugs** — cycle > zone > null is subtle. The smoke test for "cycle shadows zone" is a must-have; add a second one for "deleting the cycle row uncovers the zone default" to prove cascade behavior.
- **`op` vocabulary creep** — resist adding `op` values beyond the initial four (`out_of_range`, `below_ideal`, `above_ideal`, `inside_range`). They cover the 95% case, and every new op is a new rule-engine code path.
- **Additive-only discipline** — it will be tempting to rename `zone_setpoints` something more generic ("environmental_targets", etc.). Don't. Shipping a mediocre name on an additive table costs nothing; renaming a table post-RAG costs pain.

---

## Using this plan in a new chat (copy-paste prompt)

```text
Implement Phase 20.6 per @docs/plans/phase_20_6_stage_scoped_setpoints.plan.md.

Scope:
1) WS1 — Migration + schema mirror + sqlc queries for gr33ncore.zone_setpoints (additive only). Include the precedence resolver query GetActiveSetpointForScope.
2) WS2 — CRUD handlers + OpenAPI for /farms/{id}/setpoints and /setpoints/{id}.
3) WS3 — Extend predicate shape in internal/automation/predicates.go with a "type" discriminator ("hard" | "setpoint"); implement setpoint resolution at eval time with graceful "no_setpoint_for_scope" skip.
4) WS4 — Setpoints.vue + SetpointRow.vue; Zone detail + Crop Cycle detail entry points; extend RuleForm.vue predicate picker with "use setpoint" toggle.
5) WS5 — Smoke (CRUD, scope CHECK, precedence, graceful skip); update workflow-guide.md §4 + glossary; OpenAPI audit.

Constraints: additive schema only — no changes to existing tables, no enum changes. Reuse the Phase 20 WS2 shared predicate evaluator — extend it, don't fork. Run go test ./cmd/api/..., go test ./..., python3 -m pytest pi_client/test_gr33n_client.py -q, and npm run build in ui/ after each WS. Update this plan's YAML todo statuses when each WS lands.
```
