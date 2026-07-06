---
name: Phase 135 — Guardian RAG lifecycle (freshness, re-ingest, Settings corpus)
overview: >
  Field memories show age and staleness; operators trigger LAN-only re-ingest from
  Settings without scripts. Health and awakening surfaces corpus state. Closes the
  gap between live snapshot and stale operational RAG.
todos:
  - id: ws1-ingest-metadata
    content: "WS1: Track last_ingested_at per farm+source_type — migration or aggregate max(rag_embedding_chunks.updated_at); expose in /v1/chat/health"
    status: pending
  - id: ws2-freshness-rules
    content: "WS2: rag_corpus_ok + staleness tiers — fresh <24h, aging <7d, stale >7d for operational; field guides stale on manifest hash change"
    status: pending
  - id: ws3-reingest-job
    content: "WS3: POST /farms/{id}/guardian/reingest {scope: field_guides|platform|operational} — async job, 202 + poll status (LAN embed only)"
    status: pending
  - id: ws4-settings-corpus-card
    content: "WS4: Settings Guardian — corpus table, last run, Re-ingest buttons (admin), link guardian-bootstrap-farm for first-time"
    status: pending
  - id: ws5-awakening-warn
    content: "WS5: Farm counsel mode + awakening panel amber when operational stale or field_guide_chunks=0"
    status: pending
  - id: ws6-cron-doc
    content: "WS6: Enterprise README cron example; local-operator-bootstrap one-liner after seed"
    status: pending
  - id: ws7-tests
    content: "WS7: health freshness fields; reingest job mock; vitest Settings corpus card"
    status: pending
isProject: false
---

# Phase 135 — Guardian RAG lifecycle

**Status:** planned · **Depends on:** [129](phase_129_guardian_awakening.plan.md) WS0/WS8

**Scripts reused:** `rag-ingest-field-guides.sh`, `rag-ingest-platform-docs.sh`, `rag-ingest-farm-operational.sh`, `guardian-bootstrap-farm.sh`

---

## Problem

Operators don't know if RAG is empty or weeks old. Bootstrap is terminal-only. Farm counsel promises "field memories" without freshness honesty.

---

## WS1 — Freshness metadata

Health `awakening` / `field_assistant` extension:

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

---

## WS3 — Re-ingest API (v1)

- **Auth:** farm_manager+ on farm
- **Scope:** one of `field_guides`, `platform_docs`, `operational`, `all`
- **Implementation:** spawn goroutine running existing shell scripts or Go ingest packages (prefer in-process Go for API deployability long-term; v1 may `exec` scripts with timeout)
- **Guard:** reject if embed unreachable; set job status `running|done|failed`
- `GET /farms/{id}/guardian/reingest/status`

---

## WS4 — Settings UI

| Corpus | Count | Last ingested | Action |
|--------|-------|---------------|--------|
| Field guides | 58 | 2d ago | Re-ingest |
| Platform docs | 12 | 5d ago | Re-ingest |
| Operational | 240 | **21d ago** | Re-ingest (amber) |

Progress bar while job running (poll status).

---

## Acceptance

- [ ] Fresh seed + no ingest → Farm counsel warns "field memories not loaded"
- [ ] Re-ingest field guides from Settings increases chunk count without terminal
- [ ] Health shows `operational_last_ingested_at` after `rag-ingest-farm-operational`

---

## Non-goals

- Automatic cron in API process (document external cron only)
- Cross-farm corpus sharing
- WAN cloud embed for re-ingest
