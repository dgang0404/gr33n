# Local operator bootstrap — start here

Single happy path for standing up **Postgres → API → dashboard → optional Insert Commons receiver → optional Pi / MQTT bridge**, with explicit env templates and pointers to federation and audit docs. For farm template behavior (blank vs starter pack), see [`plans/phase_15_farm_onboarding.plan.md`](plans/phase_15_farm_onboarding.plan.md).

## Prerequisites

| Need | Native install | Docker only |
|------|----------------|-------------|
| Go | 1.23+ | Optional (API runs in container) |
| Node.js | 22+ recommended (`npm` for UI) | Optional if you only use the UI container |
| PostgreSQL | 14+ with **TimescaleDB** and **PostGIS** (schema runs `CREATE EXTENSION`) | Provided by Compose |
| Docker | — | Docker Engine + Compose v2 |

Detailed native Postgres steps (peer auth, roles): [`INSTALL.md`](../INSTALL.md).

## One-command bootstrap

From the repository root:

```bash
./scripts/bootstrap-local.sh
```

Options:

| Flag | Meaning |
|------|---------|
| `--docker` | `docker compose up -d` instead of host `psql` schema steps |
| `--seed` | Load [`db/seeds/master_seed.sql`](../db/seeds/master_seed.sql) (legacy demo **farm_id = 1**). Omit if you rely on dashboard **New farm** + template choice. |
| `--skip-schema` | Skip `psql` schema and migrations (database already provisioned) |

The script copies [`.env.example`](../.env.example) to `.env` **once** if `.env` is missing, then runs `npm ci` in `ui/`.

**Make equivalent:** `make bootstrap-local` (same as the script without flags). Use `make bootstrap-local-docker` for the Docker path.

## Order of operations

1. **Database** — Full schema: `db/schema/gr33n-schema-v2-FINAL.sql`. Upgrades on older snapshots: apply `db/migrations/*.sql` in **filename sort order** (the bootstrap script does this after the schema).
2. **Environment** — Root [`.env.example`](../.env.example): `DATABASE_URL`, `AUTH_MODE`, and for real auth `JWT_SECRET` / `PI_API_KEY`. The API loads `.env` then `.env.local` from the repo root.
3. **API** — `make run` (dev auth bypass) or `make run-auth` / production-style config; see comments in `.env.example`.
4. **UI** — `make ui` or `make dev` (API + UI). Copy [`ui/.env.example`](../ui/.env.example) to `ui/.env` if you need a non-default API URL (`VITE_API_URL`; otherwise the code defaults to `http://localhost:8080`).
5. **Optional: Insert Commons receiver** — `make run-receiver`; env and migrations: [`insert-commons-receiver-playbook.md`](insert-commons-receiver-playbook.md).
6. **Optional: Pi client / MQTT** — [`pi_client/setup.sh`](../pi_client/setup.sh), [`mqtt-edge-operator-playbook.md`](mqtt-edge-operator-playbook.md). Python deps: `pi_client/requirements.txt`.

## First user and auth

- **`AUTH_MODE=dev`** (default in `make run` / `make dev`): use the UI **Register** flow or `POST /auth/register` with `email`, `password` (minimum 8 characters), optional `full_name`.
- **Production**: set `AUTH_MODE=production`, `JWT_SECRET`, and `PI_API_KEY`; optional env-admin login via `ADMIN_USERNAME` + `ADMIN_PASSWORD_HASH` in `.env` (see `.env.example`).

## Insert Commons and custom integrators

Farm-side pipeline and **strict ingest JSON** (only six top-level keys; no extra fields): [`insert-commons-pipeline-runbook.md`](insert-commons-pipeline-runbook.md) — read **Custom senders** before POSTing from scripts or third-party tools.

## Audit and operator index

- Farm audit API and sensitive actions: [`audit-events-operator-playbook.md`](audit-events-operator-playbook.md).
- Phase 14 playbook index (MQTT, commons catalog, notifications, etc.): [`phase-14-operator-documentation.md`](phase-14-operator-documentation.md).

## Security notes

Bootstrap keeps **secrets and TLS** in your hands: the script does not generate passwords or certificates. Use real secrets in production; do not commit `.env`.
