---
name: Phase 126 — Guardian CPU efficiency (effective context cap, phi3 4096 budget, clearer errors)
overview: >
  Follow-up to Phases 118 and 122 after laptop validation with phi3:mini: farm-context
  chat is designed to work on a 16 GB CPU-only box, but grounded turns fail with a
  generic "LLM request failed" because (1) Ollama /api/show reports rope-extended
  context (131072) while runtime uses 4096, so Phase 122 prompt trimming never runs;
  (2) RAG embedding and chat models compete for RAM on one Ollama host; (3) stream
  errors are not classified for operators. This phase adds effective-context budgeting
  (including a phi3:mini 4096 override), clearer API/UI errors, and light progress
  hints — without blocking grounded chat on phi3 when prompts fit the real window.
todos:
  - id: ws1-effective-context-resolution
    content: "WS1: Effective context window — parse rope.scaling.original_context_length from /api/show; add EffectiveContextWindow on ModelInfo; built-in override map (phi3:mini→4096, tinyllama→2048); env GUARDIAN_EFFECTIVE_CONTEXT_OVERRIDES"
    status: completed
  - id: ws2-prompt-budget-uses-effective
    content: "WS2: Prompt budget uses effective window — ComputePromptBudget(contextWindow) called with EffectiveContextWindow; grounded gate still uses advertised window for 8192 minimum OR documents phi3 allowed with trim; log effective vs advertised"
    status: completed
  - id: ws3-phi3-4096-override
    content: "WS3: phi3:mini 4096 budget override — harden default override; unit test grounded demo farm prompt trims history/RAG/snapshot when effective=4096; regression: no full-size prompt sent for phi3 on CPU box"
    status: completed
  - id: ws4-classified-llm-errors
    content: "WS4: Classified LLM errors — map timeout/cancel/unreachable/context to error_code + operator_message in SSE error + JSON 502; UI GuardianChatPanel shows specific text (not generic LLM request failed)"
    status: completed
  - id: ws5-stream-progress-hints
    content: "WS5: Stream progress — optional SSE status events (embedding, retrieving, generating) before first delta; selector shows effective_context_window + trim warning for phi3"
    status: completed
  - id: ws6-ollama-contention-docs
    content: "WS6: Ollama contention — document unload embed before chat; optional GET /v1/chat/health field loaded_models; INSTALL.md CPU laptop playbook"
    status: completed
  - id: ws7-tests-smoke
    content: "WS7: Tests — ollama_show effective context parse; prompt_budget effective 4096; handler error classification; optional //go:build ollama smoke grounded phi3 trim"
    status: completed
isProject: false
---

# Phase 126 — Guardian CPU efficiency (effective context cap, phi3 4096 budget, clearer errors)

**Status: shipped** (follow-up to Phase 122, discovered on CPU laptop with phi3:mini + farm context)

**Related:** [Phase 118](phase_118_guardian_model_capabilities.plan.md) (rope quirk documented),
[Phase 122](phase_122_guardian_model_eval_and_context_budget.plan.md) (budget guard shipped),
[INSTALL.md](../../INSTALL.md) § Context budget.

---

## Problem (observed on dgang-laptop17)

| Symptom | Cause |
|---------|--------|
| First ungrounded "hi" works | phi3 loads; small prompt |
| Second turn with **Use farm context** → `LLM request failed` | Full grounded prompt (no trim) + RAG embed + CPU at 100% |
| `ollama list` shows phi3:mini installed | Not a missing model |
| `GET /guardian/models` reports `context_window: 131072` for phi3:mini | Rope max from `/api/show` |
| `ollama ps` shows `CONTEXT 4096` for phi3:mini | **Runtime** window |
| Phase 122 `ComputePromptBudget` only trims when `context_window < 8192` | phi3 advertises 131072 → **no trim** |
| UI shows generic red **LLM request failed** | Handler maps all stream errors to one string |

Farm context on a laptop with phi3:mini **is supported by design** (`INSTALL.md`); the failure mode is **efficiency + observability**, not missing capability.

---

## Goals

