---
name: Phase 27 — Farm Guardian AI layer
overview: >
  Wire the generation layer of gr33n's RAG pipeline into a conversational
  farm assistant ("Farm Guardian") powered by Llama 3.1 70B Q4 via Ollama on
  an on-premise server. Introduces a hard AI_ENABLED toggle so the system
  degrades cleanly to Lite (Pi-only, no AI) or runs Full (on-premise server
  room, full LLM inference). Builds directly on the Phase 25 RAG pipeline
  and Phase 26 content boundary/glossary work.
todos:
  - id: ws1-ollama-infra
    content: "WS1: Ollama server setup — install, GPU config, model pull, systemd override, intranet DNS, smoke test (docs/farm-guardian-ollama-setup.md)"
    status: completed
  - id: ws2-ai-toggle
    content: "WS2: AI_ENABLED — env, GET /capabilities, startup LLM reachability when configured; gates rag/answer + strips LLM client when off; POST /v1/chat stub (503/501)"
    status: completed
  - id: ws3-generation-client
    content: "WS3: Go LLM client — SSE ChatCompletionStream + LLM_TIMEOUT_SECONDS + retry/backoff on transient LLM failures (LLM_RETRY_MAX_ATTEMPTS, LLM_RETRY_BACKOFF_MS)"
    status: completed
  - id: ws4-farm-guardian-persona
    content: "WS4: Farm Guardian system prompt — persona + BuildUserMessage + RAG context injection + live farm-state snapshot block (zones / active cycles / unread alerts) on /v1/chat"
    status: completed
  - id: ws5-chat-api
    content: "WS5: Chat API endpoint — POST /v1/chat with optional farm_id RAG injection, streaming SSE, citations, DB-backed conversation_turns + multi-turn replay, GET /v1/chat/sessions[/{id}] history endpoints."
    status: completed
  - id: ws6-ui-chat
    content: "WS6: Operator UI — /capabilities store, Settings Lite/Full label, Knowledge Ask-LLM gating, /chat panel with streaming + citations + persistent session sidebar + multi-turn transcript + inline rename modal + bulk-delete"
    status: completed
isProject: false
---

# Phase 27 — Farm Guardian AI Layer

## Status

**In progress (Phase 27)** — Preconditions met (Phase 25 RAG + Phase 26 boundary/glossary).

### Shipped in-repo (WS2 + WS4 v1 + WS5 v1 + WS6 v1)

- **`AI_ENABLED`** — Parsed in **`internal/ai`**; **unset → on** (backward compatible). Explicit **`false` / `0` / `no` / `off`** → Lite mode (no LLM client wiring for synthesis or chat).
- **`cmd/api` startup** — When AI is on and **`LLM_BASE_URL`** + **`LLM_MODEL`** are set, **`GET {LLM_BASE_URL}/models`** must succeed or the process **exits** (clear failure vs silent degradation).
- **`GET /capabilities`** — Public JSON `{"ai_enabled": bool}` consumed by the UI.
- **`POST /v1/chat`** — JWT-protected non-streaming chat:
  - **AI off** → **503** `AI features are disabled on this installation`.
  - **LLM not configured** → **503** with `set LLM_BASE_URL and LLM_MODEL` hint.
  - **Happy path** → `{ "answer": "...", "llm_model": "..." }` using the Farm Guardian **persona** (`internal/farmguardian`).
- **`POST /farms/{id}/rag/answer`** — Same **503** message when AI off (generation path only; **search** still works if embeddings are configured).
- **`LLM_TIMEOUT_SECONDS`** — Chat HTTP client timeout (default 120s).
- **UI** — `ui/src/stores/capabilities.js` Pinia store auto-loads `/capabilities` at app start; **Settings → AI features** shows a read-only **Lite / Full** label; **Farm knowledge → Ask (LLM)** is disabled with a clear note when AI is off.

### Shipped after the v1 cut (WS1 doc + WS3 stream + WS5 v2/v3 + WS6 chat panel)

