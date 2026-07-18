---
name: Phase 120 — Virtual Pi (WS2): interactive wiring + relay stack visualization
overview: >
  Make the Phase 119 board editable: click a free pin to wire a sensor/actuator to
  it, click an assigned pin to edit or move it, with farm-wide conflict checks.
  Add the relay HAT stack view — DIP switch positions and channel map derived from
  the farm's actual devices instead of the static table in the Pi setup guide.
todos:
  - id: ws1-pin-click-assign
    content: "WS1: Assign from board — click free GPIO pin opens a wiring drawer (entity picker: unwired sensors/actuators on that device's zones, driver/source select from deviceTaxonomy); saves via existing PATCH wiring endpoints"
    status: completed
  - id: ws2-edit-move
    content: "WS2: Edit/move — click assigned pin shows current wiring (reuse HardwareWiringPanel form logic); 'move to pin' flow revalidates with findWiringConflict before save"
    status: completed
  - id: ws3-conflict-surface
    content: "WS3: Farm-wide conflict surface — board highlights double-booked pins/channels in red (data already detectable via findWiringConflict, today only checked at edit time per-entity); banner lists conflicts with links to both claimants"
    status: completed
  - id: ws4-relay-stack
    content: "WS4: Relay stack view — derive stack levels from relay channel numbers in use (ch 0–7 = stack 0 @0x27, ch 8–15 = stack 1 @0x26 …); render each card with its DIP switch setting and per-channel actuator labels; empty channels shown as available"
    status: completed
  - id: ws5-tests
    content: "WS5: Tests — assign flow (mock PATCH), conflict highlight from seeded double-booking, stack derivation math (channel → stack level → I2C address → DIP bits)"
    status: completed
isProject: false
---

# Phase 120 — Virtual Pi (WS2): interactive wiring + relay stack visualization

## Status

**Shipped** (2026-07-03). Click-to-wire GPIO pins, conflict banner, relay stack view with DIP hints.

## Why

Phase 119 shows the board; operators will immediately want to fix what they see.
Today wiring edits live inside each zone page (`HardwareWiringPanel.vue`,
`ActuatorWiringPanel.vue`) — fine when you start from the plant, wrong when you
start from the hardware ("this pin is free, what should I put on it?").

## Gaps found while scoping

1. **Conflicts are only checked at edit time, per entity.**
   `findWiringConflict` in `ui/src/lib/hardwareWiring.js` (and the mirrored API
   check) prevents *new* conflicts, but nothing surfaces conflicts that already
   exist in the data (e.g. created before Phase 50, or via direct API). The board
   is the natural place to show them.
2. **The DIP/stack knowledge is static.** `PiSetupGuide.vue` hardcodes the
   address table; nothing tells the operator "your farm uses channels 0–11, so
   you need 2 cards, set the second card's DIP to ID0=ON". Deriving this from
   live data closes the loop.
3. **No "unwired hardware" worklist.** Sensors/actuators with empty wiring are
   only discoverable zone by zone. The board's entity picker doubles as that
   list.

## Design notes

- Reuse, don't fork: the wiring drawer should wrap the same form logic as
  `HardwareWiringPanel` (extract a composable if the panel is too view-coupled —
  prefer extracting over duplicating validation).
- Channel→stack math: `stack = floor(ch / 8)`, `relay = ch % 8`,
  I2C address `0x27 - stack`, DIP bits = 3-bit little-endian of `stack`
  (matches the existing table in `PiSetupGuide.vue` — assert parity in tests).
- Moving a pin is a single PATCH of the wiring JSONB; no new API needed. If the
  API-side conflict check rejects, surface its message verbatim.
- Guardian tie-in (cheap win): "Ask Guardian" starter chip on the board view —
  "What's free on my Pi for a new pump?" — context ref includes the pin
  assignment summary. Skip if it bloats the phase.

## Out of scope

- Physical wiring diagrams per driver (how to hook a DHT22's 3 wires) — Phase 121
- Printable/export views — Phase 121
- pi_client config generation — Phase 121

## Acceptance

- [x] Wire a brand-new sensor to a free pin entirely from `/virtual-pi`
- [x] Move an actuator to a different pin; old pin frees, conflict check blocks a taken pin
- [x] Pre-existing double-booked pin renders red with both claimants linked
- [x] Relay stack view shows correct card count, DIP settings, and channel labels for the demo farm
- [x] Adding an actuator on ch 8 makes a second card appear with ID0=ON
- [x] Existing zone wiring editors untouched and passing (`phase-78`, `zone-feeding-water` suites)

## Files expected to change

| Area | Files |
|------|-------|
| Board interactivity | `ui/src/components/VirtualPiBoard.vue`, new `ui/src/components/PinWiringDrawer.vue` |
| Shared form logic | `ui/src/components/HardwareWiringPanel.vue` (extract composable `ui/src/composables/useWiringForm.js`) |
| Stack view | new `ui/src/components/RelayStackView.vue`, `ui/src/lib/relayStack.js` |
| View | `ui/src/views/VirtualPi.vue` |
| Tests | `ui/src/__tests__/phase-120-virtual-pi-wiring.test.js` (new) |
