---
name: Phase 20.8 Animal Husbandry Flesh-Out
overview: >
  Turns the dormant `gr33nanimals.animal_groups` table (shipped as a thin
  placeholder in Phase 14) into a working animal-husbandry surface. Adds the
  minimum columns a chicken coop / goat pen / small rabbitry actually needs
  (count, primary_zone_id, active/archived_at), plus a sibling
  `animal_lifecycle_events` table that mirrors the crop_cycles timeline shape.
  Leans on Phase 20.7's broadened input_category_enum (animal_feed / bedding /
  veterinary_supply) so feed consumption auto-logs cost exactly like fertigation
  does. Also fleshes out `gr33naquaponics.loops` with typed FKs so fish→plant
  coupling is queryable. All changes additive. Target: 4–5 days.
todos:
  - id: ws1-schema-additions
    content: "WS1: Additive migrations — animal_groups.(count, primary_zone_id, active, archived_at, archived_reason); new gr33nanimals.animal_lifecycle_events table; aquaponics.loops.(fish_tank_zone_id, grow_bed_zone_id); regenerated sqlc"
    status: pending
  - id: ws2-crud-handlers
    content: "WS2: CRUD handlers for animal_groups + animal_lifecycle_events; aquaponics.loops CRUD; OpenAPI schemas + paths"
    status: pending
  - id: ws3-feed-consumption-wiring
    content: "WS3: Wire animal feed into Phase 20.7 task-consumption flow — operator records 'fed herd X, used 2kg feed' on a task; cost auto-logs as category='feed_livestock'; verify category override path in autologger"
    status: pending
  - id: ws4-ui-animals-page
    content: "WS4: Animals.vue under Operate — list groups per farm, inline edit count, lifecycle timeline view, link from Zone detail; Aquaponics.vue for loops (or merge into Zone detail); HelpTips explaining the 'use primitives' approach for climate/feeding/watering"
    status: pending
  - id: ws5-bootstrap-upgrade
    content: "WS5: Upgrade Phase 20.5 chicken_coop_v1 + small_aquaponics_v1 bootstraps to seed an animal_group / loop row with count + zone links (idempotent, won't break existing farms that already ran the bootstrap)"
    status: pending
  - id: ws6-smoke-and-docs
    content: "WS6: Smoke — CRUD + lifecycle event insert/archive; feed consumption end-to-end via a task; bootstrap re-run idempotency. Docs: new workflow-guide.md §12 'Animals & Aquaponics'; glossary entries; OpenAPI audit"
    status: pending
isProject: false
---

# Phase 20.8 — Animal Husbandry Flesh-Out

## Why this phase

The schema has three thin placeholder tables — `gr33ncrops.plants`, `gr33nanimals.animal_groups`, `gr33naquaponics.loops` — all shipped in Phase 14 as scaffolding. Each is just `(id, farm_id, label, [species], meta JSONB, timestamps)`. Phase 14 deliberately shipped them "reserved for later." Later is here.

Phase 20.5 just added bootstrap templates for `chicken_coop_v1` and `small_aquaponics_v1`, and Phase 20.7 just broadened `input_category_enum` with `animal_feed` / `bedding` / `veterinary_supply`. What's still missing is the minimum structured state needed to answer "how many animals do I have, where are they, when did this batch of chicks arrive, when did I retire this group?" Without it, operators have to put everything in meta JSONB, which RAG can't reason about.

This phase is deliberately small and purely additive. No renames, no drops, no enum changes (Phase 20.7 already landed the enum widening this phase depends on).

## Hand-offs from earlier phases (reuse, don't re-implement)

