---
name: Phase 122 — Guardian model eval harness, context budget guard, and proposal repair
overview: >
  Three model-agnostic reliability improvements found while assessing whether
  phi3:mini is "good enough": (1) nothing measures answer quality or proposal
  validity per model today — the runtime hint from Phase 118 tells you a model
  is cold/CPU-bound but not whether it's actually good at this job; (2) the
  grounded prompt (system + snapshot + RAG chunks + read-tool blocks + history)
  is built the same size regardless of which model is resolved, so a small model
  loaded later with a smaller real context (tinyllama: 2048) gets no protection;
  (3) there is no repair path when a small model returns malformed JSON for an
  action proposal — it just fails the turn. All three help every model in the
  selector, not just phi3:mini.
todos:
  - id: ws1-eval-harness
    content: "WS1: Eval harness — cmd/guardian-eval CLI (or `//go:build ollama` smoke suite) runs a fixed set of ~20 grounded questions (drawn from docs/field-guides fixtures + seeded farm) against every installed chat-capable model; scores citation presence, groundedness (no invented zone/sensor names), and latency; writes a markdown/JSON report"
    status: pending
  - id: ws2-context-budget
    content: "WS2: Context budget guard — resolve the chat model's real context window before building the prompt (not the grounded-chat minimum gate); trim RAG chunk count / history turns / snapshot verbosity to fit a token budget derived from ContextWindow when it is well below GuardianMinContextWindow headroom; log when trimming occurs"
    status: pending
  - id: ws3-proposal-repair
    content: "WS3: Proposal JSON repair — when the model's response fails ActionProposal schema parsing, retry once with a short corrective system message showing the exact expected shape and the parse error, before failing the turn; cap at one retry; log repair attempts + outcome per model for the eval harness to pick up"
    status: pending
  - id: ws4-selector-quality-badge
    content: "WS4: Selector quality badge — surface WS1 report results in GET /guardian/models and GuardianModelSelector.vue ('tested: 18/20 grounded, 1 proposal repair avg' or 'not yet evaluated' for freshly pulled models); advisory only, same spirit as Phase 118 runtime hints"
    status: pending
  - id: ws5-docs
    content: "WS5: Docs — INSTALL.md / operator guide section on reading eval scores when choosing a model for a low-power Pi-adjacent box; note phi3:mini's rope-extended context quirk from Phase 118 findings and why the budget guard uses real window, not the reported max"
    status: pending
isProject: false
---

# Phase 122 — Guardian model eval harness, context budget guard, and proposal repair

## Why (the question this phase answers)

"Is phi3:mini actually good enough, and will this hold up when someone loads a
different model?" Today the platform has no way to answer that other than
trying it and reading the reply. Phase 111/112/118 built discovery, guardrails,
and runtime (loaded/CPU) hints — good infrastructure, but none of it measures
**answer quality**. This phase adds that measurement and fixes the two things
most likely to make a small model look bad that aren't actually the model's
fault.

## Findings from this review

1. **RAG content is small and well-scoped — that's a point in phi3:mini's favor.**
   Field guides (`docs/field-guides/*.md`) are intentionally short (the apple
   nursery guide is 4 sentences); only `RAGTopK = 8` chunks are retrieved per
   turn (`internal/farmguardian/persona.go`), not the whole manifest. The
   grounded prompt is system prompt (~350 words) + live snapshot block + up to
   8 short chunks + capped read-tool text (`ReadToolsMaxAlerts=20`,
   `ReadToolsMaxPlants=20`, etc.) + conversation history. This is a modest,
   well-bounded prompt — squarely in small-model territory for **Q&A/lookup**.

2. **Guardian tools are not native LLM function-calling.** Read tools
   (`internal/farmguardian/readtools.go`) are regex-matched against the
   operator's message server-side and injected as text — the model never has
   to emit a tool-call schema to use them. This is a good design choice for
   small models (function-calling reliability scales hard with parameter
   count). The one place structured output is required is **write actions**:
   the model must emit a valid `ActionProposal` JSON object
   (`internal/farmguardian/proposals.go`) for the operator to Confirm. Small
   models are measurably worse at strict JSON schema adherence — this is the
   most likely place phi3:mini looks unreliable with "enough back-and-forth,"
   and there's currently no repair path (WS3).

