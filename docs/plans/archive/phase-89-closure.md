# Phase 89 — closure (OC-89)

**Status:** **Shipped** on `main`.

**Canonical plan:** [`phase_89_lighting_presets_api_wiring.plan.md`](phase_89_lighting_presets_api_wiring.plan.md)

**Depends on:** Phase 35 lighting programs API (`GET /lighting-programs/presets`, `POST …/from-preset`).

---

## The one job (done)

> **Lighting preset chips and zone wizard options come from `GET /lighting-programs/presets`** — same list Guardian `create_lighting_program` uses via `lightinghandler.PresetList()`.

---

## Workstream checklist

| WS | Deliverable | Verify |
|----|-------------|--------|
| **WS1** | `ui/lib/lightingPresets.js` — fetch + session cache | `loadLightingPresets`, `mapApiPreset` |
| **WS2** | `LightingPrograms.vue` — no `PRESET_CHIPS` | loads via `loadLightingPresets(api)` |
| **WS3** | `ZoneLightingEditor.vue` — presets from loader | `loadLightingPresets(api)` on mount |
| **WS4** | `zoneSetupWizard.js` — API presets passed in; includes `peas_22_2` | `ZoneSetupWizard.vue` · `LIGHTING_PRESET_SKIP` UI-only |
| **WS5** | Vitest + Phase 35 smoke parity | `lighting-presets.test.js` · `smoke_phase35_lighting_test.go` |

---

## Operator impact

| Before | After |
|--------|-------|
| Three hardcoded preset arrays in Vue | Single loader from API |
| Zone wizard had 4 presets (missing `peas_22_2`) | All backend presets including `peas_22_2` |
| New Go preset required UI release | New preset in `PresetList()` appears after API deploy |

**UI-only exception:** `{ key: '', label: 'Skip for now' }` (`LIGHTING_PRESET_SKIP`) — not an API preset.

---

## Guardian alignment

No backend change. Guardian and UI share the same preset keys from `internal/handler/lighting/handler.go`.

---

## Automated tests

| Test | Path |
|------|------|
| Loader map + cache | `ui/src/__tests__/lighting-presets.test.js` |
| Zone wizard preset request | `ui/src/__tests__/zone-setup-wizard.test.js` |
| Presets contract (≥4, `veg_18_6`) | `cmd/api/smoke_phase35_lighting_test.go` |
| Preset apply + schedule | `cmd/api/smoke_phase35_lighting_test.go` |

---

## OC-89

Phase 89 is **closed** when smokes pass and no hardcoded preset keys remain in production Vue (except **Skip for now**).
