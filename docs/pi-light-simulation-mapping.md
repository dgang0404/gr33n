# Pi light simulation rig — sensor & actuator mapping (Phase 125 WS1)

**Status:** normative spec for WS2 (`pi_client` LED driver) and WS4 (demo runbook).

Before real plants go in, a small Raspberry Pi + WS2812 strip lets an operator
*see* the automation loop: sensor reading → comfort band → rule/schedule →
actuator command → device ack. This document defines which LED shows what, using
the same band language as the UI (`SensorTile.vue` badges).

**Non-goals:** no new API routes, no schema changes. The rig consumes existing
sensor readings and actuator state the way `pi_client` already does for relays.

---

## 1. UI parity — sensor comfort bands

Match `ui/src/components/SensorTile.vue`:

| UI badge | Condition | Rig LED behavior |
|----------|-----------|------------------|
| **OK** | `alert_threshold_low ≤ value ≤ alert_threshold_high`, and not within 15% of either edge | Solid **green** `(0, 180, 0)` at 60% brightness |
| **WARN** | In band but within **15%** of low or high threshold (same `range * 0.15` math) | Solid **amber** `(255, 160, 0)` at 70% brightness |
| **ALERT — low** | `value < alert_threshold_low` | **Cyan blink** `(0, 120, 255)` — 1 Hz (500 ms on / 500 ms off) |
| **ALERT — high** | `value > alert_threshold_high` | **Red blink** `(255, 0, 0)` — 1 Hz |
| **NO DATA** | No reading, or reading older than **3×** `reading_interval_seconds` | **Dim gray pulse** `(40, 40, 40)` — 0.5 Hz (or off if strip is full) |

Thresholds come from each sensor row (`alert_threshold_low` / `alert_threshold_high`
in the DB). Do not hardcode crop-specific bands in the driver.

---

## 2. Actuator & device states

Actuator `current_state_text` from the API is typically `on`, `off`, or `offline`.
Command queue states come from `GET /devices/{id}/commands` (pending / completed / failed).

| State | Rig LED behavior |
|-------|------------------|
| **Idle / off** | Solid dim **white** `(80, 80, 80)` at 15% brightness |
| **On / running** | **Type-colored blink** 2 Hz (250 ms on / 250 ms off) — see §4 |
| **Command queued** (pending, not yet acked) | Slow **amber pulse** 1 Hz |
| **Command failed** (`execution_status = failed`) | **Magenta fast blink** 4 Hz |
| **Device offline** (`devices.status != online`) | **Magenta fast blink** 4 Hz on that device's actuators |
| **Stale wiring** (drift badge from Phase 121) | GPIO fault LED (§3) solid amber |

### Actuator type colors (when on / blinking)

| `actuator_type` | Color RGB | Notes |
|-----------------|-----------|-------|
| `light` | `(255, 200, 80)` warm yellow | Grow light |
| `pump` | `(60, 120, 255)` blue | Irrigation / fertigation pump |
| `fan` | `(80, 220, 255)` cyan | Exhaust / circulation |
| `valve` | `(180, 80, 255)` violet | Solenoid / ball valve |
| `heater` | `(255, 80, 40)` orange | Heat mat / room heat |
| *(other)* | `(200, 200, 200)` white | Fallback |

---

## 3. Physical rig v1 (bench / demo)

Minimal parts for one complete **Veg Room** loop (moisture → alert → pump/light).

| Part | Qty | Role |
|------|-----|------|
| Raspberry Pi 4/5 (or Pi Zero 2 W) | 1 | Runs `pi_client` in `simulation` driver mode |
| WS2812B strip or ring | 8 pixels | Sensor + actuator indicators |
| 330 Ω resistor | 1 | NeoPixel data line (if not on HAT) |
| 5 V level shifter (optional) | 1 | 3.3 V → 5 V data for long strips |
| GPIO LED + 220 Ω | 2 | Heartbeat + fault (plain LEDs, not NeoPixel) |

### Pin assignment (defaults for WS2 config)

| Signal | BCM GPIO | Notes |
|--------|----------|-------|
| NeoPixel data | **18** | `board.D18` — standard for rpi_ws281x / adafruit_neopixel |
| Heartbeat LED | **17** | Slow green pulse when Pi client loop is healthy |
| Fault / drift LED | **27** | Solid amber when config drift or device offline |

### NeoPixel index map — rig v1 (8 pixels)

Left-to-right on the strip as viewed from the operator:

| Index | Entity (demo farm 1) | `sensor_type` / `actuator_type` | Zone |
|-------|----------------------|-----------------------------------|------|
| **0** | Media Moisture Indoor | `soil_moisture` | *(farm-level)* |
| **1** | EC Sensor | `conductivity` | *(farm-level)* |
| **2** | Air Temp Indoor | `temperature` | *(farm-level)* |
| **3** | Air Humidity Indoor | `humidity` | *(farm-level)* |
| **4** | pH Sensor | `ph` | *(farm-level)* |
| **5** | Veg Room Grow Light | `light` | Veg Room |
| **6** | Veg Room Irrigation Pump | `pump` | Veg Room |
| **7** | *(spare / demo)* | — | Shows aggregate “automation fired” white flash when any rule triggers |

