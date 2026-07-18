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
    content: "WS1: Eval harness — cmd/guardian-eval CLI runs ~20 grounded questions against installed chat-capable models; writes JSON report"
    status: completed
  - id: ws2-context-budget
    content: "WS2: Context budget guard — trim history/RAG/snapshot when context_window < 8192; log trims"
    status: completed
  - id: ws3-proposal-repair
    content: "WS3: Proposal JSON repair — one retry with corrective system message on parse failure"
    status: completed
  - id: ws4-selector-quality-badge
    content: "WS4: Selector quality badge — eval summary on GET /guardian/models and GuardianModelSelector.vue"
    status: completed
  - id: ws5-docs
    content: "WS5: Docs — INSTALL.md section on eval scores and phi3 rope quirk"
    status: completed
isProject: false
---

# Phase 122 — Guardian model eval harness, context budget guard, and proposal repair

**Status: shipped**

## Acceptance

- [x] `make guardian-eval` runs the fixture set against installed chat-capable models and writes `data/guardian_model_eval.json`
- [x] Report includes per-model: grounded-citation rate, out-of-scope decline rate, proposal-valid rate, mean latency
- [x] Context budget guard shrinks prompt for `ContextWindow < 4096` (unit test); large windows unchanged
- [x] One proposal-repair retry recovers malformed JSON in unit test
- [x] `GET /guardian/models` includes eval summary; selector shows scores or "not yet evaluated"
- [x] Phase 111/112/118 tests unchanged (run via `make ollama-smoke`)
