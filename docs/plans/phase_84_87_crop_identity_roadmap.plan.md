---
name: Phases 84–87 — Crop identity & knowledge base (roadmap)
overview: >
  Plants are a first-class gr33n surface: catalog-backed dropdowns (not free text),
  EC/light/watering from Postgres, farm EC tweaks in Settings, Guardian grounded
  on the same crop APIs as the UI. Phase 84 shipped; 85–87 finish plants + grows + closure.
todos:
  - id: p84-closure
    content: "Phase 84 — OC-84 closure doc"
    status: pending
  - id: p85-plants
    content: "Phase 85 — plants.crop_key + Zone Plants dropdown + catalog-only create"
    status: pending
  - id: p86-grows
    content: "Phase 86 — grow chain + zone EC strip + Guardian same crop_key path"
    status: pending
  - id: p87-guardian
    content: "Phase 87 — Guardian crop API hardening + operator runbook + OC-87"
    status: pending
isProject: false
---

# Phases 84–87 — Plants & crop knowledge base

## Why this arc exists

**Plants are core to gr33n.** Operators live in **My zones → Plants**, the **Plants** workspace, and **Start grow**. Every grow decision — EC, watering style, photoperiod, feeding program — should flow from one **farm knowledge base** in Postgres, not from typed labels or the LLM’s general training.

Your screenshots show the target UX:

| Today (wrong) | Target (this arc) |
|---------------|-------------------|
| “+ Add **strain**” | “+ Add **plant**” |
| Free-text “Crop type” (`tom` → **404**) | **Dropdown** from `GET /farms/{id}/crop-library/picker` (all `crop_library.yaml` crops seeded in DB) |
| No targets visible at pick time | **Feeding & light targets** preview (EC mS/cm, DLI, photoperiod by stage) |
| EC tuned nowhere obvious | **Settings → Crops & targets** — per-farm EC override by `crop_key` (Phase 83) |
| Guardian guesses EC from Llama weights | Guardian **`lookup_crop_targets`** reads **same DB profiles** as the picker — never invents numbers |

**404 on picker:** UI is correct; API needs `make migrate` + restart so `/farms/{id}/crop-library/picker` is registered and `CROP_CATALOG_SOURCE=db`.

---

## The one pipeline (UI + Guardian + feeding)

```
data/crop_library.yaml  ──generate──►  Postgres seed migrations
                                              │
                    ┌─────────────────────────┼─────────────────────────┐
                    ▼                         ▼                         ▼
         crop_catalog_entries      crop_profiles + stages      agronomy_field_guides
         (list, substrate,           (EC, pH, VPD, DLI,           (RAG narrative)
          watering, supported)        photoperiod per stage)
                    │                         │
                    │    farm override ◄──────┘  Settings → Crops & targets
                    ▼
              plants.crop_key     UNIQUE (farm_id, crop_key) — one catalog slot per crop
                    │
                    ▼
              crop_cycles.plant_id  →  zone grow strip  →  Water / Light tabs
                    │
                    └──────────────────►  Guardian lookup_crop_targets (same crop_key + stage)
```

**Operator rule:** Pick the **plant type** from the knowledge base. Optional **variety / cultivar** is genetics or batch label only — not a new catalog row. Extend the platform catalog (flowers, cacti, San Pedro, …) via **seed migration**, not UI typing.

---

## Architectural decisions (locked)

| Decision | Verdict |
|----------|---------|
| Runtime YAML | **No** — `CROP_CATALOG_SOURCE=db` |
| Separate unsupported table | **No** — `crop_catalog_entries.supported=false` |
| Free-text crop identity on `plants` | **No** — catalog dropdown only (Phase 85) |
| Farm EC tweak | **Settings → Crops & targets** by `crop_key` (Phase 83, shipped) |
| Guardian EC numbers | **Only** from `lookup_crop_targets` / DB profiles — persona hard rule |
| Guardian alias / unsupported | **Same DB catalog** as picker — not stale YAML at runtime |
| New crops (ornamentals, etc.) | Platform seed + `catalog_version` bump |

---

## Phase map

| Phase | Status | One job |
|-------|--------|---------|
| **[84](phase_84_crop_catalog_enterprise_db.plan.md)** | **Shipped** | Full catalog + field guides + targets in Postgres; commons + picker APIs |
| **[85](phase_85_catalog_bound_plants.plan.md)** | Planned | **Plants** = catalog slot; dropdown-only create; kill typo flooding |
| **[86](phase_86_grow_ops_catalog_chain.plan.md)** | Planned | Start grow + zone strip + Water/Light wired to `crop_key` chain |
| **[87](phase_87_crop_knowledge_operator_closure.plan.md)** | Planned | **Guardian crop API parity** + operator runbook + smokes + OC-87 |

