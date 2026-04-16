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

Apply **`db/migrations/20260426_commons_catalog.sql`** (includes one demo documentation pack).

## Licensing and attribution

Operators should set **`license_spdx`** (e.g. `CC-BY-4.0`) and **`license_notes`** for human-readable terms. The API does not enforce license compliance; downstream use is the operator’s responsibility.

## Non-goals (v1)

- User-submitted catalog entries via API (curation is DB/ops for now).
- Executable SQL in catalog `body` applied automatically by the API.
- Cross-farm analytics on catalog usage beyond per-farm import lists.

## Related

- Phase 14 plan: [`plans/phase_14_network_and_commons.plan.md`](plans/phase_14_network_and_commons.plan.md)
- Insert Commons (separate feature): [`insert-commons-pipeline-runbook.md`](insert-commons-pipeline-runbook.md)
- Phase 14 operator index: [`phase-14-operator-documentation.md`](phase-14-operator-documentation.md)
