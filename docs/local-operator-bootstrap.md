# Local operator bootstrap — start here

**Run bootstrap and Make targets from the repository root** (`cd /path/to/gr33n-platform` after `git clone`). Commands like `./scripts/bootstrap-local.sh` and `make dev` apply to **this** repo only — not from your home directory (`~`).

Single happy path for standing up **Postgres → API → dashboard → optional Insert Commons receiver → optional Pi / MQTT bridge**, with explicit env templates and pointers to federation and audit docs. For farm template behavior (blank vs starter pack), see [`plans/phase_15_farm_onboarding.plan.md`](plans/phase_15_farm_onboarding.plan.md).

## Prerequisites

| Need | Native install | Docker only |
|------|----------------|-------------|
| Go | 1.23+ | Optional (API runs in container) |
| Node.js | 22+ recommended (`npm` for UI) | Optional if you only use the UI container |
| PostgreSQL | 14+ with **TimescaleDB** and **PostGIS** (schema runs `CREATE EXTENSION`) | Provided by Compose |
| Docker | — | Docker Engine + Compose v2 |

Detailed native Postgres steps (peer auth, roles): [`INSTALL.md`](../INSTALL.md).

## First clone (recommended for new contributors)

From the repository root after `git clone`, run:

```bash
./scripts/setup-first-clone.sh
```

Or **`make first-clone`**. This runs `go mod download`, copies `.env` / `ui/.env` from examples if missing, then **`scripts/bootstrap-local.sh`**. You still need PostgreSQL created with extensions first for the native path — see [INSTALL.md](../INSTALL.md). **Debian/Ubuntu:** install Postgres stack + Node with **`./scripts/install-system-deps-debian.sh`** (sudo apt), or combine with **`./scripts/setup-first-clone.sh --install-system-deps`** (`make first-clone-install-deps`). For a machine without local Postgres, use **`./scripts/setup-first-clone.sh --docker`** or **`make first-clone-docker`**.

For how the schema is defined (and why ad-hoc ERD screenshots may be outdated), see [database-schema-overview.md](database-schema-overview.md).

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

## When localhost (DB / API / UI) is not running

**Docker:** from the repo root run `docker compose up -d --build` (or `make bootstrap-local-docker`). Dashboard: **http://localhost:5173** · API: **http://localhost:8080** (`AUTH_MODE=dev` in Compose). Postgres is exposed on **localhost:5432** (credentials in [`docker-compose.yml`](../docker-compose.yml)).

**Native:** follow [INSTALL.md](../INSTALL.md) for Postgres + extensions, then `./scripts/bootstrap-local.sh`, set **`DATABASE_URL`** in `.env`, then **`make dev`** (API + UI together) or **`make run`** and **`make ui`** in two terminals.

## Order of operations

1. **Database** — Full schema: `db/schema/gr33n-schema-v2-FINAL.sql`. Upgrades on older snapshots: apply `db/migrations/*.sql` in **filename sort order** (the bootstrap script does this after the schema).
2. **Environment** — Root [`.env.example`](../.env.example): `DATABASE_URL`, `AUTH_MODE`, and for real auth `JWT_SECRET` / `PI_API_KEY`. The API loads `.env` then `.env.local` from the repo root.
3. **API** — `make run` (dev auth bypass) or `make run-auth` / production-style config; see comments in `.env.example`.
4. **UI** — `make ui` or `make dev` (API + UI). Copy [`ui/.env.example`](../ui/.env.example) to `ui/.env` if you need a non-default API URL (`VITE_API_URL`; otherwise the code defaults to `http://localhost:8080`).
5. **Optional: Insert Commons receiver** — `make run-receiver`; env and migrations: [`insert-commons-receiver-playbook.md`](insert-commons-receiver-playbook.md).
6. **Optional: Pi client / MQTT** — OS packages: [`scripts/install-pi-edge-deps.sh`](../scripts/install-pi-edge-deps.sh). Then [`pi_client/setup.sh`](../pi_client/setup.sh), [`mqtt-edge-operator-playbook.md`](mqtt-edge-operator-playbook.md). Full topologies (edge vs all-on-Pi vs split servers): [`raspberry-pi-and-deployment-topology.md`](raspberry-pi-and-deployment-topology.md). Python deps: `pi_client/requirements.txt`.

