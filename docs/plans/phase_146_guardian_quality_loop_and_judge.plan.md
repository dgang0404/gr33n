---
name: Phase 146 — Guardian quality loop, judge & ops hardening
overview: >
  After Phase 145 embedding relevance and citation alignment, close the remaining
  quality loop: optional lightweight self-critique judge on GPU profile, CPU-safe
  warmup/eval ops, feedback-export→fixture pipeline, synthesis prompt hardening,
  and regression-suite drift coverage beyond smoke.
todos:
  - id: ws1-optional-self-critique
    content: "WS1: Optional GUARDIAN_ANSWER_CRITIQUE=1 — single yes/no+reason LLM pass on GPU; skip on cpu_laptop"
    status: completed
  - id: ws2-warmup-eval-ops
    content: "WS2: Eval warmup timeout configurable; async warmup; source-local-env in smoke Makefile"
    status: completed
  - id: ws3-feedback-to-fixture
    content: "WS3: Export thumbs-down → candidate regression fixtures; runbook triage → score tests"
    status: completed
  - id: ws4-synthesis-prompt
    content: "WS4: Harden synthesis system prompt — no source dumps, stop after answer, cite [n] only"
    status: completed
  - id: ws5-regression-drift
    content: "WS5: Apply topicDriftNote to field_guide regression fixtures; phase127 agronomy prompts"
    status: completed
  - id: ws6-infra-hygiene
    content: "WS6: Phase 138 migration note in bootstrap; guardian_counsel_model tick fix doc"
    status: completed
  - id: ws7-closure
    content: "WS7: phase-146-closure.test.js; ci-guardian-qa judge section; mark 131 deferred judge superseded for GPU"
    status: completed
isProject: false
---

# Phase 146 — Guardian quality loop, judge & ops hardening

**Status:** **Shipped.** · **Depends on:** [145](phase_145_guardian_topic_drift_and_grounding.plan.md) · [134](phase_134_guardian_answer_feedback.plan.md) · [131](phase_131_guardian_qa_harness.plan.md)

---

## Why a second phase

Phase 145 stays **embedding-first** (cheap, deterministic, CPU-safe). Some gaps need **policy**, **ops**, or **optional LLM critique** — better as a follow-on so 145 can ship incrementally.

| Item | Phase | Rationale |
|------|-------|-----------|
| Cosine relevance + citation align | **145** | Uses existing embedder; no new model call |
| Self-critique "does this answer the question?" | **146** | Extra LLM latency; GPU profile first |
| Warmup 5m timeout on CPU smoke | **146** | Ops/Makefile; not drift logic |
| Feedback → regression fixtures | **146** | Closes 134/141 human loop into code |
| Synthesis prompt rewrite | **146** | Prevention at source; test after 145 metrics exist |

---

## Workstreams

### WS1 — Optional self-critique judge ✅

**Shipped:** `internal/farmguardian/answer_critique.go` — `GUARDIAN_ANSWER_CRITIQUE=1`; `CritiqueAnswer` YES/NO gate; turn debug `critique_pass` / `critique_reason`; eval fails on NO when enabled.

### WS2 — Warmup & eval ops ✅

**Shipped:** `eval/env.go` — `WarmupTimeoutFromEnv`, `ClientTimeoutFromEnv` (+15m buffer); async warmup on smoke/phase127; `make guardian-qa-smoke` / `guardian-qa-phase127` refresh JWT via `source-local-env.sh`; `GUARDIAN_EVAL_WARMUP_TIMEOUT` in laptop tune script.

### WS3 — Feedback → fixture pipeline ✅

**Shipped:** `scripts/guardian-feedback-to-fixture.sh`; `GET /v1/chat/feedback/export?rating=down`; `eval/fixtures_feedback.go` stub; runbook § Promote feedback to regression.

### WS4 — Synthesis prompt hardening ✅

**Shipped:** `synthesis.go` system prompt — no `Sources:` dumps, max four paragraphs, cite `[n]` only; `synthesis_test.go`.

### WS5 — Regression drift coverage ✅

**Shipped:** `score.go` — drift on all `field_guide` + phase127 IDs; `score_regression_drift_test.go`.

### WS6 — Infra hygiene ✅

**Shipped:** `local-operator-bootstrap.md` — migrate before smoke (Phase 138 `guardian_counsel_model`).

### WS7 — Closure ✅

**Shipped:** `phase-146-closure.test.js`; `ci-guardian-qa.md` judge section; Phase 131 footnote superseded for GPU critique only.

---

## Acceptance

- [x] `GUARDIAN_ANSWER_CRITIQUE=0` (default): smoke unchanged except 145 gates.
- [x] `GUARDIAN_ANSWER_CRITIQUE=1` on GPU: run #3 ec-ph would fail critique YES/NO (unit test).
- [x] `make guardian-qa-smoke` refreshes eval token via `source-local-env.sh`.
- [x] Feedback export script produces fixture candidates from down-vote rows.
- [x] Synthesis system prompt includes no-source-dump rule (unit test).
- [x] Regression `field_guide` fixtures run drift note (unit tests with archived-style answers).

---

## Non-goals (Phase 146)

- Mandatory LLM judge on every PR / CPU laptop.
- Auto-merge feedback into fixtures without human promote step.
- Fine-tuning or model swap.
- Full `guardian-qa-regression` on every commit.
