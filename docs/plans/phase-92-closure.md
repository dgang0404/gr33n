# Phase 92 — closure (OC-92)

**Status:** **Shipped** on `main`.

**Canonical plan:** [`phase_92_zone_greenhouse_vocabulary.plan.md`](phase_92_zone_greenhouse_vocabulary.plan.md)

**Depends on:** [Phase 88](phase_88_domain_enums_api.plan.md) — zone/greenhouse vocabulary extends `GET /platform/domain-enums`.

---

## The one job (done)

> **Zone type** and **greenhouse profile enums** come from one API — wizard subset is a `wizard_visible` flag, not a separate hardcoded list.

---

## Workstream checklist

| WS | Deliverable | Verify |
|----|-------------|--------|
| **WS1** | Extend `GET /platform/domain-enums` | `internal/platform/domainenums/zone_vocabulary.go` |
| **WS2** | `Zones.vue` — zone type select from API | `adminZoneTypes`, `zoneTypeLabel` via `loadDomainEnums` |
| **WS3** | `zoneSetupWizard.js` — types from API | `wizardZoneTypes(getDomainEnums())`; deprecated re-exports |
| **WS4** | `ZoneGreenhouseTab.vue` — cover/policy from API | `greenhouseCoverTypes`, `greenhouseAutomationPolicies` |
| **WS5** | Guardian greenhouse labels | `domainenums.GreenhouseCoverTypeLabel` in `tools/greenhouse.go` |

---

## API extension (on domain-enums)

```json
{
  "zone_types": [
    { "value": "indoor", "label": "Indoor grow zone", "wizard_visible": true, "hint": "…" }
  ],
  "greenhouse_cover_types": [
    { "value": "film", "label": "Film / poly" }
  ],
  "greenhouse_automation_policies": [
    { "value": "auto", "label": "Auto (sensor rules)", "hint": "…" }
  ]
}
```

Eight zone types total (3 wizard-visible); three cover types including **`film`**; three automation policies.

---

## Operator impact

| Before | After |
|--------|-------|
| `Zones.vue` hardcoded 8 types | Same 8 from API with consistent labels |
| Wizard only 3 types in static JS | `wizard_visible` filter on API list |
| Duplicate greenhouse `<option>` lists | Single source in domain enums + fallback |

Legacy types (`veg`, `flower`, etc.) remain selectable in admin with `(legacy)` labels.

---

## Automated tests

| Test | Path |
|------|------|
| Domain-enums zone/GH contract | `cmd/api/smoke_phase92_test.go` |
| Payload shape (8 types, 3 covers, wizard count) | `internal/platform/domainenums/enums_test.go` |

---

## OC-92

Phase 92 is **closed** when smokes pass and zone create / greenhouse forms load vocabulary from **`GET /platform/domain-enums`**.

**Arc B note:** Phases **88–92** are now closed; **Phase 99** (CI domain parity guards) remains open in the arc.
