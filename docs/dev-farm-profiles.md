# Dev farm seed profiles

**Audience:** Operators and developers running local/staging Postgres.

**Plan:** [`plans/phase_48_dev_seed_and_small_farm_profiles.plan.md`](plans/phase_48_dev_seed_and_small_farm_profiles.plan.md)

Farm **profiles** control how much demo config `master_seed.sql` and reset scripts load. Profiles are stored on `gr33ncore.farms.meta_data->>'dev_seed_profile'`.

---

## Profiles

### `small_indoor` (daily dev + sit-in)

| Entity | Target (farm 1) |
|--------|-----------------|
| Zones | 2 active — Veg Room, Flower Room (Outdoor Garden soft-deleted) |
| Sensors | ~8–10 active (no outdoor soil / extra CO₂ duplicate sprawl) |
| Actuators | As seeded devices allow (~4–6) |
| Programs / schedules | One feeding + lighting path per active zone |
| Automation rules | Seeded rules; keep inactive for dev |
| Tasks | Handful with `zone_id` set |

**Apply:**

```bash
./scripts/dev-reset-farm.sh --farm-id 1 --profile small_indoor
```

### `demo_showcase` (operator tour + RAG demos)

Full `master_seed.sql` — JADAM inputs, three zones, Guardian demo alerts, fertigation history, greenhouse templates. This is the **default** tag on farm 1 after `make seed` / `make dev-stack`.

**Apply:**

```bash
./scripts/dev-reset-farm.sh --farm-id 1 --profile demo_showcase
# or: make seed   (idempotent; stamps demo_showcase on farm 1)
```

---

## Commands

| Command | When |
|---------|------|
| `make seed` | Re-apply idempotent `master_seed.sql` on existing volume |
| `./scripts/dev-reset-farm.sh --farm-id 1 --profile small_indoor` | Surgical reset without Docker volume wipe |
| `./scripts/dev-reset-farm.sh --farm-id 1 --profile demo_showcase --include-readings` | Full demo + truncate readings |
| `make dev-stack-fresh` | Nuclear — `docker compose down -v` + bootstrap + seed |
| `make db-sanity-report` | Sensor counts, duplicates, profile tag, readings estimate |

---

## Bootstrap templates vs profiles

| Mechanism | Use |
|-----------|-----|
| **`jadam_indoor_photoperiod_v1`** bootstrap | New farms from setup wizard — idempotent zones/sensors/schedules (~small indoor footprint) |
| **`demo_showcase` profile** | Pre-loaded farm 1 after `make seed` |
| **`small_indoor` profile** | Trim farm 1 for fast UI and Phase 45-style validation |

Bootstrap templates do **not** replace profiles — they stamp rows on **new** farms. Farm 1 demo data comes from `master_seed.sql`.

---

## Environment

| Variable | Purpose |
|----------|---------|
| `DEV_SEED_PROFILE` | Default profile for `dev-reset-farm.sh` when `--profile` omitted (`small_indoor`) |
| `TIMESCALE_RETENTION_DAYS` | When set, `./scripts/apply-dev-retention.sh` adds hypertable retention (dev/staging only) |

See [local-operator-bootstrap.md](local-operator-bootstrap.md) § Slow UI and dev DB hygiene.
