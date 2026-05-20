# Farm Guardian — Architecture & request flow

**Audience:** Developers, operators, and anyone curious about *how* Farm Guardian actually answers a chat message.

**Companion docs:**
- [`farm-guardian-ollama-setup.md`](farm-guardian-ollama-setup.md) — operator install runbook for the inference host (Ollama).
- [`rag-scope-and-threat-model.md`](rag-scope-and-threat-model.md) — privacy & scope boundaries (what RAG indexes, what it doesn't).
- [`plans/phase_27_farm_guardian_ai_layer.md`](plans/phase_27_farm_guardian_ai_layer.md) — the calendar plan that shipped this layer.

---

## 1. The 30-second mental model

Farm Guardian is a conversational AI assistant that runs **entirely on your intranet**. It combines three knowledge sources on every grounded turn:

| Layer | Source | What it gives the model |
|-------|--------|-------------------------|
| **General agronomy** | Llama 3.1 70B Q4 training weights | "What does high humidity do to bud-rot risk?" type knowledge baked in. |
| **Per-farm RAG corpus** | `gr33ncore.farm_knowledge_chunks` (pgvector embeddings) | Operator notes, sensor logs, manuals you've ingested for THIS farm. |
| **Live farm-state snapshot** | DB query at request time (zones, active cycles, unread alerts) | "Right now" context — never stale, never indexed. |

The first layer is universal; the second is private to your farm; the third reflects the database the moment the request fires. The handler combines them into a single system prompt before streaming to Ollama.

---

## 2. End-to-end request flow

What happens when an operator types **"how is my flower room cycle going?"** in the `/chat` UI:

```
1. UI (Vue: FarmGuardianChat.vue)
   ↓ POST /v1/chat { message, farm_id?, session_id?, stream: true }

2. internal/handler/chat (Go) — the orchestrator
   ├── Auth gate         JWT required (route wiring); farm-member if farm_id is set
   ├── Cost guard        Have you blown your token budget? → 429 with Retry-After
   ├── If farm_id set → grounded path:
   │   ├── Embed the question     (OpenAI-compat embedding model)
   │   ├── pgvector kNN search    (over farm_knowledge_chunks)
   │   ├── Build live snapshot    (zones / active cycles / unread alerts)
   │   └── Compose prompt         persona + snapshot + RAG instructions
   ├── Else (no farm_id) → plain path: persona only
   ├── Replay prior turns         from conversation_turns (multi-turn context)
   └── Stream to LLM

3. internal/rag/llm (Go HTTP client → Ollama)
   ├── POST /v1/chat/completions  (OpenAI-compat, stream:true,
   │                               stream_options.include_usage:true)
   ├── Retry on transient errors  (HTTP 5xx, 429, 408, network errors)
   └── Parse SSE chunks → emit text deltas back to the handler

4. Back through the handler
   ├── Stream deltas to UI        as SSE events ("delta", "done")
   ├── On 'done':
   │   ├── Persist the turn       to conversation_turns (with prompt/completion tokens)
   │   └── Update session         conversation_sessions.updated_at → sidebar reorder
   └── Log structured             slog "farm guardian chat streamed" with usage
```

No step in this flow touches the public internet in Full mode. The model lives on the intranet GPU box; the embeddings live in your local Postgres; the snapshot is a local DB query.

---

## 3. The three knowledge layers in detail

### 3.1 Layer 1 — General agronomy (the LLM weights)

Llama 3.1 70B Q4 was trained on the open internet, so it knows the basics of nutrient deficiency symptoms, EC ranges per stage, common pest patterns, IPM, and so on. This layer answers "**what does** high humidity do?" without ever looking at your farm.

You don't manage this layer — it's frozen in the model weights. To upgrade it, swap `LLM_MODEL` to a newer model.

### 3.2 Layer 2 — Your farm's RAG corpus (pgvector)

When the chat request includes `farm_id`, the handler:

1. Sends the user's question to the **embedding model** (OpenAI-compat embeddings endpoint via `internal/rag/embed`).
2. Receives back a vector (typically 1536 floats).
3. Runs a pgvector **nearest-neighbour search** against `gr33ncore.farm_knowledge_chunks` (filtered to `farm_id`), pulling the top-K most semantically similar chunks. Default `K = 6` (`farmguardian.RAGTopK`).
4. Injects those chunks into the prompt with citation markers `[1]`, `[2]`, etc. The system prompt tells the model to cite which chunk supports each claim.

This layer answers "**what did *I* note** about my last flower run?" It only contains content you've ingested via `POST /farms/{id}/rag/ingest`.

What ingest covers vs what it deliberately excludes is documented in [`rag-scope-and-threat-model.md` §9](rag-scope-and-threat-model.md). The short version: cycle notes, sensor narratives, recipe notes, alert resolutions — yes. Operational logs, raw sensor readings, user PII — no.

### 3.3 Layer 3 — Live farm-state snapshot

The RAG corpus is a point-in-time index — yesterday's notes may already be stale. To keep Guardian honest about "right now", every grounded turn pulls a fresh DB snapshot of:

- Total zone count + zone names (capped at 12 names to bound the prompt).
- Active crop cycles (capped at 8) with name, strain, current stage, started_at.
- Count of unread alerts (so Guardian can say "you have 3 open alerts").

This is built by `internal/farmguardian/snapshot.go` → `BuildSnapshot()` → `PromptBlock()`. Failures here never block the chat turn (best-effort, logged at WARN).

A grounded turn's system prompt ends up structured like:

```
{persona}

═══════════ LIVE FARM STATE ═══════════
Zones: Flower Room, Veg Room, Outdoor
Active cycles: 1 (OG Kush — Run 3, late_flower, started 2026-03-01)
Unread alerts: 2
═══════════════════════════════════════

You are answering based on retrieved context. Cite chunks as [n].

[1] {chunk content}
[2] {chunk content}
...
```

The model sees persona → snapshot → RAG instructions in that order, then the conversation history, then the user's question.

---

## 4. The cost guard explained

The **cost guard** is a rolling-window cap on accumulated token usage that prevents runaway LLM compute. Configured via three env vars:

| Env var | Default | Purpose |
|---------|---------|---------|
| `CHAT_COST_WINDOW_HOURS` | `1` | Rolling-window length (clamp 1..168). |
| `CHAT_COST_MAX_TOKENS_PER_USER` | `0` (disabled) | Max tokens per user across all their sessions in the window. |
| `CHAT_COST_MAX_TOKENS_PER_FARM` | `0` (disabled) | Max tokens per farm across all users (only enforced on grounded turns with farm_id). |

**How it works on each `POST /v1/chat`:**

1. Auth resolves the `user_id` (and optional `farm_id` for grounded turns).
2. Cost guard runs **before** any LLM work — sums `prompt_tokens + completion_tokens` from `conversation_turns` where `created_at >= NOW() - window`.
3. If either dimension's total exceeds its cap, return **HTTP 429** with:
   - `Retry-After: <window_seconds>` header.
   - JSON body `{error, reason: "per_user"|"per_farm", used_tokens, max_tokens, window_seconds, retry_after_seconds}`.
4. Per-user dimension takes precedence over per-farm so a single runaway user can't hide behind a quiet farm.

**Key safety properties:**

- **Rejected requests cost zero tokens** — the guard runs before embedding, snapshot, or LLM calls.
- **Fails open on DB errors** — if the SUM query itself errors, the request proceeds (logged at WARN). A transient Postgres outage doesn't take chat offline.
- **Defaults are disabled** — a single-operator home farm doesn't need budget gates. Turn them on for shared / multi-tenant deployments where a buggy script could blow out compute.

**When to set caps:**

- Single farm, single operator → leave both at 0.
- Small team (3–10 staff sharing one Ollama box) → consider `CHAT_COST_MAX_TOKENS_PER_USER=20000` to catch scripted abuse.
- Multi-farm shared deployment → also set `CHAT_COST_MAX_TOKENS_PER_FARM` so one farm's runaway can't starve the others.

**Operator visibility (Phase 28 WS5):** `GET /v1/chat/usage` returns the caller's rolling-window totals + remaining budget; `?farm_id=N` adds per-farm totals (farm-member-gated). The **Settings → Guardian usage** card renders two-tier progress bars and shifts to amber at 80 %, red at 100 %. Crossing 80 % of the per-user cap fires a one-shot `chat_budget_warning` alert into `gr33ncore.alerts_notifications` (debounced once per window) so the warning surfaces through the existing alert channel without operators having to poll the usage endpoint.

The full table of token-related env vars (LLM timeouts, retry, chat history TTL, cost guards) lives in [`INSTALL.md`](../INSTALL.md).

---

## 5. Code map — what lives where

For the developer reader, here's the actual module layout:

### Backend (Go)

| Module / file | Role |
|---------------|------|
| `cmd/api/main.go` | Loads `ai.Config`, verifies LLM reachability on boot, spawns the prune loop goroutine. |
| `cmd/api/routes.go` | Registers `/v1/chat`, `/v1/chat/sessions[/{id}]`, `/capabilities`, RAG endpoints. |
| `internal/ai/config.go` | Parses `AI_ENABLED`, runs the startup LLM reachability check. |
| `internal/handler/chat/handler.go` | The orchestrator — receives `POST /v1/chat`, runs every step in §2 above. |
| `internal/farmguardian/persona.go` | The system prompt that defines Guardian's voice + constraints. |
| `internal/farmguardian/snapshot.go` | Live farm-state snapshot builder (zones / cycles / alerts). |
| `internal/farmguardian/cost_guard.go` | Rolling-window token cap → 429 decision. |
| `internal/farmguardian/prune.go` | Background goroutine that TTL-prunes old chat sessions. |
| `internal/rag/llm/chat.go` | HTTP client to Ollama (streaming SSE, retries, usage capture). |
| `internal/rag/llm/retry.go` | Transient-error classifier + exponential backoff with jitter. |
| `internal/rag/embed/` | OpenAI-compatible embedding client (vectorises the question). |
| `internal/rag/synthesis/` | RAG-grounded prompt builder (citations, system instructions). |
| `internal/db/conversation_turns.sql.go` | Hand-written sqlc bindings for chat history (per-turn insert + listing + pruning + cost sums). |

### Frontend (Vue 3 + Pinia)

| File | Role |
|------|------|
| `ui/src/views/FarmGuardianChat.vue` | The `/chat` panel — session sidebar, multi-turn transcript, streaming, rename/delete/bulk-delete. |
| `ui/src/views/FarmKnowledge.vue` | The `/farm-knowledge` page — RAG search + Ask-LLM (synthesis). |
| `ui/src/stores/capabilities.js` | Loads `/capabilities` at app start; gates AI UI in Lite mode. |
| `ui/src/components/SideNav.vue` | "Guardian" + "Knowledge" entries under Monitor. |

### Database

| Table | Phase | What it stores |
|-------|-------|----------------|
| `gr33ncore.farm_knowledge_documents` | 24 | One row per ingested document (source, kind, last_indexed_at). |
| `gr33ncore.farm_knowledge_chunks` | 24 | Per-chunk content + `embedding vector(1536)` for kNN. |
| `gr33ncore.conversation_sessions` | 27 | Chat session metadata (title, owner, timestamps). |
| `gr33ncore.conversation_turns` | 27 | Each user/assistant exchange (messages, citations, token usage, grounded flag). |

---

## 6. Why this design (vs alternatives)

A few common questions:

**"Why not fine-tune Llama on the farm's data?"**  
RAG is cheap, traceable, and updates instantly when you add a note. Fine-tuning costs hours of GPU time per refresh, loses citation accuracy ("which note led to this answer?"), and requires re-training when the schema or your domain shifts. For per-farm knowledge that changes daily, RAG is the right tool.

**"Why a live snapshot AND a RAG corpus?"**  
The corpus is a *historical index* — what was true when you wrote that cycle note three weeks ago. The snapshot is *right now*. If the model only had the corpus, it'd hallucinate stale facts about which cycles are active.

**"Why does the cost guard default to 0?"**  
The platform assumes a sovereign single-operator install by default. Operators on shared deployments opt into the caps because that's where one bad script can blow up your compute budget. Defaults shouldn't punish the home-farm case.

**"What if Ollama goes down mid-chat?"**  
The retry logic in `internal/rag/llm/retry.go` covers transient blips (5xx, 429, 408, network errors) with exponential backoff. Persistent outages return HTTP 502 from `/v1/chat`; mid-stream errors close the SSE with a `done` event that contains the error. The UI surfaces the error inline and the half-finished turn is **not** persisted to `conversation_turns` (no orphan rows).

**"Can a Guardian reply leak data across farms?"**  
No. The RAG search is hard-filtered by `farm_id` at the SQL level. The snapshot is built from a single `farm_id` query path. The chat persistence is scoped by `user_id` + `session_id` so listing sessions only returns the caller's own.

---

## 7. Phase ledger

The layer was built incrementally across these phases:

- **Phase 24** — RAG retrieval system (embeddings, pgvector, `/rag/search` + `/rag/answer`).
- **Phase 25** — RAG operations & expansion (ingest breadth, incremental re-embed, CI parity).
- **Phase 26** — Operator tutorial, observability, RAG scope/threat-model doc, LLM retry/backoff.
- **Phase 27** — Farm Guardian AI layer (chat endpoint, multi-turn history, snapshot, sessions, streaming, cost guards, `/chat` UI panel). Closed 2026-05-19.
- **Phase 28** — Crop intelligence & Guardian depth. WS3 extends the snapshot with active-cycle analytics; WS4 adds alert detail; WS5 surfaces token-usage to operators. See [`plans/phase_28_crop_intelligence_guardian_depth.md`](plans/phase_28_crop_intelligence_guardian_depth.md).
