---
name: Phase 111 — Guardian model selector & Ollama discovery
overview: >
  Let farm admins pick which local Ollama model Guardian uses (DeepSeek, Llama,
  Mistral, etc.) without touching environment variables. The server queries
  Ollama /api/tags once at startup and caches the list; a farm-scoped preference
  persists in the DB; any operator can override for their own session; every
  farm-level switch is audited. Hard guardrail: Guardian refuses to use a model
  whose reported context window is below 8 192 tokens — the minimum needed for
  a full grounded prompt (live snapshot + RAG context + system prompt).
todos:
  - id: ws0-discovery
    content: "WS0: Ollama discovery — query /api/tags at startup, cache model list in memory; expose GET /guardian/models (server-wide, not farm-scoped — Ollama is a single runtime)"
    status: done
  - id: ws1-schema
    content: "WS1: Schema — ADD guardian_preferred_model TEXT NULL to gr33ncore.farms; migration 20260703_phase111_guardian_preferred_model.sql; update farm_settings struct"
    status: done
  - id: ws2-ui
    content: "WS2: UI — GuardianModelSelector.vue in Guardian panel header; shows model name + context window + speed_class; admin can save farm default; non-admin sees read-only badge"
    status: done
  - id: ws3-chat-param
    content: "WS3: /v1/chat model param — accept optional model field; resolve: session param → farm guardian_preferred_model → env LLM_MODEL; record resolved model in conversation_turns.llm_model"
    status: done
  - id: ws4-fallback
    content: "WS4: Graceful fallback — missing model → log warning + use env fallback, never silent; context_window < 8192 → reject with actionable error suggesting model switch"
    status: done
  - id: ws5-rbac
    content: "WS5: RBAC — PATCH /farms/{id}/settings guardian_preferred_model gated to farmauthz.RequireFarmAdmin; session-level model param open to any farm member"
    status: done
  - id: ws6-audit
    content: "WS6: Audit — farm-level model switch writes user_activity_log row (action_type='guardian_model_changed', details: {from, to, farm_id}); session overrides not audited (visible in conversation_turns.llm_model)"
    status: done
  - id: ws7-openapi
    content: "WS7: OpenAPI — document GET /guardian/models and model param on POST /v1/chat; note farm-scope vs server-scope distinction"
    status: done
  - id: ws8-smokes
    content: "WS8: Smokes — cmd/api/smoke_phase111_model_selector_test.go: discovery returns models, session override, fallback on missing model, RBAC denial on non-admin farm switch, context guardrail"
    status: done
isProject: false
---

# Phase 111 — Guardian model selector & Ollama discovery

## Status

**Shipped** on `main` (commit `d86dfc3`). Depends on Guardian chat (Phases 27–34) and
Ollama at `LLM_BASE_URL`. Embedding model path unchanged (`EMBEDDING_MODEL` env).

**Preconditions (all met on `main`):**
- [`internal/rag/llm/chat.go`](../../internal/rag/llm/chat.go) — `NewChatClientFromEnv`, `ModelLabel()`
- [`internal/handler/chat/handler.go`](../../internal/handler/chat/handler.go) — single `h.llm` client per server
- [`internal/farmauthz/capabilities.go`](../../internal/farmauthz/capabilities.go) — `RequireFarmAdmin`, `RequireFarmOperate`
- `gr33ncore.user_activity_log` + `guardian_tool_executed` audit pattern (Phases 30/34)
- `gr33ncore.farms.meta_data JSONB` — already exists; preferred model stored in dedicated column (not JSON blob) for indexed queries

---

## Why this phase

Guardian's model is locked at process start via `LLM_MODEL` env var — no runtime
discovery, no switching, no operator visibility. A farm running three different crops
might want DeepSeek-R1 (reasoning) for proposal work and Llama3 (fast) for quick
zone questions. Today that requires an SSH session and a server restart.

| Today | After Phase 111 |
|-------|-----------------|
| `LLM_MODEL` env, locked at boot | Farm default + session override, both changeable at runtime |
| No visibility into what's loaded in Ollama | `GET /guardian/models` shows all loaded models with metadata |
| Any model switch = restart | Switch in UI, takes effect next turn |
| No audit trail on model in use | Every farm-level switch audited; model in every `conversation_turns` row |

---

## Scope & design decisions

### Discovery endpoint — server-wide, not farm-scoped

