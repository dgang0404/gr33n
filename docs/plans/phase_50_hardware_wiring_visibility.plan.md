---
name: Phase 50 — Hardware wiring visibility (read-only)
overview: >
  Close the gap where Pi GPIO wiring (pin / source / I2C channel / serial port) lives only
  in pi_client/config.yaml and is invisible in the app. Add structured, validated wiring
  metadata to sensors/actuators, expose it read + edit via API/UI, surface "where this is
  wired" on Sensors, Controls, and the device wizard, and generate a config.yaml from the
  platform so operators stop hand-editing YAML or writing SQL. Pi runtime still reads local
  YAML this phase; live pull-from-API is deferred to Phase 51.
todos:
  - id: ws1-wiring-model
    content: "WS1: Define wiring schema in sensors.config / actuators.config JSONB (source, gpio_pin, i2c_channel, i2c_address, serial_port, notes); JSON-schema + migration backfill from config.yaml"
    status: completed
  - id: ws2-api-read-write
    content: "WS2: API — expose wiring on sensor/actuator GET; PATCH wiring (validated, conflict-checked: no two devices on same pin per Pi)"
    status: completed
  - id: ws3-ui-surface
    content: "WS3: UI — wiring panel on Sensors detail + Controls cards + device wizard step; 'Wiring: BCM GPIO 17 (relay)' badges; empty state 'Not wired yet'"
    status: completed
  - id: ws4-config-generator
    content: "WS4: Generate pi_client/config.yaml from DB per device — download/copy in device wizard; round-trips with WS1 schema"
    status: completed
  - id: ws5-validation-conflicts
    content: "WS5: Conflict + sanity checks — duplicate pin/channel per device, unknown source driver, derived-sensor inputs exist; surfaced in UI + db-sanity-report"
    status: completed
  - id: ws6-docs-tests
    content: "WS6: pi-integration-guide rewrite (DB-first), architecture §, Go handler tests, Vitest wiring panel, phase-50-closure.test.js, OC-50"
    status: completed
isProject: false
---

# Phase 50 — Hardware wiring visibility (read-only)

## Status

