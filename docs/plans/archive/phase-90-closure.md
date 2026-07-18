# Phase 90 — closure (OC-90)

**Status:** **Shipped** on `main`.

**Canonical plan:** [`phase_90_device_taxonomy_registry.plan.md`](phase_90_device_taxonomy_registry.plan.md)

**Depends on:** Phase 38 zone cockpit plant-need tabs; optional alignment with Phase 88 platform metadata pattern.

---

## The one job (done)

> **Water / Light / Climate tabs** classify sensors and actuators from a **platform registry in Postgres**, not from `plantNeeds.js` static `Set()` lists.

---

## Workstream checklist

| WS | Deliverable | Verify |
|----|-------------|--------|
| **WS1** | `gr33ncore.device_type_registry` migration + seed | `db/migrations/20260622_phase90_device_type_registry.sql` |
| **WS2** | `GET /platform/device-taxonomy` | `internal/platform/devicetaxonomy/registry.go` |
| **WS3** | UI fetch not hardcode | `deviceTaxonomy.js`, `plantNeeds.js`, `sensorTypeLabel.js` |
| **WS4** | Pi wiring sources from registry | `hardwareWiring.js` · `wiring_source_options` in API payload |
| **WS5** | Guardian zone snapshot grouping | `internal/farmguardian/readtools.go` · `devicetaxonomy.Current()` |
| **WS6** | Contract smokes + Vitest | `smoke_phase90_test.go` · `device-taxonomy.test.js` |

---

## API shape

```
GET /platform/device-taxonomy
```

Returns `sensors`, `actuators`, `by_plant_need`, and `wiring_source_options`. Each row includes `type_key`, `plant_need`, `display_label`, `supports_pulse`, optional `gh_role`, and optional `wiring_sources`.

UI caches via `loadDeviceTaxonomy`; bundled fallback in `deviceTaxonomy.fallback.js` mirrors the migration seed.

---

## Operator impact

| Before | After |
|--------|-------|
| ~20 sensor types hardcoded in `plantNeeds.js` | Full registry from Postgres (≥20 sensors, ≥15 actuators) |
| Unknown types (e.g. `temp_f`) heuristic → Climate | `temp_f` seeded under **air** with label **Temperature (°F)** |
| Duplicate labels in `sensorTypeLabel.js` | Labels from registry lookup |
| Pi wiring hints scattered | `wiring_source_options` from registry extension |

New integrator types: add a migration row — UI and Guardian pick up on deploy without Vue edits.

---

## Guardian alignment

Zone read tools group latest sensor readings by `plant_need` from the registry and use `DisplayLabel` instead of raw `sensor_type` strings in prompts.

---

## Automated tests

| Test | Path |
|------|------|
| API contract (counts, `temp_f`, `by_plant_need`, wiring) | `cmd/api/smoke_phase90_test.go` |
| Fallback registry parity | `internal/platform/devicetaxonomy/fallback_test.go` |
| Loader + `plantNeeds` / labels | `ui/src/__tests__/device-taxonomy.test.js` |
| Climate classification incl. `temp_f` | `ui/src/__tests__/plantNeeds.test.js` |

---

## OC-90

Phase 90 is **closed** when smokes pass and zone cockpit Water/Light/Climate tabs classify devices from **`GET /platform/device-taxonomy`** (or bundled fallback).
