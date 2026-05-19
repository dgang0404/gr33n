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
    content: "WS1: Ollama server setup — install, GPU config, model pull, health endpoint, systemd service, intranet DNS alias"
    status: pending
  - id: ws2-ai-toggle
    content: "WS2: AI_ENABLED — env, GET /capabilities, startup LLM reachability when configured; gates rag/answer + strips LLM client when off; POST /v1/chat stub (503/501)"
    status: completed
  - id: ws3-generation-client
    content: "WS3: Go LLM client — streaming ChatStream + retry; timeout via LLM_TIMEOUT_SECONDS (non-stream path wired)"
    status: pending
  - id: ws4-farm-guardian-persona
    content: "WS4: Farm Guardian system prompt — internal/farmguardian persona + BuildUserMessage v1 (RAG/snapshot injection still pending)"
    status: completed
  - id: ws5-chat-api
    content: "WS5: Chat API endpoint — POST /v1/chat non-streaming answer with persona; AI off/LLM not configured → 503. Session history + RAG injection pending."
    status: completed
  - id: ws6-ui-chat
    content: "WS6: Operator UI — /capabilities Pinia store, Settings Lite/Full label, FarmKnowledge Ask-LLM gating in Lite mode (chat panel still pending)"
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

### Still open

- **WS1** — Ollama infra checklist (operator doc, no K8s).
- **WS3 follow-up** — Streaming `ChatStream` + retry policy (currently non-streaming only).
- **WS4 follow-up** — RAG context injection + live farm snapshot in `BuildUserMessage`.
- **WS5 follow-up** — `session_id` + `conversation_turns` table + source attribution; full streaming SSE response.
- **WS6 follow-up** — Dedicated chat panel/route with streaming token display.

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

### Retry Policy
- Single retry on network error (not on model error or timeout)
- No retry on `context.Canceled` (user navigated away)
- Log all errors with slog at `WARN` level including model, token budget, and duration

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
