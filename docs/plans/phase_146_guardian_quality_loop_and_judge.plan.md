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
    status: pending
  - id: ws2-warmup-eval-ops
    content: "WS2: Eval warmup timeout configurable; async warmup; source-local-env in smoke Makefile"
    status: pending
  - id: ws3-feedback-to-fixture
    content: "WS3: Export thumbs-down → candidate regression fixtures; runbook triage → score tests"
    status: pending
  - id: ws4-synthesis-prompt
    content: "WS4: Harden synthesis system prompt — no source dumps, stop after answer, cite [n] only"
    status: pending
  - id: ws5-regression-drift
    content: "WS5: Apply topicDriftNote to field_guide regression fixtures; phase127 agronomy prompts"
    status: pending
  - id: ws6-infra-hygiene
    content: "WS6: Phase 138 migration note in bootstrap; guardian_counsel_model tick fix doc"
    status: pending
  - id: ws7-closure
    content: "WS7: phase-146-closure.test.js; ci-guardian-qa judge section; mark 131 deferred judge superseded for GPU"
    status: pending
isProject: false
---

# Phase 146 — Guardian quality loop, judge & ops hardening

**Status:** **Planned** · **Depends on:** [145](phase_145_guardian_topic_drift_and_grounding.plan.md) · [134](phase_134_guardian_answer_feedback.plan.md) · [131](phase_131_guardian_qa_harness.plan.md)

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

## Problem statement

| Gap | Evidence | Target |
|-----|----------|--------|
| No automated "answers the question?" | Phase 131 deferred LLM-as-judge | Optional critique on Profile D; embed-only default on Profile A |
| Smoke warmup blocks 5m then skips | Run #3 log | `GUARDIAN_EVAL_WARMUP_TIMEOUT`; Makefile uses `source-local-env.sh` |
| Thumbs-down doesn't become tests | Phase 141 runbook manual only | Script: feedback export → `eval/fixtures_feedback.go` candidates |
| Model dumps sources verbatim | Run #3 ec-ph `[6] type=field_guide…` tail | System prompt + WS4 trim backup |
| Regression lacks drift checks | Only smoke-ec-ph keywords today | All `field_guide` + phase127 agronomy in `score.go` |
| `guardian_counsel_model` tick noise | Run #3 infra table | Bootstrap checklist + migration verify |

---

## Workstreams

### WS1 — Optional self-critique judge (GPU path)

**Where:** `internal/farmguardian/answer_critique.go`, env `GUARDIAN_ANSWER_CRITIQUE=1`.

| Profile | Behavior |
|---------|----------|
| **Profile A (cpu_laptop)** | Off by default; relevance from 145 only |
| **Profile D (server/GPU)** | After finalize, one short completion: "YES/NO: Does the answer address the question using only cited farm/doc facts? One sentence why." |
| Eval | When enabled, `critique_pass` in archive; smoke fail on NO |
| Cost guard | Respect existing token caps; max 128 completion tokens |

**Not** full rubric LLM-as-judge — single binary gate + reason string for turn debug.

**Tests:** mock chat client returns YES/NO; eval scores accordingly; default off in tests.

### WS2 — Warmup & eval ops

**Where:** `eval/runner.go`, `Makefile`, `scripts/source-local-env.sh`.

- `GUARDIAN_EVAL_WARMUP_TIMEOUT` (default 5m → 90s on cpu_laptop tune).
- `guardian-qa-smoke` target: `source scripts/source-local-env.sh --refresh-eval-token` preamble.
- Document stale JWT failure mode in smoke report template.
- Optional: fire warmup async, don't block first grounded prompt.

### WS3 — Feedback → fixture pipeline

**Where:** `scripts/guardian-feedback-to-fixture.sh`, `eval/fixtures_feedback.go`, runbook.

1. `GET /v1/chat/feedback/export?since=30d&rating=down`
2. Emit Go test stub or JSON fixture candidate per row (question, answer excerpt, reason chip).
3. Operator promotes row → `fixtures_feedback.go` + `score_*_test.go` in next phase.
4. Runbook § "Promote feedback to regression" checklist.

### WS4 — Synthesis prompt hardening

**Where:** `internal/rag/synthesis/synthesis.go` `systemPrompt`.

Add explicit rules:

- Answer the question in ≤ N paragraphs; **do not** append a Sources list.
- Use only `[n]` citations from the provided list; never invent `type=field_guide` lines.
- If sources conflict or are off-topic, say so briefly — do not elaborate on unrelated chunks.

**Tests:** `synthesis_test.go` prompt contains new constraints; smoke run #4 qualitative check.

### WS5 — Regression drift coverage

**Where:** `eval/score.go` — apply `smokeTopicDriftNote` (from 145) to:

- All `Category == "field_guide"` fixtures
- Phase 127 agronomy prompts (`p128-fert-triage`, etc.)

Keyword blocklist from 144 becomes subset of drift note.

### WS6 — Infra hygiene

- `local-operator-bootstrap.md`: verify Phase 138 migration before smoke (`guardian_counsel_model`).
- `make migrate` in smoke preflight doc.
- Silence or fix background tick if column missing (graceful degrade already? verify).

### WS7 — Closure

- Update `docs/ci-guardian-qa.md` — judge section: embed default, critique optional on GPU.
- Phase 131 plan footnote: "146 supersedes deferred judge for GPU profile only."
- `phase-146-closure.test.js`.
- Mark **Shipped** when WS1–5 + docs done (WS6 ops optional if migration already applied locally).

---

## Acceptance

- [ ] `GUARDIAN_ANSWER_CRITIQUE=0` (default): smoke unchanged except 145 gates.
- [ ] `GUARDIAN_ANSWER_CRITIQUE=1` on GPU: run #3 ec-ph would fail critique YES/NO.
- [ ] `make guardian-qa-smoke` refreshes eval token via `source-local-env.sh`.
- [ ] Feedback export script produces ≥1 fixture candidate from down-vote row.
- [ ] Synthesis system prompt includes no-source-dump rule (unit test).
- [ ] Regression `field_guide` fixtures run drift note (unit tests with archived answers).

---

## Suggested implementation order

1. WS4 synthesis prompt (prevention, low risk).
2. WS2 eval ops (unblock reliable smoke).
3. WS5 regression drift (extends 145 scorer).
4. WS3 feedback pipeline (docs + script).
5. WS1 optional critique (feature-flagged).
6. WS6 infra docs.
7. WS7 closure.

---

## Non-goals (Phase 146)

- Mandatory LLM judge on every PR / CPU laptop.
- Auto-merge feedback into fixtures without human promote step.
- Fine-tuning or model swap.
- Full `guardian-qa-regression` on every commit.

---

## Roadmap slice (143 → 146)

```mermaid
flowchart LR
  P143[Phase 143 hygiene] --> P144[Phase 144 residuals]
  P144 --> P145[Phase 145 embed drift]
  P145 --> P146[Phase 146 judge + loop]
  P146 --> Smoke4[Smoke run 4 green]
```

| Phase | Focus |
|-------|--------|
| 143 | Leak, gr33n.com, pH heuristic, warmup model |
| 144 | gr33n-docs, apology trim, keyword drift |
| 145 | **Embed relevance, citation align, RAG filter, tail trim** |
| 146 | **Critique optional, ops, feedback fixtures, prompt, regression** |
