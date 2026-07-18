---
name: Phase 84 — Crop catalog enterprise DB
overview: >
  All crop_library.yaml crops + EC/light/watering targets + field guides in Postgres;
  picker and commons APIs for UI and Guardian; YAML authoring-only at runtime.
todos:
  - id: ws-b-schema
    content: "WS-B: crop_catalog_entries + aliases + agronomy_field_guides"
    status: completed
  - id: ws-c-seed
    content: "WS-C: seed SQL from YAML (~46 supported + unsupported + ornamentals path)"
    status: completed
  - id: ws-d-runtime
    content: "WS-D: LoadCatalogFromDB + CROP_CATALOG_SOURCE=db default"
    status: completed
  - id: ws-e-ingest
    content: "WS-E: Field guide ingest from DB"
    status: completed
  - id: ws-f-picker
    content: "WS-F: GET /farms/{id}/crop-library/picker — grouped dropdown + stage targets"
    status: completed
  - id: ws-g-cutover
    content: "WS-G: cutover runbook + production defaults"
    status: completed
  - id: ws-i-meta
    content: "WS-I: profile meta substrate/watering from catalog"
    status: completed
  - id: ws-j-commons
    content: "WS-J: GET /commons/crop-catalog*"
    status: completed
  - id: ws-k-parity
    content: "WS-K: check-crop-catalog-parity + check-crop-catalog-db"
    status: completed
  - id: ws-closure
    content: "OC-84: phase-84-closure.md"
    status: completed
isProject: false
---

# Phase 84 — Crop catalog enterprise DB

## Status

**Shipped on `main`.** Foundation for **Plants dropdown** and **Guardian crop grounding**.

**Closure:** **OC-84** (doc-only)

**Next:** [Phase 85](phase_85_catalog_bound_plants.plan.md) — bind `plants` to catalog; [Phase 86](phase_86_grow_ops_catalog_chain.plan.md) — Guardian cycle chain.

---

## The one job

> **Every crop in `data/crop_library.yaml` lives in Postgres with EC, pH, VPD, DLI, photoperiod, substrate, and watering metadata.** UI dropdown and Guardian read the same DB — not runtime YAML.

---

## What the Plants dropdown consumes (WS-F)

`GET /farms/{id}/crop-library/picker` returns:

- Grouped catalog entries (`category`: vegetables, herbs, cannabis, ornamentals, …)
- `crop_key`, `display_name`, `substrate`, `watering_style`
- `crop_profile_id`, `has_targets`
- Per-stage lines: EC mS/cm, DLI, photoperiod — shown in `CropLibraryPicker` under **Feeding & light targets**

**Pre-req for Zone → Plants:** migrate + API restart. 404 = route missing on old binary.

---

## Schema (shipped)

| Table | Purpose |
|-------|---------|
| `crop_catalog_entries` | Platform crops; `supported=true/false`; cousin_of for ornamentals |
| `crop_catalog_aliases` | Alias → crop_key (tom, roma → tomato) |
| `agronomy_field_guides` | Guide bodies for RAG |
| `crop_profiles` + stages | EC / light / watering targets |

Extend catalog (flowers, cacti, San Pedro): YAML → `./scripts/generate-crop-catalog-seed.sql.sh` → new migration.

---

## APIs (UI + Guardian integrators)

| Endpoint | Consumer |
|----------|----------|
| `GET /farms/{id}/crop-library/picker` | Zone Plants, Start grow, Plants workspace |
| `GET /commons/crop-catalog` | Guardian alias index; integrators |
| `GET /commons/crop-catalog/{crop_key}` | Detail + builtin profile id |
| Effective profile SQL | `lookup_crop_targets`, Settings overrides |

---

## Guardian note

Phase 84 seeds the **data** Guardian needs. Phase 86 WS5 + Phase 87 WS4 verify Guardian uses **DB catalog registry** and **same effective profiles** as the picker — not LLM-invented EC.

---

## Related

| Doc | Use |
|-----|-----|
| [phase_84_87_crop_identity_roadmap.plan.md](phase_84_87_crop_identity_roadmap.plan.md) | Full arc |
| [crop-catalog-db-cutover-runbook.md](../crop-catalog-db-cutover-runbook.md) | Operator migrate |