## API integration smoke tests

Run from repo root: `go test -tags dev ./cmd/api/...` (or `make test`, which includes this package). The `cmd/api` tests build an in-memory `httptest` server wired like production, with **`AUTH_MODE=auth_test`** and fixed test-only `JWT_SECRET` / `PI_API_KEY` (not read from your `.env`).

| Requirement | Notes |
|---------------|--------|
| **`DATABASE_URL`** | Must point at Postgres that already has **full schema** (`db/schema/gr33n-schema-v2-FINAL.sql`) and **migrations** applied (same order as bootstrap). Export it in the shell before `go test`, or rely on the Makefile default `DB_URL` when you run `make test`. |
| **`-tags dev`** | Required so `auth_test` mode is allowed (`make test` sets this). |
| **Seed data** | Recommended: [`db/seeds/master_seed.sql`](../db/seeds/master_seed.sql) (demo **farm_id = 1**, sensors, NF inputs, alerts, etc.). A few tests **skip** if expected rows are missing (e.g. “no sensors in seed”, “no NF inputs in seed data”). |
| **No database** | If the pool cannot open or ping, `TestMain` prints a **stderr hint** and exits **0** locally (so `go test ./...` without Postgres does not fail every package). In **CI** (`CI=true` or `GITHUB_ACTIONS`), the same condition exits **1** so a forgotten DB service does not look green. |
| **Unset `DATABASE_URL`** | Tests use a **Linux peer-auth default** (`postgres://davidg@/gr33n?host=/var/run/postgresql`). Override with `DATABASE_URL` if your user or socket path differs. |

Do not use `go test -shuffle=on` on this package as a gate — smoke tests share package-level state (see Phase 20.95 plan notes).

## First user and auth

- **`AUTH_MODE=dev`** (default in `make run` / `make dev`): use the UI **Register** flow or `POST /auth/register` with `email`, `password` (minimum 8 characters), optional `full_name`.
- **Production**: set `AUTH_MODE=production`, `JWT_SECRET`, and `PI_API_KEY`; optional env-admin login via `ADMIN_USERNAME` + `ADMIN_PASSWORD_HASH` in `.env` (see `.env.example`).

## Insert Commons and custom integrators

Farm-side pipeline and **strict ingest JSON** (only six top-level keys; no extra fields): [`insert-commons-pipeline-runbook.md`](insert-commons-pipeline-runbook.md) — read **Custom senders** before POSTing from scripts or third-party tools.

## Audit and operator index

- Farm audit API and sensitive actions: [`audit-events-operator-playbook.md`](audit-events-operator-playbook.md).
- Phase 14 playbook index (MQTT, commons catalog, notifications, etc.): [`phase-14-operator-documentation.md`](phase-14-operator-documentation.md).

## OpenAPI route audit

From the repo root, `make audit-openapi` runs [`scripts/openapi_route_diff.sh`](../scripts/openapi_route_diff.sh). It diffs **(HTTP method, path)** pairs from [`cmd/api/routes.go`](../cmd/api/routes.go) against [`openapi.yaml`](../openapi.yaml) and exits non-zero on any mismatch — run it after you add or rename HTTP routes.

**Edge vs dashboard auth in the spec:** paths wrapped with `requireAPIKey` in `routes.go` are **Pi / bridge** calls using header **`X-API-Key`** (same secret as `PI_API_KEY` in `.env`). `GET /farms/{id}/devices` uses **`requireJWTOrPiEdge`**: OpenAPI lists **both** `bearerAuth` and `apiKeyAuth` so operators know the Pi may poll device `config` (including `pending_command`) with the API key while the dashboard uses a JWT.

## Security notes

Bootstrap keeps **secrets and TLS** in your hands: the script does not generate passwords or certificates. Use real secrets in production; do not commit `.env`.
