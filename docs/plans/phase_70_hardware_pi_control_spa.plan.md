---
name: Phase 70 — Hardware / Pi GPIO live control SPA
overview: >
  Replace the constants-only PiSetupGuide with a live, editable Hardware workspace
  where an operator sees the whole Pi at a glance — every GPIO/relay channel mapped
  to its device, zone, role (sensor vs pump/light), live state, and next scheduled
  run — and can re-assign wiring, fire a manual pulse, and download the Pi config
  from one screen. The static wiring guide survives as a Reference tab. Backend
  work closes the real gaps found in audit: relay-HAT channels not exported to the
  Pi runtime, one-actuator-per-device limit in the Pi client, and config_version
  not bumping on /assign.
todos:
  - id: ws1-board-view
    content: "WS1: GPIO board UI — one card per pin/channel showing device, zone, role, live state, next run; reads store.devices/actuators/sensors + hardwareWiring.js; loads via store.loadAll (fix onMounted bug)"
    status: pending
  - id: ws2-inline-assign-control
    content: "WS2: Inline edit per channel — re-assign zone/role/wiring (PATCH assign|wiring), rename, manual on/off/pulse (POST command); conflict + unassigned-pin warnings"
    status: pending
  - id: ws3-export-gap
    content: "WS3 (backend): BuildPiRuntimeConfig + piconfig export include relay-HAT (hardware_identifier) actuators, not only config.wiring.gpio_pin; unified channel|pin in runtime JSON/YAML"
    status: pending
  - id: ws4-multi-actuator-pi
    content: "WS4 (Pi client): gr33n_client.py drives multiple actuators per device by payload.actuator_id; add Sequent/8relind relay-HAT driver path alongside gpiozero direct GPIO"
    status: pending
  - id: ws5-config-version-assign
    content: "WS5 (backend): bump devices.config_version on /assign + hardware_identifier change (not only config.wiring JSON) so Pi hot-reloads HAT channel changes"
    status: pending
  - id: ws6-reference-tab
    content: "WS6: Fold PiSetupGuide constants into Hardware → Reference tab; clearly label as guidance vs the live 'your farm' board; device wizard entry"
    status: pending
  - id: ws7-docs-tests
    content: "WS7: pi-integration-guide update, Go export/version tests, Pi client multi-actuator test, board Vitest, phase-70-closure.test.js, smoke_phase70_test.go, OC-70"
    status: pending
isProject: false
---

# Phase 70 — Hardware / Pi GPIO live control SPA

## Status

**Planned.** The largest phase in the [SPA workspace arc](phase_68_73_spa_workspace_roadmap.plan.md) and the only one with substantial **Go + Pi-client** work. Builds on the wiring model from [Phase 50](phase_50_hardware_wiring_visibility.plan.md), config sync from [Phase 51](phase_51_pi_config_sync.plan.md), device API keys from [Phase 57](phase_57_pi_device_api_keys.plan.md), and diagnostics read-tool from [Phase 65](phase_65_guardian_pi_diagnostics.plan.md).

**Closure:** **OC-70** — tracked in this plan's DoD + [arc hub OC table](phase_68_73_spa_workspace_roadmap.plan.md#operational-closure-oc-rows). Do not add to the archived Phase 35 closure doc.

---

## The one job

> **When setting up a Pi, see — on one screen — what every GPIO/relay channel is wired to, which zone it serves, whether it's a sensor or a pump, and when it runs. Change it there. Don't click around, don't hand-edit YAML.**

This is the operator's `GPIO_RELAY · BCM GPIO 27 → which zone → sensor or pump → schedule` request, end to end.

---

## Problem

[`ui/src/views/PiSetupGuide.vue`](../ui/src/views/PiSetupGuide.vue) (659 lines) is **~90% hardcoded display constants** — parts list, DIP table (`dipTable`), channel-map cards (`channelMapCards`), a "typical 8-channel wiring plan" (`wiringPlan`), example `config.yaml` (`yamlExample`). It has **one** live section ("Your farm channels", lines ~141–227) that reads `store.devices/actuators/sensors`, but:

