# Commons catalog (gr33n_inserts) — operator playbook

Phase 14 **WS3** + Phase 207 **auto-apply & publish**: browse published packs, **import applies automatically**, and farm admins can **publish recipe packs** from the UI.

**API:** `GET/POST /commons/catalog`, `GET /commons/catalog/{slug}`, `POST /farms/{id}/commons/catalog-imports`, `POST /farms/{id}/commons/catalog-export/recipe-pack`, `GET /farms/{id}/commons/catalog-imports`.

**UI:** **More → Help → Library → Import** (`/operator-guide?tab=catalog`) — Browse, Farm Imports, **Publish from Farm**.

## What this is (plain language)

| Term | Meaning |
|------|---------|
| **Starter pack** | A JSON document in the catalog — fertigation programs, an agronomy checklist, or documentation |
| **Browse** | See packs published on **this server** (shipped by gr33n migrations or published by your team) |
| **Import to Farm** | One click: records provenance **and applies** the pack (creates programs, verifies agronomy, etc.) |
| **Publish from Farm** | Export **this farm's** fertigation programs as a new catalog pack other farms on the same server can import |

**Not the same as:** field guides / Guardian RAG (knowledge for chat) or Insert Commons (anonymous stats **out** — see [`insert-commons-pipeline-runbook.md`](insert-commons-pipeline-runbook.md)).

## How packs get onto a server

| Source | How |
|--------|-----|
| **gr33n ship** | SQL migrations seed demo packs (Recipe Pack v7, Agronomy Seed Pack, docs) |
| **Your team** | **Help → Catalog → Publish from Farm** or `POST /commons/catalog` with validated JSON |
| **Another server** | No live federation in v1 — copy JSON + `POST /commons/catalog` on the target server, or ship a migration |

There is **no** public gr33n.com app store the UI calls over the internet. Catalog is **local to each deployment's Postgres**.

## Import flow (automatic apply)

```
Farm admin → Import to Farm
    → POST /farms/{id}/commons/catalog-imports
    → audit row in farm_commons_catalog_imports
    → apply by body.kind (see below)
    → response includes "apply": { status, message, programs_created, next_steps, ... }
```

### Pack kinds

| `body.kind` | On import |
|-------------|-----------|
| `fertigation_recipe_pack` | Creates fertigation programs **by name** (idempotent). Updates metadata if program already exists. **`is_active` stays false** unless the pack says otherwise — review in Zones → Water before enabling. |
| `agronomy_seed_pack` | Verifies platform crop catalog version + row counts. Does **not** run RAG ingest — follow `next_steps` (Settings → Field memories → Re-ingest or `make guardian-bootstrap-farm`). |
| `documentation_pack` | Audit only — no farm data changes. |

Unknown kinds: import audit recorded; apply `status: skipped`.

### API response shape

```json
{
  "import": { "id": 1, "farm_id": 1, "imported_at": "..." },
  "catalog_entry": { "slug": "...", "title": "..." },
  "apply": {
    "kind": "fertigation_recipe_pack",
    "status": "applied",
    "message": "Fertigation programs imported...",
    "programs_created": 2,
    "programs_skipped": 1,
    "next_steps": ["Open Zones → Water..."]
  }
}
```

If apply fails (e.g. unknown `crop_key`), HTTP **400** with `"error"` and partial `"apply"`.

## Publish flow (user-published packs)

### From the UI (recommended)

1. **More → Help → Catalog → Publish from Farm**
2. Enter slug (lowercase, hyphens), title, summary
3. **Export farm programs & publish** — exports all fertigation programs on the selected farm as `fertigation_recipe_pack` (inactive)

Requires **farm admin** on the source farm.

### From the API

**Custom pack:** `POST /commons/catalog`

```json
{
  "slug": "co-op-lettuce-2026",
  "title": "Co-op lettuce recipes",
  "summary": "Shared veg/flower profiles",
  "license_spdx": "CC-BY-4.0",
  "contributor_display": "River Valley Co-op",
  "body": {
    "catalog_version": "gr33n.commons_catalog.v1",
    "kind": "fertigation_recipe_pack",
    "readme_md": "# Co-op pack\n...",
    "programs": [ { "name": "...", "total_volume_liters": 2, ... } ]
  }
}
```

