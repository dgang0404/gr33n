# gr33n — Architecture

## Overview

gr33n is a local-first farm automation platform with four independent layers that communicate over HTTP. Nothing requires the internet at runtime.

```
┌─────────────────────────────────────────────────────────┐
│                    Browser / LAN                        │
│                                                         │
│  ┌──────────────────┐         ┌──────────────────────┐  │
│  │   Vue 3 Frontend │◄───────►│   Go REST API        │  │
│  │   localhost:5173 │  HTTP   │   localhost:8080      │  │
│  │                  │  Axios  │                       │  │
│  │  Pinia store     │         │  net/http stdlib      │  │
│  │  Tailwind CSS    │         │  sqlc query layer     │  │
│  └──────────────────┘         └──────────┬───────────┘  │
│                                          │               │
└──────────────────────────────────────────┼───────────────┘
                                           │ pgx/v5
                              ┌────────────▼────────────┐
                              │   PostgreSQL + TimescaleDB│
                              │   gr33ncore schema        │
                              │   sensor_readings         │
                              │   (hypertable)            │
                              └────────────▲─────────────┘
                                           │ HTTP POST
                              ┌────────────┴─────────────┐
                              │   Raspberry Pi Client     │
                              │   Python daemon           │
                              │   GPIO / I2C sensors      │
                              │   gr33n_client.py         │
                              └──────────────────────────┘
```

---

## Go API

### Structure

```
cmd/api/
  main.go      — connectDB() → registerRoutes() → ListenAndServe()
  routes.go    — all mux.HandleFunc registrations
  cors.go      — permissive CORS for local dev (tighten for production)

internal/
  db/          — sqlc output: Queries struct, all param/row types, models
  handler/     — one package per resource (farm, zone, sensor, device, actuator, task, automation, auth, fertigation, naturalfarming, sse)
  httputil/    — WriteJSON(w, status, v) and WriteError(w, status, msg)
  platform/
    commontypes/ — enums (TaskStatusEnum, DeviceStatusEnum, etc.)
  automation/   — schedule execution worker (cron evaluation + action dispatch)
```

### Request lifecycle

```
HTTP request
  → corsMiddleware (adds CORS headers)
  → http.ServeMux (pattern match)
  → handler.Method(w, r)
      → parse path value (r.PathValue("id"))
      → decode JSON body if needed
      → db.New(pool).QueryFunction(ctx, params)
      → httputil.WriteJSON(w, status, result)
```

### Automation worker lifecycle

```
process start
  → NewWorker(pool, simulationMode, opts...)
  → worker ticker (30s)
      → ListActiveSchedules()
      → cron match for current minute
      → cooldown check (skip if last success within cooldown window)
      → idempotency check (skip if run with same schedule+minute key exists)
      → ListExecutableActionsBySchedule()
      → execute action (with retry for transient errors):
          - control_actuator → actuator_events (+ actuator state update in simulation mode)
          - update_record_in_gr33n (fertigation_events) → fertigation event insert
      → insert automation_runs log row (with idempotency_key in details)
      → update schedules.last_triggered_time
```

#### Execution safeguards

| Safeguard | Behavior |
|-----------|----------|
| Same-minute dedup | `shouldTriggerNow` skips if `last_triggered_time` equals current minute |
| Cooldown | Configurable via `AUTOMATION_COOLDOWN_SECONDS` (default 120s). Skips execution if last successful run is within the window |
| Idempotency | SHA-256 key from `schedule_id:minute`. Checks `automation_runs` JSONB for existing key before execution |
| Retry | Transient errors (connection, timeout) retried up to 2x with exponential backoff. Permanent errors fail immediately |
| Error classification | `isTransient()` categorizes by error message patterns (connection refused, timeout, pgconn) |

Simulation mode is controlled by `AUTOMATION_SIMULATION_MODE` (default `true`).

### Database connection

- Uses `pgxpool` with 20 max / 2 min connections
- 5-attempt retry loop on startup (useful when DB is slow to start)
- Single `DATABASE_URL` env var — defaults to Unix socket peer auth for local dev

