# Pi Integration Guide (Pi → API → UI)

> **Scope:** How an on-farm Raspberry Pi running [`pi_client/gr33n_client.py`](../pi_client/gr33n_client.py) posts sensor readings, device status, and actuator events to the gr33n API, and how those flow into the dashboard UI.
>
> **Companion docs:**
> - API spec: [`openapi.yaml`](../openapi.yaml) — source of truth for every route used below.
> - MQTT edge playbook: [`mqtt-edge-operator-playbook.md`](mqtt-edge-operator-playbook.md) — when you run the MQTT → API bridge instead of (or alongside) direct HTTP.
> - Operator workflow narrative: [`workflow-guide.md`](workflow-guide.md) — how the pieces connect end-to-end.
> - **Hardware layout & scaling:** [`raspberry-pi-and-deployment-topology.md`](raspberry-pi-and-deployment-topology.md) — Pi OS packages, full stack on one Pi, splitting DB/API/UI onto servers or containers as the farm grows.

---

## 1. What the Pi does

The Pi client is a single-process Python program that:

1. **Reads sensors** on a per-sensor interval (temperature, humidity, EC, pH, CO₂, PAR, soil moisture). Hardware drivers are auto-stubbed on non-Pi hosts, so the exact same file runs on a dev laptop for tests.
2. **Posts readings** to the API (one-shot or batched).
3. **Reports device status** (online / offline / error) on a heartbeat.
4. **Polls for pending commands** (`GET /farms/{id}/devices` → `config.pending_command`) and executes them via GPIO.
5. **Reports actuator events** (what it actually did and when), then clears the pending command.
6. **Falls back offline** — any failed POST is stored in a local SQLite queue (`offline_queue_path`) and flushed later via the batch endpoint.

All API calls go over plain HTTP(S); there is no long-lived socket. Authentication is a pre-shared **API key** sent as the **`X-API-Key`** HTTP header on every request (same spelling the server reads in `cmd/api/auth.go`; header names are case-insensitive, but examples below use this form). The API validates it with **`requireAPIKey`** middleware (`cmd/api/routes.go`).

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

---

## 3. Routes used by the Pi

Every Pi-facing endpoint is declared with `requireAPIKey` in [`cmd/api/routes.go`](../cmd/api/routes.go):

| Method | Path | Purpose | Client method |
|--------|------|---------|---------------|
| `POST` | `/sensors/{id}/readings` | Post one reading for one sensor | `Gr33nApiClient.post_reading` |
| `POST` | `/sensors/readings/batch` | Post many readings across sensors in one request | `Gr33nApiClient.post_readings_batch` |
| `PATCH` | `/devices/{id}/status` | Heartbeat: `{"status": "online"}` | `Gr33nApiClient.patch_device_status` |
| `POST` | `/actuators/{id}/events` | Record what the actuator actually did | `Gr33nApiClient.post_actuator_event` |
| `DELETE` | `/devices/{id}/pending-command` | Clear a pending command after executing it | `Gr33nApiClient.clear_pending_command` |
| `GET` | `/farms/{id}/devices` | List devices (Pi or JWT) — Pi reads `config.pending_command` | `Gr33nApiClient.get_devices` |

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
