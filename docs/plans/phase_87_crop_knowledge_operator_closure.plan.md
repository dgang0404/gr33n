---
name: Phase 87 — Crop knowledge & Guardian closure
overview: >
  Operator runbook for Plants + knowledge base; Guardian crop API parity verified;
  smokes prove UI and Guardian share EC from Postgres; OC-87 closes 84–87 arc.
todos:
  - id: ws1-runbook
    content: "WS1: crop-knowledge-operator-runbook.md — Plants tab, picker, Settings EC, Guardian"
    status: pending
  - id: ws2-tour
    content: "WS2: operator-tour § Plants — dropdown, targets preview, no strain copy"
    status: pending
  - id: ws3-architecture
    content: "WS3: farm-guardian-architecture §7.0af — Plants + crop APIs + lookup_crop_targets"
    status: pending
  - id: ws4-guardian-smokes
    content: "WS4: Guardian smokes — EC matches UI; compare crops; ramps block; DB registry"
    status: pending
  - id: ws5-closure
    content: "WS5: phase-87-closure.md + phase-14 rows; persona platform_context mirror"
    status: pending
isProject: false
---

# Phase 87 — Crop knowledge & Guardian closure

## Status

**Planned.** Capstone for [84–87 roadmap](phase_84_87_crop_identity_roadmap.plan.md).

**Depends on:** [Phase 85](phase_85_catalog_bound_plants.plan.md), [Phase 86](phase_86_grow_ops_catalog_chain.plan.md).

**Closure:** **OC-87**

---

## The one job

> Operators and Guardian both trust **one farm knowledge base**: Postgres catalog, dropdown plants, Settings EC tweaks, structured targets in chat — no YAML, no typed crop names, no invented EC.

---

## How operators use Plants + knowledge base

| Step | Where | What happens |
|------|-------|--------------|
| 1 | **My zones → Plants** | See current grow + plants in zone |
| 2 | **+ Add plant** | Dropdown from DB (~46+ crops); see EC / light / watering preview |
| 3 | **Settings → Crops & targets** | Tune EC for `cannabis`, `tomato`, … on **this farm** |
| 4 | **Start grow** | Pick catalog plant; stage drives strip + feeding context |
| 5 | **Water / Light** tabs | Targets from same profile stages |
| 6 | **Farm Guardian** | Ask feed/light/compare questions — answers from **`lookup_crop_targets`**, not guesswork |
| 7 | **Bootstrap** | `make guardian-bootstrap-farm` — field guides for narrative depth |

**Terminology:** **Plant** and **crop** — never “strain” in operator UI.

**Extend catalog** (flowers, cacti, San Pedro): edit `data/crop_library.yaml` → regenerate seed SQL → migrate. Not typed in by operators.

---

## WS3 — Architecture (Guardian + crop APIs)

Add **`§7.0af Plants & crop knowledge chain`** to [farm-guardian-architecture.md](../farm-guardian-architecture.md):

| Layer | Guardian access |
|-------|-----------------|
| Catalog metadata | DB `crop_catalog_entries` / aliases (same as `/commons/crop-catalog`) |
| Structured targets | `lookup_crop_targets` → effective `crop_profiles` + stages (same as picker) |
| Farm EC override | Farm profile row — immediate on chat turn |
| Active grow | `crop_cycles.plant_id` → `plants.crop_key` |
| Narrative | RAG field guides — supplement only; numbers still from read tool |
| Hard rule | No EC/pH/VPD/DLI/photoperiod without read-tool output |

Mirror summary in [farm-guardian-persona-platform-context.md](../farm-guardian-persona-platform-context.md).

---

## WS1 — Operator runbook

New doc: `docs/crop-knowledge-operator-runbook.md`

1. `make migrate` + `make check-crop-catalog-parity`
2. Fix picker 404 (restart API)
3. Add plant from dropdown (Zone or Plants workspace)
4. Adjust EC in Settings
5. Start grow → verify strip chips
6. Ask Guardian — verify numbers match strip
7. Unsupported crop smoke (ramps)
8. Integrator: add catalog crops via seed migration

---

## WS4 — Guardian + UI parity smokes

| # | Check |
|---|--------|
| 1 | `GET /commons/crop-catalog` ≥ 50 entries |
| 2 | `GET /farms/1/crop-library/picker` ≥ 46 with `has_targets` |
| 3 | POST plant `crop_key=tomato` twice → one row |
| 4 | POST plant `crop_key=ramps` → 400 |
| 5 | Start grow → zone EC chip == profile stage EC |
| 6 | PUT farm override cannabis EC → picker preview + strip update |
| 7 | **Guardian:** “EC for early flower” with active cannabis cycle → matches strip (mS/cm) |
| 8 | **Guardian:** compare cannabis vs tomato → both from DB stages |
| 9 | **Guardian:** “EC for ramps” → unsupported; no invented targets |
| 10 | **Guardian:** registry uses DB when `CROP_CATALOG_SOURCE=db` (alias mention resolves) |

Optional: `cmd/api/smoke_phase87_test.go` wrapping chat read-tool injection tests.

---

## WS5 — Closure

- `docs/plans/phase-87-closure.md`
- Phase 14 index: 84–87 all **shipped**
- Re-ingest: `make rag-ingest-platform-docs` after doc edits

**Prompt loop:** `phase 87 ws4` for Guardian smokes only, or **`phase 87`** for full closure.
