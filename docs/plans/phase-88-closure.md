# Phase 88 — closure (OC-88)

**Status:** **Shipped** on `main`.

**Canonical plan:** [`phase_88_domain_enums_api.plan.md`](phase_88_domain_enums_api.plan.md)

**Depends on:** Postgres enum alignment with `croplibrary.ValidGrowthStages`.

**Follow-on:** Phase 99 CI parity guards — [`phase-99-closure.md`](phase-99-closure.md) · `make check-ui-domain-parity`; Phase 100 offline cache for domain enums.

---

## The one job (done)

> **One HTTP call** returns every platform enum the UI needs for dropdowns — growth stages (11 values including `transition` and `flush`), reservoir status, cost categories, NF inventory enums, zone/greenhouse vocabulary.

---

## Workstream checklist

| WS | Deliverable | Verify |
|----|-------------|--------|
| **WS1** | `GET /platform/domain-enums` | `internal/platform/domainenums/enums.go`, `internal/handler/platform/handler.go` |
| **WS2** | OpenAPI + contract smokes | `openapi.yaml` · `cmd/api/smoke_phase88_test.go` |
| **WS3** | `ui/lib/domainEnums.js` — fetch once, module cache | `loadDomainEnums`, `getDomainEnums` |
| **WS4** | UI drops duplicate arrays | `Setpoints.vue`, `Fertigation.vue`, `Inventory.vue`, `MoneyHub.vue`, `growHub.js`, `moneyHub.js` |
| **WS5** | Guardian / setpoint stage parity | `transition` persists — `TestPhase88_SetpointTransitionStagePersists` |
| **WS6** | Fallback snapshot for offline | `domainEnums.fallback.js` (Phase 100 cache path) |

---

## API shape

```
GET /platform/domain-enums
```

Returns `growth_stages`, `reservoir_statuses`, `cost_categories`, `application_targets`, `input_definition_categories`, `batch_statuses`, plus zone/greenhouse enums added in Phase 92 (same payload).

Labels are humanized snake_case; backend order matches Postgres enums.

---

## Operator impact

| Before | After |
|--------|-------|
| `SetpointRow` default had 9 stages (missing `transition`, `flush`) | All 11 stages from API via `Setpoints.vue` → `loadDomainEnums` |
| `Fertigation.vue` inline stage array | `growthStages` / `reservoirStatuses` from domain enums |
| `moneyHub.js` partial cost list | Full `cost_categories` from API |
| `Inventory.vue` hardcoded NF enums | `application_targets`, `input_definition_categories`, `batch_statuses` from API |

Bundled fallback in `domainEnums.fallback.js` keeps forms usable when API is briefly unavailable (Phase 100 extends this with IndexedDB).

---

## Automated tests

| Test | Path |
|------|------|
| Domain enums contract (11 stages, costs, reservoirs) | `cmd/api/smoke_phase88_test.go` |
| `transition` setpoint round-trip | `cmd/api/smoke_phase88_test.go` |
| Loader cache + growth stage values | `ui/src/__tests__/domain-enums.test.js` |

---

## OC-88

Phase 88 is **closed** when smokes pass and every stage/cost/reservoir dropdown loads from **`GET /platform/domain-enums`** (or bundled fallback). Phase **99** adds CI guards so UI fallbacks cannot drift from Go/OpenAPI again — see [`phase-99-closure.md`](phase-99-closure.md).
