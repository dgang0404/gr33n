---
name: Phase 130 — Guardian runtime orchestration (chat path hardening)
overview: >
  Phase 129 makes awakening visible and login-and-go. Phase 130 fixes what happens
  during each chat turn on CPU laptops — auto embed unload before grounded LLM,
  grounded timeout override, early SSE status through snapshot/embed/read-tools,
  single-flight busy guard, eval harness timeouts, and morning-walkthrough fixture.
  Without 130, operators still risk 777s failures even after a green awakening badge.
todos:
  - id: ws1-grounded-timeout
    content: "WS1: GUARDIAN_GROUNDED_TIMEOUT_SECONDS env — grounded streams use max(LLM_TIMEOUT, grounded, 1500); document in environment-variables.md"
    status: completed
  - id: ws2-embed-unload-on-send
    content: "WS2: Before grounded LLM call — if embed loaded and chat model cold/contended, Ollama unload embed (Phase 126 deferred); log guardian: embed unloaded for chat"
    status: completed
  - id: ws3-early-sse-phases
    content: "WS3: Refactor POST /v1/chat stream — flush SSE early; status phases: preparing, snapshot, read_tools, embedding, generating; UI guardianChat shows phase line"
    status: completed
  - id: ws4-chat-busy-lock
    content: "WS4: Single-flight grounded chat per farm (or global on laptop) — health awakening.state=busy; 429 llm_busy with clear message if second send"
    status: completed
  - id: ws5-eval-timeout
    content: "WS5: cmd/guardian-eval HTTP client uses max(120, LLM_TIMEOUT_SECONDS) or GUARDIAN_EVAL_TIMEOUT_SECONDS; morning walkthrough fixture in eval set"
    status: completed
  - id: ws6-stale-ollama-detect
    content: "WS6: Health flag stale_ollama_cli — detect orphan ollama run / high CPU with no ps models; message in awakening"
    status: completed
  - id: ws7-auto-warm-on-send
    content: "WS7: If farm_counsel send while sleeping — inline mini-warmup before build prompt (fallback when user skips awakening panel)"
    status: completed
  - id: ws8-tests-smoke
    content: "WS8: Handler tests embed unload decision, grounded timeout, busy lock; optional //go:build ollama smoke grounded morning walkthrough"
    status: completed
isProject: false
---

# Phase 130 — Guardian runtime orchestration (chat path hardening)

**Status:** **Shipped.** · Depends on [Phase 129](phase_129_guardian_awakening.plan.md) WS0 health fields

**Closes:** Phase 126 "Out of scope / later" — auto-unload embed before chat

---

## Why a separate phase

Phase 129 is **operator-visible** (modes, awakening, tune script, badges).

Phase 130 is **request-path** surgery:

- Today all grounding (snapshot, read tools, RAG embed) runs **synchronously before** the SSE stream opens — operators see nothing for minutes, then timeout at Ollama headers.
- Warmup (129) reduces cold-start; **130** ensures each heavy turn survives CPU RAM contention and timeout budgets.

Split keeps 129 shippable as UX + warmup API while 130 refactors `handler.go` stream lifecycle.

---

## WS1 — Grounded timeout

New env: `GUARDIAN_GROUNDED_TIMEOUT_SECONDS` (default: `max(1500, LLM_TIMEOUT_SECONDS)` when unset on CPU profile).

Apply only to grounded `POST /v1/chat` stream/non-stream clients in `internal/rag/llm`.

**Why separate from `LLM_TIMEOUT_SECONDS`:** ungrounded phi3 cherry-tree chat can stay at 777s fast-fail; farm counsel needs longer prefill on CPU.

Tune script (129 WS3) sets both on laptop profile.

---

## WS2 — Embed unload on send (Phase 126 deferred)

Before `retrieveChunks` + LLM on grounded turns:

1. `GET /api/ps` — if embed model loaded and chat model not (or both loaded on CPU with `size_vram=0`)
2. `POST /api/generate` unload embed or Ollama stop API
3. Log: `guardian: embed unloaded for chat`

