# Offline / intranet deployment (private LAN or VLAN)

This page is a **starting sketch** for running gr33n **without using the public internet** for database, dashboard, or **RAG** (embeddings + optional chat synthesis). Treat “offline” here as **no WAN dependency for AI services**: components still talk to each other over **HTTP on your network** (including loopback). Update this doc once your **offline-mode farm** topology is fixed so reality matches the checklist.

**Related:** [rag-scope-and-threat-model.md](rag-scope-and-threat-model.md) (trust boundaries), [INSTALL.md](../INSTALL.md) (`EMBEDDING_*`, `LLM_*`), [example-env.md](example-env.md).

## Rough layout on a private network

| Piece | Where it can live |
|--------|-------------------|
| Postgres (vectors + relational data) | LAN host / Pi / NUC |
| gr33n API + UI | Same machine or another host on the LAN |
| Embedding service (OpenAI-compatible HTTP, `/v1/embeddings`) | `http://192.168.x.x:port` or `http://127.0.0.1:...` (Ollama, LM Studio, vLLM, etc.) |
| Chat / answer service for RAG synthesis | Same idea — set `LLM_BASE_URL` to another LAN or loopback endpoint |

Point **`DATABASE_URL`** at whichever host runs Postgres; point **`EMBEDDING_BASE_URL`** / **`LLM_BASE_URL`** (and optional keys/models per [INSTALL.md](../INSTALL.md)) at the **embedding** and **chat** servers you run internally. Ingestion (`cmd/rag-ingest`) uses the same embedding settings as the API.

## VLAN / air-gap notes

- **Isolated VLAN:** ensure API and browser-reachable UI can still reach Postgres and your model servers per firewall rules; DNS on the segment should resolve internal names if you use hostnames instead of IPs.
- **True air-gap (no external routes):** run Postgres, API, UI, and model servers on hosts that only see each other; use loopback where everything is co-located on one box.
- **First-time setup:** pulling container images or OS packages may need a one-time connected step unless you mirror artifacts internally—document your org’s process here when applicable.

---

*Changelog: stub added 2026-04-21 — extend with your farm’s IPs, hostnames, and ops runbook.*
