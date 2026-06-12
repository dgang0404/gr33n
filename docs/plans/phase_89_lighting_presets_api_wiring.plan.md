---
name: Phase 89 — Lighting presets API wiring
overview: >
  UI fetches GET /lighting-programs/presets instead of three hardcoded preset arrays;
  zone wizard, LightingPrograms, ZoneLightingEditor stay aligned with Guardian.
todos:
  - id: ws1-loader
    content: "WS1: ui/lib/lightingPresets.js — fetch + cache GET /lighting-programs/presets"
    status: pending
  - id: ws2-lighting-programs
    content: "WS2: LightingPrograms.vue — remove PRESET_CHIPS"
    status: pending
  - id: ws3-zone-editor
    content: "WS3: ZoneLightingEditor.vue — presets from loader"
    status: pending
  - id: ws4-zone-wizard
    content: "WS4: zoneSetupWizard.js — remove ZONE_LIGHTING_PRESETS; include peas_22_2"
    status: pending
  - id: ws5-tests
    content: "WS5: Vitest + smoke_phase35 parity"
    status: pending
isProject: false
---

# Phase 89 — Lighting presets API wiring

## Status

**Planned.** **Quick win** — backend API **already shipped**; UI never calls it.

**Closure:** **OC-89**

---

## The one job

> **Lighting preset chips and zone wizard options come from `GET /lighting-programs/presets`** — same list Guardian `create_lighting_program` uses via `lightinghandler.PresetList()`.

---

## Gap today

| Location | Issue |
|----------|--------|
| `ui/views/LightingPrograms.vue` | `PRESET_CHIPS` hardcoded |
| `ui/components/ZoneLightingEditor.vue` | inline `presets` array |
| `ui/lib/zoneSetupWizard.js` | `ZONE_LIGHTING_PRESETS` — **4 presets, missing `peas_22_2`** |

Backend (`internal/handler/lighting/handler.go`):

```29:34:internal/handler/lighting/handler.go
var presets = map[string]presetDef{
	"peas_22_2":      {Name: "Peas 22/2 (Long-day veg)", OnHours: 22, OffHours: 2},
	"veg_18_6":       {Name: "Veg 18/6 (Vegetative)", OnHours: 18, OffHours: 6},
	"flower_12_12":   {Name: "Flower 12/12 (Flowering)", OnHours: 12, OffHours: 12},
	"seedling_16_8":  {Name: "Seedling 16/8", OnHours: 16, OffHours: 8},
}
```

Route: `GET /lighting-programs/presets` (authenticated).

---

## WS1 — Loader

```javascript
// ui/lib/lightingPresets.js
export async function loadLightingPresets(api) {
  const { data } = await api.get('/lighting-programs/presets')
  return data // [{ preset_key, name, on_hours, off_hours }, …]
}
```

Cache in memory for session; optional pinia store.

Zone wizard keeps synthetic `{ key: '', label: 'Skip for now' }` as UI-only first option.

---

## Guardian alignment

No backend change. Guardian already uses `PresetList()`. After Phase 89, operator picks same keys in UI as Guardian proposes.

**Phase 86 tie-in:** Light tab photoperiod from crop profile **and** active lighting program preset should both display — preset list consistency helps.

---

## Acceptance

- [ ] Zero hardcoded preset keys in Vue except "Skip for now"
- [ ] Zone wizard shows `peas_22_2`
- [ ] New preset added to Go map appears in UI after API deploy (no UI release)
- [ ] Existing `smoke_phase35_lighting` still green

**Prompt loop:** **`phase 89`** (small — often one session).
