# gr33n-api — Local Development Setup

**New here?** After cloning, run **`./scripts/setup-first-clone.sh`** (or **`make first-clone`**) from the repo root — it prepares env files, installs UI dependencies, and applies schema/migrations when your database is ready. Use **`./scripts/setup-first-clone.sh --docker`** if you prefer Docker Compose for Postgres. **What's in the box:** [`docs/current-state.md`](docs/current-state.md). **Readable walkthrough:** [`docs/first-session-after-clone.md`](docs/first-session-after-clone.md). Full happy-path narrative: [`docs/local-operator-bootstrap.md`](docs/local-operator-bootstrap.md). **Schema source of truth** (not informal diagrams): [`docs/database-schema-overview.md`](docs/database-schema-overview.md).

## Prerequisites

| Tool | Version | Install |
|------|---------|---------|
| Go | 1.25+ | https://go.dev/dl/ or `snap install go --classic` |
| PostgreSQL | 15+ | `sudo apt install postgresql` (schema uses `UNIQUE NULLS NOT DISTINCT`, added in Postgres 15) |
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
4. **`make dev-auth-test`** — API + UI with production-like auth.
5. Log in on the UI's Login page with the seeded demo account: **`dev@gr33n.local` / `devpassword`** — owns **gr33n Demo Farm** (farm 1), so you see real crops/animals/aquaponics data immediately. No extra setup needed; skip step 6 unless you specifically want the separate env-admin path.
6. *Optional* — env-admin password file (login **`admin`**, same underlying user via `ADMIN_BIND_USER_ID`): **`echo -n 'password' | go run scripts/gen-admin-hash.go > ~/.gr33n/admin.hash`**

**After running `make test`** against a persistent local DB, smoke tests can leave thousands of alerts and automation runs on farm 1, plus hundreds of smoke orgs/users in Settings. Reset to clean demo data without wiping Docker volumes:

```bash
./scripts/dev-reset-farm.sh --farm-id 1 --profile demo_showcase
```

That purges smoke pollution (automation runs, alerts, smoke orgs/users), re-applies `master_seed.sql`, and restores the showcase profile.

For a **full wipe** (empty RAG chunks, one farm, one user — best before UI demos):

```bash
make dev-stack-fresh
./scripts/dev-reset-farm.sh --farm-id 1 --profile demo_showcase
# optional embeddings (needs Ollama + EMBEDDING_API_KEY):
./scripts/rag-ingest-demo.sh && ./scripts/rag-ingest-platform-docs.sh
```

Then **Settings → Field memories (RAG corpus) → Re-ingest → Operational** indexes live farm rows (field guides + platform docs are separate scripts or the Re-ingest buttons).

**Start the API** with the dev-tagged binary when `.env` has `AUTH_MODE=auth_test`:

```bash
make run-auth-test   # not bare `go run ./cmd/api/` — that exits immediately
make dev-auth-test   # API + UI together
```

Restart the API after `git pull` so new routes and solar/weather handlers register.

**Weather on Today:** set `WEATHER_PROVIDER=openmeteo` in `.env`, restart API, then **Settings → Farm site** (near the top of the page) → check **Use live weather forecast** and pick **°F** or **°C** for the forecast line.

_operator narrative and troubleshooting:_ **`docs/local-operator-bootstrap.md`**. **Readable `.env` mirror:** [`docs/example-env.md`](docs/example-env.md).

---

### Raspberry Pi OS (edge daemon or experimental full stack)

- **Edge-only Pi** (sensors/actuators talking to an API elsewhere): **`./scripts/install-pi-edge-deps.sh`** (`make install-pi-edge-deps`), then **`pi_client/setup.sh`** — see **`docs/raspberry-pi-and-deployment-topology.md`**. After wiring sensors and actuators in the dashboard, use **Wiring → Virtual Pi → Download config.yaml** (or **`GET /devices/{id}/pi-config`**) instead of hand-editing pins and channels.
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

