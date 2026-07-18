---
name: Phase 147 — Smoke run #5 closure & eval isolation
overview: >
  Phase 146 added eval client timeout buffer and Makefile JWT refresh, but smoke run #4
  still ended 3/4 (smoke-ec-ph client timeout after ~103 min). Phase 147 isolates
  single-prompt re-runs, documents laptop timeout tuning, shows critique in Settings QA,
  and closes the 143–147 answer-quality arc with run #5 evidence.
todos:
  - id: ws1-prompt-id-filter
    content: "WS1: -prompt-ids / GUARDIAN_EVAL_PROMPT_IDS filter in guardian-eval"
    status: completed
  - id: ws2-make-ec-ph
    content: "WS2: make guardian-qa-smoke-ec-ph; GUARDIAN_EVAL_TIMEOUT_SECONDS=2100 in laptop tune"
    status: completed
  - id: ws3-settings-critique
    content: "WS3: Settings Guardian QA — Critique column when critique_pass in archive"
    status: completed
  - id: ws4-smoke-run5
    content: "WS4: Re-run smoke-ec-ph (run #5); update smoke report § run #5"
    status: completed
  - id: ws5-closure
    content: "WS5: phase-147-closure.test.js; architecture §8.11; phase-14 index; mark 143–147 arc"
    status: completed
isProject: false
---

# Phase 147 — Smoke run #5 closure & eval isolation

**Status:** **Shipped.** · **Depends on:** [146](phase_146_guardian_quality_loop_and_judge.plan.md) · [145](phase_145_guardian_topic_drift_and_grounding.plan.md)

---

## Why this phase

Run #4 proved Phase 145 drift stack on CPU but **`smoke-ec-ph` never scored** — eval HTTP client timed out after three long grounded prompts. Phase 146 added `ClientTimeoutFromEnv` (+15m buffer) and async warmup; operators still need a **fast path** to re-run one prompt without a full ~2h smoke.

| Gap (run #4) | Phase 147 fix |
|--------------|---------------|
| ec-ph client timeout | `GUARDIAN_EVAL_TIMEOUT_SECONDS` laptop tune + isolated re-run |
| No single-prompt Make target | `make guardian-qa-smoke-ec-ph` |
| Critique in archive, not Settings | **Critique** column on QA last-run card |
| Quality arc open after 146 | Run #5 + docs closure for phases **143–147** |

---

## Workstreams

### WS1 — Prompt ID filter ✅

**Shipped:** `eval/filter.go` — `FilterFixturesByIDs`; `guardian-eval -prompt-ids` / `GUARDIAN_EVAL_PROMPT_IDS`.

### WS2 — Make target & laptop tune ✅

**Shipped:** `make guardian-qa-smoke-ec-ph`; `scripts/tune-guardian-laptop.sh` ensures `GUARDIAN_EVAL_TIMEOUT_SECONDS>=2100` on `cpu-16gb`.

### WS3 — Settings critique column ✅

**Shipped:** `GuardianSettingsQARunCard.vue` — **Critique** column when `critique_pass` present (GPU + `GUARDIAN_ANSWER_CRITIQUE=1`).

### WS4 — Smoke run #5 ✅

Re-ran `make guardian-qa-smoke-ec-ph` — **25.1 min**, archive `20260708T130745_smoke_phi3-mini.json`. **No client timeout** (fixes run #4 ops gap). Heuristic fail on `uncited_tail` (documented in [smoke report](../guardian-qa-smoke-report-20260707.md) § run #5).

### WS5 — Closure ✅

**Shipped:** `phase-147-closure.test.js`; architecture §8.11; phase-14 index; **143–147** quality arc marked shipped.

---

## Acceptance

- [x] `-prompt-ids smoke-ec-ph` runs one fixture from smoke suite.
- [x] `make guardian-qa-smoke-ec-ph` refreshes JWT like full smoke.
- [x] Laptop tune recommends `GUARDIAN_EVAL_TIMEOUT_SECONDS=2100`.
- [x] Settings QA shows Critique when archived.
- [x] Run #5 ec-ph completed without client timeout; results in smoke report § run #5.
- [x] Phase 147 plan marked shipped; closure test green.

---

## Non-goals

- Mandatory full smoke on every commit.
- Enabling `GUARDIAN_ANSWER_CRITIQUE=1` on CPU laptop by default.
- New regression fixtures beyond run #5 evidence.