3. **The context-window gate uses the wrong number for prompt sizing.**
   `GuardianMinContextWindow = 8192` gates whether grounded chat is *allowed*.
   phi3:mini reports `context_window: 131072` via rope scaling extension
   (documented in Phase 118), so it clears the gate — but its
   `rope.scaling.original_context_length` is 4096, meaning quality past that
   point is not well-supported by training even though the API accepts more
   tokens. Meanwhile, the *prompt itself* is built to a fixed shape regardless
   of which model answers. A future smaller/shorter-context model (e.g. the
   already-installed `tinyllama:latest`, 2048 real context per Phase 112/118
   fixtures) gets the same prompt size as an 8B model — nothing shrinks it.
   WS2 fixes this in one direction only: **trim down**, never expand past what
   a bigger model already gets.

## Design notes

- **WS1 fixture set:** reuse the demo farm + existing field guides as the
  grounded corpus; write ~20 questions covering: direct field-guide lookup
  (should cite), farm-state lookup via read tools (alerts/plants/low-stock),
  a deliberately out-of-scope question (should decline, not invent), and 2–3
  write-intent prompts that should produce a valid `ActionProposal`. Score
  automatically where possible (citation present? known zone/sensor names
  only? proposal parses?), flag the rest for human read of the report.
- **WS1 placement:** prefer a small Go CLI over another `//go:build ollama`
  smoke test — this needs to run on demand against whatever models are
  installed, not gate CI. `make guardian-eval MODEL=all` is the operator
  workflow.
- **WS2 budget:** derive from `ModelInfo.ContextWindow`, not
  `GuardianMinContextWindow`. Trim order: oldest history turns first, then
  RAG chunk count (never below 3), then snapshot detail (drop per-zone program
  lists before dropping alerts). Always log what was trimmed so the eval
  harness (WS1) can correlate quality drops with trimming.
- **WS3 repair:** one retry only, corrective message states the exact JSON
  shape expected and the parser error, nothing else — keep the correction
  prompt short so it doesn't eat the budget WS2 just protected.
- **WS4:** cache eval results per model name (with the same `:latest`
  normalization from Phase 118 WS2) in the existing model cache structure or a
  small table; freshly pulled models show "not yet evaluated — run
  `make guardian-eval`" rather than a stale or fabricated score.

## Out of scope

- Automated model *recommendations* ("switch to X") — scores are informational
- Continuous/scheduled re-evaluation — operator-triggered only this phase
- Fine-tuning or prompt-per-model customization — one prompt, sized per WS2

## Acceptance

- [ ] `make guardian-eval` runs the fixture set against every installed chat-capable model and writes a report (path printed at end of run)
- [ ] Report includes per-model: grounded-citation rate, out-of-scope decline rate, proposal-valid rate, mean latency
- [ ] Context budget guard measurably shrinks prompt token estimate for a model with `ContextWindow < 4096` in a unit test (mock cache), and leaves an 8B+ model's prompt unchanged
- [ ] One proposal-repair retry recovers a seeded malformed-JSON case in a unit test; a still-invalid second attempt fails the turn cleanly (existing error path)
- [ ] `GET /guardian/models` includes eval summary fields when available; selector shows them or "not yet evaluated"
- [ ] `TestPhase111_*`, `TestPhase112_*`, `TestPhase118_*` still green

## Files expected to change

| Area | Files |
|------|-------|
| Eval CLI | new `cmd/guardian-eval/main.go`, new `internal/farmguardian/eval/` package, fixture questions under `internal/farmguardian/eval/fixtures/` |
| Context budget | `internal/farmguardian/persona.go` (or new `prompt_budget.go`), `internal/handler/chat/handler.go` (wire trimming into prompt assembly) |
| Proposal repair | `internal/farmguardian/proposals.go`, `internal/handler/chat/handler.go` |
| Model info + UI | `internal/handler/chat/models.go`, `ui/src/components/GuardianModelSelector.vue` |
| Docs | `INSTALL.md`, `docs/plans/phase_118_guardian_model_capabilities.plan.md` (cross-reference) |
| Tests | `internal/farmguardian/prompt_budget_test.go`, `internal/farmguardian/proposals_test.go` (repair case), `ui/src/__tests__/phase-122-*.test.js` |
