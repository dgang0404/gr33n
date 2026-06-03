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

## Field assistant mode (Phase 37)

When the grow site has **no WAN** (or you choose not to use cloud LLMs), point inference at **loopback or a private LAN** address:

| Variable | Typical field value | Purpose |
|----------|---------------------|---------|
| `LLM_BASE_URL` | `http://127.0.0.1:11434/v1` | Ollama on the same NUC/Pi as the API |
| `LLM_MODEL` | e.g. `llama3.2` | Chat model name |
| `EMBEDDING_BASE_URL` | same host as LLM | Local embeddings for RAG |
| `EMBEDDING_API_KEY` | any non-empty string | Required by the embed client |
| `GR33N_REPO_ROOT` | path to gr33n checkout | Loads `docs/field-guides/procedures/*.yaml` |

**Health:** `GET /v1/chat/health?farm_id=1` — `field_assistant.field_mode`, `llm_reachable`, chunk counts.

**Graceful degrade:** if the local LLM is down, field install chat still works via **guided procedures** (`start procedure wire-pi-relay-light`, `done`, `list procedures`) and **static print** (`GET /v1/field-guides/procedures/{id}/print`).

```bash
make rag-ingest-platform-docs
make rag-ingest-field-guides
```

**Single-box smoke:** Postgres + API + UI + Ollama on one host; stop Ollama to verify procedure degrade in chat.

---

*Changelog: stub 2026-04-21 — Phase 37 field assistant 2026-06-03.*