1. **Budget prompts against the window the model actually runs with** (effective cap), not rope-extended metadata alone.
2. **Ship a phi3:mini 4096 default override** so grounded demo-farm turns trim history, RAG top-K, and snapshot on CPU boxes.
3. **Replace generic LLM errors** with operator-actionable messages (timeout, unreachable, busy, context too large).
4. **Light UX** so operators know a CPU grounded turn may take minutes and which phase is running.

## Non-goals

- GPU detection / automatic model recommendation engine
- Remote cloud LLM routing
- Replacing Ollama with a different runtime
- Full Phase 122 re-open (eval harness stays as-is)

---

## WS1 — Effective context resolution

**Files:** `internal/farmguardian/ollama_show_pull.go`, `ollama_discovery.go`, `model_cache.go`

### 1a. Parse original rope context from `/api/show`

Extend `parseContextLength` (or add `parseEffectiveContextLength`) to read:

```text
phi3.rope.scaling.original_context_length  → 4096
```

Keep **advertised** `ContextWindow` as today (max `*.context_length` keys) for display and the existing 8192 grounded *gate*.

Add **`EffectiveContextWindow`** on `ModelInfo`:

| Field | Source | Used for |
|-------|--------|----------|
| `context_window` (advertised) | max `*.context_length` | Selector display, 8192 gate |
| `effective_context_window` | min(original rope, advertised) or override | **Prompt budget** |

### 1b. Built-in override map + env

```go
// defaults (overrideable)
"phi3:mini"     → 4096
"phi3:mini:latest" → 4096
"tinyllama"     → 2048
"tinyllama:latest" → 2048
```

Env (optional):

```bash
# comma-separated name=cap pairs; wins over rope parse and builtins
GUARDIAN_EFFECTIVE_CONTEXT_OVERRIDES=phi3:mini=4096,llama3.1:8b=8192
```

Expose both windows on `GET /guardian/models`:

```json
{
  "name": "phi3:mini",
  "context_window": 131072,
  "effective_context_window": 4096,
  "runtime_hint": "loaded, CPU-only — grounded prompts trimmed to 4096 tokens"
}
```

---

## WS2 — Prompt budget uses effective window

**Files:** `internal/handler/chat/completion_repair.go`, `internal/farmguardian/prompt_budget.go`

Change:

```go
contextWindow := h.contextWindowForModel(modelPreview.ModelName)
```

to resolve **effective** window (override → rope original → advertised → 0).

`ComputePromptBudget(effectiveWindow, maxHistoryTurns)` — behavior unchanged, but phi3 at **4096** hits the `< 8192` branch:

- history turns capped (8)
- RAG top-K 5
- snapshot caps reduced

Log both values:

```text
guardian: prompt budget trim model=phi3:mini advertised=131072 effective=4096 detail=...
```

### Grounded gate policy

**Keep** `GuardianMinContextWindow = 8192` on **advertised** window so phi3 remains selectable for grounded chat (INSTALL promise).

**Trim** using **effective** window so prompts fit runtime. Do **not** 400-reject phi3 solely for 4096 effective if overrides say it is an supported trimmed model.

Add unit test: phi3 effective 4096 → trim log non-empty; llama3.1 effective ≥8192 → no trim.

---

## WS3 — phi3:mini 4096 budget override (acceptance focus)

**Acceptance tests:**

1. `TestParseEffectiveContext_phi3` — fixture `model_info` with `phi3.context_length: 131072` and `phi3.rope.scaling.original_context_length: 4096` → effective 4096.
2. `TestComputePromptBudget_phi3Effective4096` — same caps as `< 8192` tier.
3. `TestHandlerGroundedPromptEstimate_phi3` — estimated tokens for demo farm snapshot + RAG after trim ≤ effective window − reserve (coarse `CharsPerTokenEstimate`).

**Manual laptop checklist (document in INSTALL):**

```bash
ollama stop rjmalagon/gte-qwen2-1.5b-instruct-embed-f16   # one model in RAM
# Guardian: Use farm context ON, phi3:mini, demo farm 1
# Expect: trim logs in API, reply within LLM_TIMEOUT_SECONDS (may still be slow on CPU)
```

