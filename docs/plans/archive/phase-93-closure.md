# Phase 93 — closure (OC-93)

**Status:** **Shipped** on `main`.

**Canonical plan:** [`phase_93_plant_identity_vocabulary_cleanup.plan.md`](phase_93_plant_identity_vocabulary_cleanup.plan.md)

**Depends on:** Phase 85 catalog-bound plants (`plants.crop_key`).

---

## The one job (done)

> **One plant identity system** — catalog `crop_key` + server `display_name`; cycle genetics/room notes use **`batch_label`** (not `strain_or_variety`); workspace tab is **`plants`**.

---

## Workstream checklist

| WS | Deliverable | Verify |
|----|-------------|--------|
| **WS1** | No client `display_name` as create identity | `plantcatalog.CreateFromRequest` — POST by `crop_key` only |
| **WS2** | `crop_cycles.batch_label` migration + alias | `20260620_phase93_crop_cycles_batch_label.sql` |
| **WS3** | `tab=strains` → `tab=plants` | `workspaces.js` `resolveWorkspaceTab`; `ZonesWorkspace.vue` redirect |
| **WS4** | OpenAPI `batch_label` primary; deprecated alias | `openapi.yaml` |
| **WS5** | Guardian/analytics use `batch_label` | `cropcycle/analytics.go`, `growHub.js` |
| **WS6** | Smokes + Vitest | `smoke_phase93_test.go`, `workspaces.test.js` |

---

## API semantics

| Field | Role |
|-------|------|
| `plants.crop_key` | Catalog identity (required on create) |
| `plants.display_name` | Server-set from catalog |
| `plants.variety_or_cultivar` | Optional genetics note |
| `crop_cycles.batch_label` | Optional batch/row label (primary) |
| `strain_or_variety` | **Deprecated write/read alias** for one release |

---

## Automated tests

| Test | Path |
|------|------|
| batch_label CRUD + alias | `cmd/api/smoke_phase93_test.go` |
| Legacy strains tab → plants | `ui/src/__tests__/workspaces.test.js` |

---

## OC-93

Phase 93 is **closed** when smokes pass and operators use **Plants** workspace tab + **batch label** on grows. Remove `strain_or_variety` alias in a future breaking release after integrators migrate.