**Export from farm:** `POST /farms/{id}/commons/catalog-export/recipe-pack`

```json
{
  "slug": "farm-1-export-v1",
  "title": "Demo farm recipes",
  "summary": "Exported from farm 1"
}
```

Validation: known `kind`, recipe `crop_key` tags must exist in `GET /commons/crop-catalog`, slug format, no duplicate slugs.

## Trust & security (read before multi-user deploy)

Import is **configuration**, not code execution. No shell, no arbitrary SQL, no eval — apply paths are whitelisted in Go (`internal/commonscatalog/apply.go`).

| Question | Answer |
|----------|--------|
| Can import run on my farm without me? | **No** — `POST .../catalog-imports` requires **farm admin** on that farm. |
| Can import touch another farm? | **No** — scoped to `{farm_id}` in the URL. |
| Can unknown JSON do anything? | **No** — unknown `body.kind` → audit row only (`apply.status: skipped`). |
| Is this RAG / field guides? | **No** — separate pipeline; catalog does not embed chunks. |
| Is this Insert Commons? | **No** — Insert Commons is stats **out** ([`insert-commons-pipeline-runbook.md`](insert-commons-pipeline-runbook.md)). |

**Real risks (operational trust, not RCE):**

1. **Publish is open to any logged-in user** on this deployment (`POST /commons/catalog`). A bad actor can list a pack everyone can browse. Mitigation today: **LAN / trusted users only**; only import packs you trust; check `contributor_display` and `published_by_user_id` on the entry.
2. **Farm admin chooses to import** — same trust as applying a bootstrap template. A malicious recipe pack could set bad EC/pH or `is_active: true`. Demo/export packs default inactive; **review Zones → Water** before enabling automation.
3. **No remote catalog federation in v1** — packs are local Postgres. Another server does not push packs over WAN unless you copy JSON or run a migration.

**vs bootstrap templates (Settings / farm setup wizard):** templates use fixed DB function keys; only gr33n migrations seed them. Catalog **publish** is looser (any auth user); catalog **apply** is narrower (mostly fertigation programs).

**ponytail ceiling:** multi-tenant sign-up needs publish restricted to platform/org admin + force `is_active: false` on import. Upgrade path: gate `POST /commons/catalog`, org-scoped catalog rows.

## Three “pack” words (don't mix these up)

| Name | Where | Direction | Creates |
|------|-------|-----------|---------|
| **Bootstrap template** | Settings, farm setup wizard | In | Zones, sensors, schedules, rules, tasks |
| **Commons Catalog pack** | Help → Catalog | In | Fertigation programs (+ agronomy verify) |
| **Insert Commons sync** | Settings | **Out** | Anonymous aggregate JSON to optional receiver |

## Enterprise scripts (still supported)

Bulk multi-farm promotion:

```bash
./scripts/enterprise/import-recipe-pack.sh --farm-ids 1,2,3
./scripts/enterprise/import-agronomy-seed-pack.sh --farm-ids 1
./scripts/enterprise/guardian-bootstrap-farm.sh --farm-id 1
```

The API import endpoint now performs the same program creation as `import-recipe-pack.sh` for a single farm.

## Data model

- **`gr33ncore.commons_catalog_entries`** — published packs (`body` JSONB, `published_by_user_id`, `source_farm_id`)
- **`gr33ncore.farm_commons_catalog_imports`** — per-farm provenance

Migration: `db/migrations/20260719_commons_catalog_publish_apply.sql`

## Testing checklist (developer / QA)

- [ ] Browse catalog — 3 seeded entries
- [ ] Import documentation pack — `apply.status = noop`
- [ ] Import Recipe Pack v7 — programs created or skipped; check Zones → Water
- [ ] Import Agronomy Seed Pack — `apply.status = verified` or clear failure message
- [ ] Publish from Farm — new entry in browse list; import on second farm
- [ ] Re-import same pack — idempotent audit + program skip by name

## Related

- Insert Commons (stats **out**): [`insert-commons-pipeline-runbook.md`](insert-commons-pipeline-runbook.md)
- Enterprise promotion: [`enterprise-catalog-promotion-model.md`](enterprise-catalog-promotion-model.md)
- Phase 14 plan: [`plans/archive/phase_14_network_and_commons.plan.md`](plans/archive/phase_14_network_and_commons.plan.md)
