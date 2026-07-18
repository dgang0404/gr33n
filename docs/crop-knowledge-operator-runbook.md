# Crop knowledge — operator runbook

How operators and Farm Guardian use the **same Postgres-backed crop catalog**: dropdown plants, Settings EC, zone grow strip, Water/Light targets, and `lookup_crop_targets`.

**Arc closure:** Phases 84–87 · **Integrator cutover:** [`crop-catalog-db-cutover-runbook.md`](crop-catalog-db-cutover-runbook.md)

---

## Prerequisites

```bash
make migrate
make check-crop-catalog-parity
```

- API running with **`CROP_CATALOG_SOURCE=db`** (default in `.env` and docker-compose).
- Restart the API after migrate so `/farms/{id}/crop-library/picker` is registered.

**Picker 404 or empty dropdown?** Run migrate, restart API, confirm parity check passes. The UI shows an amber **Knowledge base API outdated** banner when falling back to legacy profiles — do not treat that as the full catalog.

### LAN-only / brief API outage (Phase 100)

After at least one successful load while online, the UI caches the crop picker in **IndexedDB** (per farm) and domain enums locally. If the API is unreachable (network error, not 404):

- **Plants** / **Add plant** / **Start grow** dropdowns show the last cached catalog with a blue **Offline — showing cached knowledge base** banner.
- If the platform catalog was upgraded while you were offline, you may also see **New crops may be available — reconnect and reload**.
- A **404** still means migrate/restart — the UI does **not** substitute the full cached catalog for a missing picker route.

**Mobile / warehouse Wi‑Fi:** The UI ships with **vite-plugin-pwa** (service worker for static assets). Catalog data is cached separately in IndexedDB — no extra PWA config is required for picker offline mode. For Capacitor field builds, one online session before going offline is enough to populate the cache.

**Upgrading an existing farm (pre–Phase 85 plants)?** After `make migrate`, run `./scripts/merge-legacy-plants.sh` (audit) then `./scripts/merge-legacy-plants.sh --apply --audit`. This merges typo plant rows (Tomato / tomato / Romas) into one `crop_key` slot and relinks `crop_cycles.plant_id`. Rows that still lack `crop_key` need a manual catalog pick in **Zone → Plants**.

---

## Operator flow

| Step | Where | What happens |
|------|-------|--------------|
| 1 | **Zone → Plants** | See current grow + plants linked to this zone |
| 2 | **+ Add plant** | **Crop from knowledge base** dropdown (~46+ crops); EC / DLI / photoperiod preview under the picker |
| 3 | **Settings → Crops & targets** | Tune EC (and stages) per `crop_key` for **this farm** |
| 4 | **Start grow** | Pick catalog crop type → creates/links a catalog plant → active cycle with `plant_id` |
| 5 | **Zone grow strip** | EC chip for `current_stage` from effective profile |
| 6 | **Water / Light** tabs | **Crop targets** hint from the same profile stage |
| 7 | **Farm Guardian** | Grow questions → **`plant_context_bundle`** (cycle + targets + sensors + feed + light) or **`lookup_crop_targets`** alone (mS/cm) |

**Copy:** Use **plant** and **crop** — not “strain” in operator UI.

**Growth stages:** UI dropdowns load all 11 stages (`transition`, `flush`, …) from `GET /platform/domain-enums` — same vocabulary as Postgres `growth_stage_enum`, OpenAPI, and Guardian `lookup_crop_targets` / cycle tools.

**Extend the catalog** (new crops, aliases): edit `data/crop_library.yaml` → follow [`catalog-integrator-playbook.md`](catalog-integrator-playbook.md) (Phase 95). Operators do not type new crop identities in the UI.

---

## What Settings EC affects (v1 vs v2)

- **Farm-wide (v1):** A Settings override for `crop_key` (e.g. `cannabis`) applies to **all grows of that crop on this farm** when no genetics profile is linked — strip, Water/Light hints, picker preview, and Guardian on the next chat turn (no RAG re-ingest).
- **Per-genetics (v2 — Phase 94):** When a plant has **`variety_or_cultivar`** and the farm has a **genetics EC profile** for that variety, that profile wins over the farm-wide `crop_key` table. Manage via API `PUT /farms/{id}/crop-profiles/{crop_key}/genetics/{variety_slug}` or the **Tune EC for this variety →** link on Plants.
- **Not yet:** Per-batch run EC without a reusable variety label → separate farm override model or future phase.

---

## Structured truth vs RAG (Phase 97)

Guardian uses **two layers** for crop knowledge:

| Source | Updates when | Use for |
|--------|--------------|---------|
| **Structured** (`crop_profiles`, farm override, genetics) | Immediately on Settings PUT / effective API | EC, pH, VPD, DLI, photoperiod (**mS/cm**) |
| **RAG** (field guides) | After `make rag-ingest-field-guides` | Qualitative narrative — deficiency signs, timing, mistakes |

**Rule:** When `lookup_crop_targets` or **`plant_context_bundle`** runs on a chat turn, those numbers **win** over field-guide narrative EC. Farm EC overrides do **not** require RAG re-ingest.

**Phase 136 — plant context bundle:** From the zone grow strip, **How is this grow doing?** sends `crop_cycle_id` so Guardian fuses cycle stage, targets, live sensors, fertigation, lighting, and `grow_advisor` in one block (~800 token cap). Symptom questions append `lookup_crop_symptoms` in the same bundle.

### When to re-ingest field guides

| Event | Structured profile | Re-ingest RAG? |
|-------|-------------------|----------------|
| Farm EC override (Settings) | ✅ immediate | ❌ not required for numbers |
| Genetics EC profile (Phase 94) | ✅ immediate | ❌ not required for numbers |
| Platform catalog seed bump + migrate | ✅ after migrate | ✅ if guide **body** changed |
| YAML EC edit + new migration | ✅ after migrate | ✅ re-ingest affected farms |
| Operator chat (feed / EC question) | Read tool first | Narrative supplement only |

```bash
make rag-ingest-field-guides          # per farm; needs EMBEDDING_API_KEY
make rag-ingest-field-guides-dry-run  # chunk estimate only
```

RAG chunks store `crop_key` + `catalog_version` metadata (Phase 97) for stale detection after catalog bumps.

---

With an **active cannabis grow** in early flower:

1. Zone strip shows **EC target … mS/cm**.
2. Ask Guardian: *Is my EC on target for early flower?* — numbers must match the strip (mS/cm, not %).
3. Ask: *Compare cannabis and tomato EC targets* — both crops from DB stage rows.
4. Ask: *What EC for ramps?* — unsupported block; **no** invented targets.

**Bootstrap narrative depth:**

```bash
make guardian-bootstrap-farm FARM_ID=1
```

Field guides supplement RAG; structured numbers still come from `lookup_crop_targets`. See [`phase-97-closure.md`](plans/phase-97-closure.md).

**Enterprise promotion:** [Phase 98](plans/archive/phase_98_enterprise_catalog_promotion.plan.md) (when applicable).

---

## Integrator

| Task | Doc |
|------|-----|
| DB cutover | [`crop-catalog-db-cutover-runbook.md`](crop-catalog-db-cutover-runbook.md) |
| Guardian bootstrap | [`scripts/enterprise/README.md`](../scripts/enterprise/README.md) |
| Add catalog crops | Regenerate from `crop_library.yaml` → [Phase 95](plans/archive/phase_95_catalog_integrator_ops.plan.md) |
| Override audit trail (compliance) | [`audit-events-operator-playbook.md`](audit-events-operator-playbook.md) — `crop_profile_override_*` kinds; Settings → Farm audit trail |

---

## Automated tests

| Test | What it proves |
|------|----------------|
| `TestPhase85CatalogBoundPlants` | `crop_key` upsert; `ramps` → 400 |
| `TestPhase86_*` | Active cycle requires plant; Guardian EC matches profile |
| `TestPhase87_*` | Picker parity; multi-crop compare; DB registry alias |
| `TestPhase88_*` | Domain enums API; `transition` setpoint persists |
| `TestPhase101_*` | Guardian `create_plant` requires `crop_key` |
| `TestPhase103_*` | Legacy plant merge; no duplicate `crop_key` per farm |
| `TestPhase106_*` | Guardian `lookup_crop_symptoms` — tomato yellow leaves + EC targets |
| `TestPhase107_*` | Commons + picker `image_url` for ornamentals |
| `TestPhase64_*` / `TestPhase82_*` | Profile library + picker API |
| `TestPhase94_*` | Genetics EC profile beats farm `crop_key` override on effective API |
| `TestPhase95_*` | Picker `version` matches YAML; catalog crop in commons + picker |
| `TestPhase136_*` | `plant_context_bundle` on demo veg grow |
| `TestPhase97_*` | Farm override EC in `lookup_crop_targets`; stale RAG numbers stripped |
| `TestPhase98_*` | Farm A EC override does not change Farm B builtin profile |

---

## Related

- [Operator tour §6m — Plants & crop chain](operator-tour.md#6m-plants--crop-knowledge-chain-phases-8587--shipped)
- [Farm Guardian architecture §7.0af](farm-guardian-architecture.md#70af-plants--crop-knowledge-chain-phases-8587--shipped)
- [Phase 87 closure](plans/phase-87-closure.md)