**Pi device:** `demo-veg-relay-01` (platform sync). Actuators on indices 5–6 bind to
this device. Sensor LEDs poll by **sensor name** (stable across re-seeds); WS2 resolves
`sensor_id` at startup via `GET /farms/{farm_id}/sensors` or platform wiring export.

### Rig v2 expansion (optional — not required for WS2 MVP)

Add pixels 8–11 or a second strip for Flower Room path:

| Index | Entity | Type | Zone |
|-------|--------|------|------|
| 8 | Flower Room Irrigation Pump | `pump` | Flower Room |
| 9 | PAR Sensor Indoor | `par` | *(farm-level)* |
| 10 | CO2 Sensor Indoor | `co2` | *(farm-level)* |
| 11 | Berry Patch Soil Moisture | `soil_moisture` | Outdoor Berry Patch |

Device for index 8: `demo-flower-relay-01`.

---

## 4. Full demo farm catalog (farm 1 seed)

Every sensor and actuator shipped in `db/seeds/master_seed.sql`. **Rig v1** uses
the bold rows; others are documented for rig v2, multi-strip setups, or operator
custom maps.

### Sensors

| Name | Type | Alert low | Alert high | Interval (s) | Rig v1 LED |
|------|------|-----------|------------|--------------|------------|
| PAR Sensor Indoor | `par` | 100 | 1800 | 60 | v2 → 9 |
| Lux Sensor Indoor | `light_lux` | 1000 | 80000 | 60 | — |
| **Media Moisture Indoor** | `soil_moisture` | 25 | 80 | 120 | **0** |
| **Air Temp Indoor** | `temperature` | 16 | 32 | 60 | **2** |
| Root Zone Temp | `temperature` | 18 | 26 | 120 | — |
| **Air Humidity Indoor** | `humidity` | 35 | 75 | 60 | **3** |
| Soil Moisture Outdoor | `soil_moisture` | 20 | 85 | 300 | — |
| **EC Sensor** | `conductivity` | 0.5 | 3.5 | 60 | **1** |
| **pH Sensor** | `ph` | 5.5 | 7.0 | 60 | **4** |
| **CO2 Sensor Indoor** | `co2` | 400 | 1500 | 60 | v2 → 10 |
| Propagation Dome Temp | `temperature` | 20 | 29 | 60 | — |
| Herb Room Air Temp | `temperature` | 16 | 28 | 60 | — |
| Pepper Bed Soil Moisture | `soil_moisture` | 20 | 85 | 300 | — |
| Berry Patch Soil Moisture | `soil_moisture` | 25 | 80 | 300 | v2 → 11 |

### Actuators

| Name | Type | Device UID | Channel | Rig v1 LED |
|------|------|------------|---------|------------|
| **Veg Room Grow Light** | `light` | `demo-veg-relay-01` | 1 | **5** |
| **Veg Room Irrigation Pump** | `pump` | `demo-veg-relay-01` | 2 | **6** |
| Flower Room Irrigation Pump | `pump` | `demo-flower-relay-01` | 2 | v2 → 8 |

### Devices

| Name | `device_uid` | Zone | Rig role |
|------|--------------|------|----------|
| Veg Relay Controller | `demo-veg-relay-01` | Veg Room | **Primary sim Pi** |
| Flower Relay Controller | `demo-flower-relay-01` | Flower Room | v2 second Pi or shared strip |

---

## 5. `pi_client` simulation config (WS2 target schema)

Add to `config.yaml` (alongside existing `api` / `device` / `farm` bootstrap).
Platform sync still supplies sensor/actuator wiring; simulation block only maps
entities to LEDs.

```yaml
simulation:
  enabled: true
  driver: neopixel          # neopixel | gpio_only
  poll_interval_seconds: 2  # how often to refresh LED state from API

  neopixel:
    pin: 18
    count: 8
    brightness: 0.4         # 0.0–1.0 strip-wide max
    pixel_order: GRB        # WS2812B default

  gpio_leds:
    heartbeat_pin: 17
    fault_pin: 27

  sensors:
    - match_name: "Media Moisture Indoor"
      pixel: 0
    - match_name: "EC Sensor"
      pixel: 1
    - match_name: "Air Temp Indoor"
      pixel: 2
    - match_name: "Air Humidity Indoor"
      pixel: 3
    - match_name: "pH Sensor"
      pixel: 4

  actuators:
    - match_name: "Veg Room Grow Light"
      pixel: 5
    - match_name: "Veg Room Irrigation Pump"
      pixel: 6

  # Optional: flash pixel 7 white 200ms when automation enqueues any command
  activity_pixel: 7
```

