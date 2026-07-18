---
name: Phase 196 — Pending proposal revision timeline on card
overview: >
  Proposal cards show revision badge and args diff vs previous revision, but
  not the operator chat turns that drove each revision. After Phase 194 adds
  View conversation, operators still leave Pending to read a long transcript.
  This phase adds a compact inline timeline on the Pending card.
todos:
  - id: ws1-api-session-turns-for-proposal
    content: "WS1: UI loads session turns when expanding timeline — reuse GET /v1/chat/sessions/{id} (no new API unless perf requires); optional lazy fetch on expand"
    status: completed
  - id: ws2-timeline-component
    content: "WS2: GuardianProposalRevisionTimeline.vue — collapsible 'Revision history (N turns)' showing user_message snippets + which args changed per supersede (title/zone/due_date)"
    status: completed
  - id: ws3-map-turns-to-revisions
    content: "WS3: correlate turns after turn 0 with proposal revision increments — heuristic: turns with handled revise show in timeline; link proposal.revision to turn_index"
    status: completed
  - id: ws4-tests-docs
    content: "WS4: phase-196-closure.test.js + operator-tour; fixture from scenario-task-dialogue-pending session"
    status: completed
isProject: false
---

# Phase 196 — Pending proposal revision timeline

**Status:** shipped · **Depends on:** [194](phase_194_pending_view_conversation.plan.md) · **Related:** [192](phase_192_guardian_due_date_title_clobber.plan.md)

## The problem

Pending card today shows:

- `REVISION 4` badge
- `Changed from previous revision` args diff (when `previous_args` available)
- Operator-stated facts from `meta`

It does **not** show:

- Turn 1: "Put it in Veg Room…"
- Turn 2: "call it Refill calcium nitrate instead"
- Turn 3: "make it due tomorrow"

Operators reviewing a multi-turn pending request need that story **without**
reading the full Guardian assistant replies (often noisy on phi3:mini).

## What to ship

### WS1 — Data source

Use existing `GET /v1/chat/sessions/{session_id}` — returns `turns[]` with
`user_message`, `assistant_message`, `turn_index`.

Lazy-load when user expands timeline (avoid N+1 on inbox load).

### WS2 — UI component

Collapsible section on `GuardianActionProposal` when `revision > 1`:

```
Revision history ▾
  1. You: Create a task to refill calcium nitrate…
  2. You: Put it in Veg Room — zone for this task
     → zone_id set
  3. You: call it Refill calcium nitrate instead
     → title updated
  4. You: make it due tomorrow
     → due_date set
```

Show **user lines** prominently; assistant lines collapsed or omitted by default.

### WS3 — Revision correlation

Options (pick simplest in implementation):

- **A (MVP):** List all `user_message` turns from session; annotate with
  current `args` keys that differ from empty — no per-turn diff without
  supersede chain API.
- **B:** `GET /v1/chat/proposals?session_id=` returns superseded chain with
  args per revision — diff adjacent revisions for timeline annotations.

Start with **A**; upgrade to **B** if proposal list API already exposes
supersedes chain (check `ListProposals` / proposal detail).

### WS4 — Tests & docs

- Collapsed by default; expand shows turn count
- `data-test="guardian-proposal-revision-timeline"`

## Acceptance criteria

- [x] scenario-task-dialogue-pending pending card shows 4 user turns when expanded
- [x] Timeline loads without blocking Confirm/Refine
- [x] View conversation ([194]) still opens full transcript

## Out of scope

- Showing assistant message quality ratings
- Editing from timeline (use Refine)
