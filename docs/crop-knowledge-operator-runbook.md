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
| 7 | **Farm Guardian** | Feed / light / compare questions → **`lookup_crop_targets`** (mS/cm) — not guesswork |

**Copy:** Use **plant** and **crop** — not “strain” in operator UI.

**Growth stages:** UI dropdowns load all 11 stages (`transition`, `flush`, …) from `GET /platform/domain-enums` — same vocabulary as Postgres `growth_stage_enum`, OpenAPI, and Guardian `lookup_crop_targets` / cycle tools.

**Extend the catalog** (new crops, aliases): edit `data/crop_library.yaml` → regenerate seed SQL → migrate. Operators do not type new crop identities in the UI.

---

## What Settings EC affects (v1 scope)

- **Now:** A Settings override for `crop_key` (e.g. `cannabis`) applies to **all grows of that crop on this farm** — strip, Water/Light hints, picker preview, and Guardian on the next chat turn (no RAG re-ingest).
- **Not v1:** Per-genetics EC (Blue Dream vs OG Kush) → [Phase 94](plans/phase_94_genetics_batch_ec_profiles.plan.md).
- **Not v1:** Per-batch run EC → Phase 94 or a separate farm override model.

---

## Guardian checks (manual)

With an **active cannabis grow** in early flower:

1. Zone strip shows **EC target … mS/cm**.
2. Ask Guardian: *Is my EC on target for early flower?* — numbers must match the strip (mS/cm, not %).
3. Ask: *Compare cannabis and tomato EC targets* — both crops from DB stage rows.
4. Ask: *What EC for ramps?* — unsupported block; **no** invented targets.

**Bootstrap narrative depth:**

```bash
make guardian-bootstrap-farm FARM_ID=1
```

Field guides supplement RAG; structured numbers still come from `lookup_crop_targets`. See [Phase 97 — RAG vs structured truth](plans/phase_97_rag_structured_truth_governance.plan.md).

**Enterprise promotion:** [Phase 98](plans/phase_98_enterprise_catalog_promotion.plan.md) (when applicable).

---

## Integrator

| Task | Doc |
|------|-----|
| DB cutover | [`crop-catalog-db-cutover-runbook.md`](crop-catalog-db-cutover-runbook.md) |
| Guardian bootstrap | [`scripts/enterprise/README.md`](../scripts/enterprise/README.md) |
| Add catalog crops | Regenerate from `crop_library.yaml` → [Phase 95](plans/phase_95_catalog_integrator_ops.plan.md) |

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
| `TestPhase64_*` / `TestPhase82_*` | Profile library + picker API |

---

## Related

- [Operator tour §6m — Plants & crop chain](operator-tour.md#6m-plants--crop-knowledge-chain-phases-8587--shipped)
- [Farm Guardian architecture §7.0af](farm-guardian-architecture.md#70af-plants--crop-knowledge-chain-phases-8587--shipped)
- [Phase 87 closure](plans/phase-87-closure.md)
