---
name: Phase 24 RAG Retrieval System
overview: >
  Build a real retrieval layer on top of Phase 20.95 RAG-prep schema: embed
  operator-relevant records, store vectors, expose a farm-scoped query API,
  and optional LLM answer synthesis with explicit consent boundaries. Starts
  after Phase 23 exit criteria and (recommended) Phase 21 crop-cycle analytics
  so summary/compare endpoints exist before ingestion prioritization.
todos:
  - id: ws1-scope-and-threat-model
    content: "WS1: Document data classes to embed (tasks, costs, automation_runs, crop cycles, etc.), farm isolation, and what must never leave the farm without opt-in"
    status: completed
  - id: ws2-storage
    content: "WS2: Choose + migrate vector storage (e.g. pgvector extension + column(s), or external store); idempotent migrations"
    status: completed
  - id: ws3-ingestion-pipeline
    content: "WS3: Batch or incremental embedding jobs from Postgres → vectors; dedupe keys; backfill strategy"
    status: completed
  - id: ws4-retrieval-api
    content: "WS4: Authenticated POST/GET retrieval endpoint(s); hybrid filter (farm_id, module, date) + vector search"
    status: completed
  - id: ws5-optional-llm-layer
    content: "WS5: Optional — pluggable LLM (LM Studio / local) for synthesis; strict prompt + cite sources; rate limits"
    status: completed
  - id: ws6-ui-and-smoke
    content: "WS6: Minimal UI entry (e.g. Settings or Operate drawer) + smoke tests + OpenAPI + workflow-guide glossary"
    status: completed
isProject: false
---

# Phase 24 — RAG retrieval system

## Relationship to Phase 20.95

**Phase 20.95** added **RAG-prep** columns and housekeeping so *future* retrieval queries have stable joins. Phase **24** is the first phase that actually ships **embeddings + retrieval** (and optionally **generation**). Nothing here replaces human operators; it **surfaces** what is already in the database.

## Preconditions

- **[Phase 23 stabilization](phase_23_stabilization_sprint.plan.md)** exit criteria satisfied.
- **[Phase 21 crop-cycle analytics](phase_21_crop_cycle_analytics.plan.md)** shipped (recommended); coordinate if deferred.
- Clear **product decision** on which objects get embedded first (suggest: crop cycles + cost lines + automation runs + task titles, iterate).

## Non-goals (initial cut)

- Training a foundation model on customer data.
- Sending farm payloads to third-party clouds **without** explicit configuration and consent aligned with Insert Commons / audit patterns.

## Work-stream detail

### WS1 — Scope and threat model (**done**)

Deliverable: **[docs/rag-scope-and-threat-model.md](../rag-scope-and-threat-model.md)** — actors (JWT vs Pi vs dev bypass), farm isolation rules, sensitivity tiers, candidate tables/columns for embeddings, exclusions (secrets/JSON), egress/LLM/Insert Commons boundaries, **product checklist (§6)** for v1 ingestion by domain, and **§8 hand-offs** to WS3–WS6.

### WS2 — Vector storage (**done**)

**Decision:** Postgres **pgvector**, table `gr33ncore.rag_embedding_chunks` (`vector(1536)`), HNSW + btree indexes — see **[docs/rag-scope-and-threat-model.md](../rag-scope-and-threat-model.md)** §7.

Artifacts: `db/migrations/20260518_phase24_rag_pgvector.sql`, schema mirror in `db/schema/gr33n-schema-v2-FINAL.sql`, Docker `db/Dockerfile` + `docker-compose.yml` build for local dev, [INSTALL.md](../../INSTALL.md) §2c for bare-metal pgvector.

### WS3 — Ingestion pipeline (**done**)

