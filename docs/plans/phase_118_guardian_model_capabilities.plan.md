---
name: Phase 118 — Guardian model capabilities (chat-only filtering, name normalization, runtime hints)
overview: >
  Follow-on polish for Phases 111–112 found while running the E2E suite against a
  real Ollama box: the selector offers embedding-only models as chat candidates;
  model-name tag normalization ("tinyllama" vs "tinyllama:latest") lets the env
  default and the fallback path bypass the context-window guardrail; and the UI
  gives no hint that a model will be slow on CPU-only hardware.
todos:
  - id: ws1-capabilities
    content: "WS1: Capabilities filter — /api/show already returns capabilities[]; keep only models with 'completion' in the chat selector/cache resolution; expose capabilities in ModelInfo + GET /guardian/models; vision capability tagged for future vision routing"
    status: completed
  - id: ws2-name-normalization
    content: "WS2: Name normalization — cache lookups treat 'name' and 'name:latest' as the same model; ResolveChatModel fallback paths (unknown model, env default) must re-check the context guardrail after normalization; unit tests for both spellings"
    status: completed
  - id: ws3-guardrail-fallback
    content: "WS3: Guardrail on fallback — the 'unknown model → env default' branches in ResolveChatModel return without the grounded context check; route them through the same try() gate; smoke: grounded chat with env default tinyllama (ctx 2048) → 400, ungrounded → 200"
    status: completed
  - id: ws4-runtime-hints
    content: "WS4: Runtime hints — query /api/ps at discovery refresh; surface loaded-vs-cold and CPU/GPU placement per model in the selector ('loaded, CPU-only — expect slow replies'); no hard blocks, advisory only"
    status: completed
  - id: ws5-ui-polish
    content: "WS5: UI — hide embedding models; badge capabilities (chat/vision); cold-model warning before first message; keep admin pull row"
    status: completed
isProject: false
---

# Phase 118 — Guardian model capabilities (chat-only filtering, name normalization, runtime hints)

## Status

**Shipped** (2026-07-03). Chat-only model filtering, `:latest` normalization with
guardrail on env-default paths, `/api/ps` runtime hints, and selector UI polish.

---

## Findings

### 1. Embedding models offered for chat

On the verification box, 3 of 6 installed models are embedding-only
(`nomic-embed-text`, `qwen3-embedding:0.6b`, `gte-qwen2-…-embed`). Discovery lists
them all; the selector offers them as chat candidates; picking one yields a confusing
runtime error instead of never being offered. Ollama's `/api/show` (which Phase 112
WS0 already calls) returns a `capabilities` array — `["completion"]`,
`["embedding"]`, etc. We fetch the payload today and discard this field.

### 2. Tag normalization bypasses the guardrail

`LLM_MODEL=tinyllama` but Ollama reports `tinyllama:latest`. `ModelCache.Get("tinyllama")`
misses, so `ResolveChatModel` falls through to the bare-install branches
(`model_cache.go` lines ~146–152) which return the env default **without** the
grounded context-window check. Observed live: grounded chat ran on tinyllama
(context 2048 < 8192 minimum) and returned 200 where an explicit
`"model": "tinyllama:latest"` request correctly gets 400.

### 3. No runtime placement/latency hints

`speed_class` is a parameter-count guess. Real signal exists in `/api/ps`: whether the
model is loaded and whether it sits on CPU or GPU. On the CPU-only verification box a
grounded tinyllama turn took ~45 s — an operator picking a 70B model there gets no
warning at all.

Related quirk (document, don't fix): `phi3:mini` reports
`phi3.context_length: 131072` alongside `rope.scaling.original_context_length: 4096`;
`parseContextLength` takes the max, which is the correct read of extended-rope models.

---

## Design notes

- **WS1:** filter at cache-set time but keep the raw list available (an `all=true`
  query param on `GET /guardian/models` for debugging). Vision-capable models get
  `capabilities: ["completion","vision"]` — groundwork for routing `LLM_VISION_MODEL`
  through the same cache later.
- **WS2:** normalize with a single helper (`strings.TrimSuffix(name, ":latest")` on
  both sides of the comparison); do not rewrite stored audit values — normalization is
  lookup-only, `conversation_turns.llm_model` keeps whatever the client ran.
- **WS3:** this is the actual bug fix; WS2 alone would mask it on this box but not on
  a box whose env default is genuinely absent from Ollama.

### Out of scope

- Automatic model recommendations ("your box should use X") — advisory text only
- VRAM/RAM fit prediction
- Vision model selection UI (groundwork only)

---

## Acceptance

- [x] Embedding-only models absent from `GET /guardian/models` default response and the selector
- [x] `LLM_MODEL=tinyllama` with Ollama reporting `tinyllama:latest`: grounded chat → 400 context reject; ungrounded → 200
- [x] Explicit and env-default paths produce identical guardrail outcomes for the same effective model
- [x] Selector shows loaded/cold + CPU/GPU hint when Ollama exposes it; absent gracefully otherwise
- [x] Unit tests cover normalization matrix (bare/`:latest`/custom tag × session/farm/env sources)
- [x] `TestPhase112_*` still green; new `TestPhase118_*` ollama-tagged smokes for WS1–WS3

---

## Files expected to change

| Area | Files |
|------|-------|
| Discovery | `internal/farmguardian/ollama_discovery.go`, `ollama_show_pull.go` (capabilities), new `/api/ps` fetch |
| Cache/resolve | `internal/farmguardian/model_cache.go` (+ tests) |
| Handler | `internal/handler/chat/models.go` (`all=true`, capabilities in payload) |
| UI | `ui/src/components/GuardianModelSelector.vue` |
| Docs | `openapi.yaml`, INSTALL.md ollama section |
| Tests | `cmd/api/smoke_phase118_ollama_test.go` (`//go:build ollama`) |
