# gr33n 🌱

An open-source agricultural operating system designed to reclaim data, land, and autonomy.

[![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)
[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev)
[![Vue](https://img.shields.io/badge/Vue-3-4FC08D?logo=vue.js)](https://vuejs.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-14+-336791?logo=postgresql)](https://postgresql.org)

**Current focus:** **Pi and edge hookup** — the API, worker, and [`pi_client/gr33n_client.py`](pi_client/gr33n_client.py) now agree on `pending_command` JSON (`command`, `schedule_id`, optional `rule_id` / `program_id`), base64-encoded `devices.config` on `GET /farms/{id}/devices`, and **`POST /actuators/{id}/events`** provenance (`triggered_by_schedule_id`, `triggered_by_rule_id`, `program_id` → `meta_data`). Automated smoke covers the contract; **“next on the wall”** means **you** run the real Pi (or bridge) against **your** staging/prod API: set `PI_API_KEY` and `base_url` in `config.yaml`, map each `device_id` ↔ `actuator_id`, let sensors post readings and the schedule loop poll devices, then confirm in Postgres or the UI that readings and **`actuator_events`** look right after a schedule or program fires. That is **field validation**, not more repo code. **Phase 22** is done on `main`; **Phase 21** (full crop-cycle analytics in [the plan](docs/plans/phase_21_crop_cycle_analytics.plan.md)) is **not** — only **Phase 20.7**’s narrower **`GET /crop-cycles/{id}/cost-summary`** exists today. **RAG** in this repo means **Phase 20.95 “RAG-prep”** (schema/columns so a *future* retrieval stack can read clean rows) — not a deployed chatbot/RAG pipeline. Key playbooks: [`docs/workflow-guide.md`](docs/workflow-guide.md), [`docs/mqtt-edge-operator-playbook.md`](docs/mqtt-edge-operator-playbook.md), [`docs/insert-commons-pipeline-runbook.md`](docs/insert-commons-pipeline-runbook.md), [`docs/insert-commons-receiver-playbook.md`](docs/insert-commons-receiver-playbook.md), [`docs/notifications-operator-playbook.md`](docs/notifications-operator-playbook.md), [`docs/domain-modules-operator-playbook.md`](docs/domain-modules-operator-playbook.md), [`docs/mobile-distribution.md`](docs/mobile-distribution.md), [`docs/audit-events-operator-playbook.md`](docs/audit-events-operator-playbook.md), [`docs/terminology-guideline.md`](docs/terminology-guideline.md), [`docs/phase-13-operator-documentation.md`](docs/phase-13-operator-documentation.md), [`docs/phase-14-operator-documentation.md`](docs/phase-14-operator-documentation.md).

---

## What Is gr33n?

gr33n is a modular, scalable, and decentralized farm management system built for real humans — not cloud landlords. Whether you're managing a homestead on solar or automating thousands of acres, gr33n adapts to your size, ethics, and bandwidth.

It's PostgreSQL schemas + Go APIs + Vue dashboards + Raspberry Pi clients + shared insert statements.

But more than that:
it's a political stance in schema form.

---

## Why gr33n Exists

> "If your DNA, soil, labor, and climate data feed trillion-dollar industries — and you're not seeing a dime — that's not tech, that's extraction."

This project exists because:
- Big Ag is closing the loop on food systems, and we're cracking it back open.
- Data rights matter — even your soil and sunlight deserve consent.
- Billionaires shouldn't profit off your greenhouse or genome without giving back.
- Farmers, tinkerers, and off-gridders deserve tools that don't call home.

### 🔌 What Does "Don't Call Home" Mean?

gr33n will never require a permanent internet connection, forced login, or hidden check-in with third-party servers. Whether you're on an island, a mountaintop, or a mesh-netted greenhouse, gr33n works where you live, without compromise.

---

## Core Principles

- **Modularity** — Each ag domain (crops, animals, natural-farming inputs, IoT sensors) lives in its own schema. Use what you need, prune the rest. Enable modules per-farm via `gr33ncore.farm_active_modules`.

- **Connectivity Optional** — Works offline, intranet-only, or online. Supports Supabase or bare-metal Postgres with TimescaleDB/PostGIS.

- **Automation-Ready** — Schedule tasks, trigger actuators, run AI models — or run it all manually. Your tech, your tempo.

- **Insert Commons (farm-side sender)** — Per-farm opt-in in Settings; `POST /farms/{id}/insert-commons/sync` builds **coarse, pseudonymous aggregates** and optionally POSTs them to `INSERT_COMMONS_INGEST_URL` with optional `Authorization: Bearer <INSERT_COMMONS_SHARED_SECRET>`. Sync attempts are persisted (`GET /farms/{id}/insert-commons/sync-events`) with **idempotency keys**, **rate limits**, and **server-side backoff** after repeated delivery failures. A separate **farm audit trail** records sensitive actions (membership, opt-in, sync attempts, finance COA changes, cost exports, receipt access, and more) for owner/manager review via `GET /farms/{id}/audit-events` (see [`docs/audit-events-operator-playbook.md`](docs/audit-events-operator-playbook.md)). For self-hosted pilots, an optional **receiver** process (`cmd/insert-commons-receiver`, `make run-receiver`) validates payloads, enforces the shared secret, dedupes on payload hash, and stores rows in Postgres — see [`docs/insert-commons-receiver-playbook.md`](docs/insert-commons-receiver-playbook.md) and migration `db/migrations/20260417_phase13_insert_commons_receiver.sql`. Apply `db/migrations/20260415_phase11_rbac_receipts_commons.sql` and `db/migrations/20260416_phase12_insert_commons_federation.sql` on existing databases. **Custom clients** POSTing ingest JSON themselves must use the **exact** documented shape (only six top-level keys, complete `aggregates` children, boolean `includes_pii`) or validation returns **400** — see [`docs/insert-commons-pipeline-runbook.md`](docs/insert-commons-pipeline-runbook.md) (*Custom senders*).

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| API | Go 1.23 · `net/http` stdlib |
| Database | PostgreSQL 14+ · TimescaleDB · PostGIS |
| Query layer | sqlc (generated — do not edit `internal/db/`) |
| Frontend | Vue 3 · Vite · Pinia · Tailwind CSS |
| Pi client | Python 3 · RPi.GPIO / smbus2 |
| Auth | Supabase (hosted) / local peer auth (dev) |
| Schema | Multi-schema PostgreSQL — `gr33ncore` + `gr33nnaturalfarming` |

---

## Repository Layout

```
gr33n-api/
├── cmd/api/
│   ├── main.go              # Entry point, DB pool, server startup
│   ├── routes.go            # All HTTP route registrations
│   └── cors.go              # CORS middleware
├── cmd/insert-commons-receiver/
│   └── main.go              # Optional pilot ingest service for Insert Commons (`POST /v1/ingest`, `GET /v1/stats`)
├── internal/
│   ├── db/                  # sqlc-generated query layer (DO NOT EDIT)
│   ├── handler/
│   │   ├── farm/            # GET /farms/:id
│   │   ├── zone/            # Zones CRUD
│   │   ├── sensor/          # Sensors CRUD + readings endpoints
│   │   ├── device/          # Devices CRUD + status toggle
│   │   └── task/            # Tasks list + status update
│   ├── httputil/            # WriteJSON / WriteError helpers
│   ├── insertcommonsreceiver/ # Optional Insert Commons ingest HTTP handler
│   └── platform/
│       └── commontypes/     # Shared enum types for sqlc
├── db/
│   ├── schema/
│   │   └── gr33n-schema-v2-FINAL.sql   # Full PostgreSQL schema (source of truth)
│   ├── seeds/
│   │   └── master_seed.sql             # Demo farm: natural-farming inventory + JADAM-style inputs (v1.005)
│   └── queries/             # sqlc SQL source files
├── ui/                      # Vue 3 frontend
│   └── src/
│       ├── views/           # Dashboard, Zones, Sensors, Actuators, Schedules, Inventory
│       ├── stores/farm.js   # Pinia store — all API state
│       ├── api/index.js     # Axios instance → localhost:8080
│       └── router/index.js  # Vue Router
├── pi_client/
│   ├── gr33n_client.py      # Sensor daemon — reads GPIO, POSTs readings to API
│   ├── config.yaml          # Per-node hardware mapping
│   ├── gr33n.service        # systemd unit for autostart
│   └── setup.sh             # One-time Pi bootstrap
├── sqlc.yaml
├── go.mod / go.sum
├── openapi.yaml             # Full API spec (paste into editor.swagger.io for live UI)
├── INSTALL.md
├── ARCHITECTURE.md
└── SECURITY.md
```

---

## Quick Start

**Guided one-path setup (DB → env → UI deps):** [docs/local-operator-bootstrap.md](docs/local-operator-bootstrap.md) — run `./scripts/bootstrap-local.sh` or `make bootstrap-local`.

Full setup in [INSTALL.md](INSTALL.md). Short version:

```bash
# 1. Clone
git clone https://github.com/dgang0404/gr33n.git
cd gr33n-api

# 2. Create and migrate the database
sudo -u postgres psql -c "CREATE DATABASE gr33n;"
psql -d gr33n -f db/schema/gr33n-schema-v2-FINAL.sql

# 3. Seed demo data (natural farming + JADAM-style starter labels)
psql -d gr33n -f db/seeds/master_seed.sql

# 4. Run the API (from repo root)
cp .env.example .env   # once: edit .env with DATABASE_URL, JWT_SECRET, PI_API_KEY if using auth
# Or only: export DATABASE_URL="postgres://$(whoami)@/gr33n?host=/var/run/postgresql"
go run -tags dev ./cmd/api/

# 5. Run the frontend (separate terminal)
cd ui && npm install && npm run dev
```

API → `http://localhost:8080`
UI  → `http://localhost:5173`

Receipt storage defaults to local disk for development:

- `FILE_STORAGE_BACKEND=local`
- `FILE_STORAGE_DIR=./data/files`

Production deployments can switch receipts to S3-compatible object storage by setting:

- `FILE_STORAGE_BACKEND=s3`
- `S3_BUCKET=<bucket>`
- `S3_REGION=<region>`
- optional: `S3_ENDPOINT=<custom endpoint>` for MinIO / R2 / other S3-compatible providers
- optional: `S3_PREFIX=<key prefix>`
- optional: `S3_ACCESS_KEY_ID` and `S3_SECRET_ACCESS_KEY`
- optional: `S3_USE_PATH_STYLE=true`
- optional: `S3_DISABLE_HTTPS=true` for local/test endpoints only
- optional: `FILE_STORAGE_SIGNED_URL_TTL_SECONDS=300` for short-lived receipt download links

To backfill existing blobs from an old local `FILE_STORAGE_DIR` into the configured target backend before cutover:

```bash
# 1. Keep DATABASE_URL pointed at the live DB
# 2. Point the target backend env vars at the new storage location
# 3. Run a dry run first
go run ./cmd/filebackfill --source-dir /path/to/old/files --dry-run

# 4. Then copy all attachments (or only receipts)
go run ./cmd/filebackfill --source-dir /path/to/old/files
go run ./cmd/filebackfill --source-dir /path/to/old/files --file-type cost_receipt
```

The backfill preserves each attachment's existing `storage_path`, so DB rows do not change. After the copy is complete, switch the API to the new `FILE_STORAGE_BACKEND` and verify a few receipt downloads before removing the old local storage.

For operator guidance on receipt storage cutover plus DB/blob backup and restore, see `docs/receipt-storage-cutover-runbook.md`.

### PWA install + offline task writes

Phase 12 adds an offline write queue for the Tasks workflow (`create task` and `advance status`):

- when offline (or on retryable network failure), task writes are queued locally
- queued items are marked in the Tasks UI
- each queued item can be retried or discarded
- queued writes auto-sync on reconnect, and manual `Sync now` is available
- non-retryable server failures are shown as stale/conflict items for operator review

Install/offline notes:

- install the app from your browser for field use (PWA)
- keep one online sync checkpoint before long offline sessions
- after reconnect, verify queued writes drained before ending a shift

For **Play Store / App Store / MDM** distribution without replacing the PWA, use the optional Capacitor scaffold in `ui/` (`npm run build:cap`, `cap:sync`, platform add/open). See [`docs/mobile-distribution.md`](docs/mobile-distribution.md).

---

## API Endpoints

Base URL: `http://localhost:8080` — authoritative request/response schemas in [openapi.yaml](openapi.yaml). Path placeholders use `:id`, `:rid`, `:uid`, `:iid` for readability (the server matches the same paths with `{id}` style).

### Public

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | API + DB health check |
| POST | `/auth/login` | Authenticate & receive JWT |
| POST | `/auth/register` | Register a new account or set password for an **invited** user (existing email with no password yet) |
| GET | `/auth/mode` | Current auth mode (dev / production / auth_test) |

### Pi routes (API key)

Header: `X-API-Key: <PI_API_KEY>` (see env configuration for the API process).

| Method | Path | Description |
|--------|------|-------------|
| POST | `/sensors/:id/readings` | Pi posts a sensor reading |
| PATCH | `/devices/:id/status` | Pi heartbeat / status update |
| POST | `/actuators/:id/events` | Pi reports executed command |
| DELETE | `/devices/:id/pending-command` | Pi clears pending command after execution |

### Insert Commons receiver (optional separate process)

Farm API POSTs JSON to `INSERT_COMMONS_INGEST_URL`; this repo’s **pilot receiver** (`go run ./cmd/insert-commons-receiver/` or `make run-receiver`) listens on `INSERT_COMMONS_RECEIVER_LISTEN` (default **`:8765`**) and implements:

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Process liveness |
| GET | `/v1/stats` | Pilot aggregate counts (pseudonyms, daily ingests, retention) — same Bearer auth as ingest |
| POST | `/v1/ingest` | Validate payload, optional `Authorization: Bearer <INSERT_COMMONS_SHARED_SECRET>`, optional `Gr33n-Idempotency-Key` (forwarded from farm sync), persist idempotently |

Details, migration, and retention: [`docs/insert-commons-receiver-playbook.md`](docs/insert-commons-receiver-playbook.md). If you build or forward JSON manually, match the farm API’s ingest schema (no extra top-level fields; full `aggregates`; `privacy.includes_pii` as JSON boolean); `GET /farms/:id/insert-commons/preview` returns a valid example body — full rules in [`docs/insert-commons-pipeline-runbook.md`](docs/insert-commons-pipeline-runbook.md).

### Dashboard routes (JWT)

Header: `Authorization: Bearer <JWT>` (SSE also supports `?token=` on the stream URL where documented).

**Farm access:** most `/farms/:id/...` routes require the user to be the farm **owner** or a **member** (`gr33ncore.farm_memberships`). **Role caps** apply per area (for example *view* vs *edit* costs, *operate* for field workflows, *admin* for farm settings and membership). Exact checks live in `internal/farmauthz` and in [openapi.yaml](openapi.yaml) per route.

Integration tests in `cmd/api/smoke_test.go` use `AUTH_MODE=auth_test` with a real JWT.

#### Auth, profile, units

| Method | Path | Description |
|--------|------|-------------|
| PATCH | `/auth/password` | Change password (must be logged in) |
| GET | `/profile` | Current user profile |
| PUT | `/profile` | Update current user profile |
| GET | `/units` | List all measurement units |

#### Farms

| Method | Path | Description |
|--------|------|-------------|
| GET | `/farms` | List farms; use `?user_id=<uuid>` to restrict to that user’s farms (recommended for UIs). If omitted, lists **all** farms — use only in trusted operator contexts. |
| POST | `/farms` | Create farm |
| GET | `/farms/:id` | Farm detail (member or owner) |
| PUT | `/farms/:id` | Update farm record (**admin**: owner or manager) |
| DELETE | `/farms/:id` | Soft-delete farm (**admin**) |

#### Farm members (**admin**: owner or manager)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/farms/:id/members` | List members and roles |
| POST | `/farms/:id/members` | Invite or add member (`email`, `role_in_farm`, optional `full_name`) |
| PATCH | `/farms/:id/members/:uid/role` | Change member role (`:uid` = user UUID) |
| DELETE | `/farms/:id/members/:uid` | Remove member from farm |

#### Insert Commons & audit

| Method | Path | Description |
|--------|------|-------------|
| PATCH | `/farms/:id/insert-commons/opt-in` | Toggle Insert Commons aggregate sharing (**admin**) |
| GET | `/farms/:id/insert-commons/preview` | Preview validated ingest JSON only — no sync, no history (**admin**) |
| POST | `/farms/:id/insert-commons/sync` | Build aggregates and POST to `INSERT_COMMONS_INGEST_URL` when set (**admin** or **finance**) |
| GET | `/farms/:id/insert-commons/sync-events` | Paginated sync attempt history (**admin** or **finance** / anyone with cost **view**) |
| GET | `/farms/:id/audit-events` | Sensitive-action audit log (**admin** only; query `limit`, `offset`) |

#### Zones

| Method | Path | Description |
|--------|------|-------------|
| GET | `/farms/:id/zones` | List zones for farm |
| GET | `/zones/:id` | Zone detail |
| POST | `/farms/:id/zones` | Create zone |
| PUT | `/zones/:id` | Update zone |
| DELETE | `/zones/:id` | Delete zone |

#### Devices & actuators

| Method | Path | Description |
|--------|------|-------------|
| GET | `/farms/:id/devices` | List devices |
| GET | `/devices/:id` | Device detail |
| POST | `/farms/:id/devices` | Create device |
| DELETE | `/devices/:id` | Delete device |
| GET | `/farms/:id/actuators` | List actuators for farm |
| PATCH | `/actuators/:id/state` | Update actuator state (dashboard) |
| GET | `/actuators/:id/events` | Actuator event history |

#### Sensors & live stream

| Method | Path | Description |
|--------|------|-------------|
| GET | `/farms/:id/sensors` | List sensors |
| GET | `/farms/:id/sensors/stream` | **SSE** live sensor readings (JWT may be passed as query `token`) |
| GET | `/sensors/:id` | Sensor detail |
| POST | `/farms/:id/sensors` | Create sensor |
| DELETE | `/sensors/:id` | Delete sensor |
| GET | `/sensors/:id/readings/latest` | Latest reading |
| GET | `/sensors/:id/readings` | List readings (`since`, `until`, `limit`, …) |
| GET | `/sensors/:id/readings/stats` | Aggregate stats for a time range |

#### Automation (schedules & runs)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/farms/:id/schedules` | List schedules |
| PATCH | `/schedules/:id/active` | Toggle schedule active |
| GET | `/farms/:id/automation/runs` | List automation runs for farm |
| GET | `/schedules/:id/actuator-events` | Actuator events triggered by schedule |
| GET | `/automation/worker/health` | Automation worker health |

#### Tasks

| Method | Path | Description |
|--------|------|-------------|
| GET | `/farms/:id/tasks` | List tasks |
| POST | `/farms/:id/tasks` | Create task |
| PATCH | `/tasks/:id/status` | Update task status |

#### Fertigation

| Method | Path | Description |
|--------|------|-------------|
| GET | `/farms/:id/fertigation/reservoirs` | List reservoirs |
| POST | `/farms/:id/fertigation/reservoirs` | Create reservoir |
| PATCH | `/fertigation/reservoirs/:rid` | Update reservoir |
| DELETE | `/fertigation/reservoirs/:rid` | Delete reservoir |
| GET | `/farms/:id/fertigation/ec-targets` | List EC targets |
| POST | `/farms/:id/fertigation/ec-targets` | Create EC target |
| GET | `/farms/:id/fertigation/programs` | List programs |
| POST | `/farms/:id/fertigation/programs` | Create program |
| PATCH | `/fertigation/programs/:rid` | Update program |
| DELETE | `/fertigation/programs/:rid` | Delete program |
| GET | `/farms/:id/fertigation/events` | List fertigation events (`?crop_cycle_id=` optional) |
| POST | `/farms/:id/fertigation/events` | Create fertigation event (optional `crop_cycle_id`) |

#### Crop cycles

| Method | Path | Description |
|--------|------|-------------|
| GET | `/farms/:id/crop-cycles` | List crop cycles |
| POST | `/farms/:id/crop-cycles` | Create crop cycle |
| GET | `/crop-cycles/:id` | Get crop cycle |
| PUT | `/crop-cycles/:id` | Update crop cycle |
| DELETE | `/crop-cycles/:id` | Deactivate crop cycle |
| PATCH | `/crop-cycles/:id/stage` | Update growth stage |

#### Costs, finance & receipts

| Method | Path | Description |
|--------|------|-------------|
| GET | `/farms/:id/costs/summary` | Cost totals (income, expenses, net) |
| GET | `/farms/:id/costs` | List cost transactions (`limit`, `offset`, …) |
| GET | `/farms/:id/costs/export` | Download CSV (`format=csv` or `format=gl_csv`) |
| GET | `/farms/:id/finance/coa-mappings` | List COA mappings for GL export |
| PUT | `/farms/:id/finance/coa-mappings` | Save COA mapping overrides |
| DELETE | `/farms/:id/finance/coa-mappings` | Reset all COA overrides |
| DELETE | `/farms/:id/finance/coa-mappings/:category` | Reset one category override |
| POST | `/farms/:id/costs` | Create cost transaction |
| PUT | `/costs/:id` | Update cost transaction |
| DELETE | `/costs/:id` | Delete cost transaction |
| POST | `/farms/:id/cost-receipts` | Upload cost receipt (**multipart**: `file`, optional `cost_transaction_id`) |
| GET | `/file-attachments/:id/content` | Inline file bytes (cost receipt when linked) |
| GET | `/file-attachments/:id/download` | Presigned or proxied download URL JSON (backend-dependent) |

#### Alerts

| Method | Path | Description |
|--------|------|-------------|
| GET | `/farms/:id/alerts` | List alerts for farm |
| GET | `/farms/:id/alerts/unread-count` | Unread count |
| PATCH | `/alerts/:id/read` | Mark alert read |
| PATCH | `/alerts/:id/acknowledge` | Acknowledge alert |

#### Natural farming — inputs & batches

| Method | Path | Description |
|--------|------|-------------|
| GET | `/farms/:id/naturalfarming/inputs` | List input definitions |
| POST | `/farms/:id/naturalfarming/inputs` | Create input definition |
| PUT | `/naturalfarming/inputs/:id` | Update input definition |
| DELETE | `/naturalfarming/inputs/:id` | Delete input definition |
| GET | `/farms/:id/naturalfarming/batches` | List input batches |
| POST | `/farms/:id/naturalfarming/batches` | Create input batch |
| PUT | `/naturalfarming/batches/:id` | Update input batch |
| DELETE | `/naturalfarming/batches/:id` | Delete input batch |

#### Natural farming — recipes & components

| Method | Path | Description |
|--------|------|-------------|
| GET | `/farms/:id/naturalfarming/recipes` | List application recipes |
| POST | `/farms/:id/naturalfarming/recipes` | Create recipe |
| GET | `/naturalfarming/recipes/:id` | Get recipe |
| PUT | `/naturalfarming/recipes/:id` | Update recipe |
| DELETE | `/naturalfarming/recipes/:id` | Delete recipe |
| GET | `/naturalfarming/recipes/:id/components` | List recipe components |
| POST | `/naturalfarming/recipes/:id/components` | Add component |
| DELETE | `/naturalfarming/recipes/:id/components/:iid` | Remove component (`:iid` = component row id) |

---

## Seed Data (v1.004)

The master seed loads a **demo farm** with natural-farming inventory and **JADAM**-style input names (JMS, JLF, …), photoperiod schedules, and automation — verified clean against the live schema:

| Table | Rows | Contents |
|-------|------|----------|
| `farms` | 1 | gr33n Demo Farm |
| `zones` | 3 | Veg Room, Flower Room, Outdoor Garden |
| `sensors` | 10 | PAR, lux, temp, humidity, EC, pH, CO2, soil moisture |
| `input_definitions` | 15 | JMS, LAB, FPJ, FFJ, OHN, JHS, WCA, WCS, JWA, JS, JLF variants, compost tea |
| `application_recipes` | 14 | Soil drenches, foliar sprays, pest control, fungicide |
| `recipe_components` | 20 | Input-to-recipe links with dilution ratios |
| `schedules` | 14 | Light (24/0, 18/6, 16/8, 12/12) + watering programs per grow stage |
| `automation_rules` | 7 | Automated light on/off rules per grow stage |

---

## Make Commands

```bash
make help       # Show all targets
make bootstrap-local  # Guided DB + env + UI deps (see docs/local-operator-bootstrap.md)
make bootstrap-local-docker  # Same, but start stack with docker compose
make run        # Run the API server
make run-receiver # Run optional Insert Commons receiver (see docs/insert-commons-receiver-playbook.md)
make dev        # Run API + UI dev server in parallel
make ui         # Run the Vue dev server
make build      # Build the Go binary
make build-ui   # Build the Vue frontend for production
make test       # Run Go tests
make lint       # Run go vet
make sqlc       # Regenerate sqlc Go code from SQL queries
make seed       # Apply seed data to the database
make schema     # Apply the schema to the database
make up         # Start Docker Compose services
make down       # Stop Docker Compose services
make logs       # Tail Docker Compose logs
make clean      # Remove build artifacts
```

---

## Raspberry Pi Client

The Pi daemon runs four threads concurrently:

- **sensor-loop** — reads each GPIO/I2C sensor at its configured interval, POSTs to `POST /sensors/:id/readings`
- **heartbeat-loop** — PATCHes device status every 30s so the dashboard shows "online"
- **schedule-loop** — polls `GET /farms/:id/devices` for `pending_command` in device config JSONB, executes via GPIO, reports via `POST /actuators/:id/events`, then clears via `DELETE /devices/:id/pending-command`
- **flush-loop** — drains the offline SQLite queue when API becomes reachable

Configure sensors, actuators (with `device_id`), and GPIO pins in `pi_client/config.yaml`. Install as a systemd service with `pi_client/setup.sh` so it starts automatically on boot.

---

### MQTT telemetry bridge (microcontrollers)

MCUs can publish to an on-farm **MQTT broker**; a **bridge** process subscribes and forwards to **`POST /sensors/readings/batch`** using `X-API-Key` (same server `PI_API_KEY` as the Pi daemon). Reference implementation: [`pi_client/mqtt_telemetry_bridge.py`](pi_client/mqtt_telemetry_bridge.py). Topics, TLS, ACLs, and tasking: [`docs/mqtt-edge-operator-playbook.md`](docs/mqtt-edge-operator-playbook.md).

---

## 🔄 AI Augmentation with Consent

gr33n doesn't replace farm.chat — it augments it.

For users who choose to integrate local AI, gr33n offers schema-guided intelligence via LM Studio and gr33n_inserts. This AI layer respects user autonomy and privacy:

- AI is modular, never mandatory.
- Prompts are schema-aligned, not generic.
- Control is user-directed, through defined integration tiers.

| Mode | AI Role | User Control |
|------|---------|-------------|
| Ambient | Passive suggestions | Low (opt-in cues) |
| Reactive | Triggered by schema events | Medium (configurable) |
| Sovereign | Fully directed by user input | High (full control) |

---

## Roadmap Status

A phase-by-phase ledger of what's live on `main`. Each row links to the governing plan doc where one exists; undated rows predate the phase-plan convention.

| Phase | Focus | Status | Links |
|------:|-------|--------|-------|
| 10 | JWT smoke tests, farm-scoped write auth, fertigation ↔ crop cycle link, costs CSV, SensorDetail UX | ✅ Done | — |
| 11 | Farm RBAC, cost receipts + local storage, PWA shell, Insert Commons opt-in | ✅ Done | — |
| 12 | Insert Commons federation | ✅ Done | `db/migrations/20260416_phase12_insert_commons_federation.sql` |
| 13 | Platform evolution — receiver, audit/compliance, offline, finance depth, tenancy | ✅ Done | [plan](docs/plans/phase_13_platform_evolution.plan.md) · [ops doc](docs/phase-13-operator-documentation.md) |
| 14 | Field network & commons — MQTT/edge, insert pipeline, catalog, receiver, FCM, org governance, domain schema stubs | ✅ Done | [plan](docs/plans/phase_14_network_and_commons.plan.md) · [ops doc](docs/phase-14-operator-documentation.md) |
| 15 | Farm onboarding & templates | ✅ Done | [plan](docs/plans/phase_15_farm_onboarding.plan.md) |
| 18 | Platform polish | ✅ Done | [plan](docs/plans/phase_18_platform_polish.plan.md) |
| 19 | Safety & alert rules | ✅ Done | [plan](docs/plans/phase_19_safety_and_alert_rules.plan.md) |
| 20 | Automation rule engine (sensor-driven rules, dispatch, cooldowns, notifier fan-out) | ✅ Done | [plan](docs/plans/phase_20_automation_rule_engine.plan.md) |
| 20.6 | Stage-scoped setpoints (`gr33ncore.zone_setpoints`) + rule engine integration + UI | ✅ Done | — |
| 20.7 | Cost/energy rollups — nightly runtime × watts × kWh price; per-cycle P&L via `cost_transactions.crop_cycle_id` | ✅ Done | — |
| 20.8 | Animal husbandry (groups + lifecycle events), typed `aquaponics.loops`, feed autologging, bootstrap upgrade | ✅ Done | — |
| 20.9 | Labor auto-cost (timer + manual entry + profile rate); program `executable_actions` surface + `metadata.steps` backfill + `ResolveProgramActions` fallback | ✅ Done | — |
| 20.95 | RAG-prep column adds & housekeeping (executable_actions.program_id, cost/energy columns, labor schema, animal/aquaponics scope) | ✅ Done | [plan](docs/plans/phase_20_95_rag_prep_and_housekeeping.plan.md) |
| 21 | Crop cycle analytics & yield (`GET /crop-cycles/{id}/summary`, compare, UI, CSV per plan) | ⏳ Planned — plan todos still open; **partial today:** `GET /crop-cycles/{id}/cost-summary` (Phase 20.7) only | [plan](docs/plans/phase_21_crop_cycle_analytics.plan.md) |
| **22** | **Worker program-tick + final `metadata.steps` backfill sweep** — `runProgramTick` dispatches `executable_actions` per program, `automation_runs.program_id` attribution, 20260517 sweep + per-program NOTICE log, structured fallback warning | ✅ Done | — |

**In flight / next up** (no committed plan docs yet; candidates in rough priority order):

- [x] **Pi ↔ API contract pass** — smoke tests `TestPiContract*` in [`cmd/api/smoke_pi_contract_test.go`](cmd/api/smoke_pi_contract_test.go): enqueue `pending_command` → Pi-key `GET /farms/1/devices` (base64 `config` decode) → `POST /actuators/{id}/events` with `triggered_by_schedule_id`, `program_id` (→ `meta_data`), `triggered_by_rule_id`, `parameters_sent` → `DELETE` pending. [`pi_client/gr33n_client.py`](pi_client/gr33n_client.py) forwards `rule_id` / `program_id` / `schedule_id` and decodes real API `config`. `requireAPIKey` injects `PiEdgeAuth` context so `RecordEvent` can authorize like `GET /farms/{id}/devices`.
- [ ] **Phase 21** — ship cycle summary / compare / UI / CSV per [`docs/plans/phase_21_crop_cycle_analytics.plan.md`](docs/plans/phase_21_crop_cycle_analytics.plan.md) (distinct from 20.7 cost-summary).
- [ ] **Deprecate `programs.metadata.steps`** — after N deploys with zero fallback warnings, promote `action_source` checks to hard errors and drop the column.
- [ ] **Program "run now" API** — explicit trigger for unscheduled / ad-hoc programs (today only schedule-bound programs fire via the tick).
- [ ] **AI augmentation / LM Studio integration** for insert-sharing recommendations.
- [ ] **Mobile distribution polish** — Capacitor packaging, store submission checklist.

## Project Roadmap

- [x] gr33ncore schema — users, sensors, schedules, zones, automation rules
- [x] gr33nnaturalfarming schema — inputs, recipes, batches
- [x] Go REST API — farms, zones, devices, sensors, tasks, readings
- [x] Natural farming demo seed — 15 inputs, 14 recipes, full automation (JADAM-style labels)
- [x] sqlc query layer + enum types
- [x] Vue 3 frontend — Dashboard, Zones, Sensors, Actuators, Schedules, Settings, Inventory
- [x] Raspberry Pi sensor client with systemd daemon
- [x] OpenAPI spec (openapi.yaml)
- [x] Sensor readings live on dashboard (SSE stream with JWT query param auth)
- [x] Phase 10 — JWT smoke tests (`AUTH_MODE=auth_test`), farm-scoped write authorization, fertigation ↔ crop cycle link, costs CSV export, SensorDetail export UX
- [x] Phase 11 — Farm RBAC (viewer / operator / finance / manager / owner), cost receipts + local `FILE_STORAGE_DIR` storage, **PWA-first** installable shell (manifest + SW in production builds; Capacitor still an option for store-distributed apps), Insert Commons opt-in + early sync hook, OpenAPI updates
- [x] Phase 13 — Platform evolution (receiver-side Insert Commons, audit/compliance API, offline + finance depth, tenancy experiments, optional Capacitor scaffold; [`docs/phase-13-operator-documentation.md`](docs/phase-13-operator-documentation.md) indexes plans and playbooks)
- [x] Phase 14 — Field network & commons (MQTT/edge, insert pipeline, gr33n_inserts catalog, federation/receiver depth, FCM notifications, org governance, domain schema stubs; [`docs/plans/phase_14_network_and_commons.plan.md`](docs/plans/phase_14_network_and_commons.plan.md), [`docs/phase-14-operator-documentation.md`](docs/phase-14-operator-documentation.md))
- [x] Actuator control pipeline (automation worker → pending_command → Pi poll → execute → report)
- [x] Fertigation module — reservoirs, EC targets, programs, events
- [x] Natural farming inventory UI — input definitions & batch tracking
- [x] Pi heartbeat loop — devices show online/offline in real time
- [x] Docker Compose + Dockerfile for containerized deployment
- [x] Microcontroller integrations — MQTT → HTTP bridge ([`pi_client/mqtt_telemetry_bridge.py`](pi_client/mqtt_telemetry_bridge.py), [`docs/mqtt-edge-operator-playbook.md`](docs/mqtt-edge-operator-playbook.md)); field tasking unchanged (`pending_command` + Pi / bridge poll)
- [x] Data insert pipeline (Insert Commons validation, approval bundles, export — [`docs/insert-commons-pipeline-runbook.md`](docs/insert-commons-pipeline-runbook.md))
- [ ] LM Studio integration and AI scaffolds for insert-sharing
- [x] gr33n_inserts — commons catalog API (browse + farm import audit; [`docs/commons-catalog-operator-playbook.md`](docs/commons-catalog-operator-playbook.md))
- [x] Stub schemas `gr33ncrops`, `gr33nanimals`, `gr33naquaponics` (placeholder tables; enable via `farm_active_modules` — [`docs/domain-modules-operator-playbook.md`](docs/domain-modules-operator-playbook.md))
- [x] Phase 20 — automation rule engine (sensor-driven conditions, action dispatch, cooldowns, rule-driven notifications) — [plan](docs/plans/phase_20_automation_rule_engine.plan.md)
- [x] Phase 20.6–20.9 — stage-scoped setpoints, cost/energy nightly rollups, animal husbandry, labor auto-cost, program actions + `metadata.steps` backfill
- [x] Phase 20.95 — RAG-prep schema housekeeping for AI consumption — [plan](docs/plans/phase_20_95_rag_prep_and_housekeeping.plan.md)
- [ ] Phase 21 — crop cycle analytics & yield (summary, compare, UI, CSV) — [plan](docs/plans/phase_21_crop_cycle_analytics.plan.md) *(not started; see `GET /crop-cycles/{id}/cost-summary` for per-cycle costs only)*
- [x] Phase 22 — worker program-tick (`runProgramTick`), `automation_runs.program_id`, final `metadata.steps` backfill sweep, observable fallback warning

---

## Contribute

- Fork this repo
- Join the insert-sharing network (coming soon in gr33n_inserts)
- Help build bridges between sensors, dashboards, and soil
- Translate docs, test offline installs, or write a better knf_notes parser

---

## Built for the Commons

> "Built for the commons."

The commons means shared knowledge, shared code, shared resilience. It's an ancient concept — like the village well or a seed bank — remixed into digital space.

gr33n lives in this tradition:
Free to use, fork, and rebuild.
Not fenced off behind corporate toll booths.

---

## License

**GNU Affero General Public License v3.0 (AGPL-3.0)**

Use it. Fork it. Share it.
If you run it as a service — cloud, SaaS, or otherwise — you must release your modifications back to the community. No exceptions. No toll booths.

Just don't try to put a fence around the commons.

Built by farmers, hackers, and friends.
With sunlight and rage.