---

## WS4 — Classified LLM errors

**Files:** `internal/handler/chat/handler.go`, `internal/rag/llm/chat.go`, `ui/src/stores/guardianChat.js`, `GuardianChatPanel.vue`

### API error taxonomy

| `error_code` | When | `operator_message` (example) |
|--------------|------|------------------------------|
| `llm_timeout` | context deadline / client timeout | Local model is still loading or CPU is slow. Wait and retry, or switch to tinyllama. |
| `llm_unreachable` | connection refused / probe fail | Ollama is not reachable at LLM_BASE_URL. |
| `llm_busy` | 503 / queue / runner busy | Ollama is busy with another model (embedding). Retry in a moment. |
| `llm_context` | context length exceeded (if Ollama returns it) | Prompt too large for this model; try without farm context or use a larger model. |
| `llm_failed` | fallback | LLM request failed (today's message) |

SSE:

```text
event: error
data: {"error_code":"llm_timeout","error":"Local model is still loading..."}
```

Non-stream JSON 502 uses same shape.

### UI

- Show `operator_message` in red banner (`data-test="chat-error"`).
- If `error_code=llm_busy`, append hint: `ollama stop <embed-model>` (link to INSTALL).

Do **not** expose raw Go `err.Error()` to operators unless `AUTH_MODE=dev`.

---

## WS5 — Stream progress hints

**Minimal v1** (no new infra):

1. Before RAG embed: optional SSE `event: status` `{"phase":"embedding"}` (only when embedder runs).
2. Before LLM call: `{"phase":"generating"}`.
3. Selector: when `effective_context_window < 8192` and `advertised > effective`, show amber:

   > Grounded prompts trimmed to 4096 tokens (phi3 CPU mode).

**Defer:** top-bar global LLM dot (out of scope unless trivial).

---

## WS6 — Ollama contention + docs

**INSTALL.md** — new subsection *CPU laptop Guardian playbook*:

- phi3:mini + farm context is supported but **slow** (minutes per turn on CPU).
- Run `ollama stop <embed-model>` before long chat sessions.
- Use **tinyllama** for fast smoke; **phi3:mini** for quality.
- Cherry-tree / off-farm horticulture → farm context off.

**Optional:** extend `GET /v1/chat/health` with `ollama_loaded_models[]` from `/api/ps` (read-only).

---

## WS7 — Tests & smoke

| Test | Type |
|------|------|
| `parseEffectiveContextLength` rope + override | unit |
| `ComputePromptBudget` with effective 4096 | unit |
| `classifyLLMError(timeout)` | unit |
| SSE error payload includes `error_code` | handler test |
| GuardianChatPanel renders classified message | vitest |
| `//go:build ollama` grounded phi3 trim smoke (optional, CPU long timeout) | integration |

---

## Implementation order

1. **WS1 + WS2 + WS3** — effective context + phi3 4096 budget (fixes root cause).
2. **WS4** — clearer errors (fixes confusing UI).
3. **WS5 + WS6** — progress + docs (operator experience).
4. **WS7** — tests throughout.

---

## Acceptance (phase done when)

- [ ] `phi3:mini` has `effective_context_window: 4096` on `GET /guardian/models` on the laptop box.
- [ ] Grounded chat with farm context logs prompt budget trims for phi3 (`effective=4096`).
- [ ] Grounded demo-farm turn completes on CPU laptop (may be slow; not `LLM request failed` due to oversized prompt).
- [ ] Stream timeout shows `llm_timeout` message in UI, not generic failure only.
- [ ] INSTALL.md documents CPU playbook + embed unload.
- [ ] Unit tests for effective context parse and phi3 budget trim pass in CI (no ollama tag required).

---

## Out of scope / later

- Auto-unload embed model in API before chat (nice follow-up; document manual step first).
- Per-farm model policy forcing tinyllama on Lite tier.
- Phase 122 eval re-run gate in CI for phi3 quality scores.
