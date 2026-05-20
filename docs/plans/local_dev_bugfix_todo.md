---
name: Local dev bugfix backlog
overview: >
  Triage list for local operator pain discovered while standing up Farm Guardian
  on a developer machine. Most items below were fixed 2026-05-20; remaining
  items are optional follow-ups.
status: mostly-resolved
---

# Local dev bugfix backlog

Tracked while running `make dev-stack`, `make restart-local-serve`, and inspecting
the Compose Postgres volume for Guardian demo readiness.

## Fixed (2026-05-20)

| ID | Symptom | Root cause | Fix |
|----|---------|------------|-----|
| **L1** | `make dev-stack` fails on existing DB with `type "farm_scale_tier_enum" already exists` | Bootstrap always re-applied the monolithic schema file; enums lack `IF NOT EXISTS` | `scripts/bootstrap-local.sh` auto-detects provisioned schema (`farm_scale_tier_enum`) and skips the big schema file; migrations + seed still run |
| **L2** | `make dev-stack` (re-seed) fails with `more than one row returned by a subquery` at `master_seed.sql:605` | Zones/schedules used `ON CONFLICT DO NOTHING` but **no unique constraint** on `(farm_id, name)` — every re-run inserted duplicates | Seed: zones + schedules use `WHERE NOT EXISTS`; all zone/schedule subqueries use `ORDER BY id LIMIT 1` |
| **L3** | No Makefile target for “wipe + fresh demo” | Only `./scripts/dev-stack.sh --reset-volumes` was documented | Added **`make dev-stack-fresh`** → `dev-stack.sh --reset-volumes --quick` |
| **L4** | Smoke-test pollution (186k+ alerts, 45 farms) confused Guardian demo | Repeated `make test` against one long-lived DB | Documented: use **`make dev-stack-fresh`** for clean demo; idempotent **`make dev-stack`** for migration-only updates |

## Verified working

```bash
make dev-stack-fresh   # wipe volume → schema → migrations → seed → check-stack
make dev-stack         # idempotent re-run (skip schema, migrate, re-seed safely)
make restart-local-serve   # after reboot: db + sanity + API + UI
```

Clean DB after `dev-stack-fresh`:

| Table | farm_id=1 count |
|-------|-----------------|
| farms | 1 |
| zones | 3 |
| schedules | 13 |
| crop_cycles | 3 |
| alerts | 0 |
| rag_embedding_chunks | 0 (run `rag-ingest` separately) |

Login: `dev@gr33n.local` / `devpassword` (from seed).

## Still open (optional)

| ID | Item | Notes |
|----|------|-------|
| **L5** | **`rag-ingest` in bootstrap** | Seed does not populate embeddings. Consider `make dev-stack-fresh-rag` or post-seed hook when `EMBEDDING_API_KEY` is set. Phase 29 WS3 candidate. |
| **L6** | **Schema unique constraints** | Add `(farm_id, name)` unique partial indexes on `zones` and `schedules` (where `deleted_at IS NULL`) so bad `ON CONFLICT` patterns can't regress. Migration + careful on existing polluted DBs. |
| **L7** | **Smoke test DB isolation** | Long-term: dedicated test DB or transaction rollback per test package to prevent alert/farm accumulation on dev volume. |
| **L8** | **Phase 28 plan frontmatter** | Duplicate `ws6` status line in YAML — cosmetic doc fix. |

## Operator quick reference

| Goal | Command |
|------|---------|
| After reboot, start everything | `make restart-local-serve` |
| Fresh Guardian demo DB | `make dev-stack-fresh` then `rag-ingest` (see bootstrap doc) |
| Apply new migrations only | `make dev-stack` |
| DB only, no API | `make restart-local` |

See [`docs/local-operator-bootstrap.md`](../local-operator-bootstrap.md).
