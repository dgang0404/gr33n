---
name: Phase 143 — Guardian answer quality (post-smoke hygiene)
overview: >
  Smoke heuristics passed 4/4 on phi3:mini CPU (2026-07-07) but human review found
  instruction template leaks, hallucinated citation URLs, warmup 503 when env default
  model differs from eval model, missing pH in ec-ph answers, and lenient pass rules.
  Tighten persistence, citations, warmup, and eval scoring so smoke pass implies
  operator-trustworthy answers.
todos:
  - id: ws1-instruction-leak-guard
    content: "WS1: Strip/detect template echoes (## Your task, Question:) before persisting assistant turns; warn in turn debug"
    status: pending
  - id: ws2-citation-url-hygiene
    content: "WS2: Block or rewrite fake gr33n.com markdown links; prefer [source#N] only in rendered answers"
    status: pending
  - id: ws3-warmup-eval-model
    content: "WS3: guardian-eval warmup passes explicit model; fix 503 when farm counsel model rejects tinyllama ctx floor"
    status: pending
  - id: ws4-smoke-heuristics
    content: "WS4: eval score.go — no_prompt_leak, no_fake_url, ec-ph requires ph; keep walk_farm log override"
    status: pending
  - id: ws5-feedback-review-run
    content: "WS5: Document post-smoke feedback checklist; optional Settings nudge after QA archive write"
    status: pending
  - id: ws6-closure
    content: "WS6: Re-run make guardian-qa-smoke; update smoke report; architecture § pointer; phase-143-closure test"
    status: pending
isProject: false
---

# Phase 143 — Guardian answer quality

**Status:** **Planned** · **Depends on:** [131](phase_131_guardian_qa_harness.plan.md) (smoke harness), [129](phase_129_guardian_awakening.plan.md) (warmup)

**Evidence:** [`guardian-qa-smoke-report-20260707.md`](../guardian-qa-smoke-report-20260707.md) — run #2 **4/4 heuristic pass**, quality gaps documented.

**129–139 closure:** Smoke green on laptop; this phase closes the **quality** gap, not the harness gap.

---

## Problem statement

| Issue | Smoke run #2 | Risk |
|-------|----------------|------|
| Instruction template leak | `smoke-morning-walk` ends with `## Your task:Given the sources...` | Broken UX; looks like debug output |
| Fake URLs | `https://gr33n.com/tasks`, `gr33n.com/sources/field_guide` | Misleading citations |
| Warmup 503 | `POST /guardian/warmup` before grounded block; env `LLM_MODEL=tinyllama` | Cold model; relies on inline warmup |
| ec-ph missing pH | Passed on EC + citations only | Incomplete agronomy answer |
| Lenient heuristics | morning-walk pass despite leak/URLs | False confidence |

---

## Workstreams

### WS1 — Instruction leak guard

**Where:** `internal/handler/chat` turn persist path (after LLM, before DB write).

- Detect markers: `## Your task`, `Question:\n` echo of user prompt, repeated system-instruction blocks.
- **Strip** trailing leak from stored `answer` (and streamed final if applicable).
- Log `guardian: answer_leak_trimmed` with chars removed.
- Expose `leak_trimmed: true` in dev turn debug (Phase 139).

**Tests:** unit table on sample leaked morning-walk text; smoke does not regress latency.

### WS2 — Citation URL hygiene

**Where:** answer post-process or markdown render path.

- Reject/rewrite markdown links whose host is not allowlisted (empty allowlist → strip `https://gr33n.com/*` hallucinations).
- Keep numeric refs `[1]`, `[source#N]`, `per our field guide` phrasing.
- Optional: normalize `([live farm snapshot](#))` to plain text.

**Tests:** `answer_sanitize_test.go` with fake URL fixtures.

### WS3 — Warmup + eval model alignment

**Where:** `internal/farmguardian/eval/warmup.go`, `POST /guardian/warmup` body.

- `WarmupFarmCounsel` accepts optional `chat_model` from eval runner (same as `-models` flag).
- When unset, use farm `guardian_counsel_model` or eval model, not only env default.
- Document: smoke should pre-warm **phi3:mini** even when `.env` has `tinyllama:latest`.

**Tests:** eval warmup unit with mock HTTP; awakening test for counsel model resolution.

### WS4 — Tighter smoke heuristics

**Where:** `internal/farmguardian/eval/score.go`, `fixtures_smoke.go` notes.

| Prompt | Add |
|--------|-----|
| `smoke-morning-walk` | Fail if `looksLikePromptLeak(a)` or `fakeCitationURL(a)` |
| `smoke-ec-ph` | Require `ph` in answer (in addition to EC/citation) |
| All grounded | Optional: `no_fake_url` shared helper |

Keep `walk_farm` log-evidence override only when answer passes leak/URL checks.

**Tests:** `score_test.go` with archived run #2 morning-walk text → should **fail** new rules until WS1–2 ship.

### WS5 — Feedback review loop

**Where:** docs only (+ optional UI toast).

- Extend [`guardian-feedback-review-runbook.md`](../guardian-feedback-review-runbook.md) with smoke quality checklist (leak, URLs, pH).
- After `SaveQARunArchive`, log reminder line (already have `feedback_review_prompt` in JSON).

### WS6 — Closure

- Re-run `make guardian-qa-smoke MODEL=phi3:mini FARM_ID=1` on CPU profile.
- Update smoke report with run #3 results.
- One paragraph in `farm-guardian-architecture.md` § answer hygiene.
- `ui/src/__tests__/phase-143-closure.test.js` — plan file + report link present.

---

## Acceptance

- [ ] `make guardian-qa-smoke` **4/4** with **no** prompt leak or fake URL on morning-walk (archived JSON proof).
- [ ] `smoke-ec-ph` answer mentions pH targets or ranges.
- [ ] Eval warmup returns 200/202 (not 503) when `MODEL=phi3:mini` and env default is tinyllama.
- [ ] Run #2 morning-walk text **fails** new `score.go` tests (regression guard).

---

## Non-goals

- LLM-as-judge (Phase 131 deferred).
- Full `make guardian-qa-regression` on every PR.
- Model swap (still phi3:mini on CPU).
- Production turn debugger for all users.
- Git history secret purge (operator task if repo was public).

---

## Suggested implementation order

1. WS4 tests first (red) using archived run #2 answers.
2. WS1 + WS2 (make tests green).
3. WS3 warmup.
4. WS5 docs + WS6 closure smoke.
