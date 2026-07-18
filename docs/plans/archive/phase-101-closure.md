# Phase 101 — closure (OC-101)

**Status:** **Shipped** on `main`.

**Canonical plan:** [`phase_101_guardian_write_tools_crop_key.plan.md`](phase_101_guardian_write_tools_crop_key.plan.md)

**Depends on:** [Phase 85](phase_85_catalog_bound_plants.plan.md) catalog-bound plants.

**Closes:** Guardian write path aligned with UI — no free-text plant identity via chat.

---

## The one job (done)

> **`create_plant`** and grow setup pack plant sections require **`crop_key`** from the knowledge base — server sets `display_name`; unsupported keys and client `display_name` are rejected.

---

## Workstream checklist

| WS | Deliverable | Verify |
|----|-------------|--------|
| **WS1** | `execCreatePlant` → `plantcatalog.CreateOrGet` | `internal/farmguardian/tools/plants.go` |
| **WS2** | `apply_grow_setup_pack` requires `plant.crop_key` | `grow_setup_pack.go` |
| **WS3** | Impact summary cites `crop_key` | `impact.go`, `proposals_revise_test.go` |
| **WS4** | OpenAPI `GuardianCreatePlantArgs` | `openapi.yaml` — `required: [crop_key]` |
| **WS5** | Smokes + unit tests | `smoke_phase101_test.go`, `grow_create_test.go` |

---

## Tool contract

| Arg | Rule |
|-----|------|
| `crop_key` | **Required** (or legacy `crop_profile_id`); catalog supported |
| `display_name` | **Rejected** — server sets from catalog |
| `variety_or_cultivar` | Optional genetics / batch note |

Same upsert semantics as `POST /farms/{id}/plants` (one `crop_key` per farm).

---

## Persona / grounding

`CropTargetsGroundingRule` in `readtools_crop.go` states plant writes require catalog `crop_key`.

---

## Automated tests

| Test | Path |
|------|------|
| Confirm without crop_key → 400 | `cmd/api/smoke_phase101_test.go` |
| Upsert + ramps block + display_name reject | same |
| Args validation | `internal/farmguardian/tools/grow_create_test.go` |

---

## OC-101

Phase 101 is **closed** when Guardian cannot create typo plant rows — write tools mirror Phase 85 plant API rules.