- **Phase 14** landed the placeholder tables and the `gr33nanimals` / `gr33naquaponics` schemas. This phase fleshes them out; don't create new schemas.
- **Phase 20.5** shipped `chicken_coop_v1` and `small_aquaponics_v1` bootstraps that create zones / sensors / actuators / rules / tasks but don't touch the `animal_groups` / `loops` tables. WS5 here upgrades those bootstraps to seed one row in each.
- **Phase 20.7** widened `input_category_enum` with `animal_feed`, `bedding`, `veterinary_supply` AND built the `task_input_consumptions` flow with automatic cost logging. This phase doesn't reinvent that — it reuses it. Animal feed is just another `input_definition` with `category = animal_feed`; feeding the herd is a task with a `task_input_consumption` row pointing at the feed batch. Cost auto-logs with `category = 'feed_livestock'` (from the `cost_category_enum`). Mapping `animal_feed` input → `feed_livestock` cost category is a one-line lookup in `internal/costing/autologger.go`.
- **Crop cycle timeline shape** — `gr33nfertigation.crop_cycles` has `started_at`, `harvested_at`, `is_active`, `current_stage`, `cycle_notes`. `animal_lifecycle_events` mirrors this pattern row-by-row; RAG queries written against crop_cycles generalise trivially.

## Scope

| WS | Focus | Location in repo |
|----|-------|------------------|
| **WS1** | Additive migrations | `db/migrations/2026xxxx_phase208_animal_husbandry.sql` + schema mirror; sqlc regen |
| **WS2** | CRUD + OpenAPI | `internal/handler/animal/handler.go` (new), `internal/handler/aquaponics/handler.go` (new), `cmd/api/routes.go`, `openapi.yaml` |
| **WS3** | Feed consumption wiring | `internal/costing/autologger.go` extension, input→cost-category lookup table |
| **WS4** | UI | `ui/src/views/Animals.vue` (new), `ui/src/views/Aquaponics.vue` (new or Zone-detail-embedded) |
| **WS5** | Bootstrap upgrade | `db/migrations/2026xxxx_phase208_bootstrap_upgrade.sql` (PL/pgSQL patch to the existing template functions, idempotent) |
| **WS6** | Smoke + docs | `cmd/api/smoke_test.go`, `docs/workflow-guide.md` new §12, glossary |

## Work-stream detail

### WS1 — Additive migrations

```sql
-- animal_groups — operational columns
ALTER TABLE gr33nanimals.animal_groups
  ADD COLUMN IF NOT EXISTS count              INTEGER CHECK (count IS NULL OR count >= 0),
  ADD COLUMN IF NOT EXISTS primary_zone_id    BIGINT REFERENCES gr33ncore.zones(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS active             BOOLEAN NOT NULL DEFAULT TRUE,
  ADD COLUMN IF NOT EXISTS archived_at        TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS archived_reason    TEXT;

CREATE INDEX IF NOT EXISTS idx_animal_groups_zone
  ON gr33nanimals.animal_groups (primary_zone_id)
  WHERE deleted_at IS NULL;

-- lifecycle events (mirrors crop_cycles timeline shape)
CREATE TABLE IF NOT EXISTS gr33nanimals.animal_lifecycle_events (
  id               BIGSERIAL PRIMARY KEY,
  farm_id          BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
  animal_group_id  BIGINT NOT NULL REFERENCES gr33nanimals.animal_groups(id) ON DELETE CASCADE,
  event_type       TEXT NOT NULL,        -- 'added', 'born', 'died', 'sold', 'culled', 'health_event', 'weight_check', 'note'
  event_time       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  delta_count      INTEGER,              -- +N for added/born, -N for died/sold/culled, NULL for note/health
  notes            TEXT,
  recorded_by      UUID REFERENCES gr33ncore.profiles(user_id) ON DELETE SET NULL,
  related_task_id  BIGINT REFERENCES gr33ncore.tasks(id) ON DELETE SET NULL,
  meta             JSONB NOT NULL DEFAULT '{}',
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_lifecycle_events_group_time
  ON gr33nanimals.animal_lifecycle_events (animal_group_id, event_time DESC);

-- aquaponics.loops typed FKs (in addition to existing free-form label + meta)
ALTER TABLE gr33naquaponics.loops
  ADD COLUMN IF NOT EXISTS fish_tank_zone_id  BIGINT REFERENCES gr33ncore.zones(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS grow_bed_zone_id   BIGINT REFERENCES gr33ncore.zones(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS active             BOOLEAN NOT NULL DEFAULT TRUE;
```

`event_type` is TEXT not an enum — deliberately. Different farms track different events (weight checks for livestock, vaccination for rabbits, water-quality notes for aquaponics fish). Keeping it TEXT lets operators and later playbooks converge organically; we can tighten to an enum in a future phase once the real vocabulary emerges.