**CI / staging parity:** GitHub Actions uses the same Compose **`db`** image and **`bootstrap-local.sh`** path as local dev; hosted environments must still provide **Timescale + PostGIS + pgvector** where applicable — see **[docs/rag-ci-and-staging-parity.md](../docs/rag-ci-and-staging-parity.md)**.

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
| `AI_ENABLED` | `POST /farms/{id}/rag/answer`, `POST /v1/chat` | When **`false`**, runs **Lite mode**: no synthesis and chat returns **503**. When **unset**, defaults **on** (backward compatible). **`GET /capabilities`** exposes `ai_enabled`. |
| `EMBEDDING_API_KEY` | `GET/POST /farms/{id}/rag/search` and `/rag/answer` | OpenAI-compatible `/v1/embeddings` (see also `EMBEDDING_BASE_URL`, `EMBEDDING_MODEL`) |
| `LLM_BASE_URL` | `POST /farms/{id}/rag/answer` | OpenAI-compatible base URL, e.g. `https://api.openai.com/v1` or `http://127.0.0.1:1234/v1` (LM Studio) |
| `LLM_MODEL` | Answer synthesis | Chat model id (required with `LLM_BASE_URL` for answers) |
| `LLM_API_KEY` | Answer synthesis | Set if the chat server requires `Authorization: Bearer`; many local servers need no key |
| `LLM_TEMPERATURE` | Answer synthesis | Default `0.2` |
| `LLM_MAX_TOKENS` | Answer synthesis | Default `1024` |
| `LLM_TIMEOUT_SECONDS` | Answer synthesis | Chat HTTP client timeout; default **666** |
| `LLM_RETRY_MAX_ATTEMPTS` | Answer synthesis + `/v1/chat` | Total tries including the first attempt; default **3**, clamped **1..8**. **1** disables retry. Retries HTTP 408/425/429/5xx, per-attempt timeout, network errors. Never retries caller cancellation or other 4xx. |
| `LLM_RETRY_BACKOFF_MS` | Answer synthesis + `/v1/chat` | Initial backoff in ms; default **500**, clamped **10..30000**. Exponential doubling up to 10s with ±25% jitter. |
| `LLM_VISION_MODEL` | `/v1/chat` with `attachment_ids` | Optional multimodal model (e.g. `llava`). When set, `GET /capabilities` exposes `vision_chat_enabled`. Uses `LLM_VISION_BASE_URL` or `LLM_BASE_URL`. |
| `LLM_VISION_BASE_URL` | Vision chat | Optional; defaults to `LLM_BASE_URL`. |
| `LLM_VISION_API_KEY` | Vision chat | Optional; defaults to `LLM_API_KEY`. |
| `CHAT_SESSION_TTL_DAYS` | `/v1/chat` history retention | Sessions whose newest turn is older than this are pruned by a background loop. Default **30**, clamped **0..3650**. **0** disables pruning (history grows forever). |
| `CHAT_SESSION_PRUNE_INTERVAL_HOURS` | Prune loop cadence | How often the loop runs after the startup delay. Default **24**, clamped **1..168**. |
| `CHAT_SESSION_PRUNE_STARTUP_DELAY_SECONDS` | Prune loop startup | Delay before the first prune at API boot. Default **30**, clamped **0..600**. Lets RAG ingest / automation worker settle before the loop touches the chat tables. |
| `CHAT_COST_WINDOW_HOURS` | `/v1/chat` cost guards | Rolling-window length for token accumulation. Default **24** when cost guard enabled (Phase 113), else **1**. Clamped **1..168**. |
| `CHAT_COST_MAX_TOKENS_PER_USER` | `/v1/chat` cost guards | Total prompt+completion tokens a single user may burn within the window before **429**. Default **200000** when cost guard enabled (Phase 113). Set **`GUARDIAN_COST_GUARD=off`** to disable caps. |
| `CHAT_COST_MAX_TOKENS_PER_FARM` | `/v1/chat` cost guards | Same as above but per `farm_id`. Default **0** (disabled). |
| `GUARDIAN_COST_GUARD` | `/v1/chat` cost guards | **`off`** disables token caps. When unset: **on** in production, **off** in dev/auth_test. |

### Phase 113 — auth hardening

