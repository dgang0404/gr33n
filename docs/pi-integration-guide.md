# Pi Integration Guide (Pi → API → UI)

> **Scope:** How an on-farm Raspberry Pi running [`pi_client/gr33n_client.py`](../pi_client/gr33n_client.py) posts sensor readings, device status, and actuator events to the gr33n API, and how those flow into the dashboard UI.
>
> **Phase 50 (DB-first wiring):** Set GPIO / I2C / serial wiring in the **dashboard** (Sensors, Controls, device wizard), then **download** a generated `config.yaml` — see [§2a](#2a-db-first-wiring-and-config-generation-phase-50--recommended). Manual YAML editing remains a fallback; live config pull from the API is [Phase 51](plans/phase_51_pi_config_sync.plan.md).
>
> **Companion docs:**
> - API spec: [`openapi.yaml`](../openapi.yaml) — source of truth for every route used below.
> - MQTT edge playbook: [`mqtt-edge-operator-playbook.md`](mqtt-edge-operator-playbook.md) — MQTT → API bridge; **room-scale warehouse topics (Phase 31 WS4)** in § Room-scale pattern.
> - Operator workflow narrative: [`workflow-guide.md`](workflow-guide.md) — how the pieces connect end-to-end.
> - **Hardware layout & scaling:** [`raspberry-pi-and-deployment-topology.md`](raspberry-pi-and-deployment-topology.md) — Pi OS packages, full stack on one Pi, splitting DB/API/UI onto servers or containers as the farm grows.
> - **Laptop stub loop first (Phase 31 WS1):** [`local-operator-bootstrap.md`](local-operator-bootstrap.md) — *Edge loop in 5 commands* before wiring GPIO.
> - **Enterprise zone naming (hypothetical):** [`hypothetical-enterprise-topology.md`](hypothetical-enterprise-topology.md) — one plastic room → multiple zones.

---

## 1. What the Pi does

The Pi client is a single-process Python program that:

1. **Reads sensors** on a per-sensor interval (temperature, humidity, EC, pH, CO₂, PAR, soil moisture). Hardware drivers are auto-stubbed on non-Pi hosts, so the exact same file runs on a dev laptop for tests.
2. **Posts readings** to the API (one-shot or batched).
3. **Reports device status** (online / offline / error) on a heartbeat.
4. **Drains the device command queue** (`GET /devices/{id}/commands/next` → execute → `POST …/commands/{cid}/ack`) — **actuator**, **pulse**, and **mix_batch** (see §1.1–§1.2). Falls back to legacy `config.pending_command` if the queue is empty.
5. **Reports actuator events** (what it did and when) and, for **mix_batch**, posts a **mixing event** audit row.
6. **Falls back offline** — any failed POST is stored in a local SQLite queue (`offline_queue_path`) and flushed later via the batch endpoint.

All API calls go over plain HTTP(S); there is no long-lived socket. Authentication is a pre-shared **API key** sent as the **`X-API-Key`** HTTP header on every request (same spelling the server reads in `cmd/api/auth.go`; header names are case-insensitive, but examples below use this form). The API validates it with **`requireAPIKey`** middleware (`cmd/api/routes.go`).

### 1.1 Device command queue (Phase 39 — preferred)

Writers enqueue to **`gr33ncore.device_commands`** (FIFO per device). The Pi polls **`GET /devices/{id}/commands/next`**, which atomically marks the oldest **`pending`** row **`in_progress`**, runs it, then **`POST /devices/{id}/commands/{cid}/ack`** with `status: completed` or `failed`.

| `command_type` | Pi behavior |
|----------------|-------------|
| **`actuator`** | Instant on/off/deploy/retract on the bound actuator |
| **`pulse`** | **on → wait `duration_seconds` → off** (Phase 38 shape in payload) |
| **`mix_batch`** | Run **`mix_plan.steps[]`** sequentially — each step: channel → relay on → `run_seconds` → off; then **`POST /farms/{id}/fertigation/mixing-events`** |

**Typical fertigation program fire (cloud):** **`mix_batch`** (if program has recipe + reservoir + base EC) **then** **`pulse`** irrigate — two queue rows, one FIFO drain.

**`mix_batch` payload** (JSON):

| Field | Purpose |
|-------|---------|
| `mix_plan` | Cloud-calculated **`MixPlan`** — `steps[]` with `channel`, `run_seconds`, `volume_ml` |
| `program_id`, `reservoir_id` | Provenance for mixing events |

**Channel map (Pi `config.yaml`):** optional `mix_channels: [actuator_id, …]` — index 0 = channel 1. Without it, channels map to actuators sorted by id (demo only).

Operator enqueue: **`POST /farms/{id}/fertigation/mix-jobs`** (preview or enqueue). Calculator lives in Go (`internal/fertigation/mixplan`); Pi does not recompute doses.

### 1.2 Legacy `pending_command` (backward compat)

Older deployments and one-release mirroring still use **`devices.config.pending_command`** (base64 JSON on `GET /farms/{id}/devices`). The Pi drains the **queue first**; if empty, it reads **`pending_command`** and clears via **`DELETE /devices/{id}/pending-command`**.

**Payload** (actuator / pulse):

| Field | Purpose |
|-------|---------|
| `command` | `on`, `off`, `deploy`, `retract`, … |
| `actuator_id` | Which relay/pump |
| `duration_seconds` | Timed pulse |
| `schedule_id`, `rule_id`, `program_id`, `source` | Provenance |

**Do not rely on concurrent writes to `pending_command` alone** — use the queue for multi-step automation.

---

## 2. Environment & configuration

The Pi reads its config from `pi_client/config.yaml` (see `DEFAULT_CONFIG` in [`gr33n_client.py`](../pi_client/gr33n_client.py)):

```yaml
api:
  base_url: http://192.168.1.100:8080   # or https://gr33n-api.example.com
  timeout_seconds: 5
  api_key: "replace-with-PI_API_KEY"    # must match the API's PI_API_KEY env var
farm:
  farm_id: 1
offline_queue_path: /var/lib/gr33n/queue.db
offline_flush_interval_seconds: 60
```

For the **MQTT bridge** (`pi_client/mqtt_telemetry_bridge.py`) the same values are read from environment variables:

| Var | Purpose |
|-----|---------|
| `GR33N_API_URL` | Base URL of the API (no trailing slash) |
| `GR33N_FARM_ID` | Farm ID this edge reports for; messages whose topic `farm_id` differs are dropped |
| `PI_API_KEY` (or `MQTT_BRIDGE_API_KEY`) | Pre-shared key; must equal the API's `PI_API_KEY` env var |

The API side must have `PI_API_KEY` set (see `cmd/api/main.go` / deployment docs). Any Pi-tagged route (`requireAPIKey`) will reject requests without a matching header with **401**.

### 2a. DB-first wiring and config generation (Phase 50 — recommended)

**Happy path:** register hardware in gr33n, record **where each sensor and actuator is wired**, generate `config.yaml`, copy to the Pi. No hand-editing pin lists in YAML and no SQL.

| Step | Where | What |
|------|--------|------|
| 1 | **Settings → Connect edge device** (`/farms/{id}/devices/new`) | Register the Pi (`device_uid`, zone). |
| 2 | **Sensors** list + **sensor detail → Hardware wiring** | Set driver (`dht22`, `ads1115`, …), BCM GPIO, I2C channel, or serial port; assign the **edge device**. |
| 3 | **Controls** cards | Wiring badges show pin summary; edit wiring via sensor detail or API (`PATCH /actuators/{id}/wiring`). |
| 4 | Device wizard **Pi config** step | **Download config.yaml** or copy — generated from DB via `GET /devices/{id}/pi-config`. |
| 5 | On the Pi | `scp` the file to `pi_client/config.yaml`, set `api.api_key`, restart `gr33n` systemd unit. |

**Data model:** wiring lives in `sensors.config.wiring` and `actuators.config.wiring` (JSONB). API responses also expose a top-level `wiring` field. Validation rejects unknown drivers, duplicate pins per device (with an exception: multiple **DHT22** logical sensors may share one physical GPIO), and broken **derived** input references.

**Operator checks:**

- `make db-sanity-report` — prints wiring coverage and **fails** on pin/channel conflicts.
- Demo farm backfill: migration `db/migrations/20260607_phase50_hardware_wiring_backfill.sql` (idempotent).

**Manual fallback:** edit [`pi_client/config.yaml`](../pi_client/config.yaml) directly (§2, §8.3 step 4b). Use when the UI is unreachable or for one-off experiments. Keep `sensor_id` / `actuator_id` aligned with the database ([`scripts/print-demo-sensor-ids.sh`](../scripts/print-demo-sensor-ids.sh)).

**API (JWT, farm-scoped):**

| Method | Path | Purpose |
|--------|------|---------|
| `PATCH` | `/sensors/{id}/wiring` | Merge validated wiring into `sensors.config` |
| `PATCH` | `/actuators/{id}/wiring` | Merge validated wiring into `actuators.config` |
| `GET` | `/devices/{id}/pi-config` | Generate full `config.yaml` for that edge device |

Plan: [`plans/phase_50_hardware_wiring_visibility.plan.md`](plans/phase_50_hardware_wiring_visibility.plan.md). Runtime pull-from-API: Phase 51.

---

## 3. Routes used by the Pi

Every Pi-facing endpoint is declared with `requireAPIKey` in [`cmd/api/routes.go`](../cmd/api/routes.go):

| Method | Path | Purpose | Client method |
|--------|------|---------|---------------|
| `POST` | `/sensors/{id}/readings` | Post one reading for one sensor | `Gr33nApiClient.post_reading` |
| `POST` | `/sensors/readings/batch` | Post many readings across sensors in one request | `Gr33nApiClient.post_readings_batch` |
| `PATCH` | `/devices/{id}/status` | Heartbeat: `{"status": "online"}` | `Gr33nApiClient.patch_device_status` |
| `POST` | `/actuators/{id}/events` | Record what the actuator actually did | `Gr33nApiClient.post_actuator_event` |
| `GET` | `/devices/{id}/commands/next` | Dequeue head command (204 if empty) | `Gr33nApiClient.get_next_command` |
| `POST` | `/devices/{id}/commands/{cid}/ack` | Mark command completed/failed | `Gr33nApiClient.ack_command` |
| `POST` | `/farms/{id}/fertigation/mixing-events` | Audit row after automated mix | `Gr33nApiClient.post_mixing_event` |
| `DELETE` | `/devices/{id}/pending-command` | Legacy clear after `pending_command` fallback | `Gr33nApiClient.clear_pending_command` |
| `GET` | `/farms/{id}/devices` | List devices — legacy `pending_command` fallback | `Gr33nApiClient.get_devices` |

Request/response shapes for each of these are defined in [`openapi.yaml`](../openapi.yaml). The Pi client struct-matches those shapes one-for-one.

### Single reading

```http
POST /sensors/42/readings
X-API-Key: <PI_API_KEY>
Content-Type: application/json

{
  "sensor_id": 42,
  "value_raw": 22.5,
  "reading_time": "2026-03-03T10:00:00+00:00",
  "is_valid": true
}
```

### Batch readings (preferred for flush and high-frequency sensors)

```http
POST /sensors/readings/batch
X-API-Key: <PI_API_KEY>
Content-Type: application/json

[
  {"sensor_id": 1, "value_raw": 22.5, "reading_time": "2026-03-03T10:00:00+00:00", "is_valid": true},
  {"sensor_id": 2, "value_raw": 58.1, "reading_time": "2026-03-03T10:00:01+00:00", "is_valid": true},
  {"sensor_id": 3, "value_raw": 1.42, "reading_time": "2026-03-03T10:00:02+00:00", "is_valid": true}
]
```

The batch endpoint is also how the offline queue drains: on reconnect the client calls `pop_batch(50)` from SQLite, posts the whole array, and `ack()`s only on `2xx`. On `5xx` or network failure the rows stay queued, which is covered by the unit tests `test_post_readings_batch_success` and `test_batch_posts_to_correct_path_and_handles_500` in [`pi_client/test_gr33n_client.py`](../pi_client/test_gr33n_client.py).

### Actuator event + clear

When the Pi sees `config.pending_command` on a device and executes it:

```http
POST /actuators/7/events
{
  "event_type": "actuator_on",
  "source": "schedule_trigger",
  "schedule_id": 12
}

DELETE /devices/7/pending-command
```

The schedule ID threads the event back to the automation run that triggered it, so the Schedules page can show **"actuator X went on at HH:MM:SS because of schedule Y"** under the `/schedules/{id}/actuator-events` endpoint (JWT-protected, consumed by the UI).

---

## 4. How the data reaches the UI

1. **`POST /sensors/{id}/readings`** (or `/batch`) writes a row into `gr33ncore.sensor_readings`. It also:
   - evaluates alert rules → inserts into `gr33ncore.sensor_alerts` if thresholds are breached → emits push notifications via the registered FCM tokens.
   - pushes the reading onto the SSE channel for that farm.
2. **Dashboard** subscribes to `GET /farms/{id}/sensors/stream` (SSE) and renders live values without polling. The latest value per sensor also feeds the Zone cards via `GET /sensors/{id}/readings/latest`.
3. **Zone detail & Monitor views** chart history via `GET /sensors/{id}/readings` and `GET /sensors/{id}/readings/stats`.
4. **Schedules page** shows every actuator event the Pi posted via `GET /schedules/{id}/actuator-events` — the audit trail of what hardware did.
5. **Devices page** shows a device as **online** when its latest `PATCH /devices/{id}/status` is recent; **offline** otherwise.
6. **Alerts** bell in the TopBar is hydrated from `GET /farms/{id}/alerts/unread-count`; clicking shows the list from `GET /farms/{id}/alerts`. Both are driven entirely by readings the Pi posted.

No UI view polls the Pi directly — the API is always the middle layer, which keeps auth, validation, and alerting in one place.

---

## 5. Testing the integration

1. **On the Pi host (or laptop):** set `api.base_url` and `api.api_key` in `pi_client/config.yaml`, then `python3 pi_client/gr33n_client.py --config config.yaml` — hardware calls are stubbed if no GPIO is present.
2. **Unit tests** (no hardware, no live API):
   ```bash
   cd pi_client
   python3 -m pytest test_gr33n_client.py -v
   ```
   Covers reading post, batch post + empty-list short-circuit, 500 handling, device status heartbeat, actuator event dispatch, pending-command clear, and the offline queue.
3. **End-to-end smoke against a dev API:**
   - Start the API in `AUTH_MODE=auth_test` with `PI_API_KEY=devkey`.
   - `curl -H "X-API-Key: devkey" -H "Content-Type: application/json" \
     -d '[{"sensor_id":1,"value_raw":22.5,"is_valid":true}]' \
     http://localhost:8080/sensors/readings/batch` → **201**.
   - The reading should appear in the Dashboard under the farm's sensor stream within ~1s.

---

## 6. Operational tips

- **API key rotation:** see [§7](#7-pi-api-key-security-middleware-and-least-privilege) for an ordered runbook. Short form: update `PI_API_KEY` on the API and every edge (`config.yaml`, bridge env, systemd drop-ins), restart; Pis requeue on `401` until the key matches — queued readings are not dropped.
- **Clock skew:** `reading_time` is generated by the Pi in UTC. If a Pi's clock is wrong, charts look wrong; NTP is recommended.
- **Backpressure:** the offline queue is unbounded SQLite. For farms with many high-frequency sensors, lower `offline_flush_interval_seconds` and rely on `/sensors/readings/batch` to drain efficiently.
- **Multiple Pis, one farm:** each Pi uses the same `PI_API_KEY` and the same `farm_id`; sensor IDs must be unique across the farm. The API has no per-Pi identity beyond the device rows the Pi heartbeats.

## 7. Pi API key security, middleware, and least privilege

This section maps **what the code actually enforces** (`cmd/api/auth.go`, `internal/farmauthz/farmauthz.go`, handlers) so operators can rotate keys and reason about exposure. It is not a substitute for network TLS and LAN segmentation.

### Middleware (edge HTTP)

| Middleware | When it runs | Effect |
|------------|----------------|--------|
| **`requireAPIKey`** | Pi-only routes in `routes.go` (`POST` sensor readings, `PATCH` device status, `POST` actuator events, `DELETE` pending-command, …) | If not `AUTH_MODE=dev` bypass: requires header **`X-API-Key`** equal to process env **`PI_API_KEY`**. Missing key → **401** (`X-API-Key required`). Wrong key → **403** (`invalid API key`). On success, sets **`PiEdgeAuth`** on the request context (`internal/authctx`). |
| **`requireJWTOrPiEdge`** | `GET /farms/{id}/devices` only | Accepts **either** a valid dashboard **JWT** (`Authorization: Bearer …`) **or** the same **`X-API-Key`** as above. API key path also sets **`PiEdgeAuth`**. |

### Farm and resource checks (handlers)

| Route (edge) | After middleware | Additional check |
|----------------|------------------|------------------|
| `POST /sensors/{id}/readings`, `POST /sensors/readings/batch` | `requireAPIKey` | **None** — any holder of `PI_API_KEY` can write readings for **any numeric `sensor_id`**. Farm membership is **not** re-checked here. |
| `PATCH /devices/{id}/status`, `DELETE /devices/{id}/pending-command` | `requireAPIKey` | **None** — valid key can update/clear **any device id** it knows. |
| `POST /actuators/{id}/events` | `requireAPIKey` | **`RequireFarmMemberOrPiEdge`** for the **actuator's farm** (`internal/handler/actuator/handler.go`). If context has **`PiEdgeAuth`**, membership is **skipped** (same trust as ingest: key + knowing resource IDs). |
| `GET /farms/{id}/devices` | `requireJWTOrPiEdge` | **`RequireFarmMemberOrPiEdge`** for **`farm_id` in the URL** (`internal/handler/device/handler.go`). **`PiEdgeAuth`** bypasses JWT membership for that farm id. |

So today **`PI_API_KEY` is a single shared farm-network secret**, not a per-device credential in the database. **Least privilege in practice:** keep the key only on trusted hosts (Pi, bridge PC), prefer **MQTT → bridge** so microcontrollers never hold it, use **TLS** to the API, firewall the API from the public internet, and restrict who can read deployment secrets (systemd, Kubernetes Secret, `.env` permissions).

### Rotation runbook (zero code change)

1. **Generate** a new long random secret (password manager or `openssl rand -hex 32`).
2. **Apply on the API:** set `PI_API_KEY` to the new value in the API environment, **restart** the API process. Until edges update, they will see **401/403** on Pi routes — expected.
3. **Rolling update edges:** for each Pi, MQTT bridge, or test laptop, update `api.api_key` / `PI_API_KEY` / `MQTT_BRIDGE_API_KEY` to the **same** new value, then restart that unit. Order can be “API first, then edges”; the Pi client **requeues** failed posts in SQLite until the key matches.
4. **Verify:** `GET /auth/mode` (public), then one **`X-API-Key`** call (e.g. `PATCH /devices/{id}/status`) returns **200**; dashboard JWT login still works unchanged (`JWT_SECRET` is separate).
5. **Revoke old material:** delete the previous secret from config repos, shell history, and chat logs.

`AUTH_MODE=dev` (with a **`-tags dev`** build) bypasses key/JWT checks for local development — **never** use that combination on an internet-exposed host.

---

## 8. Field checklist — first Pi on a real bench (Phase 31 WS2)

**Goal:** An operator with a Raspberry Pi knows **exactly** what to wire and configure first — before mains-powered pumps or warehouse-scale rollouts.

**Prerequisites:** Complete the **laptop stub loop** once ([`local-operator-bootstrap.md` — Edge loop in 5 commands](local-operator-bootstrap.md)) so API, `PI_API_KEY`, and sensor IDs are proven before GPIO.

### 8.1 Safety (read before GPIO)

| Rule | Why |
|------|-----|
| **No mains AC on a breadboard** | Bench work is **3.3 V / 5 V logic only** — DHT22, I2C, **3.3 V relay modules**, **LED + resistor** for actuator proof. |
| **Mains loads through rated enclosures** | Pumps, contactors, and line-voltage lighting belong in **listed relay panels** with fuses, strain relief, and physical E-stop — not on the Pi desk. |
| **Fail-safe defaults** | Many optocoupler relay boards are **active-LOW** (gr33n client uses `active_high=False`). Power loss should **de-energize** loads that can flood or overheat — **operator wiring responsibility**; software v1 does not cut GPIO on comms loss. |
| **One relay / one load first** | Prove **`pending_command`** with an **LED or small 5 V fan** before a solenoid or pump (Phase 31 WS3). |
| **E-stop story** | Keep a manual way to cut power to the load under test; document who is allowed to Confirm Guardian actuator PRs on production farms. |

gr33n documents safe paths; it does **not** certify hardware. See also [`recommended-hardware-and-sizing.md`](recommended-hardware-and-sizing.md) (edge vs central server).

### 8.2 Zone naming — one plastic room, three tiers (example)

For a single **plastic grow room** with **three vertical shelves**, pick **one convention** and keep sensor/actuator IDs aligned in [`pi_client/config.yaml`](../pi_client/config.yaml) and the dashboard:

| Physical layout | gr33n mapping (recommended for independent EC/light per tier) |
|-----------------|------------------------------------------------------------------|
| Room envelope | One **`farm_id`** (e.g. demo **gr33n Demo Farm** = `1`) |
| Bottom / middle / top shelf | Three **`zones`**, e.g. `Room A — Tier 1`, `Room A — Tier 2`, `Room A — Tier 3` |
| One Pi per room | One **`devices`** row + actuators per relay channel; sensors tagged to tier **`zone_id`** |
| Alternative (single zone) | One zone + three **`sensor_id`** rows (PAR/temp/humidity per shelf) — simpler UI, less per-tier automation |

Enterprise-scale naming patterns: [`hypothetical-enterprise-topology.md`](hypothetical-enterprise-topology.md) (*Plastic grow room → zones*). **MQTT room-scale ingest:** same three zones can publish on `gr33n/farm/{farm_id}/zone/{zone_id}/sensor/{id}` — [`mqtt-edge-operator-playbook.md`](mqtt-edge-operator-playbook.md#room-scale-warehouse-pattern-phase-31-ws4).

Create zones in the dashboard (**Zones**) or via seed/template **before** wiring sensors. Prefer **§2a**: set wiring in the UI, download config, then deploy to the Pi. For manual YAML, run [`scripts/print-demo-sensor-ids.sh`](../scripts/print-demo-sensor-ids.sh) or `./scripts/run-edge-stub-client.sh` on a laptop to confirm **`sensor_id`** values match DB names.

#### Wiring sketch — room layout (logical, not electrical)

One **plastic room**, **one Pi**, **three shelf tiers** — sensors and relays per tier (names illustrative):

```
                    ┌── Plastic room (farm_id=1) ──────────────────┐
                    │  Tier 3  zone "Room A — Tier 3"               │
                    │    [DHT22] [PAR]     relay CH3 → light strip  │
                    │  ───────────────────────────────────────────  │
                    │  Tier 2  zone "Room A — Tier 2"               │
                    │    [DHT22] [PAR]     relay CH2 → light strip  │
                    │  ───────────────────────────────────────────  │
                    │  Tier 1  zone "Room A — Tier 1"               │
                    │    [DHT22] [EC/pH]   relay CH1 → pump (rated) │
                    │                                               │
                    │  Raspberry Pi (device_uid) — edge_gateway     │
                    │    I2C/SPI → sensors   GPIO → relay module    │
                    └───────────────────────────────────────────────┘
                              │ HTTP + X-API-Key
                              ▼
                    LAN server (Postgres + API + UI)
```

Wire **one channel first** (§8.6) before copying this layout to all three tiers.

#### Wiring sketch — bench actuator proof (LED, no mains)

Safe first GPIO test — matches default **`gpio_pin: 17`** in [`pi_client/config.yaml`](../pi_client/config.yaml). Client uses **`active_high=False`** (common on optocoupler relay boards).

**Option A — LED directly (simplest):**

```
Pi BCM 17 ──[330 Ω]──►|── LED ──► GND
```

**Option B — 3.3 V relay module (prep for WS3):**

```
Pi 3.3 V ──► relay VCC          Pi GND ──► relay GND
Pi BCM 17 ──► relay IN          (module switches separate 5 V load — fan/LED on COM/NO)
```

Do **not** wire mains AC to the breadboard. Rated enclosures only for line-voltage loads.

### 8.3 Copy-paste field checklist

Use this on first deploy; tick items in order.

```
[ ] 0. Laptop stub loop OK — make edge-smoke-help; dashboard Live Sensors show values
[ ] 1. Server — API reachable on LAN (https:// or http://); AUTH_MODE=production or auth_test;
       PI_API_KEY set on API; JWT_SECRET set; firewall allows Pi → API port only (not Postgres to internet)
[ ] 2. Pi OS — Raspberry Pi OS (64-bit); SSH; NTP/chrony (correct reading_time UTC)
[ ] 3. Packages — from repo on Pi: ./scripts/install-pi-edge-deps.sh
       (see docs/raspberry-pi-and-deployment-topology.md §2)
[ ] 4a. Pi config (DB-first, Phase 50) — in device wizard step 2: set wiring on sensors/actuators,
       Download config.yaml, scp to Pi pi_client/config.yaml; set api.api_key on the Pi only
[ ] 4b. Pi config (manual fallback) — cd pi_client && ./setup.sh; hand-edit config.yaml:
       api.base_url = http://<api-lan-ip>:8080
       api.api_key  = <same as server PI_API_KEY>
       farm.farm_id = <your farm>
       sensors[]    = sensor_id values from DB (print-demo-sensor-ids.sh)
       actuators[]  = device_id / actuator_id from dashboard Devices (after bench device exists)
[ ] 5. systemd — sudo systemctl enable --now gr33n; journalctl -u gr33n -f
[ ] 6. Readings — dashboard Live Sensors update within ~1 interval; Devices show online after heartbeat
[ ] 7. Offline queue drill — stop API 2 min; client logs queue; start API; readings flush via batch (§8.5)
[ ] 8. One-relay safe test — LED on GPIO via pending_command (§8.6); then consider real load in enclosure
[ ] 9. Contract smokes green — make test (TestPiContract*) on CI/dev DB documents same HTTP shapes (§8.7)
```

### 8.4 Power, network, and secrets

| Topic | Check |
|-------|--------|
| **Pi power** | Official **5 V / 3 A+** PSU (USB-C Pi 4/5); avoid undersized phone chargers under sensor + Wi-Fi load. |
| **Relay module** | **3.3 V logic**, separate **5 V** relay coil supply if the board requires it; common ground with Pi. |
| **`PI_API_KEY`** | Same secret on API **and** `pi_client/config.yaml` (or MQTT bridge env). Rotate per [§7](#7-pi-api-key-security-middleware-and-least-privilege). |
| **LAN firewall** | Pi → **API port only**; do not expose Postgres or Ollama to the field VLAN unless intentional. |
| **TLS** | Production: reverse proxy (Caddy/nginx) with TLS in front of API; set `api.base_url` to `https://…`. |
| **Wi-Fi vs Ethernet** | Prefer **Ethernet** on greenhouse edges; if Wi-Fi, expect offline queue use during drops. |

Topology ladder (edge-only Pi vs all-on-one-Pi): [`raspberry-pi-and-deployment-topology.md`](raspberry-pi-and-deployment-topology.md).

### 8.5 Offline queue drill (operator)

Confirms [`offline_queue_path`](../pi_client/gr33n_client.py) behavior before you rely on it in the field.

1. Pi client running; dashboard shows live readings.
2. **Stop the API** (or block Pi → API with firewall) for **2–3 minutes**.
3. Confirm client still runs — readings accumulate in SQLite (`/var/lib/gr33n/queue.db` or path in config).
4. **Restore API**; within **`offline_flush_interval_seconds`** (default 60), queued rows post via **`POST /sensors/readings/batch`**.
5. Dashboard catches up; no duplicate manual steps required.

Unit coverage: `test_gr33n_client.py` (offline queue + batch flush).

### 8.6 One-relay safe test (bench, before WS3 E2E)

Minimal GPIO proof **without** Guardian or automation worker:

1. Wire **one LED + resistor** (or 3.3 V relay module input) to the GPIO pin in `config.yaml` `actuators[].gpio_pin` (default **BCM 17**).
2. Ensure a **`devices`** + **`actuators`** row exists in the dashboard for that `device_id` / `actuator_id`.
3. Enqueue **`pending_command`** — automation rule, confirmed Guardian PR, or bench script (§9).
4. Pi polls **`GET /farms/{id}/devices`** with **`X-API-Key`** → sees **`config.pending_command`** (base64 JSON in API responses — client decodes automatically).
5. Client executes → **`POST /actuators/{id}/events`** → **`DELETE /devices/{id}/pending-command`**.

Full E2E walkthrough: **§9** below.

### 8.7 API contract smokes (same shapes as the Pi)

Run on a dev machine with Postgres + seed; documents the HTTP contract the Pi must speak:

```bash
export DATABASE_URL='postgres://gr33n:gr33n@127.0.0.1:5433/gr33n?sslmode=disable'
make test   # includes cmd/api — tags dev
# Or narrow:
go test -tags dev ./cmd/api/ -run TestPiContract
```

| Test | What it proves for field ops |
|------|------------------------------|
| `TestPiContractScheduleAndProgramFeedback` | `pending_command` with schedule + program → Pi POST event with provenance |
| `TestPiContractRuleFeedback` | Rule-triggered pending → actuator event + audit fields |
| `TestPiContractRejectRuleAndProgramTogether` | Invalid dual provenance rejected (Pi should not send both) |
| `TestPiContractRejectUnknownSchedule` | Bad schedule id rejected |

Source: [`cmd/api/smoke_pi_contract_test.go`](../cmd/api/smoke_pi_contract_test.go).

Python client contract (no DB): `cd pi_client && python3 -m pytest test_gr33n_client.py -v`.

### 8.8 When something fails

| Symptom | Check |
|---------|--------|
| **401 / 403** on readings | `PI_API_KEY` mismatch; API restarted after `.env` change? |
| **404** on sensor post | Wrong **`sensor_id`** — run `print-demo-sensor-ids.sh` |
| **Devices stay offline** | Heartbeat uses `device_id` from actuators config; PATCH path reachable? |
| **pending_command never clears** | `GET /farms/{id}/devices` returns config? Actuator `device_id` matches config? |
| **Command “lost” after automation + manual test** | Use **FIFO queue** (`GET …/commands/next`) — not concurrent `pending_command` writes alone |
| **Pulse ignored** | Confirm `duration_seconds` in pending JSON; pump/relay actuator type; check Pi log for pulse thread |
| **Queue grows forever** | API down or batch POST failing — see §8.5 (sensor offline SQLite queue, not actuator command queue) |

Broader ops: [`operator-troubleshooting.md`](operator-troubleshooting.md), [`mqtt-edge-operator-playbook.md`](mqtt-edge-operator-playbook.md) if using MQTT instead of direct HTTP.

---

## 9. Safe actuator E2E — pending_command round-trip (Phase 31 WS3)

**Goal:** Prove one relay or LED path end-to-end: something enqueues **`devices.config.pending_command`** → **`pi_client`** polls → GPIO (or stub) → **`POST /actuators/{id}/events`** → **`DELETE /devices/{id}/pending-command`**.

This matches [`TestPiContract*`](../../cmd/api/smoke_pi_contract_test.go) and Phase 30 **`enqueue_actuator_command`** after Confirm.

### 9.1 Preconditions

| Item | Why |
|------|-----|
| API up with **`PI_API_KEY`** | Pi routes need matching key ([§7](#7-pi-api-key-security-middleware-and-least-privilege)) |
| **`demo-veg-relay-01`** seeded | Master seed device + **Veg Room Grow Light** actuator ([`print-demo-actuator-ids.sh`](../scripts/print-demo-actuator-ids.sh)) |
| **`AUTOMATION_SIMULATION_MODE=false`** (optional) | Simulation mode skips real **`pending_command`** enqueue from the worker — use direct/G Guardian enqueue for bench |
| Safety read | [§8.1](#81-safety-read-before-gpio) + [operator-troubleshooting §5](operator-troubleshooting.md#5-edge-actuator-safety-phase-31-ws3) |

### 9.2 Automated smoke (laptop or Pi)

Terminal 1 — API (if not already running):

```bash
make dev-auth-test
```

Terminal 2 — one-shot E2E (starts pi_client, enqueues, verifies event):

```bash
./scripts/run-edge-actuator-smoke.sh --direct
# Guardian PR path (Confirm → pending_command):
./scripts/run-edge-actuator-smoke.sh --guardian
```

Success prints the latest **`gr33ncore.actuator_events`** row and confirms **`pending_command`** was cleared.

### 9.3 Manual two-terminal bench (real GPIO)

**Terminal A** — actuator-only client (polls every 5s by default):

```bash
./scripts/run-edge-actuator-client.sh
# Optional: GR33N_GPIO_PIN=17 GR33N_SCHEDULE_POLL_SECONDS=5
```

**Terminal B** — enqueue after client is running:

```bash
./scripts/enqueue-demo-pending-command.sh on
# Turn off:
./scripts/enqueue-demo-pending-command.sh off
# Clear without executing:
./scripts/enqueue-demo-pending-command.sh --clear
```

**Verify:**

- Client log: `Executing scheduled command 'on' for device_id=…`
- Dashboard **Devices** → device stays **online**; pending clears on next poll
- DB or API: `GET /actuators/{id}/events` shows new row with `reported_by: pi_client` in `meta_data`

### 9.4 Guardian PR path (production-shaped)

1. Dashboard → **Guardian** → ask to turn on a grow light (or open an actuator PR from inbox).
2. **Confirm** the high-tier proposal — executes **`enqueue_actuator_command`** → writes **`pending_command`** with `source: guardian` and `proposal_id`.
3. Pi client running on the bench picks it up on the next **`schedule_poll_interval_seconds`** tick.
4. Event posts with **`source: manual_api_call`** and `meta_data.proposal_id` for audit join back to the PR.

Automated equivalent: `./scripts/run-edge-actuator-smoke.sh --guardian`.

### 9.5 Flow diagram

```
  Automation worker / Guardian Confirm / bench script
              │
              ▼
   devices.config.pending_command  (JSON: command, actuator_id, optional duration_seconds, …)
              │
              ▼
   pi_client GET /farms/{id}/devices  (X-API-Key)
              │
              ▼
   GPIO execute (instant or pulse on→off)  — stubbed off-Pi
              │
              ├── POST /actuators/{id}/events
              └── DELETE /devices/{id}/pending-command
```

Help text only: **`make edge-actuator-smoke-help`**.

