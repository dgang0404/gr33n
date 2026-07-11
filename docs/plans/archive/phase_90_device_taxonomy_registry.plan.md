---
name: Phase 90 — Device taxonomy registry
overview: >
  Platform DB registry maps sensor_type and actuator_type to plant need (water/light/climate),
  display labels, pulse/GH flags; UI plantNeeds.js and Guardian zone tools consume API.
todos:
  - id: ws1-schema
    content: "WS1: gr33ncore.device_type_registry migration + seed from plantNeeds.js sets"
    status: completed
  - id: ws2-api
    content: "WS2: GET /platform/device-taxonomy — sensors + actuators grouped by plant_need"
    status: completed
  - id: ws3-ui
    content: "WS3: plantNeeds.js, sensorTypeLabel.js, ZoneGreenhouseTab — fetch not hardcode"
    status: completed
  - id: ws4-pi-setup
    content: "WS4: deviceSetupWizard + hardwareWiring — wiring sources from registry extension"
    status: completed
  - id: ws5-guardian
    content: "WS5: Guardian summarize_zone_* uses registry for sensor/actuator grouping in prompt"
    status: completed
  - id: ws6-smokes
    content: "WS6: New sensor type in seed appears in correct Water/Light/Climate tab"
    status: completed
isProject: false
---

# Phase 90 — Device taxonomy registry

## Status

**Shipped.** Zone cockpit and Guardian classify devices from Postgres registry.

**Closure:** [`phase-90-closure.md`](phase-90-closure.md) · **OC-90**

---

## The one job

> **Water / Light / Climate tabs** classify sensors and actuators from a **platform registry in Postgres**, not from `plantNeeds.js` static `Set()` lists.

---

## Gap today

`ui/lib/plantNeeds.js` hardcodes:

- `WATER_SENSOR`, `LIGHT_SENSOR`, `AIR_SENSOR` (~20 types)
- `WATER_ACTUATOR`, `LIGHT_ACTUATOR`, `AIR_ACTUATOR`, `PULSE_ACTUATOR`
- Heuristic fallback: unknown → `air`

Custom types (`temp_f`, `rh_pct`, new Pi drivers) → wrong tab, wrong comfort grouping, Guardian zone snapshot groups incorrectly.

`sensorTypeLabel.js` duplicates display names.

---

## Target schema

```sql
CREATE TABLE gr33ncore.device_type_registry (
  type_key       TEXT PRIMARY KEY,  -- e.g. soil_moisture, exhaust_fan
  device_class   TEXT NOT NULL CHECK (device_class IN ('sensor','actuator')),
  plant_need     TEXT NOT NULL CHECK (plant_need IN ('water','light','air')),
  display_label  TEXT NOT NULL,
  supports_pulse BOOLEAN NOT NULL DEFAULT false,
  gh_role        TEXT,              -- shade | vent | fan | null
  wiring_sources JSONB,             -- optional Pi wiring hints
  sort_order     INT NOT NULL DEFAULT 0
);
```

**Seed migration:** export current `plantNeeds.js` + `sensorTypeLabel.js` + greenhouse GH types into INSERT rows.

**Extend:** integrator adds row via migration — UI and Guardian pick up on deploy.

---

## API

```
GET /platform/device-taxonomy
```

```json
{
  "sensors": [{ "type_key": "ec", "plant_need": "water", "display_label": "EC", … }],
  "actuators": [ … ],
  "by_plant_need": { "water": […], "light": […], "air": […] }
}
```

UI: replace `sensorPlantNeed()` / `actuatorPlantNeed()` with registry lookup (cached).

---

## Guardian (WS5)

When building zone snapshot / read-tool blocks:

- Group sensors by `plant_need` from registry
- `summarize_zone_greenhouse_climate` — shade/vent/fan from `gh_role`
- Persona: cite registry label for unknown types instead of raw `sensor_type` string

**Does not replace** live readings — only classification metadata.

---

## Acceptance

- [x] Seed ≥ current hardcoded type count
- [x] Add `temp_f` sensor in seed → appears under Climate tab
- [x] Pulse UI shows for actuators with `supports_pulse=true`
- [x] Guardian zone question routes to correct need tab in guidance text

**Prompt loop:** `phase 90 ws1` … or **`phase 90`**.