- **`docs/farm-guardian-ollama-setup.md`** — Compose + systemd operator runbook for Ollama (WS1, no K8s).
- **`internal/rag/llm`** — `StreamingChatCompleter` interface + `ChatCompletionStream` that parses OpenAI-compatible SSE chunks (`data: {…}\n\n` … `data: [DONE]`), forwards content deltas, surfaces upstream `error.message`, honours `ctx.Done()`.
- **`POST /v1/chat`** —
  - Optional **`farm_id`** triggers JWT farm-membership check + pgvector retrieval (`farmguardian.RAGTopK`) + grounded prompt (persona + `synthesis.SystemPrompt()` + `synthesis.BuildUserMessage`).
  - Optional **`stream: true`** switches the response to `text/event-stream` with `event: delta` / `event: done` / `event: error` blocks ending in `data: [DONE]`.
  - Response (and `done` event) include **`citations`** (`[ref, chunk_id, source_type, source_id, excerpt]`), **`context_count`**, **`embedding_model_id`**, and the echoed **`session_id`** for client correlation.
- **UI** —
  - `/chat` (sidebar: **Guardian**) — single-turn panel with streaming text, citation list, **Use farm context** checkbox (gated by the farm context store), Lite-mode banner driven by the capabilities store.

### Shipped after WS5 v3 (multi-turn history)

- **`db/migrations/20260519_phase27_conversation_turns.sql`** + matching schema entry — `gr33ncore.conversation_turns` (session_id UUID, user_id FK auth.users, optional farm_id, monotonic turn_index, user/assistant messages, llm_model, grounded, context_count, citations JSONB, created_at) with the indexes needed by both lookups.
- **`db/queries/conversation_turns.sql`** + hand-maintained `internal/db/conversation_turns.sql.go` — `InsertConversationTurn`, `ListConversationTurnsBySession`, `ListRecentConversationSessions` (LATERAL joins to keep first/last messages strongly typed instead of `interface{}`).
- **`internal/rag/llm`** — public `Message` type + `ChatCompletionMessages` / `ChatCompletionStreamMessages` for multi-turn; existing single-turn entry points stay as thin wrappers; new `MessagesChatCompleter` and `MessagesStreamingChatCompleter` interfaces.
- **`POST /v1/chat`** now:
  - Validates `session_id` as a UUID (generates a fresh one when omitted).
  - Loads prior turns for `(session_id, user_id)` and replays up to `MaxHistoryTurns = 20` `(user, assistant)` pairs into the prompt **after** the persona / RAG system message and **before** the current user turn.
  - Persists every successful turn (streaming or not) with monotonic `turn_index` via `COALESCE(MAX+1, 0)`.
  - Returns the assigned `turn_index` in the JSON response and the SSE `done` event.
- **`GET /v1/chat/sessions`** — most-recently-active sessions for the caller (cap `MaxRecentSessions = 50`).
- **`GET /v1/chat/sessions/{session_id}`** — full ordered turn history for the caller; returns 400 on bad UUID, scoped by `user_id` so a session_id guess cannot leak another operator's chat.
- **UI `/chat`** — sidebar with persistent session list (active-state highlight, first user message preview, turn count, grounded chip, last-active timestamp, **New** button) + scrollable multi-turn transcript with per-turn citations; replaces the single-turn answer card.

### Shipped after WS5 follow-up (live farm snapshot)

- **`internal/farmguardian/snapshot.go`** — `Snapshot{ZoneCount, ZoneNames, ActiveCycles, UnreadAlerts}` plus `BuildSnapshot(ctx, q, farmID)` (zones + crop cycles + unread-alerts count, best-effort: a failing sub-query never blocks the chat turn) and a prompt-ready `PromptBlock()` that prepends a header so the model knows the snapshot is background context and is not subject to the `[n]` citation rule.
- **`POST /v1/chat` grounded path** — system message now layers as `persona → live farm snapshot → synthesis instructions`. Plain (no `farm_id`) turns are unchanged. Cap of 12 zone names and 8 active cycles in the rendered block keeps the prompt budget predictable for larger farms.

### Shipped after WS4 follow-up (session lifecycle + token usage)