---

## PostgreSQL Schema

Schema namespace: `gr33ncore`

### Core tables

| Table | Purpose |
|-------|---------|
| `farms` | Top-level farm entity |
| `zones` | Spatial subdivisions of a farm (rooms, beds) |
| `sensors` | Sensor registry — type, unit, zone, thresholds |
| `sensor_readings` | TimescaleDB hypertable — time-series readings |
| `devices` | Physical devices — Pi nodes, controllers |
| `actuators` | Controllable outputs linked to devices |
| `tasks` | Farm task management with status workflow |
| `schedules` | Cron-based automation schedules |
| `units` | Master unit reference with conversion factors |
| `profiles` | User accounts (Supabase-compatible auth.users) |

### sensor_readings hypertable

```sql
CREATE TABLE gr33ncore.sensor_readings (
  reading_time  TIMESTAMPTZ NOT NULL,   -- partition key (time col first)
  sensor_id     BIGINT NOT NULL,
  value_raw     NUMERIC NOT NULL,
  value_normalized NUMERIC,             -- auto-filled by trigger
  is_valid      BOOLEAN DEFAULT TRUE,
  ...
  PRIMARY KEY (reading_time, sensor_id)
);
-- TimescaleDB partitions this by reading_time automatically
```

A `BEFORE INSERT` trigger normalizes `value_raw` to the base unit for that sensor type using the `units` table conversion factors.

### Enum types

All enums live in `commontypes` package (sqlc override):

| Enum | Values |
|------|--------|
| `DeviceStatusEnum` | online, offline, error_comms, error_hardware, maintenance_mode, initializing, unknown, decommissioned, pending_activation |
| `TaskStatusEnum` | todo, in_progress, on_hold, completed, cancelled, blocked_requires_input, pending_review |
| `UserRoleEnum` | user, farm_manager, farm_worker, gr33n_system_admin |

---

## Vue 3 Frontend

### State management (Pinia)

All API state lives in one store: `ui/src/stores/farm.js`

```
state:
  farm             — current farm object
  zones[]          — all zones for farm
  sensors[]        — all sensors for farm
  devices[]        — all devices for farm
  actuators[]      — all actuators for farm
  schedules[]      — automation schedules
  automationRuns[] — execution history
  tasks[]          — all tasks for farm
  readings{}       — map: sensor_id → latest reading

actions:
  loadAll(farmId)                    — parallel fetch: farm + zones + sensors + devices + actuators
  loadTasks / loadSchedules          — fetch tasks or schedules
  loadAutomationRuns                 — fetch automation run history
  refreshReadings()                  — loop sensors, GET /sensors/:id/readings/latest
  toggleDevice(id, status)           — PATCH /devices/:id/status
  toggleActuator(id, stateText)      — PATCH /actuators/:id/state
  updateTaskStatus(id, s)            — PATCH /tasks/:id/status
  loadReservoirs / createReservoir   — fertigation CRUD
  loadEcTargets / createEcTarget     — fertigation CRUD
  loadFertigationPrograms / Events   — fertigation CRUD
  loadActuatorEvents(id, opts)       — GET /actuators/:id/events
```

### Views

| View | Route | Data source |
|------|-------|-------------|
| Dashboard | `/` | store.farm, store.sensors, store.devices |
| Zones | `/zones` | store.zones, store.sensorsByZone(), store.devicesByZone() |
| Zone Detail | `/zones/:id` | Operator console: live sensor readings, actuator toggles, actuator event timeline, fertigation summary |
| Sensors | `/sensors` | store.sensors, store.readings, store.zones |
| Actuators | `/actuators` | store.actuators, store.zones |
| Schedules | `/schedules` | store.schedules, store.automationRuns, worker health |
| Tasks | `/tasks` | store.tasks (kanban-style status columns) |
| Fertigation | `/fertigation` | Tabbed: reservoirs, EC targets, programs, events with create forms |
| Inventory | `/inventory` | hardcoded JADAM inputs (stub) |
| Login | `/login` | Auth store login action (public route) |
| Settings | `/settings` | Account info, password change, sign out |