**Depends on:** [Phase 82](phase_82_guardian_crop_grounding_hardening.plan.md), [Phase 83](phase_83_enterprise_agronomy_seed_pack.plan.md) (farm EC overrides UI).

---

## Guardian must use the crop APIs (cross-cutting)

Guardian must **not** answer EC, pH, VPD, DLI, photoperiod, watering style, or “is this crop supported?” from general LLM knowledge. It must use the **same sources as the UI**:

| Source | API / path | Used by |
|--------|------------|---------|
| Catalog list + metadata | `GET /commons/crop-catalog`, `GET /commons/crop-catalog/{crop_key}` | Integrators; Guardian alias/unsupported resolution |
| Farm picker (grouped + targets preview) | `GET /farms/{id}/crop-library/picker` | Zone Plants, Start grow, Plants workspace |
| Effective targets (incl. farm override) | `crop_profiles` + stages via sqlc; `PUT /farms/{id}/crop-profiles/{crop_key}` | UI strip, Water tab, **`lookup_crop_targets`** |
| Active grow context | `plants.crop_key` on cycle’s `plant_id` | Zone strip + Guardian when user asks about “this grow” |
| Narrative depth | `agronomy_field_guides` → RAG | Guardian chat (bootstrap ingest) |

**Read tool contract (Phase 82 → hardened in 86/87):**

- **`lookup_crop_targets`** — fires on feed/light/water/compare intent; returns mS/cm EC from **effective farm profile**; honest block for unsupported catalog rows.
- **Registry** — `CROP_CATALOG_SOURCE=db` + `SetRuntimeCatalogQuerier` at API boot so aliases match Postgres, not a stale checkout.
- **Acceptance:** Guardian answer for “EC in early flower” **matches** zone strip chip after farm override; compare cannabis vs tomato uses DB stages, not invented percentages.

Implementation spans **Phase 86 WS5** (cycle → plant → crop_key) and **Phase 87 WS3–WS4** (architecture doc + Guardian smokes).

---

## Operator surfaces (plants-first)

| Surface | Phase | Behavior |
|---------|-------|----------|
| **Zone → Plants** tab | 85–86 | “Plants in this zone”; **+ Add plant**; catalog dropdown + target preview |
| **Plants** workspace | 85 | All farm catalog slots; link to Settings for EC |
| **Start grow** wizard | 86 | Requires catalog plant; optional variety; feeding program follows profile |
| **Settings → Crops & targets** | 83 ✅ | Adjust EC per `crop_key` for this farm |
| **Farm Guardian** | 86–87 | Grounded on same profiles; cites field guides when bootstrapped |

---

## Manual prompt loop

| Prompt | Scope |
|--------|--------|
| `phase 85 ws1` | `plants.crop_key` migration + backfill |
| `phase 85 ws3` | Zone Plants dropdown + “plant” copy (your screenshot flow) |
| `phase 86 ws4` | Zone EC strip + Water/Light from profile |
| `phase 86 ws5` | Guardian resolves active cycle → `plants.crop_key` |
| `phase 87 ws3` | Architecture §7.0af Guardian + crop APIs |
| `phase 87 ws4` | Guardian smokes: EC matches UI; ramps blocked |

Or **`phase 85`**, **`phase 86`**, **`phase 87`** for full phases.

---

## Related docs

| Doc | Use |
|-----|-----|
| [crop-catalog-db-cutover-runbook.md](../crop-catalog-db-cutover-runbook.md) | Migrate + parity |
| [phase-83-closure.md](phase-83-closure.md) | Farm EC override UI |
| [phase-14-operator-documentation.md](../phase-14-operator-documentation.md) | Index |

---

## Continuation — platform data gaps (Phases 88–92)

After plants/crops (85–87), the UI still hardcodes domain enums that **already exist** in Postgres/OpenAPI. See **[phase_88_92_platform_data_gaps_roadmap.plan.md](phase_88_92_platform_data_gaps_roadmap.plan.md)**.

| Phase | Focus |
|-------|--------|
| **88** | Domain enums API (growth stages, reservoir, cost categories) |
| **89** | Lighting presets — wire existing `GET /lighting-programs/presets` |
| **90** | Device taxonomy (sensor/actuator → water/light/climate) + Guardian |
| **91** | Bootstrap template catalog |
| **92** | Zone types + greenhouse enums |

---

## Out of scope (future phases OK)

| Topic | Notes |
|-------|--------|
| Per-genetics EC (OG Kush vs Wedding Cake) | Separate genetics profile phase |
| Operator UI to add catalog rows | Integrator seed migration only |
| More ornamentals (flowers, cacti, San Pedro) | Add to YAML → regenerate seed SQL → migrate |
