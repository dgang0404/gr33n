# Phase 212 WS5 — Negative controls

Evidence that knowledge does **not** live-sync between Install A and Install B.
Both installs ran their own schema/migrations/seed — catalog tables can match without any cross-install traffic.

| Check | Install A | Install B | Verdict |
|---|---|---|---|
| `rag_embedding_chunks` total | 336 | 0 | **PASS** — embeddings are local ingest only |
| platform RAG chunks | 67 | 0 | **PASS** |
| farm_id=1 RAG chunks | 336 | 0 | **PASS** |
| Field guide rows (`agronomy_field_guides`) | 77 | 77 | **Expected** — same migrations/seed on each DB, not live sync |
| Symptom entries | 10 | 10 | **Expected** — migration/seed catalog on each DB |
| Hand-carried Commons pack `phase212-handcarry-jadam-indoor-starter-v1` | 0 | 1 | **PASS** — only after WS4 manual copy |
| Insert Commons receiver `distinct_pseudonyms` | — | — | **2** (see `receiver-stats.json`) |

## Interpretation

- **RAG embeddings / operational field memories** do not appear on B until B runs its own ingest — this is the real install boundary.
- **Field guide / symptom catalog rows** ship with schema/migrations on every install; they are not "synced" from A. B never received A's embeddings or chat corpus.
- Install B runs `AI_ENABLED=false`, so `/v1/chat/health` awakening corpus is empty even if rows exist — another local boundary.

## API surface

- A: `chat-health-a.json` — field_guide / platform / operational corpus present
- B: `chat-health-b.json` — awakening null / empty (AI_ENABLED=false)
