---
name: Phase 125 — Pi light/LED simulation rig for sensors + actuators (pre-plant dry run)
overview: >
  Before real plants go into any bed, give the operator a physical way to see
  the platform "think" without anything alive on the line. A Raspberry Pi
  (the same Virtual Pi hardware from Phase 121, or a second one) drives a
  small strip/array of addressable LEDs (WS2812/NeoPixel) plus a couple of
  plain GPIO LEDs. Each light or blink pattern stands in for one sensor
  reading or actuator action, so a person standing in front of the rig can
  watch the automation loop (sensor reading -> target band -> automation
  decision -> pump/light/fan command -> device) happen in real time, entirely
  in simulation. This is a dry-run harness, not new backend logic — it should
  reuse the existing device/actuator/sensor APIs and command queue exactly as
  a real Pi would, so the same wiring later drives real relays/pumps with no
  code changes on the platform side, only a driver swap on the Pi.
todos:
  - id: ws1-mapping-spec
    content: "WS1: Sensor/actuator -> light mapping spec (doc): which color/zone LED = which sensor; which blink pattern = which actuator action/state (idle, running, fault)"
    status: completed
  - id: ws2-pi-light-driver
    content: "WS2: pi_client light driver — consumes existing device/actuator state (poll or SSE) and drives NeoPixel/GPIO instead of a real relay; config-driven pin/zone mapping, no new server-side endpoints"
    status: completed
  - id: ws3-simulated-sensor-loopback
    content: "WS3: Simulated sensor loopback mode — pi_client can optionally generate synthetic sensor readings (e.g. slow sine-wave EC/moisture drift, or a manual test-panel script) and POST them through the normal /sensors/{id}/readings path so the full automation loop fires for real"
    status: completed
  - id: ws4-demo-script
    content: "WS4: Guided demo script/checklist — a short runbook for triggering each automation path on the rig (e.g. drop simulated soil moisture -> watch irrigation light blink -> confirm task/alert appears in UI) for hands-on walkthroughs"
    status: completed
  - id: ws5-docs
    content: "WS5: Docs — wiring diagram, parts list (Pi, LED strip, resistors/level shifter if needed), and README for swapping simulation driver for real relay driver later"
    status: completed
isProject: false
---

# Phase 125 — Pi light/LED simulation rig for sensors + actuators (pre-plant dry run)

**Status: shipped** — [`pi-light-simulation-mapping.md`](../pi-light-simulation-mapping.md) · [`pi-light-simulation-runbook.md`](../pi-light-simulation-runbook.md) · [`pi_client/README-simulation-rig.md`](../pi_client/README-simulation-rig.md)

## Why

Real plants are slow, expensive to mess up, and hide problems (a bad
automation rule might not visibly hurt anything for days). A lights-only rig
lets the operator abuse the automation logic — trip thresholds, force
faults, watch the pump/light/fan command path — and see it happen instantly
and safely, with zero risk to a crop. It's also just a good demo: anyone can
walk up, watch colored lights react, and understand what the platform does
without reading a dashboard.

## Design sketch (normative detail in [`pi-light-simulation-mapping.md`](../pi-light-simulation-mapping.md))

- **One light/zone per sensor** (e.g. soil moisture, EC, pH, temp, humidity)
  — color encodes "in band" (green) vs "low" (blue-ish/blink) vs "high"
  (red/blink), matching whatever band language the UI already uses.
- **One light per actuator** (pump, light, fan) — solid = idle, blink = the
  device is actively running/commanded on, fast-blink or distinct color =
  fault/offline.
- Runs on the same `pi_client` used for Phase 121's Virtual Pi hookup, in a
  new "simulation" driver mode — same config export / heartbeat / drift
  badge machinery already shipped, just a different bottom-layer driver
  (LEDs instead of real GPIO relays) so swapping to real hardware later is a
  config change, not a rewrite.
- Simulated sensor input can be either scripted (deterministic demo path) or
  manual (a small local test panel / CLI to nudge a reading up/down on
  demand) — both go through the real `/sensors/{id}/readings` endpoint so
  the automation engine, alerts, and task creation all fire exactly as they
  would for a live sensor.

## Explicit non-goals for this phase

- No new backend endpoints or schema — this is a client-side driver + a
  demo script, reusing Phase 121's config/heartbeat contract.
- No attempt to model real plant physiology — thresholds can be arbitrary/
  fast for demo purposes (e.g. a 30-second "day" instead of a real one).
- Not a substitute for real sensor calibration work — this is strictly a
  before-plants confidence/demo tool.

## Acceptance (draft — refine once WS1 mapping spec exists)

- [x] Mapping spec document lists every sensor/actuator type in the demo
      farm and its light/color/blink assignment
- [x] `pi_client` simulation driver mode boots, reads config the same way
      the real driver does, and drives at least one LED per mapped
      sensor/actuator
- [x] A synthetic sensor reading pushed through the normal ingestion path
      visibly changes the correct light within a few seconds
- [x] A forced out-of-band reading triggers the same alert/automation path
      as production and the corresponding actuator light blinks
- [x] Demo runbook exists and was walked through end-to-end at least once
- [x] README documents the exact steps to swap the simulation driver for a
      real relay/pump driver with no platform-side changes
