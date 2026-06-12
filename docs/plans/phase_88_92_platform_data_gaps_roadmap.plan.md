---
name: Phases 88–92 — Platform data gaps (UI static → DB/API)
overview: >
  Audit of UI hardcoded domain constants that should come from Postgres/API so
  operators, zone cockpit, and Guardian stay aligned. Phases 84–87 cover plants/crops;
  88–92 cover enums, lighting, devices, bootstrap, zones.
todos:
  - id: p88-enums
    content: "Phase 88 — GET /platform/domain-enums; UI drops duplicate stage/reservoir/cost lists"
    status: completed
  - id: p89-lighting
    content: "Phase 89 — Wire GET /lighting-programs/presets in UI (API exists, unused)"
    status: pending
  - id: p90-devices
    content: "Phase 90 — Device taxonomy registry (sensor/actuator → water/light/climate)"
    status: pending
  - id: p91-bootstrap
    content: "Phase 91 — Bootstrap template catalog API (replace bootstrapTemplates.js)"
    status: pending
  - id: p92-zones
    content: "Phase 92 — Zone types + greenhouse enums from API"
    status: pending
isProject: false
---

# Phases 88–92 — Platform data gaps (UI static → DB/API)

## Why this arc exists

Phases **84–87** move **plants & crop knowledge** into Postgres. The UI still carries **dozens of parallel copies** of domain vocabulary — growth stages, lighting presets, sensor types, bootstrap keys — that **already exist** in Postgres enums or backend handlers but are **not fetched at runtime**.

When UI and DB drift:

- Operators pick stages setpoints cannot store (`SetpointRow` missing `transition` / `flush`)
- New lighting preset on API never appears in zone wizard
- Custom Pi sensor types land in wrong Water/Light/Climate tab
- Guardian tools use backend maps; UI uses different lists → wrong advice context

**Rule:** Domain lists operators depend on → **API or DB**. UI may cache; it must not be source of truth.

**Not in scope here:** Product IA (`workspaces.js`, nav tabs), farmer vocabulary lint, icon maps, cron helpers — those stay in frontend.

---

## Full gap audit (UI scan)

| # | Category | Hardcoded where | DB/API today | Severity | Phase |
|---|----------|-----------------|--------------|----------|-------|
| 1 | **Growth stages** | `growHub.js`, `Fertigation.vue` (duplicate), `SetpointRow.vue` (**9/11**, missing transition/flush) | Postgres `growth_stage_enum`; OpenAPI `GrowthStageEnum`; `croplibrary.ValidGrowthStages` | **High** | **88** |
| 2 | **Lighting presets** | `LightingPrograms.vue`, `ZoneLightingEditor.vue`, `zoneSetupWizard.js` (missing `peas_22_2`) | **`GET /lighting-programs/presets`** — UI never calls it | **High** | **89** |
| 3 | **Sensor → plant need** | `plantNeeds.js`, `sensorTypeLabel.js`, `ZoneGreenhouseTab.vue` | `sensor_type` free-text; no registry | **High** | **90** |
| 4 | **Actuator → plant need** | `plantNeeds.js`, `deviceSetupWizard.js`, greenhouse GH types | `actuator_type` free-text; no registry | **High** | **90** |
| 5 | **Bootstrap templates** | `constants/bootstrapTemplates.js` (+ summaries in 3 views) | DB `apply_farm_bootstrap_template`; **no list API** | **Medium** | **91** |
| 6 | **Zone types** | `Zones.vue` (8 values), `zoneSetupWizard.js` (3 values) | `zone_type` free-text on zones | **Medium** | **92** |
| 7 | **Greenhouse enums** | `zoneSetupWizard.js`, `ZoneGreenhouseTab.vue` (duplicate cover/policy lists) | OpenAPI enums on greenhouse meta | **Medium** | **92** |
| 8 | **Reservoir status** | `Fertigation.vue` select, `feedingAdminHub.js` labels | `ReservoirStatusEnum` in OpenAPI | **Medium** | **88** |
| 9 | **Cost categories** | `moneyHub.js` — 6 spend + all income → `miscellaneous` | Full `CostCategoryEnum` | **Medium** | **88** |
| 10 | **Inventory NF enums** | `Inventory.vue` — categories, batch status, application targets | OpenAPI enums | **Low** | **88** (bundle) |
| 11 | **Pi wiring sources** | `hardwareWiring.js` `SENSOR_WIRING_SOURCES` | config JSON only | **Low** | **90** (optional) |
| 12 | **Task Kanban columns** | `Tasks.vue` `COLUMNS` | `TaskStatusEnum` | **Low** | defer |
| 13 | **Crop category order** | `cropLibraryPicker.js` `CATEGORY_ORDER` (dead code) | API picker `categoryOrder` | **Low** | delete in 85/87 |
| 14 | **Workspaces / nav** | `workspaces.js`, `navGroups.js` | Product IA | **N/A** | — |

