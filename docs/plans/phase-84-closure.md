# Phase 84 — closure (OC-84)

**Status:** **Shipped** on `main`.

**Canonical plan:** [`phase_84_crop_catalog_enterprise_db.plan.md`](phase_84_crop_catalog_enterprise_db.plan.md)

**Depends on:** Phase 82 crop library YAML + field guides; Phase 83 enterprise seed pack (optional agronomy import).

---

## The one job (done)

> Every supported crop in `data/crop_library.yaml` lives in Postgres with EC, pH, VPD, DLI, photoperiod, substrate, and watering metadata. The Plants picker, Settings targets, and Guardian read the same DB — not runtime YAML.

---

## Workstream checklist

| WS | Deliverable | Verify |
|----|-------------|--------|
| **WS-B** | `crop_catalog_entries`, aliases, `agronomy_field_guides` schema | `make migrate` |
| **WS-C** | Seed SQL from YAML (~46 supported + unsupported) | `db/migrations/20260616_phase84_crop_catalog_seed.sql` |
| **WS-D** | `LoadCatalogFromDB` + `CROP_CATALOG_SOURCE=db` default | `internal/croplibrary/catalog_db.go` |
| **WS-E** | Field guide ingest from DB | Guardian RAG ingest scripts |
| **WS-F** | `GET /farms/{id}/crop-library/picker` | `CropLibraryPicker.vue` |
| **WS-G** | Cutover runbook + production defaults | [`crop-catalog-db-cutover-runbook.md`](../crop-catalog-db-cutover-runbook.md) |
| **WS-I** | Profile `meta` substrate/watering from catalog | Phase 84 meta migration |
| **WS-J** | `GET /commons/crop-catalog*` | `smoke_commons_crop_catalog_test.go` |
| **WS-K** | `make check-crop-catalog-parity` + DB check | CI / operator scripts |

---

## Operator quick start

```bash
make migrate
make check-crop-catalog-parity

# Picker + targets (farm member JWT):
curl -H "Authorization: Bearer $TOKEN" "$API/farms/1/crop-library/picker"

# Commons integrator index:
curl "$API/commons/crop-catalog"
```

**Cutover:** [`crop-catalog-db-cutover-runbook.md`](../crop-catalog-db-cutover-runbook.md)

**Farm overrides (Phase 83/105):** Settings → **Crops & targets** or `PUT /farms/{id}/crop-profiles/{crop_key}` — override changes are auditable (Phase 105).

---

## Automated tests

| Test | Path |
|------|------|
| Commons crop catalog | `cmd/api/smoke_commons_crop_catalog_test.go` |
| Picker + profile parity | `cmd/api/smoke_phase87_test.go` |
| Catalog-bound plants | `cmd/api/smoke_phase85_test.go` |
| Parity scripts | `make check-crop-catalog-parity` |

---

## Documentation index

| Doc | Topic |
|-----|--------|
| [`crop-catalog-db-cutover-runbook.md`](../crop-catalog-db-cutover-runbook.md) | DB cutover |
| [`crop-knowledge-operator-runbook.md`](../crop-knowledge-operator-runbook.md) | Operator + integrator hub |
| [`plans/phase_84_87_crop_identity_roadmap.plan.md`](phase_84_87_crop_identity_roadmap.plan.md) | Phases 84–87 arc |
| [`plans/phase-87-closure.md`](phase-87-closure.md) | Full crop knowledge closure |

---

## OC-84

Phase 84 is **closed** when migrate + parity check pass and the picker/commons APIs return DB-backed catalog rows with stage targets. Runtime YAML is authoring-only; production defaults to Postgres (`CROP_CATALOG_SOURCE=db`).
