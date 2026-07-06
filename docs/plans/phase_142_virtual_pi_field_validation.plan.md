---
name: Phase 142 — Virtual Pi field validation arc
overview: >
  Bridge Virtual Pi wiring (119–123), LED simulation rig (125), and Phase 31 field
  validation into one operator path: configure on /virtual-pi, dry-run on the simulation
  rig, then promote to real relays with the same config export contract.
todos:
  - id: ws1-validation-path-doc
    content: "WS1: docs/virtual-pi-field-validation-path.md — ordered checklist virtual-pi → simulation → live Pi"
    status: completed
  - id: ws2-virtual-pi-ui-banner
    content: "WS2: Virtual Pi page — validation status banner (wiring complete, config exported, drift clear)"
    status: pending
  - id: ws3-settings-link
    content: "WS3: Settings edge card links simulation runbook + Phase 31 smokes"
    status: pending
  - id: ws4-smoke-virtual-pi-validation
    content: "WS4: cmd/api smoke — virtual pi export + demo device heartbeat path"
    status: pending
  - id: ws5-closure
    content: "WS5: phase-142-closure test + phase-14 index"
    status: pending
isProject: false
---

# Phase 142 — Virtual Pi field validation arc

**Status:** planned · **Depends on:** [125](phase_125_pi_sensor_actuator_light_simulation.plan.md), [31](phase_31_field_validation_and_edge.plan.md), [121](phase_121_virtual_pi_hookup_export.plan.md)

---

## The one job

> **Wire in the browser → prove on the LED rig → swap driver for real relays** without changing platform APIs.

---

## Wave order

1. **WS1** — Single doc tying `/virtual-pi`, `config.yaml` export, and [pi-light-simulation-runbook.md](../pi-light-simulation-runbook.md)
2. **WS2–WS3** — UI surfaces so operators don't hunt docs
3. **WS4** — Automated smoke for export + config_version bump
4. **WS5** — Closure

---

## Non-goals

- New backend endpoints (reuse Phase 121 push-config + device heartbeat)
- Replacing Phase 31 GPIO hardware smokes

---

## Acceptance (draft)

- [ ] New operator follows one doc from Virtual Pi to simulation demo A
- [ ] Virtual Pi shows clear “ready for dry run” vs “needs wiring” state
- [ ] Smoke covers config export non-empty for demo-veg-relay-01
