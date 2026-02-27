# gr33n

An open-source agricultural operating system designed to reclaim data, land, and autonomy.

[![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)

---

## What Is gr33n?

gr33n is a modular, scalable, and decentralized farm management system built for real humans—not cloud landlords. Whether you're managing a homestead on solar or automating thousands of acres, gr33n adapts to your size, ethics, and bandwidth.

It's PostgreSQL schemas + Go APIs + front ends + microcontrollers + shared insert statements.

But more than that:  
it's a political stance in schema form.

---

## Why gr33n Exists

> "If your DNA, soil, labor, and climate data feed trillion-dollar industries—and you're not seeing a dime—that's not tech, that's extraction."

This project exists because:
- Big Ag is closing the loop on food systems, and we're cracking it back open.
- Data rights matter—even your soil and sunlight deserve consent.
- Billionaires shouldn't profit off your greenhouse or genome without giving back.
- Farmers, tinkerers, and off-gridders deserve tools that don't call home.

### 🔌 What Does "Don't Call Home" Mean?

That means gr33n will never require a permanent internet connection, forced login, or hidden "check-in" with third-party servers. Whether you're on an island, a mountaintop, or a mesh-netted greenhouse, gr33n works where you live, without compromise.

---

## Core Principles

- **Modularity**  
  Each ag domain (crops, animals, KNF inputs, IoT sensors) lives in its own schema. Use what you need, prune the rest.

- **Connectivity Optional**  
  Works offline, intranet-only, or online. Supports Supabase or bare-metal Postgres with TimescaleDB/PostGIS.

- **Automation-Ready**  
  Schedule tasks, trigger actuators, run AI models—or run it all manually. Your tech, your tempo.

- **Insert Commons (Coming Soon)**  
  A sibling repo for community-contributed data (pest trials, IMO recipes, soil logs) with scrubbers and staging.

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.23 |
| Database | PostgreSQL 14+ with TimescaleDB |
| Query layer | sqlc (generated — do not edit) |
| HTTP | Go standard library `net/http` |
| Auth | Supabase (hosted) / local bootstrap (dev) |
| Schema | Multi-schema PostgreSQL — `gr33ncore` + `gr33nnaturalfarming` |

---

## Repository Layout

```
gr33n/
├── cmd/api/
│   ├── main.go              # Entry point, DB connection, server startup
│   └── routes.go            # HTTP route registration
├── internal/
│   ├── db/                  # sqlc-generated query layer (do not edit)
│   ├── handler/             # HTTP handlers
│   │   ├── farm/
│   │   ├── zone/
│   │   ├── device/
│   │   └── sensor/
│   ├── httputil/            # Shared response helpers
│   └── platform/
│       └── commontypes/
│           └── enums.go     # Shared enum types used by sqlc
├── db/
│   ├── schema/
│   │   └── gr33n-schema-v2-FINAL.sql   # Full PostgreSQL schema
│   ├── seeds/
│   │   └── master_seed.sql             # JADAM seed data v1.004 (verified clean)
│   └── queries/             # sqlc SQL query source files
├── sqlc.yaml
├── go.mod
├── go.sum
├── INSTALL.md               # Full local dev setup guide
└── README.md
```

---

## Quick Start

Full setup instructions are in [INSTALL.md](INSTALL.md). Short version:

```bash
# 1. Clone
git clone https://github.com/dgang0404/gr33n.git
cd gr33n

# 2. Create and migrate the database
sudo -u postgres psql -c "CREATE DATABASE gr33n;"
psql -d gr33n -f db/schema/gr33n-schema-v2-FINAL.sql

# 3. Seed with JADAM demo data
psql -d gr33n -f db/seeds/master_seed.sql

# 4. Set env and run
export DATABASE_URL="postgres://$(whoami)@/gr33n?host=/var/run/postgresql"
go run ./cmd/api
```

Server starts at `http://localhost:8080`

---

## Seed Data (v1.004)

The master seed loads a complete JADAM natural farming demo dataset — verified clean against the live schema:

| Table | Rows | Contents |
|-------|------|----------|
| `input_definitions` | 15 | JMS, LAB, FPJ, FFJ, OHN, JHS, WCA, WCS, JWA, JS, JLF variants, compost tea |
| `application_recipes` | 14 | Soil drenches, foliar sprays, pest control, fungicide |
| `recipe_components` | 20 | Input-to-recipe links with dilution ratios |
| `schedules` | 14 | Light (24/0, 18/6, 16/8, 12/12) + watering programs per grow stage |
| `automation_rules` | 7 | Automated light on/off rules per grow stage |
| `sensors` | 10 | PAR, lux, temp, humidity, EC, pH, CO2, soil moisture templates |

---

## 🔄 AI Augmentation with Consent

gr33n doesn't replace farm.chat—it augments it.

For users who choose to integrate local AI, gr33n offers schema-guided intelligence via LM Studio and gr33n_inserts. This AI layer respects user autonomy and privacy, operating as a consent-based augmentation system:

- AI is modular, never mandatory.
- Prompts are schema-aligned, not generic.
- Control is user-directed, through defined integration tiers.

### Augmentation Tiers

| Mode      | AI Role                       | User Control          |
|-----------|-------------------------------|-----------------------|
| Ambient   | Passive suggestions           | Low (opt-in cues)     |
| Reactive  | Triggered by schema events    | Medium (configurable) |
| Sovereign | Fully directed by user input  | High (full control)   |

---

## Project Roadmap

- [x] gr33ncore schema — users, sensors, schedules, zones, automation rules
- [x] gr33nnaturalfarming schema — inputs, recipes, batches
- [x] Go backend with REST API — farms, zones, devices, sensors
- [x] JADAM natural farming seed data — 15 inputs, 14 recipes, full automation
- [x] sqlc query layer + enum types
- [ ] Front end with complexity pruning by farm type
- [ ] Microcontroller integrations (MQTT + field tasking)
- [ ] Data insert pipeline (scrubbing, approval, federation-ready)
- [ ] LM Studio integration and AI scaffolds for insert-sharing
- [ ] gr33n_inserts — community contributed data commons

---

## Got DNA?

Yeah, we talked about that too.  
If our genes are data, and our farms are extensions of that data, then gr33n is just the next step in reclaiming the means of biological production.  
Sovereignty begins with seeds—and schemas.

---

## Contribute

- Fork this repo
- Join the insert-sharing network (coming soon in gr33n_inserts)
- Help build bridges between sensors, dashboards, and soil
- Translate docs, test offline installs, or write a better knf_notes parser

---

## Built for the Commons

> "Built for the commons."

The commons means shared knowledge, shared code, shared resilience. It's an ancient concept—like the village well or a seed bank—remixed into digital space.

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