---

## Raspberry Pi Client

### Threading model

```
main thread
  └── SensorLoop thread      — reads sensors, POSTs readings
  └── HeartbeatLoop thread   — PATCHes device status every 30s
  └── CommandPollLoop thread — polls for pending commands, triggers GPIO
```

### Sensor reading flow

```
1. SensorLoop wakes up (sleep interval from config)
2. Reads GPIO / I2C pin for sensor hardware_identifier
3. POST /sensors/{sensor_id}/readings
   { "value_raw": 22.5, "is_valid": true, "battery_level_percent": 87 }
4. API trigger normalizes value_raw → value_normalized
5. Frontend refreshReadings() picks it up on next poll
```

### config.yaml structure

```yaml
api_base_url: http://192.168.1.100:8080
device_uid: pi-node-01
sensors:
  - sensor_id: 1
    hardware_identifier: GPIO4
    sensor_type: temperature
    interval_seconds: 60
  - sensor_id: 5
    hardware_identifier: 0x40   # I2C address for HTU21D
    sensor_type: humidity
    interval_seconds: 60
```

---

## Data Flow: Sensor Reading End-to-End

```
Pi reads GPIO pin
  → POST /sensors/1/readings { value_raw: 22.5 }
  → InsertSensorReading (sqlc)
  → normalize_sensor_reading TRIGGER fires
      → converts 22.5°C → 22.5°C (already celsius base unit)
      → sets value_normalized, normalized_unit_id
  → 201 Created { reading_time, sensor_id, value_raw, value_normalized, ... }

Frontend (every 30s via refreshReadings())
  → GET /sensors/1/readings/latest
  → store.readings[1] = { value_normalized: 22.5, is_valid: true, ... }
  → Sensors.vue reactively updates Last Reading column: "22.50"
  → status badge: "ok"
```

---

## Authentication

gr33n supports two explicit modes controlled by `AUTH_MODE`:

| Mode | `AUTH_MODE` | Behavior |
|------|-------------|----------|
| Dev | `dev` (default) | JWT and API key middleware pass through when secrets are unset. Top bar shows "DEV MODE" banner. |
| Production | `production` | Fatal on startup if `JWT_SECRET` or `PI_API_KEY` are missing. Full auth enforcement. |

`GET /auth/mode` (public) returns the current mode for frontend awareness.

### Environment variables

| Variable | Default | Purpose |
|----------|---------|---------|
| `DATABASE_URL` | `postgres://<user>@/gr33n?host=/var/run/postgresql` | PostgreSQL connection string |
| `PORT` | `8080` | API listen port |
| `AUTH_MODE` | `dev` | `dev` or `production` |
| `JWT_SECRET` | (empty) | HMAC-SHA256 signing key for dashboard JWTs |
| `PI_API_KEY` | (empty) | Shared secret for Pi client `X-API-Key` header |
| `CORS_ORIGIN` | `http://localhost:5173` | Allowed CORS origin |
| `ADMIN_USERNAME` | `admin` | Login username |
| `ADMIN_PASSWORD_HASH` | (empty) | bcrypt hash (or read from `~/.gr33n/admin.hash`) |
| `AUTOMATION_SIMULATION_MODE` | `true` | Worker simulates hardware instead of sending real commands |
| `AUTOMATION_COOLDOWN_SECONDS` | (empty) | Minimum seconds between successful schedule executions |

---

## Development Commands

```bash
# Run API
DATABASE_URL="postgres://$(whoami)@/gr33n?host=/var/run/postgresql" go run ./cmd/api/

# Run frontend
cd ui && npm run dev

# Regenerate sqlc query layer (after editing db/queries/*.sql)
sqlc generate

# Build check
go build ./...

# Run all Go tests (requires local DB with seed data)
go test ./... -count=1

# Run Pi client tests
cd pi_client && python3 -m pytest test_gr33n_client.py -v

# Reset and reseed database
psql -d gr33n -f db/schema/gr33n-schema-v2-FINAL.sql
psql -d gr33n -f db/seeds/master_seed.sql
```