Guard: skip on GPU server profile when chat already loaded.

Reuse contention heuristic from 129 `embed_blocks_chat`.

---

## WS3 — Early SSE phases

Refactor `Handler.PostChat` when `stream=true`:

1. Open SSE immediately after auth/farm checks
2. Emit status phases as work completes:

| phase | When | UI message |
|-------|------|------------|
| `preparing` | Start | "Preparing counsel…" |
| `snapshot` | After BuildSnapshot | "Reading live farm…" |
| `read_tools` | After EnrichPromptBlock | "Checking alerts and devices…" |
| `embedding` | Before retrieveChunks | "Searching field memories…" |
| `generating` | Before LLM stream | "Composing answer…" |

`guardianChat.js` already handles `event: status` — extend `streamingStatus` to show phase-specific copy.

**Risk:** larger handler refactor — feature-flag `GUARDIAN_EARLY_SSE=1` for rollout.

---

## WS4 — Chat busy lock

Laptop profile: one grounded stream at a time (in-memory mutex per process).

- Second concurrent POST → `429` + `error_code: chat_busy`
- Health `awakening.state=busy` while stream open
- UI: disable Send + show "Guardian is answering…"

Prevents stacked 777s × retry attempts.

---

## WS5 — Eval harness alignment

`internal/farmguardian/eval/runner.go`:

- HTTP client timeout: `GUARDIAN_EVAL_TIMEOUT_SECONDS` or `max(120, LLM_TIMEOUT_SECONDS)`
- Add fixture: `farm-morning-walkthrough` — "What should I check first on a morning walkthrough of this farm today?" — expect `walk_farm` enrichment in logs (optional log scrape) or answer mentions alerts/devices

Unblocks Phase 128 WS4 without false failures at 120s.

---

## WS6 — Stale Ollama CLI detect

Health field `stale_ollama_hint`:

- `pgrep -f 'ollama run'` while `/api/ps` empty or stuck
- Message: "Close stray terminal ollama run sessions"

Playbook advanced fallback — surfaced in awakening failed state.

---

## WS7 — Auto-warm on send

If user sends Farm counsel while `awakening.state=sleeping`:

- Inline warmup (same as POST /guardian/warmup) with 60s cap
- Then proceed with prompt build
- SSE status: `awakening` phase

Bridges users who skip the awakening panel.

---

## WS8 — Tests & smoke

| Test | Coverage |
|------|----------|
| Grounded timeout > ungrounded | unit |
| embed unload decision | unit with mock ps |
| chat_busy 429 | handler test |
| Early SSE phase order | handler integration |
| eval client timeout env | unit |
| Morning walkthrough fixture | eval score smoke |

---

## Acceptance

1. Morning walkthrough on laptop completes after 129 tune + 130 timeout (no manual `ollama stop`).
2. Logs show `embed unloaded for chat` when embed was loaded at send time.
3. UI shows phase line through snapshot → embedding → generating.
4. Second send while first streaming → busy message, not parallel timeout.
5. `make guardian-eval MODEL=phi3:mini` does not fail all questions at 120s.

---

## Implementation order

1. WS1 + WS2 (timeout + embed unload) — highest impact, smaller diff
2. WS4 + WS7 (busy + auto-warm on send)
3. WS3 (early SSE — largest refactor)
4. WS5 + WS6 + WS8

---

## Relationship to Phase 129

| Concern | Phase |
|---------|-------|
| "Is Guardian awake?" | 129 |
| "Why did my question timeout?" | 130 |
| Mode cards / druid copy | 129 |
| In-turn phase messages | 130 |
| `make guardian-laptop-tune` | 129 |
| `GUARDIAN_GROUNDED_TIMEOUT_SECONDS` | 130 (+ tune script sets it) |

**Recommended ship:** 129 WS3+WS0+WS2 together with 130 WS1+WS2 as one laptop fix PR; UI workstreams 129 WS4–8 can follow.
