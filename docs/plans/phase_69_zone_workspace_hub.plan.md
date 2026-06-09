---
name: Phase 69 — Zone workspace as the single operational hub
overview: >
  Make the zone the place where a whole job lives. ZoneDetail already composes
  sensors, automation, controls, lighting and feeding per need-tab, but it only
  *summarizes* them and links out to farm-wide pages to edit. This phase makes the
  zone need-tabs fully edit-capable inline (wire/assign sensors, toggle + tune
  controls, edit the lighting program) so the operator never has to bounce to
  /sensors, /actuators or /lighting for zone work. The farm-wide pages collapse
  into a single "Fleet" tab in the Zones workspace for cross-zone admin. UI-only.
todos:
  - id: ws1-zone-tab-inline-edit
    content: "WS1: ZoneNeedSection — promote summary tiles to inline editors (sensor wiring/assign, actuator tune, comfort target edit) using existing PATCH endpoints; no link-out for in-zone work"
    status: pending
  - id: ws2-zone-lighting-inline
    content: "WS2: Light tab edits the zone's lighting program inline (photoperiod, schedule) instead of read-only summary + /lighting link"
    status: pending
  - id: ws3-fleet-tab
    content: "WS3: Zones workspace 'Fleet' tab — farm-wide Sensors + Controls + Lighting as one cross-zone admin surface (filter/group by zone); absorbs Sensors.vue/Actuators.vue/LightingPrograms.vue"
    status: pending
  - id: ws4-zone-overview-spine
    content: "WS4: Zone Overview becomes the spine — pin GPIO/device pipeline, next runs, alerts, active grow; one-screen 'what is this zone doing right now'"
    status: pending
  - id: ws5-detail-redirects
    content: "WS5: Redirect /sensors, /actuators, /lighting → /zones?tab=fleet; keep /sensors/:id, /zones/:id detail; update cross-links + wiggle targets"
    status: pending
  - id: ws6-docs-tests
    content: "WS6: zone inline-edit Vitest, fleet-tab test, phase-69-closure.test.js; operator-tour zone section; OC-69"
    status: pending
isProject: false
---

# Phase 69 — Zone workspace as the single operational hub

## Status

**Planned.** Builds on the [workspace shell](phase_68_workspace_shell_spa_nav.plan.md) (Phase 68). UI-only — reuses existing wiring/command/setpoint endpoints; no schema, no new API, no Pi.

**Closure:** **OC-69** — tracked in this plan's DoD + [arc hub OC table](phase_68_73_spa_workspace_roadmap.plan.md#operational-closure-oc-rows). Do not add to the archived Phase 35 closure doc.

---

## The one job

> **Open a zone, see and *change* everything it does — sensors, pumps, lights, targets, schedules — without leaving the page.**

This is the screen the operator pointed at: the Zone Details view "has the most info, but from there you've got to jump around." Phase 69 removes the jumping.

---

## Problem

[`ui/src/views/ZoneDetail.vue`](../ui/src/views/ZoneDetail.vue) is already the integration hub — Overview + Water / Light / Climate need-tabs (driven by [`ui/src/lib/plantNeeds.js`](../ui/src/lib/plantNeeds.js)), each composing ~10 shared components ([`ZoneNeedSection.vue`](../ui/src/components/ZoneNeedSection.vue), [`ZoneAutomationPanel.vue`](../ui/src/components/ZoneAutomationPanel.vue), [`ZoneWaterGrowStory.vue`](../ui/src/components/ZoneWaterGrowStory.vue), `ZoneComfortTargets.vue`, `SensorTile.vue`, actuator tiles, …).

But the need-tabs **summarize and link out**:

| In a zone need-tab today | To actually edit, you leave to |
|--------------------------|-------------------------------|
| Sensor tiles (read live value) | `/sensors` → `/sensors/:id` (wiring panel, rules) |
| Actuator toggles (on/off + pulse) | `/actuators` (Controls — wiring, rename) |
| Lighting summary (read-only) | `/lighting?zone_id=…` (program CRUD) |
| Comfort targets (mostly inline ✅) | `/setpoints` for raw scope |

So the farm-wide pages (`Sensors`, `Controls`, `Lighting`) exist mostly to **edit things that belong to a zone** — which is exactly the overlap the operator flagged ("sensor / control / plant / lights all kinda the same as the zone page"). The fix: let the zone edit its own hardware inline, and keep one **cross-zone Fleet** admin surface for the genuinely farm-wide view (audit all pins, bulk actions).