- It calls `store.loadFarm?.()` in `onMounted` — **the store action is `loadAll`** — so the live section often shows nothing unless another view already populated the store.
- The prominent content is the *typical* plan, **not** the operator's actual deployment.

Meanwhile the **real source of truth is the database**, edited through scattered panels:

| Truth | Where | Endpoint |
|-------|-------|----------|
| Direct GPIO relay | `actuators.config.wiring` (`gpio_pin`, `source: gpio_relay`) | `PATCH /actuators/{id}/wiring` |
| Relay-HAT channel | `actuators.hardware_identifier` (0–63) | `PATCH /actuators/{id}/assign` |
| Sensor wiring | `sensors.config.wiring` | `PATCH /sensors/{id}/wiring` |
| Actuator → zone/device | `actuators.zone_id`, `device_id` | actuator CRUD |
| Schedule → actuator | `executable_actions.target_actuator_id` | schedules API |
| Generated Pi config | DB → YAML/JSON | `GET /devices/{id}/pi-config`, `GET /devices/by-uid/{uid}/config` |

So there is **no single screen** that shows pin → device → zone → role → schedule, and three real backend gaps make the picture inconsistent (below). This phase builds that screen and closes the gaps.

---

## Audit: the three real backend gaps

Found while mapping the Pi config path. These must be fixed or the board will show data the Pi doesn't actually honor:

1. **Relay-HAT channels aren't exported to the Pi.** `BuildPiRuntimeConfig` ([`internal/hardware/piconfig_json.go`](../internal/hardware/piconfig_json.go)) and the YAML generator ([`internal/hardware/piconfig.go`](../internal/hardware/piconfig.go)) only include actuators with `config.wiring.gpio_pin`. Actuators assigned a **HAT channel via `hardware_identifier`** are dropped from the runtime config.
2. **Pi client is single-actuator + GPIO-direct only.** [`pi_client/gr33n_client.py`](../pi_client/gr33n_client.py) maps one actuator per `device_id` and ignores `payload.actuator_id`; `ActuatorController` always uses `gpiozero.OutputDevice(gpio_pin)` — no Sequent/`8relind` relay-HAT driver.
3. **`config_version` doesn't bump on `/assign`.** The DB trigger ([`db/migrations/20260608_phase51_device_config_version.sql`](../db/migrations/20260608_phase51_device_config_version.sql)) only fires on `config.wiring` JSON changes, so changing a HAT channel via `hardware_identifier` does **not** trigger a Pi hot-reload.

---

## Design principles

1. **DB is truth; the board reflects it.** The board renders live `devices/actuators/sensors` + schedules. The static guide is clearly separated as "Reference," not "your farm."
2. **Pin-first.** The Hardware workspace is organized by the physical board (stacks → channels → pins). Phase 69 stays zone-first; this is the complementary hardware-first view.
3. **What you see is what runs.** No showing a HAT-channel mapping the Pi will silently ignore — WS3/WS4/WS5 make the runtime honor everything the board shows.
4. **One screen to set up a Pi.** Assign wiring, name it, test it (manual pulse), download config — all here.
5. **Safe control.** Manual on/off/pulse routes through the existing command queue ([`POST /actuators/{id}/command`](../internal/handler/actuator/)); destructive re-assigns confirm and conflict-check.

---

## WS1 — GPIO board view

New `ui/src/views/hardware/GpioBoard.vue` (the Hardware workspace "GPIO board" tab from Phase 68):

- One card per **pin / relay channel**, grouped by device (Pi) then stack/bank. Each card shows:
  - **Identity:** `GPIO_RELAY · BCM GPIO 27` or `HAT stack 0 · ch 3` (from [`ui/src/lib/hardwareWiring.js`](../ui/src/lib/hardwareWiring.js)).
  - **Mapped to:** device → zone → actuator/sensor **name** + **role** (pump / light / fan / sensor type).
  - **Live state:** on/off (actuator) or last reading + staleness (sensor).
  - **Next run:** the next scheduled action targeting this actuator (from schedules/lighting programs).
  - **Empty channels:** "Unassigned — wire something here."
