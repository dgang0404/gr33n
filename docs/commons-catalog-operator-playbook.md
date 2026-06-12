# Commons catalog (gr33n_inserts) — operator playbook

Phase 14 **WS3** introduces a **published catalog** of contribution-style packs: metadata, **license hints**, and a JSON **`body`** per entry. This is **not** a marketplace: there are no payments, ratings, or third-party listings in v1.

**API:** `GET /commons/catalog`, `GET /commons/catalog/{slug}`, `GET /farms/{id}/commons/catalog-imports`, `POST /farms/{id}/commons/catalog-imports` (see [`openapi.yaml`](../openapi.yaml), tag **commons**).

## Goals

| Goal | Mechanism |
|------|-----------|
| **Discover** | Authenticated users browse published rows (search `q`, pagination `limit` / `offset`). |
| **Inspect** | Detail endpoint returns full `body` JSON (e.g. `catalog_version`, `kind`, `readme_md`, `related_urls`). |
| **Provenance** | Farm **admins** record an **import** (idempotent per farm + entry); list imports per farm. |
| **No auto-exec** | Import does **not** run SQL against the database; it records audit-friendly linkage and returns the payload for tools or future workers. |

## Data model

- **`gr33ncore.commons_catalog_entries`** — curator-published rows (`slug`, `title`, `summary`, `body` JSONB, `contributor_*`, `license_spdx`, `tags`, `published`).
- **`gr33ncore.farm_commons_catalog_imports`** — `(farm_id, catalog_entry_id)` unique; `imported_by`, `imported_at`, optional `note`.

Apply **`db/migrations/20260426_commons_catalog.sql`** (includes one demo documentation pack). Phase 31 WS5 adds **`20260527_phase31_commons_recipe_pack_v7.sql`** (fertigation recipe pack demo) — promote with [`scripts/enterprise/import-recipe-pack.sh`](../../scripts/enterprise/import-recipe-pack.sh). Phase 108 tags pack programs with **`recommended_crop_keys`** / **`recommended_stages`** (Phase 102 metadata) via **`20260627_phase108_commons_recipe_pack_crop_tags.sql`**. Phase 83 adds **`20260618_phase83_cultivator_seed_pack_v1.sql`** — agronomy seed pack — promote with [`scripts/enterprise/import-agronomy-seed-pack.sh`](../../scripts/enterprise/import-agronomy-seed-pack.sh).

## Recipe pack (`gr33n-recipe-pack-v7-lettuce-veg`)

**Kind:** `fertigation_recipe_pack` in catalog `body`.

| Field | Meaning |
|-------|---------|
| `programs[]` | Fertigation program payloads promoted per farm |
| `programs[].recommended_crop_keys` | Phase 102 tags — validated against `GET /commons/crop-catalog` on import |
| `programs[].recommended_stages` | Growth stages for program suggest / fit warnings |
| `programs[].profile_ec_source` | Optional `{ crop_key, stage }` for EC band provenance |
| `programs[].ec_band_mscm` | Optional denormalized EC min/max (mS/cm) |

**Import semantics:** `import-recipe-pack.sh` records **`POST /farms/{id}/commons/catalog-imports`**, creates programs idempotently by **`name`**, and applies metadata via **`PATCH /fertigation/programs/{id}/metadata`**. Unknown **`crop_key`** values fail the import (parity with catalog).

## Agronomy seed pack (`gr33n-cultivator-seed-pack-v1`)

**Kind:** `agronomy_seed_pack` in catalog `body`.

**Promotion model (Phase 98):** This pack is **not** the platform catalog. Platform crops live in Postgres after **`make migrate`**. The commons import is an **optional org audit record** per farm — see [`enterprise-catalog-promotion-model.md`](enterprise-catalog-promotion-model.md).

| Field | Meaning |
|-------|---------|
| `platform_catalog_version` | Expected Postgres `crop_catalog_*` version after migrate |
| `expected_counts` | Sanity checks (supported crops, field guides) — import script verifies |
| `readme_md` | Operator-facing notes; not executed by API |

**Import semantics:** same as any commons entry — `POST /farms/{id}/commons/catalog-imports` records audit linkage and returns the body JSON. Import does **not** run migrations or RAG ingest. Integrators follow with:

1. `make check-crop-catalog-parity` (platform DB already seeded by migrate)
2. [`guardian-bootstrap-farm.sh`](../../scripts/enterprise/guardian-bootstrap-farm.sh) for RAG ingest
3. Optional [`apply-agronomy-overrides.sh`](../../scripts/enterprise/apply-agronomy-overrides.sh) or Settings **Crops & targets**

**Idempotency:** one import row per `(farm_id, catalog_entry_id)`; re-import updates audit timestamp only.

See [`scripts/enterprise/README.md`](../../scripts/enterprise/README.md) and [`plans/phase-83-closure.md`](plans/phase-83-closure.md).

## Licensing and attribution

Operators should set **`license_spdx`** (e.g. `CC-BY-4.0`) and **`license_notes`** for human-readable terms. The API does not enforce license compliance; downstream use is the operator’s responsibility.

## Non-goals (v1)

- User-submitted catalog entries via API (curation is DB/ops for now).
- Executable SQL in catalog `body` applied automatically by the API.
- Cross-farm analytics on catalog usage beyond per-farm import lists.

## Related

- Phase 14 plan: [`plans/phase_14_network_and_commons.plan.md`](plans/phase_14_network_and_commons.plan.md)
- **Enterprise promotion model (Phase 98):** [`enterprise-catalog-promotion-model.md`](enterprise-catalog-promotion-model.md)
- Insert Commons (separate feature): [`insert-commons-pipeline-runbook.md`](insert-commons-pipeline-runbook.md)
- Phase 14 operator index: [`phase-14-operator-documentation.md`](phase-14-operator-documentation.md)
