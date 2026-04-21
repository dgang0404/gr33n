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
    status: completed
  - id: ws2-ingestion-breadth
    content: "WS2: Extend rag-ingest + document builders for agreed domains (e.g. crop cycles, programs, costs rollups)—sanitizers per domain"
    status: completed
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
    content: "WS6: UI empty/degraded states (503), optional nav polish; README/roadmap checkbox; cross-links from workflow guide; schema ERD text doc (schema-erd-text.md) refreshed when graph changes"
    status: pending
isProject: false
---

# Phase 25 — RAG operations and expansion

## Status

**Planning — open questions below are answered** for this pass. **WS1 (engineering)** is satisfied by **[docs/rag-scope-and-threat-model.md](../rag-scope-and-threat-model.md) §6** — engineering-defaults checklist and implementation order (product may still reprioritize later). This document is the hand-off after [Phase 24 — RAG retrieval system](phase_24_rag_retrieval_system.plan.md). Todos flip to `in_progress` / `completed` as work lands.

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

## Decisions (agreed before implementation)

Answers below align with **[docs/rag-scope-and-threat-model.md](../rag-scope-and-threat-model.md)** §4.2 / §6 defaults and Phase 25 WS1–WS6.

### 1. Domain priority — Phase 25 vs later

**Already ingested end of Phase 24:** tasks (`gr33ncore.tasks`), automation runs (`automation_runs` with scrubbed `details`).

**Phase 25 ingestion expansion (in order):**

1. **Crop cycles** (`gr33nfertigation.crop_cycles`) — highest operator value next; align text with Phase 21 surfaces where summaries exist; sanitize sensitive note fields per §4.2.
2. **Fertigation programs** (`programs` name/description; **metadata** only via an explicit allowlist) — ship in Phase 25 if allowlist + tests land in the same window; otherwise first follow-up release.
3. **Explicitly later (not Phase 25 exit criteria):** task labor notes (PII), raw line-level costs / inventory unit costs, high-volume alerts, automation rules/schedules until `action_parameters` and JSON paths are scrubbed — per §6 “Later” defaults unless product reprioritizes.

**WS1 output:** §6 now carries engineering defaults and order (see threat-model doc). Optional later: product sign-off column if stakeholders want an explicit Yes/No/Later audit trail.

### 2. Incremental re-embed model

**Chosen:** **Polling with cursors** — watermark on `updated_at` and/or monotonic `(source_type, source_id)` progress — implemented inside **`rag-ingest`** (CLI flags or env) and/or a **cron-friendly** periodic run. Same process remains idempotent via existing `(farm_id, source_type, source_id, chunk_index)` upsert.

**Deferred:** PostgreSQL **NOTIFY** (nice for lower latency; adds long-lived listeners), **logical replication** (heavy ops), and a dedicated **outbox** table — only reconsider if cron + cursor proves insufficient or multiple consumers need a queue.

### 3. CI depth

**Both**, staged:

1. **Parity baseline (WS4):** CI applies migrations against a **pgvector-capable** Postgres image so extension + `rag_embedding_chunks` never silently drift from prod.
2. **Phase 25 tests (WS5):** **Integration-style tests with mocked HTTP** for embedding (and optional chat) clients — exercise handlers + DB + farm isolation **without** calling external APIs; keep or extend existing smoke patterns where appropriate.

Pure “pgvector job only” is insufficient as an exit criterion if handlers are untested against a real DB shape.

### 4. Synthesis rate limits

**Phase 25:** Keep the **global** `RAG_SYNTHESIS_MAX_PER_MINUTE` (and related env) as the primary control — one knob for operators.

**If we add granularity in this phase:** Prefer **per-farm** ceilings next (fairness on shared deployments; aligns with tenant boundaries). **Per-user** quotas are heavier (JWT identity plumbing, UX) — only if product requires abuse mitigation beyond farm-level.

### 5. Legacy / alternate schemas

**Rule:** Implementation and ingestion **only** follow **`db/schema/gr33n-schema-v2-FINAL.sql`** + **`db/migrations/`** (see **[docs/database-schema-overview.md](../database-schema-overview.md)**). Old informal ERDs are not authoritative. A **new** diagram is fine as **documentation** if regenerated from current SQL or a DB built from those files — optional under **WS6** (date it and cite the migration or baseline commit).

---

## Work-stream detail (stub)

### WS1 — Product and scope lock

**Done (engineering baseline):** **[docs/rag-scope-and-threat-model.md](../rag-scope-and-threat-model.md) §6** lists engineering defaults per domain and suggested ingest order; **Decisions §1** (“Domain priority”) above matches it. Phase 21 alignment applies when implementing **crop cycles** ingestion (WS2). Optional follow-up: formal product sign-off if stakeholders want an explicit audit trail beyond engineering defaults.

### WS2 — Ingestion breadth

Extend **`cmd/rag-ingest`** / **`internal/rag/ingest`** for approved domains; reuse **`internal/rag/sanitize`** patterns; add sqlc queries as needed per source table.

**Shipped:** **crop cycles**, **programs**, **schedules / rules / executable actions**, **cost transactions** (no money in text), **input definitions & batches** (`-inventory-definitions`, `-inventory-batches`; no unit cost / qty numerics), **alerts** (`-alerts`; `ListAlertsByFarmAfterID`). WS2 ingestion breadth checklist is effectively **complete** unless product asks for §6 tweaks.

### WS3 — Incremental re-embed

Design and implement **something better than full manual backfill only** for agreed tables (cursor flags, cron contract, or worker hook—TBD).

### WS4 — Ops and CI

Ensure **pgvector** and migrations run in **CI/staging/prod** the same way; document **minimum env** for a healthy Knowledge tab; optional short **operator** section (link from INSTALL or workflow guide).

### WS5 — Quality, observability, limits

Stronger **automated tests** around farm isolation and handler behavior; **metrics** optional; tighten **rate limits** and production **error/logging** behavior as needed.

### WS6 — UX and docs

503 / missing-key messaging in UI; nav/README/roadmap updates so Phase 25 exit is visible; **schema diagram:** [`docs/schema-erd-text.md`](../schema-erd-text.md) (ASCII + optional Mermaid from current SQL — refresh date/migrations note when the graph shifts); **Pi / deployment topology** narrative lives in **`docs/raspberry-pi-and-deployment-topology.md`** (iterate with hardware reality).

---

*Next step:* Execute WS1–WS6 using **Decisions** above; flip individual `todos` to `in_progress` / `completed` as work lands (same convention as Phase 24).