**Shipped.** WS1–WS6 complete on `main`. Wiring metadata + read/edit UI + config generation. **Pi still reads local YAML** — live pull-from-API is **[Phase 51](#relationship-to-phase-51)**.

**Roadmap:** [farmer_ux_roadmap_40_plus.plan.md](farmer_ux_roadmap_40_plus.plan.md) (edge/Pi track).

**Closure:** **OC-50** in [phase_35_37_operational_closure.plan.md](phase_35_37_operational_closure.plan.md).

---

## Problem

**Where a sensor or actuator is physically wired exists in exactly one place: the Pi's local YAML.**

```17:21:pi_client/config.yaml
sensors:
  - sensor_id: 3
    sensor_type: temperature
    source: dht22
    pin: 4                    # BCM GPIO pin for DHT22 data line — Air Temp Indoor
```

```81:90:pi_client/config.yaml
actuators:
  - actuator_id: 1
    device_id: 1
    device_type: light
    gpio_pin: 17

  - actuator_id: 2
    device_id: 2
    device_type: irrigation
    gpio_pin: 27
```

Meanwhile the database only has freeform text — [`gr33ncore.sensors.hardware_identifier TEXT`](../db/schema/gr33n-schema-v2-FINAL.sql) and a generic `config JSONB`; same on `actuators`. **The UI shows nothing about wiring** (Sensors list, Controls cards, device wizard all omit pins/channels).

Consequences operators hit today:
- To set up or re-wire a Pi you **hand-edit YAML on the device** or write **SQL** — there is no in-app path.
- No one can answer "which GPIO is the flower-room pump on?" from the UI.
- Pin conflicts (two devices on the same GPIO) are only discovered when hardware misbehaves.
- The device wizard registers a device and prints generic hints, but never captures the actual pin map.

This phase makes wiring a **first-class, validated, visible** property of each sensor/actuator — without yet changing how the Pi fetches its runtime config.

---

## Design principles

1. **Structured, not freeform.** Wiring is typed JSON with a published schema — not a `TEXT` blob — so it validates, renders, and round-trips to YAML.
2. **DB is the source of truth for *intent*.** The platform records what *should* be wired where; the Pi YAML is generated from it. (Pi still reads local YAML until Phase 51.)
3. **Conflict-aware.** No two devices on the same pin/channel of the same Pi. Surface conflicts in UI and sanity report.
4. **Non-destructive rollout.** Existing rows keep working; wiring is optional and backfilled from the current `config.yaml` where ids match.
5. **No contract break for the Pi.** Reading/edit/generation only — the running edge client is untouched this phase.

---

## WS1 — Wiring data model

Store wiring under a reserved key in the existing `config JSONB` on `sensors` and `actuators` (no new columns; keeps migrations light and is forward-compatible):

```jsonc
// sensors.config.wiring
{
  "wiring": {
    "source": "dht22",          // driver: dht22 | ads1115 | mhz19 | bh1750 | derived | gpio_relay ...
    "gpio_pin": 4,              // BCM pin (digital sensors / relays)
    "i2c_channel": 0,           // ADS1115 channel where applicable
    "i2c_address": "0x48",      // optional
    "serial_port": "/dev/ttyS0",// optional (mh-z19 etc.)
    "device_id": 1,             // which Pi/edge device this wiring belongs to
    "notes": "Air Temp Indoor data line"
  }
}
```

- Publish a JSON schema (`docs/schemas/wiring.schema.json` or a Go struct + validator) shared by API and config generator.
- **Migration**: additive only — no column changes. A one-time backfill script maps the seeded `config.yaml` entries onto matching `sensor_id` / `actuator_id` rows for farm 1 so the demo shows real wiring.
- `actuators.config.wiring` uses `gpio_pin` + `source: gpio_relay` by default.

---

## WS2 — API read + edit

- **GET** sensor/actuator responses include a typed `wiring` object (null when unset).
- **PATCH** `wiring` on sensor/actuator (auth + farm scope as existing handlers): validates against the schema, normalizes pin/channel types, and **rejects conflicts** (same `device_id` + same `gpio_pin`/`i2c_channel`).
- Reuse existing handlers ([`internal/handler/sensor/handler.go`](../internal/handler/sensor/handler.go), [`internal/handler/actuator/handler.go`](../internal/handler/actuator/handler.go)); extend OpenAPI.
- No new endpoints if PATCH on the existing resource suffices.

---

## WS3 — UI surfacing

| Surface | What it shows |
|---------|---------------|
| Sensors detail ([`Actuators.vue`](../ui/src/views/Actuators.vue) / sensor detail) | "Wiring: BCM GPIO 4 · DHT22" badge; edit affordance |
| Controls cards ([`Actuators.vue`](../ui/src/views/Actuators.vue), `ActuatorCard.vue`) | Pin badge per actuator; "Not wired yet" empty state |
| Device wizard ([`DeviceSetupWizard.vue`](../ui/src/views/DeviceSetupWizard.vue)) | New step: assign pins/channels to the device's sensors/actuators |

- Inline editor PATCHes wiring; shows conflict errors from WS2.
- Empty state copy: "Not wired yet — set the GPIO pin so this {sensor|pump} knows its hardware."
- Plain-language: "BCM GPIO" with a help tip, not raw schema words.

---

## WS4 — config.yaml generator

- "Download Pi config" / "Copy config" in the device wizard (and optionally a per-device API endpoint) that renders a valid [`pi_client/config.yaml`](../pi_client/config.yaml) from the DB wiring for that `device_id`.
- Output must round-trip: generated YAML parses back to the same WS1 schema.
- Includes sensors, actuators, derived-sensor stubs, and the standard poll/queue keys.
- This is the operator's new happy path: register device → set wiring in UI → download config → drop on Pi. **No hand-editing, no SQL.**

---

## WS5 — Validation & conflicts

- Reject/flag: duplicate `gpio_pin` or `i2c_channel` per `device_id`; unknown `source` driver; derived-sensor `inputs` referencing missing sensors.
- Extend [`scripts/sql/db_sanity_report.sql`](../scripts/sql/db_sanity_report.sql) + [`db-sanity-report.sh`](../scripts/db-sanity-report.sh): "sensors/actuators with wiring", "pin conflicts per device" (exit non-zero on conflict).
- UI shows conflicts inline before save.

---

## WS6 — Docs, tests, closure (OC-50)

| Artifact | Content |
|----------|---------|
| [pi-integration-guide.md](../pi-integration-guide.md) | Rewrite to **DB-first**: set wiring in UI → generate config; keep manual YAML as fallback |
| [farm-guardian-architecture.md](../farm-guardian-architecture.md) | New § — wiring metadata model + generator |
| Go tests | Handler validation, conflict rejection, generator round-trip |
| Vitest | Wiring panel render + edit; empty state; conflict error |
| `ui/src/__tests__/phase-50-closure.test.js` | Closure bundle |

**OC-50** row added and closed when WS1–WS6 ship.

---

## Relationship to Phase 51

Phase 50 is **read + edit + generate** with the Pi still reading **local** YAML. **Phase 51** makes the Pi client **pull its config from the API** (and lets the UI push changes live), fully closing "you need SQL/YAML to set up a Pi." Phase 50's WS1 schema and WS4 generator are the contract Phase 51 builds on, so 50 must land first.

---

## Out of scope

- Pi client fetching config from the API at runtime (Phase 51).
- Live actuator command routing / OTA firmware (existing edge execution paths unchanged).
- Auto-discovery of connected hardware (I2C scan, etc.).
- Multi-Pi orchestration UI beyond per-device config generation.
- New columns on sensors/actuators (we reuse `config JSONB`).

---

## Definition of done

- [x] Wiring is a typed, validated object on sensors/actuators (in `config JSONB`)
- [x] API returns wiring on GET and accepts validated, conflict-checked PATCH
- [x] Sensors, Controls, and device wizard show "where this is wired" + edit
- [x] Operator can generate a working `config.yaml` from the UI (round-trips)
- [x] Sanity report flags pin/channel conflicts; demo farm backfilled from current YAML
- [x] pi-integration-guide is DB-first; OC-50 closed

---

## Suggested implementation order

1. WS1 schema + backfill (data foundation)
2. WS2 API read, then PATCH + conflict validation
3. WS5 validation/sanity (lets you trust the data early)
4. WS3 UI surfacing + edit
5. WS4 config generator (the payoff)
6. WS6 docs + closure

---

## Related

| Doc | Use |
|-----|-----|
| [pi-integration-guide.md](../pi-integration-guide.md) | Current manual wiring guide to replace |
| [pi_client/config.yaml](../pi_client/config.yaml) | Source format the generator must produce |
| [phase_44_getting_started_edge_wizard.plan.md](phase_44_getting_started_edge_wizard.plan.md) | Device wizard this phase extends |
| [phase_49_sidebar_nav_polish.plan.md](phase_49_sidebar_nav_polish.plan.md) | Companion UI-polish phase |
| [db/schema/gr33n-schema-v2-FINAL.sql](../db/schema/gr33n-schema-v2-FINAL.sql) | sensors/actuators `config JSONB` + `hardware_identifier` |

---

## Using this in a new chat

> Read `docs/plans/phase_50_hardware_wiring_visibility.plan.md`. Phase 50 is **shipped** (OC-50 closed). For Pi live config pull, start **Phase 51** (`phase_51_pi_config_sync.plan.md`).
