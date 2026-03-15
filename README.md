# gr33n 🌱

An open-source agricultural operating system designed to reclaim data, land, and autonomy.

[![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)
[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://go.dev)
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

- **Insert Commons (Coming Soon)** — A sibling repo for community-contributed data (pest trials, IMO recipes, soil logs) with scrubbers and staging.

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

# 4. Run the API
export DATABASE_URL="postgres://$(whoami)@/gr33n?host=/var/run/postgresql"
go run ./cmd/api/

# 5. Run the frontend (separate terminal)
cd ui && npm install && npm run dev
```

API → `http://localhost:8080`
UI  → `http://localhost:5173`

---

## API Endpoints

Base URL: `http://localhost:8080` — full spec in [openapi.yaml](openapi.yaml).

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | API + DB health check |
| GET | `/farms/:id` | Farm detail |
| GET | `/farms/:id/zones` | List zones |
| GET | `/farms/:id/sensors` | List sensors |
| GET | `/farms/:id/devices` | List devices |
| GET | `/farms/:id/tasks` | List tasks |
| GET | `/sensors/:id/readings/latest` | Latest reading for a sensor |
| POST | `/sensors/:id/readings` | Pi client posts a reading |
| PATCH | `/devices/:id/status` | Toggle actuator ON/OFF |
| PATCH | `/tasks/:id/status` | Advance task on Kanban board |

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

## Raspberry Pi Client

The Pi daemon runs three threads concurrently:

- **sensor-loop** — reads each GPIO/I2C sensor at its configured interval, POSTs to `POST /sensors/:id/readings`
- **heartbeat-loop** — PATCHes device status every 30s so the dashboard shows "online"
- **command-poll** — polls for pending actuator commands and triggers GPIO output

Configure sensors and GPIO pins in `pi_client/config.yaml`. Install as a systemd service with `pi_client/setup.sh` so it starts automatically on boot.

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
- [x] Vue 3 frontend — Dashboard, Zones, Sensors, Actuators, Schedules
- [x] Raspberry Pi sensor client with systemd daemon
- [x] OpenAPI spec (openapi.yaml)
- [ ] Sensor readings live on dashboard (Pi connection)
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
