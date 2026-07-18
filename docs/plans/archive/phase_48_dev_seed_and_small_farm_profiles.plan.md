---
name: Phase 48 — Dev seed hygiene & small farm profiles
overview: >
  Stop local/staging databases from accumulating duplicate sensors, bloated demo config,
  and unbounded time-series rows. Introduce named farm profiles (small_indoor vs
  demo_showcase), idempotent seeding, dev reset scripts, and optional Timescale retention
  for non-production. Unblocks fast UI dev, Phase 44 wizards, and Phase 45 sit-ins.
todos:
  - id: ws1-profiles-spec
    content: "WS1: Profile spec — small_indoor (2–3 zones, ~12 sensors) vs demo_showcase; document entity counts and which bootstrap templates each applies"
    status: completed
  - id: ws2-seed-idempotency
    content: "WS2: master_seed idempotency — unique (farm_id, name) on sensors/zones where safe; upsert-by-name instead of blind INSERT; fix ON CONFLICT DO NOTHING on sensors"
    status: completed
  - id: ws3-dev-reset-script
    content: "WS3: scripts/dev-reset-farm.sh — reset farm 1 config + readings without docker volume wipe; document vs dev-stack --reset-volumes"
    status: completed
  - id: ws4-bootstrap-alignment
    content: "WS4: Bootstrap template guards audit — apply_farm_bootstrap_template idempotency; default new dev farms to small_indoor profile"
    status: completed
  - id: ws5-timescale-retention
    content: "WS5: Optional dev retention — add_retention_policy on sensor_readings/actuator_events (env-gated); document in workflow-guide + operator-logging-runbook cross-link"
    status: completed
  - id: ws6-sanity-report
    content: "WS6: Extend db-sanity-report — sensor count per farm, duplicate names, readings row count, profile tag in farms.meta_data"
    status: completed
  - id: ws7-docs-tests
    content: "WS7: local-operator-bootstrap § dev profiles; operator-tour note; architecture §7.0n; smoke idempotent re-seed; OC-48"
    status: completed
isProject: false
---

# Phase 48 — Dev seed hygiene & small farm profiles

## Status

**Shipped.** WS1–WS7 complete — profiles, idempotent seed, `dev-reset-farm.sh`, sanity metrics, optional retention, OC-48 closed.

**Roadmap:** [farmer_ux_roadmap_40_plus.plan.md](farmer_ux_roadmap_40_plus.plan.md) (dev hygiene track)

**Closure:** **OC-48** in [phase_35_37_operational_closure.plan.md](phase_35_37_operational_closure.plan.md)

---

## Problem

Local dev databases **accumulate** even when operators never “add junk” manually:

| Cause | Effect |
|-------|--------|
| Docker volume persists across reboots | `restart-local.sh` / `make dev-auth-test` do **not** wipe Postgres |
| Re-running `master_seed.sql` | Sensors lack `(farm_id, name)` unique key — each run can insert **duplicate** rows with new serial IDs |
| Multiple bootstrap template applies | Mostly idempotent by name, but stacks zones/rules/devices over months |
| Edge stub / Pi client | `sensor_readings` hypertable grows without bound |
| Timescale hypertables without retention | Chunks partition data but **do not delete** old rows |

Symptoms: slow dashboard boot, hundreds of `/sensors/*/readings/latest` calls (mitigated by batch endpoint), automation worker noise, confusing Guardian context.

**Not the same as:** application log rotation ([operator-logging-runbook.md](../operator-logging-runbook.md)) — that is stdout/docker logs, not DB rows.

---

## Design principles

1. **Profiles, not one mega seed** — `small_indoor` for daily dev and sit-in; `demo_showcase` for tours and RAG demos.
2. **Idempotent by name** — re-running seed or bootstrap must not duplicate config entities.
3. **Surgical reset** — `dev-reset-farm-1` without `--reset-volumes` for most days; full volume wipe stays rare and documented.
4. **Production-safe defaults** — retention policies and truncate scripts are **dev/staging gated** (`DEV_SEED_PROFILE`, `TIMESCALE_RETENTION_DAYS`, or explicit `--i-know-this-is-dev`).
5. **No farmer UI** — this phase is operator/docs/scripts/migrations only.

---

## WS1 — Farm profiles spec

Define two first-class profiles (extensible later):

### `small_indoor` (default for new dev)

| Entity | Target count (farm 1) |
|--------|------------------------|
| Zones | 2–3 (Veg, Flower; optional Outdoor) |
| Sensors | ~10–12 (one per need per zone, no duplicates) |
| Actuators | ~4–6 (pump, light, fan/shade per active zone) |
| Programs / schedules | One feeding + one lighting program per active zone |
| Automation rules | ≤6 active (comfort + one GH template optional) |
| Tasks | Handful of open tasks with `zone_id` set |

### `demo_showcase` (current master_seed ambition)

Full JADAM inputs, multi-zone fertigation, Guardian demo alerts, greenhouse templates — for operator tour and conference demos. **Not** the default after Phase 48.

**Storage:** `gr33ncore.farms.meta_data->>'dev_seed_profile'` or bootstrap apply flag. `db-sanity-report` prints active profile.

**Deliverable:** `docs/dev-farm-profiles.md` (or section in local-operator-bootstrap) with entity counts and apply commands.

---

## WS2 — Seed idempotency (migration + master_seed)

1. **Migration:** add unique indexes where safe, e.g. `(farm_id, name)` on `gr33ncore.sensors` **where deleted_at IS NULL** (partial unique), same pattern for zones if not already enforced.
2. **Refactor `master_seed.sql`:**
   - Replace blind `INSERT … ON CONFLICT DO NOTHING` on sensors with `INSERT … ON CONFLICT (farm_id, name) DO UPDATE` or `WHERE NOT EXISTS` by name (match zones pattern).
   - Split seed into `master_seed_small.sql` + `master_seed_showcase.sql` **or** profile flag at top of one file.
