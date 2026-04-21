# Database schema overview

This page explains **where the real schema lives** so contributors are not blocked by stale diagrams or old snapshots.

## Canonical sources

| Artifact | Role |
|----------|------|
| **[db/schema/gr33n-schema-v2-FINAL.sql](../db/schema/gr33n-schema-v2-FINAL.sql)** | Full baseline for a **new** database (extensions, schemas, tables). |
| **[db/migrations/*.sql](../db/migrations/)** | Incremental changes — applied in **lexicographic filename order** after the baseline on upgrades. **[scripts/bootstrap-local.sh](../scripts/bootstrap-local.sh)** applies schema then sorted migrations. |

If a drawing or PDF disagrees with those files, **trust the SQL**.

## Logical layout (multi-schema Postgres)

The platform uses multiple schemas (examples from the baseline file — always confirm in SQL):

- `auth` — authentication-related objects where present
- `gr33ncore` — farms, users, tasks, devices, automation, costs, RAG chunks, etc.
- `gr33nnaturalfarming` — natural-farming inputs and related objects
- `gr33nfertigation` — programs, crop cycles, irrigation linkage
- `gr33ncrops`, `gr33nanimals`, `gr33naquaponics` — domain modules as implemented

Farm-scoped isolation and module toggles follow the same patterns as the rest of the API (see threat model and farm auth docs).

## Diagrams (ERDs) and screenshots

**Entity-relationship diagrams are optional documentation.** Older or external ERDs may reflect an earlier iteration of the product. Do **not** treat them as migration instructions.

When you need a visual:

1. Prefer regenerating from the live database or from `gr33n-schema-v2-FINAL.sql` using a tool you trust (e.g. SchemaSpy, pgAdmin, or schema-as-code exporters).
2. If you publish a diagram in-repo, date it and say **which migration filename** or commit it was generated from.

Good documentation drives adoption; **accurate** schema docs beat pretty but wrong pictures.

## Related reading

- Local database setup: [INSTALL.md](../INSTALL.md), [local-operator-bootstrap.md](local-operator-bootstrap.md)
- RAG storage (pgvector): [rag-scope-and-threat-model.md](rag-scope-and-threat-model.md)
- Pi edge vs full stack vs scaling DB/API/UI: [raspberry-pi-and-deployment-topology.md](raspberry-pi-and-deployment-topology.md)
