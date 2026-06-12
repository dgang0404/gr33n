# Phase 107 — closure (OC-107)

**Status:** **Shipped** on `main`.

**Canonical plan:** [`phase_107_crop_catalog_photos.plan.md`](phase_107_crop_catalog_photos.plan.md)

**Depends on:** [Phase 84](phase_84_crop_catalog_enterprise_db.plan.md) catalog DB; [Phase 95](phase_95_catalog_integrator_ops.plan.md) integrator cadence for optional assets.

**Closes:** Visual picker UX for ornamentals and specialty crops — thumbnail + name, backward compatible when `image_url` is null.

---

## The one job (done)

> **Pick San Pedro cactus** from a catalog row with a thumbnail and display name — not a text-only dropdown. Crops without images stay text-only (no broken icons).

---

## Workstream checklist

| WS | Deliverable | Verify |
|----|-------------|--------|
| **WS0** | Integrator asset convention | [`catalog-integrator-playbook.md`](../catalog-integrator-playbook.md) § photos |
| **WS1** | `crop_catalog_entries.image_url` | `20260626_phase107_crop_catalog_photos.sql` |
| **WS2** | Demo SVG assets (5 ornamentals) | `ui/public/assets/crops/*.svg` |
| **WS3** | Picker + commons API return `image_url` | `picker.go`, `GET /commons/crop-catalog/{key}` |
| **WS4** | Thumbnail in picker list + selected row | `CropLibraryPicker.vue` |
| **WS5** | Alt text + null fallback | `cropImageAlt()`; placeholder block when no image |

---

## Seeded ornamentals (≥5)

`san_pedro`, `succulent`, `phalaenopsis`, `chrysanthemum`, `rose` → `/assets/crops/{crop_key}.svg`

Food crops (e.g. `tomato`) keep `image_url` null — picker omits thumbnail gracefully.

---

## Automated tests

| Test | Path |
|------|------|
| Commons + picker `image_url` contract | `cmd/api/smoke_phase107_test.go` |
| Picker renders thumbnail slot | `ui/src/__tests__/crop-catalog-photos.test.js` |

---

## OC-107

Phase 107 is **closed** when ≥5 ornamentals show thumbnails after migrate, picker and commons API expose `image_url`, and null URLs remain backward compatible.
