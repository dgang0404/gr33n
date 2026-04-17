---
name: Phase 20.5 Husbandry & Climate Bootstraps
overview: >
  Turn the primitives Phase 20 just finished (rules, schedules, sensors, actuators)
  into working operator surfaces for the farm patterns we don't have bootstraps
  for yet: animal enclosures, greenhouses / tents with dew-point control, drying
  rooms, and small aquaponics loops. No schema migrations — everything reuses
  existing tables. Also teaches the Pi client to compute derived sensor
  channels (dew point, VPD, heat index) so the rule engine can key off them
  without any backend schema change. Target: 3–4 days.
todos:
  - id: ws1-pi-derived-sensors
    content: "WS1: Pi client — new `source: derived` sensor type that computes dew_point, vpd, heat_index from configured source sensors; reports as its own sensor channel; unit tests in pi_client/test_gr33n_client.py"
    status: completed
  - id: ws2-bootstrap-templates
    content: "WS2: Four new farm bootstrap templates in `gr33ncore.apply_farm_bootstrap_template` — chicken_coop_v1, greenhouse_climate_v1, drying_room_v1, small_aquaponics_v1; sidebar picker via BOOTSTRAP_STARTER_OPTIONS"
    status: completed
  - id: ws3-playbook-docs
    content: "WS3: Playbook section per pattern in docs/ (one new file `pattern-playbooks.md`), cross-linked from workflow-guide.md §2 (Zones)"
    status: completed
  - id: ws4-smoke
    content: "WS4: Smoke tests — bootstrap each new template against a blank farm; assert the expected zones / sensors / actuators / schedules / rules land; pi_client unit tests for derived-sensor math"
    status: completed
isProject: false
---

# Phase 20.5 — Husbandry & Climate Bootstraps

## Why this phase

Phase 20 gave us a real rule engine, but day-one operators still have to hand-wire the 30+ rows (zones, sensors, actuators, schedules, rules) needed to run a chicken coop or a drying room. That work is mechanical and identical across farms — the perfect job for a bootstrap template. And the schema doesn't need a single migration: every pattern the user listed (fencing, hoppers, fans, humidifiers, dehumidifiers) is already expressible with the existing free-text `sensor_type` / `actuator_type` columns.

The one genuine gap is **derived sensor channels**. Cannabis flower-room and drying-room climate control is keyed off *dew point* and *VPD*, neither of which is directly measured — they're computed from temperature + humidity. Rather than teach the backend rule evaluator to do math (which would need a new predicate shape), we push the computation to the edge: the Pi reports `dew_point` as if it were a physical channel. Zero backend churn, and the edge can recompute locally when the network is flaky.

This phase is deliberately small. It ships operator value without touching the schema so the surface RAG has to index (Phase 21) stays stable.

## Hand-offs from earlier phases (reuse, don't re-implement)

- **Bootstrap dispatch** — `gr33ncore.apply_farm_bootstrap_template(farm_id, template_key)` (see `db/migrations/20260423_farm_bootstrap_templates.sql` and Phase 15 plan). Each new template is a new `IF template_key = '...' THEN ... END IF;` branch in that function. The existing `jadam_indoor_photoperiod_v1` is the reference implementation — idempotent inserts, `ON CONFLICT DO NOTHING` on every row.
- **Picker UI** — `ui/src/constants/bootstrapTemplates.js` exports `BOOTSTRAP_STARTER_OPTIONS` + a per-template `*_SUMMARY` object; `Farms.vue` and `OrgDefaultsForm.vue` already read from this. New templates just add entries.
- **Rule engine** — Phase 20 landed `automation_rules` + `executable_actions`. Climate-control bootstraps can pre-seed rules like "if dew_point > 15°C for 10 min → turn on dehumidifier" out of the box.
- **Pi sensor reader** — `pi_client/gr33n_client.py::SensorReader` dispatches on `cfg['source']`. WS1 adds a new `'derived'` source that reads two other sensor channels locally and emits a computed value.

## Scope

