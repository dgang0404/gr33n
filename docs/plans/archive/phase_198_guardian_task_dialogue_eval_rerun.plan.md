---
name: Phase 198 — Re-run scenario-task-dialogue-pending + docs/CI notes
overview: >
  Verify Phase 192 fixes the live eval failure on scenario-task-dialogue-pending
  (title "due tomorrow" vs want "Refill calcium nitrate"). Document expected
  CPU runtime and operator steps for guardian-qa-change-requests-ui on laptop
  stacks. Not a code-heavy phase — primarily validation + documentation.
todos:
  - id: ws1-rerun-eval
    content: "WS1: run make guardian-qa-change-requests-ui with -prompt-ids scenario-task-dialogue-pending -leave-pending after Phase 192 ships"
    status: completed
  - id: ws2-archive-assertions
    content: "WS2: confirm data/guardian_qa_runs archive shows passed=true and proposal args title + WantDueDateOffsetDays"
    status: completed
  - id: ws3-docs-ci
    content: "WS3: update docs/ci-guardian-qa.md, current-state.md, operator-tour §7s with runtime note (~2h CPU, 4 turns) and failure signature from 2026-07-16 run"
    status: completed
  - id: ws4-closure
    content: "WS4: phase-198-closure.test.js — documents eval command + scenario expectations in repo"
    status: completed
isProject: false
---

# Phase 198 — Re-run task dialogue eval + docs

**Status:** shipped · **Depends on:** [192](phase_192_guardian_due_date_title_clobber.plan.md) · **Blocks:** operator sign-off on sit-in arc

## The problem

2026-07-16 live run:

```
eval: scenario "scenario-task-dialogue-pending" fail (proposals=1)
notes: proposal title "due tomorrow" want "Refill calcium nitrate"
```

All four turns completed (~2.3 hours on CPU). Proposal left pending correctly
at revision 4, but **title assertion failed** — blocking confidence in the
187→192 revise arc.

## What to ship

### WS1 — Re-run eval

```bash
make guardian-qa-change-requests-ui
# or subset:
# guardian-eval -suite change-requests-ui -prompt-ids scenario-task-dialogue-pending -leave-pending
```

Prerequisites: API + Ollama + DB up; `GUARDIAN_EVAL_TOKEN` set.

### WS2 — Pass criteria

Archive JSON must show:

- `passed: true`
- `WantTitle` match: `Refill calcium nitrate`
- `WantDueDateOffsetDays: 1` → `due_date` tomorrow UTC
- `RequireTaskZone` satisfied
- `MinRevision: 4`
- `leave_pending` row visible in UI Pending tab

### WS3 — Documentation

Add to `docs/ci-guardian-qa.md`:

| Scenario | Turns | CPU time (phi3:mini laptop) | Notes |
|----------|-------|----------------------------|-------|
| scenario-task-dialogue-pending | 4 | ~90–120 min/turn observed | leave-pending for UI |

Document 2026-07-16 failure as **fixed in Phase 192** (title clobber).

### WS4 — Closure test

Vitest reads fixture + Makefile target strings — no live LLM in CI.

## Acceptance criteria

- [x] Eval scenario passes after 192 (2026-07-17: `passed: true` after API restart; stale-API re-run documented)
- [x] Pending proposal in DB has correct title + due_date + zone (asserted by eval `WantTitle` / `RequireTaskZone` / `WantDueDateOffsetDays`)
- [x] Docs reflect observed runtime (not a CI gate — manual/optional target)

## Out of scope

- Speeding up eval (GPU, smaller model) — separate backlog
- Running full 5-scenario change-requests-ui suite in CI
