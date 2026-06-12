# Phase 104 — closure (OC-104)

**Status:** **Shipped** on `main`.

**Canonical plan:** [`phase_104_harvest_analytics_by_crop_key.plan.md`](phase_104_harvest_analytics_by_crop_key.plan.md)

**Depends on:** [Phase 86](phase_86_grow_ops_catalog_chain.plan.md) plant-linked cycles; [Phase 93](phase_93_plant_identity_vocabulary_cleanup.plan.md) vocabulary cleanup.

**Closes:** Compare grows and farm rollups bucket by catalog **`crop_key`**, not display name or legacy strain fields.

---

## The one job (done)

> **“Compare my last two cannabis runs”** resolves via **`crop_key=cannabis`** from `plant_id` — compare picker, analytics rollup, and Guardian read tools share the same bucket.

---

## Workstream checklist

| WS | Deliverable | Verify |
|----|-------------|--------|
| **WS1** | Compare/summary includes `crop_key` | `internal/handler/cropcycle/analytics.go` |
| **WS2** | Farm rollup API | `GET /farms/{id}/crop-analytics` |
| **WS3** | UI compare + Money | `CropCycleCompare.vue` crop filter + grouping; `farm.js` loaders |
| **WS4** | Guardian read tool | `summarize_farm_crops_by_key` in `readtools_crop.go` |
| **WS5** | Smokes | `cmd/api/smoke_phase104_test.go` |

---

## API contract

| Route | Behavior |
|-------|----------|
| `GET /farms/{id}/crop-cycles?crop_key=` | Filter cycles by plant catalog key |
| `GET /farms/{id}/crop-cycles/compare?ids=` | Each summary includes `crop_key`, `catalog_display_name`, `batch_label` |
| `GET /farms/{id}/crop-analytics` | Rollup: yield / cost / cycle count per `crop_key` |

Identity resolved via `cropcycle.ResolveCycleCropIdentity` (plant → catalog).

---

## Operator behavior

| Surface | Behavior |
|---------|----------|
| **Compare cycles** | Filter chips by crop; cycles grouped under catalog display name |
| **Money / grows** | Uses `loadCropAnalytics` for crop-key buckets |
| **Guardian** | “Compare last two tomato runs” → `summarize_farm_crops_by_key` |

---

## Automated tests

| Test | Path |
|------|------|
| Compare + analytics + filter | `cmd/api/smoke_phase104_test.go` |
| Store compare URL | `ui/src/__tests__/crop-cycle-analytics.test.js` |

---

## OC-104

Phase 104 is **closed** when compare and farm analytics group by `crop_key`, UI picker filters by catalog crop, and Guardian compare questions use the read tool.