| WS | Focus | Location in repo |
|----|-------|------------------|
| **WS1** | Pi-side derived sensors (dew_point, vpd, heat_index) | `pi_client/gr33n_client.py`, `pi_client/config.yaml`, `pi_client/test_gr33n_client.py` |
| **WS2** | Four new bootstrap templates | New migration `db/migrations/2026xxxx_phase205_bootstraps.sql` (+ mirror in schema file); `ui/src/constants/bootstrapTemplates.js` |
| **WS3** | Playbook docs per pattern | `docs/pattern-playbooks.md` (new); cross-link from `docs/workflow-guide.md` |
| **WS4** | Smoke + pi_client unit tests | `cmd/api/smoke_test.go`, `pi_client/test_gr33n_client.py` |

## Work-stream detail

### WS1 — Pi-side derived sensors

- Add a new `source: derived` branch in `SensorReader._init_hardware` / `.read()`. Config shape:
  ```yaml
  - sensor_id: 42
    source: derived
    sensor_type: dew_point        # or vpd, heat_index
    inputs:
      temperature_c: 37           # sensor_id of the temp sensor on this Pi
      humidity_pct: 38            # sensor_id of the RH sensor on this Pi
  ```
- The reader caches the most recent raw reading it posted for each "input" sensor (keep a `self._latest_by_id: dict[int, tuple[float, datetime]]` on the client) and computes on demand. If either input is older than `config.derived_input_max_age_seconds` (default 120s), `read()` returns `None` and the client logs — don't emit stale derived values.
- **Math** (well-known formulas — keep them in one function each, fully unit-tested):
  - `dew_point_c = 243.04 * (ln(rh/100) + (17.625*t)/(243.04+t)) / (17.625 - ln(rh/100) - (17.625*t)/(243.04+t))` (Magnus-Tetens)
  - `vpd_kpa = 0.6108 * exp(17.27*t/(t+237.3)) * (1 - rh/100)`
  - `heat_index_c` via the Rothfusz regression; fall back to `t` when `t < 27°C`.
- **No backend change** — derived sensors are registered in `gr33ncore.sensors` exactly like physical ones (free-text `sensor_type`), they just happen to be driven by a Pi with `source=derived`. The ingest endpoint doesn't need to know the difference.

### WS2 — Four new bootstrap templates

Each new template is a branch in the PL/pgSQL bootstrap function. All follow the idempotent-insert pattern of `jadam_indoor_photoperiod_v1`.

**`chicken_coop_v1`** — one "Chicken Coop" zone, sensors `water_level` + `feed_level` + `coop_temperature` + `coop_humidity`, actuators `feeder_hopper` (relay) + `water_valve` + `coop_exhaust_fan` + `coop_heat_lamp`, schedules "06:00 feed" + "18:00 feed" + "weekly egg check" task, rules "water_level < 20 → create task 'refill waterer'" + "feed_level < 15 → create task 'refill hopper'" + "coop_temperature > 32 → exhaust_fan on" + "coop_temperature < 5 → heat_lamp on".

**`greenhouse_climate_v1`** — one "Greenhouse" zone, sensors `temperature` + `humidity` + `co2_ppm` + derived `dew_point` + derived `vpd`, actuators `exhaust_fan` + `humidifier` + `dehumidifier` + `shade_cloth_motor` + `co2_injector`, rules keyed off VPD and dew point (placeholder thresholds — operator tunes per crop). One "weekly CO₂ bottle check" task.

**`drying_room_v1`** — one "Drying Room" zone with *tight* humidity/temperature range, sensors `temperature` + `humidity` + derived `dew_point`, actuators `dehumidifier` + `circulation_fan`, rules "dew_point > 12°C → dehumidifier on" + "dew_point < 7°C → dehumidifier off" + "humidity < 55 → dehumidifier off" (cannabis dry-cure window 55–62% RH @ 15–21°C). Includes a clear HelpTip note that these are cannabis-oriented defaults and should be re-tuned for basil, orchids, etc.

**`small_aquaponics_v1`** — two zones ("Fish Tank", "Grow Bed") + one `gr33naquaponics.loops` row linking them, sensors `water_temperature` + `ph` + `ammonia_ppm` + `nitrate_ppm` on the tank, `ph` + `ec` on the bed, actuator `return_pump` + `air_pump`, rules "ammonia > 0.5ppm → create task 'ammonia spike, check fish'" + "water_temperature < 18 → create task 'tank heater check'", plus a daily "feed fish" task.

