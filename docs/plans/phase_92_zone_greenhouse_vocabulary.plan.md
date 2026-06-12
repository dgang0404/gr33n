---
name: Phase 92 — Zone & greenhouse vocabulary API
overview: >
  Zone types and greenhouse cover/automation enums served from platform API;
  align Zones admin, zone setup wizard, and ZoneGreenhouseTab.
todos:
  - id: ws1-api
    content: "WS1: Extend domain-enums or GET /platform/zone-vocabulary — zone_types, gh_cover, gh_policy"
    status: pending
  - id: ws2-zones-admin
    content: "WS2: Zones.vue — zone type select from API (include legacy veg/flower with labels)"
    status: pending
  - id: ws3-wizard
    content: "WS3: zoneSetupWizard.js — remove ZONE_SETUP_TYPES, GREENHOUSE_* constants"
    status: pending
  - id: ws4-greenhouse-tab
    content: "WS4: ZoneGreenhouseTab.vue — cover/policy options from API"
    status: pending
  - id: ws5-guardian
    content: "WS5: summarize_zone_greenhouse_climate uses canonical enum labels"
    status: pending
isProject: false
---

# Phase 92 — Zone & greenhouse vocabulary

## Status

**Planned.** Aligns **zone create** paths and **greenhouse climate** forms.

**Depends on:** [Phase 88](phase_88_domain_enums_api.plan.md) (shared enums endpoint pattern).

**Closure:** **OC-92**

---

## The one job

> **Zone type** and **greenhouse profile enums** come from one API — wizard subset is a `wizard_visible` flag, not a separate hardcoded list.

---

## Gap today

| File | Values |
|------|--------|
| `Zones.vue` | `indoor`, `outdoor`, `greenhouse`, `nursery`, `seedling`, `veg`, `flower`, `storage` |
| `zoneSetupWizard.js` | Only `indoor`, `greenhouse`, `outdoor` |
| `zoneSetupWizard.js` | `GREENHOUSE_COVER_TYPES`, `GREENHOUSE_AUTOMATION_POLICIES` |
| `ZoneGreenhouseTab.vue` | Duplicate inline `<option>` for cover/policy |

OpenAPI already documents greenhouse cover type and automation policy enums.

---

## API extension

Add to `GET /platform/domain-enums` (Phase 88):

```json
{
  "zone_types": [
    { "value": "indoor", "label": "Indoor grow zone", "wizard_visible": true },
    { "value": "veg", "label": "Veg room (legacy)", "wizard_visible": false }
  ],
  "greenhouse_cover_types": [ … ],
  "greenhouse_automation_policies": [ … ]
}
```

---

## Guardian (WS5)

Greenhouse read tool output uses same labels as UI when describing cover type and automation policy on zone meta.

---

## Acceptance

- [ ] Wizard and Zones admin show consistent labels
- [ ] `film` cover available everywhere greenhouse meta is edited
- [ ] Legacy zone types still selectable in admin for existing farms

**Prompt loop:** **`phase 92`**.
