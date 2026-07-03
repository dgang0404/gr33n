---
name: Phase 119 — Virtual Pi (WS1): pin map foundation + graphical board view
overview: >
  A graphical Raspberry Pi in the UI that renders the real 40-pin header and shows
  what this farm has wired to each pin, driven entirely by existing wiring data
  (sensor/actuator wiring JSONB + relay hardware_identifier). Read-only in this
  phase; gets its own sidebar entry. Foundation for interactive editing (Phase 120)
  and guided hookup/export (Phase 121).
todos:
  - id: ws1-pin-map-lib
    content: "WS1: Pin map library — ui/src/lib/piPinMap.js with the canonical 40-pin table (physical position ↔ BCM number, power/ground/reserved roles, I2C/SPI/UART buses); pure data + helpers, unit tested"
    status: pending
  - id: ws2-board-svg
    content: "WS2: Board component — VirtualPiBoard.vue renders the 2x20 header as SVG/CSS grid; pins colored by role (power/ground/reserved/assigned/free); assignment data resolved from store wiring via hardwareWiring.js"
    status: pending
  - id: ws3-view-route
    content: "WS3: View + route — /virtual-pi view with per-device selector (farm Pis from store.devices), board + assigned-pin legend, link to zone wiring editors; empty state when no Pi registered"
    status: pending
  - id: ws4-sidebar
    content: "WS4: Sidebar — add Wiring entry to navGroups 'Grow & operate' (icon 🔌, navTitle 'Virtual Pi — see what's wired to every pin'); mobile stays unchanged; update phase-78 sidebar tests"
    status: pending
  - id: ws5-tests
    content: "WS5: Tests — pin map table integrity (40 pins, no duplicate BCM, known power/ground positions), assignment resolution from mock store, route + sidebar presence"
    status: pending
isProject: false
---

# Phase 119 — Virtual Pi (WS1): pin map foundation + graphical board view

## Why

Operators already enter wiring per sensor/actuator (Phase 50/51: `wiring.source`,
`gpio_pin`, `i2c_channel`, `serial_port`, `device_id`; relay channels via
`hardware_identifier`). But the farm-wide picture is text-only:

- `ui/src/views/hardware/GpioBoard.vue` (Phase 70) lists rows per device — no
  physical geometry, buried at `/hardware?tab=board`, and `/hardware` was removed
  from the sidebar in Phase 78.
- `ui/src/views/PiSetupGuide.vue` has a nice static reference (stack diagram, DIP
  table, channel map) but it is generic — not this farm's wiring.

A graphical Pi that mirrors the physical board makes "which pin do I plug this
into" a look-at-the-screen job instead of a cross-referencing job.

## Gap found while scoping

**Nothing in the codebase maps BCM GPIO ↔ physical pin position.** Wiring stores
BCM numbers only (`BCM GPIO 4`), but a person standing at the Pi counts physical
pins (pin 7). The pin map library is the real deliverable of this phase; the SVG
is a renderer on top of it.

## Design notes

- **WS1 `piPinMap.js`:** export a 40-entry array:
  `{ physical, bcm|null, role: 'gpio'|'power3v3'|'power5v'|'ground'|'reserved', buses: ['i2c1','spi0','uart0'...] }`.
  Helpers: `pinByBcm(n)`, `pinByPhysical(n)`, `assignmentsForDevice(deviceId, sensors, actuators)`
  (reuses `resolveWiring` from `hardwareWiring.js`).
- **WS2 rendering:** CSS grid of 2×20 is enough (no need for a full SVG
  illustration); each pin is a cell with color + tooltip. Orientation matches
  the board photo convention (pin 1 top-left, USB ports down).
- **I2C devices** (relay HATs, ADS1115, BH1750): render as a chip attached to the
  I2C bus pins (3/5) rather than claiming individual GPIO pins. Relay channels
  listed under the attached HAT.
- **Read-only this phase.** Clicking an assigned pin links to the owning zone's
  wiring editor (`zoneHardwareRoute(zoneId)` from Phase 118 work). Editing from
  the board itself is Phase 120.
- Keep `GpioBoard.vue` as-is (list view still useful); the new view links to it.

## Out of scope (later phases)

- Assigning/editing wiring by clicking a pin (Phase 120)
- DIP-switch/stack visualization derived from registered devices (Phase 120)
- Per-driver hookup diagrams, print view, config.yaml export (Phase 121)

## Acceptance

- [ ] `/virtual-pi` renders a 40-pin header for the selected farm Pi
- [ ] Sensors/actuators with `wiring.gpio_pin` appear on the correct physical pin (BCM→physical verified against pinout.xyz)
- [ ] Power/ground/reserved pins visually distinct and never shown as assignable
- [ ] I2C-attached hardware (relay HAT, ADS1115, BH1750) shown on the bus, not on random pins
- [ ] Sidebar shows the Wiring entry; navigating works on desktop + mobile drawer
- [ ] Empty state when the farm has no registered Pi devices, linking to Pi setup
- [ ] Unit tests: pin table integrity, BCM↔physical mapping, assignment resolution

## Files expected to change

| Area | Files |
|------|-------|
| Pin map lib | `ui/src/lib/piPinMap.js` (new) |
| Components | `ui/src/components/VirtualPiBoard.vue` (new), `ui/src/views/VirtualPi.vue` (new) |
| Routing/nav | `ui/src/router/index.js`, `ui/src/lib/navGroups.js` |
| Tests | `ui/src/__tests__/phase-119-virtual-pi.test.js` (new), `phase-78-closure.test.js` (sidebar expectations) |
