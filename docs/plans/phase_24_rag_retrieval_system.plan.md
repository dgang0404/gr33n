---
name: Phase 24 RAG Retrieval System
overview: >
  Build a real retrieval layer on top of Phase 20.95 RAG-prep schema: embed
  operator-relevant records, store vectors, expose a farm-scoped query API,
  and optional LLM answer synthesis with explicit consent boundaries. Starts
  only after Phase 23 stabilization exit criteria are met.
todos:
  - id: ws1-scope-and-threat-model
    content: "WS1: Document data classes to embed (tasks, costs, automation_runs, crop cycles, etc.), farm isolation, and what must never leave the farm without opt-in"
    status: pending
  - id: ws2-storage
    content: "WS2: Choose + migrate vector storage (e.g. pgvector extension + column(s), or external store); idempotent migrations"
    status: pending
  - id: ws3-ingestion-pipeline
    content: "WS3: Batch or incremental embedding jobs from Postgres → vectors; dedupe keys; backfill strategy"
    status: pending
  - id: ws4-retrieval-api
    content: "WS4: Authenticated POST/GET retrieval endpoint(s); hybrid filter (farm_id, module, date) + vector search"
    status: pending
  - id: ws5-optional-llm-layer
    content: "WS5: Optional — pluggable LLM (LM Studio / local) for synthesis; strict prompt + cite sources; rate limits"
    status: pending
  - id: ws6-ui-and-smoke
    content: "WS6: Minimal UI entry (e.g. Settings or Operate drawer) + smoke tests + OpenAPI + workflow-guide glossary"
    status: pending
isProject: false
---

# Phase 24 — RAG retrieval system

## Relationship to Phase 20.95

**Phase 20.95** added **RAG-prep** columns and housekeeping so *future* retrieval queries have stable joins. Phase **24** is the first phase that actually ships **embeddings + retrieval** (and optionally **generation**). Nothing here replaces human operators; it **surfaces** what is already in the database.

## Preconditions

- **[Phase 23 stabilization](phase_23_stabilization_sprint.plan.md)** exit criteria satisfied.
- Clear **product decision** on which objects get embedded first (suggest: crop cycles + cost lines + automation runs + task titles, iterate).

## Non-goals (initial cut)

- Training a foundation model on customer data.
- Sending farm payloads to third-party clouds **without** explicit configuration and consent aligned with Insert Commons / audit patterns.

## Work-stream detail (high level)

Details will be filled in during kickoff after Phase 23; the todos above track the spine.
