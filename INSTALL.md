# gr33n-api — Local Development Setup

**New here?** After cloning, run **`./scripts/setup-first-clone.sh`** (or **`make first-clone`**) from the repo root — it prepares env files, installs UI dependencies, and applies schema/migrations when your database is ready. Use **`./scripts/setup-first-clone.sh --docker`** if you prefer Docker Compose for Postgres. Full happy-path narrative: [`docs/local-operator-bootstrap.md`](docs/local-operator-bootstrap.md). **Schema source of truth** (not informal diagrams): [`docs/database-schema-overview.md`](docs/database-schema-overview.md).

## Prerequisites

| Tool | Version | Install |
|------|---------|---------|
| Go | 1.25+ | https://go.dev/dl/ or `snap install go --classic` |
| PostgreSQL | 14+ | `sudo apt install postgresql` |
| PostGIS | 3.x (match Postgres) | `sudo apt install postgresql-14-postgis-3` (version as needed) |
| TimescaleDB | 2.x | https://docs.timescale.com/self-hosted/latest/install/ |
| pgvector | Match Postgres major | Required for Phase 24 RAG (`CREATE EXTENSION vector`). Install per [pgvector](https://github.com/pgvector/pgvector#installation), or use the repo `docker compose` database image (`db/Dockerfile` builds pgvector). |
| sqlc | latest | `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest` |
| Node.js (UI) | 22+ | https://nodejs.org/ or your OS package manager |

### Debian / Ubuntu (automated)

On **Linux** with **apt** (Debian, Ubuntu, Mint, Pop!\_OS, …), you can install **PostgreSQL 16** (official PGDG apt), **PostGIS**, **pgvector**, **TimescaleDB**, and **Node.js 22** with sudo — you will be prompted for your password:

```bash
./scripts/install-system-deps-debian.sh
# or: make install-deps-debian
```

This adds the PostgreSQL PGDG and TimescaleDB apt repositories, then installs packages. It does **not** install **Go** (distro packages are often too old for `go 1.25` in `go.mod`); install Go from [go.dev/dl](https://go.dev/dl/) or snap, then `go install … sqlc` as above.

To run that script **and** the first-clone bootstrap in one flow:

```bash
./scripts/setup-first-clone.sh --install-system-deps
# or: make first-clone-install-deps
```

Use **`./scripts/install-system-deps-debian.sh --skip-node`** if you already manage Node with nvm/fnm.

---

## Docker Compose DB + `AUTH_TEST` + demo seed (laptop / QA parity)

Typical flow when you want **Timescale + PostGIS + pgvector** without a native Postgres install:

1. From the repo root: **`sg docker -c 'docker compose up -d db'`** — Postgres is published on the host at **`127.0.0.1:5433`** (see `docker-compose.yml`).
2. Copy **`.env.example` → `.env`**. Set **`DATABASE_URL=postgres://gr33n:gr33n@127.0.0.1:5433/gr33n?sslmode=disable`**, **`AUTH_MODE=auth_test`**, **`JWT_SECRET`**, **`PI_API_KEY`**, and optional **`ADMIN_BIND_USER_ID` / `ADMIN_BIND_EMAIL`** (env-admin JWT needs a real `user_id` for farm routes — defaults match `master_seed.sql`).
3. **`./scripts/bootstrap-local.sh --seed`** — applies schema, migrations, and **`db/seeds/master_seed.sql`**.
4. Env-admin password file (login **`admin`**): **`echo -n 'password' | go run scripts/gen-admin-hash.go > ~/.gr33n/admin.hash`**
5. **`make dev-auth-test`** — API + UI with production-like auth.

_operator narrative and troubleshooting:_ **`docs/local-operator-bootstrap.md`**. **Readable `.env` mirror:** [`docs/example-env.md`](docs/example-env.md).

---

### Raspberry Pi OS (edge daemon or experimental full stack)

- **Edge-only Pi** (sensors/actuators talking to an API elsewhere): **`./scripts/install-pi-edge-deps.sh`** (`make install-pi-edge-deps`), then **`pi_client/setup.sh`** — see **`docs/raspberry-pi-and-deployment-topology.md`**.
- **Docker on the Pi** (for `docker compose` experiments): **`./scripts/install-pi-edge-deps.sh --with-docker`** or **`make install-pi-edge-deps-docker`**.
- Pi OS is Debian-derived; **do not** run `install-system-deps-debian.sh` on a small Pi unless you intend to host Postgres locally — see the topology doc for RAM/storage warnings.

---

## 1. Clone the repo

```bash
git clone https://github.com/YOUR_ORG/gr33n.git
cd gr33n
```

---

## 2. PostgreSQL setup

### 2a. Create the database

```bash
sudo -u postgres psql -c "CREATE DATABASE gr33n;"
```

### 2b. Enable TimescaleDB on the database

```bash
sudo -u postgres psql -d gr33n -c "CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;"
```

### 2c. Enable pgvector (Phase 24 RAG)

The bundled schema enables `vector` for `gr33ncore.rag_embedding_chunks`. Install the pgvector package for your Postgres version first, then:

```bash
sudo -u postgres psql -d gr33n -c "CREATE EXTENSION IF NOT EXISTS vector;"
```

If Postgres reports that the extension is not available, follow [pgvector installation](https://github.com/pgvector/pgvector#installation) for your OS, or run Postgres via **`docker compose`** in this repo (the `db` service builds TimescaleDB + pgvector).

### 2d. Create a local dev user matching your Linux username

PostgreSQL on Linux uses **peer authentication** by default — the connecting
OS user must match a PostgreSQL role of the same name.

```bash
sudo -u postgres psql -c "CREATE USER $USER WITH SUPERUSER;"
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE gr33n TO $USER;"
```

Verify it works (no password, no sudo needed):

```bash
psql -d gr33n -c "SELECT current_user, current_database();"
# Expected:  current_user | current_database
#            davidg       | gr33n
```

---

## 3. Apply database schema

For a **new** database, load the full schema (includes `CREATE EXTENSION` for PostGIS, TimescaleDB, and **vector** — those packages must be installed on the server):

```bash
psql -d gr33n -v ON_ERROR_STOP=1 -f db/schema/gr33n-schema-v2-FINAL.sql
```

**Upgrading** an older database that was created from an earlier snapshot: apply SQL files under `db/migrations/` in **lexicographic (filename) order**:

```bash
for f in $(printf '%s\n' db/migrations/*.sql | LC_ALL=C sort); do
  echo "==> $f"
  psql -d gr33n -v ON_ERROR_STOP=1 -f "$f"
done
```

Or run `./scripts/bootstrap-local.sh` from the repo root (schema + sorted migrations + optional `--seed`); see [`docs/local-operator-bootstrap.md`](docs/local-operator-bootstrap.md).

---

## 4. Environment variables

The API reads one required env var at startup:

| Variable | Description | Default (dev) |
|----------|-------------|---------------|
| `DATABASE_URL` | PostgreSQL connection string | see below |
| `PORT` | HTTP listen port | `8080` |

For local development with peer auth (no password):

```bash
export DATABASE_URL="postgres://$USER@/gr33n?host=/var/run/postgresql"
```

Add this to your `~/.bashrc` or `~/.zshrc` to avoid typing it every time.

### Optional: RAG search and answer synthesis (Phase 24)

| Variable | Used for | Notes |
|----------|----------|--------|
| `EMBEDDING_API_KEY` | `GET/POST /farms/{id}/rag/search` and `/rag/answer` | OpenAI-compatible `/v1/embeddings` (see also `EMBEDDING_BASE_URL`, `EMBEDDING_MODEL`) |
| `LLM_BASE_URL` | `POST /farms/{id}/rag/answer` | OpenAI-compatible base URL, e.g. `https://api.openai.com/v1` or `http://127.0.0.1:1234/v1` (LM Studio) |
| `LLM_MODEL` | Answer synthesis | Chat model id (required with `LLM_BASE_URL` for answers) |
| `LLM_API_KEY` | Answer synthesis | Set if the chat server requires `Authorization: Bearer`; many local servers need no key |
| `LLM_TEMPERATURE` | Answer synthesis | Default `0.2` |
| `LLM_MAX_TOKENS` | Answer synthesis | Default `1024` |
| `RAG_SYNTHESIS_MAX_PER_MINUTE` | Answer endpoint | Default `30` (per API process) |

### Optional: observability (sit-in logging)

| Variable | Used for | Notes |
|----------|----------|--------|
| `LOG_FORMAT` | `cmd/api` access + automation logs | Set to `json` for **JSON** log lines (default is **text** `key=value` from `log/slog`). |
| `AUTH_DEBUG_LOG` | Auth middleware | Set to `true` to log **`auth_rejected`** with a **reason** code when login fails (missing bearer, bad JWT, bad API key). Never logs token values. |

---

## 5. Build and run

```bash
go mod tidy
go run ./cmd/api/
```

Expected output:

```
2026/02/26 16:41:55 ✅ Connected to gr33n database
2026/02/26 16:41:55 🌱 gr33n API running on http://localhost:8080
```

---

## 6. Smoke test

```bash
# Health check
curl http://localhost:8080/health
# → {"service":"gr33n-api","status":"ok"}

# All units of measure
curl http://localhost:8080/units

# Units filtered by type
curl "http://localhost:8080/units?type=temperature"

# Devices
curl http://localhost:8080/devices
```

---

## 7. Code generation (sqlc)

If you modify any `.sql` query files under `internal/db/`, regenerate the
Go query layer:

```bash
sqlc generate
```

Generated files live in `internal/db/` — do **not** edit them by hand.

---

## Common issues

### `could not connect to database after 5 attempts`

The error message will now print the real cause on each attempt.
Most common root causes:

- **Peer auth mismatch** — your Linux username has no matching PostgreSQL role.
  Fix: run step 2c above.
- **Socket path wrong** — make sure `?host=/var/run/postgresql` is in the URL
  (not `localhost:5432`, which forces TCP and fails peer auth).
- **PostgreSQL not running** — `sudo systemctl start postgresql`

### `package gr33n-api/internal/platform/commontypes is not in std`

The `enums.go` file is missing. Copy it into place:

```bash
mkdir -p ~/gr33n-api/internal/platform/commontypes
cp ~/Downloads/enums.go ~/gr33n-api/internal/platform/commontypes/
go mod tidy
```

### `could not change directory … Permission denied` (sudo -u postgres)

Harmless warning — postgres can't `cd` into your home dir when you run `sudo`
from inside it. The command itself still executes correctly.

---

## Repository layout

```
gr33n-api/
├── cmd/
│   └── api/
│       ├── main.go          # Entry point, DB connection, server startup
│       └── routes.go        # HTTP route registration
├── internal/
│   ├── db/                  # sqlc-generated query layer (do not edit)
│   ├── handlers/            # HTTP handler functions
│   └── platform/
│       └── commontypes/
│           └── enums.go     # Shared enum types used by sqlc
├── db/
│   ├── migrations/          # Incremental SQL migrations (apply in filename order on upgrades)
│   └── schema/              # Full schema snapshot (greenfield installs)
├── sqlc.yaml
├── go.mod
└── go.sum
```