---

## Duplication map (same truth, many files)

```
growth_stage_enum
├── Postgres / OpenAPI GrowthStageEnum
├── internal/croplibrary/catalog.go ValidGrowthStages
├── ui/lib/growHub.js GROWTH_STAGES
├── ui/views/Fertigation.vue (inline duplicate)
└── ui/components/SetpointRow.vue (INCOMPLETE — bug)

lighting presets
├── internal/handler/lighting/handler.go presets map
├── GET /lighting-programs/presets (Guardian create_lighting_program uses PresetList())
├── ui/views/LightingPrograms.vue PRESET_CHIPS
├── ui/components/ZoneLightingEditor.vue
└── ui/lib/zoneSetupWizard.js (subset, no peas_22_2)

sensor/actuator taxonomy
├── ui/lib/plantNeeds.js (water/light/air sets)
├── ui/lib/sensorTypeLabel.js
└── Guardian read tools (implicit via zone snapshot — no shared registry)
```

---

## Phase map

| Phase | One job | Depends on |
|-------|---------|------------|
| **[88](phase_88_domain_enums_api.plan.md)** | Single **domain enums API**; UI imports one loader | — |
| **[89](phase_89_lighting_presets_api_wiring.plan.md)** | UI fetches presets API (quick win) | — |
| **[90](phase_90_device_taxonomy_registry.plan.md)** | DB registry for sensor/actuator roles + Guardian | 88 optional |
| **[91](phase_91_bootstrap_template_catalog.plan.md)** | List bootstrap templates from DB/commons | — |
| **[92](phase_92_zone_greenhouse_vocabulary.plan.md)** | Zone types + GH cover/policy from API | 88 |

**Continues:** [Phases 84–87](phase_84_87_crop_identity_roadmap.plan.md) (plants & crop knowledge).

---

## Guardian impact (cross-cutting)

| Gap | Guardian risk |
|-----|----------------|
| Growth stages | `lookup_crop_targets` stage vs setpoint stage mismatch |
| Lighting presets | `create_lighting_program` preset keys ≠ UI wizard keys |
| Device taxonomy | Wrong zone context in `summarize_zone_*` enrichment |
| Bootstrap templates | Grow setup pack proposals reference unknown template keys |

Phase **90** should expose taxonomy to Guardian read tools the same way crop catalog feeds `lookup_crop_targets`.

---

## Quick wins (no new phase)

| Fix | Effort |
|-----|--------|
| `SetpointRow` import `GROWTH_STAGES` from `growHub.js` | 1 line default prop |
| Remove `Fertigation.vue` inline stage array | import shared |
| Delete dead `CATEGORY_ORDER` in `cropLibraryPicker.js` | cleanup |
| Wire lighting presets (Phase 89) | API already exists |

---

## Prompt loop

`phase 88 ws1`, … or `phase 88` for full phase. Same for 89–92.

**Suggested order:** **89** (fast) → **88** (foundation) → **90** (cockpit) → **91** → **92** → **99**.

---

## Continuation — blind spots (Phases 93–100)

**Master roadmap:** [phase_84_100_master_roadmap.plan.md](phase_84_100_master_roadmap.plan.md)

| Phase | Blind spot |
|-------|------------|
| **93** | Identity vocabulary, `batch_label`, `tab=plants` |
| **94** | Genetics / batch EC |
| **95** | Catalog integrator cadence |
| **96** | Feeding program vs stage |
| **97** | RAG vs structured truth |
| **98** | Enterprise promotion model |
| **99** | CI domain parity guards |
| **100** | Offline catalog cache |

---

## Out of scope (future phases OK)

| Topic | Notes |
|-------|--------|
| Per-farm custom sensor types | Registry is platform-wide; farm picks from list |
| Full OpenAPI codegen for UI | Optional; hand-rolled enum endpoint is enough for v1 |
| Workspace/nav in DB | Product shell stays in code |