`delta_count` is the one calculated field — a computed-view-or-trigger approach could keep `animal_groups.count` in sync with the sum of `delta_count` over lifecycle events. **Don't** ship that trigger in this phase. Keep `count` as a manually-edited field and render "sum of lifecycle deltas" as a UI sanity check. Trigger-based reconciliation is tempting but fragile; defer until real customers complain.

### WS2 — CRUD + OpenAPI

- Routes (all JWT, member authz):
  - `GET /farms/{id}/animal-groups`, `POST /farms/{id}/animal-groups`
  - `GET /animal-groups/{id}`, `PUT /animal-groups/{id}`, `DELETE /animal-groups/{id}` (soft — sets `deleted_at`; uses the existing partial-index pattern)
  - `PATCH /animal-groups/{id}/archive` → sets `active=false, archived_at=NOW(), archived_reason=...` (distinct from soft delete — archiving preserves lifecycle history; deleting is for mistake-entry cleanup)
  - `GET /animal-groups/{id}/lifecycle-events`, `POST /animal-groups/{id}/lifecycle-events`
  - `DELETE /lifecycle-events/{id}` (admin-only; lifecycle events should be corrected via new compensating events, not deleted, but preserve the hatch)
  - `GET /farms/{id}/aquaponics-loops`, `POST /farms/{id}/aquaponics-loops`, `GET|PUT|DELETE /aquaponics-loops/{id}`
- OpenAPI schemas: `AnimalGroup`, `AnimalGroupCreate`, `AnimalGroupUpdate`, `LifecycleEvent`, `LifecycleEventCreate`, `AquaponicsLoop`, `AquaponicsLoopCreate`.
- Validation: `primary_zone_id` and `fish_tank_zone_id` / `grow_bed_zone_id` must belong to the farm in the URL (same farm-scoping check used elsewhere).

### WS3 — Feed consumption wiring

- In `internal/costing/autologger.go`, the existing `LogTaskConsumption` computes cost category. Extend it with a lookup: `input_definitions.category` → `cost_category_enum` mapping.
  - `animal_feed` → `feed_livestock`
  - `bedding` → `feed_livestock` (close enough for pre-RAG; split later if needed)
  - `veterinary_supply` → `veterinary_services`
  - All existing fertilizer categories → `fertilizers_soil_amendments` (unchanged)
- Update the OpenAPI note on `cost_transactions.category` to describe the auto-mapping.
- **No new endpoints needed** — operators create a `task` linked to an animal zone, attach a `task_input_consumption` row with the feed batch, and the cost logs automatically.

### WS4 — UI

- **`Animals.vue`** under Operate → Animals:
  - List all animal_groups for a farm (active + archived filterable).
  - Each row: label, species, count, primary zone (linked), lifecycle timeline count, last event.
  - Click → detail drawer with timeline of `lifecycle_events` (reverse chronological, color-coded by event_type), inline "add event" form.
  - Archive button + confirmation modal.
- **`Aquaponics.vue`** — small; lists loops with fish_tank_zone and grow_bed_zone as clickable zone badges. Alternatively embed into Zone detail as a "this zone is part of aquaponics loop X" card. Decide in implementation; both are fine.
- **HelpTips** explain that feeding, watering, and climate all work through the existing primitives (sensors / actuators / tasks / rules) — the animal_group row is just a way to *count heads and track lifecycle*, not to duplicate the hardware layer.
- **SideNav** gets Operate → "Animals" entry (next to Plants / Rules / Schedules). Visible only when the farm has the `gr33nanimals` domain module active in `farm_active_modules` (Phase 14 pattern — reuse).

### WS5 — Bootstrap upgrade

- Patch `gr33ncore.apply_farm_bootstrap_template` so the `chicken_coop_v1` branch inserts one `gr33nanimals.animal_groups` row (`label='Layer Flock', species='chicken', count=12, primary_zone_id=<coop zone>`). Same for `small_aquaponics_v1` (loop row with `fish_tank_zone_id` + `grow_bed_zone_id`).
- Idempotent: use `ON CONFLICT DO NOTHING` + a new partial unique index on `(farm_id, label) WHERE deleted_at IS NULL` if one doesn't already exist.
- Farms that already ran the Phase 20.5 bootstraps can re-run with no effect on existing rows — the patch adds new rows only where the `animal_groups` / `loops` table is empty for that farm.

