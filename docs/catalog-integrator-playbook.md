# Catalog integrator playbook (Phase 95)

**Audience:** Platform team (YAML authoring) and site integrators (migrate + RAG + smoke).

**Closes blind spot #4:** The Plants dropdown stays frozen unless someone runs a **repeatable cadence** from git YAML → Postgres → picker/Guardian.

**Closure:** [Phase 95](plans/archive/phase_95_catalog_integrator_ops.plan.md) · [`phase-95-closure.md`](plans/archive/phase-95-closure.md)

---

## Roles

| Role | Owns |
|------|------|
| **Platform team** | `data/crop_library.yaml`, field guides, seed SQL, `catalog_version` bump, CI |
| **Site integrator** | `make migrate` on each deployment, RAG re-ingest, API restart, smoke |
| **Operators** | Pick crops from the catalog — **never** type new crop identities in UI |

---

## When to use this

- Add a supported crop (ornamental, cactus, San Pedro, new fruiting crop, …)
- Change EC stage targets or substrate metadata for a **platform** crop
- Add or update a field guide body tied to a crop
- Bump **`catalog_version`** so enterprise packs and Phase 109 notifications stay aligned

Farm-specific EC tweaks without new catalog rows → **Settings → Crops & targets** or [Phase 94 genetics profiles](plans/archive/phase_94_genetics_batch_ec_profiles.plan.md).

---

## Integrator checklist

```bash
# ── Platform (git PR) ─────────────────────────────────────────
# 1. Edit sources
vim data/crop_library.yaml
vim docs/field-guides/crop-<key>-nutrition.md   # if new supported crop

# 2. Bump version (required when catalog rows change)
#    version: N   ← monotonic int in crop_library.yaml header

# 3. Regenerate committed SQL (CI drift gate)
./scripts/generate-crop-catalog-seed.sql.sh -o db/seed/crop_catalog_from_yaml.sql
./scripts/generate-crop-catalog-seed.sql.sh -o db/migrations/YYYYMMDD_catalog_<slug>.sql

# 4. Pre-migrate validation (no DB)
make add-crop-check

# 5. If enterprise agronomy pack references platform_catalog_version, bump:
#    scripts/enterprise/sample-cultivator-seed-pack-v1.body.json
#    db/migrations/20260618_phase83_cultivator_seed_pack_v1.sql

# ── Each site (after PR merge) ────────────────────────────────
make migrate
make check-catalog-release
make rag-ingest-field-guides          # if guide body changed; needs EMBEDDING_API_KEY
# restart API pods (CROP_CATALOG_SOURCE=db default)
go test -tags dev ./cmd/api/ -run TestPhase95 -count=1   # optional smoke

# Enterprise bundle
./scripts/enterprise/guardian-bootstrap-farm.sh --farm-id 1   # after major catalog bump
```

---

## `catalog_version` contract

| Location | Field | Rule |
|----------|-------|------|
| `data/crop_library.yaml` | `version:` | **Bump** on every catalog release |
| Generated seed SQL header | `-- crop_library version: N` | Matches YAML |
| `crop_catalog_entries.catalog_version` | per row | Set from YAML on UPSERT |
| `agronomy_field_guides.catalog_version` | per row | Same N on guide UPSERT |
| Agronomy seed pack JSON | `platform_catalog_version` | Must be `<=` DB max after migrate |
| Phase 109 notifications | `platform_catalog_state` | API compares max DB version on startup |

Integrators **do not** hand-edit `catalog_version` in Postgres — always ship a migration.

---

## Make targets

| Target | Needs DB | Purpose |
|--------|----------|---------|
| `make add-crop-check` | No | YAML validate + seed drift gate |
| `make check-catalog-seed-drift` | No | Regenerated SQL vs `db/seed/crop_catalog_from_yaml.sql` |
| `make check-catalog-release` | Yes | Full post-authoring checklist (`add-crop-check` + DB parity) |
| `make check-crop-catalog-parity` | Yes | Phase 84 YAML + row counts (existing) |

---

## Example: add San Pedro cactus

San Pedro (`san_pedro`) is already in the catalog — use this as a **PR walkthrough template**:

1. **YAML** — crop block under `crops:` with stages, substrate, aliases (`trichocereus`, `echinopsis` in top-level `aliases:`).
2. **Field guide** — `docs/field-guides/crop-san-pedro-nutrition.md` listed in `docs/rag/field-guide-manifest.yaml`.
3. **Photos (optional)** — `image_url` in YAML or follow-on migration ([Phase 107](plans/archive/phase_107_crop_catalog_photos.plan.md)).
4. **Regenerate** both canonical seed and dated migration.
5. **PR checklist** — [`docs/templates/add-crop-pr-checklist.md`](templates/add-crop-pr-checklist.md).

Verify after migrate:

```bash
curl -H "Authorization: Bearer $JWT" "$API/commons/crop-catalog/san_pedro"
curl -H "Authorization: Bearer $JWT" "$API/farms/1/crop-library/picker"
```

---

## CI gates

- **`make check-catalog-seed-drift`** — fails when YAML/guides changed without updating `db/seed/crop_catalog_from_yaml.sql`
- **`make check-crop-catalog-parity`** — fails when DB row counts drift after migrate (CI bootstrap)

---

## Related docs

- [Crop catalog DB cutover runbook](crop-catalog-db-cutover-runbook.md)
- [Crop knowledge operator runbook](crop-knowledge-operator-runbook.md) — operator-facing (not integrator)
- [Enterprise README](../scripts/enterprise/README.md)
- [Commons catalog operator playbook](commons-catalog-operator-playbook.md)
- [Phase 14 operator doc index](phase-14-operator-documentation.md)