- **`db/migrations/20260520_phase27_session_metadata.sql`** — new `gr33ncore.conversation_sessions` (id UUID PK, user_id FK auth.users, nullable title, created_at/updated_at + `set_updated_at` trigger). Adds `prompt_tokens` + `completion_tokens` columns to `conversation_turns`.
- **Queries** — `UpsertConversationSession` (touch on every turn), `UpdateConversationSessionTitle` (returns row), `DeleteConversationTurnsBySession`, `DeleteConversationSession`. `InsertConversationTurn` extended to record token usage; `ListRecentConversationSessions` LEFT JOINs sessions for title and SUMs token usage across the session.
- **LLM client** — new `llm.Usage{PromptTokens, CompletionTokens, TotalTokens}` and `ChatCompletionMessagesWithUsage`; new `UsageAwareChatCompleter` interface. Old `ChatCompletion` / `ChatCompletionMessages` are thin wrappers that discard usage.
- **`POST /v1/chat`** — captures token usage on the non-streaming path (streaming Usage{} is logged; full streaming usage capture lives behind `stream_options.include_usage` and is deferred). `prompt_tokens` + `completion_tokens` flow into both the JSON response and the persisted row.
- **New endpoints**:
  - `PATCH /v1/chat/sessions/{session_id}` — `{ "title": "…" }`. Whitespace-only / null title clears it (UI falls back to the first user message). 120-rune cap with ellipsis. 404 when the session doesn't belong to the caller.
  - `DELETE /v1/chat/sessions/{session_id}` — removes all turns + the metadata row; idempotent (204 even for non-existent UUIDs so the API doesn't confirm session existence to outsiders).
- **`GET /v1/chat/sessions`** — now returns `title`, `total_prompt_tokens`, `total_completion_tokens` per session.
- **UI `/chat` sidebar** — per-session pencil (✎) and ✕ buttons that appear on hover, wired to the new endpoints. Token totals render as `<n> tok` chips with a prompt/completion tooltip. Transcript turns also show per-turn token chips. Sessions display their title when set, otherwise fall back to the first user message.
- **Smoke harness** — `initMigrations` now applies the Phase 27 migrations (`20260519_phase27_conversation_turns.sql` + `20260520_phase27_session_metadata.sql`) so tests stay self-contained on fresh DBs.

### Shipped after WS5 follow-up (streaming token usage)

- **`internal/rag/llm/chat.go`** — streaming request body now sets `stream_options: {"include_usage": true}` so OpenAI-compatible servers (OpenAI + recent Ollama) emit a terminal SSE chunk with the canonical token-usage block before `data: [DONE]`. New `ChatCompletionStreamMessagesWithUsage` parses any chunk carrying non-zero usage (last-write-wins, matches the OpenAI contract). The legacy `ChatCompletionStreamMessages` is now a thin wrapper that discards usage — back-compat for any caller still on the old signature. New `UsageAwareStreamingChatCompleter` interface exposes the surface.
- **`internal/handler/chat/handler.go`** — streaming path now prefers `UsageAwareStreamingChatCompleter` and falls back to the legacy interface via a local `streamFn` adapter that returns `Usage{}`. Token counts flow into the SSE `done` event payload (`prompt_tokens` + `completion_tokens`) and into the persisted `conversation_turns` row, closing the asymmetry where non-streaming turns recorded usage and streaming turns didn't. Backends that don't honour `include_usage` still work — the row lands with zero tokens, same as before.
- **Tests** — `internal/rag/llm/chat_stream_usage_test.go` covers the contract: terminal usage chunk parsed into the return value, request body asserted to carry `stream_options.include_usage`, backward-compat when the server emits no usage chunk, legacy method still works. `internal/handler/chat/handler_test.go` adds `fakeUsageStreamingLLM` + a happy-path test asserting the `done` SSE event contains `prompt_tokens` / `completion_tokens` and the usage-aware branch was selected.

### Shipped after WS5 follow-up (TTL pruning for conversation history)

- **`db/queries/conversation_turns.sql`** + manually maintained `internal/db/conversation_turns.sql.go` — `DeleteStaleConversationTurns(cutoff)` removes turns from sessions whose `MAX(created_at)` is older than cutoff; `DeleteStaleConversationSessions(cutoff)` drops the matching session metadata rows. Cutoff is computed in Go (`NOW() - TTL`) so the SQL is parameter-only and portable.
- **`internal/farmguardian/prune.go`** — `PruneConfig{TTLDays, Interval, StartupDelay}` + `LoadPruneConfigFromEnv()` reading **`CHAT_SESSION_TTL_DAYS`** (default 30, clamp 0..3650, **0 disables**), **`CHAT_SESSION_PRUNE_INTERVAL_HOURS`** (default 24, clamp 1..168), **`CHAT_SESSION_PRUNE_STARTUP_DELAY_SECONDS`** (default 30, clamp 0..600). `PruneOnce` runs the two DELETEs sequentially (turns first so the visible row count drops before the metadata side), returns `{TurnsDeleted, SessionsDeleted, Cutoff, Duration}` for logging. Turn-pass errors short-circuit so we never orphan dangling metadata. `StartPruneLoop` sleeps the startup delay (ctx-aware), runs the opening prune, then ticks on `Interval` until ctx is done.
- **`cmd/api/main.go`** — spawns `farmguardian.StartPruneLoop(...)` as a goroutine **only when `AI_ENABLED=true`** (no point pruning a feature the operator turned off) and `PruneConfig.Enabled()` (TTL > 0). Logs a one-line summary at boot: `🧹 Chat session prune loop: ttl=30d interval=24h0m0s startup_delay=30s`. Per-prune outcomes log via `slog.Info` only when something was removed — quiet happy path keeps small installs from drowning in 30-day no-op lines.
- **Tests** — `internal/farmguardian/prune_test.go` covers env defaults + clamps + garbage-fallback + `Enabled()` semantics, `PruneOnce` happy path + turn-error short-circuit, `StartPruneLoop` disabled-returns-immediately, ctx-cancel-during-startup-delay, and opening-prune-then-cancel. `cmd/api/smoke_prune_test.go` exercises the real-DB path: seeds a 100-day-old session + a fresh session, runs `PruneOnce` with a 30-day TTL, verifies the stale one is gone (turn + metadata) and the fresh one survives.

### Shipped after WS6 follow-up (inline rename modal + bulk delete)

- **`ui/src/views/FarmGuardianChat.vue`** — replaces the `window.prompt` rename flow with an in-page modal (`<dialog>`-style overlay, role/aria wired, click-outside + Esc to close, autofocus + select on open, max-length 120 with helper copy, empty input clears the title and falls back to the first user message). Submit goes through the form (Enter or the Save button both work) so keyboard-only operators never need the mouse. API errors render **inside** the modal instead of in the page error strip, and the modal stays open so the operator can correct the title.
- **`ui/src/__tests__/chat-rename-modal.test.js`** — mounts the chat panel against a mocked `/capabilities` + `/v1/chat/sessions` seed, then exercises: modal opens pre-filled, `window.prompt` is **never** called, save PATCHes the right body and closes, cancel discards the draft, server errors stay in the modal with the original title intact, empty title clears (UI falls back to the first message).
- **Bulk delete** — `Select` button in the sessions sidebar header flips a **select mode** that swaps per-row ✎/✕ for a checkbox + a top toolbar (`N of M selected · Select all · Cancel · Delete N`). The Delete button opens an aria-wired confirm modal; submitting fans out `Promise.allSettled` DELETEs. Succeeded rows drop out of the sidebar; if the active session was among them the transcript is cleared. Failed rows stay selected so the operator can retry without re-picking, and an inline error reports `Failed to delete N of M`. Cancel exits select mode without firing any DELETE.
- **`ui/src/__tests__/chat-bulk-delete.test.js`** — covers entering select mode (checkboxes appear, ✎/✕ hide), live selection count, full-confirm-flow with the active session deleted (transcript cleared), partial-failure path (modal stays open, only failed row stays selected), Cancel discards selection, Select all picks every row.

### Shipped after WS3 follow-up (retry / backoff)

- **`internal/rag/llm/retry.go`** — `RetryConfig{MaxAttempts, InitialBackoff, MaxBackoff, Sleeper}` + `retryConfigFromEnv()` reading **`LLM_RETRY_MAX_ATTEMPTS`** (default 3, clamped 1..8) and **`LLM_RETRY_BACKOFF_MS`** (default 500ms, clamped 10ms..30s). `IsTransientLLMError` classifies retryable failures: HTTP 408/425/429/5xx (via the new `*HTTPStatusError` carrying status + truncated body), `context.DeadlineExceeded` (per-attempt timeout), `net.Error` / `*url.Error`. Caller `context.Canceled` is **never** retried.
- **Backoff** — exponential `initial * 2^(N-1)` capped at `MaxBackoff` (default 10s) with ±25% jitter so a fleet of restarted Pis doesn't thunder the LLM at the same millisecond.
- **Non-streaming** — `ChatCompletionMessagesWithUsage` retries the full request body each attempt (the body is buffered; each retry is a fresh `bytes.Reader`). Caller sees only the first success or the final error.
- **Streaming connect** — `ChatCompletionStreamMessages` retries only the **connect + status-check** phase (factored into `openStream`). Once the SSE body has yielded any delta to the caller, mid-stream errors fall through directly — replaying after visible content would duplicate text.
- **Tests** — `retry_test.go` covers the classifier (every HTTP code we care about + net error + canceled vs deadline), `retryOp` (transient retried, permanent not retried, max-attempts honoured, ctx cancel during backoff), an `httptest`-backed end-to-end non-streaming "503, 503, 200" round-trip with usage assertions, an end-to-end SSE "503 → 200 + delta + [DONE]" connect-retry round-trip, and env clamp behaviour.

### Shipped after WS5 follow-up (cost guards)

- **`db/queries/conversation_turns.sql`** + `internal/db/conversation_turns.sql.go` — `SumChatTokensSinceForUser(user_id, since)` and `SumChatTokensSinceForFarm(farm_id, since)` each return `{prompt_tokens, completion_tokens, total_tokens}` for the window. Sums are coalesced to `0::bigint` so callers never have to handle SQL NULLs.
- **`internal/farmguardian/cost_guard.go`** — `CostGuardConfig{Window, PerUserMaxTokens, PerFarmMaxTokens}` + `LoadCostGuardConfigFromEnv()` reading **`CHAT_COST_WINDOW_HOURS`** (default 1, clamp 1..168), **`CHAT_COST_MAX_TOKENS_PER_USER`** / **`CHAT_COST_MAX_TOKENS_PER_FARM`** (default 0 = disabled, clamp 0..100_000_000). `CheckBudget` returns a `Decision{Allowed, Reason, UsedTokens, MaxTokens, WindowSeconds, RetryAfter}`; per-user takes precedence over per-farm so a runaway user can't hide behind a quiet farm. Disabled config short-circuits without touching the DB.
- **`internal/handler/chat/handler.go`** — `PostV1` runs `checkCostBudget` immediately after resolving `user_id` + `farm_id` and before any LLM work, so over-budget requests cost nothing. Response is **HTTP 429** with `Retry-After` (in seconds, = window length) and a JSON body `{error, reason: "per_user"|"per_farm", used_tokens, max_tokens, window_seconds, retry_after_seconds}`. Fails open (allows the request) when the SUM query itself errors so a Postgres hiccup never takes chat offline; the failure is logged at WARN.
- **Tests** — `cost_guard_test.go` covers env parsing/clamping and `CheckBudget` against a fake querier (allowed below cap, per-user fires first, per-farm fires when user is fine, farm dimension skipped when `farm_id=0`, error propagation). `cmd/api/smoke_cost_guard_test.go` pins the SQL contract against a real Postgres: rolled-up totals across sessions, per-farm rollup on grounded turns, and confirms the `created_at >= since` clause excludes ancient turns.
- **Docs** — `INSTALL.md` env reference + `.env.example` block + `docs/workstreams/sit-in-operator-experience.md` changelog entry.

### Still open

_None — Phase 27 backend slices are complete. Future ideas (operator dashboard for token usage, alert-channel hook for budget rejections) are sit-in-experience scope and will land via that workstream document._

---

## Goals

1. **Ollama on-premise inference** — Deploy Llama 3.1 70B Q4 via Ollama on the farm's inference server. Expose it on the intranet at a stable DNS alias (`ollama.farm.local` or equivalent) so the Go API never hardcodes an IP.
2. **Hard AI toggle** — A single `AI_ENABLED` env var gates all AI features. When `false`, the system runs in Lite mode: full operational capability (schedules, rules, tasks, alerts, fertigation programs, inventory) with zero LLM dependency. No partial states, no degraded AI — it is either on or off.
3. **Go LLM client** — A thin, reusable HTTP client in the Go API that speaks the OpenAI-compatible chat completions interface Ollama exposes. Allows swapping to a cloud endpoint (OpenAI, Mistral API) by changing two env vars — useful for local development on a laptop without a GPU server.
4. **Farm Guardian persona** — A system prompt that grounds the LLM in gr33n's domain: crop cycles, fertigation schedules, zone/sensor/control terminology, task and alert states. The farmer's conversational counterpart — confident, calm, farm-specific, never generic.
5. **RAG-backed chat endpoint** — A `POST /v1/chat` endpoint that retrieves relevant farm-scoped chunks from pgvector, assembles a grounded prompt, streams the response, and returns source attribution so the operator can verify what Farm Guardian drew on.
6. **Operator chat UI** — A conversation panel in the gr33n frontend. Streaming token display, cited sources below each response, and the AI toggle visible in settings. Farmers who want it off never have to see it.

---

## Two-Mode System

### Lite Mode (`AI_ENABLED=false`)

The full operational system without any LLM dependency. Runs entirely on a Raspberry Pi if needed.

- Schedules, rules, automation engine: fully operational
- Tasks, alerts, inventory: fully operational
- Fertigation programs: operator-configured, no AI suggestions
- Knowledge / Help: static tutorial copy and glossary (Phase 26 WS1)
- No Ollama dependency, no pgvector queries for generation
- Chat UI: hidden or replaced with a static help panel

### Full Mode (`AI_ENABLED=true`)

Requires on-premise server room with Ollama + GPU inference server.

- Everything in Lite, plus:
- Conversational Farm Guardian assistant
- RAG-backed answers grounded in the farm's own data
- Fertigation and schedule suggestions with reasoning
- Inventory and task summaries on request
- Alert context and troubleshooting guidance

There is no middle tier. Operators choose one or the other at deploy time.

---

## Deployment Architecture

```
┌─────────────────────────────────────────────────────┐
│                  Farm Intranet                      │
│                                                     │
│  ┌──────────────┐    ┌──────────────────────────┐   │
│  │ Raspberry Pi │───▶│  API Server (Go)         │   │
│  │   (client)   │    │  gr33n API + site        │   │
│  └──────────────┘    └────────────┬─────────────┘   │
│                                   │                 │
│                          ┌────────┴──────────┐      │
│                          │                   │      │
│               ┌──────────▼───────┐  ┌────────▼────┐ │
│               │  DB Server       │  │  Inference  │ │
│               │  PostgreSQL      │  │  Server     │ │
│               │  pgvector        │  │  Ollama     │ │
│               │  TimescaleDB     │  │  Llama 3.1  │ │
│               └──────────────────┘  │  70B Q4     │ │
│                                     └─────────────┘ │
└─────────────────────────────────────────────────────┘
```

All traffic stays on the farm intranet. No external API calls in production Full mode.

---

## WS1: Ollama Server Setup

### Hardware Minimum (Full Mode)
- GPU: RTX 3090 (24GB VRAM) or equivalent
- RAM: 64GB system RAM
- Storage: 50GB free for model weights
- OS: Ubuntu 22.04 LTS or Debian 12

### Install + Model Pull
```bash
# Install Ollama
curl -fsSL https://ollama.com/install.sh | sh

# Pull the target model
ollama pull llama3.1:70b-instruct-q4_K_M

# Verify GPU offload
ollama run llama3.1:70b-instruct-q4_K_M "ping"
```

### Systemd Service
Ollama ships with a systemd service. Bind it to the intranet interface only:
```ini
# /etc/systemd/system/ollama.service.d/override.conf
[Service]
Environment="OLLAMA_HOST=0.0.0.0:11434"
Environment="OLLAMA_NUM_GPU=1"
```

### Intranet DNS
Register `ollama.farm.local` → inference server IP in the farm's internal DNS or `/etc/hosts` on all servers. The Go API references this alias, never a raw IP.

### Health Endpoint
Ollama exposes `GET /api/tags` — use this for the API's health check dependency. If `AI_ENABLED=true` and Ollama is unreachable at startup, log a fatal error with a clear message rather than silently failing.

---

## WS2: AI Toggle

### Env Vars
```bash
AI_ENABLED=true                          # Master switch
LLM_BASE_URL=http://ollama.farm.local:11434  # Ollama intranet URL
LLM_MODEL=llama3.1:70b-instruct-q4_K_M  # Exact model tag
LLM_TIMEOUT_SECONDS=120                  # Generation timeout
LLM_MAX_CONTEXT_TOKENS=8192             # Budget per request
```

For local development (no GPU server):
```bash
AI_ENABLED=true
LLM_BASE_URL=https://api.openai.com/v1
LLM_MODEL=gpt-4.1-mini
LLM_API_KEY=sk-...
```

Zero code changes between environments — only env vars differ.

### Feature-Flag Middleware
A Go middleware reads `AI_ENABLED` at startup and registers it on the app context. Any handler or service that touches AI gates on this flag:

```go
// config/ai.go
type AIConfig struct {
    Enabled      bool
    BaseURL      string
    Model        string
    TimeoutSecs  int
    MaxTokens    int
    APIKey       string // empty for Ollama (no auth needed on intranet)
}

func AIEnabled(cfg AIConfig) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ctx := context.WithValue(r.Context(), ctxKeyAIEnabled, cfg.Enabled)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

### Degradation Contract
When `AI_ENABLED=false`:
- `POST /v1/chat` returns `HTTP 503` with body `{"error": "AI features are disabled on this installation"}`
- Chat UI hides the conversation panel and shows the static help/glossary instead
- No pgvector embedding queries are issued for generation purposes
- Startup skips Ollama health check entirely

---

## WS3: Go LLM Client

A thin wrapper around the OpenAI-compatible chat completions API. Ollama implements this spec identically to OpenAI, so the same client works for both.

```go
// internal/llm/client.go

type Message struct {
    Role    string `json:"role"`    // "system" | "user" | "assistant"
    Content string `json:"content"`
}

type ChatRequest struct {
    Model    string    `json:"model"`
    Messages []Message `json:"messages"`
    Stream   bool      `json:"stream"`
}

type Client struct {
    baseURL    string
    model      string
    apiKey     string
    httpClient *http.Client
}

func NewClient(cfg AIConfig) *Client {
    return &Client{
        baseURL: cfg.BaseURL,
        model:   cfg.Model,
        apiKey:  cfg.APIKey,
        httpClient: &http.Client{
            Timeout: time.Duration(cfg.TimeoutSecs) * time.Second,
        },
    }
}

func (c *Client) ChatStream(ctx context.Context, messages []Message, onToken func(string)) error {
    // POST to /v1/chat/completions with stream: true
    // Parse SSE chunks, call onToken for each delta
    // Return error on timeout, context cancel, or non-2xx
}
```

### Context Window Budget
Each request must stay within `LLM_MAX_CONTEXT_TOKENS`. The prompt assembly step (WS4/WS5) is responsible for trimming RAG chunks to fit. Never silently truncate mid-chunk — drop the least-relevant chunks first (lowest cosine similarity score).

### Retry Policy (shipped — WS3 follow-up)
- `LLM_RETRY_MAX_ATTEMPTS` (default 3, clamped 1..8) total tries including the first attempt.
- `LLM_RETRY_BACKOFF_MS` (default 500, clamped 10..30000) initial backoff. Exponential doubling up to a 10s cap with ±25% jitter.
- Retryable: HTTP 408/425/429/5xx, per-attempt `context.DeadlineExceeded` (the request-level timeout, not the caller's), `net.Error` / `*url.Error` (DNS, connect, reset, dropped conn).
- Never retried: `context.Canceled` (operator gave up), HTTP 4xx other than the above (bad input, auth — replaying won't help), JSON decode errors, mid-stream SSE failures after the first delta has been forwarded to the caller.

---

## WS4: Farm Guardian System Prompt

The system prompt is the persona contract. It is assembled once at request time from a static template + dynamic farm context.

### Persona Definition
```
You are Farm Guardian, the on-farm intelligence for {farm_name}.
You know this farm's crops, zones, sensors, schedules, fertigation programs,
tasks, and alerts because you have access to its real operational data.

Your role:
- Answer questions about what is happening on the farm right now
- Suggest schedule adjustments, rule changes, and fertigation tweaks
- Summarize tasks and alert states clearly
- Help operators understand why something happened

Constraints:
- Only draw on the farm data provided in the context below
- If the answer is not in the context, say so — do not guess
- Use the glossary terms consistently: setpoint (target value), live reading
  (current sensor value), schedule (time-based trigger), rule (condition-based
  trigger), cycle (named grow period), zone (physical area with sensors/controls)
- Be direct. Farmers are busy. No filler.
- Never mention that you are an LLM or reference your training data
```

### Dynamic Context Block
Appended after the persona, before the conversation history:

```
--- Farm Context (retrieved) ---
{rag_chunks}
--- End Context ---

Current farm snapshot:
- Active zones: {zone_list}
- Open alerts: {alert_summary}
- Current cycle: {cycle_name} (day {day_of_cycle})
- Timestamp: {now}
```

The RAG chunks come from pgvector retrieval (Phase 25 pipeline). The snapshot fields are cheap DB queries injected at request time.

---

## WS5: Chat API Endpoint

### Route
```
POST /v1/chat
Authorization: Bearer {session_token}
Content-Type: application/json

{
  "message": "Why did zone 3 alert this morning?",
  "session_id": "uuid",          // optional — for conversation history
  "stream": true
}
```

### Pipeline
```
1. Auth check
2. Check AI_ENABLED — 503 if false
3. Embed user message → pgvector similarity search → top-K chunks
4. Fetch farm snapshot (zones, alerts, active cycle)
5. Assemble prompt:
     system prompt (persona + glossary)
     + context block (RAG chunks + snapshot)
     + conversation history (last N turns, trimmed to token budget)
     + user message
6. POST to Ollama /v1/chat/completions (streaming)
7. Stream SSE tokens back to client
8. On completion: persist turn to conversation_history table
9. Return source attribution (chunk IDs used in context)
```

### Conversation History Table
```sql
CREATE TABLE conversation_turns (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id  UUID NOT NULL,
    role        TEXT NOT NULL CHECK (role IN ('user', 'assistant')),
    content     TEXT NOT NULL,
    rag_chunk_ids UUID[],         -- which chunks grounded this turn
    model       TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ON conversation_turns (session_id, created_at DESC);
```

History is trimmed to the last N turns that fit within the context window budget before each request.

### Source Attribution Response
```json
{
  "session_id": "uuid",
  "content": "Zone 3 triggered a high-temperature alert at 07:14...",
  "sources": [
    { "chunk_id": "uuid", "table": "automation_events", "summary": "Zone 3 alert 2026-05-09 07:14" },
    { "chunk_id": "uuid", "table": "sensor_readings",   "summary": "Zone 3 temp sensor history" }
  ]
}
```

---

## WS6: Operator Chat UI

### Conversation Panel
- Collapsible side panel or dedicated `/chat` route — operator preference
- Streaming token display (append tokens as SSE events arrive)
- Each assistant message shows a collapsible "Sources" section listing the RAG chunks used
- Conversation history persists per session; new session button clears context
- Typing indicator while generation is in progress

### AI Toggle in Settings
- Settings page exposes the `AI_ENABLED` state as a read-only indicator
- Operators can see whether their installation has AI enabled
- Disabling AI requires a server-side config change (env var) and restart — not a runtime toggle. This is intentional: it prevents accidental mid-session disabling and makes the mode an infrastructure decision, not an operator decision.

### Empty State (AI Off)
When `AI_ENABLED=false`, the chat panel shows:
```
Farm Guardian is not available on this installation.
Your farm is running in Lite mode — all operational
features are fully active.
```
No broken UI, no error state — just an honest, clean message.

---

## Preconditions

- **Phase 25 complete**: RAG ingestion pipeline stable, pgvector populated with farm-scoped domain data, embedding model chosen and deployed
- **Phase 26 WS1 complete**: Glossary finalized — Farm Guardian system prompt references these terms directly
- **Phase 26 WS3 complete**: **`rag-scope-and-threat-model.md` §9** — education vs DB RAG vs ops logs boundary documented
- **Inference server provisioned**: RTX 3090+ box on the farm intranet, Ollama installed, `llama3.1:70b-instruct-q4_K_M` pulled and verified

---

## References

- [Phase 25 — RAG operations and expansion](phase_25_rag_operations_and_expansion.plan.md)
- [Phase 26 — Operator tutorial, observability evolution, RAG education layer](phase_26_operator_tutorial_observability_rag.plan.md)
- [RAG scope and threat model](../rag-scope-and-threat-model.md)
- [Ollama documentation](https://ollama.com/library/llama3.1)

---

*Phase 27 execution: WS2 + WS4 v1 + WS5 v1 + WS6 v1 in code. Streaming, RAG injection, sessions, and Ollama infra doc tracked above.*