`GET /guardian/models` returns models from the **Ollama runtime**, which is a single
process on the server. There is one Ollama, shared by all farms on this instance —
it is not farm-scoped. The endpoint lists what is **loaded** right now.

```
GET /guardian/models
→ 200 { available_models: [...], server_default: "llama3.1:70b" }
```

No `farm_id` parameter. Any authenticated user can call it (the list is not sensitive
— it's just model names). Farm preference is a separate concept stored in `farms`.

### Farm preference — farm-scoped

```
PATCH /farms/{id}/settings
{ "guardian_preferred_model": "deepseek-r1" }
→ 403 unless farmauthz.RequireFarmAdmin
```

The server validates that the requested model is present in the Ollama cache before
accepting the write. Storing an unloaded model name would cause confusing failures.

### Context window minimum — 8 192 tokens

**Why 8 192?**
A full grounded Guardian prompt contains:
- System prompt + persona: ~800 tokens
- Live farm snapshot (zones, cycles, programs, alerts): ~1 500–2 500 tokens
- RAG context chunks (top-5 at ~200 tokens each): ~1 000 tokens
- Recent conversation history (last 4 turns × ~200 tokens): ~800 tokens
- Operator question: ~200 tokens
- Buffer for response: ~1 500 tokens

**Total: ~5 800–7 300 tokens.** The minimum of **8 192** gives a safe headroom margin
and is the baseline context window of the smallest practical Ollama model (Llama3 8B
default). Any model reporting a smaller window than this is too constrained for
grounded Guardian and will be rejected at request time with:

```
"Model 'phi3:mini' context window (4096) is below the minimum required for
grounded Guardian chat (8192). Switch to a larger model or use non-grounded chat."
```

Models with `context_window: null` (Ollama did not report it) are **allowed with a
warning** — treat unknown as sufficient rather than rejecting valid models that omit
the field.

### Resolution order for `POST /v1/chat`

```
1. Request body: { "model": "deepseek-r1" }   ← session override, any member
2. farms.guardian_preferred_model              ← farm default, admin-set
3. env LLM_MODEL                              ← server default, always present
```

### Fallback on missing model

If the resolved model is not found in the Ollama cache at request time:
1. Log `WARN guardian: model "deepseek-r1" not found in Ollama cache, falling back`
2. Fall back to `env LLM_MODEL`
3. Include `"model_fallback": true` in the chat response JSON so the UI can show
   a one-time toast: *"Model not available — using server default"*
4. **Do not** silently serve a garbled response under the wrong model.

---

## Workstream detail

### WS0 — Ollama model discovery

**Deliverables:**
- `internal/guardian/ollama_discovery.go`
  - `DiscoverModels(ctx, baseURL) ([]ModelInfo, error)` — hits `/api/tags`
  - `ModelInfo` struct: `Name`, `ContextWindow int` (0 if not reported), `ParameterCount int64`, `SpeedClass string`
  - Speed class heuristics: `> 30B params` → `"general"`, `≤ 7B` → `"fast"`, model name contains `r1|reasoning` → `"reasoning"`
- `internal/guardian/model_cache.go` — in-memory cache populated at server init; `Refresh()` callable on demand
- `GET /guardian/models` route in the Guardian/chat router
  - Response: `{ available_models: [ModelInfo...], server_default: string }`
  - Auth: any authenticated operator (no farm scope required)

**Verify:** `curl -H "Authorization: Bearer <token>" /guardian/models` returns the
models currently loaded in Ollama. Logged at startup: `guardian: discovered N models`.

---

### WS1 — Schema

**Deliverable:**
- Migration `db/migrations/20260703_phase111_guardian_preferred_model.sql`:
  ```sql
  ALTER TABLE gr33ncore.farms
    ADD COLUMN IF NOT EXISTS guardian_preferred_model TEXT NULL;
  COMMENT ON COLUMN gr33ncore.farms.guardian_preferred_model IS
    'Farm-default Ollama model for Guardian chat. NULL = use server LLM_MODEL env.';
  ```
- Update `db/schema/gr33n-schema-v2-FINAL.sql` to match
- `sqlc generate` — `Gr33ncoreFarm.GuardianPreferredModel *string`

**Verify:** `SELECT guardian_preferred_model FROM gr33ncore.farms LIMIT 1;` → `NULL`.

---

### WS2 — UI model selector

**Deliverable:** `ui/src/components/GuardianModelSelector.vue`
- Fetches `/guardian/models` when Guardian panel opens
- Displays currently active model (resolved from farm pref or server default)
- Farm admin: dropdown + Save button → `PATCH /farms/{id}/settings`
- Non-admin: read-only chip showing model name
- Shows `context_window` and `speed_class` as sub-text per option
- One-line note: *"Session overrides apply to your chat only and don't change the farm default"*

Include `GuardianModelSelector` in the Guardian panel header
(`ui/src/components/GuardianPanel.vue` or equivalent).

**Verify:** Admin sees dropdown; non-admin sees badge. Selecting a model and saving
reflects immediately on next chat turn.

---

### WS3 — `/v1/chat` model param

**Deliverable:** Extend `POST /v1/chat` request body to accept `"model": string?`.
In `internal/handler/chat/handler.go`:
1. Read `body.Model` (optional)
2. Resolve: `body.Model` → `farm.GuardianPreferredModel` → `env LLM_MODEL`
3. If resolved model ≠ `h.llm.ModelLabel()` and model is in cache → construct a
   per-request `llm.Client` with the override model (reuse `LLM_BASE_URL` + `LLM_API_KEY`)
4. Pass resolved model label into `conversation_turns.llm_model` (already recorded)
5. Include `"model_used"` in streaming/non-streaming response JSON

**Verify:** `POST /v1/chat { "model": "mistral:7b" }` → response `llm_model` field
shows `"mistral:7b"`. Omitting `model` → server/farm default. Farm switch reflected
on next turn without session `model` param.

---

### WS4 — Graceful fallback & context guardrail

**Deliverable:** In the model resolution path:

```go
const GuardianMinContextWindow = 8192

func resolveModel(requested, farmPref, envDefault string, cache *ModelCache) (string, bool fallback) {
    // 1. Validate requested/farmPref against cache
    // 2. Context window check
    // 3. Fall back to envDefault if validation fails
}
```

- Missing model in cache → fallback + warn + set `model_fallback: true` in response
- Context window < `GuardianMinContextWindow` and window is known → reject 400 with
  message pointing operator to a larger model
- Unknown context window (`0`) → allow with `WARN` log, no rejection

**Verify:** Select deleted model → next chat turn gets fallback toast. Select `phi3:mini`
(4 096 window) → 400 with message. Select model with no window reported → succeeds.

---

### WS5 — RBAC

**Deliverable:** `PATCH /farms/{id}/settings` endpoint (create if not yet present,
or extend existing farm settings handler):
- Calls `farmauthz.RequireFarmAdmin(w, r, h.q, farmID)` — 403 for non-admins
- Validates model name is in Ollama cache before writing
- Updates `farms.guardian_preferred_model`

Session-level `model` param on `POST /v1/chat` requires only farm membership
(`RequireFarmMember`) — already enforced by the chat handler.

**Verify:** Non-admin PATCH → 403. Non-admin `/v1/chat` with `model` param → 200
(session only). Admin PATCH → 200, subsequent chat without `model` param uses new farm default.

---

### WS6 — Audit trail

**Deliverable:** On successful `PATCH /farms/{id}/settings` changing
`guardian_preferred_model`, write to `gr33ncore.user_activity_log`:

```json
{
  "action_type": "guardian_model_changed",
  "farm_id": 1,
  "details": {
    "from": "llama3.1:70b",
    "to": "deepseek-r1",
    "changed_by_user_id": "<uuid>"
  }
}
```

Session-level overrides do **not** produce audit rows — they're ephemeral and
already visible in `conversation_turns.llm_model` per turn.

**Verify:** Admin switches farm model → `SELECT * FROM gr33ncore.user_activity_log WHERE action_type = 'guardian_model_changed'` returns one row.

---

### WS7 — OpenAPI

**Deliverable:** In `openapi.yaml`:
- Add `GET /guardian/models` — response schema `GuardianModelsResponse`
- Add `model` (string, optional) to `ChatRequest` body schema
- Add `model_used` and `model_fallback` to `ChatResponse` / streaming chunk schema
- Note in `/guardian/models` description: *"Lists all models currently loaded in the
  Ollama runtime. This is server-wide — not farm-scoped. Farm preference is set via
  PATCH /farms/{id}/settings."*

**Verify:** `make lint-openapi` (or equivalent) passes; new endpoints visible in generated docs.

---

### WS8 — Smoke tests

**Deliverable:** `cmd/api/smoke_phase111_model_selector_test.go`

Tests (all skip if `LLM_BASE_URL` not set — same pattern as other Guardian smokes):
1. `TestPhase111_ModelDiscovery` — `GET /guardian/models` returns `available_models` array with ≥1 entry
2. `TestPhase111_SessionOverride` — `/v1/chat` with explicit `model` param records that model in `conversation_turns.llm_model`
3. `TestPhase111_FallbackOnMissingModel` — farm set to `"nonexistent-model:99b"` → chat response contains `"model_fallback": true`, turn recorded with env default model
4. `TestPhase111_RBACDenial` — non-admin PATCH of `guardian_preferred_model` → 403
5. `TestPhase111_ContextWindowGuardrail` — inject a fake cache entry with `context_window: 512` → chat request with that model → 400 with error message
6. `TestPhase111_AuditOnFarmSwitch` — admin switches farm model → `user_activity_log` row with `guardian_model_changed`

**Verify:** `go test -tags dev ./cmd/api/... -run TestPhase111 -v` passes green.

---

## Acceptance

- [x] `GET /guardian/models` returns all models loaded in Ollama at runtime (server-wide, documented as such)
- [x] `farms.guardian_preferred_model` column exists; admin PATCH persists it; non-admin PATCH → 403
- [x] `/v1/chat` respects session → farm → env resolution order; `llm_model` in `conversation_turns` always reflects the model that actually served the response
- [x] Missing model falls back gracefully with `model_fallback: true` in response — never silent
- [x] Model with known context window < 8 192 rejected with actionable error message (grounded turns)
- [x] Farm-level model switch produces `guardian_model_changed` audit row
- [x] `GuardianModelSelector.vue` in Guardian panel; admin editable, non-admin read-only
- [x] OpenAPI documents `/guardian/models` and `model` param with scope note
- [x] `smoke_phase111_*` green; unit tests for model resolution

---

## Out of scope

- **Model auto-pull** — if the selected model is not in Ollama, Phase 111 fails gracefully; pulling via `ollama pull` is a future phase
- **RAG / embedding model selection** — embedding model stays in `EMBEDDING_MODEL` env; this phase governs the chat/reasoning model only. Conflating the two would break retrieval quality silently
- **Per-session cost tracking** — local Ollama has no token billing; speed class in the UI is the only cost-proxy surfaced
- **Retroactive reprocessing** — switching models mid-session affects new turns only; earlier turns are not replayed under the new model

---

## Implementation order

WS0 (discovery) → WS1 (schema + sqlc) → WS3 (chat param, needs both) → WS4 (fallback, needs WS0 cache) → WS5 (RBAC, needs WS1 column) → WS6 (audit, needs WS5) → WS2 (UI, needs WS0+WS5 API) → WS7 (OpenAPI) → WS8 (smokes, needs all)

WS4 can be built alongside WS3.

---

## Files expected to change

| Area | Files |
|------|-------|
| DB | `db/migrations/20260703_phase111_guardian_preferred_model.sql`, `db/schema/gr33n-schema-v2-FINAL.sql`, `db/queries/farms.sql`, `internal/db/*.go` |
| Discovery | `internal/guardian/ollama_discovery.go`, `internal/guardian/model_cache.go` |
| Handler | `internal/handler/chat/handler.go`, new `internal/handler/guardian/models_handler.go` |
| RBAC / settings | `internal/handler/farm/handler.go` (or new settings sub-handler) |
| Audit | reuse `gr33ncore.user_activity_log` insert pattern from Phase 30 |
| UI | `ui/src/components/GuardianModelSelector.vue`, `ui/src/components/GuardianPanel.vue` |
| Contract | `openapi.yaml` |
| Tests | `cmd/api/smoke_phase111_model_selector_test.go` |

---

## Related

| Doc | Role |
|-----|------|
| [`farm-guardian-architecture.md`](../farm-guardian-architecture.md) | Guardian request flow; model is resolved before prompt assembly |
| [`phase_84_100_master_roadmap.plan.md`](phase_84_100_master_roadmap.plan.md) | Phase 111+ slot — this is the first entry |
| [`internal/rag/llm/chat.go`](../../internal/rag/llm/chat.go) | `NewChatClientFromEnv`, `ModelLabel()` — base to extend |
| [`internal/farmauthz/capabilities.go`](../../internal/farmauthz/capabilities.go) | `RequireFarmAdmin` reuse (WS5) |