| Variable | Description |
|----------|-------------|
| `REGISTRATION_MODE` | **`open`**, **`invite`** (default in production), or **`closed`**. dev/auth_test default **open**. Invite codes: **`POST /auth/invites`** (JWT) then register with **`invite_code`**. |
| `AUTH_LOGIN_MAX_PER_MINUTE` | Brute-force cap on **`POST /auth/login`** (default **10**/min per IP+username). |
| `PI_LEGACY_KEY_DISABLED` | Set **`true`** after migrating Pi clients to per-device keys (disables shared **`PI_API_KEY`** auth). |
| `SECURITY_HSTS_ENABLED` | Set **`true`** when this process terminates TLS and should emit **Strict-Transport-Security**. |

Per-device Pi keys: mint from device settings in the UI; see **`SECURITY.md`**. Legacy **`PI_API_KEY`** logs a startup deprecation warning until disabled.


**Operator-facing usage dashboard** (Phase 28 WS5): when `AI_ENABLED=true`, **`GET /v1/chat/usage`** (JWT) returns the caller's rolling-window totals + remaining budget. Pass `?farm_id=N` to additionally include per-farm totals (farm-member auth required). The Settings → **Guardian usage** card calls this endpoint and renders progress bars. Crossing **80 %** of the per-user cap fires a one-shot `chat_budget_warning` alert into `gr33ncore.alerts_notifications` so the existing alerts UI surfaces the warning without operators having to poll the endpoint.
| `RAG_SYNTHESIS_MAX_PER_MINUTE` | Answer endpoint | Default `30` requests/minute per API process (rolling minute window). |
| `RAG_SYNTHESIS_MAX_PER_MINUTE_PER_FARM` | Answer endpoint | Optional. When set to an integer **>0**, each `farm_id` gets its own cap per minute (replaces the global-only limiter). Use on shared deployments for fairness. |

When **`AI_ENABLED=true`** and both **`LLM_BASE_URL`** and **`LLM_MODEL`** are set, `cmd/api` **verifies** the backend with `GET {LLM_BASE_URL}/models` at startup and **exits** if that check fails (see Phase 27 plan).

Full on-prem **Ollama** setup — install, systemd override, model pull, intranet DNS — is in **[docs/farm-guardian-ollama-setup.md](docs/farm-guardian-ollama-setup.md)** (Phase 27 WS1). For the **request-flow** explainer (UI → handler → RAG → snapshot → LLM → SSE → persistence) and the cost-guard reasoning, see **[docs/farm-guardian-architecture.md](docs/farm-guardian-architecture.md)**. **Hardware sizing** (will chat feel laggy? GPU/RAM for DB vs Ollama): **[docs/recommended-hardware-and-sizing.md](docs/recommended-hardware-and-sizing.md)**. **Phases 129–139** (awakening through turn debugger): **[docs/plans/archive/phase_129_139_guardian_next_level_roadmap.plan.md](docs/plans/archive/phase_129_139_guardian_next_level_roadmap.plan.md)** · **[closure checklist](docs/plans/archive/phase-129-139-closure.md)** · optional nightly QA: **[docs/ci-guardian-qa.md](docs/ci-guardian-qa.md)**.

### Optional: observability (sit-in logging)

| Variable | Used for | Notes |
|----------|----------|--------|
| `LOG_FORMAT` | `cmd/api` access + automation logs | Set to `json` for **JSON** log lines (default is **text** `key=value` from `log/slog`). |
| `AUTH_DEBUG_LOG` | Auth middleware | Set to `true` to log **`auth_rejected`** with a **reason** code when login fails (missing bearer, bad JWT, bad API key). Never logs token values. |

Full capture, Docker/systemd rotation, **optional Loki stack (`docker-compose.logging.yml`)**, aggregation, and archival patterns: **[docs/operator-logging-runbook.md](../docs/operator-logging-runbook.md)** (Phase 26 WS2).

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

### Go tests and the `@hardware` lane (Phase 33 WS4)

```bash
# Default suite (includes cmd/api smoke tests; auth-bypass via the dev tag).
# GPIO / live-hardware tests are NOT compiled here.
make test            # == go test -tags dev ./... -v -count=1

# Phase 99 — UI enum lists must match backend/OpenAPI (growth stages, lighting presets).
make check-ui-domain-parity
```

Live GPIO / edge-hardware tests live behind the **`hardware` build tag** and are
excluded from `make test` and CI by default. Run them only on a Raspberry Pi /
relay bench with the API running, `PI_API_KEY` set, and the seeded
`demo-veg-relay-01` actuator:

```bash
GR33N_HARDWARE_TEST=1 go test -tags 'dev hardware' -run Hardware ./cmd/api/ -count=1 -v
```

This drives the real `pending_command → pi_client → actuator_events → clear`
round-trip via [`scripts/run-edge-actuator-smoke.sh`](scripts/run-edge-actuator-smoke.sh).
In CI it is a **manual** `hardware-smoke` job (`workflow_dispatch`, self-hosted
`gr33n-hardware` runner) — it never runs on push/PR.

### Ollama E2E smokes (Phase 112)

**Database first:** smokes need a migrated Postgres with `auth.users` and demo seed.
If you see `relation "auth.users" does not exist`, run once:

```bash
./scripts/bootstrap-local.sh --seed
# Docker Compose instead: make dev-stack
```

Ensure `.env` has `DATABASE_URL` pointing at that database (same as `make dev-auth-test`).

**Stop `make dev-auth-test` before smokes** — the dev API and Ollama both consume RAM;
on a 16 GB laptop you may see Ollama `502` / “requires more system memory” if both run.

Set **`LLM_MODEL=tinyllama`** in `.env` for Phase 118 guardrail smokes (not `llama3.1:8b`).

Guardian model-selector E2E tests (`TestPhase112_*`) compile only with the
`ollama` build tag and expect a running Ollama at `LLM_BASE_URL`:

```bash
# Local (Ollama running, models pulled)
export AI_ENABLED=true
export LLM_BASE_URL=http://127.0.0.1:11434/v1
export LLM_MODEL=tinyllama
ollama pull tinyllama
ollama pull phi3:mini   # context-window guardrail test

go test -tags 'dev ollama' ./cmd/api/ -run TestPhase112 -count=1 -v
```

Phase 118 adds `TestPhase118_*` (capabilities filter, tag normalization guardrail,
runtime hints). Makefile shortcuts:

```bash
make ollama-smoke          # Phase 112 + 118 smokes
make ollama-smoke-cpu        # CPU-only box (40m timeout, LLM_MAX_TOKENS=60)
make ollama-smoke-help       # print prerequisites
```

**Model quality eval (Phase 122):** compare grounded answers across installed Ollama models:

```bash
# API + Ollama running; JWT from dashboard login or smoke helper
export GUARDIAN_EVAL_TOKEN="<jwt>"
make guardian-eval                    # all chat-capable models, demo farm id 1
make guardian-eval MODEL=phi3:mini  # one model
```

Report: `data/guardian_model_eval.json`. Scores surface in **Guardian model selector**
(`GET /guardian/models` → `eval` field). Re-run after pulling a new model.

**Context budget (Phase 122 + 126):** prompt trimming uses the model's **effective**
context window (`effective_context_window` on `GET /guardian/models`), not rope-extended
metadata alone. Built-in overrides include `phi3:mini` → 4096 and `tinyllama` → 2048;
override more with `GUARDIAN_EFFECTIVE_CONTEXT_OVERRIDES=phi3:mini=4096`. Models whose
effective window is below 8192 get trimmed history, RAG top-K, and snapshot detail before
the prompt is sent. The grounded-chat *gate* still uses advertised `context_window` (8192
minimum): `phi3:mini` reports rope-extended 131072 via Ollama but runs at 4096 on CPU —
Phase 126 trims grounded prompts accordingly.

**CPU laptop Guardian playbook (Phase 126):** full troubleshooting, RAG bring-up, and what the UI
means by “CPU” → **[docs/guardian-ollama-laptop-playbook.md](docs/guardian-ollama-laptop-playbook.md)**.

Summary:

- `phi3:mini` + **Use farm context** is supported but **slow** on CPU-only boxes (minutes
  per grounded turn). First ungrounded message is a good warm-up while the model loads.
- Before long chat sessions, free RAM: `ollama stop <embed-model>` (embedding and chat
  models compete on one Ollama host). The UI **does not** run this when you switch models.
- Check for stale jobs: `pgrep -a 'ollama run'` (kill any PIDs — they can wedge Ollama for hours).
- Use **tinyllama** for fast ungrounded smoke; **phi3:mini** for grounded quality when you can wait.
- Off-farm horticulture (e.g. cherry tree, forest garden) → turn **farm context off**;
  Guardian still answers from general knowledge.