### WS6 — Smoke + docs

- Smoke:
  - Animal group CRUD + archive (verify archived rows still appear in lifecycle-event list but not in default list).
  - Lifecycle event insert with `delta_count=+3` and another with `-1`, then fetch the group; `count` stays whatever operator set (we don't auto-reconcile in this phase).
  - Feed-consumption end-to-end: create `animal_feed` input + batch with unit_cost, create task with a `task_input_consumption` row, verify one `cost_transactions` row lands with `category='feed_livestock'`, `related_record_id = <consumption row>`, and `current_quantity_remaining` decremented.
  - Bootstrap re-idempotency: run `chicken_coop_v1` twice against the same farm; assert exactly one `animal_groups` row, not two.
- Docs:
  - New `docs/workflow-guide.md` §12 "Animals & Aquaponics" — describes the animal_group as *the head-count and timeline anchor*, not a replacement for sensors/actuators/rules. Example walkthrough: "a 12-hen layer flock that auto-feeds at 06:00, alerts on low water, and logs feed cost automatically."
  - Glossary: `animal_group`, `lifecycle_event`, `aquaponics_loop`.
  - Cross-link `docs/pattern-playbooks.md` (from Phase 20.5) to the new §12.

## After Phase 20.8

- Every thin Phase-14 placeholder table is now fleshed out just enough to be useful.
- RAG Phase 21 can cross-reference crop cycles and animal groups uniformly when summarising a mixed-use farm.
- Cost auto-attribution (Phase 20.7) now covers animal operations, not just plant fertigation.

## Risks / things to watch

- **count-drift** — the manual `animal_groups.count` vs the summed `delta_count` from lifecycle_events will diverge in practice. UI must surface both ("you've recorded +14 / -2 = 12 events, stored count is 11 — reconcile?"). Do NOT add a trigger in this phase. Operators make the call.
- **event_type freeform** — resist the urge to enum it now. Let the field populate from real use for one or two customer sites, then enum in a later phase once the vocabulary has settled.
- **Bootstrap migration safety** — patching an existing PL/pgSQL function that farms already ran against is fine (functions are replaceable), but make sure the patch NEVER creates duplicate rows on re-run.

---

## Using this plan in a new chat (copy-paste prompt)

```text
Implement Phase 20.8 per @docs/plans/phase_20_8_animal_husbandry_flesh_out.plan.md.

Scope:
1) WS1 — Additive migrations: animal_groups new columns, animal_lifecycle_events (new table), aquaponics.loops typed FKs. Mirror in schema file. Regenerate sqlc.
2) WS2 — CRUD handlers for animal-groups + lifecycle-events + aquaponics-loops; OpenAPI schemas + paths; farm-scoping validation on zone FKs.
3) WS3 — Extend internal/costing/autologger.go with input_definitions.category → cost_category_enum lookup (animal_feed → feed_livestock, veterinary_supply → veterinary_services).
4) WS4 — Animals.vue + Aquaponics.vue (or Zone-detail embed); HelpTips stating the "use primitives for hardware, animal_group for head-count+timeline" model; SideNav entry gated on gr33nanimals domain module.
5) WS5 — Patch apply_farm_bootstrap_template: chicken_coop_v1 seeds an animal_groups row; small_aquaponics_v1 seeds a loops row. Idempotent.
6) WS6 — Smoke (CRUD, lifecycle, feed end-to-end, bootstrap re-run idempotency). Workflow-guide §12 + glossary. OpenAPI audit.

Constraints: additive schema only — new columns (nullable) + new table + new FK columns. NO trigger-based count reconciliation. NO enum for event_type. Reuse Phase 20.7's autologger and task_input_consumptions. Run go test ./cmd/api/..., go test ./..., python3 -m pytest pi_client/test_gr33n_client.py -q, and npm run build in ui/ after each WS. Update this plan's YAML todo statuses when each WS lands.
```
