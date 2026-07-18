---
name: Phase 200 — accuracy_note round-trip verification and gaps
overview: >
  Phase 152 added live AnswerAccuracyNote on every chat turn; Phase 159 added
  accuracy_note column and handler persistence. Architecture doc §8.16 still
  claimed accuracy_note is not persisted. This phase audited the full round-trip
  (compute → persist → session reload → UI banner), fixed eval-archive gap, and
  updated stale docs.
todos:
  - id: ws1-audit-persist-reload
    content: "WS1: verify persistTurn writes accuracy_note; GET session returns it; GuardianChatPanel shows banner on reload not only on live SSE done event"
    status: completed
  - id: ws2-revise-turn-gap
    content: "WS2: if revise-only turns skip accuracy_note pipeline, wire applyAnswerAccuracyNote on those code paths too"
    status: completed
  - id: ws3-eval-archive
    content: "WS3: guardian-eval QA archive includes accuracy_note per turn when present — for offline quality review"
    status: completed
  - id: ws4-docs-closure
    content: "WS4: fix farm-guardian-architecture.md §8.16 limitation; phase-200-closure.test.js"
    status: completed
isProject: false
---

# Phase 200 — accuracy_note round-trip

**Status:** shipped · **Depends on:** [152](phase_152_guardian_live_accuracy_guardrails.plan.md), [159](phase_159_guardian_citation_completeness.plan.md)

## The problem

`docs/farm-guardian-architecture.md` §8.16 states:

> `accuracy_note` isn't persisted to `conversation_turns` yet

But Phase 159 shipped:

- Migration `20260711_phase159_accuracy_note.sql`
- `persistTurn(..., accuracyNote)` in handler
- Session reload maps `row.AccuracyNote` → turn payload
- UI: `accuracyNoteMessage(t.accuracy_note)` banner

**Gap may be documentation drift**, not missing column — or subtle bugs:

- Banner only on `finalEvent.accuracy_note` from live stream, not reloaded turns?
- Revise-handled turns that bypass full synthesis?
- Eval archive omitting notes?

Live DB sample (2026-07-16): some turns have `accuracy_note` populated
(e.g. `citation_number_mismatch`), others empty.

## What to ship

### WS1 — End-to-end audit

1. Send grounded message that triggers `dangling_list_intro` or `citation_number_mismatch`
2. Confirm DB row has `accuracy_note`
3. Reload session — banner must still show
4. Fix `GuardianChatPanel` if reload path drops the field

### WS2 — Revise path

Trace `tryReviseActiveProposal` handled turns — do they persist turns with
accuracy notes when LLM still answers? Ensure consistency.

### WS3 — Eval archive (optional)

Include `accuracy_note` in `guardian_qa_runs` JSON per scored turn for
operator quality review runbook.

### WS4 — Docs

- Remove incorrect "not persisted" from architecture doc
- `current-state.md` one-liner under 188–191 arc or 152

## Acceptance criteria

- [x] accuracy_note survives session switch and page refresh
- [x] Architecture doc matches reality
- [x] phase-200-closure.test.js references migration + handler + UI reload path

## Out of scope

- Storing accuracy_note on `guardian_action_proposals`
- Historical backfill of old turns
