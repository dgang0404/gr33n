---
name: Phase 85 — Catalog-bound plants (gr33n Plants UX)
overview: >
  Plants are a first-class gr33n surface: Zone Plants and Plants workspace use a
  DB-backed catalog dropdown (all crop_library crops + EC/watering/light preview),
  plants.crop_key enforces one slot per crop per farm, Settings adjusts EC.
todos:
  - id: ws0-deps
    content: "WS0: migrate + picker API ≥46 crops with targets; no 404"
    status: pending
  - id: ws1-schema
    content: "WS1: plants.crop_key FK + UNIQUE(farm_id,crop_key) + backfill"
    status: pending
  - id: ws2-api
    content: "WS2: POST plants by crop_key; server display_name; upsert; reject unsupported"
    status: pending
  - id: ws3-ui
    content: "WS3: Zone Plants + Plants workspace — dropdown only, plant copy, target preview"
    status: pending
  - id: ws4-smokes
    content: "WS4: API smokes + Vitest; duplicate tomato → one row"
    status: pending
  - id: ws5-docs
    content: "WS5: operator-tour Plants tab; architecture § plants = knowledge base slot"
    status: pending
isProject: false
---

# Phase 85 — Catalog-bound plants

## Status

**Planned.** This is the **big Plants phase** for gr33n — everything operators touch when they say “what am I growing?”

**Depends on:** [Phase 84](phase_84_crop_catalog_enterprise_db.plan.md) (full catalog in DB), [Phase 83](phase_83_enterprise_agronomy_seed_pack.plan.md) (Settings → Crops & targets).

**Closure:** **OC-85**

---

## The one job

> **Zone → Plants → + Add plant** opens a catalog dropdown populated from Postgres — every crop in `crop_library.yaml` with EC, watering, and light targets visible at pick time. Operators never type a crop name. One farm row per `crop_key`.

---

## Operator UX (matches your screenshots)

### Zone → Plants tab (`/zones/{id}?tab=plants`)

```
┌─ Current grow strip ─────────────────────────────────────┐
│  Flower run (12/12) · Early Flower · day 19              │
└──────────────────────────────────────────────────────────┘

┌─ Plants in this zone ──────────────────── [ + Add plant ]┐
│  (list of catalog plants linked to this zone)             │
│  All farm plants →                                        │
└───────────────────────────────────────────────────────────┘
```

### New plant modal

| Field | Control | Rule |
|-------|---------|------|
| **Crop from knowledge base** | `<select>` via `CropLibraryPicker` | Required; calls `GET /farms/{id}/crop-library/picker` |
| **Target preview** | Read-only under dropdown | EC mS/cm, DLI, photoperiod by stage — same data Guardian uses |
| **Plant label** (optional) | Text | Room/batch nickname only — **not** crop identity |
| **Variety / cultivar** | Text optional | Genetics ("Blue Dream", "Cherokee Purple") |

**Copy rules:** Never “strain”. Use **plant**, **crop**, **knowledge base**.

**EC adjustment:** Link to **Settings → Crops & targets** — “tune EC for this crop on your farm” (Phase 83). That override applies to all grows of that `crop_key`, including Guardian answers.

### 404 fix (pre-req WS0)

If picker shows `Request failed with status code 404`:

```bash
make migrate
make check-crop-catalog-parity
# restart API — CROP_CATALOG_SOURCE=db (default)
```

---

## Problem today

```sql
plants (display_name TEXT NOT NULL, crop_profile_id BIGINT NULL)
```

Free text creates duplicate crops and breaks Guardian grounding. Screenshot symptom: typing `tom` in a text box instead of choosing **Tomato** from the seeded catalog.

---

## Target schema

```sql
ALTER TABLE gr33ncrops.plants
  ADD COLUMN crop_key TEXT REFERENCES gr33ncrops.crop_catalog_entries(crop_key);

CREATE UNIQUE INDEX idx_plants_farm_crop_key
  ON gr33ncrops.plants (farm_id, crop_key)
  WHERE deleted_at IS NULL AND crop_key IS NOT NULL;
```

- `display_name` → server-set from `crop_catalog_entries.display_name`
- `crop_profile_id` → denormalized effective profile for farm + `crop_key`
- Backfill: join existing `crop_profile_id` → `crop_profiles.crop_key`

---

## API contract

### POST `/farms/{id}/plants`

| Field | Required | Rule |
|-------|----------|------|
| `crop_key` | **Yes** | Catalog row; `supported=true` |
| `display_name` | No | Server sets from catalog |
| `variety_or_cultivar` | No | Genetics only |
| `crop_profile_id` | No | Server resolves effective profile |

**Upsert:** duplicate `crop_key` on farm → return existing plant (200).

**Reject:** unsupported (`ramps`, …) with catalog `unsupported_reason`.

### Data the dropdown needs (already Phase 84)

`GET /farms/{id}/crop-library/picker` returns grouped entries with:

- `crop_key`, `display_name`, `category`
- `substrate`, `watering_style` (catalog metadata)
- `crop_profile_id`, `has_targets`, per-stage EC/DLI/photoperiod lines

All ~46+ supported crops from `data/crop_library.yaml` must be in DB after migrate.

---

## UI surfaces (WS3)

| Surface | Change |
|---------|--------|
| `ZonePlantsSection.vue` | “Plants in this zone”; **+ Add plant**; `CropLibraryPicker` required |
| `Plants.vue` workspace | Same picker; list shows catalog name + variety |
| Terminology sweep | Remove “strain” from operator copy (`workspaces.js`, routes, empty states) |
| Target preview | Keep “Feeding & light targets (by stage)” block under dropdown |

---

## Guardian hook (prep for Phase 86/87)

Once `plants.crop_key` exists, Guardian can resolve:

`active cycle.plant_id` → `plants.crop_key` → effective `crop_profiles` (same as picker).

Phase 85 **does not** change `lookup_crop_targets` logic — it gives Guardian a reliable `crop_key` on every plant row. Phase 86 WS5 wires the cycle chain.

---

## Acceptance

- [ ] Picker loads ≥46 crops; no 404 after migrate + API restart
- [ ] Modal shows dropdown + EC/DLI preview (not free-text crop type)
- [ ] Two adds of `tomato` → one DB row
- [ ] `crop_key=ramps` → 400 with honest reason
- [ ] Settings EC override for `cannabis` visible in picker preview after save
- [ ] No “strain” in Zone Plants operator copy

---

## Out of scope

- Per-genetics EC profiles
- Operator-created catalog rows from UI
- Start grow / cycle binding (Phase 86)

**Prompt loop:** `phase 85 ws1` … `phase 85 ws5` or **`phase 85`**.