**Entity resolution:** prefer `match_name` (stable for demo farm). Optional
`match_sensor_id` / `match_actuator_id` overrides for farms where names differ.

**Driver swap later:** set `simulation.enabled: false` (or remove block) and use
normal relay actuators from platform wiring — no platform-side changes.

---

## 6. Data sources (existing API — no new endpoints)

| LED input | Source | Refresh |
|-----------|--------|---------|
| Sensor band color | **Local `ReadingCache`** — values this `pi_client` posts (or WS3 synthetic loopback) + thresholds from `simulation.sensors[]` in `config.yaml` | Every `poll_interval_seconds` |
| Actuator on/off | Local `SimulationActuatorController.state` (commands via normal queue) | Same poll |
| Command queued/failed | Local flags set when `get_next_command` / ack runs | Same poll |
| Device offline | `api.is_reachable()` — fault GPIO when API down | Same poll |
| Heartbeat GPIO | Local — `pi_client` main loop alive | 1 Hz toggle |

> **WS2 note:** Pi credentials cannot call JWT-only `GET /farms/{id}/sensors` routes,
> so sensor LEDs use the local reading cache plus thresholds copied into
> `simulation.sensors[]` (defaults in [`config.simulation.example.yaml`](../pi_client/config.simulation.example.yaml)).
> Thresholds should match the DB row; re-copy after changing alert bands in the UI.

Stale reading rule: if `now - reading_time > 3 * reading_interval_seconds`, treat
as **NO DATA** (gray pulse).

**Implementation:** [`pi_client/light_simulation.py`](../pi_client/light_simulation.py) —
enable with `simulation.enabled: true` in `config.yaml`.

### WS3 — synthetic sensor loopback

Add `simulation.synthetic_sensors[]` to post generated values through the same
`POST /sensors/{id}/readings` path as real hardware. Modes: `sine`, `hold`,
`step`, `demo_moisture` (3-minute moisture drop/recover cycle).

| Tool | Role |
|------|------|
| `synthetic_sensors.py` | In-process loop inside `gr33n_client` when configured |
| `nudge_sensor.py` | One-shot manual POST (`--sensor-id N --value V`) |
| `run_demo_moisture_loop.py` | Scripted WS4 walkthrough (external to daemon) |

Example — see [`config.simulation.example.yaml`](../pi_client/config.simulation.example.yaml).

**WS4 runbook:** [`pi-light-simulation-runbook.md`](pi-light-simulation-runbook.md)  
**WS5 hardware + relay swap:** [`pi_client/README-simulation-rig.md`](../pi_client/README-simulation-rig.md)

---

## 7. Primary demo loop (pointer to WS4)

Rig v1 is sized for this walkthrough:

1. **Warm start** — all sensor LEDs green (synthetic or real readings in band).
2. **Drop moisture** — WS3 script or manual POST lowers Media Moisture below 25%.
   - Pixel **0** → cyan blink (ALERT low).
3. **Automation fires** — irrigation rule/schedule enqueues pump command.
   - Pixel **6** → amber pulse (queued), then blue blink (on).
   - Pixel **7** — brief white flash (activity).
4. **UI check** — alert on Alerts page; optional task on Tasks.
5. **Recover** — raise moisture back in band; pixel **0** returns solid green;
   pixel **6** returns dim white (off).

Secondary paths for WS4: EC/pH out of band (pixels 1, 4), Veg light schedule
(pixel 5 yellow blink at photoperiod ON).

---

## 8. WS2 implementation notes

- Run LED refresh on a **daemon thread**; do not block sensor posting or command ack.
- Off-Pi dev: stub NeoPixel as log lines (`LED[3] = AMBER solid`) — same pattern
  as GPIO stubs in `gr33n_client.py`.
- Blink timing: use monotonic clock; phase-align all pixels of the same pattern.
- When `simulation.enabled` and platform sync both present, **actuator commands still
  flow through the normal queue** — simulation driver reflects state on LEDs instead
  of toggling relay GPIO (or toggles both if `simulation.mirror_relay_gpio: true`
  for bench debugging).

---

## Related docs

- [Phase 125 plan](plans/phase_125_pi_sensor_actuator_light_simulation.plan.md)
- [Demo runbook (WS4)](pi-light-simulation-runbook.md)
- [Hardware & relay swap (WS5)](../pi_client/README-simulation-rig.md)
- [Pi integration guide](pi-integration-guide.md)
- [Virtual Pi config export](plans/phase_121_virtual_pi_hookup_export.plan.md) — `GET /devices/{id}/pi-config`
- [Connectivity requirements](connectivity-requirements.md) — LAN-only; no WAN for LED driver
