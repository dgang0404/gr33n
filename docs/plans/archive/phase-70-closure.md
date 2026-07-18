# Phase 70 — closure (OC-70)

**Status:** **Shipped** on `main` (v1 — live board + backend/Pi gaps closed; inline per-card edit deferred).

**Canonical plan:** [`phase_70_hardware_pi_control_spa.plan.md`](phase_70_hardware_pi_control_spa.plan.md)

**Depends on:** [Phase 50](phase_50_hardware_wiring_visibility.plan.md) wiring model; [Phase 51](phase_51_pi_config_sync.plan.md) config sync; [Phase 68](phase_68_spa_workspace_shell.plan.md) Hardware workspace shell.

**Closes:** Live GPIO/relay board in the Hardware workspace, relay-HAT export to Pi runtime, multi-actuator Pi client dispatch, and `config_version` bump on HAT assign.

---

## The one job (done)

> **See every Pi channel mapped to zone, role, and live state on one screen** — and have the Pi runtime honor HAT channels the same way it honors direct GPIO wiring.

---

## Workstream checklist

| WS | Deliverable | Verify |
|----|-------------|--------|
| **WS1** | Live `GpioBoard.vue` — relay/GPIO/sensor rows per device | `ui/src/views/hardware/GpioBoard.vue`; `store.loadAll()` |
| **WS2** | Inline assign + pulse per card | **Deferred v2** — v1 read-only board + links; use Controls / Zone hardware panels today |
| **WS3** | Export `hardware_identifier` actuators as `driver: relay_hat` | `internal/hardware/piconfig.go`, `piconfig_json.go` |
| **WS4** | Pi client multi-actuator + relay-HAT driver | `pi_client/gr33n_client.py` — `resolve_actuator_for_command`, `RelayHATActuatorController` |
| **WS5** | `config_version` bump on `/assign` | `db/migrations/20260629_phase70_config_version_assign.sql` |
| **WS6** | Reference tab banner + `loadAll` fix | `PiSetupGuide.vue` embedded mode; Hardware → Reference |
| **WS7** | Tests + smoke | See below |

---

## Automated tests

| Test | Path |
|------|------|
| HAT actuator export shape | `internal/hardware/piconfig_json_test.go` |
| `parseRelayHATChannel` | `internal/hardware/piconfig_test.go` |
| Assign bumps version + pi-config includes `relay_hat` | `cmd/api/smoke_phase70_test.go` |
| Pi multi-actuator + relay-HAT factory | `pi_client/test_gr33n_client.py` (`TestPhase70MultiActuatorDispatch`) |
| Board + Reference tab closure | `ui/src/__tests__/phase-70-closure.test.js` |

---

## Operator path

1. **Hardware → GPIO board** — live farm wiring by Pi device.
2. **Hardware → Reference** — typical HAT diagrams and parts (not farm-specific).
3. Edit wiring via **Controls** or **Zone → hardware** (`ActuatorWiringPanel` / `HardwareWiringPanel`).
4. Pi pulls updated runtime config after assign (`config_version` hot-reload).

---

## OC-70

Phase 70 is **closed** when the GPIO board renders live DB wiring, relay-HAT actuators appear in Pi runtime JSON, the Pi client dispatches by `actuator_id` with a relay-HAT driver path, and assign bumps `config_version`. Inline per-card edit (WS2) remains a follow-up enhancement.
