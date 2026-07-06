---
name: Phase 135 — Guardian RAG lifecycle (freshness, re-ingest, Settings corpus)
overview: >
  Field memories show age and staleness; operators trigger LAN-only re-ingest from
  Settings without scripts. Health and awakening surfaces corpus state. Closes the
  gap between live snapshot and stale operational RAG.
todos:
  - id: ws1-ingest-metadata
    content: "WS1: Track last_ingested_at per farm+source_type — migration or aggregate max(rag_embedding_chunks.updated_at); expose in /v1/chat/health"
    status: completed
  - id: ws2-freshness-rules
    content: "WS2: rag_corpus_ok + staleness tiers — fresh <24h, aging <7d, stale >7d for operational; field guides stale on manifest hash change"
    status: completed
  - id: ws3-reingest-job
    content: "WS3: POST /farms/{id}/guardian/reingest {scope: field_guides|platform|operational} — async job, 202 + poll status (LAN embed only)"
    status: completed
  - id: ws4-settings-corpus-card
    content: "WS4: Settings Guardian — corpus table, last run, Re-ingest buttons (admin), link guardian-bootstrap-farm for first-time"
    status: completed
  - id: ws5-awakening-warn
    content: "WS5: Farm counsel mode + awakening panel amber when operational stale or field_guide_chunks=0"
    status: completed
  - id: ws6-cron-doc
    content: "WS6: Enterprise README cron example; local-operator-bootstrap one-liner after seed"
    status: completed
  - id: ws7-tests
    content: "WS7: health freshness fields; reingest job mock; vitest Settings corpus card"
    status: completed
isProject: false
---

# Phase 135 — Guardian RAG lifecycle

**Status:** shipped · **Depends on:** [129](phase_129_guardian_awakening.plan.md) WS0/WS8

**Scripts reused:** `rag-ingest-field-guides.sh`, `rag-ingest-platform-docs.sh`, `rag-ingest-farm-operational.sh`, `guardian-bootstrap-farm.sh`

---

## Problem

Operators don't know if RAG is empty or weeks old. Bootstrap is terminal-only. Farm counsel promises "field memories" without freshness honesty.

---

## WS1 — Freshness metadata

Health `awakening.corpus` extension:

```json
{
  "corpus": {
    "field_guide_chunks": 58,
    "field_guide_last_ingested_at": "2026-07-05T10:00:00Z",
    "platform_doc_chunks": 12,
    "platform_last_ingested_at": "2026-07-01T08:00:00Z",
    "operational_chunks": 240,
    "operational_last_ingested_at": "2026-06-20T12:00:00Z",
    "staleness": "operational_stale"
  }
}
```

Aggregates: `GetRagCorpusStatsByFarm` (max `updated_at` per tier).

---

## WS3 — Re-ingest API (v1)

- **Auth:** farm admin on farm
- **Scope:** `field_guides`, `platform_docs`, `operational`, `all`
- **Implementation:** in-process `internal/rag/reingest` goroutine (same ingest worker as CLI)
- **Guard:** reject if embed unreachable or non-LAN `EMBEDDING_BASE_URL`
- `POST /farms/{id}/guardian/reingest` → 202 + job
- `GET /farms/{id}/guardian/reingest/status`

---

## WS4 — Settings UI

**Settings → Field memories (RAG corpus)** — table with Re-ingest buttons (admin), progress while job runs.

---

## Verify

```bash
go test ./internal/farmguardian/... -run Corpus -count=1
go test ./internal/handler/guardian/... -count=1
cd ui && npm test -- --run src/__tests__/guardian-settings-corpus.test.js
# Live: GET /v1/chat/health?farm_id=1&mode=farm_counsel → awakening.corpus
```

---

## Acceptance

- [x] Fresh seed + no ingest → Farm counsel warns "field memories not loaded"
- [x] Re-ingest field guides from Settings (API) without terminal
- [x] Health shows `operational_last_ingested_at` after operational ingest

---

## Non-goals

- Automatic cron in API process (document external cron only)
- Cross-farm corpus sharing
- WAN cloud embed for re-ingest
- Field-guide manifest hash staleness (time-based tiers only in v1)
