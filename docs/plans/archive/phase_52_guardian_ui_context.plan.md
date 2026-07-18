---
name: Phase 52 — Guardian UI context & Pi setup discoverability
overview: >
  Guardian knows where the operator is in the app (current route, last 3 pages, page-specific
  framing) without the user saying "I'm on this screen." Pi + HAT setup gets a farmer-facing
  in-app guide, sidebar placement, and nav-hint wiggle chains from wiring/offline badges.
todos:
  - id: ws1-nav-history
    content: "WS1: guardianPanel.navHistory + POST nav_history; renderRouteContext breadcrumb"
    status: completed
  - id: ws2-pi-setup-guide
    content: "WS2: PiSetupGuide.vue at /pi-setup + docs/pi-sequent-hat-setup.md"
    status: completed
  - id: ws3-sidebar-wiggles
    content: "WS3: Guide child link, empty-state hints, wiring/offline/config-stale nav-hint chains"
    status: completed
  - id: ws4-starter-cleanup
    content: "WS4: Setup starters drop redundant location prefix; operator-tour mention"
    status: completed
isProject: false
---

# Phase 52 — Guardian UI context & Pi setup discoverability

## Status

**Shipped** on `main` (commits through `804e930`). **OC-52** closed in [operational closure](phase_35_37_operational_closure.plan.md).

**Not the same as:** Phase 51 plan [Out of scope](phase_51_pi_config_sync.plan.md#out-of-scope) **"Phase 52+ per-device API keys"** — shipped as **[Phase 57](phase_57_pi_device_api_keys.plan.md)**.

**Historical:** "Next in arc" below is deprecated — phases 53–55 also shipped.

---

## The one job

> **Guardian and the sidebar should know what screen the operator is on — and where they came from.**

| Farmer moment | Before | After |
|---------------|--------|-------|
| Clicks "Walk me through setup" | Message says "I'm in farm setup…" | Starter is action-only; route + nav trail in system prompt |
| On Sensors, asks about wiring | No page context | `context_ref` route + Sensors framing + nav trail |
| Needs Pi hardware help | Link → generic Operator guide | `/pi-setup` Sequent HAT guide + sidebar under Guide |
| Hovers wiring / offline | No affordance | Sidebar wiggles Pi setup → Sensors + Controls |

---

## WS1 — Navigation history (Guardian)

| Artifact | Purpose |
|----------|---------|
| `guardianPanel.navHistory` | Last 3 unique routes before current (deduped) |
| `POST /v1/chat` `nav_history` | Breadcrumb trail in `renderRouteContext` |
| `ContextRefPromptBlock` | Per-page framing hints (sensors, actuators, setup wizards, …) |

---

## WS2 — Pi + HAT setup guide

| Artifact | Purpose |
|----------|---------|
| `ui/src/views/PiSetupGuide.vue` | Visual Sequent stack guide (DIP table, channel map, wiring plan) |
| `docs/pi-sequent-hat-setup.md` | Repo reference |
| `emptyStateHint` `no_telemetry` | Links to `/pi-setup` |

### WS2 shipped extension — live farm channels (post-57)

**Shipped.** `/pi-setup` was upgraded from a static reference into a **live wiring view** using `useFarmStore`:

| Surface | Behaviour |
|---------|-----------|
| **Reference channel-map cards** | A channel that is wired to an actuator shows the actuator name and becomes a `router-link` to its detail (with `v-nav-hint` sidebar wiggle). `slotActuator(channel)` resolves the overlay. |
| **"Your farm channels"** section (`data-test="pi-setup-live-wiring"`) | Lists each wired device with its actuators (by relay channel) and sensors (by GPIO/source), each linking to the actuator/sensor page. Empty state links to Settings → Devices / Controls / Sensors. |
| Computeds | `actuatorByChannel`, `sensorByDevicePin`, `wiredDevices`, `deviceActuators`, `deviceSensors`, `slotActuator` |

This pairs with the **Phase 50 shipped extension** (Controls → **Edit wiring** actuator editor + `PATCH /actuators/{id}/assign`): the guide now reads the same `hardware_identifier` / `config.wiring` data operators edit on the Controls page.

---

## WS3 — Sidebar discoverability & wiggle chains

| Artifact | Purpose |
|----------|---------|
| `navGroups.js` `children` under Guide | Pi + HAT setup sub-link |
| `SideNav.vue` | Indented sub-items when expanded |
| `HardwareWiringBadge` `v-nav-hint` | Wiring → Pi setup |
| `navRelations.js` | `/pi-setup` ↔ sensors/actuators; operator-guide → pi-setup |
| High-impact batch | Offline badge, config stale, comfort targets link, Operator guide hints |

---

## Definition of done

- [x] Grounded chat sends `context_ref` + `nav_history` on every turn
- [x] `/pi-setup` reachable from sidebar, zone empty states, operator tour
- [x] Wiring and hardware problem surfaces wiggle Pi setup in sidebar
- [x] Go + Vitest tests for nav history and context_ref trail

---

## Related

- [phase_44_getting_started_edge_wizard.plan.md](phase_44_getting_started_edge_wizard.plan.md) — device wizard (complement Pi guide)
- [phase_51_pi_config_sync.plan.md](phase_51_pi_config_sync.plan.md) — platform config sync (API keys = future phase)
