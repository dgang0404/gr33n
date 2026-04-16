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
3. **Tasking**: automation writes `config.pending_command` on a device row. The **same bridge** (or the existing Pi client) polls **`GET /farms/{farm_id}/devices`** with **`X-API-Key: PI_API_KEY`**, executes the command locally (GPIO, relay, downstream MQTT publish), then **`DELETE /devices/{id}/pending-command`** and records actuator events as today.

## Broker choice: Mosquitto vs managed

| Concern | Self-hosted Mosquitto | Managed MQTT (generic) |
|--------|------------------------|---------------------------|
| **Cost / data residency** | You operate; data stays on-farm | Vendor SLA; check regions |
| **Auth** | `password_file`, TLS, ACL patterns | Often IAM / JWT / per-device certs |
| **Ops** | Backups, upgrades, monitoring | Less infra; vendor limits apply |

gr33n does **not** embed a broker; operators pick one and configure the bridge.

## Suggested topic layout (convention, not enforced)

Use a stable prefix so ACLs are simple:

```text
gr33n/<farm_id>/<device_uid>/telemetry/<sensor_slug|sensor_id>
gr33n/<farm_id>/<device_uid>/cmd/<name>          # optional: downstream commands from bridge to MCU
```

- **`device_uid`** matches `gr33ncore.devices.device_uid` when possible.
- Payloads should stay small. Example JSON: `{"v":22.5,"t":"2026-04-16T12:00:00Z"}` or even a single float string for ultra-constrained nodes.

Bridges map `sensor_id` (integer primary key in `gr33ncore.sensors`) before calling the API.

- **Farm id in the topic** must match **`GR33N_FARM_ID`** on the bridge. The reference bridge **drops** messages where the topic’s `farm_id` segment differs (defense against misconfigured publishers).

## Reference implementation (in-repo)

Maintained script: **[`pi_client/mqtt_telemetry_bridge.py`](../pi_client/mqtt_telemetry_bridge.py)** (`pip install -r pi_client/requirements.txt` includes `paho-mqtt`).

| Artifact | Purpose |
|----------|---------|
| [`mqtt_bridge_map.example.yaml`](../pi_client/mqtt_bridge_map.example.yaml) | YAML `(device_uid, slug) → sensor_id` |
| [`mqtt-bridge.example.env`](../pi_client/mqtt-bridge.example.env) | Environment template for production |
| [`mqtt-bridge.example.service`](../pi_client/mqtt-bridge.example.service) | systemd unit sketch |

Run (development):

```bash
export GR33N_API_URL=http://127.0.0.1:8080 GR33N_FARM_ID=1 PI_API_KEY=...
export MQTT_HOST=127.0.0.1 MQTT_PORT=1883
python3 pi_client/mqtt_telemetry_bridge.py --sensor-map pi_client/mqtt_bridge_map.example.yaml
```

Optional **`MQTT_BATCH_MS`**: coalesce readings for that many milliseconds (still caps at **64** readings per HTTP request). **`MQTT_TOPIC_PREFIX`**: defaults to `gr33n` if you need a different first path segment.

Unit tests (no broker): `python3 -m pytest pi_client/test_mqtt_telemetry_bridge.py -v`

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

## Tasking and schedules

- **Automation worker** sets `pending_command` on the target device when a schedule fires (see README schedule-loop overview).
- Edge code must poll **`GET /farms/{farm_id}/devices`** with the API key, inspect each device’s `config`, act on `pending_command`, then clear it.
- Human **tasks** (`/farms/{id}/tasks`) remain dashboard-centric for now; MQTT hooks are aimed at **device/actuator** execution paths.

## Troubleshooting

| Symptom | Likely cause |
|---------|----------------|
| Bridge logs “dropping message: topic farm_id≠…” | Publisher path uses wrong farm segment; fix topic or `GR33N_FARM_ID` |
| HTTP **401/403** from API | Missing/wrong `X-API-Key`; API must have `PI_API_KEY` set (not `AUTH_MODE=dev` bypass on prod) |
| Readings missing | Unmapped `(device_uid, slug)` — add YAML row or publish numeric slug equal to `sensor_id` |
| Bursts overload HTTP | Raise `MQTT_BATCH_MS` so the bridge coalesces into fewer `POST /sensors/readings/batch` calls |

## Related documents

- [`openapi.yaml`](../openapi.yaml) — machine-readable contracts.
- [`docs/phase-13-operator-documentation.md`](phase-13-operator-documentation.md) — operator index.
- Phase 14 plan: [`plans/phase_14_network_and_commons.plan.md`](plans/phase_14_network_and_commons.plan.md).