- **sqlc:** `db/queries/rag.sql` (upsert, delete by source/type, count, nearest-neighbor search for WS4), `ListAutomationRunsByFarmAfterID` in `db/queries/automation.sql`; `sqlc.yaml` maps `vector` → `pgvector.Vector`.
- **pgx:** `internal/pgxutil.RegisterVectorTypes` hooks `pgvector-go/pgx` into the pool (`cmd/api` and `cmd/rag-ingest`).
- **Sanitize:** `internal/rag/sanitize` — plain-note trimming + `AutomationDetailsJSON` (drops sensitive JSON keys / URL-like strings).
- **Embed:** `internal/rag/embed` — OpenAI-compatible `/v1/embeddings` via `EMBEDDING_API_KEY`, optional `EMBEDDING_BASE_URL`, `EMBEDDING_MODEL`, `EMBEDDING_DIMENSION`.
- **Worker:** `internal/rag/ingest` — task + automation_run document builders and `Worker.IngestFarmTasks` / `IngestFarmAutomationRuns`.
- **CLI:** `cmd/rag-ingest` — `-farm-id`, `-tasks`, `-automation-runs`, `-dry-run`, cursor flags for runs.

### WS4 — Retrieval API (**done**)

- **Routes:** `GET` and `POST /farms/{id}/rag/search` — JWT + `farmauthz.RequireFarmMember`; farm id from path only.
- **Handler:** `internal/handler/rag/handler.go` — embeds the user query via the same OpenAI-compatible client as WS3 (`EMBEDDING_*` env); calls `SearchRagNearestNeighborsFiltered` (`farm_id`, optional `metadata.module`, `created_at` range, limit).
- **sqlc:** `SearchRagNearestNeighborsFiltered` in `db/queries/rag.sql`.
- **OpenAPI:** `openapi.yaml` — `RagSearchRequest`, `RagSearchResponse`, `RagSearchResult`, tag `rag`. Run `make audit-openapi` after route changes.

### WS5 — Optional LLM synthesis (**done**)

- **Route:** `POST /farms/{id}/rag/answer` — same auth and farm scope as WS4; requires **embedding** (`EMBEDDING_*`) plus **`LLM_BASE_URL`** + **`LLM_MODEL`** (optional **`LLM_API_KEY`** for providers that need it; omit for many local gateways).
- **Flow:** filtered vector retrieval (same as search, `max_context_chunks` default **8**, max **15**) → numbered context blocks → OpenAI-compatible **`/v1/chat/completions`** → parse bracket citations **`[n]`** → response includes **`citations`** with chunk id / source / excerpt.
- **Rate limit:** global per-process **`RAG_SYNTHESIS_MAX_PER_MINUTE`** (default **30**).
- **Code:** `internal/rag/llm/chat.go`, `internal/rag/synthesis/synthesis.go`, `internal/handler/rag/limiter.go`, `internal/handler/rag/handler.go` (`Answer`).
- **OpenAPI:** `RagAnswerRequest`, `RagAnswerResponse`, `RagCitation`. Env tuning: `LLM_TEMPERATURE`, `LLM_MAX_TOKENS`.

### WS6 — UI, smoke, glossary (**done**)

- **UI:** **Monitor → Knowledge** (`/farm-knowledge`) — `ui/src/views/FarmKnowledge.vue`: query + optional filters, **Search chunks** (`POST /rag/search`), **Ask (LLM)** (`POST /rag/answer`, extended timeout). Router + `SideNav` + mobile drawer (`App.vue`).
- **Smoke:** `cmd/api/smoke_rag_test.go` — unauthorized → **401**; with JWT and no embedding/LLM env in the smoke process → **503** (documents degraded config).
- **OpenAPI:** Already covered in WS4/WS5; `make audit-openapi` unchanged from WS6.
- **Workflow guide:** §10.6 Farm knowledge (RAG); glossary rows **RAG**, **Embedding chunk**, **pgvector**, **Knowledge (UI)**; link to [`rag-scope-and-threat-model.md`](../rag-scope-and-threat-model.md) in §12.