- **Fix the load bug:** call `store.loadAll()` (not the nonexistent `loadFarm`) in `onMounted`; show a skeleton while loading.
- Reuse the channel math already in `PiSetupGuide.vue` (`channelMapCards`, stack/I²C addressing) but overlay **live** assignments as the primary content.

---

## WS2 — Inline assign + control per channel

Each card is editable, reusing shipped panels/endpoints:

- **Re-assign:** open [`ActuatorWiringPanel.vue`](../ui/src/components/ActuatorWiringPanel.vue) (HAT channel `PATCH /assign` ↔ direct GPIO `PATCH /wiring`) and [`HardwareWiringPanel.vue`](../ui/src/components/HardwareWiringPanel.vue) for sensors — inline.
- **Set zone/role/name:** assign which zone the channel serves and what it is.
- **Test it:** manual on / off / pulse via `POST /actuators/{id}/command` so the operator confirms the wiring physically works during setup.
- **Warnings:** pin/channel conflicts per device (existing conflict checks), unassigned-but-scheduled actuators, sensors with stale readings.
- **Download config:** per-device "Download Pi config" (`GET /devices/{id}/pi-config`) right on the board.

---

## WS3 — Export HAT-channel actuators (backend)

Close gap #1 in [`internal/hardware/piconfig_json.go`](../internal/hardware/piconfig_json.go) + [`internal/hardware/piconfig.go`](../internal/hardware/piconfig.go):

- Include actuators that have a `hardware_identifier` (HAT channel) **as well as** those with `config.wiring.gpio_pin`.
- Emit a unified runtime descriptor per actuator: `{ actuator_id, device_id, driver: "gpio"|"relay_hat", gpio_pin?|hat_stack+channel?, ... }`.
- Round-trip with the WS1 schema from Phase 50; extend `pi-config` response + OpenAPI.

---

## WS4 — Multi-actuator + relay-HAT driver (Pi client)

Close gap #2 in [`pi_client/gr33n_client.py`](../pi_client/gr33n_client.py):

- Build an **actuator registry keyed by `actuator_id`** (not one per `device_id`); when a command arrives, dispatch on `payload.actuator_id`.
- Add a **relay-HAT driver** path (Sequent `8relind` / I²C) selected by the runtime descriptor's `driver: relay_hat`, alongside the existing `gpiozero` GPIO-direct path.
- Keep the local-YAML fallback working; honor `config_version` hot-reload (Phase 51).
- Report events back with `actuator_id` provenance (already supported by `POST /actuators/{id}/events`).

---

## WS5 — `config_version` on assign (backend)

Close gap #3:

- Extend the bump so changing `actuators.hardware_identifier` (via `/assign`) or `zone_id`/`device_id` mappings bumps the owning device's `config_version` — via DB trigger update (migration) or in the `/assign` handler ([`internal/handler/actuator/handler.go`](../internal/handler/actuator/handler.go)).
- Result: a HAT-channel re-assign on the board triggers a Pi hot-reload, same as a GPIO wiring edit does today.

---

## WS6 — Reference tab (the old constants, in their place)

- Move `PiSetupGuide.vue`'s static content (parts, DIP table, channel map, typical plan, scaling tiers, `config.yaml` example) into the Hardware workspace **Reference** tab (Phase 68 declared `/pi-setup → /hardware?tab=reference`).
- Header copy makes the distinction explicit: **"Reference — typical wiring. Your actual farm is on the GPIO board tab."**
- Device wizard ([`DeviceSetupWizard.vue`](../ui/src/views/DeviceSetupWizard.vue)) links into the board after registering a Pi.

---