Or run directly (same as the make targets):

**Model selector notes (Phase 118):** `GET /guardian/models` returns chat-capable
models only; embedding models appear with `?all=true`. The UI shows Ollama
`runtime_hint` (loaded/cold, CPU vs GPU). `LLM_MODEL=tinyllama` resolves to
`tinyllama:latest` for guardrails — grounded chat rejects models below 8192 context.
Extended-rope models like `phi3:mini` report the max `*.context_length` from
Ollama `/api/show` (intentional).

**CPU-only Ollama boxes** (no GPU): the full grounding prompt + tinyllama can exceed
go-test's default 10-minute deadline. Use a longer test timeout and cap generation:

```bash
go test -tags 'dev ollama' ./cmd/api/ -run TestPhase112 -count=1 -v \
  -timeout 40m \
  LLM_TIMEOUT_SECONDS=150 LLM_MAX_TOKENS=60
```

Smoke helpers use a shared 60 s HTTP client (`cmd/api/smoke_helpers_test.go`) so a
wedged handler fails one test instead of hanging the suite. The default LLM client
timeout is 120 s (`internal/rag/llm/chat.go`); override with `LLM_TIMEOUT_SECONDS`.

Optional farm-settings auto-pull:

```bash
export GUARDIAN_OLLAMA_AUTO_PULL=true
export GUARDIAN_OLLAMA_PULL_TIMEOUT_SECONDS=600
```

In CI, run the manual **`ollama-smoke`** job from the Actions tab
(`workflow_dispatch`) — it starts an Ollama service container, pulls
`tinyllama` and `phi3:mini`, then runs the Phase 112 smokes. It never runs on
push/PR.

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

## Scripts & Make targets reference

Quick-reference for what to run and when. Full details: `make help` or read the target comments in the Makefile.

### Daily dev (you will use these)

| Command | When to run |
|---------|-------------|
| `make dev-auth-test` | **Start the stack** — API on `:8080` + UI on `:5173` (Ctrl+C to stop) |
| `make migrate` | After `git pull` brings new migrations |
| `make test` | Run all Go unit + smoke tests |
| `make ollama-smoke-cpu` | Smoke-test Guardian/Ollama integration (stop `dev-auth-test` first) |
| `make lint` | Quick sanity check before committing |
| `scripts/restart-local.sh` | **After a reboot** — starts Postgres, waits, runs `db-sanity-report` |

### One-time setup (already done on this machine)

| Command | What it did |
|---------|-------------|
| `make install-deps-debian` | Installed Postgres, PostGIS, pgvector, TimescaleDB, Node via apt |
| `make first-clone` | `go mod download`, `.env` from template, `npm ci` |
| `scripts/bootstrap-local.sh --seed` | Applied schema + migrations + demo seed data |
| `echo -n 'password' \| go run scripts/gen-admin-hash.go > ~/.gr33n/admin.hash` | Set env-admin login password |

### LLM / Guardian model defaults

`.env` controls the model `make dev-auth-test` uses.
Your laptop `.env` should have `LLM_MODEL=phi3:mini` (2.2 GB, fits 16 GB RAM, supports grounded farm chat).
The production server uses `LLM_MODEL=llama3.1:8b`.
Smokes (`make ollama-smoke-cpu`) always default to `tinyllama` regardless of `.env`.

### You probably won't touch these

| Script / target | What it's for |
|-----------------|---------------|
| `generate-crop-*.sh` | Regenerate crop catalog seed SQL from YAML |
| `rag-ingest-*.sh` | Ingest documents into the vector/RAG store |
| `run-edge-*.sh` / `install-pi-edge-deps.sh` | Raspberry Pi sensor & actuator integration |
| `scripts/enterprise/` | Multi-tenant operator scripts (future) |
| `sit-in-*.sh` | QA session prep & dry-run |
| `make sqlc` | Regenerate Go DB query layer from SQL (after schema changes) |
| `make audit-openapi` / `make audit-env` | Pre-release validation checks |
| `make build` | Compile a production binary (for server deploy) |

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