---

## Design principles

1. **In-zone work stays in-zone.** Anything scoped to one zone (wire a sensor, tune a pump, set a photoperiod) is editable in the zone tab. No link-out for zone-scoped edits.
2. **Reuse, don't rebuild.** Embed the existing editors (`HardwareWiringPanel.vue`, `ActuatorWiringPanel.vue`, lighting program form) inside the zone tabs — same PATCH endpoints, just hosted in context.
3. **Fleet = the only farm-wide view.** Sensors/Controls/Lighting collapse into one "Fleet" tab grouped by zone, for cross-zone admin and conflict auditing — not three sidebar entries.
4. **Overview is the spine.** "What is this zone doing right now" — pipeline (GPIO → device → zone), next scheduled runs, alerts, active grow — answerable without opening a tab.
5. **Contract-safe.** No deleted routes; detail pages survive; redirects + wiggle updated.

---

## WS1 — Inline edit in zone need-tabs

Promote [`ui/src/components/ZoneNeedSection.vue`](../ui/src/components/ZoneNeedSection.vue) tiles from read-only to edit-in-place, reusing shipped panels:

- **Sensors:** each `SensorTile` gains an "Edit wiring" affordance that opens [`HardwareWiringPanel.vue`](../ui/src/components/HardwareWiringPanel.vue) inline → `PATCH /sensors/{id}/wiring`. Assign a sensor to the zone/device without leaving.
- **Controls:** actuator tiles embed [`ActuatorWiringPanel.vue`](../ui/src/components/ActuatorWiringPanel.vue) (HAT channel via `PATCH /actuators/{id}/assign` or GPIO via `PATCH /actuators/{id}/wiring`) + rename, alongside the existing toggle/pulse (`POST /actuators/{id}/command`).
- **Comfort targets:** keep the existing inline `ZoneComfortTargets` editing; add a "raw scope" expander for power users instead of bouncing to `/setpoints`.
- **Automation/schedules:** `ZoneAutomationPanel` already lists rules/schedules; add inline enable/pause (existing `PATCH /schedules/{id}`, rule patch) so the operator doesn't open `/schedules` or `/automation` for a zone toggle.

Link-outs to farm-wide pages remain only as "see all zones" affordances, not as the way to edit *this* zone.

---

## WS2 — Lighting inline in the Light tab

Today the Light tab shows a read-only program summary and links to `/lighting?zone_id=…`. Embed the lighting program editor (from [`ui/src/views/LightingPrograms.vue`](../ui/src/views/LightingPrograms.vue), refactored into a reusable `LightingProgramForm.vue`) directly in the zone Light tab:

- Edit photoperiod / on-off times for this zone's program (creates/updates the paired ON/OFF schedules via the existing lighting endpoints — see [`internal/handler/lighting/`](../internal/handler/lighting/)).
- The farm-wide lighting list moves to the Fleet tab (WS3).

---

## WS3 — Zones workspace "Fleet" tab

The Phase 68 `ZonesWorkspace` has tabs **Rooms** (zone list) + **Fleet**. Build Fleet as the single cross-zone admin surface that absorbs the three farm-wide pages:

| Fleet sub-view | Source today | Purpose |
|----------------|--------------|---------|
| Sensors | [`Sensors.vue`](../ui/src/views/Sensors.vue) | All sensors, live readings, wiring badges, grouped/filtered by zone; link to `/sensors/:id` |
| Controls | [`Actuators.vue`](../ui/src/views/Actuators.vue) | All actuators farm-wide, toggle/pulse, wiring; flag pin conflicts |
| Lighting | [`LightingPrograms.vue`](../ui/src/views/LightingPrograms.vue) | All photoperiod programs across zones |

Fleet is for **cross-zone** work ("show me every pump," "audit all GPIO pins," "which zones have stale sensors"). Single-zone work belongs in the zone tabs (WS1/WS2).

---

## WS4 — Overview as the spine