## WS7 — Docs, tests, closure (OC-70)

| Artifact | Content |
|----------|---------|
| [pi-integration-guide.md](../pi-integration-guide.md) | Board-first setup: register → assign on board → test pulse → download config |
| Go test | `BuildPiRuntimeConfig` includes HAT actuators; runtime descriptor shape; `config_version` bumps on `/assign` |
| `cmd/api/smoke_phase70_test.go` (new) | End-to-end: assign HAT channel → version bumps → pi-config includes it |
| Pi client test | Multi-actuator dispatch by `actuator_id`; relay-HAT driver selection (mockable) |
| `ui/src/__tests__/phase-70-closure.test.js` (new) | Board renders live mapping; inline assign; reference tab present; loadAll called |
| `docs/pi-sequent-hat-setup.md` | Cross-reference the live board |

**OC-70** added and closed when WS1–WS7 ship.

---

## Out of scope

- Auto-discovery of connected hardware (I²C scan).
- OTA firmware / Pi client self-update.
- Multi-Pi orchestration beyond per-device config + the board.
- New "Pi hardware profile" entity (e.g. "this Pi has 2 HATs at 0x27/0x26") — the board infers from assigned channels; a first-class profile table is a possible future phase if needed.
- Zone-first wiring edits (those are [Phase 69](phase_69_zone_workspace_hub.plan.md); this phase is the hardware-first complement).

---

## Definition of done

- [ ] GPIO board shows every pin/channel → device → zone → role → live state → next run, from live DB (loadAll bug fixed)
- [ ] Re-assign, rename, and manual-pulse a channel inline; conflicts/unassigned flagged
- [ ] Relay-HAT (`hardware_identifier`) actuators are exported to the Pi runtime config
- [ ] Pi client drives multiple actuators per device by `actuator_id`; relay-HAT driver path exists
- [ ] `config_version` bumps on `/assign` so HAT changes hot-reload
- [ ] PiSetupGuide constants live in the Reference tab, clearly labeled vs the live board
- [ ] Go + Pi + Vitest green; smoke_phase70 green; OC-70 closed

---

## Suggested implementation order

1. WS1 board view (read-only first; fix loadAll) — immediate value, proves data
2. WS3 export HAT actuators (backend) — makes board data truthful
3. WS5 config_version on assign — hot-reload correctness
4. WS4 Pi client multi-actuator + HAT driver — edge honors it
5. WS2 inline assign + control + test pulse
6. WS6 reference tab
7. WS7 docs/tests/closure

---

## Related

| Doc | Use |
|-----|-----|
| [phase_50_hardware_wiring_visibility.plan.md](phase_50_hardware_wiring_visibility.plan.md) | Wiring schema + panels |
| [phase_51_pi_config_sync.plan.md](phase_51_pi_config_sync.plan.md) | Config pull-sync + version hot-reload |
| [phase_57_pi_device_api_keys.plan.md](phase_57_pi_device_api_keys.plan.md) | Per-device auth |
| [phase_65_guardian_pi_diagnostics.plan.md](phase_65_guardian_pi_diagnostics.plan.md) | Guardian reads the same wiring/health |
| [ui/src/views/PiSetupGuide.vue](../ui/src/views/PiSetupGuide.vue) | Constants moved to Reference tab |
| [pi_client/gr33n_client.py](../pi_client/gr33n_client.py) | Multi-actuator + HAT driver |
| [internal/hardware/piconfig_json.go](../internal/hardware/piconfig_json.go) | Runtime export gap |

---

## Using this in a new chat

> Read `docs/plans/phase_70_hardware_pi_control_spa.plan.md`. Build the live GPIO board (pin → device → zone → role → schedule, editable). Backend: export relay-HAT actuators to the Pi runtime, make the Pi client multi-actuator + relay-HAT capable, and bump config_version on /assign. Fix the PiSetupGuide loadAll bug; move its constants to a Reference tab. Add smoke_phase70_test.go.
