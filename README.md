# gr33n 🌱

An open-source agricultural operating system designed to reclaim data, land, and autonomy.

[![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)
[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev)
[![Vue](https://img.shields.io/badge/Vue-3-4FC08D?logo=vue.js)](https://vuejs.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-14+-336791?logo=postgresql)](https://postgresql.org)

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

- **Modularity** — Each ag domain (crops, animals, KNF inputs, IoT sensors) lives in its own schema. Use what you need, prune the rest. Enable modules per-farm via `gr33ncore.farm_active_modules`.

- **Connectivity Optional** — Works offline, intranet-only, or online. Supports Supabase or bare-metal Postgres with TimescaleDB/PostGIS.

- **Automation-Ready** — Schedule tasks, trigger actuators, run AI models — or run it all manually. Your tech, your tempo.

- **Insert Commons (MVP)** — Per-farm opt-in in Settings; `POST /farms/{id}/insert-commons/sync` records a sync timestamp (full outbound adapters are still future work). Apply `db/migrations/20260415_phase11_rbac_receipts_commons.sql` on existing databases.

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
├── internal/
│   ├── db/                  # sqlc-generated query layer (DO NOT EDIT)
│   ├── handler/
│   │   ├── farm/            # GET /farms/:id
│   │   ├── zone/            # Zones CRUD
│   │   ├── sensor/          # Sensors CRUD + readings endpoints
│   │   ├── device/          # Devices CRUD + status toggle
│   │   └── task/            # Tasks list + status update
│   ├── httputil/            # WriteJSON / WriteError helpers
│   └── platform/
│       └── commontypes/     # Shared enum types for sqlc
├── db/
│   ├── schema/
│   │   └── gr33n-schema-v2-FINAL.sql   # Full PostgreSQL schema (source of truth)
│   ├── seeds/
│   │   └── master_seed.sql             # JADAM demo data v1.004
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

Full setup in [INSTALL.md](INSTALL.md). Short version:

```bash
# 1. Clone
git clone https://github.com/dgang0404/gr33n.git
cd gr33n-api

# 2. Create and migrate the database
sudo -u postgres psql -c "CREATE DATABASE gr33n;"
psql -d gr33n -f db/schema/gr33n-schema-v2-FINAL.sql

# 3. Seed with JADAM demo data
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

---

## API Endpoints

Base URL: `http://localhost:8080` — full spec in [openapi.yaml](openapi.yaml).

### Public

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | API + DB health check |
| POST | `/auth/login` | Authenticate & receive JWT |
| GET | `/auth/mode` | Current auth mode (dev/production) |

### Pi Routes (API key)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/sensors/:id/readings` | Pi posts a sensor reading |
| PATCH | `/devices/:id/status` | Pi heartbeat / status update |
| POST | `/actuators/:id/events` | Pi reports executed command |
| DELETE | `/devices/:id/pending-command` | Pi clears pending command after execution |

### Dashboard Routes (JWT)

JWT is required for all routes in this section. Mutations (POST, PUT, PATCH, DELETE) also require the user to be a **farm member** (see `gr33ncore.farm_memberships`) or the farm **owner** (`gr33ncore.farms.owner_user_id`). Integration tests in `cmd/api/smoke_test.go` run with `AUTH_MODE=auth_test` and a real JWT.

| Method | Path | Description |
|--------|------|-------------|
| PATCH | `/auth/password` | Change password |
| GET | `/units` | List all measurement units |
| GET | `/farms/:id` | Farm detail |
| GET | `/farms/:id/zones` | List zones |
| GET | `/farms/:id/devices` | List devices |
| GET | `/farms/:id/sensors` | List sensors |
| GET | `/farms/:id/actuators` | List actuators |
| GET | `/farms/:id/schedules` | List schedules |
| GET | `/farms/:id/tasks` | List tasks |
| GET | `/farms/:id/automation/runs` | List automation runs |
| GET | `/farms/:id/sensors/stream` | SSE live sensor readings |
| GET | `/sensors/:id` | Sensor detail |
| POST | `/farms/:id/sensors` | Create sensor |
| DELETE | `/sensors/:id` | Delete sensor |
| GET | `/sensors/:id/readings/latest` | Latest reading |
| GET | `/sensors/:id/readings` | List readings (since/until/limit) |
| GET | `/sensors/:id/readings/stats` | Aggregate stats for a time range |
| GET | `/devices/:id` | Device detail |
| POST | `/farms/:id/devices` | Create device |
| DELETE | `/devices/:id` | Delete device |
| PATCH | `/actuators/:id/state` | Update actuator state |
| GET | `/actuators/:id/events` | Actuator event history |
| PATCH | `/schedules/:id/active` | Toggle schedule active |
| GET | `/automation/worker/health` | Automation worker status |
| GET | `/zones/:id` | Zone detail |
| POST | `/farms/:id/zones` | Create zone |
| DELETE | `/zones/:id` | Delete zone |
| PATCH | `/tasks/:id/status` | Update task status |
| GET | `/schedules/:id/actuator-events` | Events by schedule |
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
| GET | `/farms/:id/crop-cycles` | List crop cycles |
| POST | `/farms/:id/crop-cycles` | Create crop cycle |
| GET | `/crop-cycles/:id` | Get crop cycle |
| PUT | `/crop-cycles/:id` | Update crop cycle |
| DELETE | `/crop-cycles/:id` | Deactivate crop cycle |
| PATCH | `/crop-cycles/:id/stage` | Update growth stage |
| GET | `/farms/:id/costs/summary` | Cost totals (income, expenses, net) |
| GET | `/farms/:id/costs` | List cost transactions |
| GET | `/farms/:id/costs/export` | Download costs as CSV |
| POST | `/farms/:id/costs` | Create cost transaction |
| PUT | `/costs/:id` | Update cost transaction |
| DELETE | `/costs/:id` | Delete cost transaction |
| GET | `/farms/:id/naturalfarming/recipes` | List application recipes |
| POST | `/farms/:id/naturalfarming/recipes` | Create recipe |
| GET | `/naturalfarming/recipes/:id` | Get recipe |
| PUT | `/naturalfarming/recipes/:id` | Update recipe |
| DELETE | `/naturalfarming/recipes/:id` | Delete recipe |
| GET | `/farms/:id/naturalfarming/inputs` | List NF input definitions |
| GET | `/farms/:id/naturalfarming/batches` | List NF input batches |

---

## Seed Data (v1.004)

The master seed loads a complete JADAM natural farming demo dataset — verified clean against the live schema:

| Table | Rows | Contents |
|-------|------|----------|
| `farms` | 1 | gr33n Demo Farm |
| `zones` | 4 | Seedling Room, Veg Room, Flower Room, Outdoor Beds |
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
make run        # Run the API server
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

## Project Roadmap

- [x] gr33ncore schema — users, sensors, schedules, zones, automation rules
- [x] gr33nnaturalfarming schema — inputs, recipes, batches
- [x] Go REST API — farms, zones, devices, sensors, tasks, readings
- [x] JADAM natural farming seed data — 15 inputs, 14 recipes, full automation
- [x] sqlc query layer + enum types
- [x] Vue 3 frontend — Dashboard, Zones, Sensors, Actuators, Schedules, Settings, Inventory
- [x] Raspberry Pi sensor client with systemd daemon
- [x] OpenAPI spec (openapi.yaml)
- [x] Sensor readings live on dashboard (SSE stream with JWT query param auth)
- [x] Phase 10 — JWT smoke tests (`AUTH_MODE=auth_test`), farm-scoped write authorization, fertigation ↔ crop cycle link, costs CSV export, SensorDetail export UX
- [x] Phase 11 — Farm RBAC (viewer / operator / finance / manager / owner), cost receipts + local `FILE_STORAGE_DIR` storage, **PWA-first** installable shell (manifest + SW in production builds; Capacitor still an option for store-distributed apps), Insert Commons opt-in + sync stub, OpenAPI updates
- [x] Actuator control pipeline (automation worker → pending_command → Pi poll → execute → report)
- [x] Fertigation module — reservoirs, EC targets, programs, events
- [x] Natural farming inventory UI — input definitions & batch tracking
- [x] Pi heartbeat loop — devices show online/offline in real time
- [x] Docker Compose + Dockerfile for containerized deployment
- [ ] Microcontroller integrations (MQTT + field tasking)
- [ ] Data insert pipeline (scrubbing, approval, federation-ready)
- [ ] LM Studio integration and AI scaffolds for insert-sharing
- [ ] gr33n_inserts — community contributed data commons
- [ ] gr33n_crops, gr33n_animals, gr33n_aquaponics module schemas

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
