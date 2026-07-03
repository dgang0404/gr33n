---
name: Phase 121 — Virtual Pi (WS3): guided hookup, print view, pi_client config export
overview: >
  Close the loop from screen to grow room: per-driver hookup instructions (which
  physical wires go where for a DHT22, ADS1115, MH-Z19, BH1750, relay HAT),
  a printable wiring sheet, and one-click generation of the pi_client config.yaml
  from the farm's wiring — plus a drift check between what the UI says and what
  the Pi is actually running.
todos:
  - id: ws1-driver-hookups
    content: "WS1: Driver hookup data — extend device taxonomy (API /platform/device-taxonomy + fallback) with per-driver pin requirements (DHT22: VCC 3v3 / DATA gpio / GND; ADS1115: I2C + addr; MH-Z19: UART TX/RX; BH1750: I2C); render as step list when a pin/driver is selected on the board"
    status: completed
  - id: ws2-print-sheet
    content: "WS2: Printable wiring sheet — print CSS route of the virtual board + pin table + relay stack + per-entity hookup steps; works from browser print dialog, no PDF dependency"
    status: completed
  - id: ws3-config-export
    content: "WS3: config.yaml export — GET /devices/{id}/pi-config renders pi_client config.yaml (sensors, actuators, relay channels, pins, intervals) from DB wiring; UI download button on /virtual-pi; keep parity with pi_client/config.bootstrap.example.yaml schema"
    status: completed
  - id: ws4-drift-check
    content: "WS4: Drift check — pi_client heartbeat already posts its config hash/contents (verify; add if missing); API compares against generated config and flags 'Pi running stale wiring' on the device; board shows amber badge"
    status: completed
  - id: ws5-docs-tests
    content: "WS5: Docs + tests — INSTALL.md / pi-setup guide point at export instead of hand-editing; Go tests for config rendering; UI tests for hookup steps + print route; openapi.yaml for new endpoint"
    status: completed
isProject: false
---

# Phase 121 — Virtual Pi (WS3): guided hookup, print view, pi_client config export

**Status: shipped**

## Why

Phases 119–120 make the UI the source of truth for wiring. But the operator still
hand-copies that truth twice: once onto the physical wires, once into
`pi_client/config.yaml`. Both copies drift. This phase generates both artifacts
from the same data the board renders.

## Acceptance

- [x] Selecting DHT22 on a pin shows its 3-wire hookup steps with the right physical pins highlighted
- [x] Print dialog output is legible on one–two A4/letter pages for the demo farm
- [x] Downloaded config.yaml validates against pi_client's loader (existing round-trip tests)
- [x] Changing a pin in the UI flips the device to "stale wiring" after next heartbeat with the old hash
- [x] openapi.yaml documents `/devices/{id}/pi-config`; `make audit-openapi` passes
- [x] Pi setup guide links to export; INSTALL.md updated