UI: extend `BOOTSTRAP_STARTER_OPTIONS` with four new entries + four `*_SUMMARY` objects in `bootstrapTemplates.js`. `Farms.vue` / `OrgDefaultsForm.vue` already render dynamically.

### WS3 — Playbook docs per pattern

New file `docs/pattern-playbooks.md` with one short section per bootstrap template: what hardware to buy, what sensors and actuators wire to what, the defaults the template ships with, and what to tune. Cross-linked from `docs/workflow-guide.md` §2 (Zones) as "Patterns: if you're running X, start here." Format matches the existing operator playbooks in `docs/`.

Also: add a short note in the workflow guide §3 (Schedules & automation runs, which Phase 20 WS5 just rewrote) pointing operators at dew_point / VPD as derived channels — "you don't need a dew point sensor, the Pi computes it for you."

### WS4 — Smoke + unit tests

- **Smoke** (`cmd/api/smoke_test.go`): for each new template key, POST `/farms` with `bootstrap_template=<key>`, then assert row counts in the expected tables (zones, sensors, actuators, schedules, automation_rules, executable_actions, tasks) match the template's summary. Mirrors the existing `TestFarmBootstrapOnCreate` pattern — one test per template is fine.
- **pi_client** (`test_gr33n_client.py`): unit tests for the derived-sensor math (known-good input/output pairs from published references) + a behaviour test that `read()` returns `None` when an input reading is stale.

## After Phase 20.5

- Four one-click patterns cover ~80% of the farm types we expect early. Operators can still build custom farms from scratch — bootstraps are additive, not mandatory.
- The Pi now reports derived channels that Phase 20.6 will use as first-class sensor types when wiring up stage-scoped setpoints.
- No schema changes. Zero RAG-indexing risk. Every row written is through the existing CRUD surface.

## Risks / things to watch

- **Bootstrap drift** — every new template is more code in a single PL/pgSQL function. Keep branches alphabetized and each branch self-contained. If the function balloons past ~800 lines, split per-template PL/pgSQL files and have the dispatcher CALL them.
- **Derived-sensor staleness** — if the Pi computes dew_point off two sensors and one is broken, the rule reading off dew_point silently stops firing. The `input_max_age_seconds` guard + a clear log line is the first line of defense. A follow-up (Phase 21 or later) could surface "this derived sensor hasn't reported in N minutes" as an alert, same way sensor_readings staleness is surfaced today.
- **Cannabis-specific defaults** — the drying-room template ships with cannabis dew-point windows. Document loudly in the HelpTip + playbook that these are defaults, not gospel, and that basil/orchid/herb operators should retune. Don't ship strain-specific rules until Phase 20.6 makes stage setpoints a first-class concept.

---

## Using this plan in a new chat (copy-paste prompt)

```text
Implement Phase 20.5 per @docs/plans/phase_20_5_husbandry_and_climate_bootstraps.plan.md.

Scope:
1) WS1 — Pi-side derived sensors: add `source: derived` to pi_client/gr33n_client.py's SensorReader, with dew_point / vpd / heat_index computed from other configured sensors on the same Pi; include input-staleness guard; unit tests in pi_client/test_gr33n_client.py.
2) WS2 — Four new farm bootstrap templates (chicken_coop_v1, greenhouse_climate_v1, drying_room_v1, small_aquaponics_v1) as branches in gr33ncore.apply_farm_bootstrap_template. Mirror each new branch in db/schema/gr33n-schema-v2-FINAL.sql. Extend BOOTSTRAP_STARTER_OPTIONS + *_SUMMARY in ui/src/constants/bootstrapTemplates.js.
3) WS3 — New docs/pattern-playbooks.md with one section per pattern; cross-link from workflow-guide.md §2.
4) WS4 — Smoke tests per template (row-count assertions) + pi_client derived-sensor math tests.

Constraints: NO schema migrations except the new bootstrap-template branch (additive PL/pgSQL only — no new tables, no new columns, no enum changes). Reuse the existing idempotent-insert pattern from jadam_indoor_photoperiod_v1. Run go test ./cmd/api/..., go test ./..., python3 -m pytest pi_client/test_gr33n_client.py -q, and npm run build in ui/ after each WS. Update this plan's YAML todo statuses when each WS lands.
```
