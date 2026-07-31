# Dual-farm federation test runbook (Phase 212)

**Status:** Executed on laptop 2026-07-30 · Evidence: [`docs/evidence/phase212/`](evidence/phase212/)  
**Glossary:** [`workflow-guide.md`](workflow-guide.md) §11 / §11a

## What this proves

Two independent gr33n installs talk only through:

1. **Insert Commons receiver** (live HTTP aggregates)
2. **Commons Catalog pack hand-carry** (manual JSON copy)

Field guides / symptom catalog rows can appear on both installs because **each** runs the same migrations/seed — that is **not** sync. RAG **embeddings** stay local until each install ingests.

```
Install A (Farm A / Org A) ──POST /v1/ingest──▶ Receiver :8765
Install B (Farm B / Org B) ──POST /v1/ingest──▶ Receiver :8765
Commons pack JSON ──manual copy──▶ Install B catalog → import
```

## Port map

| Service | Install A | Install B |
|---------|-----------|-----------|
| Postgres | 5433 | 5434 |
| API | 8080 | 8081 |
| UI | 5173 | 5174 (optional) |
| Receiver | 8765 (shared) | — |

Install A = hybrid (Compose DB + `go run` / `make dev-auth-test`).  
Install B = hybrid (Compose DB via `docker-compose.phase212-b.yml` + host API). Full Compose API image hit Go 1.26 toolchain pin — documented below as Tier A, fixed with `GOTOOLCHAIN=auto` in [`Dockerfile`](../Dockerfile); B still uses hybrid for RAM.

## Prerequisites

- Free RAM ≳ 2 GB available
- `./scripts/guardian-power.sh sleep` (needs sudo) or Settings → Rest now — Ollama not required
- Install A DB up: `make compose-db-up`

## WS1 — Install B bring-up

```bash
# From Install A repo
./scripts/phase212-install-b-setup.sh
# → clones ~/gr33n-platform-b, starts DB :5434, schema+seed, farm_b_seed.sql

cd ~/gr33n-platform-b && ./scripts/phase212-install-b-serve.sh   # API :8081
# optional UI:
# cd ~/gr33n-platform-b/ui && npm run dev -- --host 0.0.0.0 --port 5174
```

Accept: `curl -sf :8080/health` and `curl -sf :8081/health` both OK.

## WS2 — Organizations + farm names

```bash
# Install A
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f db/seeds/farm_a_org_assign.sql
# → Farm A + Organization A

# Install B (also applied by phase212-install-b-setup.sh)
psql 'postgres://gr33n:gr33n@127.0.0.1:5434/gr33n?sslmode=disable' \
  -v ON_ERROR_STOP=1 -f db/seeds/farm_b_seed.sql
# → Farm B (America/Chicago, CAD) + Organization B
```

## WS3 — Insert Commons receiver

```bash
# One-time DB on Install A Postgres
psql 'postgres://gr33n:gr33n@127.0.0.1:5433/postgres' \
  -c "CREATE DATABASE gr33n_insertcommons;"
psql 'postgres://gr33n:gr33n@127.0.0.1:5433/gr33n_insertcommons?sslmode=disable' \
  -c "CREATE SCHEMA IF NOT EXISTS gr33ncore;"
psql '…/gr33n_insertcommons?sslmode=disable' \
  -f db/migrations/20260417_phase13_insert_commons_receiver.sql
psql '…/gr33n_insertcommons?sslmode=disable' \
  -f db/migrations/20260425_insert_commons_receiver_idempotency_stats.sql

# Same INSERT_COMMONS_SHARED_SECRET on A .env, B .env, receiver env
# Both farms: INSERT_COMMONS_INGEST_URL=http://127.0.0.1:8765/v1/ingest
# Prefer different INSERT_COMMONS_PSEUDONYM_KEY per install → distinct pseudonyms

DATABASE_URL='postgres://gr33n:gr33n@127.0.0.1:5433/gr33n_insertcommons?sslmode=disable' \
INSERT_COMMONS_SHARED_SECRET='…' INSERT_COMMONS_RECEIVER_LISTEN=:8765 \
  go run ./cmd/insert-commons-receiver/

# Restart Install A API after editing .env (process must reload env)

# Opt-in + sync (each install)
curl -X PATCH -H "Authorization: Bearer $JWT" -H 'Content-Type: application/json' \
  -d '{"insert_commons_opt_in":true}' http://127.0.0.1:8080/farms/1/insert-commons/opt-in
curl -X POST -H "Authorization: Bearer $JWT" -H 'Idempotency-Key: a-1' \
  http://127.0.0.1:8080/farms/1/insert-commons/sync
# repeat on :8081

curl -H "Authorization: Bearer $SECRET" http://127.0.0.1:8765/v1/stats
# expect distinct_pseudonyms == 2
```

Evidence: [`evidence/phase212/receiver-stats.json`](evidence/phase212/receiver-stats.json).

## WS4 — Commons Catalog hand-carry

No live “publish to remote server” API (**Tier B** — expected).

```bash
# A: GET a pack body
curl -H "Authorization: Bearer $JWT_A" \
  http://127.0.0.1:8080/commons/catalog/jadam-indoor-starter-recipes-v1 > pack.json

# B: insert catalog row (SQL — POST /commons/catalog returned 500 in this run; Tier C)
# then:
curl -X POST -H "Authorization: Bearer $JWT_B" -H 'Content-Type: application/json' \
  -d '{"slug":"phase212-handcarry-jadam-indoor-starter-v1"}' \
  http://127.0.0.1:8081/farms/1/commons/catalog-imports
```

Accept: import `status=applied` (may skip rows already seeded).

## WS5 — Negative controls

See [`evidence/phase212/negative-controls.md`](evidence/phase212/negative-controls.md).

Summary: RAG embeddings **336 on A / 0 on B**. Field guide / symptom **row counts match** because both bootstrapped the same migrations — not because of sync.

## Incidents (Tier A / B / C)

| Tier | WS | Symptom | Root cause | Fix / note |
|------|----|---------|------------|------------|
| A | WS1 | Compose API build fail `go >= 1.26.5` | Dockerfile `golang:1.25` + `GOTOOLCHAIN=local` | `ENV GOTOOLCHAIN=auto` in Dockerfile; Install B used hybrid serve |
| A | WS1 | B DB bind `5433 already allocated` | Compose **merges** ports from override | Dedicated [`docker-compose.phase212-b.yml`](../docker-compose.phase212-b.yml) (no merge) |
| B | WS4 | No live catalog sync API | Product design | Documented; SQL + import path |
| B | WS5 | Field guide rows equal on A/B | Migrations/seed on each install | Documented — embeddings are the real boundary |
| C | WS4 | `POST /commons/catalog` 500 on B | Insert error swallowed in handler | Used SQL insert; backlog improve error surfacing |
| B | WS3 | First A sync `skipped_no_receiver` | API started before `.env` ingest URL | Restart API after env change |

## Post-test state

- Scripts committed: `scripts/phase212-install-b-setup.sh`, `phase212-install-b-serve.sh`, `phase212-teardown.sh`
- Seeds: `db/seeds/farm_a_org_assign.sql`, `db/seeds/farm_b_seed.sql`
- **WS7 done (2026-07-30):** Install B clone removed; receiver + `:8081` stopped; Install A `.env` Insert Commons knobs stripped; only Install A DB `:5433` + API `:8080` remain.
- Farm 1 left named **Farm A** under **Organization A** (harmless for next phases; volume wipe + re-seed resets if you want the old demo name).