Tighten `ZoneDetail` Overview into a one-screen "right now" status (the operator's "most info" screen, made self-sufficient):

- **Connection pipeline** front and center: GPIO/channel → device → this zone, with online/offline (reuse the existing pipeline + [`ui/src/lib/hardwareWiring.js`](../ui/src/lib/hardwareWiring.js) labels like `GPIO_RELAY · BCM GPIO 27`).
- **Next runs:** the next scheduled feed/light/climate action per need, inline.
- **Alerts + active grow + tasks** strips (already present) kept above the fold.
- Each summary tile deep-links to its need-tab (`?tab=water|light|air`) — the wiggle points there.

---

## WS5 — Redirects & cross-links

- `/sensors`, `/actuators`, `/lighting` → `redirect` to `/zones?tab=fleet` (added in Phase 68 WS4; confirm targets land on the right Fleet sub-view via a secondary query, e.g. `?tab=fleet&fleet=controls`).
- **Keep** `/sensors/:id` (`SensorDetail.vue`) and `/zones/:id` — detail pages.
- Update `v-nav-hint` targets and [`navRelations.js`](../ui/src/lib/navRelations.js): in-zone "see all" links wiggle the Zones workspace; remove now-dead hints to `/lighting` etc.
- Audit Guardian route refs in [`context_ref.go`](../internal/farmguardian/context_ref.go) for `/sensors`/`/actuators`/`/lighting` and re-point to zone tabs or Fleet.

---

## WS6 — Docs, tests, closure (OC-69)

| Artifact | Content |
|----------|---------|
| `ui/src/__tests__/zone-inline-edit.test.js` (new) | Sensor wiring panel + actuator assign render and PATCH inside zone tab |
| `ui/src/__tests__/zone-fleet-tab.test.js` (new) | Fleet tab groups sensors/controls/lighting by zone; pin-conflict flag |
| `ui/src/__tests__/phase-69-closure.test.js` (new) | Closure: inline edit present, fleet tab present, farm-wide routes redirect |
| [operator-tour.md](../operator-tour.md) | Zone section: "edit everything here; Fleet for cross-zone" |

**OC-69** added and closed when WS1–WS6 ship.

---

## Out of scope

- The **live GPIO board** view (pin-first, all zones at once) — that's the Hardware workspace, [Phase 70](phase_70_hardware_pi_control_spa.plan.md). Phase 69 keeps wiring edits *zone-first*.
- Feed/water plan consolidation — [Phase 71](phase_71_feed_water_unification.plan.md) (the Water tab keeps `ZoneWaterGrowStory`; deep "Advanced feeding" goes to the Feed & Water workspace).
- Any schema/API/Pi change (reuses existing PATCH/command endpoints).

---

## Definition of done

- [ ] Sensors/controls in a zone are wired, assigned, renamed, toggled and tuned **inline** — no link-out for zone-scoped edits
- [ ] Zone Light tab edits the zone's lighting program inline
- [ ] Zones workspace Fleet tab presents farm-wide sensors + controls + lighting grouped by zone
- [ ] Overview answers "what is this zone doing right now" without opening a tab
- [ ] `/sensors`, `/actuators`, `/lighting` redirect to Fleet; detail/param routes still resolve
- [ ] Vitest green; OC-69 closed

---

## Suggested implementation order

1. WS3 Fleet tab (move existing views into the workspace — low risk, immediate dedupe)
2. WS1 inline sensor/control edit (embed shipped panels)
3. WS2 lighting inline (extract `LightingProgramForm`)
4. WS4 Overview spine
5. WS5 redirects + wiggle cleanup
6. WS6 closure

---

## Related

| Doc | Use |
|-----|-----|
| [phase_68_workspace_shell_spa_nav.plan.md](phase_68_workspace_shell_spa_nav.plan.md) | Shell + Zones workspace tabs |
| [phase_40_unified_farmer_ux_zone_cockpit.plan.md](phase_40_unified_farmer_ux_zone_cockpit.plan.md) | Original zone cockpit design |
| [phase_50_hardware_wiring_visibility.plan.md](phase_50_hardware_wiring_visibility.plan.md) | Wiring panels reused inline |
| [ui/src/components/ZoneNeedSection.vue](../ui/src/components/ZoneNeedSection.vue) | Primary component edited |

---

## Using this in a new chat

> Read `docs/plans/phase_69_zone_workspace_hub.plan.md`. UI-only. Make zone need-tabs edit sensors/controls/lighting inline by embedding the existing wiring panels (same PATCH endpoints). Collapse farm-wide Sensors/Controls/Lighting into a Zones-workspace Fleet tab. Keep detail routes; redirect the farm-wide list routes.
