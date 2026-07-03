---
name: Phase 114 — Pi/edge integrity (device health, actuator truth, calibration)
overview: >
  Fix the gaps in the sensor → API → UI chain found in the July 2026 audit: devices
  never marked offline on stale heartbeat, ActuatorCard toggling the wrong endpoint,
  Pi never reporting firmware version or battery/signal, mixing-event audit posting to
  a JWT-only route from the Pi, relay-HAT driver that logs instead of switching, no
  calibration workflow, and no per-command queue inspector.
todos:
  - id: ws1-stale-heartbeat
    content: "WS1: Offline detection — worker job marks devices offline when last_heartbeat older than threshold (default 3× report interval, env override); Pi graceful-shutdown PATCH offline; alert on transition"
    status: pending
  - id: ws2-actuator-card-fix
    content: "WS2: ActuatorCard bug — replace toggleDevice() PATCH /devices/{id}/status (Pi-key-only route!) with enqueueActuatorCommand; reflect queued/confirmed state"
    status: pending
  - id: ws3-firmware-telemetry
    content: "WS3: Edge telemetry — Pi status PATCH reports client_version + uptime; accept battery_level_percent / signal_strength_dbm in readings; offline flush uses batch endpoint; surface version + battery in DeviceDetail"
    status: pending
  - id: ws4-mix-audit-route
    content: "WS4: Mixing-event audit — Pi-key route POST for mixing events (or fold into command ack payload); Pi client posts with device key; UI 'Mix now' button calls existing mix-jobs API"
    status: pending
  - id: ws5-relay-hat
    content: "WS5: Relay HAT — implement smbus I/O in RelayHATActuatorController (Sequent 8-relay register map) or hard-disable HAT channels in wizard UI until implemented; bench-test doc"
    status: pending
  - id: ws6-calibration
    content: "WS6: Sensor calibration — API to save calibration points (raw vs reference), compute slope/offset into sensors.calibration_data; Pi pulls calibration in config sync and applies instead of hardcoded EC/pH formulas; SensorDetail calibration wizard"
    status: pending
  - id: ws7-queue-inspector
    content: "WS7: Command queue inspector — per-device command list UI (status, payload, age); cancel pending; resulting_state + execution_status written on Pi ack; error_comms/error_hardware used by Pi client"
    status: pending
isProject: false
---

# Phase 114 — Pi/edge integrity (device health, actuator truth, calibration)

## Status

**Planned.** From the July 2026 audit, Pi-chain findings. Sibling phases:
[113](phase_113_security_hardening.plan.md) security,
[115](phase_115_schema_utilization.plan.md) schema surfacing,
[116](phase_116_docs_refresh.plan.md) docs, [117](phase_117_test_depth.plan.md) tests.

---

## Findings driving this phase

| # | Finding | Why an operator cares |
|---|---------|----------------------|
| 1 | Pi only ever PATCHes `status=online`; nothing marks stale devices offline | Powered-down Pi stays green on the dashboard forever; automation keeps queueing commands into the void |
| 2 | `ActuatorCard.vue` calls `toggleDevice()` → `PATCH /devices/{id}/status`, a **Pi-key-only** route | The primary hardware card 401s or flips device online/offline instead of the relay |
| 3 | Pi never reports `firmware_version`, battery, or signal; offline flush replays readings one-by-one instead of batch | Field bugs can't be correlated to client versions; wireless sensor health invisible; slow reconnect drain |
| 4 | Pi `post_mixing_event()` targets a JWT-only route → silently fails with device key; UI never calls `mix-jobs` enqueue API | Edge fertigation runs with **no audit rows** — a traceability gap |
| 5 | `RelayHATActuatorController.turn_on/off` only logs — no smbus I/O | Sequent HAT deployments look wired in the UI but never actuate |
| 6 | `sensors.calibration_data` schema exists; Pi uses hardcoded EC/pH linear formulas; no UI workflow | EC/pH drift silently corrupts automation thresholds and Guardian advice |
| 7 | Queue UI shows aggregate depth only; `resulting_state_*`, `execution_status`, `error_comms`/`error_hardware` enum values never populated | "2 queued" but no way to see what's stuck, failed, or why |

---

## Design notes

### WS1 — Offline detection

The automation worker already polls every ~1 min; add a sweep: devices with
`last_heartbeat < now() - threshold` transition `online → offline` and raise an
alert (`device_offline`). Threshold default `DEVICE_OFFLINE_AFTER_SECONDS=900`,
overridable per farm later. Pi client installs a SIGTERM hook to PATCH `offline`
on clean shutdown so planned reboots don't page anyone.

### WS2 — ActuatorCard

This is a straight bug fix and ships first. Card gets the same pulse/queue affordances
as `ActuatorPulseControl` (on / off / pulse N seconds), and shows *queued* vs
*confirmed* (confirmed = Pi ack with `resulting_state`, which WS7 starts writing).

### WS6 — Calibration

Two-point calibration for EC/pH (buffer solutions), single-offset for temperature.
Operator enters reference value while sensor sits in buffer; API stores points and
computed slope/offset in `sensors.calibration_data` JSONB, stamps
`last_calibration_date`, sets `is_calibrated`. Pi config sync already hot-reloads —
extend payload with calibration coefficients; Pi applies them in the read path,
falling back to current formulas when absent. Guardian read tools can then cite
"calibrated 3 days ago" vs "never calibrated" in fertigation answers.

### Out of scope

- New sensor hardware drivers (this phase fixes the chain, not the catalog)
- MQTT transport (polling stays; a transport swap is its own phase)
- Multi-Pi orchestration / fleet dashboards

---

## Acceptance

- [ ] Unplugging a Pi flips its device card to offline within threshold + one poll; alert raised; replug restores online
- [ ] ActuatorCard on/off/pulse enqueues device commands; no JWT calls to Pi-key routes remain in `ui/src`
- [ ] Device detail shows client version + last heartbeat age; battery/signal shown when reported
- [ ] Pi `mix_batch` execution produces a mixing-event row visible in fertigation history; UI "Mix now" enqueues via `mix-jobs`
- [ ] Relay HAT either switches real relays (bench-verified, `-tags hardware` smoke) or is visibly marked unsupported in the wizard
- [ ] Calibrating a pH sensor changes subsequent computed readings; calibration date shown in UI and Guardian zone summary
- [ ] Queue inspector lists per-command status incl. failed with error class; cancel works on pending commands
- [ ] Offline flush uses the batch readings endpoint (single POST per flush cycle)

---

## Files expected to change

| Area | Files |
|------|-------|
| Worker | `internal/automation/*` (offline sweep) |
| Handlers | `internal/handler/device/*`, `internal/handler/actuator/*`, fertigation mixing route |
| Pi client | `pi_client/gr33n_client.py` (version report, batch flush, SIGTERM hook, calibration, HAT smbus, error enums) |
| UI | `ui/src/components/ActuatorCard.vue`, `SensorDetail.vue`, device detail, queue inspector view |
| Schema | migration only if invite-style additions needed (calibration uses existing columns) |
| Tests | `cmd/api/smoke_phase114_*.go`, Pi client unit tests, `-tags hardware` bench smoke |
