---
name: Phase 144 — Guardian answer quality (run #3 residuals)
overview: >
  Phase 143 run #3 passed 4/4 heuristics but human review found gr33n-docs/ fake
  markdown links, model apology tails on morning-walk, and ec-ph topic drift into
  unrelated endocrine-disruptor RAG. Extend finalize hygiene and smoke heuristics.
todos:
  - id: ws1-gr33n-docs-citations
    content: "WS1: Sanitize gr33n-docs/ hallucinated markdown links like gr33n.com"
    status: completed
  - id: ws2-meta-correction-trim
    content: "WS2: Trim/detect apology + 'updated answer' tails before persist; turn debug flag"
    status: completed
  - id: ws3-ecph-topic-drift
    content: "WS3: smoke-ec-ph heuristic — fail off-topic endocrine / lake drift"
    status: completed
  - id: ws4-docs-tests-closure
    content: "WS4: Runbook + architecture pointer; score tests; phase-144-closure.test.js"
    status: completed
isProject: false
---

# Phase 144 — Guardian answer quality (run #3 residuals)

**Status:** **Shipped.** · **Depends on:** [143](phase_143_guardian_answer_quality.plan.md)

**Evidence:** [`guardian-qa-smoke-report-20260707.md`](../guardian-qa-smoke-report-20260707.md) run #3 human-review gaps.

---

## Problem statement

| Issue | Run #3 | Risk |
|-------|--------|------|
| `gr33n-docs/…` links | morning-walk | Fake doc paths look like real citations |
| Meta apology tail | morning-walk ends with “I apologize… Here's an updated answer” | Broken UX |
| ec-ph topic drift | Opening OK; tail hallucinates endocrine / Lake Erie content | Misleading agronomy |

---

## Workstreams

### WS1 — `gr33n-docs` citation hygiene ✅

**Shipped:** `isHallucinatedCitationURL` includes `gr33n-docs`; smoke morning-walk fails `AnswerContainsFakeCitationURL`.

### WS2 — Meta correction trim ✅

**Shipped:** `TrimMetaCorrection` in finalize chain; `meta_correction_trimmed` on dev turn debug; smoke morning-walk fails `AnswerContainsMetaCorrection`.

### WS3 — ec-ph topic drift ✅

**Shipped:** `smokeECPHQualityNote` — fail on endocrine / lake drift terms from run #3 archive.

### WS4 — Docs & closure ✅

**Shipped:** Runbook checklist updated; architecture §8.8 pointer; `score_smoke_quality_test.go` run #3 fixtures; `phase-144-closure.test.js`.

---

## Acceptance

- [x] Run #3 morning-walk text **fails** new morning-walk heuristics (gr33n-docs + apology).
- [x] Run #3 ec-ph drift excerpt **fails** `smokeECPHQualityNote`.
- [x] `SanitizeCitationURLs` rewrites `gr33n-docs` links to plain labels.
- [x] Persist path trims apology tails before `conversation_turns`.

---

## Non-goals

- LLM-as-judge for all prompts.
- Re-run full smoke on CPU in this phase (operator task).
- Blocking all long answers — only ec-ph drift keywords v1.
