---
name: Phase 38 — Plant-needs UI + timed actuator commands
overview: >
  Reorganize the operator UI around what a plant needs (Water / Light / Air & Climate)
  with the zone as the primary lens, and add real timed/pulse actuator commands
  (e.g. pump on for 2s) end-to-end.
todos:
  - id: ws1-need-helper
    content: "WS1: Frontend need-classification helper (ui/src/lib/plantNeeds.js)"
    status: done
  - id: ws2-zone-hub
    content: "WS2: ZoneDetail need tabs (Overview/Water/Light/Air) for all zones"
    status: done
  - id: ws3-connection-card
    content: "WS3: Per-need connection card (reading -> target -> automation -> control)"
    status: done
  - id: ws4-nav-ia
    content: "WS4: SideNav/App IA — Grow primary, Advanced devices group"
    status: done
  - id: ws5-pulse-backend
    content: "WS5: Backend duration_seconds in pending_command + fertigation wiring"
    status: done
  - id: ws6-pulse-pi
    content: "WS6: Pi pulse execution (on -> wait -> off)"
    status: done
  - id: ws7-pulse-ui
    content: "WS7: UI run-for-N-seconds in zone Water + Controls"
    status: done
  - id: ws8-docs-tests
    content: "WS8: operator-tour, OpenAPI, smokes"
    status: done
isProject: false
---

# Phase 38 — Plant-needs UI + timed actuator commands

## Status

**Shipped.** Zone hub uses plant-need tabs; navigation regrouped; `duration_seconds` on `pending_command` with Pi pulse support.

## Related

| Doc | Use |
|-----|-----|
| [phase_35_lighting_domain.plan.md](phase_35_lighting_domain.plan.md) | Light need |
| [phase_36_greenhouse_climate.plan.md](phase_36_greenhouse_climate.plan.md) | Air/climate need |
| [phase_39_edge_fertigation_execution.plan.md](phase_39_edge_fertigation_execution.plan.md) | **Next:** mix jobs + device command queue (fixes last-write-wins); automated recipe→Pi |
| [phase_40_unified_farmer_ux_zone_cockpit.plan.md](phase_40_unified_farmer_ux_zone_cockpit.plan.md) | **After 39:** zone cockpit — inline setpoints, alerts, today strip (not DB-shaped UI) |
| [phase_41_farm_hub_coherence.plan.md](phase_41_farm_hub_coherence.plan.md) | **After 40:** Dashboard + farm-wide pages + why-empty |
| [pre_development_gaps_index.plan.md](pre_development_gaps_index.plan.md) | Gap tracker before dev |
| [operator-tour.md](../operator-tour.md) | § Plant needs (Phase 38) |

## Not in Phase 38

- **Device command queue** — still one `pending_command` slot until Phase 39 WS1.
- **Automated EC mixing on the Pi** — mixing remains operator-logged via API; programs only pulse pumps when `run_duration_seconds` is set.
