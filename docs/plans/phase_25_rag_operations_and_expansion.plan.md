---
name: Phase 25 RAG operations and expansion
overview: >
  Harden and broaden the Phase 24 RAG slice: operational readiness (CI, environments,
  secrets), product-finalized ingestion domains beyond tasks + automation runs,
  incremental or scheduled re-embedding, observability and limits, and UX/docs polish.
  Scope and naming are intentionally provisional—refine in follow-up planning prompts
  before execution.
todos:
  - id: ws1-product-and-scope-lock
    content: "WS1: Lock v1 embedding checklist (per-domain yes/no) in rag-scope-and-threat-model.md §6; align with Phase 21 crop-cycle surfaces where relevant"
    status: pending
  - id: ws2-ingestion-breadth
    content: "WS2: Extend rag-ingest + document builders for agreed domains (e.g. crop cycles, programs, costs rollups)—sanitizers per domain"
    status: pending
  - id: ws3-incremental-reembed
    content: "WS3: Incremental re-embed strategy (cursors/outbox/triggers—not only manual CLI); backfill and idempotency verified at scale"
    status: pending
  - id: ws4-ops-and-ci
    content: "WS4: CI/staging/prod parity for pgvector; smoke or integration path; operator runbook snippet (env, migrate, ingest)"
    status: pending
  - id: ws5-quality-obs-limits
    content: "WS5: Integration tests with mocked embedding/LLM; optional metrics; synthesis rate limits (per-farm or per-user if needed); log/error hygiene"
    status: pending
  - id: ws6-ux-docs
    content: "WS6: UI empty/degraded states (503), optional nav polish; README/roadmap checkbox; cross-links from workflow guide; optional regenerated schema diagrams dated to migrations"
    status: pending
isProject: false
---

# Phase 25 — RAG operations and expansion

## Status

**Planning.** This document is the hand-off for the next implementation pass after [Phase 24 — RAG retrieval system](phase_24_rag_retrieval_system.plan.md). Todos above are **pending** until refined and agreed.

## Preconditions

- [Phase 24](phase_24_rag_retrieval_system.plan.md) shipped: pgvector storage, ingest worker (current: **tasks** + **automation runs**), search + optional answer API, minimal Knowledge UI, smoke tests, threat model in **[docs/rag-scope-and-threat-model.md](../rag-scope-and-threat-model.md)**.
- **[Phase 21 crop-cycle analytics](phase_21_crop_cycle_analytics.plan.md)** (recommended): richer cycle text/summaries improve retrieval quality when those surfaces are stable.

## Goals (draft)

- Turn the RAG stack from a **vertical slice** into something **operators can rely on** in real environments (migrations, extensions, env, rerun/repair).
- **Expand ingestion** only where product + §4.2 checklist explicitly allow—each new domain gets sanitization rules and tests.
- Reduce **manual-only** ingestion over time via **incremental or scheduled** updates (exact mechanism TBD).

## Non-goals (draft)

- Pi / `X-API-Key` RAG API unless explicitly designed (same as Phase 24 default).
- Training or fine-tuning models on tenant data.

## Open questions (fill in during planning prompts)

1. **Priority order** among domains in §4.2 (crop cycles, programs, costs, inventory, alerts, …)—which ship in Phase 25 vs later?
2. **Incremental model:** polling cursors (`after_id`), logical replication, NOTIFY, or application outbox—what fits gr33n’s ops and dev velocity?
3. **CI depth:** pgvector job only, or full integration tests with mocked HTTP for embeddings/chat?
4. **Synthesis limits:** keep global only, or add per-farm / per-user quotas first?
5. **Legacy / alternate schemas** (e.g. historical ERDs): in scope only if they map to **current** `gr33ncore` / `gr33nfertigation` tables—confirm with schema owners.

## Work-stream detail (stub)

### WS1 — Product and scope lock

Finalize **yes/no per domain** for v1 in **[docs/rag-scope-and-threat-model.md](../rag-scope-and-threat-model.md)** §6 (checklist), consistent with §4.2 sensitivity and Phase 21 alignment.

### WS2 — Ingestion breadth

Extend **`cmd/rag-ingest`** / **`internal/rag/ingest`** for approved domains; reuse **`internal/rag/sanitize`** patterns; add sqlc queries as needed per source table.

### WS3 — Incremental re-embed

Design and implement **something better than full manual backfill only** for agreed tables (cursor flags, cron contract, or worker hook—TBD).

### WS4 — Ops and CI

Ensure **pgvector** and migrations run in **CI/staging/prod** the same way; document **minimum env** for a healthy Knowledge tab; optional short **operator** section (link from INSTALL or workflow guide).

### WS5 — Quality, observability, limits

Stronger **automated tests** around farm isolation and handler behavior; **metrics** optional; tighten **rate limits** and production **error/logging** behavior as needed.

### WS6 — UX and docs

503 / missing-key messaging in UI; nav/README/roadmap updates so Phase 25 exit is visible; optional regenerated schema diagrams dated to migrations; **Pi / deployment topology** narrative lives in **`docs/raspberry-pi-and-deployment-topology.md`** (iterate with hardware reality).

---

*Next step:* Edit the **open questions** and **WS*** stubs in follow-up prompts, then flip individual `todos` to `in_progress` / `completed` as work lands (same convention as Phase 24).
