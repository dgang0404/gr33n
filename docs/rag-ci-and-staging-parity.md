# RAG / pgvector — CI, staging, and production parity

Phase **25 WS4** ensures the **same Postgres capabilities** used for **Knowledge (RAG)** in development—**pgvector**, plus the relational stack the schema expects—are present in **CI**, **staging**, and **production**, and that **migrations** apply cleanly everywhere.

## What CI does

GitHub Actions workflow **`.github/workflows/ci.yml`**:

1. Builds and starts **`docker compose` `db`** from **`db/Dockerfile`** (TimescaleDB **PG16** + **pgvector** + PostGIS — same image families as local dev).
2. Runs **`./scripts/bootstrap-local.sh --seed`** against  
   `postgres://gr33n:gr33n@127.0.0.1:5433/gr33n?sslmode=disable` (host port from **`docker-compose.yml`**).
3. Asserts the **`vector`** extension exists and **`gr33ncore.rag_embedding_chunks`** is present.
4. Runs **`go test -tags dev ./...`** (includes **`internal/handler/rag`** integration tests when `DATABASE_URL` is set — mocked embedding/LLM, real pgvector).
5. Runs **`ui`** Vitest in a separate job (`npm ci` + `npm test`).

If CI is green, migrations and pgvector-backed tables are not silently drifting from what operators run locally via Compose.

## Staging / production checklist

Use the **same migration ordering** as bootstrap: monolithic schema file (if fresh DB) plus **`db/migrations/*.sql`** in lexical order — see **`scripts/bootstrap-local.sh`** and **[INSTALL.md](../INSTALL.md)**.

| Requirement | Notes |
|-------------|--------|
| **PostgreSQL** | Version aligned with repo expectations (see **`db/Dockerfile`** / INSTALL). |
| **Extensions** | **`timescaledb`**, **`postgis`**, **`vector`** — Knowledge/RAG requires **`vector`**; API startup registers pgvector types against the pool. |
| **`DATABASE_URL`** | Single URL for API, **`rag-ingest`**, and **`make check-stack`** / **`scripts/check-local-stack.sh`**. |
| **Secrets for ingest / Knowledge** | **`EMBEDDING_API_KEY`**, optional **`EMBEDDING_*`**, **`LLM_*`** for Ask — see **[`.env.example`](../.env.example)** and **[workflow-guide §10.6](workflow-guide.md#106-farm-knowledge-rag-retrieval)**. |

**Hosted Postgres:** enable the **vector** extension (package name varies by provider; see [pgvector install](https://github.com/pgvector/pgvector#installation)). If the provider does not offer pgvector, use the repo **`docker-compose`** **`db`** image or another pgvector-capable image — do not assume “Postgres only” is enough for RAG.

## Operator commands (after DB is up)

```bash
# From repo root; DATABASE_URL matches your environment
export DATABASE_URL='postgres://…'

# Sanity-check extension + optional API health (see docs/local-operator-bootstrap.md)
make check-stack

# Load schema + migrations (+ optional seed) — same script CI uses
./scripts/bootstrap-local.sh --seed   # or omit --seed on empty prod DB

# Farm-scoped embedding ingest (requires embedding provider env)
make rag-ingest-help
go run ./cmd/rag-ingest -farm-id 1 -tasks …   # see workflow-guide §10.6
```

## Related docs

- Threat model + storage shape: **[rag-scope-and-threat-model.md](rag-scope-and-threat-model.md)**
- Full operator bootstrap: **[local-operator-bootstrap.md](local-operator-bootstrap.md)**
- Phase plan: **[plans/phase_25_rag_operations_and_expansion.plan.md](plans/phase_25_rag_operations_and_expansion.plan.md)**