3. **Regression:** run seed twice; sensor count for farm 1 must not increase.

---

## WS3 — Dev reset script

**New:** `scripts/dev-reset-farm.sh`

```bash
# Example intent (implementation in WS3):
#   ./scripts/dev-reset-farm.sh --farm-id 1 --profile small_indoor
#   ./scripts/dev-reset-farm.sh --farm-id 1 --profile demo_showcase --include-readings
```

| Mode | Behavior |
|------|----------|
| Default | Truncate farm-scoped config children + re-apply profile seed; keep auth users |
| `--include-readings` | Also truncate `sensor_readings` / `actuator_events` for farm’s sensors |
| Full wipe | Still `./scripts/dev-stack.sh --reset-volumes` — document when needed |

Link from [local-operator-bootstrap.md](../local-operator-bootstrap.md) § troubleshooting slow dev UI.

---

## WS4 — Bootstrap template alignment

- Audit `gr33ncore.apply_farm_bootstrap_template` branches for `NOT EXISTS` vs duplicate inserts.
- Dashboard “Apply template” should show profile impact (“adds ~8 sensors”).
- [Phase 44](phase_44_getting_started_edge_wizard.plan.md) wizard should call **`small_indoor`** by default, `demo_showcase` as optional “full demo pack”.

---

## WS5 — Timescale retention (dev/staging)

Hypertables today ([schema](../db/schema/gr33n-schema-v2-FINAL.sql)): `sensor_readings`, `actuator_events`, `weather_data`, `user_activity_log`, `system_logs`.

**WS5 adds** (migration, env-gated apply):

```sql
-- Illustrative — exact intervals in implementation
SELECT add_retention_policy('gr33ncore.sensor_readings', INTERVAL '90 days');
SELECT add_retention_policy('gr33ncore.actuator_events', INTERVAL '90 days');
```

Only when `TIMESCALE_RETENTION_DAYS` set or `DEV_SEED_PROFILE` present — **never** auto-apply on production farms without explicit ops runbook.

Cross-link [workflow-guide.md](../workflow-guide.md) and [operator-logging-runbook.md](../operator-logging-runbook.md) § data vs log retention.

---

## WS6 — Sanity report extensions

Extend [scripts/sql/db_sanity_report.sql](../scripts/sql/db_sanity_report.sql):

- Sensors per farm (active vs soft-deleted)
- Duplicate `(farm_id, name)` sensor rows
- `sensor_readings` approximate row count
- `dev_seed_profile` from farm meta_data
- **Warn** if farm 1 sensor count > 2× profile target

Exit non-zero on duplicate names (already zones); add sensor duplicate check.

---

## WS7 — Docs, tests, closure (OC-48)

| Artifact | Content |
|----------|---------|
| [local-operator-bootstrap.md](../local-operator-bootstrap.md) | Profiles, reset script, when to use `--reset-volumes` |
| [operator-tour.md](../operator-tour.md) | Note: tour assumes `demo_showcase` |
| [farm-guardian-architecture.md](../farm-guardian-architecture.md) | §7.0n dev seed profiles |
| Go smoke | Re-seed idempotency; batch latest readings on small profile |
| Vitest | Optional — none required unless reset script gets JS helper |

**OC-48** closed (shipped).

---

## Relationship to other phases

| Phase | Relationship |
|-------|----------------|
| **43** | Operations hub — benefits from sane low-stock seed counts |
| **44** | Getting started wizard — **consumes** profile definitions from WS1/WS4 |
| **45** | Sit-in — **requires** `small_indoor` for realistic validation |
| **15** | Bootstrap API — profiles extend template keys, not replace |

---

## Out of scope

- Multi-tenant production data lifecycle (enterprise retention SLAs)
- Automatic orphan GC for soft-deleted entities across all tables
- Replacing Timescale with another TSDB
- Farmer-facing “delete my farm data” UI (privacy/export is a separate product decision)

---

## Definition of done

- [x] `small_indoor` and `demo_showcase` documented with entity targets
- [x] Re-running seed does not duplicate sensors on farm 1
- [x] `dev-reset-farm.sh` restores small profile without volume wipe
- [x] `db-sanity-report` flags bloat and duplicates
- [x] Optional dev retention policy documented and env-gated
- [x] local-operator-bootstrap + OC-48 closed

---

## Suggested implementation order

1. WS1 spec + WS6 sanity metrics (read-only, immediate value)
2. WS2 migration + seed split
3. WS3 reset script
4. WS4 bootstrap audit
5. WS5 retention (optional env)
6. WS7 docs + smokes

---

## Related

| Doc | Use |
|-----|-----|
| [phase_15_farm_onboarding.plan.md](phase_15_farm_onboarding.plan.md) | Bootstrap templates origin |
| [phase_44_getting_started_edge_wizard.plan.md](phase_44_getting_started_edge_wizard.plan.md) | Wizard picks profile |
| [phase_45_farmer_validation_whole_app_polish.plan.md](phase_45_farmer_validation_whole_app_polish.plan.md) | Sit-in on small farm |
| [db/seeds/master_seed.sql](../db/seeds/master_seed.sql) | Primary seed file to refactor |
| [scripts/dev-stack.sh](../scripts/dev-stack.sh) | `--reset-volumes` nuclear option |

---

## Using this in a new chat

> Read `docs/plans/archive/phase_48_dev_seed_and_small_farm_profiles.plan.md`. Implement one workstream (WS1–WS7). Prefer migrations + scripts over UI. Gate destructive ops behind dev env flags. Do not change production retention without runbook update.
