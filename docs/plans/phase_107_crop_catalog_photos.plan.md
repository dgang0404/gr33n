---
name: Phase 107 — Crop catalog photos (picker UX)
overview: >
  Optional image_url per crop_catalog_entries; picker shows thumbnails for
  ornamentals (cacti, San Pedro, flowers) — after Phase 95 integrator cadence.
todos:
  - id: ws0-deps
    content: "WS0: Phase 95 integrator playbook — assets live in repo or CDN path convention"
    status: completed
  - id: ws1-schema
    content: "WS1: crop_catalog_entries.image_url + optional icon_key"
    status: completed
  - id: ws2-assets
    content: "WS2: ui/public/crops/ or S3 manifest; seed San Pedro, succulent, flower examples"
    status: completed
  - id: ws3-api
    content: "WS3: picker + commons API return image_url"
    status: completed
  - id: ws4-ui
    content: "WS4: CropLibraryPicker — thumbnail in select/option or grouped grid mode"
    status: completed
  - id: ws5-a11y
    content: "WS5: alt text from display_name; graceful fallback without image"
    status: completed
isProject: false
---

# Phase 107 — Crop catalog photos

## Status

**Shipped** on `main`. Closure: [`phase-107-closure.md`](phase-107-closure.md) (**OC-107**).

Improves **ornamental / specialty** picker UX (flowers, cacti, San Pedro).

**Depends on:** [Phase 95](phase_95_catalog_integrator_ops.plan.md), [Phase 84](phase_84_crop_catalog_enterprise_db.plan.md).

**Closure:** **OC-107**

---

## The one job

> **Pick San Pedro cactus** from a visual catalog row — thumbnail + name + target preview — not text-only dropdown.

---

## Asset convention (WS2)

```
data/crop-images/{crop_key}.webp   → copied or referenced in seed
crop_catalog_entries.image_url     → /assets/crops/{crop_key}.webp
```

Integrator doc (Phase 95) adds: new crop = YAML + optional image + seed SQL.

---

## UI (WS4)

- Default: small thumbnail left of label in picker
- Optional `grid` mode for Plants workspace "browse knowledge base"
- No image → existing text-only row (no broken icons)

---

## Acceptance

- [ ] ≥5 ornamentals show thumbnails after migrate
- [ ] Picker works with `image_url` null (backward compatible)
- [ ] Commons API includes `image_url`

**Prompt loop:** **`phase 107`**.
