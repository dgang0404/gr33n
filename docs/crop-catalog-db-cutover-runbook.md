# Crop catalog & field guides — DB cutover runbook (Phase 84 WS-E / WS-G / WS-H)

Operators and integrators moving from repo YAML/MD at runtime to **Postgres as source of truth**.

**Related:** [sit-in-crop-catalog-enterprise-db.md](workstreams/sit-in-crop-catalog-enterprise-db.md)

---

## Defaults (WS-G)

After Phase 84 WS-G, production defaults are:

| Env | Default | Legacy override |
|-----|---------|-----------------|
| `CROP_CATALOG_SOURCE` | `db` | `yaml` (authoring / offline validation only) |
| `AGRONOMY_FIELD_GUIDES_SOURCE` | `db` | `file` (deprecated — reads `docs/rag/field-guide-manifest.yaml`) |

API pods and `rag-ingest` **do not need the git checkout** when defaults are used — only `DATABASE_URL` and migrations applied.

---

## One-time cutover checklist

```bash
make migrate
make check-crop-catalog-parity   # YAML authoring still valid in git
make check-crop-catalog-db       # Postgres row counts + WS-I meta

# Optional explicit (defaults are db after WS-G):
export CROP_CATALOG_SOURCE=db
export AGRONOMY_FIELD_GUIDES_SOURCE=db

make rag-ingest-field-guides     # per farm; needs EMBEDDING_API_KEY

# Enterprise one-command bootstrap (Phase 83):
# ./scripts/enterprise/guardian-bootstrap-farm.sh --farm-id 1
```

Restart the API after migrate so Guardian and the crop picker load catalog from DB.

---

## WS-H — Re-ingest field guides (required once per farm)

### Why

Historically, `rag-ingest -field-guides` used **path-hash** `source_id` values (FNV of manifest path). DB-backed guides use **`agronomy_field_guides.id`** as `source_id`.

Until you re-ingest, farms may have **stale or duplicate** `field_guide` chunks (old hash ids vs new guide ids).

### What to run

For each farm with embeddings configured:

```bash
export AGRONOMY_FIELD_GUIDES_SOURCE=db   # default after WS-G
make rag-ingest-field-guides             # farm 1 by default
# or:
go run ./cmd/rag-ingest -farm-id N -field-guides
```

Dry-run (no API key):

```bash
make rag-ingest-field-guides-dry-run
```

Ingest **upserts** by `(farm_id, source_type, source_id)` — re-running is safe. Old path-hash chunks are **not** auto-deleted; optional cleanup:

```sql
-- Inspect stale field_guide chunks (example farm_id=1):
SELECT source_id, count(*)
FROM gr33ncore.rag_embedding_chunks
WHERE farm_id = 1 AND source_type = 'field_guide'
GROUP BY source_id ORDER BY source_id;

-- After confirming new guide ids exist, delete orphans manually if needed.
```

### When to re-ingest again

| Event | Action |
|-------|--------|
| First DB cutover | Full field-guide ingest per farm |
| `body_md` updated in DB / new seed migration | Re-ingest affected farms |
| New farm bootstrap | Ingest as part of guardian bootstrap |
| Alias-only catalog change | No re-ingest |

---

## Authoring workflow (unchanged in git)

1. Edit `data/crop_library.yaml` and `docs/field-guides/*.md`
2. Optional catalog thumbnail: add `ui/public/assets/crops/{crop_key}.svg` and set `crop_catalog_entries.image_url` in seed SQL (Phase 107 — `/assets/crops/{crop_key}.svg`)
3. Regenerate seed: `./scripts/generate-crop-catalog-seed.sql.sh -o db/migrations/…`
4. Add migration + `make migrate`
5. Re-ingest farms if guide text changed

`make check-crop-catalog-parity` validates YAML/manifest **before** SQL generation — keep running in CI.

---

## API verification

JWT required:

```bash
curl -s -H "Authorization: Bearer $TOKEN" "$API/commons/crop-catalog" | jq '.count'
curl -s -H "Authorization: Bearer $TOKEN" "$API/commons/crop-catalog/tomato" | jq '.crop_profile_id'
curl -s -H "Authorization: Bearer $TOKEN" "$API/commons/agronomy-field-guides?crop_key=tomato" | jq 'length'
```

---

## Troubleshooting

| Symptom | Fix |
|---------|-----|
| `crop catalog empty — run migrate` | `make migrate` (Phase 84 seed migrations) |
| Guardian still missing crops | Restart API; confirm `CROP_CATALOG_SOURCE` not `yaml` |
| RAG cites wrong / missing crop narrative | Re-ingest with `AGRONOMY_FIELD_GUIDES_SOURCE=db` |
| Picker works, no substrate in chat | WS-I migration `20260617_phase84_crop_profile_meta_from_catalog.sql` |
| Air-gapped API without repo | Must use `db` defaults; ship migrations in image |
