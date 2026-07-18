# MQTT and field edge — operator playbook

Phase 14 **WS1** documents how MQTT-based and microcontroller field gear fits the existing gr33n model: **devices**, **sensors**, **sensor readings**, and **pending_command** tasking. The API stays vendor-neutral; brokers are interchangeable.

## Goals

- **No cloud lock-in**: self-hosted [Eclipse Mosquitto](https://mosquitto.org/) or any MQTT 3.1.1/5 broker; managed options (HiveMQ, AWS IoT Core, etc.) are alternate deployments of the same pattern.
- **Reuse existing HTTP contract**: bridges translate MQTT messages into `POST /sensors/{id}/readings`, **`POST /sensors/readings/batch`**, and poll **`GET /farms/{id}/devices`** for automation tasking.
- **Security at the boundary**: field MQTT is usually authenticated (username/password, TLS client certs, or broker ACLs). The **bridge** holds `PI_API_KEY` and talks to gr33n over HTTPS; tiny MCUs never need the farm API secret if they only speak MQTT.

## Reference architecture

```text
[MCU / sensor] --MQTT (TLS)--> [Broker] --subscribe--> [Bridge on Pi / PC]
                                                           |
                                                           +--> HTTPS + X-API-Key gr33n-api
```

1. **Microcontrollers** publish telemetry to MQTT topics (short payloads, QoS 1 where you need at-least-once).
2. **Bridge** (Python, Node, or Go on a Pi/edge PC) subscribes, validates, maps topic or payload → `sensor_id`, calls the API.
3. **Tasking**: automation writes `config.pending_command` on a device row. The **same bridge** (or the existing Pi client) polls **`GET /farms/{farm_id}/devices`** with **`X-API-Key: PI_API_KEY`**, executes the command locally (GPIO, relay, downstream MQTT publish), then **`DELETE /devices/{id}/pending-command`** and **`POST /actuators/{id}/events`** with provenance. The Go API JSON-encodes `devices.config` (`[]byte`) as a **base64 string**; [`pi_client/gr33n_client.py`](../pi_client/gr33n_client.py) decodes that automatically in `_schedule_loop`. Echo **`schedule_id`**, **`rule_id`**, and **`program_id`** from the pending JSON into the event body (`triggered_by_schedule_id`, `triggered_by_rule_id`, `program_id`) so `gr33ncore.actuator_events` rows stay auditable — see smoke tests `TestPiContract*` in `cmd/api/smoke_pi_contract_test.go`.

## Broker choice: Mosquitto vs managed

| Concern | Self-hosted Mosquitto | Managed MQTT (generic) |
|--------|------------------------|---------------------------|
| **Cost / data residency** | You operate; data stays on-farm | Vendor SLA; check regions |
| **Auth** | `password_file`, TLS, ACL patterns | Often IAM / JWT / per-device certs |
| **Ops** | Backups, upgrades, monitoring | Less infra; vendor limits apply |

gr33n does **not** embed a broker; operators pick one and configure the bridge.

## Suggested topic layout (convention, not enforced)

Two layouts are supported by [`mqtt_telemetry_bridge.py`](../pi_client/mqtt_telemetry_bridge.py) (`MQTT_TOPIC_LAYOUT`):

**Device gateway (default)** — one edge host per `device_uid`:

```text
gr33n/<farm_id>/<device_uid>/telemetry/<sensor_slug|sensor_id>
gr33n/<farm_id>/<device_uid>/cmd/<name>          # optional: downstream commands from bridge to MCU
```

**Room-scale warehouse (Phase 31 WS4)** — zone-centric multi-shelf rooms:

```text
gr33n/farm/<farm_id>/zone/<zone_id>/sensor/<sensor_id_or_slug>
```

See [Room-scale warehouse pattern](#room-scale-warehouse-pattern-phase-31-ws4) below for env, batching, and ACL examples.

- **`device_uid`** matches `gr33ncore.devices.device_uid` when possible.
- Payloads should stay small. Example JSON: `{"v":22.5,"t":"2026-04-16T12:00:00Z"}` or even a single float string for ultra-constrained nodes.

Bridges map `sensor_id` (integer primary key in `gr33ncore.sensors`) before calling the API.

- **Farm id in the topic** must match **`GR33N_FARM_ID`** on the bridge. The reference bridge **drops** messages where the topic’s `farm_id` segment differs (defense against misconfigured publishers).

## Reference implementation (in-repo)

Maintained script: **[`pi_client/mqtt_telemetry_bridge.py`](../pi_client/mqtt_telemetry_bridge.py)** (`pip install -r pi_client/requirements.txt` includes `paho-mqtt`).

| Artifact | Purpose |
|----------|---------|
| [`mqtt_bridge_map.example.yaml`](../pi_client/mqtt_bridge_map.example.yaml) | YAML `(device_uid, slug) → sensor_id` |
| [`mqtt_bridge_map.room-scale.example.yaml`](../pi_client/mqtt_bridge_map.room-scale.example.yaml) | Phase 31 WS4 — `zone_sensor_map` for room layout |
| [`mqtt-bridge.example.env`](../pi_client/mqtt-bridge.example.env) | Environment template for production |
| [`mqtt-bridge.room-scale.example.env`](../pi_client/mqtt-bridge.room-scale.example.env) | Room-scale warehouse env template (WS4) |
| [`mqtt-bridge.example.service`](../pi_client/mqtt-bridge.example.service) | systemd unit sketch |

Run (development):

```bash
export GR33N_API_URL=http://127.0.0.1:8080 GR33N_FARM_ID=1 PI_API_KEY=...
export MQTT_HOST=127.0.0.1 MQTT_PORT=1883
python3 pi_client/mqtt_telemetry_bridge.py --sensor-map pi_client/mqtt_bridge_map.example.yaml
```

Optional **`MQTT_BATCH_MS`**: coalesce readings for that many milliseconds (still caps at **64** readings per HTTP request). **`MQTT_TOPIC_PREFIX`**: defaults to `gr33n` if you need a different first path segment. **`MQTT_TOPIC_LAYOUT`**: `device` (default) or `room` for warehouse zone topics (Phase 31 WS4).

Unit tests (no broker): `python3 -m pytest pi_client/test_mqtt_telemetry_bridge.py -v`

---

## Room-scale warehouse pattern (Phase 31 WS4)

**Goal:** A plastic room with **many MCUs** (one per shelf tier or sensor cluster) publishes telemetry without custom HTTP code on each node — one **bridge** batches into gr33n.

### When to use which topic layout

| Layout | Topic example | Best for |
|--------|---------------|----------|
| **device** (default) | `gr33n/1/gw-a/telemetry/temp` | Pi gateway, single `device_uid`, slug map |
| **room** (WS4) | `gr33n/farm/1/zone/12/sensor/101` | Warehouse shelves mapped to **`zone_id`**; ACL per zone |

Both layouts hit the same API: **`POST /sensors/readings/batch`** (max **64** readings per request — see `maxBatchReadings` in `internal/handler/sensor/handler.go`).

### Room topic convention

```text
<prefix>/farm/<farm_id>/zone/<zone_id>/sensor/<sensor_id_or_slug>
```

- **`farm_id`** must match **`GR33N_FARM_ID`** on the bridge (misconfigured publishers are dropped).
- **`zone_id`** matches `gr33ncore.zones.id` — aligns with Phase 31 three-tier room naming ([`pi-integration-guide.md` §8.2](pi-integration-guide.md#82-zone-naming--one-plastic-room-three-tiers-example)).
- **`sensor_id_or_slug`**: if all digits, treated as `gr33ncore.sensors.id`; otherwise lookup **`zone_sensor_map`** in YAML.

Example MCU publish (QoS 1 recommended):

```text
Topic: gr33n/farm/1/zone/2/sensor/par
Payload: {"v": 420.0, "t": "2026-05-27T12:00:00Z"}
```

Or publish numeric sensor id directly (no YAML row required):

```text
Topic: gr33n/farm/1/zone/2/sensor/1
Payload: 380.5
```

### Load and batching (not a performance guarantee)

Illustrative math for ops planning only:

| Scenario | Raw MQTT msg/s | With `MQTT_BATCH_MS=250` |
|----------|----------------|---------------------------|
| 3 zones × 4 sensors × 1 Hz | 12 | ~4–5 HTTP batch POSTs/s (well under API limits) |
| 20 zones × 10 sensors × 1 Hz | 200 | ~8+ POSTs/s — tune batch window; watch API CPU |

Rules of thumb:

- Set **`MQTT_BATCH_MS=250`** (or 500) on busy rooms so the bridge coalesces before HTTP.
- Each batch caps at **64** readings; overflow spills to the next POST automatically.
- gr33n does **not** ship a broker — Mosquitto, HiveMQ, AWS IoT Core, etc. are interchangeable.

### Example env block (room layout)

Copy [`pi_client/mqtt-bridge.room-scale.example.env`](../pi_client/mqtt-bridge.room-scale.example.env):

```bash
export GR33N_API_URL=http://192.168.1.50:8080
export GR33N_FARM_ID=1
export PI_API_KEY=your-shared-edge-secret
export MQTT_HOST=127.0.0.1
export MQTT_PORT=1883
export MQTT_TOPIC_LAYOUT=room
export MQTT_TOPIC_PREFIX=gr33n
export MQTT_SENSOR_MAP_PATH=pi_client/mqtt_bridge_map.room-scale.example.yaml
export MQTT_BATCH_MS=250

python3 pi_client/mqtt_telemetry_bridge.py \
  --topic-layout room \
  --sensor-map "$MQTT_SENSOR_MAP_PATH"
```

YAML map: [`mqtt_bridge_map.room-scale.example.yaml`](../pi_client/mqtt_bridge_map.room-scale.example.yaml) (`zone_sensor_map` rows).

### Broker ACL sketch (Mosquitto)

Per-shelf MCU credentials publish **only** to their zone:

```text
user tier2-mcu
topic write gr33n/farm/1/zone/2/sensor/#
```

Bridge service account subscribes:

```text
user gr33n-bridge
topic read gr33n/farm/1/zone/+/sensor/+
```

### Actuator / tasking (unchanged)

MQTT WS4 covers **telemetry ingest** only. **`pending_command`** still flows through **`GET /farms/{id}/devices`** on the Pi client or a bridge extension — see [Tasking and schedules](#tasking-and-schedules) above and Phase 31 WS3 ([`pi-integration-guide.md` §9](pi-integration-guide.md#9-safe-actuator-e2e--pending_command-round-trip-phase-31-ws3)).

Enterprise topology context: [`hypothetical-enterprise-topology.md`](hypothetical-enterprise-topology.md).

---

## Security checklist

| Practice | Why |
|----------|-----|
| **No anonymous MQTT** in production | Prevents neighborhood publishes into your telemetry namespace |
| **TLS to broker** (`MQTT_USE_TLS`, optional `MQTT_CA_FILE` on the bridge) | Protects credentials and payload on the wire |
| **Broker ACLs** | Limit each device credential to `gr33n/<its_farm_id>/<its_uid>/telemetry/#` (pattern varies by broker) |
| **Rotate `PI_API_KEY`** | Compromised bridge key can POST readings for any known `sensor_id` and list devices for any `farm_id` |
| **Segment broker from WAN** | MQTT often stays on farm LAN; only the bridge needs HTTPS egress to the API |
| **Least privilege OS user** for the bridge service | Limits blast radius if the host is compromised |

**Mosquitto (illustrative):** use `allow_anonymous false`, `password_file` or TLS client certs, and `acl_file` patterns per device user. Exact syntax depends on Mosquitto version; test ACLs with `mosquitto_pub` before rolling MCU firmware.

## HTTP endpoints (edge)

| Method | Path | Purpose |
|--------|------|---------|
| `POST` | `/sensors/{id}/readings` | Single reading (existing Pi client) |
| `POST` | `/sensors/readings/batch` | Up to **64** readings in one transaction; preferred for bursty MQTT fan-in |
| `GET` | `/farms/{id}/devices` | **JWT** (dashboard) **or** **`X-API-Key`** (edge): list devices and read `config.pending_command` |
| `DELETE` | `/devices/{id}/pending-command` | Clear command after execution (existing) |
| `PATCH` | `/devices/{id}/status` | Heartbeat / status (existing) |

Environment variable: **`PI_API_KEY`** on the API; edge clients send **`X-API-Key`** (see [`openapi.yaml`](../openapi.yaml)).

**Trust model:** holding `PI_API_KEY` allows posting readings for any `sensor_id` and listing devices for any `farm_id` you request—same as knowing UUIDs/API surface today. Rotate the key if a bridge is compromised; prefer network segmentation between broker and bridge.

**Middleware vs handlers:** `requireAPIKey` / `requireJWTOrPiEdge` live in `cmd/api/auth.go`; farm bypass for the key is `RequireFarmMemberOrPiEdge` in `internal/farmauthz/farmauthz.go`. Exact matrix (which routes re-check farm membership, rotation steps, least-privilege posture) is in [`pi-integration-guide.md`](pi-integration-guide.md#7-pi-api-key-security-middleware-and-least-privilege).

## Tasking and schedules

- **Automation worker** sets `pending_command` on the target device when a schedule fires (see README schedule-loop overview).
- Edge code must poll **`GET /farms/{farm_id}/devices`** with the API key, **base64-decode** each device’s `config` field from the JSON list (then parse JSON) to read **`pending_command`**, execute, then clear it. Same rule as the Pi daemon — the HTTP response uses a base64 string for `config` bytes, not an inline JSON object ([`workflow-guide.md`](workflow-guide.md#field-edge-troubleshooting-for-pi-and-mqtt)).
- Human **tasks** (`/farms/{id}/tasks`) remain dashboard-centric for now; MQTT hooks are aimed at **device/actuator** execution paths.

## Troubleshooting

| Symptom | Likely cause |
|---------|----------------|
| Bridge logs “dropping message: topic farm_id≠…” | Publisher path uses wrong farm segment; fix topic or `GR33N_FARM_ID` |
| HTTP **401/403** from API | Missing/wrong `X-API-Key`; API must have `PI_API_KEY` set (not `AUTH_MODE=dev` bypass on prod) |
| Readings missing | Unmapped `(device_uid, slug)` — add YAML row or publish numeric slug equal to `sensor_id` |
| Bursts overload HTTP | Raise `MQTT_BATCH_MS` so the bridge coalesces into fewer `POST /sensors/readings/batch` calls |
| `config` looks like gibberish in `curl` / logs | Expected — value is **base64**; decode to JSON before looking for `pending_command` |
| Automation fires but actuator events show no schedule/rule | Bridge must echo **`triggered_by_schedule_id`**, **`triggered_by_rule_id`**, **`program_id`** from pending JSON into **`POST /actuators/{id}/events`** (cannot send **both** `triggered_by_rule_id` and `program_id`) |
| Wrong farm’s devices in response | `farm_id` in URL must match deployment; bridge env **`GR33N_FARM_ID`** must match topic layout |

## Related documents

- [`workflow-guide.md`](workflow-guide.md#field-edge-troubleshooting-for-pi-and-mqtt) — device tasking, base64 `config`, actuator provenance, edge troubleshooting (§2).
- [`pi-integration-guide.md`](pi-integration-guide.md#7-pi-api-key-security-middleware-and-least-privilege) — Pi + bridge auth, `PI_API_KEY` rotation, `requireAPIKey` / `RequireFarmMemberOrPiEdge` behavior.
- [`openapi.yaml`](../openapi.yaml) — machine-readable contracts.
- [`docs/phase-13-operator-documentation.md`](phase-13-operator-documentation.md) — operator index.
- Phase 14 plan: [`plans/archive/phase_14_network_and_commons.plan.md`](plans/archive/phase_14_network_and_commons.plan.md).
